package etcdv3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"golang.org/x/net/context"

	e "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/etcdserver"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/common"
	ec "github.com/portworx/kvdb/etcd/common"
	"github.com/portworx/kvdb/mem"
)

const (
	// Name is the name of this kvdb implementation.
	Name                      = "etcdv3-kv"
	defaultKvRequestTimeout   = 10 * time.Second
	defaultMaintenanceTimeout = 5 * time.Second
	// defaultSessionTimeout in seconds is used for etcd watch
	// to detect connectivity issues
	defaultSessionTimeout = 120
	// All the below timeouts are similar to the ones set in etcdctl
	// and are mainly used for etcd client's load balancing.
	defaultDialTimeout      = 2 * time.Second
	defaultKeepAliveTime    = 2 * time.Second
	defaultKeepAliveTimeout = 6 * time.Second
	urlPrefix               = "http://"
	// timeoutMaxRetry is maximum retries before faulting
	timeoutMaxRetry = 30
)

var (
	defaultMachines = []string{"http://127.0.0.1:2379"}
)

// watchQ to collect updates without blocking
type watchQ struct {
	// q is the producer consumer q
	q common.WatchUpdateQueue
	// opaque is returned with the callbacl
	opaque interface{}
	// cb is the watch callback
	cb kvdb.WatchCB
	// watchRet returns error on channel to indicate stopping of watch
	watchRet chan error
	// done is true if watch has finished and no longer active
	done bool
}

func newWatchQ(o interface{}, cb kvdb.WatchCB, watchRet chan error) *watchQ {
	q := &watchQ{q: common.NewWatchUpdateQueue(), opaque: o, cb: cb,
		watchRet: watchRet, done: false}
	go q.start()
	return q
}

func (w *watchQ) enqueue(key string, kvp *kvdb.KVPair, err error) bool {
	w.q.Enqueue(key, kvp, err)
	return !w.done
}

func (w *watchQ) start() {
	for {
		key, kvp, err := w.q.Dequeue()
		err = w.cb(key, w.opaque, kvp, err)
		if err != nil {
			w.done = true
			logrus.Infof("Watch cb for key %v returned err: %v", key, err)
			// Indicate the caller that watch has been canceled
			_ = w.cb(key, w.opaque, nil, kvdb.ErrWatchStopped)
			// Indicate that watch is returning.
			close(w.watchRet)
			break
		}
	}
}

func init() {
	if err := kvdb.Register(Name, New, ec.Version); err != nil {
		panic(err.Error())
	}
}

type etcdKV struct {
	common.BaseKvdb
	kvClient          *e.Client
	authClient        e.Auth
	maintenanceClient e.Maintenance
	domain            string
	ec.EtcdCommon
}

// New constructs a new kvdb.Kvdb.
func New(
	domain string,
	machines []string,
	options map[string]string,
	fatalErrorCb kvdb.FatalErrorCB,
) (kvdb.Kvdb, error) {
	if len(machines) == 0 {
		machines = defaultMachines
	}

	etcdCommon := ec.NewEtcdCommon(options)
	tls, username, password, err := etcdCommon.GetAuthInfoFromOptions()
	if err != nil {
		return nil, err
	}

	tlsCfg, err := tls.ClientConfig()
	if err != nil {
		return nil, err
	}

	cfg := e.Config{
		Endpoints:            machines,
		Username:             username,
		Password:             password,
		DialTimeout:          defaultDialTimeout,
		TLS:                  tlsCfg,
		DialKeepAliveTime:    defaultKeepAliveTime,
		DialKeepAliveTimeout: defaultKeepAliveTimeout,

		// The time required for a request to fail - 30 sec
		//HeaderTimeoutPerRequest: time.Duration(10) * time.Second,
	}
	c, err := e.New(cfg)
	if err != nil {
		return nil, err
	}
	if domain != "" && !strings.HasSuffix(domain, "/") {
		domain = domain + "/"
	}
	return &etcdKV{
		common.BaseKvdb{FatalCb: fatalErrorCb},
		c,
		e.NewAuth(c),
		e.NewMaintenance(c),
		domain,
		etcdCommon,
	}, nil
}

func (et *etcdKV) String() string {
	return Name
}

func (et *etcdKV) Capabilities() int {
	return kvdb.KVCapabilityOrderedUpdates
}

func (et *etcdKV) Context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultKvRequestTimeout)
}

func (et *etcdKV) MaintenanceContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(getContextWithLeaderRequirement(), defaultMaintenanceTimeout)
}

func (et *etcdKV) Get(key string) (*kvdb.KVPair, error) {
	var (
		err    error
		result *e.GetResponse
	)
	key = et.domain + key
	for i := 0; i < et.GetRetryCount(); i++ {
		ctx, cancel := et.Context()
		result, err = et.kvClient.Get(ctx, key)
		cancel()
		if err == nil && result != nil {
			kvs := et.handleGetResponse(result, false)
			if len(kvs) == 0 {
				return nil, kvdb.ErrNotFound
			}
			return kvs[0], nil
		}

		switch err {
		case context.DeadlineExceeded:
			logrus.Errorf("[get %v]: kvdb deadline exceeded error: %v, retry count: %v\n", key, err, i)
			time.Sleep(ec.DefaultIntervalBetweenRetries)
		case etcdserver.ErrTimeout:
			logrus.Errorf("kvdb error: %v, retry count: %v \n", err, i)
			time.Sleep(ec.DefaultIntervalBetweenRetries)
		case etcdserver.ErrUnhealthy:
			logrus.Errorf("kvdb error: %v, retry count: %v \n", err, i)
			time.Sleep(ec.DefaultIntervalBetweenRetries)
		default:
			if err == rpctypes.ErrGRPCEmptyKey {
				return nil, kvdb.ErrNotFound
			}
			return nil, err
		}
	}
	return nil, err
}

func (et *etcdKV) GetVal(key string, val interface{}) (*kvdb.KVPair, error) {
	kvp, err := et.Get(key)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(kvp.Value, val); err != nil {
		return kvp, kvdb.ErrUnmarshal
	}
	return kvp, nil
}

func (et *etcdKV) Put(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
	b, err := common.ToBytes(val)
	if err != nil {
		return nil, err
	}
	return et.setWithRetry(key, string(b), ttl)
}

func (et *etcdKV) Create(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
	pathKey := et.domain + key
	opts := []e.OpOption{}
	if ttl > 0 {
		if ttl < 5 {
			return nil, kvdb.ErrTTLNotSupported
		}
		leaseCtx, leaseCancel := et.Context()
		leaseResult, err := et.kvClient.Grant(leaseCtx, int64(ttl))
		leaseCancel()
		if err != nil {
			return nil, err
		}
		opts = append(opts, e.WithLease(leaseResult.ID))

	}
	b, _ := common.ToBytes(val)
	ctx, cancel := et.Context()
	// Txn
	// If key exist before
	// Then do nothing (txnResponse.Succeeded == true)
	// Else put/create the key (txnResponse.Succeeded == false)
	txnResponse, txnErr := et.kvClient.Txn(ctx).If(
		e.Compare(e.CreateRevision(pathKey), ">", 0),
	).Then().Else(
		e.OpPut(pathKey, string(b), opts...),
		e.OpGet(pathKey),
	).Commit()
	cancel()
	if txnErr != nil {
		return nil, txnErr
	}
	if txnResponse.Succeeded == true {
		// The key did exist before
		return nil, kvdb.ErrExist
	}

	rangeResponse := txnResponse.Responses[1].GetResponseRange()
	kvPair := et.resultToKv(rangeResponse.Kvs[0], "create")
	return kvPair, nil
}

func (et *etcdKV) Update(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
	pathKey := et.domain + key
	opts := []e.OpOption{}
	if ttl > 0 {
		if ttl < 5 {
			return nil, kvdb.ErrTTLNotSupported
		}
		leaseCtx, leaseCancel := et.Context()
		leaseResult, err := et.kvClient.Grant(leaseCtx, int64(ttl))
		leaseCancel()
		if err != nil {
			return nil, err
		}
		opts = append(opts, e.WithLease(leaseResult.ID))

	}
	b, _ := common.ToBytes(val)
	ctx, cancel := et.Context()
	// Txn
	// If key exist before
	// Then update key (txnResponse.Succeeded == true)
	// Else put/create the key (txnResponse.Succeeded == false)
	txnResponse, txnErr := et.kvClient.Txn(ctx).If(
		e.Compare(e.CreateRevision(pathKey), ">", 0),
	).Then(
		e.OpPut(pathKey, string(b), opts...),
		e.OpGet(pathKey),
	).Else().Commit()
	cancel()
	if txnErr != nil {
		return nil, txnErr
	}
	if txnResponse.Succeeded == false {
		// The key did not exist before
		return nil, kvdb.ErrNotFound
	}

	rangeResponse := txnResponse.Responses[1].GetResponseRange()
	kvPair := et.resultToKv(rangeResponse.Kvs[0], "update")
	return kvPair, nil
}

func (et *etcdKV) Enumerate(prefix string) (kvdb.KVPairs, error) {
	prefix = et.domain + prefix
	var err error

	for i := 0; i < et.GetRetryCount(); i++ {
		ctx, cancel := et.Context()
		result, err := et.kvClient.Get(
			ctx,
			prefix,
			e.WithPrefix(),
			e.WithSort(e.SortByKey, e.SortAscend),
		)
		cancel()
		if err == nil && result != nil {
			kvs := et.handleGetResponse(result, true)
			return kvs, nil
		}

		switch err {
		case context.DeadlineExceeded:
			logrus.Errorf("[enumerate %v]: kvdb deadline exceeded error: %v, retry count: %v\n", prefix, err, i)
			time.Sleep(ec.DefaultIntervalBetweenRetries)
		case etcdserver.ErrTimeout:
			logrus.Errorf("kvdb error: %v, retry count: %v \n", err, i)
			time.Sleep(ec.DefaultIntervalBetweenRetries)
		case etcdserver.ErrUnhealthy:
			logrus.Errorf("kvdb error: %v, retry count: %v \n", err, i)
			time.Sleep(ec.DefaultIntervalBetweenRetries)
		default:
			if err == rpctypes.ErrGRPCEmptyKey {
				return nil, kvdb.ErrNotFound
			}
			return nil, err
		}
	}
	return nil, err
}

func (et *etcdKV) Delete(key string) (*kvdb.KVPair, error) {
	// Delete does not return the prev kv value even after setting
	// the WithPrevKV OpOption.
	kvp, err := et.Get(key)
	if err != nil {
		return nil, err
	}
	key = et.domain + key

	ctx, cancel := et.Context()
	result, err := et.kvClient.Delete(
		ctx,
		key,
		e.WithPrevKV(),
	)
	cancel()
	if err == nil {
		if result.Deleted != 1 {
			return nil, fmt.Errorf("Incorrect number of keys: %v deleted", key)
		}
		kvp.Action = kvdb.KVDelete
		return kvp, nil
	}

	if err == rpctypes.ErrGRPCEmptyKey {
		return nil, kvdb.ErrNotFound
	}

	return nil, err
}

func (et *etcdKV) DeleteTree(prefix string) error {
	prefix = et.domain + prefix

	ctx, cancel := et.Context()
	_, err := et.kvClient.Delete(
		ctx,
		prefix,
		e.WithPrevKV(),
		e.WithPrefix(),
	)
	cancel()
	return err
}

func (et *etcdKV) Keys(prefix, sep string) ([]string, error) {
	var (
		err    error
		result *e.GetResponse
	)
	if "" == sep {
		sep = "/"
	}
	lenPrefix := len(prefix)
	lenSep := len(sep)
	if lenPrefix > 0 && prefix[lenPrefix-lenSep:] != sep {
		prefix += sep
		lenPrefix += lenSep
	}
	retList := make([]string, 0, 10)
	nextKey := et.domain + prefix
	for i := 0; i < et.GetRetryCount(); i++ {
		for {
			ctx, cancel := et.Context()
			// looking for the first key immediately after "nextKey"
			result, err = et.kvClient.Get(ctx, nextKey, e.WithFromKey(), e.WithLimit(1), e.WithKeysOnly())
			if err == nil && result != nil {
				kvs := et.handleGetResponse(result, false)
				if len(kvs) == 0 {
					// no more keys, break out
					break
				}
				key := kvs[0].Key
				if lenPrefix > 0 {
					if strings.HasPrefix(key, prefix) {
						// strip prefix (if used)
						key = key[lenPrefix:]
					} else {
						// .. no longer our prefix, stop the scan
						break
					}
				}
				if idx := strings.Index(key, sep); idx > 0 {
					// extract key's first "directory"
					key = key[:idx]
				}
				retList = append(retList, key)
				// reset nextKey to "prefix/<last_found>~" (note: "~" is at the end of printable ascii(7))
				nextKey = et.domain + prefix + key + "~"
				continue
			}

			cancel()
			switch err {
			case context.DeadlineExceeded:
				logrus.Errorf("[%v.%v]: kvdb deadline exceeded error: %v, retry count: %v\n", prefix, sep, err, i)
				time.Sleep(ec.DefaultIntervalBetweenRetries)
			case etcdserver.ErrTimeout:
				logrus.Errorf("kvdb error: %v, retry count: %v \n", err, i)
				time.Sleep(ec.DefaultIntervalBetweenRetries)
			case etcdserver.ErrUnhealthy:
				logrus.Errorf("kvdb error: %v, retry count: %v \n", err, i)
				time.Sleep(ec.DefaultIntervalBetweenRetries)
			default:
				if err == rpctypes.ErrGRPCEmptyKey {
					break
				}
				return nil, err
			}
		}
		return retList, nil
	}
	return nil, err
}

func (et *etcdKV) CompareAndSet(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
	prevValue []byte,
) (*kvdb.KVPair, error) {
	var (
		leaseResult *e.LeaseGrantResponse
		txnResponse *e.TxnResponse
		txnErr, err error
	)
	key := et.domain + kvp.Key

	opts := []e.OpOption{}
	if (flags & kvdb.KVTTL) != 0 {
		leaseCtx, leaseCancel := et.Context()
		leaseResult, err = et.kvClient.Grant(leaseCtx, int64(kvp.TTL))
		leaseCancel()
		if err != nil {
			return nil, err
		}
		opts = append(opts, e.WithLease(leaseResult.ID))
	}

	cmp := e.Compare(e.Value(key), "=", string(prevValue))
	if (flags & kvdb.KVModifiedIndex) != 0 {
		cmp = e.Compare(e.ModRevision(key), "=", int64(kvp.ModifiedIndex))
	}
retry:
	for i := 0; i < timeoutMaxRetry; i++ {
		ctx, cancel := et.Context()
		txnResponse, txnErr = et.kvClient.Txn(ctx).
			If(cmp).
			Then(e.OpPut(key, string(kvp.Value), opts...)).
			Commit()
		cancel()
		if txnErr != nil {
			// Check if we need to retry
			switch txnErr {
			case context.DeadlineExceeded, etcdserver.ErrTimeout, etcdserver.ErrUnhealthy:
				logrus.Errorf("[cas %v]: kvdb error: %v, retry count: %v\n", key, txnErr, i)
				time.Sleep(ec.DefaultIntervalBetweenRetries)
			default:
				if err == rpctypes.ErrGRPCEmptyKey {
					return nil, kvdb.ErrNotFound
				} else if strings.Contains(txnErr.Error(), rpctypes.ErrGRPCTimeout.Error()) {
					logrus.Errorf("[cas: %v] kvdb grpc timeout: %v, retry count %v \n", key, txnErr, i)
					time.Sleep(ec.DefaultIntervalBetweenRetries)
				} else {
					// For all other errors return immediately
					return nil, txnErr
				}
			}

			// server timeout
			kvPair, err := et.Get(kvp.Key)
			if err != nil {
				return nil, txnErr
			}
			if kvPair.ModifiedIndex == kvp.ModifiedIndex {
				// update did not succeed, retry
				if i == (timeoutMaxRetry - 1) {
					et.FatalCb("Too many server retries for CAS: %v", *kvp)
				}
				continue retry
			} else if bytes.Compare(kvp.Value, kvPair.Value) == 0 {
				return kvPair, nil
			}
			// else someone else updated the value, return error
			return nil, txnErr
		}
		if txnResponse.Succeeded == false {
			if len(txnResponse.Responses) == 0 {
				logrus.Infof("Etcd did not return any transaction responses "+
					"for key (%v)", kvp.Key)
			} else {
				for i, responseOp := range txnResponse.Responses {
					logrus.Infof("Etcd transaction Response: %v %v", i,
						responseOp.String())
				}
			}
			if (flags & kvdb.KVModifiedIndex) != 0 {
				return nil, kvdb.ErrModified
			} else {
				return nil, kvdb.ErrValueMismatch
			}
		}
		break
	}

	kvPair, err := et.Get(kvp.Key)
	if err != nil {
		return nil, err
	}
	return kvPair, nil
}

func (et *etcdKV) CompareAndDelete(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
) (*kvdb.KVPair, error) {
	key := et.domain + kvp.Key
	ctx, cancel := et.Context()
	if (flags & kvdb.KVModifiedIndex) != 0 {
		txnResponse, txnErr := et.kvClient.Txn(ctx).
			If(e.Compare(e.ModRevision(key), "=", int64(kvp.ModifiedIndex))).
			Then(e.OpDelete(key)).
			Commit()
		cancel()
		if txnErr != nil {
			return nil, txnErr
		}
		if txnResponse.Succeeded == false {
			if len(txnResponse.Responses) == 0 {
				logrus.Infof("Etcd did not return any transaction responses for key (%v)", kvp.Key)
			} else {
				for i, responseOp := range txnResponse.Responses {
					logrus.Infof("Etcd transaction Response: %v %v", i, responseOp.String())
				}
			}
			return nil, kvdb.ErrModified
		}
	} else {
		txnResponse, txnErr := et.kvClient.Txn(ctx).
			If(e.Compare(e.Value(key), "=", string(kvp.Value))).
			Then(e.OpDelete(key)).
			Commit()
		cancel()
		if txnErr != nil {
			return nil, txnErr
		}
		if txnResponse.Succeeded == false {
			if len(txnResponse.Responses) == 0 {
				logrus.Infof("Etcd did not return any transaction responses for key (%v)", kvp.Key)
			} else {
				for i, responseOp := range txnResponse.Responses {
					logrus.Infof("Etcd transaction Response: %v %v", i, responseOp.String())
				}
			}
			return nil, kvdb.ErrValueMismatch
		}
	}
	return kvp, nil
}

func (et *etcdKV) WatchKey(
	key string,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) error {
	key = et.domain + key
	go et.watchStart(key, false, waitIndex, opaque, cb)
	return nil
}

func (et *etcdKV) WatchTree(
	prefix string,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) error {
	prefix = et.domain + prefix
	go et.watchStart(prefix, true, waitIndex, opaque, cb)
	return nil
}

func (et *etcdKV) Lock(key string) (*kvdb.KVPair, error) {
	return et.LockWithID(key, "locked")
}

func (et *etcdKV) LockWithID(key string, lockerID string) (
	*kvdb.KVPair,
	error,
) {
	return et.LockWithTimeout(key, lockerID, kvdb.DefaultLockTryDuration, et.GetLockTimeout())
}

func (et *etcdKV) LockWithTimeout(
	key string,
	lockerID string,
	lockTryDuration time.Duration,
	lockHoldDuration time.Duration,
) (*kvdb.KVPair, error) {
	key = et.domain + key
	duration := time.Second
	ttl := uint64(ec.DefaultLockTTL)
	lockTag := ec.LockerIDInfo{LockerID: lockerID}
	kvPair, err := et.Create(key, lockTag, ttl)
	startTime := time.Now()
	for count := 0; err != nil; count++ {
		time.Sleep(duration)
		kvPair, err = et.Create(key, lockTag, ttl)
		if count > 0 && count%15 == 0 && err != nil {
			currLockerTag := ec.LockerIDInfo{LockerID: ""}
			if _, errGet := et.GetVal(key, &currLockerTag); errGet == nil {
				logrus.Warnf("Lock %v locked for %v seconds, tag: %v, err: %v",
					key, count, currLockerTag, err)
			}
		}
		if err != nil && time.Since(startTime) > lockTryDuration {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	kvPair.TTL = int64(ttl)
	kvPair.Lock = &ec.EtcdLock{Done: make(chan struct{})}
	go et.refreshLock(kvPair, lockerID, lockHoldDuration)
	return kvPair, err
}

func (et *etcdKV) Unlock(kvp *kvdb.KVPair) error {
	l, ok := kvp.Lock.(*ec.EtcdLock)
	if !ok {
		return fmt.Errorf("Invalid lock structure for key %v", string(kvp.Key))
	}
	l.Lock()
	// Don't modify kvp here, CompareAndDelete does that.
	_, err := et.CompareAndDelete(kvp, kvdb.KVFlags(0))
	if err == nil {
		l.Unlocked = true
		l.Unlock()
		l.Done <- struct{}{}
		return nil
	}
	l.Unlock()
	return err
}

func (et *etcdKV) TxNew() (kvdb.Tx, error) {
	return nil, kvdb.ErrNotSupported
}

func (et *etcdKV) getAction(action string) kvdb.KVAction {
	switch action {

	case "create":
		return kvdb.KVCreate
	case "set", "update", "compareAndSwap":
		return kvdb.KVSet
	case "delete", "compareAndDelete":
		return kvdb.KVDelete
	case "get":
		return kvdb.KVGet
	default:
		return kvdb.KVUknown
	}
}

func (et *etcdKV) resultToKv(resultKv *mvccpb.KeyValue, action string) *kvdb.KVPair {
	kvp := &kvdb.KVPair{
		Value:         resultKv.Value,
		ModifiedIndex: uint64(resultKv.ModRevision),
		CreatedIndex:  uint64(resultKv.ModRevision),
	}

	kvp.Action = et.getAction(action)
	key := string(resultKv.Key[:])
	kvp.Key = strings.TrimPrefix(key, et.domain)
	return kvp
}

func isHidden(key string) bool {
	tokens := strings.Split(key, "/")
	keySuffix := tokens[len(tokens)-1]
	return keySuffix != "" && keySuffix[0] == '_'
}

func (et *etcdKV) handleGetResponse(result *e.GetResponse, removeHidden bool) kvdb.KVPairs {
	kvs := []*kvdb.KVPair{}
	for i := range result.Kvs {
		if removeHidden && isHidden(string(result.Kvs[i].Key[:])) {
			continue
		}
		kvs = append(kvs, et.resultToKv(result.Kvs[i], "get"))
	}
	return kvs
}

func (et *etcdKV) handlePutResponse(result *e.PutResponse, key string) (*kvdb.KVPair, error) {
	kvPair, err := et.Get(key)
	if err != nil {
		return nil, err
	}
	kvPair.Action = kvdb.KVSet
	return kvPair, nil
}

func (et *etcdKV) setWithRetry(key, value string, ttl uint64) (*kvdb.KVPair, error) {
	var (
		err    error
		i      int
		result *e.PutResponse
	)
	pathKey := et.domain + key
	if ttl > 0 && ttl < 5 {
		return nil, kvdb.ErrTTLNotSupported
	}
	for i = 0; i < et.GetRetryCount(); i++ {
		if ttl > 0 {
			var leaseResult *e.LeaseGrantResponse
			leaseCtx, leaseCancel := et.Context()
			leaseResult, err = et.kvClient.Grant(leaseCtx, int64(ttl))
			leaseCancel()
			if err != nil {
				goto handle_error
			}
			ctx, cancel := et.Context()
			result, err = et.kvClient.Put(ctx, pathKey, value, e.WithLease(leaseResult.ID))
			cancel()
			if err == nil && result != nil {
				kvp, err := et.handlePutResponse(result, key)
				if err != nil {
					return nil, err
				}
				kvp.TTL = int64(ttl)
				return kvp, nil
			}
			goto handle_error
		} else {
			ctx, cancel := et.Context()
			result, err = et.kvClient.Put(ctx, pathKey, value)
			cancel()
			if err == nil && result != nil {
				kvp, err := et.handlePutResponse(result, key)
				if err != nil {
					return nil, err
				}
				kvp.TTL = 0
				return kvp, nil

			}
			goto handle_error
		}
	handle_error:
		switch err {
		case context.DeadlineExceeded:
			logrus.Errorf("[set %v]: kvdb deadline exceeded error: %v, retry count: %v\n", key, err, i)
			time.Sleep(ec.DefaultIntervalBetweenRetries)
		case etcdserver.ErrTimeout:
			logrus.Errorf("kvdb error: %v, retry count: %v \n", err, i)
			time.Sleep(ec.DefaultIntervalBetweenRetries)
		case etcdserver.ErrUnhealthy:
			logrus.Errorf("kvdb error: %v, retry count: %v \n", err, i)
			time.Sleep(ec.DefaultIntervalBetweenRetries)
		default:
			goto out
		}
	}

out:
	outErr := err
	// It's possible that update succeeded but the re-update failed.
	// Check only if the original error was a cluster error.
	if i > 0 && i < et.GetRetryCount() && err != nil {
		kvp, err := et.Get(key)
		if err == nil && bytes.Equal(kvp.Value, []byte(value)) {
			return kvp, nil
		}
	}

	return nil, outErr
}

func (et *etcdKV) refreshLock(
	kvPair *kvdb.KVPair,
	tag string,
	lockHoldDuration time.Duration,
) {
	l := kvPair.Lock.(*ec.EtcdLock)
	ttl := kvPair.TTL
	refresh := time.NewTicker(ec.DefaultLockRefreshDuration)
	var (
		keyString      string
		currentRefresh time.Time
		prevRefresh    time.Time
		startTime      time.Time
	)
	if kvPair != nil {
		keyString = kvPair.Key
	}
	startTime = time.Now()
	lockMsgString := keyString + ",tag=" + tag
	defer refresh.Stop()
	for {
		select {
		case <-refresh.C:
			l.Lock()
			for !l.Unlocked {
				et.CheckLockTimeout(lockMsgString, startTime, lockHoldDuration)
				kvPair.TTL = ttl
				kvp, err := et.CompareAndSet(
					kvPair,
					kvdb.KVTTL|kvdb.KVModifiedIndex,
					kvPair.Value,
				)
				currentRefresh = time.Now()
				if err != nil {
					et.FatalCb(
						"Error refreshing lock. [Key %v] [Err: %v]"+
							" [Current Refresh: %v] [Previous Refresh: %v]",
						keyString, err, currentRefresh, prevRefresh,
					)
					l.Err = err
					l.Unlock()
					return
				}
				prevRefresh = currentRefresh
				kvPair.ModifiedIndex = kvp.ModifiedIndex
				break
			}
			l.Unlock()
		case <-l.Done:
			return
		}
	}
}

func (et *etcdKV) watchStart(
	key string,
	recursive bool,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) {
	opts := []e.OpOption{}
	opts = append(opts, e.WithCreatedNotify())
	if recursive {
		opts = append(opts, e.WithPrefix())
	}
	if waitIndex != 0 {
		opts = append(opts, e.WithRev(int64(waitIndex+1)))
	}
	sessionChan := make(chan int, 1)
	var (
		session *concurrency.Session
		err     error
	)
	go func() {
		session, err = concurrency.NewSession(
			et.kvClient,
			concurrency.WithTTL(defaultSessionTimeout))
		close(sessionChan)
	}()

	select {
	case <-sessionChan:
		if err != nil {
			logrus.Errorf("Failed to establish session for etcd client watch: %v", err)
			_ = cb(key, opaque, nil, kvdb.ErrWatchStopped)
			return
		}
	case <-time.After(defaultKvRequestTimeout):
		logrus.Errorf("Failed to establish session for etcd client watch." +
			" Timeout!. Etcd cluster not reachable")
		_ = cb(key, opaque, nil, kvdb.ErrWatchStopped)
		return
	}
	ctx, watchCancel := context.WithCancel(context.Background())
	watchRet := make(chan error)
	watchChan := et.kvClient.Watch(ctx, key, opts...)
	watchQ := newWatchQ(opaque, cb, watchRet)
	go func() {
		for wresp := range watchChan {
			if wresp.Created == true {
				continue
			}
			if wresp.Canceled == true {
				// Watch is canceled. Notify the watcher
				logrus.Errorf("Watch on key %v cancelled. Error: %v", key,
					wresp.Err())
				watchQ.enqueue(key, nil, kvdb.ErrWatchStopped)
				break
			} else {
				for _, ev := range wresp.Events {
					var action string
					if ev.Type == mvccpb.PUT {
						if ev.Kv.Version == 1 {
							action = "create"
						} else {
							action = "set"
						}
					} else if ev.Type == mvccpb.DELETE {
						action = "delete"
					} else {
						action = "unknown"
					}
					if !watchQ.enqueue(key, et.resultToKv(ev.Kv, action), err) {
						return
					}
				}
			}
		}
	}()

	select {
	case <-session.Done(): // closed by etcd
		// Indicate the caller that watch has been canceled
		logrus.Errorf("Watch closing session for key: %v", key)
		watchQ.enqueue(key, nil, kvdb.ErrWatchStopped)
		watchCancel()
	case <-watchRet: // error in watcher
		// Close the context
		watchCancel()
		session.Close()
		logrus.Errorf("Watch for %v stopped", key)
		return
	}
}

func (et *etcdKV) Snapshot(prefix string) (kvdb.Kvdb, uint64, error) {
	// Create a new bootstrap key
	var updates []*kvdb.KVPair
	finalPutDone := false
	var lowestKvdbIndex, highestKvdbIndex uint64
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
			if err == kvdb.ErrWatchStopped && finalPutDone {
				return nil
			}
			logrus.Errorf("Watch returned error: %v", err)
			watchErr = err
			sendErr = err
			goto errordone
		}

		if kvp == nil {
			logrus.Infof("Snapshot error, nil kvp")
			watchErr = fmt.Errorf("kvp is nil")
			sendErr = watchErr
			goto errordone
		}

		m, ok = opaque.(*sync.Mutex)
		if !ok {
			logrus.Infof("Snapshot error, failed to get mutex")
			watchErr = fmt.Errorf("Failed to get mutex")
			sendErr = watchErr
			goto errordone
		}

		m.Lock()
		defer m.Unlock()
		updates = append(updates, kvp)
		if finalPutDone {
			if kvp.ModifiedIndex >= highestKvdbIndex {
				// Done applying changes.
				logrus.Infof("Snapshot complete")
				watchErr = fmt.Errorf("done")
				sendErr = nil
				goto errordone
			}
		}

		return nil
	errordone:
		done <- sendErr
		return watchErr
	}

	if err := et.WatchTree("", 0, mutex, cb); err != nil {
		return nil, 0, fmt.Errorf("Failed to start watch: %v", err)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano())).Int63()
	bootStrapKeyLow := ec.Bootstrap + strconv.FormatInt(r, 10) +
		strconv.FormatInt(time.Now().UnixNano(), 10)
	kvPair, err := et.Put(bootStrapKeyLow, time.Now().UnixNano(), 0)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create snap bootstrap key %v, "+
			"err: %v", bootStrapKeyLow, err)
	}
	lowestKvdbIndex = kvPair.ModifiedIndex

	kvPairs, err := et.Enumerate(prefix)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to enumerate %v: err: %v", prefix,
			err)
	}
	snapDb, err := mem.New(
		et.domain,
		nil,
		map[string]string{mem.KvSnap: "true"},
		et.FatalCb,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create in-mem kv store: %v", err)
	}

	for i := 0; i < len(kvPairs); i++ {
		kvPair := kvPairs[i]
		if len(kvPair.Value) > 0 {
			// Only create a leaf node
			_, err := snapDb.SnapPut(kvPair)
			if err != nil {
				return nil, 0, fmt.Errorf("Failed creating snap: %v", err)
			}
		} else {
			newKvPairs, err := et.Enumerate(kvPair.Key)
			if err != nil {
				return nil, 0, fmt.Errorf("Failed to get child keys: %v", err)
			}
			if len(newKvPairs) == 0 {
				// empty value for this key
				_, err := snapDb.SnapPut(kvPair)
				if err != nil {
					return nil, 0, fmt.Errorf("Failed creating snap: %v", err)
				}
			} else if len(newKvPairs) == 1 {
				// empty value for this key
				_, err := snapDb.SnapPut(newKvPairs[0])
				if err != nil {
					return nil, 0, fmt.Errorf("Failed creating snap: %v", err)
				}
			} else {
				kvPairs = append(kvPairs, newKvPairs...)
			}
		}
	}

	// Create bootrap key : highest index
	bootStrapKeyHigh := ec.Bootstrap + strconv.FormatInt(r, 10) +
		strconv.FormatInt(time.Now().UnixNano(), 10)
	kvPair, err = et.Put(bootStrapKeyHigh, time.Now().UnixNano(), 0)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create snap bootstrap key %v, "+
			"err: %v", bootStrapKeyHigh, err)
	}
	highestKvdbIndex = kvPair.ModifiedIndex

	mutex.Lock()
	finalPutDone = true
	mutex.Unlock()

	// wait until watch finishes
	err = <-done
	if err != nil {
		return nil, 0, err
	}

	// apply all updates between lowest and highest kvdb index
	for _, kvPair := range updates {
		if kvPair.ModifiedIndex < highestKvdbIndex &&
			kvPair.ModifiedIndex > lowestKvdbIndex {
			if kvPair.Action == kvdb.KVDelete {
				_, err = snapDb.Delete(kvPair.Key)
				// A Delete key was issued between our first lowestKvdbIndex Put
				// and Enumerate APIs in this function
				if err == kvdb.ErrNotFound {
					err = nil
				}
			} else {
				_, err = snapDb.SnapPut(kvPair)
			}
			if err != nil {
				return nil, 0, fmt.Errorf("Failed to apply update to snap: %v", err)
			}

		}
	}

	_, err = et.Delete(bootStrapKeyLow)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to delete snap bootstrap key: %v, "+
			"err: %v", bootStrapKeyLow, err)
	}
	_, err = et.Delete(bootStrapKeyHigh)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to delete snap bootstrap key: %v, "+
			"err: %v", bootStrapKeyHigh, err)
	}

	return snapDb, highestKvdbIndex, nil
}

func (et *etcdKV) EnumerateWithSelect(
	prefix string,
	enumerateSelect kvdb.EnumerateSelect,
	copySelect kvdb.CopySelect,
) ([]interface{}, error) {
	return nil, kvdb.ErrNotSupported
}

func (et *etcdKV) GetWithCopy(
	key string,
	copySelect kvdb.CopySelect,
) (interface{}, error) {
	return nil, kvdb.ErrNotSupported
}

func (et *etcdKV) SnapPut(snapKvp *kvdb.KVPair) (*kvdb.KVPair, error) {
	return nil, kvdb.ErrNotSupported
}

func (et *etcdKV) AddUser(username string, password string) error {
	// Create a role for this user
	roleName := username
	_, err := et.authClient.RoleAdd(context.Background(), roleName)
	if err != nil {
		return err
	}
	// Create the user
	_, err = et.authClient.UserAdd(context.Background(), username, password)
	if err != nil {
		return err
	}
	// Assign role to user
	_, err = et.authClient.UserGrantRole(context.Background(), username, roleName)
	return err
}

func (et *etcdKV) RemoveUser(username string) error {
	// Revoke user from this role
	roleName := username
	_, err := et.authClient.UserRevokeRole(context.Background(), username, roleName)
	if err != nil {
		return err
	}
	// Remove the role defined for this user
	_, err = et.authClient.RoleDelete(context.Background(), roleName)
	if err != nil {
		return err
	}
	// Remove the user
	_, err = et.authClient.UserDelete(context.Background(), username)
	return err
}

func (et *etcdKV) GrantUserAccess(username string, permType kvdb.PermissionType, subtree string) error {
	var domain string
	if et.domain[0] == '/' {
		domain = et.domain
	} else {
		domain = "/" + et.domain
	}
	subtree = domain + subtree
	etcdPermType, err := getEtcdPermType(permType)
	if err != nil {
		return err
	}
	// A role for this user has already been created
	// Just assign the subtree to this role
	roleName := username
	_, err = et.authClient.RoleGrantPermission(context.Background(), roleName, subtree, "", e.PermissionType(etcdPermType))
	return err
}

func (et *etcdKV) RevokeUsersAccess(username string, permType kvdb.PermissionType, subtree string) error {
	var domain string
	if et.domain[0] == '/' {
		domain = et.domain
	} else {
		domain = "/" + et.domain
	}
	subtree = domain + subtree
	roleName := username
	// A role for this user should ideally exist
	// Revoke the specfied permission for that subtree
	_, err := et.authClient.RoleRevokePermission(context.Background(), roleName, subtree, "")
	return err
}

func (et *etcdKV) AddMember(
	nodeIP string,
	nodePeerPort string,
	nodeName string,
) (map[string][]string, error) {
	peerURLs := et.listenPeerUrls(nodeIP, nodePeerPort)
	ctx, cancel := et.MaintenanceContext()
	_, err := et.kvClient.MemberAdd(ctx, peerURLs)
	cancel()
	if err != nil {
		return nil, err
	}
	resp := make(map[string][]string)
	ctx, cancel = et.MaintenanceContext()
	memberListResponse, err := et.kvClient.MemberList(ctx)
	cancel()
	if err != nil {
		return nil, err
	}
	for _, member := range memberListResponse.Members {
		if member.Name == "" {
			// Newly added member
			resp[nodeName] = member.PeerURLs
		} else {
			resp[member.Name] = member.PeerURLs
		}
	}
	return resp, nil
}

func (et *etcdKV) RemoveMember(
	nodeID string,
) error {
	ctx, cancel := et.MaintenanceContext()
	memberListResponse, err := et.kvClient.MemberList(ctx)
	cancel()
	if err != nil {
		return err
	}
	var (
		newClientUrls []string
		memberID      uint64
	)

	for _, member := range memberListResponse.Members {
		if member.Name == nodeID {
			memberID = member.ID
		} else {
			for _, clientURL := range member.ClientURLs {
				newClientUrls = append(newClientUrls, clientURL)
			}
		}
	}
	et.kvClient.SetEndpoints(newClientUrls...)
	ctx, cancel = et.MaintenanceContext()
	_, err = et.kvClient.MemberRemove(ctx, memberID)
	cancel()
	if err != nil {
		return err
	}
	return nil
}

func (et *etcdKV) ListMembers() (map[string]*kvdb.MemberInfo, error) {
	ctx, cancel := et.MaintenanceContext()
	memberListResponse, err := et.kvClient.MemberList(ctx)
	cancel()
	if err != nil {
		return nil, err
	}
	resp := make(map[string]*kvdb.MemberInfo)
	for _, member := range memberListResponse.Members {
		var (
			leader     bool
			dbSize     int64
			isHealthy  bool
			clientURLs []string
		)
		// etcd versions < v3.2.15 will return empty ClientURLs if
		// the node is unhealthy. For versions >= v3.2.15 they populate
		// ClientURLs but return an error status
		if len(member.ClientURLs) != 0 {
			ctx, cancel = et.MaintenanceContext()
			endpointStatus, err := et.maintenanceClient.Status(
				ctx,
				member.ClientURLs[0],
			)
			cancel()
			if err == nil {
				if member.ID == endpointStatus.Leader {
					leader = true
				}
				dbSize = endpointStatus.DbSize
				isHealthy = true
				// Only set the urls if status is healthy
				clientURLs = member.ClientURLs
			}
		}
		resp[member.Name] = &kvdb.MemberInfo{
			PeerUrls:   member.PeerURLs,
			ClientUrls: clientURLs,
			Leader:     leader,
			DbSize:     dbSize,
			IsHealthy:  isHealthy,
		}
	}
	return resp, nil
}

func (et *etcdKV) Serialize() ([]byte, error) {

	kvps, err := et.Enumerate("")
	if err != nil {
		return nil, err
	}
	return et.SerializeAll(kvps)
}

func (et *etcdKV) Deserialize(b []byte) (kvdb.KVPairs, error) {
	return et.DeserializeAll(b)
}

func (et *etcdKV) SetEndpoints(endpoints []string) error {
	et.kvClient.SetEndpoints(endpoints...)
	return nil
}

func (et *etcdKV) GetEndpoints() []string {
	return et.kvClient.Endpoints()
}

func (et *etcdKV) listenPeerUrls(ip string, port string) []string {
	return []string{et.constructURL(ip, port)}
}

func (et *etcdKV) constructURL(ip string, port string) string {
	ip = strings.TrimPrefix(ip, urlPrefix)
	return urlPrefix + ip + ":" + port
}

func getContextWithLeaderRequirement() context.Context {
	return e.WithRequireLeader(context.Background())
}

func getEtcdPermType(permType kvdb.PermissionType) (e.PermissionType, error) {
	switch permType {
	case kvdb.ReadPermission:
		return e.PermissionType(e.PermRead), nil
	case kvdb.WritePermission:
		return e.PermissionType(e.PermWrite), nil
	case kvdb.ReadWritePermission:
		return e.PermissionType(e.PermReadWrite), nil
	default:
		return -1, kvdb.ErrUnknownPermission
	}
}
