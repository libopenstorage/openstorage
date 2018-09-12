package alert

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
)

var (
	// ErrNotSupported implemenation of a specific function is not supported.
	ErrNotSupported = errors.New("Implementation not supported")
	// ErrNotFound raised if Key is not found.
	ErrNotFound = errors.New("Key not found")
	// ErrExist raised if key already exists.
	ErrExist = errors.New("Key already exists")
	// ErrUnmarshal raised if Get fails to unmarshal value.
	ErrUnmarshal = errors.New("Failed to unmarshal value")
	// ErrIllegal raised if object is not valid.
	ErrIllegal = errors.New("Illegal operation")
	// ErrNotInitialized raised if alert not initialized.
	ErrNotInitialized = errors.New("Alert not initialized")
	// ErrAlertClientNotFound raised if no client implementation found.
	ErrAlertClientNotFound = errors.New("Alert client not found")
	// ErrResourceNotFound raised if ResourceType is not found>
	ErrResourceNotFound = errors.New("Resource not found in Alert")
	// ErrSubscribedRaise raised if unable to raise a subscribed alert
	ErrSubscribedRaise = errors.New("Could not raise alert and its subscribed alerts")

	instances = make(map[string]Alert)
	drivers   = make(map[string]InitFunc)

	lock sync.RWMutex
)

// InitFunc initialization function for alert.
// Deprecated: Kept temporarily for backward compatibility.
type InitFunc func(kv kvdb.Kvdb, clusterID string) (Alert, error)

// AlertWatcherFunc is a function type used as a callback for KV WatchTree.
// Deprecated: Kept temporarily for backward compatibility.
type AlertWatcherFunc func(*api.Alert, api.AlertActionType, string, string) error

// Alert interface for Alert API.
// Deprecated: Kept temporarily for backward compatibility.
type Alert interface {
	fmt.Stringer

	// Shutdown.
	// Deprecated: Kept temporarily for backward compatibility.
	Shutdown()

	// GetKvdbInstance.
	// Deprecated: Kept temporarily for backward compatibility.
	GetKvdbInstance() kvdb.Kvdb

	// Raise raises an Alert.
	// Deprecated: Kept temporarily for backward compatibility.
	Raise(alert *api.Alert) error

	// Raise raises an Alert only if another alert with given resource type,
	// resource id, and unqiue_tage doesnt exists already.
	// Deprecated: Kept temporarily for backward compatibility.
	RaiseIfNotExist(alert *api.Alert) error

	// Subscribe allows a child (dependent) alert to subscribe to a parent alert
	// Deprecated: Kept temporarily for backward compatibility.
	Subscribe(parentAlertType int64, childAlert *api.Alert) error

	// Retrieve retrieves specific Alert.
	// Deprecated: Kept temporarily for backward compatibility.
	Retrieve(resourceType api.ResourceType, id int64) (*api.Alert, error)

	// Enumerate enumerates Alert.
	// Deprecated: Kept temporarily for backward compatibility.
	Enumerate(filter *api.Alert) ([]*api.Alert, error)

	// EnumerateWithinTimeRange enumerates Alert between timeStart and timeEnd.
	// Deprecated: Kept temporarily for backward compatibility.
	EnumerateWithinTimeRange(
		timeStart time.Time,
		timeEnd time.Time,
		resourceType api.ResourceType,
	) ([]*api.Alert, error)

	// Erase erases an Alert.
	// Deprecated: Kept temporarily for backward compatibility.
	Erase(resourceType api.ResourceType, alertID int64) error

	// Clear an Alert.
	// Deprecated: Kept temporarily for backward compatibility.
	Clear(resourceType api.ResourceType, alertID int64, ttl uint64) error

	// Clear an Alert for a resource with unique tag.
	// Deprecated: Kept temporarily for backward compatibility.
	ClearByUniqueTag(
		resourceType api.ResourceType,
		resourceId string,
		uniqueTag string,
		ttl uint64,
	) error

	// Watch on all Alerts for the given clusterID. It uses the global kvdb
	// options provided while creating the alertClient object to access this
	// cluster
	// Deprecated: Kept temporarily for backward compatibility.
	Watch(clusterID string, alertWatcher AlertWatcherFunc) error
}

// Shutdown the alert instance.
// Deprecated: Kept temporarily for backward compatibility.
func Shutdown() {
	lock.Lock()
	defer lock.Unlock()
	for _, v := range instances {
		v.Shutdown()
	}
}

// New returns a new alert instance tied with a clusterID and kvdb.
// Deprecated: Kept temporarily for backward compatibility.
func New(name string, clusterID string, kv kvdb.Kvdb) (Alert, error) {
	lock.Lock()
	defer lock.Unlock()

	if initFunc, exists := drivers[name]; exists {
		driver, err := initFunc(kv, clusterID)
		if err != nil {
			return nil, err
		}
		instances[name] = driver
		return driver, err
	}
	return nil, ErrNotSupported
}

// Register an alert interface.
// Deprecated: Kept temporarily for backward compatibility.
func Register(name string, initFunc InitFunc) error {
	lock.Lock()
	defer lock.Unlock()
	if _, exists := drivers[name]; exists {
		return ErrExist
	}
	drivers[name] = initFunc
	return nil
}
