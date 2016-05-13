// Package consul implements the KVDB interface based on consul.
// Code from docker/libkv was leveraged to build parts of this module.
package consul

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/common"
	"github.com/portworx/kvdb/mem"
)

const (
	// Name is the name of this kvdb implementation.
	Name = "consul-kv"
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
	key = path.Join(kv.domain, key)
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
	key = path.Join(kv.domain, key)
	b, err := common.ToBytes(val)
	if err != nil {
		return nil, err
	}
	pair := &api.KVPair{
		Key:   key,
		Value: b,
	}
	if _, err := kv.client.KV().Put(pair, nil); err != nil {
		return nil, err
	}
	return kv.pairToKv("put", pair, nil), nil
}

func (kv *consulKV) Create(key string, val interface{}, ttl uint64) (*kvdb.KVPair, error) {
	if _, err := kv.Get(key); err == nil {
		return nil, kvdb.ErrExist
	}
	key = path.Join(kv.domain, key)
	b, err := common.ToBytes(val)
	if err != nil {
		return nil, err
	}
	pair := &api.KVPair{
		Key:   key,
		Value: b,
	}
	if _, err := kv.client.KV().Put(pair, nil); err != nil {
		return nil, err
	}
	return kv.pairToKv("create", pair, nil), nil
}

func (kv *consulKV) Update(key string, val interface{}, ttl uint64) (*kvdb.KVPair, error) {
	if _, err := kv.Get(key); err != nil {
		return nil, err
	}
	key = path.Join(kv.domain, key)
	b, err := common.ToBytes(val)
	if err != nil {
		return nil, err
	}
	pair := &api.KVPair{
		Key:   key,
		Value: b,
	}
	if _, err := kv.client.KV().Put(pair, nil); err != nil {
		return nil, err
	}
	return kv.pairToKv("update", pair, nil), nil
}

func (kv *consulKV) Enumerate(prefix string) (kvdb.KVPairs, error) {
	prefix = path.Join(kv.domain, prefix)
	pairs, _, err := kv.client.KV().List(prefix, nil)
	if err != nil {
		return nil, err
	}
	return kv.pairToKvs("enumerate", pairs, nil), nil
}

func (kv *consulKV) Delete(key string) (*kvdb.KVPair, error) {
	pair, err := kv.Get(key)
	if err != nil {
		return nil, err
	}
	key = path.Join(kv.domain, key)
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
	return nil, kvdb.ErrNotSupported
}

func (kv *consulKV) CompareAndDelete(kvp *kvdb.KVPair, flags kvdb.KVFlags) (*kvdb.KVPair, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *consulKV) WatchKey(key string, waitIndex uint64, opaque interface{}, cb kvdb.WatchCB) error {
	return kvdb.ErrNotSupported
}

func (kv *consulKV) WatchTree(prefix string, waitIndex uint64, opaque interface{}, cb kvdb.WatchCB) error {
	return kvdb.ErrNotSupported
}

func (kv *consulKV) Lock(key string, ttl uint64) (*kvdb.KVPair, error) {
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

func (kv *consulKV) createKv(pair *api.KVPair) *kvdb.KVPair {
	return &kvdb.KVPair{
		Key:   strings.TrimPrefix(pair.Key, kv.domain),
		Value: []byte(pair.Value),
	}
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
	key = path.Join(kv.domain, key)
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

func (kv *consulKV) Snapshot(prefix string) (kvdb.Kvdb, uint64, error) {
	options := &api.QueryOptions{
		AllowStale:        false,
		RequireConsistent: true,
	}
	pairs, _, err := kv.client.KV().List(path.Join(kv.domain, prefix), options)
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

	return snapDb, 0, nil
}
