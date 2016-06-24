package kvdb

import (
	"errors"
)

const (
	// KVSet signifies the KV was modified.
	KVSet KVAction = 1 << iota
	// KVCreate set if the KV pair was created.
	KVCreate
	// KVGet set when the key is fetched from the KV store
	KVGet
	// KVDelete set when the key is deleted from the KV store
	KVDelete
	// KVExpire set when the key expires
	KVExpire
	// KVUknown operation on KV pair
	KVUknown
)

const (
	// KVPrevExists flag to check key already exists
	KVPrevExists KVFlags = 1 << iota
	// KVCreatedIndex flag compares with passed in index (possibly in KVPair)
	KVCreatedIndex
	// KVModifiedIndex flag compares with passed in index (possibly in KVPair)
	KVModifiedIndex
	// KVTTL uses TTL val from KVPair.
	KVTTL
)

var (
	// ErrNotSupported implemenation of a specific function is not supported.
	ErrNotSupported = errors.New("implementation not supported")
	// ErrWatchStopped is raised when user stops watch.
	ErrWatchStopped = errors.New("Watch Stopped")
	// ErrNotFound raised if Key is not found
	ErrNotFound = errors.New("Key not found")
	// ErrExist raised if key already exists
	ErrExist = errors.New("Key already exists")
	// ErrUnmarshal raised if Get fails to unmarshal value.
	ErrUnmarshal = errors.New("Failed to unmarshal value")
	// ErrIllegal raised if object is not valid.
	ErrIllegal = errors.New("Illegal operation")
	// ErrValueMismatch raised if existing KVDB value mismatches with user provided value
	ErrValueMismatch = errors.New("Value mismatch")
	// ErrModified raised during an atomic operation if the index does not match the one in the store
	ErrModified = errors.New("Key Index mismatch")
	// ErrSetTTLFailed raised if unable to set ttl value for a key create/put/update action
	ErrSetTTLFailed = errors.New("Unable to set ttl value")
	// ErrTTLNotSupported if kvdb implementation doesn't support TTL
	ErrTTLNotSupported = errors.New("TTL value not supported")
	// ErrInvalidLock Lock and unlock operations don't match.
	ErrInvalidLock = errors.New("Invalid lock/unlock operation")
)

// KVAction specifies the action on a KV pair. This is useful to make decisions
// from the results of  a Watch.
type KVAction int

// KVFlags options for operations on KVDB
type KVFlags uint64

// WatchCB is called when a watched key or tree is modified. If the callback
// returns an error, then watch stops and the cb is called one last time
// with ErrWatchStopped.
type WatchCB func(prefix string, opaque interface{}, kvp *KVPair, err error) error

// DatastoreInit is called to activate a backend KV store.
type DatastoreInit func(domain string, machines []string, options map[string]string) (Kvdb, error)

// KVPair represents the results of an operation on KVDB.
type KVPair struct {
	// Key for this kv pair.
	Key string
	// Value for this kv pair
	Value []byte
	// Action the last action on this KVPair.
	Action KVAction
	// TTL value after which this key will expire from KVDB
	TTL int64
	// KVDBIndex A Monotonically index updated at each modification operation.
	KVDBIndex uint64
	// CreatedIndex for this kv pair
	CreatedIndex uint64
	// ModifiedIndex for this kv pair
	ModifiedIndex uint64
	// Lock is a generic interface to represent a lock held on a key.
	Lock interface{}
}

// KVPairs list of KVPairs
type KVPairs []*KVPair

// Tx Interface to transactionally apply updates to a set of keys.
type Tx interface {
	// Put specified key value pair in TX.
	Put(key string, value interface{}, ttl uint64) (*KVPair, error)
	// Get returns KVPair in this TXs view. If not found, returns value from
	// backing KVDB.
	Get(key string) (*KVPair, error)
	// Get same as get except that value has the unmarshalled value.
	GetVal(key string, value interface{}) (*KVPair, error)
	// Prepare returns an error it transaction cannot be logged.
	Prepare() error
	// Commit propagates updates to the KVDB. No operations on this Tx are
	// allowed after commit.
	Commit() error
	// Abort aborts this transaction.  No operations on this Tx are allowed
	// afer commit.
	Abort() error
}

// Kvdb interface implemented by backing datastores.
type Kvdb interface {
	// String representation of backend datastore.
	String() string
	// Get returns KVPair that maps to specified key or ErrNotFound.
	Get(key string) (*KVPair, error)
	// Get returns KVPair that maps to specified key or ErrNotFound. If found
	// value contains the unmarshalled result or error is ErrUnmarshal
	GetVal(key string, value interface{}) (*KVPair, error)
	// Put inserts value at key in kvdb. If value is a runtime.Object, it is
	// marshalled. If Value is []byte it is set directly. If Value is a string,
	// its byte representation is stored.
	Put(key string, value interface{}, ttl uint64) (*KVPair, error)
	// Create is the same as Put except that ErrExist is returned if the key exists.
	Create(key string, value interface{}, ttl uint64) (*KVPair, error)
	// Update is the same as Put except that ErrNotFound is returned if the key
	// does not exist.
	Update(key string, value interface{}, ttl uint64) (*KVPair, error)
	// Enumerate returns a list of KVPair for all keys that share the specified prefix.
	Enumerate(prefix string) (KVPairs, error)
	// Delete deletes the KVPair specified by the key. ErrNotFound is returned
	// if the key is not found. The old KVPair is returned if successful.
	Delete(key string) (*KVPair, error)
	// DeleteTree same as Delete execpt that all keys sharing the prefix are
	// deleted.
	DeleteTree(prefix string) error
	// Keys returns an array of keys that share specified prefix.
	Keys(prefix, key string) ([]string, error)
	// CompareAndSet updates value at kvp.Key if the previous resident
	// satisfies conditions set in flags and optional prevValue.
	CompareAndSet(kvp *KVPair, flags KVFlags, prevValue []byte) (*KVPair, error)
	// CompareAndDelete deletes value at kvp.Key if the previous resident matches
	// satisfies conditions set in flags.
	CompareAndDelete(kvp *KVPair, flags KVFlags) (*KVPair, error)
	// WatchKey calls watchCB everytime a value at key is updated. waitIndex
	// is the oldest ModifiedIndex of a KVPair for which updates are requestd.
	WatchKey(key string, waitIndex uint64, opaque interface{}, watchCB WatchCB) error
	// WatchTree is the same as WatchKey except that watchCB is triggered
	// for updates on all keys that share the prefix.
	WatchTree(prefix string, waitIndex uint64, opaque interface{}, watchCB WatchCB) error
	// Lock specfied key. The KVPair returned should be used to unlock.
	Lock(key string) (*KVPair, error)
	// Unlock kvp previously acquired through a call to lock.
	Unlock(kvp *KVPair) error
	// TxNew returns a new Tx coordinator object or ErrNotSupported
	TxNew() (Tx, error)
	// Snapshot returns a kvdb snapshot and its version.
	Snapshot(prefix string) (Kvdb, uint64, error)
}
