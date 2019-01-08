package common

import (
	"container/list"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/portworx/kvdb"
	"github.com/sirupsen/logrus"
)

var (
	path = "/var/cores/"
)

// ToBytes converts to value to a byte slice.
func ToBytes(val interface{}) ([]byte, error) {
	switch val.(type) {
	case string:
		return []byte(val.(string)), nil
	case []byte:
		b := make([]byte, len(val.([]byte)))
		copy(b, val.([]byte))
		return b, nil
	default:
		return json.Marshal(val)
	}
}

// BaseKvdb provides common functionality across kvdb types
type BaseKvdb struct {
	// LockTimeout is the maximum time any lock can be held
	LockTimeout time.Duration
	// FatalCb invoked for fatal errors
	FatalCb kvdb.FatalErrorCB
	// lock to guard updates to timeout and fatalCb
	lock sync.Mutex
}

// SetFatalCb callback is invoked when an unrecoverable KVDB error happens.
func (b *BaseKvdb) SetFatalCb(f kvdb.FatalErrorCB) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.FatalCb = f
}

// SetLockTimeout has property such that if the lock is held past this duration,
// then a configured fatal callback is called.
func (b *BaseKvdb) SetLockTimeout(timeout time.Duration) {
	b.lock.Lock()
	defer b.lock.Unlock()
	logrus.Infof("Setting lock timeout to: %v", timeout)
	b.LockTimeout = timeout
}

// CheckLockTimeout checks lock timeout.
func (b *BaseKvdb) CheckLockTimeout(
	key string,
	startTime time.Time,
	lockTimeout time.Duration,
) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.LockTimeout > 0 && time.Since(startTime) > lockTimeout {
		b.lockTimedout(key)
	}
}

// GetLockTimeout gets lock timeout.
func (b *BaseKvdb) GetLockTimeout() time.Duration {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.LockTimeout
}

// LockTimedout does lock timedout.
func (b *BaseKvdb) LockTimedout(key string) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.lockTimedout(key)
}

// lockTimedout function is invoked if lock is held past configured timeout.
func (b *BaseKvdb) lockTimedout(key string) {
	b.FatalCb("Lock %s hold timeout triggered", key)
}

// SerializeAll Serializes all key value pairs to a byte array.
func (b *BaseKvdb) SerializeAll(kvps kvdb.KVPairs) ([]byte, error) {
	out, err := json.Marshal(kvps)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DeserializeAll Unmarshals a byte stream created from serializeAll into the kvdb tree.
func (b *BaseKvdb) DeserializeAll(out []byte) (kvdb.KVPairs, error) {
	var kvps kvdb.KVPairs
	if err := json.Unmarshal(out, &kvps); err != nil {
		return nil, err
	}
	return kvps, nil
}

// watchUpdate refers to an update to this kvdb
type watchUpdate struct {
	// key is the key that was updated
	key string
	// kvp is the key-value that was updated
	kvp *kvdb.KVPair
	// err is any error on update
	err error
}

// WatchUpdateQueue is a producer consumer queue.
type WatchUpdateQueue interface {
	// Enqueue will enqueue an update. It is non-blocking.
	Enqueue(key string, kvp *kvdb.KVPair, err error)
	// Dequeue will either return an element from front of the queue or
	// will block until element becomes available
	Dequeue() (string, *kvdb.KVPair, error)
}

// watchQueue implements WatchUpdateQueue interface for watchUpdates
type watchQueue struct {
	// updates is the list of updates
	updates *list.List
	// m is the mutex to protect updates
	m *sync.Mutex
	// cv is used to coordinate the producer-consumer threads
	cv *sync.Cond
}

// NewWatchUpdateQueue returns WatchUpdateQueue
func NewWatchUpdateQueue() WatchUpdateQueue {
	mtx := &sync.Mutex{}
	return &watchQueue{
		m:       mtx,
		cv:      sync.NewCond(mtx),
		updates: list.New()}
}

// Dequeue removes from queue.
func (w *watchQueue) Dequeue() (string, *kvdb.KVPair, error) {
	w.m.Lock()
	for {
		if w.updates.Len() > 0 {
			el := w.updates.Front()
			w.updates.Remove(el)
			w.m.Unlock()
			update := el.Value.(*watchUpdate)
			return update.key, update.kvp, update.err
		}
		w.cv.Wait()
	}
}

// Enqueue enqueues and never blocks
func (w *watchQueue) Enqueue(key string, kvp *kvdb.KVPair, err error) {
	w.m.Lock()
	w.updates.PushBack(&watchUpdate{key: key, kvp: kvp, err: err})
	w.cv.Signal()
	w.m.Unlock()
}

// PrunePrefixes will return all the top level prefixes from a given list
// so that any enumerate on these prefixes will not end up returning duplicate keys
func PrunePrefixes(prefixes []string) []string {
	prunedPrefixes := []string{}
	for i := 0; i < len(prefixes); i++ {
		foundPrefix := false
		for j := 0; j < len(prefixes); j++ {
			if i == j {
				continue
			}
			if strings.HasPrefix(prefixes[i], prefixes[j]) {
				foundPrefix = true
			}
		}
		if !foundPrefix {
			prunedPrefixes = append(prunedPrefixes, prefixes[i])
		}
	}
	return prunedPrefixes
}
