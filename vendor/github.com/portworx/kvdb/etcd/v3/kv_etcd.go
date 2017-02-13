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

	"github.com/Sirupsen/logrus"

	"golang.org/x/net/context"

	e "github.com/coreos/etcd/clientv3"
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
	Name                  = "etcdv3-kv"
	defaultRequestTimeout = 10 * time.Second
)

var (
	defaultMachines = []string{"http://127.0.0.1:2379"}
)

func init() {
	if err := kvdb.Register(Name, New, ec.Version); err != nil {
		panic(err.Error())
	}
}

type etcdKV struct {
	common.BaseKvdb
	kvClient   *e.Client
	authClient e.Auth
	domain     string
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
		Endpoints:   machines,
		Username:    username,
		Password:    password,
		DialTimeout: ec.DefaultDialTimeout,
		TLS:         tlsCfg,
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
	return context.WithTimeout(context.Background(), defaultRequestTimeout)
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
			logrus.Errorf("kvdb deadline exceeded error: %v, retry count: %v\n", err, i)
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
			logrus.Errorf("kvdb deadline exceeded error: %v, retry count: %v\n", err, i)
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

func (et *etcdKV) Keys(prefix, key string) ([]string, error) {
	return nil, kvdb.ErrNotSupported
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

	ctx, cancel := et.Context()
	if (flags & kvdb.KVModifiedIndex) != 0 {
		txnResponse, txnErr = et.kvClient.Txn(ctx).
			If(e.Compare(e.ModRevision(key), "=", int64(kvp.ModifiedIndex))).
			Then(e.OpPut(key, string(kvp.Value), opts...)).
			Commit()
		cancel()
		if txnErr != nil {
			return nil, txnErr
		}
		if txnResponse.Succeeded == false {
			if len(txnResponse.Responses) == 0 {
				logrus.Infof("Etcd did not return any transaction responses")
			} else {
				for i, responseOp := range txnResponse.Responses {
					logrus.Infof("Etcd transaction Response: %v %v", i, responseOp.String())
				}
			}
			return nil, kvdb.ErrModified
		}

	} else {
		txnResponse, txnErr = et.kvClient.Txn(ctx).
			If(e.Compare(e.Value(key), "=", string(prevValue))).
			Then(e.OpPut(key, string(kvp.Value), opts...)).
			Commit()
		cancel()
		if txnErr != nil {
			return nil, txnErr
		}
		if txnResponse.Succeeded == false {
			if len(txnResponse.Responses) == 0 {
				logrus.Infof("Etcd did not return any transaction responses")
			} else {
				for i, responseOp := range txnResponse.Responses {
					logrus.Infof("Etcd transaction Response: %v %v", i, responseOp.String())
				}
			}
			return nil, kvdb.ErrValueMismatch
		}
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
				logrus.Infof("Etcd did not return any transaction responses")
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
				logrus.Infof("Etcd did not return any transaction responses")
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
	key = et.domain + key
	duration := time.Second
	ttl := uint64(ec.DefaultLockTTL)
	count := 0
	lockTag := ec.LockerIDInfo{LockerID: lockerID}
	kvPair, err := et.Create(key, lockTag, ttl)

	for maxCount := 300; err != nil && count < maxCount; count++ {
		time.Sleep(duration)
		kvPair, err = et.Create(key, lockTag, ttl)
		if count > 0 && count%15 == 0 && err != nil {
			currLockerTag := ec.LockerIDInfo{LockerID: ""}
			if _, errGet := et.GetVal(key, &currLockerTag); errGet == nil {
				logrus.Warnf("Lock %v locked for %v seconds, tag: %v",
					key, count, currLockerTag)
			}
		}
	}
	if err != nil {
		return nil, err
	}
	if count >= 10 {
		logrus.Warnf("ETCD: spent %v iterations locking %v\n", count, key)
	}
	kvPair.TTL = int64(time.Duration(ttl) * time.Second)
	kvPair.Lock = &ec.EtcdLock{Done: make(chan struct{})}
	go et.refreshLock(kvPair)
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
			result, err := et.kvClient.Put(ctx, pathKey, value)
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
			logrus.Errorf("kvdb deadline exceeded error: %v, retry count: %v\n", err, i)
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

func (et *etcdKV) refreshLock(kvPair *kvdb.KVPair) {
	l := kvPair.Lock.(*ec.EtcdLock)
	ttl := kvPair.TTL
	refresh := time.NewTicker(ec.DefaultLockRefreshDuration)
	var (
		keyString      string
		currentRefresh time.Time
		prevRefresh    time.Time
	)
	if kvPair != nil {
		keyString = kvPair.Key
	}
	defer refresh.Stop()
	for {
		select {
		case <-refresh.C:
			l.Lock()
			for !l.Unlocked {
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
		opts = append(opts, e.WithRev(int64(waitIndex)))
	}
	watcher := e.NewWatcher(et.kvClient)
	watchChan := watcher.Watch(context.Background(), key, opts...)
	for wresp := range watchChan {
		if wresp.Created == true {
			continue
		}
		if wresp.Canceled == true {
			// Watch is canceled. Notify the watcher
			logrus.Errorf("Watch on key %v cancelled. Error: %v", key, wresp.Err())
			_ = cb(key, opaque, nil, kvdb.ErrWatchStopped)
		} else {
			for _, ev := range wresp.Events {
				if waitIndex != 0 {
					if ev.Kv.ModRevision < int64(waitIndex) {
						// Skip this updte
						continue
					}
				}
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
				err := cb(key, opaque, et.resultToKv(ev.Kv, action), nil)
				if err != nil {
					closeErr := watcher.Close()
					// etcd server might close the context before us.
					if closeErr != context.Canceled && closeErr != nil {
						logrus.Errorf("Unable to close the watcher channel for key %v : %v", key, closeErr)
					}
					// Indicate the caller that watch has been canceled
					_ = cb(key, opaque, nil, kvdb.ErrWatchStopped)
					break
				}
			}
		}
	}
}

func (et *etcdKV) Snapshot(prefix string) (kvdb.Kvdb, uint64, error) {
	// Create a new bootstrap key

	r := rand.New(rand.NewSource(time.Now().UnixNano())).Int63()
	bootStrapKeyLow := ec.Bootstrap + strconv.FormatInt(r, 10) +
		strconv.FormatInt(time.Now().UnixNano(), 10)
	kvPair, err := et.Put(bootStrapKeyLow, time.Now().UnixNano(), 0)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create snap bootstrap key %v, "+
			"err: %v", bootStrapKeyLow, err)
	}
	lowestKvdbIndex := kvPair.ModifiedIndex

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
			if len(newKvPairs) > 0 {
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
				if err == kvdb.ErrWatchStopped {
					return nil
				}
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

			if kvp.ModifiedIndex >= highestKvdbIndex {
				// Done applying changes.
				watchErr = fmt.Errorf("done")
				sendErr = nil
				goto errordone
			}

			_, err = snapDb.SnapPut(kvp)
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

		if err := et.WatchTree("", lowestKvdbIndex, mutex,
			cb); err != nil {
			return nil, 0, fmt.Errorf("Failed to start watch: %v", err)
		}
		err = <-done
		if err != nil {
			return nil, 0, err
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
