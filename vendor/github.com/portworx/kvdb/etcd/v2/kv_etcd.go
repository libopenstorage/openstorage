package etcdv2

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

	e "github.com/coreos/etcd/client"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/common"
	ec "github.com/portworx/kvdb/etcd/common"
	"github.com/portworx/kvdb/mem"
)

const (
	// Name is the name of this kvdb implementation.
	Name = "etcd-kv"
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
	client   e.KeysAPI
	authUser e.AuthUserAPI
	authRole e.AuthRoleAPI
	domain   string
	ec.EtcdCommon
	kvdb.Controller
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
	tr, err := transport.NewTransport(tls, ec.DefaultDialTimeout)
	if err != nil {
		return nil, err
	}
	cfg := e.Config{
		Endpoints: machines,
		Transport: tr,
		Username:  username,
		Password:  password,
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
		common.BaseKvdb{FatalCb: fatalErrorCb},
		e.NewKeysAPI(c),
		e.NewAuthUserAPI(c),
		e.NewAuthRoleAPI(c),
		domain,
		etcdCommon,
		kvdb.ControllerNotSupported,
	}, nil
}

func (kv *etcdKV) String() string {
	return Name
}

func (kv *etcdKV) Capabilities() int {
	return kvdb.KVCapabilityOrderedUpdates
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
	for i := 0; i < kv.GetRetryCount(); i++ {
		result, err := kv.client.Get(context.Background(), prefix, &e.GetOptions{
			Recursive: true,
			Sort:      true,
			Quorum:    true,
		})
		if err == nil {
			return kv.resultToKvs(result), nil
		}
		switch err.(type) {
		case *e.ClusterError:
			logrus.Errorf("kvdb set error: %v, retry count: %v\n", err, i)
			time.Sleep(ec.DefaultIntervalBetweenRetries)
		case e.Error:
			etcdErr := err.(e.Error)
			if etcdErr.Code == e.ErrorCodeKeyNotFound {
				// Return an empty array
				return kvdb.KVPairs{}, nil
			}
			return nil, err
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

	switch err.(type) {
	case e.Error:
		etcdErr := err.(e.Error)
		if etcdErr.Code == e.ErrorCodeKeyNotFound {
			return nil, kvdb.ErrNotFound
		}
		return nil, err
	default:
		return nil, err
	}
}

func (kv *etcdKV) DeleteTree(prefix string) error {
	prefix = kv.domain + prefix

	_, err := kv.client.Delete(context.Background(), prefix, &e.DeleteOptions{
		Recursive: true,
	})
	return err
}

func (kv *etcdKV) Keys(prefix, sep string) ([]string, error) {
	// etcd-v2 supports only '/' separator
	sep = "/"
	prefix = kv.domain + prefix
	lenPrefix := len(prefix)
	if prefix[0:1] != sep {
		prefix = sep + prefix
		lenPrefix++
	}
	if prefix[lenPrefix-1:] != sep {
		prefix += sep
		lenPrefix++
	}
	var err error
	for i := 0; i < kv.GetRetryCount(); i++ {
		result, err := kv.client.Get(context.Background(), prefix, &e.GetOptions{
			Recursive: false,
			Sort:      true,
			Quorum:    true,
		})
		if err == nil {
			num := len(result.Node.Nodes)
			var keys []string
			if result.Node.Dir && num > 0 {
				keys = make([]string, num)
				for j := range result.Node.Nodes {
					key := result.Node.Nodes[j].Key
					if strings.HasPrefix(key, prefix) {
						key = key[lenPrefix:]
					}
					keys[j] = key
				}
			}
			return keys, nil
		}
		switch err.(type) {
		case *e.ClusterError:
			logrus.Errorf("kvdb set error: %v, retry count: %v\n", err, i)
			time.Sleep(ec.DefaultIntervalBetweenRetries)
		case e.Error:
			etcdErr := err.(e.Error)
			if etcdErr.Code == e.ErrorCodeKeyNotFound {
				// Return an empty array
				return []string{}, nil
			}
			return nil, err
		default:
			return nil, err
		}
	}
	return nil, err
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

func (kv *etcdKV) Lock(key string) (*kvdb.KVPair, error) {
	return kv.LockWithID(key, "locked")
}

func (kv *etcdKV) LockWithID(key string, lockerID string) (
	*kvdb.KVPair,
	error,
) {
	key = kv.domain + key
	duration := time.Second
	ttl := uint64(ec.DefaultLockTTL)
	count := 0
	lock := &ec.EtcdLock{Done: make(chan struct{}), Tag: lockerID}
	lockTag := ec.LockerIDInfo{LockerID: fmt.Sprintf("%p:%s", lock, lockerID)}
	kvPair, err := kv.Create(key, lockTag, ttl)
	for maxCount := 300; err != nil && count < maxCount; count++ {
		time.Sleep(duration)
		kvPair, err = kv.Create(key, lockTag, ttl)
		if count > 0 && count%15 == 0 && err != nil {
			currLockerTag := ec.LockerIDInfo{LockerID: ""}
			if _, errGet := kv.GetVal(key, &currLockerTag); errGet == nil {
				logrus.Warnf("Lock %v locked for %v seconds, tag: %v",
					key, count, currLockerTag)
			}
		}
	}
	if err != nil {
		return nil, err
	}
	kvPair.TTL = int64(time.Duration(ttl) * time.Second)
	kvPair.Lock = lock
	go kv.refreshLock(kvPair, lockerID)
	if count >= 10 {
		logrus.Warnf("ETCD: spent %v iterations locking %v\n", count, key)
	}

	return kvPair, err
}

func (kv *etcdKV) Unlock(kvp *kvdb.KVPair) error {
	l, ok := kvp.Lock.(*ec.EtcdLock)
	if !ok {
		return fmt.Errorf("Invalid lock structure for key %v", string(kvp.Key))
	}
	l.Lock()
	// Don't modify kvp here, CompareAndDelete does that.
	_, err := kv.CompareAndDelete(kvp, kvdb.KVFlags(0))
	if err == nil {
		l.Unlocked = true
		l.Unlock()
		l.Done <- struct{}{}
		return nil
	}
	l.Unlock()
	logrus.Errorf("Unlock failed for key: %s, tag: %s, error: %s, addr: %p",
		kvp.Key, l.Tag, err.Error(), l)
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
	case "set", "update", "compareAndSwap":
		kvp.Action = kvdb.KVSet
	case "delete", "compareAndDelete":
		kvp.Action = kvdb.KVDelete
	case "get":
		kvp.Action = kvdb.KVGet
	case "expire":
		kvp.Action = kvdb.KVExpire
	default:
		logrus.Warnf("unhandled kvdb operation %q", result.Action)
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
	for i := 0; i < kv.GetRetryCount(); i++ {
		result, err = kv.client.Get(context.Background(), key, &e.GetOptions{
			Recursive: recursive,
			Sort:      sort,
			Quorum:    true,
		})
		if err == nil {
			return kv.resultToKv(result), nil
		}

		switch err.(type) {
		case *e.ClusterError:
			logrus.Errorf("kvdb get error: %v, retry count: %v\n", err, i)
			time.Sleep(ec.DefaultIntervalBetweenRetries)
		case e.Error:
			etcdErr := err.(e.Error)
			if etcdErr.Code == e.ErrorCodeKeyNotFound {
				return nil, kvdb.ErrNotFound
			}
			return nil, err
		default:
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
	for i = 0; i < kv.GetRetryCount(); i++ {
		result, err = kv.client.Set(ctx, key, value, opts)
		if err == nil {
			return kv.resultToKv(result), nil
		}
		switch err.(type) {
		case *e.ClusterError:
			cerr := err.(*e.ClusterError)
			logrus.Errorf("kvdb set error: %v %v, retry count: %v\n",
				err, cerr.Detail(), i)
			time.Sleep(ec.DefaultIntervalBetweenRetries)
		default:
			goto out
		}
	}

out:
	outErr := err
	if outErr != nil && strings.Contains(outErr.Error(), kvdb.ErrExist.Error()) {
		outErr = kvdb.ErrExist
	}
	// It's possible that update succeeded but the re-update failed.
	// Check only if the original error was a cluster error.
	if i > 0 && i < kv.GetRetryCount() && err != nil {
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

	return nil, outErr
}

func (kv *etcdKV) refreshLock(kvPair *kvdb.KVPair, tag string) {
	l := kvPair.Lock.(*ec.EtcdLock)
	ttl := kvPair.TTL
	refresh := time.NewTicker(ec.DefaultLockRefreshDuration)
	var (
		keyString      string
		currentRefresh time.Time
		prevRefresh    time.Time
		startTime      time.Time
	)
	startTime = time.Now()
	if kvPair != nil {
		keyString = kvPair.Key
	}
	lockMsgString := keyString + ",tag=" + tag
	defer refresh.Stop()
	for {
		select {
		case <-refresh.C:
			l.Lock()
			for !l.Unlocked {
				kv.CheckLockTimeout(lockMsgString, startTime)
				kvPair.TTL = ttl
				kvp, err := kv.CompareAndSet(
					kvPair,
					kvdb.KVTTL|kvdb.KVModifiedIndex,
					kvPair.Value,
				)
				currentRefresh = time.Now()
				if err != nil {
					kv.FatalCb(
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
				logrus.Errorf("Etcd error code: [%d] Message: [%s] Cause: [%s] Index: [%d]\n",
					e.Code, e.Message, e.Cause, e.Index)
			} else {
				logrus.Errorf("Etcd returned an error : %s\n", watchErr.Error())
			}
			_ = cb(key, opaque, nil, watchErr)
			cancel()
			isCancelSent = true
		} else if watchErr == nil && !isCancelSent {
			err := cb(key, opaque, kv.resultToKv(r), nil)
			if err != nil {
				// Cancel the context
				cancel()
				isCancelSent = true
			}
		} else {
			// Ignore return values since the watch is stopping
			_ = cb(key, opaque, nil, kvdb.ErrWatchStopped)
			break
		}
	}
}

func (kv *etcdKV) Snapshot(prefix string) (kvdb.Kvdb, uint64, error) {

	var updates []*kvdb.KVPair
	done := make(chan error)
	mutex := &sync.Mutex{}
	finalPutDone := false
	var lowestKvdbIndex, highestKvdbIndex uint64

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
		updates = append(updates, kvp)
		if finalPutDone {
			if kvp.ModifiedIndex >= highestKvdbIndex {
				// Done applying changes.
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

	if err := kv.WatchTree("", 0, mutex, cb); err != nil {
		return nil, 0, fmt.Errorf("Failed to start watch: %v", err)
	}

	// Create a new bootstrap key
	r := rand.New(rand.NewSource(time.Now().UnixNano())).Int63()
	bootStrapKey := ec.Bootstrap + strconv.FormatInt(r, 10) +
		strconv.FormatInt(time.Now().UnixNano(), 10)
	kvPair, err := kv.Put(bootStrapKey, time.Now().UnixNano(), 0)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create snap bootstrap key %v, "+
			"err: %v", bootStrapKey, err)
	}
	lowestKvdbIndex = kvPair.ModifiedIndex

	kvPairs, err := kv.Enumerate(prefix)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to enumerate %v: err: %v", prefix,
			err)
	}
	snapDb, err := mem.New(
		kv.domain,
		nil,
		map[string]string{mem.KvSnap: "true"},
		kv.FatalCb,
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
			newKvPairs, err := kv.Enumerate(kvPair.Key)
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

	kvPair, err = kv.Delete(bootStrapKey)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to delete snap bootstrap key: %v, "+
			"err: %v", bootStrapKey, err)
	}
	highestKvdbIndex = kvPair.ModifiedIndex

	mutex.Lock()
	finalPutDone = true
	mutex.Unlock()

	// wait until the watch finishes
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

	return snapDb, highestKvdbIndex, nil
}

func (kv *etcdKV) EnumerateWithSelect(
	prefix string,
	enumerateSelect kvdb.EnumerateSelect,
	copySelect kvdb.CopySelect,
) ([]interface{}, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *etcdKV) GetWithCopy(
	key string,
	copySelect kvdb.CopySelect,
) (interface{}, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *etcdKV) SnapPut(snapKvp *kvdb.KVPair) (*kvdb.KVPair, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *etcdKV) AddUser(username string, password string) error {
	// Create a role for this user
	roleName := username
	err := kv.authRole.AddRole(context.Background(), roleName)
	if err != nil {
		return err
	}
	// Create the user
	err = kv.authUser.AddUser(context.Background(), username, password)
	if err != nil {
		return err
	}
	// Assign role to user
	_, err = kv.authUser.GrantUser(context.Background(), username, []string{roleName})
	return err
}

func (kv *etcdKV) RemoveUser(username string) error {
	// Revoke user from this role
	roleName := username
	_, err := kv.authUser.RevokeUser(context.Background(), username, []string{roleName})
	if err != nil {
		return err
	}
	// Remove the role defined for this user
	err = kv.authRole.RemoveRole(context.Background(), roleName)
	if err != nil {
		return err
	}
	// Remove the user
	return kv.authUser.RemoveUser(context.Background(), username)
}

func (kv *etcdKV) GrantUserAccess(username string, permType kvdb.PermissionType, subtree string) error {
	var domain string
	if kv.domain[0] == '/' {
		domain = kv.domain
	} else {
		domain = "/" + kv.domain
	}
	subtree = domain + subtree
	etcdPermType, err := getEtcdPermType(permType)
	if err != nil {
		return err
	}
	// A role for this user has already been created
	// Just assign the subtree to this role
	roleName := username
	_, err = kv.authRole.GrantRoleKV(context.Background(), roleName, []string{subtree}, etcdPermType)
	return err
}

func (kv *etcdKV) RevokeUsersAccess(username string, permType kvdb.PermissionType, subtree string) error {
	var domain string
	if kv.domain[0] == '/' {
		domain = kv.domain
	} else {
		domain = "/" + kv.domain
	}
	subtree = domain + subtree
	etcdPermType, err := getEtcdPermType(permType)
	if err != nil {
		return err
	}
	roleName := username
	// A role for this user should ideally exist
	// Revoke the specfied permission for that subtree
	_, err = kv.authRole.RevokeRoleKV(context.Background(), roleName, []string{subtree}, etcdPermType)
	return err
}

func (kv *etcdKV) Serialize() ([]byte, error) {
	var allKvps kvdb.KVPairs
	kvps, err := kv.Enumerate("")
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(kvps); i++ {
		kvPair := kvps[i]
		if len(kvPair.Value) > 0 {
			allKvps = append(allKvps, kvPair)
		} else {
			newKvps, err := kv.Enumerate(kvPair.Key)
			if err != nil {
				return nil, err
			}
			if len(newKvps) == 0 {
				allKvps = append(allKvps, kvPair)
			} else if len(newKvps) == 1 {
				allKvps = append(allKvps, newKvps[0])
			} else {
				kvps = append(kvps, newKvps...)
			}
		}
	}
	return kv.SerializeAll(allKvps)
}

func (kv *etcdKV) Deserialize(b []byte) (kvdb.KVPairs, error) {
	return kv.DeserializeAll(b)
}

func getEtcdPermType(permType kvdb.PermissionType) (e.PermissionType, error) {
	switch permType {
	case kvdb.ReadPermission:
		return e.ReadPermission, nil
	case kvdb.WritePermission:
		return e.WritePermission, nil
	case kvdb.ReadWritePermission:
		return e.ReadWritePermission, nil
	default:
		return -1, kvdb.ErrUnknownPermission
	}
}
