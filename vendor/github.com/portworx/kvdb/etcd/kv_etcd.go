package etcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"

	"golang.org/x/net/context"

	e "github.com/coreos/etcd/client"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/common"
	"github.com/portworx/kvdb/mem"
)

const (
	// Name is the name of this kvdb implementation.
	Name = "etcd-kv"

	defaultRetryCount             = 60
	defaultIntervalBetweenRetries = time.Millisecond * 500
	bootstrap                     = "kvdb/bootstrap"
)

var (
	defaultMachines = []string{"http://127.0.0.1:4001"}
)

func init() {
	if err := kvdb.Register(Name, New); err != nil {
		panic(err.Error())
	}
}

type etcdKV struct {
	client e.KeysAPI
	domain string
}

type etcdLock struct {
	done     chan struct{}
	unlocked bool
	err      error
	sync.Mutex
}

// New constructs a new kvdb.Kvdb.
func New(
	domain string,
	machines []string,
	options map[string]string,
) (kvdb.Kvdb, error) {
	if len(machines) == 0 {
		machines = defaultMachines
	}
	cfg := e.Config{
		Endpoints: machines,
		Transport: e.DefaultTransport,
		// The time required for a request to fail - 30 sec
		HeaderTimeoutPerRequest: time.Duration(10) * time.Second,
	}
	c, err := e.New(cfg)
	if err != nil {
		return nil, err
	}
	if domain != "" && !strings.HasSuffix(domain, "/") {
		domain = domain + "/"
	}
	return &etcdKV{
		e.NewKeysAPI(c),
		domain,
	}, nil
}

func (kv *etcdKV) String() string {
	return Name
}

func (kv *etcdKV) Get(key string) (*kvdb.KVPair, error) {
	key = kv.domain + key
	return kv.get(key, false, false)
}

func (kv *etcdKV) GetVal(key string, val interface{}) (*kvdb.KVPair, error) {
	kvp, err := kv.Get(key)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(kvp.Value, val); err != nil {
		return kvp, kvdb.ErrUnmarshal
	}
	return kvp, nil
}

func (kv *etcdKV) Put(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {

	key = kv.domain + key
	b, err := common.ToBytes(val)
	if err != nil {
		return nil, err
	}
	return kv.setWithRetry(
		context.Background(),
		key,
		string(b),
		&e.SetOptions{
			TTL: time.Duration(ttl) * time.Second,
		})
}

func (kv *etcdKV) Create(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {

	key = kv.domain + key

	b, err := common.ToBytes(val)
	if err != nil {
		return nil, err
	}
	return kv.setWithRetry(context.Background(), key, string(b), &e.SetOptions{
		TTL:       time.Duration(ttl) * time.Second,
		PrevExist: e.PrevNoExist,
	})
}

func (kv *etcdKV) Update(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {

	key = kv.domain + key

	b, err := common.ToBytes(val)
	if err != nil {
		return nil, err
	}
	return kv.setWithRetry(context.Background(), key, string(b), &e.SetOptions{
		TTL:       time.Duration(ttl) * time.Second,
		PrevExist: e.PrevExist,
	})
}

func (kv *etcdKV) Enumerate(prefix string) (kvdb.KVPairs, error) {
	prefix = kv.domain + prefix
	var err error
	for i := 0; i < defaultRetryCount; i++ {
		result, err := kv.client.Get(context.Background(), prefix, &e.GetOptions{
			Recursive: true,
			Sort:      true,
		})
		if err == nil {
			return kv.resultToKvs(result), nil
		}
		switch err.(type) {
		case *e.ClusterError:
			logrus.Errorf("kvdb set error: %v, retry count: %v\n", err, i)
			time.Sleep(defaultIntervalBetweenRetries)
		default:
			return nil, err
		}
	}
	return nil, err
}

func (kv *etcdKV) Delete(key string) (*kvdb.KVPair, error) {
	key = kv.domain + key

	result, err := kv.client.Delete(context.Background(), key, &e.DeleteOptions{
		Recursive: false,
	})
	if err == nil {
		return kv.resultToKv(result), err
	}
	return nil, err
}

func (kv *etcdKV) DeleteTree(prefix string) error {
	prefix = kv.domain + prefix

	_, err := kv.client.Delete(context.Background(), prefix, &e.DeleteOptions{
		Recursive: true,
	})
	return err
}

func (kv *etcdKV) Keys(prefix, key string) ([]string, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *etcdKV) CompareAndSet(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
	prevValue []byte,
) (*kvdb.KVPair, error) {

	opts := &e.SetOptions{
		PrevValue: string(prevValue),
	}
	if (flags & kvdb.KVModifiedIndex) != 0 {
		opts.PrevIndex = kvp.ModifiedIndex
	}
	if (flags & kvdb.KVTTL) != 0 {
		opts.TTL = time.Duration(kvp.TTL)
	}
	result, err := kv.client.Set(
		context.Background(),
		kv.domain+kvp.Key,
		string(kvp.Value),
		opts,
	)
	if err != nil {
		return nil, err
	}
	return kv.resultToKv(result), err
}

func (kv *etcdKV) CompareAndDelete(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
) (*kvdb.KVPair, error) {

	result, err := kv.client.Delete(
		context.Background(),
		kv.domain+kvp.Key,
		&e.DeleteOptions{
			PrevValue: string(kvp.Value),
			PrevIndex: kvp.ModifiedIndex,
		})
	if err != nil {
		return nil, err
	}
	return kv.resultToKv(result), err
}

func (kv *etcdKV) WatchKey(
	key string,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) error {

	key = kv.domain + key
	go kv.watchStart(key, false, waitIndex, opaque, cb)
	return nil
}

func (kv *etcdKV) WatchTree(
	prefix string,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) error {

	prefix = kv.domain + prefix
	go kv.watchStart(prefix, true, waitIndex, opaque, cb)
	return nil
}

func (kv *etcdKV) Lock(key string, ttl uint64) (*kvdb.KVPair, error) {
	key = kv.domain + key
	duration := time.Second

	count := 0
	kvPair, err := kv.Create(key, "locked", ttl)
	for maxCount := 300; err != nil && count < maxCount; count++ {
		time.Sleep(duration)
		kvPair, err = kv.Create(key, "locked", ttl)
	}
	if err != nil {
		return nil, err
	}
	if count > 10 {
		logrus.Warnf("ETCD: spent %v iterations locking %v\n", count, key)
	}
	kvPair.TTL = int64(time.Duration(ttl) * time.Second)
	kvPair.Lock = &etcdLock{done: make(chan struct{})}
	go kv.refreshLock(kvPair)

	return kvPair, err
}

func (kv *etcdKV) Unlock(kvp *kvdb.KVPair) error {
	l, ok := kvp.Lock.(*etcdLock)
	if !ok {
		return fmt.Errorf("Invalid lock structure for key %v", string(kvp.Key))
	}
	l.Lock()
	// Don't modify kvp here, CompareAndDelete does that.
	_, err := kv.CompareAndDelete(kvp, kvdb.KVFlags(0))
	if err == nil {
		l.unlocked = true
		l.Unlock()
		l.done <- struct{}{}
		return nil
	}
	l.Unlock()
	return err
}

func (kv *etcdKV) TxNew() (kvdb.Tx, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *etcdKV) nodeToKv(node *e.Node) *kvdb.KVPair {
	kvp := &kvdb.KVPair{
		Value:         []byte(node.Value),
		TTL:           node.TTL,
		ModifiedIndex: node.ModifiedIndex,
		CreatedIndex:  node.CreatedIndex,
	}

	// Strip out the leading '/'
	if len(node.Key) != 0 {
		kvp.Key = node.Key[1:]
	} else {
		kvp.Key = node.Key
	}
	kvp.Key = strings.TrimPrefix(kvp.Key, kv.domain)
	return kvp
}

func (kv *etcdKV) resultToKv(result *e.Response) *kvdb.KVPair {
	kvp := kv.nodeToKv(result.Node)
	switch result.Action {
	case "create":
		kvp.Action = kvdb.KVCreate
	case "set", "update":
		kvp.Action = kvdb.KVSet
	case "delete":
		kvp.Action = kvdb.KVDelete
	case "get":
		kvp.Action = kvdb.KVGet
	default:
		kvp.Action = kvdb.KVUknown
	}
	kvp.KVDBIndex = result.Index
	return kvp
}

func (kv *etcdKV) resultToKvs(result *e.Response) kvdb.KVPairs {
	kvs := make([]*kvdb.KVPair, len(result.Node.Nodes))
	for i := range result.Node.Nodes {
		kvs[i] = kv.nodeToKv(result.Node.Nodes[i])
		kvs[i].KVDBIndex = result.Index
	}
	return kvs
}

func (kv *etcdKV) get(key string, recursive, sort bool) (*kvdb.KVPair, error) {
	var err error
	var result *e.Response
	for i := 0; i < defaultRetryCount; i++ {
		result, err = kv.client.Get(context.Background(), key, &e.GetOptions{
			Recursive: recursive,
			Sort:      sort,
		})
		if err == nil {
			return kv.resultToKv(result), nil
		}

		switch err.(type) {
		case *e.ClusterError:
			logrus.Errorf("kvdb get error: %v, retry count: %v\n", err, i)
			time.Sleep(defaultIntervalBetweenRetries)
		default:
			etcdErr := err.(e.Error)
			if etcdErr.Code == e.ErrorCodeKeyNotFound {
				return nil, kvdb.ErrNotFound
			}
			return nil, err
		}
	}
	return nil, err
}

func (kv *etcdKV) setWithRetry(ctx context.Context, key, value string,
	opts *e.SetOptions) (*kvdb.KVPair, error) {
	var (
		err    error
		i      int
		result *e.Response
	)
	for i = 0; i < defaultRetryCount; i++ {
		result, err = kv.client.Set(ctx, key, value, opts)
		if err == nil {
			return kv.resultToKv(result), nil
		}
		switch err.(type) {
		case *e.ClusterError:
			cerr := err.(*e.ClusterError)
			logrus.Errorf("kvdb set error: %v %v, retry count: %v\n",
				err, cerr.Detail(), i)
			time.Sleep(defaultIntervalBetweenRetries)
		default:
			goto out
		}
	}

out:
	// It's possible that update succeeded but the re-update failed.
	if i > 0 && i < defaultRetryCount && err != nil {
		kvp, err := kv.get(key, false, false)
		if err == nil && bytes.Equal(kvp.Value, []byte(value)) {
			if opts.PrevExist == e.PrevNoExist {
				kvp.Action = kvdb.KVCreate
			} else {
				kvp.Action = kvdb.KVSet
			}
			return kvp, nil
		}
	}

	return nil, err
}

func (kv *etcdKV) refreshLock(kvPair *kvdb.KVPair) {
	l := kvPair.Lock.(*etcdLock)
	ttl := kvPair.TTL
	refresh := time.NewTicker(time.Duration(kvPair.TTL) / 4)
	var keyString string
	if kvPair != nil {
		keyString = kvPair.Key
	}
	defer refresh.Stop()
	for {
		select {
		case <-refresh.C:
			l.Lock()
			for !l.unlocked {
				kvPair.TTL = ttl
				kvp, err := kv.CompareAndSet(
					kvPair,
					kvdb.KVTTL|kvdb.KVModifiedIndex,
					kvPair.Value,
				)
				if err != nil {
					logrus.Errorf(
						"Error refreshing lock for key %v: %v\n",
						keyString, err,
					)
					l.err = err
					l.Unlock()
					return
				}
				kvPair.ModifiedIndex = kvp.ModifiedIndex
				break
			}
			l.Unlock()
		case <-l.done:
			return
		}
	}
}

func (kv *etcdKV) watchStart(
	key string,
	recursive bool,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) {
	ctx, cancel := context.WithCancel(context.Background())
	watcher := kv.client.Watcher(key, &e.WatcherOptions{
		AfterIndex: waitIndex,
		Recursive:  recursive,
	})
	isCancelSent := false
	for {
		r, watchErr := watcher.Next(ctx)
		if watchErr != nil && !isCancelSent {
			e, ok := watchErr.(e.Error)
			if ok {
				logrus.Errorf("Etcd error code %d Message %s Cause %s Index %d\n",
					e.Code, e.Message, e.Cause, e.Index)
			} else {
				logrus.Errorf("Etcd returned an error : %s\n", watchErr.Error())
			}
			// TODO: handle error
			_ = cb(key, opaque, nil, watchErr)
		} else if watchErr == nil && !isCancelSent {
			err := cb(key, opaque, kv.resultToKv(r), nil)
			if err != nil {
				// Cancel the context
				cancel()
				isCancelSent = true
			}
		} else {
			// TODO: handle error
			_ = cb(key, opaque, nil, kvdb.ErrWatchStopped)
			break
		}
	}
}

func (kv *etcdKV) Snapshot(prefix string) (kvdb.Kvdb, uint64, error) {
	// Create a new bootstrap key
	r := rand.New(rand.NewSource(time.Now().UnixNano())).Int63()
	bootStrapKey := bootstrap + strconv.FormatInt(r, 10) +
		strconv.FormatInt(time.Now().UnixNano(), 10)
	kvPair, err := kv.Put(bootStrapKey, time.Now().UnixNano(), 0)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create snap bootstrap key %v, "+
			"err: %v", bootStrapKey, err)
	}
	lowestKvdbIndex := kvPair.ModifiedIndex

	kvPairs, err := kv.Enumerate(prefix)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to enumerate %v: err: %v", prefix,
			err)
	}
	snapDb, err := mem.New(kv.domain, nil, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create in-mem kv store: %v", err)
	}

	for i := 0; i < len(kvPairs); i++ {
		kvPair := kvPairs[i]
		if len(kvPair.Value) > 0 {
			// Only create a leaf node
			_, err := snapDb.Put(kvPair.Key, kvPair.Value, 0)
			if err != nil {
				return nil, 0, fmt.Errorf("Failed creating snap: %v", err)
			}
		} else {
			newKvPairs, err := kv.Enumerate(kvPair.Key)
			if err != nil {
				return nil, 0, fmt.Errorf("Failed to get child keys: %v", err)
			}
			if len(newKvPairs) > 0 {
				kvPairs = append(kvPairs, newKvPairs...)
			}
		}
	}

	kvPair, err = kv.Delete(bootStrapKey)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to delete snap bootstrap key: %v, "+
			"err: %v", bootStrapKey, err)
	}
	highestKvdbIndex := kvPair.ModifiedIndex
	if lowestKvdbIndex+1 != highestKvdbIndex {
		// create a watch to get all changes
		// between lowestKvdbIndex and highestKvdbIndex
		done := make(chan error)
		mutex := &sync.Mutex{}
		cb := func(
			prefix string,
			opaque interface{},
			kvp *kvdb.KVPair,
			err error,
		) error {
			var watchErr error
			var sendErr error
			var m *sync.Mutex
			ok := false

			if err != nil {
				watchErr = err
				sendErr = err
				goto errordone
			}

			if kvp == nil {
				watchErr = fmt.Errorf("kvp is nil")
				sendErr = watchErr
				goto errordone

			}

			m, ok = opaque.(*sync.Mutex)
			if !ok {
				watchErr = fmt.Errorf("Failed to get mutex")
				sendErr = watchErr
				goto errordone
			}

			m.Lock()
			defer m.Unlock()

			if kvp.ModifiedIndex > highestKvdbIndex {
				// done applying changes, return
				watchErr = fmt.Errorf("done")
				sendErr = nil
				goto errordone
			}

			_, err = snapDb.Put(kvp.Key, kvp.Value, 0)
			if err != nil {
				watchErr = fmt.Errorf("Failed to apply update to snap: %v", err)
				sendErr = watchErr
				goto errordone
			}

			return nil
		errordone:
			done <- sendErr
			return watchErr
		}

		if err := kv.WatchTree("", lowestKvdbIndex, mutex,
			cb); err != nil {
			return nil, 0, fmt.Errorf("Failed to start watch: %v", err)
		}
		err = <-done
		if err != nil {
			return nil, 0, err
		}
	}

	return snapDb, highestKvdbIndex, nil
}
