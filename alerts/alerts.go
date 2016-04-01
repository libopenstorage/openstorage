package alerts

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
	"go.pedge.io/dlog"
)

const (
	// AlertDeleteAction is an alerts watch action for delete
	AlertDeleteAction AlertAction = iota
	// AlertCreateAction is an alerts watch action for create
	AlertCreateAction
	// AlertUpdateAction is an alerts watch action for update
	AlertUpdateAction
)

const (
	//Unknown Resource
	Unknown Resource = iota
	// Volume Resource
	Volume
	// Node Resource
	Node
	// Cluster Resource
	Cluster
)

var (
	// ErrNotSupported implemenation of a specific function is not supported.
	ErrNotSupported = errors.New("implementation not supported")
	// ErrNotFound raised if Key is not found
	ErrNotFound = errors.New("Key not found")
	// ErrExist raised if key already exists
	ErrExist = errors.New("Key already exists")
	// ErrUnmarshal raised if Get fails to unmarshal value.
	ErrUnmarshal = errors.New("Failed to unmarshal value")
	// ErrIllegal raised if object is not valid.
	ErrIllegal = errors.New("Illegal operation")
	// ErrNotInitialized raised if alerts not initialized
	ErrNotInitialized = errors.New("Alerts not initialized")
	// ErrAlertsClientNotFound raised if no client implementation found
	ErrAlertsClientNotFound = errors.New("Alerts client not found")
	// ErrResourceNotFound raised if ResourceType is not found
	ErrResourceNotFound = errors.New("Resource not found in Alerts")

	instances = make(map[string]AlertsClient)
	drivers   = make(map[string]InitFunc)

	mutex sync.Mutex
)

// AlertAction used to indicate the action performed on a KV pair
type AlertAction int

// Resource is equaivalent to api.ResourceType and is used in the alerts instance
// so that callers of the instance don't have to worry about api.*
type Resource int

// InitFunc initialization function for alerts
type InitFunc func(string, string, []string, string) (AlertsClient, error)

// AlertsWatcherFunc is a function type used as a callback for KV WatchTree
type AlertsWatcherFunc func(*api.Alerts, AlertAction, string, string) error

// AlertsClient interface for Alerts API
type AlertsClient interface {
	fmt.Stringer

	// Shutdown
	Shutdown()

	// GetKvdbInstance
	GetKvdbInstance() kvdb.Kvdb

	// Raise raises an Alerts
	Raise(alert api.Alerts) (api.Alerts, error)

	// Retrieve retrieves specific  Alerts
	Retrieve(resourceType api.ResourceType, id int64) (api.Alerts, error)

	// Enumerate enumerates Alerts
	Enumerate(filter api.Alerts) ([]*api.Alerts, error)

	// EnumerateWithinTimeRange enumerates Alertss between timeStart and timeEnd
	EnumerateWithinTimeRange(timeStart time.Time, timeEnd time.Time, resourceType api.ResourceType) ([]*api.Alerts, error)

	// Erase erases an Alerts
	Erase(resourceType api.ResourceType, alertID int64) error

	// Clear an Alerts
	Clear(resourceType api.ResourceType, alertID int64) error

	// Watch on all Alerts
	Watch(clusterId string, alertsWatcher AlertsWatcherFunc) error
}

type AlertsInstance interface {
	// Clear clears an alert
	Clear(resource Resource, resourceId string, alertID int64)

	// Alarm raises an alert with severity : ALARM
	Alarm(name string, msg string, resource Resource, resourceId string) (int64, error)

	// Notify raises an alert with severity : NOTIFY
	Notify(name string, msg string, resource Resource, resourceId string) (int64, error)

	// Warn raises an alert with severity : WARNING
	Warn(name string, msg string, resource Resource, resourceId string) (int64, error)

	// Alert :  Keeping this function for backward compatibility
	// until we remove all calls to this function
	Alert(name string, msg string) error
}

// Shutdown the alerts instance
func Shutdown() {
	mutex.Lock()
	defer mutex.Unlock()
	for _, v := range instances {
		v.Shutdown()
	}
}

// Get an alerts instance
func Get(name string) (AlertsClient, error) {
	if v, ok := instances[name]; ok {
		return v, nil
	}
	return nil, ErrAlertsClientNotFound
}

// New returns a new alerts instance
func New(name string, kvdbName string, kvdbBase string, kvdbMachines []string, clusterId string) (AlertsClient, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := instances[name]; ok {
		return nil, ErrExist
	}
	if initFunc, exists := drivers[name]; exists {
		driver, err := initFunc(kvdbName, kvdbBase, kvdbMachines, clusterId)
		if err != nil {
			return nil, err
		}
		instances[name] = driver
		return driver, err
	}
	return nil, ErrNotSupported
}

// NewAlertsInstance creates a new singleton istance of AlertsInstance
func NewAlertsInstance(version, nodeId, clusterId, kvdbName, kvdbBase string, kvdbMachines []string) {
	kva, err := Get(Name)
	if err != nil {
		kva, err = New(Name, kvdbName, kvdbBase, kvdbMachines, clusterId)
		if err != nil {
			dlog.Errorf("Failed to initialize an AlertsInstance ")
		}
	}
	newAlertsInstance(nodeId, clusterId, version, kva)
}

// Instance returns the singleton AlertsInstance
func Instance() AlertsInstance {
	return instance()
}

// Register an alerts interface
func Register(name string, initFunc InitFunc) error {
	mutex.Lock()
	defer mutex.Unlock()
	if _, exists := drivers[name]; exists {
		return ErrExist
	}
	drivers[name] = initFunc
	return nil
}
