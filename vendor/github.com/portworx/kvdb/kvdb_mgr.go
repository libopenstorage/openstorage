package kvdb

import (
	"fmt"
	"sync"
)

var (
	instance          Kvdb
	datastores        = make(map[string]DatastoreInit)
	datastoreVersions = make(map[string]DatastoreVersion)
	lock              sync.RWMutex
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
func New(
	name string,
	domain string,
	machines []string,
	options map[string]string,
	errorCB FatalErrorCB,
) (Kvdb, error) {
	lock.RLock()
	defer lock.RUnlock()

	if dsInit, exists := datastores[name]; exists {
		kvdb, err := dsInit(domain, machines, options, errorCB)
		return kvdb, err
	}
	return nil, ErrNotSupported
}

// Register adds specified datastore backend to the list of options.
func Register(name string, dsInit DatastoreInit, dsVersion DatastoreVersion) error {
	lock.Lock()
	defer lock.Unlock()
	if _, exists := datastores[name]; exists {
		return fmt.Errorf("Datastore provider %q is already registered", name)
	}
	datastores[name] = dsInit

	if _, exists := datastoreVersions[name]; exists {
		return fmt.Errorf("Datastore provider's %q version function already registered", name)
	}
	datastoreVersions[name] = dsVersion
	return nil
}

// Version returns the supported version for the provided kvdb endpoint.
func Version(name string, url string, kvdbOptions map[string]string) (string, error) {
	lock.RLock()
	defer lock.RUnlock()

	if dsVersion, exists := datastoreVersions[name]; exists {
		return dsVersion(url, kvdbOptions)
	}
	return "", ErrNotSupported
}
