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

	"github.com/Sirupsen/logrus"
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

	defaultRetryCount             = 60
	defaultIntervalBetweenRetries = time.Millisecond * 500
	bootstrap                     = "kvdb/bootstrap"
	// the maximum amount of time a dial will wait for a connection to setup.
	// 30s is long enough for most of the network conditions.
	defaultDialTimeout = 30 * time.Second
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
	client   e.KeysAPI
	authUser e.AuthUserAPI
	authRole e.AuthRoleAPI
	domain   string
}

type etcdLock struct {
	done     chan struct{}
	unlocked bool
	err      error
	sync.Mutex
}

// LockerIDInfo id of locker
type LockerIDInfo struct {
	LockerID string
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
	var username, password, caFile string
	// options provided. Probably auth options
	if options != nil || len(options) > 0 {
		var ok bool
		// Check if username provided
		username, ok = options[kvdb.UsernameKey]
		if ok {
			// Check if password provided
			password, ok = options[kvdb.PasswordKey]
			if !ok {
				return nil, kvdb.ErrNoPassword
			}
			// Check if certificate provided
			caFile, ok = options[kvdb.CAFileKey]
			if !ok {
				return nil, kvdb.ErrNoCertificate
			}
		}
	}
	tls := transport.TLSInfo{
		CAFile: caFile,
	}
	tr, err := transport.NewTransport(tls, defaultDialTimeout)
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
		e.NewKeysAPI(c),
		e.NewAuthUserAPI(c),
		e.NewAuthRoleAPI(c),
		domain,
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
	for i := 0; i < defaultRetryCount; i++ {
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

func (kv *etcdKV) Lock(key string) (*kvdb.KVPair, error) {
	return kv.LockWithID(key, "locked")
}

func (kv *etcdKV) LockWithID(key string, lockerID string) (
	*kvdb.KVPair,
	error,
) {
	key = kv.domain + key
	duration := time.Second
	ttl := uint64(4)
	count := 0
	lockTag := LockerIDInfo{LockerID: lockerID}
	kvPair, err := kv.Create(key, lockTag, ttl)
	for maxCount := 300; err != nil && count < maxCount; count++ {
		time.Sleep(duration)
		kvPair, err = kv.Create(key, lockTag, ttl)
		if count > 0 && count%15 == 0 && err != nil {
			currLockerTag := LockerIDInfo{LockerID: ""}
			if _, errGet := kv.GetVal(key, &currLockerTag); errGet == nil {
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
	for i := 0; i < defaultRetryCount; i++ {
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
	snapDb, err := mem.New(
		kv.domain,
		nil,
		map[string]string{mem.KvSnap: "true"},
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
