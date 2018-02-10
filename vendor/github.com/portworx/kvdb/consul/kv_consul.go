// Package consul implements the KVDB interface based on consul.
// Code from docker/libkv was leveraged to build parts of this module.
package consul

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hashicorp/consul/api"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/common"
	"github.com/portworx/kvdb/mem"
)

const (
	// Name is the name of this kvdb implementation.
	Name      = "consul-kv"
	bootstrap = "kvdb/bootstrap"
	// MaxRenewRetries to renew TTL.
	MaxRenewRetries = 5
)

var (
	defaultMachines = []string{"127.0.0.1:8500"}
)

// CKVPairs sortable KVPairs
type CKVPairs api.KVPairs

func (c CKVPairs) Len() int {
	return len(c)
}

func (c CKVPairs) Less(i, j int) bool {
	return c[i].ModifyIndex < c[j].ModifyIndex
}

func (c CKVPairs) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func init() {
	if err := kvdb.Register(Name, New, Version); err != nil {
		panic(err.Error())
	}
}

func stripConsecutiveForwardslash(key string) string {
	// Replace consecutive occurences of forward slash with single occurrence
	re := regexp.MustCompile("(//*)")
	return re.ReplaceAllString(key, "/")
}

type consulKV struct {
	common.BaseKvdb
	client *api.Client
	config *api.Config
	domain string
	kvdb.Controller
}

type consulLock struct {
	lock   *api.Lock
	doneCh chan struct{}
	tag    interface{}
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
	} else {
		if strings.HasPrefix(machines[0], "http://") {
			machines[0] = strings.TrimPrefix(machines[0], "http://")
		} else if strings.HasPrefix(machines[0], "https://") {
			machines[0] = strings.TrimPrefix(machines[0], "https://")
		}
	}

	var token string
	// options provided. Probably auth options
	if options != nil || len(options) > 0 {
		var ok bool
		// Check if username provided
		_, ok = options[kvdb.UsernameKey]
		if ok {
			return nil, kvdb.ErrAuthNotSupported
		}
		// Check if password provided
		_, ok = options[kvdb.PasswordKey]
		if ok {
			return nil, kvdb.ErrAuthNotSupported
		}
		// Check if certificate provided
		_, ok = options[kvdb.CAFileKey]
		if ok {
			return nil, kvdb.ErrAuthNotSupported
		}
		// Get the ACL token if provided
		token, ok = options[kvdb.ACLTokenKey]

	}

	config := api.DefaultConfig()
	config.HttpClient = http.DefaultClient
	config.Address = machines[0]
	config.Scheme = "http"
	config.Token = token

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	if domain != "" && !strings.HasSuffix(domain, "/") {
		domain = domain + "/"
	}

	return &consulKV{
		common.BaseKvdb{FatalCb: fatalErrorCb},
		client,
		config,
		domain,
		kvdb.ControllerNotSupported,
	}, nil
}

// Version returns the supported version for consul api
func Version(url string, kvdbOptions map[string]string) (string, error) {
	// Currently we support only v1
	return kvdb.ConsulVersion1, nil
}

func (kv *consulKV) String() string {
	return Name
}

func (kv *consulKV) Capabilities() int {
	return 0
}

func (kv *consulKV) Get(key string) (*kvdb.KVPair, error) {
	options := &api.QueryOptions{
		AllowStale:        false,
		RequireConsistent: true,
	}
	key = kv.domain + key
	key = stripConsecutiveForwardslash(key)
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

func (kv *consulKV) createTTLSession(
	key string,
	val interface{},
	ttl uint64,
	noCreate bool,
) (*api.KVPair, error) {
	pathKey := kv.domain + key
	pathKey = stripConsecutiveForwardslash(pathKey)
	b, err := common.ToBytes(val)
	if err != nil {
		return nil, err
	}
	pair := &api.KVPair{
		Key:   pathKey,
		Value: b,
	}

	if ttl > 0 {
		if ttl < 20 {
			return nil, kvdb.ErrTTLNotSupported
		}
		// Future Use : To Support TTL values
		for retries := 1; retries <= MaxRenewRetries; retries++ {
			// Consul doubles the ttl value. Hence we divide it by 2
			// Consul does not support ttl values less than 10.
			// Hence we set our lower limit to 20
			session, err := kv.renewSession(pair, ttl/2, noCreate)
			if err == nil {
				pair.Session = session
				break
			}
			if retries == MaxRenewRetries {
				return nil, kvdb.ErrSetTTLFailed
			}
		}
	}
	return pair, nil
}

func (kv *consulKV) Put(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
	pair, err := kv.createTTLSession(key, val, ttl, false)
	if err != nil {
		return nil, err
	}
	if ttl == 0 {
		if _, err := kv.client.KV().Put(pair, nil); err != nil {
			return nil, err
		}
	} else {
		// It is unclear why err == nil but ok == false. We always
		// delete any existing sessions on Put, so this should work fine.
		ok, _, err := kv.client.KV().Acquire(pair, nil)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("Acquire failed")
		}
	}

	kvPair, err := kv.Get(key)
	if err != nil {
		return nil, err
	}
	kvPair.Action = kvdb.KVSet
	return kvPair, nil
}

func (kv *consulKV) Create(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
	sessionPair, err := kv.createTTLSession(key, val, ttl, true)
	if err != nil {
		return nil, err
	}
	kvPair := &kvdb.KVPair{Key: key, Value: sessionPair.Value}
	kvPair, err = kv.CompareAndSet(kvPair, kvdb.KVModifiedIndex, nil)
	if err == nil {
		kvPair.Action = kvdb.KVCreate
		if ttl > 0 {
			var ok bool
			ok, _, err = kv.client.KV().Acquire(sessionPair, nil)
			if ok && err == nil {
				return kvPair, err
			}
			kv.client.Session().Destroy(sessionPair.Session, nil)
			kv.Delete(key)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("Failed to set ttl")
			}
		}
	}
	if err == kvdb.ErrModified {
		// key already exists since compare and set with index 0 failed.
		err = kvdb.ErrExist
	}
	return kvPair, err
}

func (kv *consulKV) Update(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
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
	prefix = stripConsecutiveForwardslash(prefix)
	pairs, meta, err := kv.client.KV().List(prefix, nil)
	if err != nil {
		return nil, err
	}
	return kv.pairToKvs("enumerate", pairs, meta), nil
}

func (kv *consulKV) Delete(key string) (*kvdb.KVPair, error) {
	pair, err := kv.Get(key)
	if err != nil {
		return nil, err
	}
	key = kv.domain + key
	key = stripConsecutiveForwardslash(key)
	if _, err := kv.client.KV().Delete(key, nil); err != nil {
		return nil, err
	}
	return pair, nil
}

func (kv *consulKV) DeleteTree(key string) error {
	key = kv.domain + key
	key = stripConsecutiveForwardslash(key)
	if _, err := kv.client.KV().DeleteTree(key, nil); err != nil {
		return err
	}
	return nil
}

func (kv *consulKV) Keys(prefix, sep string) ([]string, error) {
	if "" == sep {
		sep = "/"
	}
	prefix = kv.domain + prefix
	prefix = stripConsecutiveForwardslash(prefix)
	lenPrefix := len(prefix)
	lenSep := len(sep)
	if prefix[lenPrefix-lenSep:] != sep {
		prefix += sep
		lenPrefix += lenSep
	}
	list, _, err := kv.client.KV().Keys(prefix, sep, nil)
	if err != nil {
		return nil, err
	}
	var retList []string
	if len(list) > 0 {
		retList = make([]string, len(list))
		for i, key := range list {
			if strings.HasPrefix(key, prefix) {
				key = key[lenPrefix:]
			}
			if lky := len(key); lky > lenSep && key[lky-lenSep:] == sep {
				key = key[0 : lky-lenSep]
			}
			retList[i] = key
		}
	}
	return retList, nil
}

func (kv *consulKV) CompareAndSet(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
	prevValue []byte,
) (*kvdb.KVPair, error) {
	key := kv.domain + kvp.Key
	key = stripConsecutiveForwardslash(key)
	pair := &api.KVPair{
		Key:   key,
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

func (kv *consulKV) CompareAndDelete(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
) (*kvdb.KVPair, error) {
	key := kv.domain + kvp.Key
	key = stripConsecutiveForwardslash(key)
	pair := &api.KVPair{
		Key:   key,
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

func (kv *consulKV) WatchKey(
	key string,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) error {
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

func (kv *consulKV) Lock(key string) (*kvdb.KVPair, error) {
	return kv.LockWithID(key, "locked")
}

func (kv *consulKV) LockWithID(key string, lockerID string) (
	*kvdb.KVPair,
	error,
) {
	key = stripConsecutiveForwardslash(key)
	// Strip of the leading slash or else consul throws error
	if key[0] == '/' {
		key = key[1:]
	}

	l, err := kv.getLock(key, lockerID, 20*time.Second)
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
	l, ok := kvp.Lock.(*consulLock)
	if !ok {
		return fmt.Errorf("Invalid lock structure for key: %v", string(kvp.Key))
	}
	_, err := kv.Delete(kvp.Key)
	if err == nil {
		_ = l.lock.Unlock()
		if l.doneCh != nil {
			close(l.doneCh)
		}
		return nil
	}
	logrus.Errorf("Unlock failed for key: %s, tag: %s, error: %s", kvp.Key,
		l.tag, err.Error())
	return err
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

	listKey := kv.domain + prefix
	listKey = stripConsecutiveForwardslash(listKey)
	pairs, _, err := kv.client.KV().List(listKey, options)
	if err != nil {
		return nil, 0, err
	}

	kvPairs := kv.pairToKvs("enumerate", pairs, nil)
	snapDb, err := mem.New(
		kv.domain,
		nil,
		map[string]string{mem.KvSnap: "true"},
		kv.FatalCb,
	)
	if err != nil {
		return nil, 0, err
	}

	for _, kvPair := range kvPairs {
		_, err := snapDb.SnapPut(kvPair)
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

			if kvp.ModifiedIndex > highestKvdbIndex {
				// done applying changes, just return
				watchErr = fmt.Errorf("done")
				sendErr = nil
				goto errordone
			} else if kvp.ModifiedIndex == highestKvdbIndex {
				// last update that we needed. Put it inside snap db
				// and return
				_, err = snapDb.SnapPut(kvp)
				if err != nil {
					watchErr = fmt.Errorf("Failed to apply update to snap: %v", err)
					sendErr = watchErr
				} else {
					watchErr = fmt.Errorf("done")
					sendErr = nil
				}
				goto errordone
			} else {
				if kvp.Action == kvdb.KVDelete {
					_, err = snapDb.Delete(kvp.Key)
					// A Delete key was issued between our first lowestKvdbIndex Put
					// and Enumerate APIs in this function
					if err == kvdb.ErrNotFound {
						err = nil
					}

				} else {
					_, err = snapDb.SnapPut(kvp)
				}
				if err != nil {
					watchErr = fmt.Errorf("Failed to apply update to snap: %v", err)
					sendErr = watchErr
					goto errordone
				}
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

	_, err = kv.Delete(bootStrapKeyLow)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to delete snap bootstrap key: %v, "+
			"err: %v", bootStrapKeyLow, err)
	}
	_, err = kv.Delete(bootStrapKeyHigh)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to delete snap bootstrap key: %v, "+
			"err: %v", bootStrapKeyHigh, err)
	}
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

func (kv *consulKV) EnumerateWithSelect(
	prefix string,
	enumerateSelect kvdb.EnumerateSelect,
	copySelect kvdb.CopySelect,
) ([]interface{}, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *consulKV) GetWithCopy(
	key string,
	copySelect kvdb.CopySelect,
) (interface{}, error) {
	return nil, kvdb.ErrNotSupported
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

func isHidden(key string) bool {
	tokens := strings.Split(key, "/")
	keySuffix := tokens[len(tokens)-1]
	return keySuffix != "" && keySuffix[0] == '_'
}

func (kv *consulKV) pairToKvs(
	action string,
	pairs []*api.KVPair,
	meta *api.QueryMeta,
) kvdb.KVPairs {
	kvs := []*kvdb.KVPair{}
	for _, pair := range pairs {
		// Ignore hidden keys.
		if isHidden(pair.Key) {
			continue
		}
		kvs = append(kvs, kv.pairToKv(action, pair, meta))
	}
	return kvs
}

func (kv *consulKV) renewLockSession(
	key string,
	initialTTL string,
	session string,
	doneCh chan struct{},
	tag interface{},
) {
	go func() {
		_ = kv.client.Session().RenewPeriodic(initialTTL, session, nil, doneCh)
	}()
	if kv.LockTimeout > 0 {
		go func() {
			timeout := time.After(kv.LockTimeout)
			for {
				select {
				case <-timeout:
					kv.LockTimedout(fmt.Sprintf("Key:%s,Tag:%v", key, tag))
				case <-doneCh:
					return
				}
			}
		}()
	}
}

func (kv *consulKV) getLock(key string, tag interface{}, ttl time.Duration) (
	*consulLock,
	error,
) {
	key = kv.domain + key
	tagValue, err := common.ToBytes(tag)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert tag: %v, error: %v", tag,
			err)
	}
	lockOpts := &api.LockOptions{
		Key:   key,
		Value: tagValue,
	}
	lock := &consulLock{}
	entry := &api.SessionEntry{
		Behavior:  api.SessionBehaviorRelease, // Release the lock when the session expires
		TTL:       (ttl / 2).String(),         // Consul multiplies the TTL by 2x
		LockDelay: 0,                          // Virtually disable lock delay
	}

	// Create the key session
	session, _, err := kv.client.Session().Create(entry, nil)
	if err != nil {
		return nil, err
	}

	// Place the session on lock
	lockOpts.Session = session
	lock.doneCh = make(chan struct{})
	lock.tag = tag

	l, err := kv.client.LockOpts(lockOpts)
	if err != nil {
		return nil, err
	}

	kv.renewLockSession(key, entry.TTL, session, lock.doneCh, tag)
	lock.lock = l
	return lock, nil
}

func (kv *consulKV) watchTreeStart(
	prefix string,
	prefixExisted bool,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) {
	prefix = stripConsecutiveForwardslash(prefix)
	opts := &api.QueryOptions{
		WaitIndex:         waitIndex,
		RequireConsistent: true,
	}
	prefixDeleted := false
	prevIndex := uint64(0)
	var cbCreateErr, cbUpdateErr error

	checkIndex := func(prevIndex *uint64, pair *api.KVPair, newIndex uint64,
		msg string, lastIndex, waitIndex uint64) {
		if *prevIndex != 0 && newIndex <= *prevIndex {
			kv.FatalCb(msg+" with index invoked twice: %v, prevIndex: %d"+
				" newIndex: %d, lastIndex: %d, waitIndex: %d", *pair,
				*prevIndex, newIndex, lastIndex, waitIndex)
		}
		*prevIndex = newIndex
	}

	for {
		// Make a blocking List query
		kvPairs, meta, err := kv.client.KV().List(prefix, opts)
		pairs := CKVPairs(kvPairs)
		sort.Sort(pairs)
		if err != nil {
			logrus.Errorf("Consul returned an error : %s\n", err.Error())
			cbUpdateErr = cb(prefix, opaque, nil, err)
		} else if pairs == nil && prefixExisted && !prefixDeleted {
			// Got a delete on the prefix of the tree (Last Key under the tree being deleted)
			pair := &api.KVPair{
				Key:   prefix,
				Value: nil,
			}
			kvPair := kv.pairToKv("delete", pair, meta)
			kvPair.ModifiedIndex = meta.LastIndex
			checkIndex(&prevIndex, pair, kvPair.ModifiedIndex,
				"delete", meta.LastIndex, opts.WaitIndex)

			// Callback with a delete action
			cbUpdateErr = cb(prefix, opaque, kvPair, nil)
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
		} else {
			// Same waitIndex as previous. Out of blocking call because
			// waitTime timeouted. (This should not happen)
			if opts.WaitIndex >= meta.LastIndex {
				continue
			}
			// Find the key value pair(s) that was(were) added/modified/deleted
			found := false
			for _, pair := range pairs {
				// Check if pair's ModifyIndex lies between the wait index and the last modified index
				if pair.ModifyIndex > opts.WaitIndex {
					if pair.CreateIndex == pair.ModifyIndex {
						// Callback with a create action
						checkIndex(&prevIndex, pair, pair.CreateIndex,
							"Create", meta.LastIndex, opts.WaitIndex)
						cbCreateErr = cb(prefix, opaque, kv.pairToKv("create", pair, meta), nil)
						prefixDeleted = false
						prefixExisted = true
					} else if (pair.CreateIndex > opts.WaitIndex) && (pair.ModifyIndex > pair.CreateIndex) {
						// In this single update from consul we have got both a create action and
						// update action for this kvpair. Calling two callback functions with different actions
						checkIndex(&prevIndex, pair, pair.ModifyIndex,
							"Create", meta.LastIndex, opts.WaitIndex)
						cbCreateErr = cb(prefix, opaque, kv.pairToKv("create", pair, meta), nil)
						prefixDeleted = false
						prefixExisted = true
						// Callback with an update action
						cbUpdateErr = cb(prefix, opaque, kv.pairToKv("update", pair, meta), nil)
					} else {
						// Callback with an update action
						checkIndex(&prevIndex, pair, pair.ModifyIndex,
							"Update", meta.LastIndex, opts.WaitIndex)
						cbUpdateErr = cb(prefix, opaque, kv.pairToKv("update", pair, meta), nil)
					}
					found = true
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
				checkIndex(&prevIndex, pair, kvPair.ModifiedIndex, "delete",
					meta.LastIndex, opts.WaitIndex)
				cbUpdateErr = cb(prefix, opaque, kvPair, nil)
			}
			// Set the waitIndex so that we block on the next List call
			opts.WaitIndex = meta.LastIndex
		}
		if cbUpdateErr != nil || cbCreateErr != nil {
			_ = cb(prefix, opaque, nil, kvdb.ErrWatchStopped)
			break
		}
	}
}

func (kv *consulKV) watchKeyStart(
	key string,
	keyExisted bool,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) {
	key = stripConsecutiveForwardslash(key)
	opts := &api.QueryOptions{
		WaitIndex: waitIndex,
	}

	// Flags used to detect a deleted key
	keyDeleted := false
	var cbErr error
	for {
		// Make a blocking Get query
		pair, meta, err := kv.client.KV().Get(key, opts)
		if err != nil {
			logrus.Errorf("Consul returned an error : %s\n", err.Error())
			cbErr = cb(key, opaque, nil, err)
		} else if pair == nil && keyExisted && !keyDeleted {
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
func (kv *consulKV) renewSession(
	pair *api.KVPair,
	ttl uint64,
	noCreate bool,
) (string, error) {
	// Check if there is any previous session with an active TTL
	session, err := kv.getActiveSession(pair.Key)
	if err != nil {
		logrus.Infof("Failed to find session: %v", err)
		return "", err
	}

	if session != "" {
		if noCreate {
			// Do not create new session for the key
			return "", kvdb.ErrModified
		}
		// Destroy the existing session associated with the key
		_, err := kv.client.Session().Destroy(session, nil)
		if err != nil {
			return "", err
		}
	}

	durationTTL := time.Duration(int64(ttl)) * time.Second

	entry := &api.SessionEntry{
		Behavior:  api.SessionBehaviorDelete, // Delete the key when the session expires
		TTL:       durationTTL.String(),
		LockDelay: 0, // Virtually disable lock delay
	}

	// Create the key session
	session, _, err = kv.client.Session().Create(entry, nil)
	if err != nil {
		return "", err
	}

	// Session timer is started after a call to "Renew"
	_, _, err = kv.client.Session().Renew(session, nil)
	return session, err
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

func (kv *consulKV) SnapPut(snapKvp *kvdb.KVPair) (*kvdb.KVPair, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *consulKV) AddUser(username string, password string) error {
	return kvdb.ErrNotSupported
}

func (kv *consulKV) RemoveUser(username string) error {
	return kvdb.ErrNotSupported
}

func (kv *consulKV) GrantUserAccess(
	username string,
	permType kvdb.PermissionType,
	subtree string,
) error {
	return kvdb.ErrNotSupported
}

func (kv *consulKV) RevokeUsersAccess(
	username string,
	permType kvdb.PermissionType,
	subtree string,
) error {
	return kvdb.ErrNotSupported
}

func (kv *consulKV) Serialize() ([]byte, error) {

	kvps, err := kv.Enumerate("")
	if err != nil {
		return nil, err
	}
	return kv.SerializeAll(kvps)
}

func (kv *consulKV) Deserialize(b []byte) (kvdb.KVPairs, error) {
	return kv.DeserializeAll(b)
}
