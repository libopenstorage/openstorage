package osdconfig

import (
	"sync"

	"github.com/portworx/kvdb"
)

// configManager implements ConfigManager
type configManager struct {
	// wrap a handle to kvdb
	kv kvdb.Kvdb

	// hashmap for callback bookkeeping
	cbCluster map[string]CallbackClusterConfigFunc
	cbNode    map[string]CallbackNodeConfigFunc

	// mutex for locking during key operations
	sync.Mutex
}

// type for callback func for cluster
type CallbackClusterConfigFunc func(config *ClusterConfig) error
type CallbackNodeConfigFunc func(config *NodeConfig) error

// Watcher is a classifier for registering function
type Watcher string

// dataToKvdb is data to be sent to kvdb as a state to run on
type dataToKvdb struct {
	Type Watcher
}

// data is data to be sent to callbacks
// The contents here are populated based on what is received from kvdb
// Callback sends an instance of this on a channel that others can only write on
type data struct {
	// kvdb key received in kvdb.KvPair
	Key string

	// kvdb byte buffer received in kvdb.KvPair
	Value []byte

	// Type
	Type Watcher
}
