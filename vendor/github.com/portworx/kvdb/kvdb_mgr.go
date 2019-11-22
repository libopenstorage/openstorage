package kvdb

import (
	"fmt"
	"sync"
)

var (
	instance          Kvdb
	datastores        = make(map[string]DatastoreInit)
	datastoreVersions = make(map[string]DatastoreVersion)
	wrappers          = make(map[WrapperName]WrapperInit)
	lock              sync.RWMutex
)

// Instance returns instance set via SetInstance, nil if none was set.
func Instance() Kvdb {
	return instance
}

// SetInstance sets the singleton instance.
func SetInstance(kvdb Kvdb) error {
	instance = kvdb
	return nil
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
		return dsInit(domain, machines, options, errorCB)
	}
	return nil, ErrNotSupported
}

// AddWrapper adds wrapper is it is not already added
func AddWrapper(
	wrapper WrapperName,
	kvdb Kvdb,
	options map[string]string,
) (Kvdb, error) {
	lock.Lock()
	defer lock.Unlock()

	for w := kvdb; w != nil; w = w.WrappedKvdb() {
		if w.WrapperName() == wrapper {
			return kvdb, fmt.Errorf("wrapper %v already present in kvdb",
				wrapper)
		}
	}
	if initFn, ok := wrappers[wrapper]; !ok {
		return kvdb, fmt.Errorf("wrapper %v not found", wrapper)
	} else {
		// keep log wrapper at the top if it exists
		if kvdb.WrapperName() == Wrapper_Log {
			newWrapper, err := initFn(kvdb.WrappedKvdb(), options)
			if err == nil {
				kvdb.SetWrappedKvdb(newWrapper)
				return kvdb, nil
			} else {
				return kvdb, err
			}
		}
		return initFn(kvdb, options)
	}
}

// RemoveWrapper adds wrapper is it is not already added
func RemoveWrapper(
	wrapper WrapperName,
	kvdb Kvdb,
) (Kvdb, error) {
	var prevWrapper Kvdb
	for w := kvdb; w != nil; w = w.WrappedKvdb() {
		if w.WrapperName() == wrapper {
			w.Removed()
			if prevWrapper != nil {
				prevWrapper.SetWrappedKvdb(w.WrappedKvdb())
				return kvdb, nil
			} else {
				// removing the top-most wrapper
				return w.WrappedKvdb(), nil
			}
		}
		prevWrapper = w
	}
	return kvdb, nil // did not find the wrapper to remove
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

// Register wrapper
func RegisterWrapper(name WrapperName, initFn WrapperInit) error {
	lock.Lock()
	defer lock.Unlock()
	wrappers[name] = initFn
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
