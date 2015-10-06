package kvdb

import (
	"fmt"
	"sync"
)

// DatastoreInit is called to activate a backend KV store.
type DatastoreInit func(domain string,
	machines []string,
	options map[string]string) (Kvdb, error)

var (
	instance   Kvdb
	datastores map[string]DatastoreInit
	lock       sync.Mutex
)

// Instance returns instance set via SetInstance, nil if none was set.
func Instance() Kvdb {
	return instance
}

// SetInstance sets the singleton instance.
func SetInstance(kvdb Kvdb) error {
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		if instance == nil {
			instance = kvdb
			return nil
		}
	}
	return fmt.Errorf("Kvdb instance is already set to %q", instance.String())
}

// New return a new instance of KVDB as specified by datastore name.
// If domain is set all requests to KVDB are prefixed by domain.
// options is interpreted by backend KVDB.
func New(name string,
	domain string,
	machines []string,
	options map[string]string) (Kvdb, error) {

	if dsInit, exists := datastores[name]; exists {
		kvdb, err := dsInit(domain, machines, options)
		return kvdb, err
	}
	return nil, ErrNotSupported
}

// Register adds specified datastore backend to the list of options.
func Register(name string, dsInit DatastoreInit) error {
	if _, exists := datastores[name]; exists {
		return fmt.Errorf("Datastore provider %q is already registered", name)
	}
	datastores[name] = dsInit
	return nil
}

func init() {
	datastores = make(map[string]DatastoreInit)
}
