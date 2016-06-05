// Package consul implements the KVDB interface based on consul.
// Code from docker/libkv was leveraged to build parts of this module.
package consul

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/hashicorp/consul/api"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/common"
	"github.com/portworx/kvdb/mem"
)

const (
	// Name is the name of this kvdb implementation.
	Name      = "consul-kv"
	bootstrap = "kvdb/bootstrap"

	MaxRenewRetries = 5
)

var (
	defaultMachines = []string{"127.0.0.1:8500"}
)

func init() {
	if err := kvdb.Register(Name, New); err != nil {
		panic(err.Error())
	}
}

type consulKV struct {
	client *api.Client
	config *api.Config
	domain string
}

type consulLock struct {
	lock    *api.Lock
	renewCh chan struct{}
}

// New constructs a new kvdb.Kvdb.
func New(
	domain string,
	machines []string,
	options map[string]string,
) (kvdb.Kvdb, error) {
	if len(machines) == 0 {
		machines = defaultMachines
	} else {
		if strings.HasPrefix(machines[0], "http://") {
			machines[0] = strings.TrimPrefix(machines[0], "http://")
		} else if strings.HasPrefix(machines[0], "https://") {
			machines[0] = strings.TrimPrefix(machines[0], "https://")
		}
	}
	config := api.DefaultConfig()
	config.HttpClient = http.DefaultClient
	config.Address = machines[0]
	config.Scheme = "http"

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	if domain != "" && !strings.HasSuffix(domain, "/") {
		domain = domain + "/"
	}

	return &consulKV{
		client,
		config,
		domain,
	}, nil
}

func (kv *consulKV) String() string {
	return Name
}

func (kv *consulKV) Get(key string) (*kvdb.KVPair, error) {
	options := &api.QueryOptions{
		AllowStale:        false,
		RequireConsistent: true,
	}
	key = kv.domain + key
	pair, meta, err := kv.client.KV().Get(key, options)
	if err != nil {
		return nil, err
	}
	if pair == nil {
		return nil, kvdb.ErrNotFound
	}
	return kv.pairToKv("get", pair, meta), nil
}

func (kv *consulKV) GetVal(key string, val interface{}) (*kvdb.KVPair, error) {
	kvp, err := kv.Get(key)
	if err != nil {
		return nil, err
	}
	return kvp, json.Unmarshal(kvp.Value, val)
}

func (kv *consulKV) Put(key string, val interface{}, ttl uint64) (*kvdb.KVPair, error) {
	pathKey := kv.domain + key
	b, err := common.ToBytes(val)
	if err != nil {
		return nil, err
	}
	pair := &api.KVPair{
		Key:   pathKey,
		Value: b,
	}

	if ttl > 0 {
		return nil, kvdb.ErrTTLNotSupported

		// Future Use : To Support TTL values
		/* for retries := 1; retries <= MaxRenewRetries; retries++ {
			err := kv.renewSession(pair, ttl)
			if err == nil {
				break
			}
			if retries == MaxRenewRetries {
				return nil, kvdb.ErrSetTTLFailed
			}
		}*/
	}

	if _, err := kv.client.KV().Put(pair, nil); err != nil {
		return nil, err
	}

	kvPair, err := kv.Get(key)
	if err != nil {
		return nil, err
	}
	kvPair.Action = kvdb.KVSet
	return kvPair, nil
}

func (kv *consulKV) Create(key string, val interface{}, ttl uint64) (*kvdb.KVPair, error) {
	if _, err := kv.Get(key); err == nil {
		return nil, kvdb.ErrExist
	}

	kvPair, err := kv.Put(key, val, ttl)
	if err != nil {
		return nil, err
	}
	kvPair.Action = kvdb.KVCreate
	return kvPair, nil
}

func (kv *consulKV) Update(key string, val interface{}, ttl uint64) (*kvdb.KVPair, error) {
	if _, err := kv.Get(key); err != nil {
		return nil, err
	}

	kvPair, err := kv.Put(key, val, ttl)
	if err != nil {
		return nil, err
	}
	kvPair.Action = kvdb.KVSet
	return kvPair, nil
}

func (kv *consulKV) Enumerate(prefix string) (kvdb.KVPairs, error) {
	prefix = kv.domain + prefix
	pairs, meta, err := kv.client.KV().List(prefix, nil)
	if err != nil {
		return nil, err
	}
	if pairs == nil {
		return nil, kvdb.ErrNotFound
	}
	return kv.pairToKvs("enumerate", pairs, meta), nil
}

func (kv *consulKV) Delete(key string) (*kvdb.KVPair, error) {
	pair, err := kv.Get(key)
	if err != nil {
		return nil, err
	}
	key = kv.domain + key
	if _, err := kv.client.KV().Delete(key, nil); err != nil {
		return nil, err
	}
	return pair, nil
}

func (kv *consulKV) DeleteTree(key string) error {
	key = kv.domain + key
	if _, err := kv.client.KV().DeleteTree(key, nil); err != nil {
		return err
	}
	return nil
}

func (kv *consulKV) Keys(prefix, key string) ([]string, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *consulKV) CompareAndSet(kvp *kvdb.KVPair, flags kvdb.KVFlags, prevValue []byte) (*kvdb.KVPair, error) {
	pair := &api.KVPair{
		Key:   kv.domain + kvp.Key,
		Value: kvp.Value,
		Flags: api.LockFlagValue,
	}
	if (flags & kvdb.KVModifiedIndex) != 0 {
		pair.ModifyIndex = kvp.ModifiedIndex
	} else if (flags&kvdb.KVModifiedIndex) == 0 && prevValue != nil {
		kvPair, err := kv.Get(kvp.Key)
		if err != nil {
			return nil, err
		}

		// Prev Value not equal to current value in etcd
		if bytes.Compare(kvPair.Value, prevValue) != 0 {
			return nil, kvdb.ErrValueMismatch
		}
		pair.ModifyIndex = kvPair.ModifiedIndex
	} else {
		pair.ModifyIndex = 0
	}

	ok, _, err := kv.client.KV().CAS(pair, nil)
	if err != nil {
		return nil, err
	}

	if !ok {
		if (flags & kvdb.KVModifiedIndex) == 0 {
			return nil, kvdb.ErrValueMismatch
		}
		return nil, kvdb.ErrModified
	}

	kvPair, err := kv.Get(kvp.Key)
	if err != nil {
		return nil, err
	}
	return kvPair, nil
}

func (kv *consulKV) CompareAndDelete(kvp *kvdb.KVPair, flags kvdb.KVFlags) (*kvdb.KVPair, error) {
	pair := &api.KVPair{
		Key:   kv.domain + kvp.Key,
		Value: kvp.Value,
		Flags: api.LockFlagValue,
	}

	if (flags & kvdb.KVModifiedIndex) == 0 {
		return nil, kvdb.ErrNotSupported
	}
	pair.ModifyIndex = kvp.ModifiedIndex

	ok, _, err := kv.client.KV().DeleteCAS(pair, nil)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, kvdb.ErrModified
	}
	return kvp, nil
}

func (kv *consulKV) WatchKey(key string, waitIndex uint64, opaque interface{}, cb kvdb.WatchCB) error {
	var keyExist bool
	kvp, err := kv.Get(key)
	if err == kvdb.ErrNotFound {
		keyExist = false
	} else if err != nil {
		return err
	} else {
		keyExist = true
	}

	if waitIndex == 0 && kvp != nil {
		waitIndex = kvp.KVDBIndex
	}

	key = kv.domain + key
	go kv.watchKeyStart(key, keyExist, waitIndex, opaque, cb)
	return nil
}

func (kv *consulKV) WatchTree(prefix string, waitIndex uint64, opaque interface{}, cb kvdb.WatchCB) error {
	var prefixExist bool
	kvps, err := kv.Enumerate(prefix)
	if err == kvdb.ErrNotFound {
		prefixExist = false
	} else if err != nil {
		return err
	} else {
		prefixExist = true
	}

	if waitIndex == 0 && kvps != nil && len(kvps) != 0 {
		waitIndex = kvps[0].KVDBIndex
	}

	prefix = kv.domain + prefix
	go kv.watchTreeStart(prefix, prefixExist, waitIndex, opaque, cb)
	return nil
}

func (kv *consulKV) Lock(key string, ttl uint64) (*kvdb.KVPair, error) {
	// Strip of the leading slash or else consul throws error
	if key[0] == '/' {
		key = key[1:]
	}

	l, err := kv.getLock(key, ttl)
	if err != nil {
		return nil, err
	}
	if _, err := l.lock.Lock(nil); err != nil {
		return nil, err
	}
	return &kvdb.KVPair{
		Key:  key,
		Lock: l,
	}, nil
}

func (kv *consulKV) Unlock(kvp *kvdb.KVPair) error {
	return kvp.Lock.(*consulLock).lock.Unlock()
}

func (kv *consulKV) TxNew() (kvdb.Tx, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *consulKV) Snapshot(prefix string) (kvdb.Kvdb, uint64, error) {
	// Create a new bootstrap key : lowest index
	r := rand.New(rand.NewSource(time.Now().UnixNano())).Int63()
	bootStrapKeyLow := bootstrap + strconv.FormatInt(r, 10) +
		strconv.FormatInt(time.Now().UnixNano(), 10)
	val, _ := common.ToBytes(time.Now().UnixNano())
	kvPair, err := kv.Put(bootStrapKeyLow, val, 0)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create snap bootstrap key %v, "+
			"err: %v", bootStrapKeyLow, err)
	}
	lowestKvdbIndex := kvPair.ModifiedIndex

	options := &api.QueryOptions{
		AllowStale:        false,
		RequireConsistent: true,
	}
	pairs, _, err := kv.client.KV().List(kv.domain+prefix, options)
	if err != nil {
		return nil, 0, err
	}

	kvPairs := kv.pairToKvs("enumerate", pairs, nil)
	snapDb, err := mem.New(kv.domain, nil, nil)
	if err != nil {
		return nil, 0, err
	}

	for _, kvPair := range kvPairs {
		_, err := snapDb.Put(kvPair.Key, kvPair.Value, 0)
		if err != nil {
			return nil, 0, fmt.Errorf("Failed creating snap: %v", err)
		}
	}

	// Create bootrap key : highest index
	bootStrapKeyHigh := bootstrap + strconv.FormatInt(r, 10) +
		strconv.FormatInt(time.Now().UnixNano(), 10)
	val, _ = common.ToBytes(time.Now().UnixNano())

	kvPair, err = kv.Put(bootStrapKeyHigh, val, 0)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create snap bootstrap key %v, "+
			"err: %v", bootStrapKeyHigh, err)
	}
	highestKvdbIndex := kvPair.ModifiedIndex

	// In consul Delete does not increment kvdb index.
	// Hence the put (bootstrap) key and delete both return the same index
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

	kv.Delete(bootStrapKeyLow)
	kv.Delete(bootStrapKeyHigh)
	return snapDb, highestKvdbIndex, nil
}

func (kv *consulKV) createKv(pair *api.KVPair) *kvdb.KVPair {
	kvp := &kvdb.KVPair{
		Value:         []byte(pair.Value),
		ModifiedIndex: pair.ModifyIndex,
		CreatedIndex:  pair.CreateIndex,
	}
	// Strip out the leading '/'
	if len(pair.Key) != 0 {
		kvp.Key = pair.Key[1:]
	} else {
		kvp.Key = pair.Key
	}
	kvp.Key = strings.TrimPrefix(pair.Key, kv.domain)
	return kvp
}

func (kv *consulKV) pairToKv(action string, pair *api.KVPair, meta *api.QueryMeta) *kvdb.KVPair {
	kvp := kv.createKv(pair)
	switch action {
	case "create":
		kvp.Action = kvdb.KVCreate
	case "set", "update", "put":
		kvp.Action = kvdb.KVSet
	case "delete":
		kvp.Action = kvdb.KVDelete
	case "get":
		kvp.Action = kvdb.KVGet
	default:
		kvp.Action = kvdb.KVUknown
	}
	if meta != nil {
		kvp.KVDBIndex = meta.LastIndex
	}
	return kvp
}

func (kv *consulKV) pairToKvs(action string, pair []*api.KVPair, meta *api.QueryMeta) kvdb.KVPairs {
	kvs := make([]*kvdb.KVPair, len(pair))
	for i := range pair {
		kvs[i] = kv.pairToKv(action, pair[i], meta)
		if meta != nil {
			kvs[i].KVDBIndex = meta.LastIndex
		}
	}
	return kvs
}

func (kv *consulKV) getLock(key string, ttl uint64) (*consulLock, error) {
	key = kv.domain + key
	lockOpts := &api.LockOptions{
		Key: key,
	}
	lock := &consulLock{}
	if ttl != 0 {
		TTL := time.Duration(0)
		entry := &api.SessionEntry{
			Behavior:  api.SessionBehaviorRelease, // Release the lock when the session expires
			TTL:       (TTL / 2).String(),         // Consul multiplies the TTL by 2x
			LockDelay: 1 * time.Millisecond,       // Virtually disable lock delay
		}

		// Create the key session
		session, _, err := kv.client.Session().Create(entry, nil)
		if err != nil {
			return nil, err
		}

		// Place the session on lock
		lockOpts.Session = session

		// Renew the session ttl lock periodically
		go func() {
			// TODO: do something with the error
			_ = kv.client.Session().RenewPeriodic(entry.TTL, session, nil, nil)
		}()
	}

	l, err := kv.client.LockOpts(lockOpts)
	if err != nil {
		return nil, err
	}
	lock.lock = l
	return lock, nil
}

func (kv *consulKV) watchTreeStart(prefix string, prefixExisted bool, waitIndex uint64, opaque interface{}, cb kvdb.WatchCB) {
	opts := &api.QueryOptions{
		WaitIndex: waitIndex,
	}
	prefixDeleted := false
	var cbErr error
	for {
		// Make a blocking List query
		pairs, meta, err := kv.client.KV().List(prefix, opts)
		if pairs == nil && prefixExisted && !prefixDeleted {
			// Got a delete on the prefix of the tree (Last Key under the tree being deleted)
			pair := &api.KVPair{
				Key:   prefix,
				Value: nil,
			}
			kvPair := kv.pairToKv("delete", pair, meta)
			kvPair.ModifiedIndex = meta.LastIndex

			// Callback with a delete action
			cbErr = cb(prefix, opaque, kvPair, nil)
			prefixDeleted = true

			// Set the wait index so that we block on the next List call
			opts.WaitIndex = meta.LastIndex

		} else if pairs == nil && !prefixExisted {
			// Prefix Never existed or hasn't been created yet
			continue
		} else if pairs == nil && prefixDeleted {
			// Prefix has been deleted and this is a recurring list.
			// Do not execute callback
			// Set the waitIndex so that we block on the next Get call
			opts.WaitIndex = meta.LastIndex
			continue
		} else if err != nil {
			logrus.Errorf("Consul returned an error : %s\n", err.Error())
			cbErr = cb(prefix, opaque, nil, err)
		} else {
			// Same waitIndex as previous. Out of blocking call because
			// waitTime timeouted. (This should not happen)
			if opts.WaitIndex == meta.LastIndex {
				continue
			}
			// Set the waitIndex so that we block on the next List call
			opts.WaitIndex = meta.LastIndex

			// Find the key value pair that was added/modified/deleted
			found := false
			for _, pair := range pairs {
				if pair.ModifyIndex == meta.LastIndex {
					if pair.CreateIndex == pair.ModifyIndex {
						// Callback with a create action
						cbErr = cb(prefix, opaque, kv.pairToKv("create", pair, meta), nil)
						prefixDeleted = false
						prefixExisted = true
					} else {
						// Callback with an update action
						cbErr = cb(prefix, opaque, kv.pairToKv("update", pair, meta), nil)
					}
					found = true
					break
				}
			}
			if found != true {
				// We had a sub-key delete
				pair := &api.KVPair{
					Key:   prefix,
					Value: nil,
				}
				kvPair := kv.pairToKv("delete", pair, meta)
				kvPair.ModifiedIndex = meta.LastIndex
				cbErr = cb(prefix, opaque, kvPair, nil)
			}
		}
		if cbErr != nil {
			_ = cb(prefix, opaque, nil, kvdb.ErrWatchStopped)
			break
		}
	}
}

func (kv *consulKV) watchKeyStart(key string, keyExisted bool, waitIndex uint64, opaque interface{}, cb kvdb.WatchCB) {
	opts := &api.QueryOptions{
		WaitIndex: waitIndex,
	}

	// Flags used to detect a deleted key
	keyDeleted := false
	var cbErr error
	for {
		// Make a blocking Get query
		pair, meta, err := kv.client.KV().Get(key, opts)
		if pair == nil && keyExisted && !keyDeleted {
			// Key being Deleted for the first time after its creation
			pair = &api.KVPair{
				Key:   key,
				Value: nil,
			}
			kvPair := kv.pairToKv("delete", pair, meta)
			kvPair.ModifiedIndex = meta.LastIndex

			// Callback with a delete action
			cbErr = cb(key, opaque, kvPair, nil)

			// Set the waitIndex so that we block on the next Get call
			opts.WaitIndex = meta.LastIndex
			keyDeleted = true
		} else if pair == nil && !keyExisted {
			// Key Never existed or hasn't been created yet
			// Set the waitIndex so that we block on the next Get call
			opts.WaitIndex = meta.LastIndex
			continue
		} else if pair == nil && keyDeleted {
			// Key has been deleted and this is a recurring get.
			// Do not execute callback
			// Set the waitIndex so that we block on the next Get call
			opts.WaitIndex = meta.LastIndex
			continue
		} else if err != nil {
			logrus.Errorf("Consul returned an error : %s\n", err.Error())
			cbErr = cb(key, opaque, nil, err)
		} else {
			// If LastIndex didn't change it means Get returned because
			// of Wait timeout
			if opts.WaitIndex == meta.LastIndex {
				continue
			}
			// Set the waitIndex so that we block on the next Get call
			opts.WaitIndex = meta.LastIndex

			if pair.CreateIndex == pair.ModifyIndex {
				// Callback with a create action
				cbErr = cb(key, opaque, kv.pairToKv("create", pair, meta), nil)
				keyDeleted = false
				keyExisted = true
			} else {
				// Callback with a update action
				cbErr = cb(key, opaque, kv.pairToKv("update", pair, meta), nil)
			}
		}
		if cbErr != nil {
			_ = cb(key, opaque, nil, kvdb.ErrWatchStopped)
			break
		}
	}
}

// Future Use : Support for ttl values in create/update/put
/*
func (kv *consulKV) renewSession(pair *api.KVPair, ttl uint64) error {
	// Check if there is any previous session with an active TTL
	session, err := kv.getActiveSession(pair.Key)
	if err != nil {
		return err
	}

	if session != "" {
		// Destroy the existing session associated with the key
		_, err := kv.client.Session().Destroy(session, nil)
		if err != nil {
			return err
		}
	}

	//if ttl < 20 {
		// Consul requires minimum of 10 sec of ttl
		//ttl = 20
	//}
	// Consul multiplies the TTL by 2x
	durationTTL := time.Duration(int64(ttl)) * time.Second

	entry := &api.SessionEntry{
		Behavior:  api.SessionBehaviorDelete, // Delete the key when the session expires
		TTL:       durationTTL.String(),
		LockDelay: 1 * time.Millisecond,      // Virtually disable lock delay
	}

	// Create the key session
	session, _, err = kv.client.Session().Create(entry, nil)
	if err != nil {
		return err
	}

	lockOpts := &api.LockOptions{
		Key:     pair.Key,
		Session: session,
	}

	// Lock and ignore if lock is held
	// It's just a placeholder for the
	// ephemeral behavior
	lock, _ := kv.client.LockOpts(lockOpts)
	if lock != nil {
		lock.Lock(nil)
	}

	_, _, err = kv.client.Session().Renew(session, nil)
	return err
}

// getActiveSession checks if the key already has
// a session attached
func (kv *consulKV) getActiveSession(key string) (string, error) {
	pair, _, err := kv.client.KV().Get(key, nil)
	if err != nil {
		return "", err
	}
	if pair != nil && pair.Session != "" {
		return pair.Session, nil
	}
	return "", nil
}
*/
