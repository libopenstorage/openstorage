package alerts

import (
	"errors"
	"time"
	"sync"
	"github.com/libopenstorage/openstorage/api"
)

// AlertAction used to indicate the action performed on a KV pair
type AlertAction int
// InitFunc initialization function for alerts
type InitFunc func() (AlertsClient, error)
// AlertsWatcherFunc is a function type used as a callback for KV WatchTree
type AlertsWatcherFunc func(*api.Alerts, AlertAction, string, string) (error)


const (
	// AlertDeleteAction is an alerts watch action for delete
	AlertDeleteAction AlertAction = iota
	// AlertCreateAction is an alerts watch action for create
	AlertCreateAction
	// AlertUpdateAction is an alerts watch action for update
	AlertUpdateAction
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
	instances             map[string]AlertsClient
	drivers               map[string]InitFunc
	mutex                 sync.Mutex
)

// AlertsClient interface for Alerts API
type AlertsClient interface {
	//String
	String() string

	// Shutdown
	Shutdown()

	// Raise raises an Alerts
	Raise(alert api.Alerts) (api.Alerts, error)

	// RaiseWithGenerateId raises an Alerts with a custom generateId function for alertIds
	RaiseWithGenerateId(alert api.Alerts, generateId func() (int64, error)) (api.Alerts, error)

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
	Watch(alertsWatcher AlertsWatcherFunc) error
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
func New(name string) (AlertsClient, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := instances[name]; ok {
		return nil, ErrExist
	}
	if initFunc, exists := drivers[name]; exists {
		driver, err := initFunc()
		if err != nil {
			return nil, err
		}
		instances[name] = driver
		return driver, err
	}
	return nil, ErrNotSupported
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

func init() {
	drivers = make(map[string]InitFunc)
	instances = make(map[string]AlertsClient)
}
