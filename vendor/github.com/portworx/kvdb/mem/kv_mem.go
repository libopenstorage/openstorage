package mem

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/common"
)

const (
	// Name is the name of this kvdb implementation.
	Name = "kv-mem"
	// KvSnap is an option passed to designate this kvdb as a snap.
	KvSnap = "KvSnap"
	// KvUseInterface is an option passed that configures the mem to store
	// the values as interfaces instead of bytes. It will not create a
	// copy of the interface that is passed in. USE WITH CAUTION
	KvUseInterface = "KvUseInterface"
	bootstrapKey   = "bootstrap"
)

var (
	// ErrSnap is returned if an operation is not supported on a snap.
	ErrSnap = errors.New("operation not supported on snap")
	// ErrSnapWithInterfaceNotSupported is returned when a snap kv-mem is
	// created with KvUseInterface flag on
	ErrSnapWithInterfaceNotSupported = errors.New("snap kvdb not supported with interfaces")
	// ErrIllegalSelect is returned when an incorrect select function
	// implementation is detected.
	ErrIllegalSelect = errors.New("Illegal Select implementation")
)

func init() {
	if err := kvdb.Register(Name, New, Version); err != nil {
		panic(err.Error())
	}
}

type memKV struct {
	common.BaseKvdb
	// m is the key value database
	m map[string]*memKVPair
	// updates is the list of latest few updates
	dist WatchDistributor
	// mutex protects m, w, wt
	mutex sync.Mutex
	// index current kvdb index
	index  uint64
	domain string
	// locks is the map of currently held locks
	locks map[string]chan int
	// noByte will store all the values as interface
	noByte bool
	kvdb.Controller
}

type memKVPair struct {
	kvdb.KVPair
	// ivalue is the value for this kv pair stored as an interface
	ivalue interface{}
}

func (mkvp *memKVPair) copy() *kvdb.KVPair {
	copyKvp := mkvp.KVPair
	if mkvp.Value == nil && mkvp.ivalue != nil {
		copyKvp.Value, _ = common.ToBytes(mkvp.ivalue)
	}
	return &copyKvp
}

type snapMem struct {
	*memKV
}

// watchUpdate refers to an update to this kvdb
type watchUpdate struct {
	// key is the key that was updated
	key string
	// kvp is the key-value that was updated
	kvp memKVPair
	// err is any error on update
	err error
}

// WatchUpdateQueue is a producer consumer queue.
type WatchUpdateQueue interface {
	// Enqueue will enqueue an update. It is non-blocking.
	Enqueue(update *watchUpdate)
	// Dequeue will either return an element from front of the queue or
	// will block until element becomes available
	Dequeue() *watchUpdate
}

// WatchDistributor distributes updates to the watchers
type WatchDistributor interface {
	// Add creates a new watch queue to send updates
	Add() WatchUpdateQueue
	// Remove removes an existing watch queue
	Remove(WatchUpdateQueue)
	// NewUpdate is invoked to distribute a new update
	NewUpdate(w *watchUpdate)
}

// distributor implements WatchDistributor interface
type distributor struct {
	sync.Mutex
	// updates is the list of latest few updates
	updates []*watchUpdate
	// watchers watch for updates
	watchers []WatchUpdateQueue
}

// NewWatchDistributor returns a new instance of
// the WatchDistrubtor interface
func NewWatchDistributor() WatchDistributor {
	return &distributor{}
}

func (d *distributor) Add() WatchUpdateQueue {
	d.Lock()
	defer d.Unlock()
	q := NewWatchUpdateQueue()
	for _, u := range d.updates {
		q.Enqueue(u)
	}
	d.watchers = append(d.watchers, q)
	return q
}

func (d *distributor) Remove(r WatchUpdateQueue) {
	d.Lock()
	defer d.Unlock()
	for i, q := range d.watchers {
		if q == r {
			copy(d.watchers[i:], d.watchers[i+1:])
			d.watchers[len(d.watchers)-1] = nil
			d.watchers = d.watchers[:len(d.watchers)-1]
		}
	}
}

func (d *distributor) NewUpdate(u *watchUpdate) {
	d.Lock()
	defer d.Unlock()
	// collect update
	d.updates = append(d.updates, u)
	if len(d.updates) > 100 {
		d.updates = d.updates[100:]
	}
	// send update to watchers
	for _, q := range d.watchers {
		q.Enqueue(u)
	}
}

// watchQueue implements WatchUpdateQueue interface for watchUpdates
type watchQueue struct {
	// updates is the list of updates
	updates []*watchUpdate
	// m is the mutex to protect updates
	m *sync.Mutex
	// cv is used to coordinate the producer-consumer threads
	cv *sync.Cond
}

// NewWatchUpdateQueue returns an instance of WatchUpdateQueue
func NewWatchUpdateQueue() WatchUpdateQueue {
	mtx := &sync.Mutex{}
	return &watchQueue{
		m:       mtx,
		cv:      sync.NewCond(mtx),
		updates: make([]*watchUpdate, 0)}
}

func (w *watchQueue) Dequeue() *watchUpdate {
	w.m.Lock()
	for {
		if len(w.updates) > 0 {
			update := w.updates[0]
			w.updates = w.updates[1:]
			w.m.Unlock()
			return update
		}
		w.cv.Wait()
	}
}

// Enqueue enqueues and never blocks
func (w *watchQueue) Enqueue(update *watchUpdate) {
	w.m.Lock()
	w.updates = append(w.updates, update)
	w.cv.Signal()
	w.m.Unlock()
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
		BaseKvdb:   common.BaseKvdb{FatalCb: fatalErrorCb},
		m:          make(map[string]*memKVPair),
		dist:       NewWatchDistributor(),
		domain:     domain,
		Controller: kvdb.ControllerNotSupported,
		locks:      make(map[string]chan int),
	}

	var noByte bool
	if _, noByte = options[KvUseInterface]; noByte {
		mem.noByte = true
	}
	if _, ok := options[KvSnap]; ok && !noByte {
		return &snapMem{memKV: mem}, nil
	} else if ok && noByte {
		return nil, ErrSnapWithInterfaceNotSupported
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

func (kv *memKV) get(key string) (*memKVPair, error) {
	key = kv.domain + key
	v, ok := kv.m[key]
	if !ok {
		return nil, kvdb.ErrNotFound
	}
	return v, nil
}

func (kv *memKV) exists(key string) (*memKVPair, error) {
	return kv.get(key)
}

func (kv *memKV) Get(key string) (*kvdb.KVPair, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	v, err := kv.get(key)
	if err != nil {
		return nil, err
	}
	return v.copy(), nil
}

func (kv *memKV) Snapshot(prefix string) (kvdb.Kvdb, uint64, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	_, err := kv.put(bootstrapKey, time.Now().UnixNano(), 0)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create snap bootstrap key: %v", err)
	}
	data := make(map[string]*memKVPair)
	for key, value := range kv.m {
		if !strings.HasPrefix(key, prefix) && strings.Contains(key, "/_") {
			continue
		}
		snap := &memKVPair{}
		snap.KVPair = value.KVPair
		cpy := value.copy()
		snap.Value = make([]byte, len(cpy.Value))
		copy(snap.Value, cpy.Value)
		data[key] = snap
	}
	highestKvPair, _ := kv.delete(bootstrapKey)
	// Snapshot only data, watches are not copied.
	return &memKV{
		m:      data,
		domain: kv.domain,
	}, highestKvPair.ModifiedIndex, nil
}

func (kv *memKV) put(
	key string,
	value interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {

	var (
		kvp  *memKVPair
		b    []byte
		err  error
		ival interface{}
	)

	suffix := key
	key = kv.domain + suffix
	index := atomic.AddUint64(&kv.index, 1)

	// Either set bytes or interface value
	if !kv.noByte {
		b, err = common.ToBytes(value)
		if err != nil {
			return nil, err
		}
	} else {
		ival = value
	}
	if old, ok := kv.m[key]; ok {
		old.Value = b
		old.ivalue = ival
		old.Action = kvdb.KVSet
		old.ModifiedIndex = index
		old.KVDBIndex = index
		kvp = old
	} else {
		kvp = &memKVPair{
			KVPair: kvdb.KVPair{
				Key:           key,
				Value:         b,
				TTL:           int64(ttl),
				KVDBIndex:     index,
				ModifiedIndex: index,
				CreatedIndex:  index,
				Action:        kvdb.KVCreate,
			},
			ivalue: ival,
		}
		kv.m[key] = kvp
	}

	kv.normalize(&kvp.KVPair)
	kv.dist.NewUpdate(&watchUpdate{key, *kvp, nil})

	if ttl != 0 {
		time.AfterFunc(time.Second*time.Duration(ttl), func() {
			// TODO: handle error
			kv.mutex.Lock()
			defer kv.mutex.Unlock()
			_, _ = kv.delete(suffix)
		})
	}

	return kvp.copy(), nil
}

func (kv *memKV) Put(
	key string,
	value interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {

	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	return kv.put(key, value, ttl)
}

func (kv *memKV) GetVal(key string, v interface{}) (*kvdb.KVPair, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	kvp, err := kv.get(key)
	if err != nil {
		return nil, err
	}

	cpy := kvp.copy()
	err = json.Unmarshal(cpy.Value, v)
	return cpy, err
}

func (kv *memKV) Create(
	key string,
	value interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	result, err := kv.exists(key)
	if err != nil {
		return kv.put(key, value, ttl)
	}
	return &result.KVPair, kvdb.ErrExist
}

func (kv *memKV) Update(
	key string,
	value interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if _, err := kv.exists(key); err != nil {
		return nil, kvdb.ErrNotFound
	}
	return kv.put(key, value, ttl)
}

func (kv *memKV) Enumerate(prefix string) (kvdb.KVPairs, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	return kv.enumerate(prefix)
}

// enumerate returns a list of values and creates a copy if specified
func (kv *memKV) enumerate(prefix string) (kvdb.KVPairs, error) {
	var kvp = make(kvdb.KVPairs, 0, 100)
	prefix = kv.domain + prefix

	for k, v := range kv.m {
		if strings.HasPrefix(k, prefix) && !strings.Contains(k, "/_") {
			kvpLocal := v.copy()
			kv.normalize(kvpLocal)
			kvp = append(kvp, kvpLocal)
		}
	}

	return kvp, nil
}

func (kv *memKV) delete(key string) (*kvdb.KVPair, error) {
	kvp, err := kv.get(key)
	if err != nil {
		return nil, err
	}
	kvp.KVDBIndex = atomic.AddUint64(&kv.index, 1)
	kvp.ModifiedIndex = kvp.KVDBIndex
	kvp.Action = kvdb.KVDelete
	delete(kv.m, kv.domain+key)
	kv.dist.NewUpdate(&watchUpdate{kv.domain + key, *kvp, nil})
	return &kvp.KVPair, nil
}

func (kv *memKV) Delete(key string) (*kvdb.KVPair, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	return kv.delete(key)
}

func (kv *memKV) DeleteTree(prefix string) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	kvp, err := kv.enumerate(prefix)
	if err != nil {
		return err
	}
	for _, v := range kvp {
		// TODO: multiple errors
		if _, iErr := kv.delete(v.Key); iErr != nil {
			err = iErr
		}
	}
	return err
}

func (kv *memKV) Keys(prefix, sep string) ([]string, error) {
	if "" == sep {
		sep = "/"
	}
	prefix = kv.domain + prefix
	lenPrefix := len(prefix)
	lenSep := len(sep)
	if prefix[lenPrefix-lenSep:] != sep {
		prefix += sep
		lenPrefix += lenSep
	}
	seen := make(map[string]bool)
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	for k := range kv.m {
		if strings.HasPrefix(k, prefix) && !strings.Contains(k, "/_") {
			key := k[lenPrefix:]
			if idx := strings.Index(key, sep); idx > 0 {
				key = key[:idx]
			}
			seen[key] = true
		}
	}
	retList := make([]string, len(seen))
	i := 0
	for k := range seen {
		retList[i] = k
		i++
	}

	return retList, nil
}

func (kv *memKV) CompareAndSet(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
	prevValue []byte,
) (*kvdb.KVPair, error) {

	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	result, err := kv.exists(kvp.Key)
	if err != nil {
		return nil, err
	}
	if prevValue != nil {
		cpy := result.copy()
		if !bytes.Equal(cpy.Value, prevValue) {
			return nil, kvdb.ErrValueMismatch
		}
	}
	if flags == kvdb.KVModifiedIndex {
		if kvp.ModifiedIndex != result.ModifiedIndex {
			return nil, kvdb.ErrValueMismatch
		}
	}
	return kv.put(kvp.Key, kvp.Value, 0)
}

func (kv *memKV) CompareAndDelete(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
) (*kvdb.KVPair, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if flags != kvdb.KVFlags(0) {
		return nil, kvdb.ErrNotSupported
	}
	result, err := kv.exists(kvp.Key)
	if err != nil {
		return nil, err
	}
	cpy := result.copy()
	if !bytes.Equal(cpy.Value, kvp.Value) {
		return nil, kvdb.ErrNotFound
	}
	return kv.delete(kvp.Key)
}

func (kv *memKV) WatchKey(
	key string,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	key = kv.domain + key
	go kv.watchCb(kv.dist.Add(), key,
		&watchData{cb: cb, waitIndex: waitIndex, opaque: opaque},
		false)
	return nil
}

func (kv *memKV) WatchTree(
	prefix string,
	waitIndex uint64,
	opaque interface{},
	cb kvdb.WatchCB,
) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	prefix = kv.domain + prefix
	go kv.watchCb(kv.dist.Add(), prefix,
		&watchData{cb: cb, waitIndex: waitIndex, opaque: opaque},
		true)
	return nil
}

func (kv *memKV) Lock(key string) (*kvdb.KVPair, error) {
	return kv.LockWithID(key, "locked")
}

func (kv *memKV) LockWithID(
	key string,
	lockerID string,
) (*kvdb.KVPair, error) {
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

	lockChan := make(chan int)
	kv.mutex.Lock()
	kv.locks[key] = lockChan
	kv.mutex.Unlock()
	if kv.GetLockTimeout() > 0 {
		go func() {
			timeout := time.After(kv.GetLockTimeout())
			for {
				select {
				case <-timeout:
					kv.LockTimedout(key)
				case <-lockChan:
					return
				}
			}
		}()
	}

	return result, err
}

func (kv *memKV) Unlock(kvp *kvdb.KVPair) error {
	kv.mutex.Lock()
	lockChan, ok := kv.locks[kvp.Key]
	if ok {
		delete(kv.locks, kvp.Key)
	}
	kv.mutex.Unlock()
	if lockChan != nil {
		close(lockChan)
	}
	_, err := kv.CompareAndDelete(kvp, kvdb.KVFlags(0))
	return err
}

func (kv *memKV) EnumerateWithSelect(
	prefix string,
	enumerateSelect kvdb.EnumerateSelect,
	copySelect kvdb.CopySelect,
) ([]interface{}, error) {
	if enumerateSelect == nil || copySelect == nil {
		return nil, ErrIllegalSelect
	}
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	var kvi []interface{}
	prefix = kv.domain + prefix
	for k, v := range kv.m {
		if strings.HasPrefix(k, prefix) && !strings.Contains(k, "/_") {
			if enumerateSelect(v.ivalue) {
				cpy := copySelect(v.ivalue)
				if cpy == nil {
					return nil, ErrIllegalSelect
				}
				kvi = append(kvi, cpy)
			}
		}
	}
	return kvi, nil
}

func (kv *memKV) GetWithCopy(
	key string,
	copySelect kvdb.CopySelect,
) (interface{}, error) {
	if copySelect == nil {
		return nil, ErrIllegalSelect
	}
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	kvp, err := kv.get(key)
	if err != nil {
		return nil, err
	}
	return copySelect(kvp.ivalue), nil
}

func (kv *memKV) TxNew() (kvdb.Tx, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *memKV) normalize(kvp *kvdb.KVPair) {
	kvp.Key = strings.TrimPrefix(kvp.Key, kv.domain)
}

func copyWatchKeys(w map[string]*watchData) []string {
	keys := make([]string, len(w))
	i := 0
	for key := range w {
		keys[i] = key
		i++
	}
	return keys
}

func (kv *memKV) watchCb(
	q WatchUpdateQueue,
	prefix string,
	v *watchData,
	treeWatch bool,
) {
	for {
		update := q.Dequeue()
		if ((treeWatch && strings.HasPrefix(update.key, prefix)) ||
			(!treeWatch && update.key == prefix)) &&
			(v.waitIndex == 0 || v.waitIndex < update.kvp.ModifiedIndex) {
			kvpCopy := update.kvp.copy()
			err := v.cb(update.key, v.opaque, kvpCopy, update.err)
			if err != nil {
				_ = v.cb("", v.opaque, nil, kvdb.ErrWatchStopped)
				kv.dist.Remove(q)
				return
			}
		}
	}
}

func (kv *memKV) SnapPut(snapKvp *kvdb.KVPair) (*kvdb.KVPair, error) {
	return nil, kvdb.ErrNotSupported
}

func (kv *snapMem) SnapPut(snapKvp *kvdb.KVPair) (*kvdb.KVPair, error) {
	var kvp *memKVPair

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
		kvp = &memKVPair{
			KVPair: kvdb.KVPair{
				Key:           key,
				Value:         snapKvp.Value,
				TTL:           0,
				KVDBIndex:     snapKvp.KVDBIndex,
				ModifiedIndex: snapKvp.ModifiedIndex,
				CreatedIndex:  snapKvp.CreatedIndex,
				Action:        kvdb.KVCreate,
			},
		}
		kv.m[key] = kvp
	}

	kv.normalize(&kvp.KVPair)
	return &kvp.KVPair, nil
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

func (kv *snapMem) Delete(snapKey string) (*kvdb.KVPair, error) {
	key := kv.domain + snapKey
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	kvp, ok := kv.m[key]
	if !ok {
		return nil, kvdb.ErrNotFound
	}
	kvPair := kvp.KVPair
	delete(kv.m, key)
	return &kvPair, nil
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

func (kv *memKV) GrantUserAccess(
	username string,
	permType kvdb.PermissionType,
	subtree string,
) error {
	return kvdb.ErrNotSupported
}

func (kv *memKV) RevokeUsersAccess(
	username string,
	permType kvdb.PermissionType,
	subtree string,
) error {
	return kvdb.ErrNotSupported
}

func (kv *memKV) Serialize() ([]byte, error) {

	kvps, err := kv.Enumerate("")
	if err != nil {
		return nil, err
	}
	return kv.SerializeAll(kvps)
}

func (kv *memKV) Deserialize(b []byte) (kvdb.KVPairs, error) {
	return kv.DeserializeAll(b)
}
