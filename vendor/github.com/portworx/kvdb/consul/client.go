package consul

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/portworx/kvdb"
	"github.com/sirupsen/logrus"
)

const (
	// httpError is a substring returned by consul during such http errors.
	// Ideally such errors should be provided as consul constants
	httpError = "Unexpected response code: 500"
	// eofError is also a substring returned by consul during EOF errors.
	eofError = "EOF"
	// connRefused connection refused
	connRefused = "connection refused"
	// keyIndexMismatch indicates consul error for key index mismatch
	keyIndexMismatch = "Key Index mismatch"
	// nameResolutionError indicates no host found, can be temporary
	nameResolutionError = "no such host"
	// connReset connection reset by peer
	connReset = "connection reset by peer"
)

// clientConsul defines methods that a px based consul client should satisfy.
type consulClient interface {
	kvOperations
	// sessionOperations includes methods methods from that interface.
	sessionOperations
	// metaOperations includes methods from that interface.
	metaOperations
	// lockOptsOperations includes methods from that interface.
	lockOptsOperations
}

type kvOperations interface {
	// Get exposes underlying KV().Get but with reconnect on failover.
	Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error)
	// Put exposes underlying KV().Put but with reconnect on failover.
	Put(p *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error)
	// Acquire exposes underlying KV().Acquire but with reconnect on failover.
	Acquire(p *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error)
	// Delete exposes underlying KV().Delete but with reconnect on failover.
	Delete(key string, w *api.WriteOptions) (*api.WriteMeta, error)
	// DeleteTree exposes underlying KV().DeleteTree but with reconnect on failover.
	DeleteTree(prefix string, w *api.WriteOptions) (*api.WriteMeta, error)
	// Keys exposes underlying KV().Keys but with reconnect on failover.
	Keys(prefix, separator string, q *api.QueryOptions) ([]string, *api.QueryMeta, error)
	// List exposes underlying KV().List but with reconnect on failover.
	List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error)
}

type sessionOperations interface {
	// Create exposes underlying Session().Create but with reconnect on failover.
	Create(se *api.SessionEntry, q *api.WriteOptions) (string, *api.WriteMeta, error)
	// Destroy exposes underlying Session().Destroy but with reconnect on failover.
	Destroy(id string, q *api.WriteOptions) (*api.WriteMeta, error)
	// Renew exposes underlying Session().Renew but with reconnect on failover.
	Renew(id string, q *api.WriteOptions) (*api.SessionEntry, *api.WriteMeta, error)
	// RenewPeriodic exposes underlying Session().RenewPeriodic but with reconnect on failover.
	RenewPeriodic(initialTTL string, id string, q *api.WriteOptions, doneCh chan struct{}) error
}

type metaOperations interface {
	// CreateMeta is a meta writer wrapping KV().Acquire and Session().Destroy but with reconnect on failover.
	CreateMeta(id string, p *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, bool, error)
	// CompareAndSet is a meta func wrapping KV().CAS and KV().Get but with reconnect on failover.
	CompareAndSet(id string, value []byte, p *api.KVPair, q *api.WriteOptions) (bool, *api.WriteMeta, error)
	// CompareAndDelete is a meta func wrapping KV().DeleteCAS and KV().Get but with reconnect on failover.
	CompareAndDelete(id string, value []byte, p *api.KVPair, q *api.WriteOptions) (bool, *api.WriteMeta, error)
}

type lockOptsOperations interface {
	// LockOpts returns pointer to underlying Lock object and an error.
	LockOpts(opts *api.LockOptions) (*api.Lock, error)
}

// consulConnection stores current consul connection state
type consulConnection struct {
	// config is the configuration used to create consulClient
	config *api.Config
	// client provides access to consul api
	client *api.Client
	// once is used to reconnect consulClient only once among concurrently running threads
	once *sync.Once
}

// consulClient wraps config information and consul client along with sync functionality to reconnect it once.
// consulClient also satisfies interface defined above.
type consulClientImpl struct {
	// conn current consul connection state
	conn *consulConnection
	// connParams holds all params required to obtain new api client
	connParams connectionParams
	// reconnectDelay is the time duration to wait between machines
	reconnectDelay time.Duration
	// maxRetries is the number of times reconnect should be tried.
	maxRetries int
}

// newConsulClient provides an instance of clientConsul interface.
func newConsulClient(config *api.Config,
	client *api.Client,
	reconnectDelay time.Duration,
	p connectionParams,
) consulClient {
	c := &consulClientImpl{
		conn: &consulConnection{
			config: config,
			client: client,
			once:   new(sync.Once)},
		connParams:     p,
		reconnectDelay: reconnectDelay,
	}
	c.maxRetries = 12 /// with default 5 second delay this would be a minute
	return c
}

// LockOpts returns pointer to underlying Lock object and an error.
func (c *consulClientImpl) LockOpts(opts *api.LockOptions) (*api.Lock, error) {
	return c.conn.client.LockOpts(opts)
}

// reconnect reconnectes to any online and healthy consul server..
func (c *consulClientImpl) reconnect(conn *consulConnection) error {
	var err error

	// once.Do executes func() only once across concurrently executing threads
	conn.once.Do(func() {
		var config *api.Config
		var client *api.Client

		for _, machine := range c.connParams.machines {
			if strings.HasPrefix(machine, "http://") {
				machine = strings.TrimPrefix(machine, "http://")
			} else if strings.HasPrefix(machine, "https://") {
				machine = strings.TrimPrefix(machine, "https://")
			}

			// sleep for requested delay before testing new connection
			time.Sleep(c.reconnectDelay)
			if config, client, err = newKvClient(machine, c.connParams); err == nil {
				c.conn = &consulConnection{
					client: client,
					config: config,
					once:   new(sync.Once),
				}
				logrus.Infof("%s: %s\n", "successfully connected to", machine)
				break
			} else {
				logrus.Errorf("failed to reconnect client on: %s", machine)
			}
		}
	})

	if err != nil {
		logrus.Infof("Failed to reconnect client: %v", err)
	}
	return err
}

// isConsulErrNeedingRetry is a type of consul error on which we should try reconnecting consul client.
func isConsulErrNeedingRetry(err error) bool {
	return strings.Contains(err.Error(), httpError) ||
		strings.Contains(err.Error(), eofError) ||
		strings.Contains(err.Error(), connRefused) ||
		strings.Contains(err.Error(), nameResolutionError) ||
		strings.Contains(err.Error(), connReset)
}

// isKeyIndexMismatchErr returns true if error contains key index mismatch substring
func isKeyIndexMismatchErr(err error) bool {
	return strings.Contains(err.Error(), keyIndexMismatch)
}

// newKvClient constructs new kvdb.Kvdb given a single end-point to connect to.
func newKvClient(machine string, p connectionParams) (*api.Config, *api.Client, error) {
	config := api.DefaultConfig()
	config.HttpClient = http.DefaultClient
	config.Address = machine
	config.Scheme = "http"
	config.Token = p.options[kvdb.ACLTokenKey]

	// check if TLS is required
	if p.options[kvdb.TransportScheme] == "https" {
		tlsConfig := &api.TLSConfig{
			CAFile:             p.options[kvdb.CAFileKey],
			CertFile:           p.options[kvdb.CertFileKey],
			KeyFile:            p.options[kvdb.CertKeyFileKey],
			Address:            p.options[kvdb.CAAuthAddress],
			InsecureSkipVerify: strings.ToLower(p.options[kvdb.InsecureSkipVerify]) == "true",
		}

		consulTLSConfig, err := api.SetupTLSConfig(tlsConfig)
		if err != nil {
			logrus.Fatal(err)
		}

		config.Scheme = p.options[kvdb.TransportScheme]
		config.HttpClient = new(http.Client)
		config.HttpClient.Transport = &http.Transport{
			TLSClientConfig: consulTLSConfig,
		}
	}

	client, err := api.NewClient(config)
	if err != nil {
		logrus.Info("consul: failed to get new api client: %v", err)
		return nil, nil, err
	}

	// check health to ensure communication with consul are working
	if _, _, err := client.Health().State(api.HealthAny, nil); err != nil {
		logrus.Errorf("consul: health check failed for %v : %v", machine, err)
		return nil, nil, err
	}

	return config, client, nil
}

// consulFunc runs a consulFunc operation and returns true if needs to be retried
type consulFunc func() bool

// runWithRetry runs consulFunc with retries if required
func (c *consulClientImpl) runWithRetry(f consulFunc) {
	for i := 0; i < c.maxRetries; i++ {
		if !f() {
			break
		}
	}
}

// writeFunc defines an update operation for consul with this signature
type writeFunc func(conn *consulConnection) (*api.WriteMeta, error)

// writeRetryFunc runs writeFunc with retries if required
func (c *consulClientImpl) writeRetryFunc(f writeFunc) (*api.WriteMeta, error) {
	var err error
	var meta *api.WriteMeta
	retry := false
	c.runWithRetry(func() bool {
		conn := c.conn
		meta, err = f(conn)
		retry, _ = c.reconnectIfConnectionError(conn, err)
		return retry
	})
	return meta, err
}

/// reconnectIfConnectionError returns (retry, error), retry is true is client reconnected
func (c *consulClientImpl) reconnectIfConnectionError(conn *consulConnection, err error) (bool, error) {
	if err == nil {
		return false, nil
	} else if isConsulErrNeedingRetry(err) {
		logrus.Errorf("consul connection error: %v, trying to reconnect..", err)
		if clientErr := c.reconnect(conn); clientErr != nil {
			return false, clientErr
		} else {
			logrus.Infof("consul connection success, returning true")
			return true, nil
		}
	} else {
		return false, err
	}
}

func (c *consulClientImpl) Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error) {
	var pair *api.KVPair
	var meta *api.QueryMeta
	var err error
	retry := false

	c.runWithRetry(func() bool {
		conn := c.conn
		pair, meta, err = conn.client.KV().Get(key, q)
		retry, _ = c.reconnectIfConnectionError(conn, err)
		return retry
	})

	return pair, meta, err
}

func (c *consulClientImpl) Put(p *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error) {
	return c.writeRetryFunc(func(conn *consulConnection) (*api.WriteMeta, error) {
		return conn.client.KV().Put(p, q)
	})
}

func (c *consulClientImpl) Delete(key string, w *api.WriteOptions) (*api.WriteMeta, error) {
	return c.writeRetryFunc(func(conn *consulConnection) (*api.WriteMeta, error) {
		return conn.client.KV().Delete(key, w)
	})
}

func (c *consulClientImpl) DeleteTree(prefix string, w *api.WriteOptions) (*api.WriteMeta, error) {
	return c.writeRetryFunc(func(conn *consulConnection) (*api.WriteMeta, error) {
		return conn.client.KV().DeleteTree(prefix, w)
	})
}

func (c *consulClientImpl) Keys(prefix, separator string, q *api.QueryOptions) ([]string, *api.QueryMeta, error) {
	var list []string
	var meta *api.QueryMeta
	var err error
	retry := false

	c.runWithRetry(func() bool {
		conn := c.conn
		list, meta, err = conn.client.KV().Keys(prefix, separator, q)
		retry, _ = c.reconnectIfConnectionError(conn, err)
		return retry
	})

	return list, meta, err
}

func (c *consulClientImpl) List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error) {
	var pairs api.KVPairs
	var meta *api.QueryMeta
	var err error
	retry := false

	c.runWithRetry(func() bool {
		conn := c.conn
		pairs, meta, err = conn.client.KV().List(prefix, q)
		retry, _ = c.reconnectIfConnectionError(conn, err)
		return retry
	})

	return pairs, meta, err
}

func (c *consulClientImpl) Acquire(p *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error) {
	var err error
	var meta *api.WriteMeta
	var ok bool
	retry := false
	c.runWithRetry(func() bool {
		conn := c.conn
		ok, meta, err = conn.client.KV().Acquire(p, q)
		retry, _ = c.reconnectIfConnectionError(conn, err)
		return retry
	})

	// *** this error is created in loop above
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("acquire failed")
	}

	return meta, err
}

func (c *consulClientImpl) Create(se *api.SessionEntry, q *api.WriteOptions) (string, *api.WriteMeta, error) {
	var session string
	var meta *api.WriteMeta
	var err error
	retry := false

	c.runWithRetry(func() bool {
		conn := c.conn
		session, meta, err = conn.client.Session().Create(se, q)
		retry, _ = c.reconnectIfConnectionError(conn, err)
		return retry
	})

	return session, meta, err
}

func (c *consulClientImpl) Destroy(id string, q *api.WriteOptions) (*api.WriteMeta, error) {
	return c.writeRetryFunc(func(conn *consulConnection) (*api.WriteMeta, error) {
		return conn.client.Session().Destroy(id, q)
	})
}

func (c *consulClientImpl) Renew(id string, q *api.WriteOptions) (*api.SessionEntry, *api.WriteMeta, error) {
	var entry *api.SessionEntry
	var meta *api.WriteMeta
	var err error
	retry := false

	c.runWithRetry(func() bool {
		conn := c.conn
		entry, meta, err = conn.client.Session().Renew(id, q)
		retry, _ = c.reconnectIfConnectionError(conn, err)
		return retry
	})

	return entry, meta, err
}

func (c *consulClientImpl) RenewPeriodic(
	initialTTL string,
	id string,
	q *api.WriteOptions,
	doneCh chan struct{},
) error {
	var err error
	retry := false

	c.runWithRetry(func() bool {
		conn := c.conn
		err = conn.client.Session().RenewPeriodic(initialTTL, id, q, doneCh)
		retry, _ = c.reconnectIfConnectionError(conn, err)
		return retry
	})

	return err
}

func (c *consulClientImpl) CreateMeta(
	id string,
	p *api.KVPair,
	q *api.WriteOptions,
) (*api.WriteMeta, bool, error) {
	var ok bool
	var meta *api.WriteMeta
	var err error
	connError := false

	for i := 0; i < c.maxRetries; i++ {
		conn := c.conn
		ok, meta, err = conn.client.KV().Acquire(p, q)
		if ok && err == nil {
			return nil, ok, err
		}
		if _, err := conn.client.Session().Destroy(p.Session, nil); err != nil {
			logrus.Error(err)
		}
		if _, err := c.Delete(id, nil); err != nil {
			logrus.Error(err)
		}
		connError, err = c.reconnectIfConnectionError(conn, err)
		if connError {
			continue
		} else {
			break
		}
	}

	if !ok {
		return nil, ok, fmt.Errorf("failed to set ttl: %v", err)
	}

	return meta, ok, err
}

func (c *consulClientImpl) CompareAndSet(
	id string,
	value []byte,
	p *api.KVPair,
	q *api.WriteOptions,
) (bool, *api.WriteMeta, error) {
	var ok bool
	var meta *api.WriteMeta
	var err error
	retried := false
	connError := false

	for i := 0; i < c.maxRetries; i++ {
		conn := c.conn
		ok, meta, err = conn.client.KV().CAS(p, q)
		connError, err = c.reconnectIfConnectionError(conn, err)
		if connError {
			retried = true
			continue
		} else if err != nil && isKeyIndexMismatchErr(err) && retried {
			kvPair, _, getErr := conn.client.KV().Get(id, nil)
			if getErr != nil {
				// failed to get value from kvdb
				return false, nil, err
			}

			// Prev Value not equal to current value in consul
			if bytes.Compare(kvPair.Value, value) != 0 {
				return false, nil, err
			} else {
				// kvdb has the new value that we are trying to set
				err = nil
				break
			}
		} else {
			break
		}
	}

	return ok, meta, err
}

func (c *consulClientImpl) CompareAndDelete(
	id string,
	value []byte,
	p *api.KVPair,
	q *api.WriteOptions,
) (bool, *api.WriteMeta, error) {
	var ok bool
	var meta *api.WriteMeta
	var err error
	retried := false
	connError := false

	for i := 0; i < c.maxRetries; i++ {
		conn := c.conn
		ok, meta, err = conn.client.KV().DeleteCAS(p, q)
		connError, err = c.reconnectIfConnectionError(conn, err)
		if connError {
			retried = true
			continue
		} else if retried && err == kvdb.ErrNotFound {
			// assuming our delete went through, there is no way
			// to figure out who deleted it
			err = nil
			break
		} else {
			break
		}
	}

	return ok, meta, err
}
