package mem

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/common"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"fmt"
)

const (
	// Name is the name of this kvdb implementation.
	Name = "kv-mem"
	// KvSnap is an option passed to designate this kvdb as a snap.
	KvSnap = "KvSnap"
	bootstrapKey = "bootstrap"
)

var (
	// ErrSnap is returned if an operation is not supported on a snap.
	ErrSnap = errors.New("Operation not supported on snap.")
)

func init() {
	if err := kvdb.Register(Name, New, Version); err != nil {
		panic(err.Error())
	}
}

type memKV struct {
	common.BaseKvdb
	m      map[string]*kvdb.KVPair
	w      map[string]*watchData
	wt     map[string]*watchData
	mutex  sync.Mutex
	index  uint64
	domain string
}

type snapMem struct {
	*memKV
}

type watchData struct {
	cb        kvdb.WatchCB
	opaque    interface{}
	waitIndex uint64
}

// New constructs a new kvdb.Kvdb.
func New(
	domain string,
	machines []string,
	options map[string]string,
	fatalErrorCb kvdb.FatalErrorCB,
) (kvdb.Kvdb, error) {
	if domain != "" && !strings.HasSuffix(domain, "/") {
		domain = domain + "/"
	}

	mem := &memKV{
		BaseKvdb: common.BaseKvdb{FatalCb: fatalErrorCb},
		m:        make(map[string]*kvdb.KVPair),
		w:        make(map[string]*watchData),
		wt:       make(map[string]*watchData),
		domain:   domain,
	}

	if _, ok := options[KvSnap]; ok {
		return &snapMem{memKV: mem}, nil
	}
	return mem, nil

}

// Version returns the supported version of the mem implementation
func Version(url string, kvdbOptions map[string]string) (string, error) {
	return kvdb.MemVersion1, nil
}

func (kv *memKV) String() string {
	return Name
}

func (kv *memKV) Capabilities() int {
	return kvdb.KVCapabilityOrderedUpdates
}

func (kv *memKV) Get(key string) (*kvdb.KVPair, error) {

	key = kv.domain + key
	v, ok := kv.m[key]
	if !ok {
		return nil, kvdb.ErrNotFound
	}
	return v, nil
}

func (kv *memKV) Snapshot(prefix string) (kvdb.Kvdb, uint64, error) {
	_, err := kv.Put(bootstrapKey, time.Now().UnixNano(), 0)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create snap bootstrap key: %v", err)
	}
	kv.mutex.Lock()
	data := make(map[string]*kvdb.KVPair)
	for key, value := range kv.m {
		if !strings.HasPrefix(key, prefix) && strings.Contains(key, "/_") {
			continue
		}
		snap := &kvdb.KVPair{}
		*snap = *value
		snap.Value = make([]byte, len(value.Value))
		copy(snap.Value, value.Value)
		data[key] = snap
	}
	highestKvPair, _ := kv.Delete(bootstrapKey)
	kv.mutex.Unlock()
	// only snapshot data, watches are not copied
	return &memKV{
		m:      data,
		w:      make(map[string]*watchData),
		wt:     make(map[string]*watchData),
		domain: kv.domain,
	}, highestKvPair.ModifiedIndex, nil

}

func (kv *memKV) Put(
	key string,
	value interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {

	var kvp *kvdb.KVPair

	suffix := key
	key = kv.domain + suffix
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	index := atomic.AddUint64(&kv.index, 1)
	if ttl != 0 {
		time.AfterFunc(time.Second*time.Duration(ttl), func() {
			// TODO: handle error
			_, _ = kv.Delete(suffix)
		})
	}
	b, err := common.ToBytes(value)
	if err != nil {
		return nil, err
	}
	if old, ok := kv.m[key]; ok {
		old.Value = b
		old.Action = kvdb.KVSet
		old.ModifiedIndex = index
		old.KVDBIndex = index
		kvp = old

	} else {
		kvp = &kvdb.KVPair{
			Key:           key,
			Value:         b,
			TTL:           int64(ttl),
			KVDBIndex:     index,
			ModifiedIndex: index,
			CreatedIndex:  index,
			Action:        kvdb.KVCreate,
		}
		kv.m[key] = kvp
	}

	kv.normalize(kvp)
	go kv.fireCB(key, *kvp, nil)
	return kvp, nil
}

func (kv *memKV) GetVal(key string, v interface{}) (*kvdb.KVPair, error) {
	kvp, err := kv.Get(key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(kvp.Value, v)
	return kvp, err
}

func (kv *memKV) Create(
	key string,
	value interface{},
	ttl uint64) (*kvdb.KVPair, error) {

	result, err := kv.Get(key)
	if err != nil {
		return kv.Put(key, value, ttl)
	}
	return result, kvdb.ErrExist
}

func (kv *memKV) Update(
	key string,
	value interface{},
	ttl uint64) (*kvdb.KVPair, error) {

	if _, err := kv.Get(key); err != nil {
		return nil, kvdb.ErrNotFound
	}
	return kv.Put(key, value, ttl)
}

func (kv *memKV) Enumerate(prefix string) (kvdb.KVPairs, error) {
	var kvp = make(kvdb.KVPairs, 0, 100)
	prefix = kv.domain + prefix

	for k, v := range kv.m {
		if strings.HasPrefix(k, prefix) && !strings.Contains(k, "/_") {
			kvpLocal := *v
			kv.normalize(&kvpLocal)
			kvp = append(kvp, &kvpLocal)
		}
	}

	return kvp, nil
}

func (kv *memKV) Delete(key string) (*kvdb.KVPair, error) {
	kvp, err := kv.Get(key)
	if err != nil {
		return nil, err
	}
	kvp.KVDBIndex = atomic.AddUint64(&kv.index, 1)
	kvp.ModifiedIndex = kvp.KVDBIndex
	kvp.Action = kvdb.KVDelete
	go kv.fireCB(kv.domain+key, *kvp, nil)
	delete(kv.m, kv.domain+key)
	return kvp, nil
}

func (kv *memKV) DeleteTree(prefix string) error {

	kvp, err := kv.Enumerate(prefix)
	if err != nil {
		return err
	}
	for _, v := range kvp {
		// TODO: multiple errors
		if _, iErr := kv.Delete(v.Key); iErr != nil {
			err = iErr
		}
	}
	return err
}

func (kv *memKV) Keys(prefix, key string) ([]string, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *memKV) CompareAndSet(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
	prevValue []byte) (*kvdb.KVPair, error) {

	kv.mutex.Lock()

	result, err := kv.Get(kvp.Key)
	if err != nil {
		kv.mutex.Unlock()
		return nil, err
	}
	if prevValue != nil {
		if !bytes.Equal(result.Value, prevValue) {
			kv.mutex.Unlock()
			return nil, kvdb.ErrValueMismatch
		}
	}
	if flags == kvdb.KVModifiedIndex {
		if kvp.ModifiedIndex != result.ModifiedIndex {
			kv.mutex.Unlock()
			return nil, kvdb.ErrValueMismatch
		}
	}
	kv.mutex.Unlock()
	return kv.Put(kvp.Key, kvp.Value, 0)
}

func (kv *memKV) CompareAndDelete(kvp *kvdb.KVPair, flags kvdb.KVFlags) (*kvdb.KVPair, error) {
	if flags != kvdb.KVFlags(0) {
		return nil, kvdb.ErrNotSupported
	}
	if result, err := kv.Get(kvp.Key); err != nil {
		return nil, err
	} else if !bytes.Equal(result.Value, kvp.Value) {
		return nil, kvdb.ErrNotFound
	}
	return kv.Delete(kvp.Key)
}

func (kv *memKV) WatchKey(
	key string,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	key = kv.domain + key
	if _, ok := kv.w[key]; ok {
		return kvdb.ErrExist
	}
	kv.w[key] = &watchData{cb: cb, waitIndex: waitIndex, opaque: opaque}
	return nil
}

func (kv *memKV) WatchTree(
	prefix string,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB) error {

	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	prefix = kv.domain + prefix
	if _, ok := kv.wt[prefix]; ok {
		return kvdb.ErrExist
	}
	kv.wt[prefix] = &watchData{cb: cb, waitIndex: waitIndex, opaque: opaque}
	return nil
}

func (kv *memKV) Lock(key string) (*kvdb.KVPair, error) {
	return kv.LockWithID(key, "locked")
}

func (kv *memKV) LockWithID(key string, lockerID string) (
	*kvdb.KVPair,
	error,
) {
	key = kv.domain + key
	duration := time.Second

	result, err := kv.Create(key, lockerID, uint64(duration*3))
	count := 0
	for err != nil {
		time.Sleep(duration)
		result, err = kv.Create(key, lockerID, uint64(duration*3))
		if err != nil && count > 0 && count%15 == 0 {
			var currLockerID string
			if _, errGet := kv.GetVal(key, currLockerID); errGet == nil {
				logrus.Infof("Lock %v locked for %v seconds, tag: %v",
					key, count, currLockerID)
			}
		}
	}

	if err != nil {
		return nil, err
	}
	return result, err
}

func (kv *memKV) Unlock(kvp *kvdb.KVPair) error {
	_, err := kv.CompareAndDelete(kvp, kvdb.KVFlags(0))
	return err
}

func (kv *memKV) TxNew() (kvdb.Tx, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *memKV) normalize(kvp *kvdb.KVPair) {
	kvp.Key = strings.TrimPrefix(kvp.Key, kv.domain)
}

func (kv *memKV) fireCB(key string, kvp kvdb.KVPair, err error) {
	for k, v := range kv.w {
		if k == key && (v.waitIndex == 0 || v.waitIndex < kvp.ModifiedIndex) {
			err := v.cb(key, v.opaque, &kvp, err)
			if err != nil {
				// TODO: handle error
				_ = v.cb("", v.opaque, nil, kvdb.ErrWatchStopped)
				delete(kv.w, key)

			}
			return
		}
	}
	for k, v := range kv.wt {
		if strings.HasPrefix(key, k) &&
			(v.waitIndex == 0 || v.waitIndex < kvp.ModifiedIndex) {
			err := v.cb(key, v.opaque, &kvp, err)
			if err != nil {
				// TODO: handle error
				_ = v.cb("", v.opaque, nil, kvdb.ErrWatchStopped)
				delete(kv.wt, key)
			}
		}
	}
}

func (kv *memKV) SnapPut(snapKvp *kvdb.KVPair) (*kvdb.KVPair, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *snapMem) SnapPut(snapKvp *kvdb.KVPair) (*kvdb.KVPair, error) {
	var kvp *kvdb.KVPair

	key := kv.domain + snapKvp.Key
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if old, ok := kv.m[key]; ok {
		old.Value = snapKvp.Value
		old.Action = kvdb.KVSet
		old.ModifiedIndex = snapKvp.ModifiedIndex
		old.KVDBIndex = snapKvp.KVDBIndex
		kvp = old

	} else {
		kvp = &kvdb.KVPair{
			Key:           key,
			Value:         snapKvp.Value,
			TTL:           0,
			KVDBIndex:     snapKvp.KVDBIndex,
			ModifiedIndex: snapKvp.ModifiedIndex,
			CreatedIndex:  snapKvp.CreatedIndex,
			Action:        kvdb.KVCreate,
		}
		kv.m[key] = kvp
	}

	kv.normalize(kvp)
	return kvp, nil
}

func (kv *snapMem) Put(
	key string,
	value interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
	return nil, ErrSnap

}

func (kv *snapMem) Create(
	key string,
	value interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
	return nil, ErrSnap
}

func (kv *snapMem) Update(
	key string,
	value interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
	return nil, ErrSnap
}

func (kv *snapMem) Delete(key string) (*kvdb.KVPair, error) {
	return nil, ErrSnap
}

func (kv *snapMem) DeleteTree(prefix string) error {
	return ErrSnap
}

func (kv *snapMem) CompareAndSet(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
	prevValue []byte,
) (*kvdb.KVPair, error) {
	return nil, ErrSnap
}

func (kv *snapMem) CompareAndDelete(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
) (*kvdb.KVPair, error) {
	return nil, ErrSnap
}

func (kv *snapMem) WatchKey(
	key string,
	waitIndex uint64,
	opaque interface{},
	watchCB kvdb.WatchCB,
) error {
	return ErrSnap
}

func (kv *snapMem) WatchTree(
	prefix string,
	waitIndex uint64,
	opaque interface{},
	watchCB kvdb.WatchCB,
) error {
	return ErrSnap
}

func (kv *memKV) AddUser(username string, password string) error {
	return kvdb.ErrNotSupported
}

func (kv *memKV) RemoveUser(username string) error {
	return kvdb.ErrNotSupported
}

func (kv *memKV) GrantUserAccess(username string, permType kvdb.PermissionType, subtree string) error {
	return kvdb.ErrNotSupported
}

func (kv *memKV) RevokeUsersAccess(username string, permType kvdb.PermissionType, subtree string) error {
	return kvdb.ErrNotSupported
}
