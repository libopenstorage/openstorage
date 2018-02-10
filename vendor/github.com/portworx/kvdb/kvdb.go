package kvdb

import (
	"errors"
	"time"

	"github.com/sirupsen/logrus"
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
	// KVCapabilityOrderedUpdates support requires watch to send an watch update
	// for every put - instead of coalescing multiple puts in one update.
	KVCapabilityOrderedUpdates = 1 << iota
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

const (
	// ReadPermission for read only access
	ReadPermission = iota
	// WritePermission for write only access
	WritePermission
	// ReadWritePermission for read-write access
	ReadWritePermission
)
const (
	// UsernameKey for an authenticated kvdb endpoint
	UsernameKey = "Username"
	// PasswordKey for an authenticated kvdb endpoint
	PasswordKey = "Password"
	// CAFileKey is the CA file path for an authenticated kvdb endpoint
	CAFileKey = "CAFile"
	// CertFileKey is the certificate file path for an authenticated kvdb endpoint
	CertFileKey = "CertFile"
	// CertKeyFileKey is the key to the certificate
	CertKeyFileKey = "CertKeyFile"
	// TrustedCAFileKey is the key for the trusted CA.
	TrustedCAFileKey = "TrustedCAFile"
	// ClientCertAuthKey is the boolean value indicating client authenticated certificate.
	ClientCertAuthKey = "ClientCertAuth"
	// RetryCountKey is the integer value indicating the retry count of etcd operations
	RetryCountKey = "RetryCount"
	// ACLTokenKey is the token value for ACL based KV stores
	ACLTokenKey = "ACLToken"
)

// List of kvdb endpoints supported versions
const (
	// ConsulVersion1 key
	ConsulVersion1 = "consulv1"
	// EtcdBaseVersion key
	EtcdBaseVersion = "etcd"
	// EtcdVersion3 key
	EtcdVersion3 = "etcdv3"
	// MemVersion1 key
	MemVersion1 = "memv1"
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
	// ErrNoPassword provided
	ErrNoPassword = errors.New("Username provided without any password")
	// ErrAuthNotSupported authentication not supported for this kvdb implementation
	ErrAuthNotSupported = errors.New("Kvdb authentication not supported")
	// ErrNoCertificate no certificate provided for authentication
	ErrNoCertificate = errors.New("Certificate File Path not provided")
	// ErrUnknownPermission raised if unknown permission type
	ErrUnknownPermission = errors.New("Unknown Permission Type")
	// ErrMemberDoesNotExist returned when an operation fails for a member
	// which does not exist
	ErrMemberDoesNotExist = errors.New("Kvdb member does not exist")
)

// KVAction specifies the action on a KV pair. This is useful to make decisions
// from the results of  a Watch.
type KVAction int

// KVFlags options for operations on KVDB
type KVFlags uint64

// PermissionType for user access
type PermissionType int

// WatchCB is called when a watched key or tree is modified. If the callback
// returns an error, then watch stops and the cb is called one last time
// with ErrWatchStopped.
type WatchCB func(prefix string, opaque interface{}, kvp *KVPair, err error) error

// FatalErrorCB callback is invoked incase of fatal errors
type FatalErrorCB func(format string, args ...interface{})

// DatastoreInit is called to activate a backend KV store.
type DatastoreInit func(domain string, machines []string, options map[string]string,
	cb FatalErrorCB) (Kvdb, error)

// DatastoreVersion is called to get the version of a backend KV store
type DatastoreVersion func(url string, kvdbOptions map[string]string) (string, error)

// EnumerateSelect function is a callback function provided to EnumerateWithSelect API
// This fn is executed over all the keys and only those values are returned by Enumerate for which
// this function return true.
type EnumerateSelect func(val interface{}) bool

// CopySelect function is a callback function provided to EnumerateWithSelect API
// This fn should perform a deep copy of the input interface and return the copy
type CopySelect func(val interface{}) interface{}

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
	Controller
	// String representation of backend datastore.
	String() string
	// Capbilities - see KVCapabilityXXX
	Capabilities() int
	// Get returns KVPair that maps to specified key or ErrNotFound.
	Get(key string) (*KVPair, error)
	// Get returns KVPair that maps to specified key or ErrNotFound. If found
	// value contains the unmarshalled result or error is ErrUnmarshal
	GetVal(key string, value interface{}) (*KVPair, error)
	// GetWithCopy returns a copy of the value as an interface for the specified key
	GetWithCopy(key string, copySelect CopySelect) (interface{}, error)
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
	// EnumerateWithSelect returns a copy of all values under the prefix that satisfy the select
	// function in the provided output array of interfaces
	EnumerateWithSelect(prefix string, enumerateSelect EnumerateSelect, copySelect CopySelect) ([]interface{}, error)
	// Delete deletes the KVPair specified by the key. ErrNotFound is returned
	// if the key is not found. The old KVPair is returned if successful.
	Delete(key string) (*KVPair, error)
	// DeleteTree same as Delete execpt that all keys sharing the prefix are
	// deleted.
	DeleteTree(prefix string) error
	// Keys returns an array of keys that share specified prefix (ie. "1st level directory").
	// sep parameter defines a key-separator, and if not provided the "/" is assumed.
	Keys(prefix, sep string) ([]string, error)
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
	// Snapshot returns a kvdb snapshot and its version.
	Snapshot(prefix string) (Kvdb, uint64, error)
	// SnapPut records the key value pair including the index.
	SnapPut(kvp *KVPair) (*KVPair, error)
	// Lock specfied key and associate a lockerID with it, probably to identify
	// who acquired the lock. The KVPair returned should be used to unlock.
	LockWithID(key string, lockerID string) (*KVPair, error)
	// Lock specfied key. The KVPair returned should be used to unlock.
	Lock(key string) (*KVPair, error)
	// Unlock kvp previously acquired through a call to lock.
	Unlock(kvp *KVPair) error
	// TxNew returns a new Tx coordinator object or ErrNotSupported
	TxNew() (Tx, error)
	// AddUser adds a new user to kvdb
	AddUser(username string, password string) error
	// RemoveUser removes a user from kvdb
	RemoveUser(username string) error
	// GrantUserAccess grants user access to a subtree/prefix based on the permission
	GrantUserAccess(username string, permType PermissionType, subtree string) error
	// RevokeUsersAccess revokes user's access to a subtree/prefix based on the permission
	RevokeUsersAccess(username string, permType PermissionType, subtree string) error
	// SetFatalCb sets the function to be called in case of fatal errors
	SetFatalCb(f FatalErrorCB)
	// SetLockTimeout sets maximum time a lock may be held
	SetLockTimeout(timeout time.Duration)
	// Serialize serializes all the keys under the domain and returns a byte array
	Serialize() ([]byte, error)
	// Deserialize deserializes the given byte array into a list of kv pairs
	Deserialize([]byte) (KVPairs, error)
}

// ReplayCb provides info required for replay
type ReplayCb struct {
	// Prefix is the watch key/tree prefix
	Prefix string
	// WaitIndex is the index after which updates must be returned
	WaitIndex uint64
	// Opaque is a hint returned by the caller
	Opaque interface{}
	// WatchCB is the watch callback
	WatchCB WatchCB
}

// UpdatesCollector collects updates from kvdb.
type UpdatesCollector interface {
	// Stop collecting updates
	Stop()
	// ReplayUpdates replays the collected updates.
	// Returns the version until the replay's were done
	// and any errors it encountered.
	ReplayUpdates(updateCb []ReplayCb) (uint64, error)
}

// NewUpdatesCollector creates new Kvdb collector that collects updates
// starting at startIndex + 1 index.
func NewUpdatesCollector(
	db Kvdb,
	prefix string,
	startIndex uint64,
) (UpdatesCollector, error) {
	collector := &updatesCollectorImpl{updates: make([]*kvdbUpdate, 0),
		startIndex: startIndex}
	logrus.Infof("Starting collector watch at %v", startIndex)
	if err := db.WatchTree(prefix, startIndex, nil, collector.watchCb); err != nil {
		return nil, err
	}
	return collector, nil
}

// List of kvdb controller ports
const (
	// PeerPort is the port on which peers identify themselves
	PeerPort = "2380"
	// ClientPort is the port on which clients send requests to kvdb.
	ClientPort = "2379"
)

// MemberInfo represents a member of the kvdb cluster
type MemberInfo struct {
	PeerUrls   []string
	ClientUrls []string
	Leader     bool
	DbSize     int64
	IsHealthy  bool
}

// Controller interface provides APIs to manage Kvdb Cluster and Kvdb Clients.
type Controller interface {
	// AddMember adds a new member to an existing kvdb cluster. Add should be
	// called on a kvdb node where kvdb is already running. It should be
	// followed by a Setup call on the actual node
	// Returns: map of nodeID to peerUrls of all members in the initial cluster or error
	AddMember(nodeIP, nodePeerPort, nodeName string) (map[string][]string, error)

	// RemoveMember removes a member from an existing kvdb cluster
	// Returns: error if it fails to remove a member
	RemoveMember(nodeID string) error

	// ListMembers enumerates the members of the kvdb cluster
	// Returns: the nodeID  to memberInfo mappings of all the members
	ListMembers() (map[string]*MemberInfo, error)

	// SetEndpoints set the kvdb endpoints for the client
	SetEndpoints(endpoints []string) error

	// GetEndpoints returns the kvdb endpoints for the client
	GetEndpoints() []string
}
