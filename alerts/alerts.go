package alerts

import (
	"errors"
	"time"
)

// ResourceType indicates the resource type to which an alert is associated
type ResourceType int

const (
	// Unknown resource type undefined.
	Unknown ResourceType = iota
	// Volume Resource Type
	Volumes
	// Node Resource Type
	Node
	// Cluster Resource Type
	Cluster
)

// Alert object holds alert information
type Alert struct {
	// Id for Alert
	Id int64
	// Severity of Alert
	Severity int
	// Name of Alert
	Name string
	// Message describing Alert
	Message string
	// Timestamp when Alert occured
	Timestamp time.Time
	// ResourceId where Alert occured
	ResourceId string
	// Resource where Alert occured
	Resource ResourceType
	// Cleared Flag
	Cleared bool
}

const (
	// Name of this module
	Name = "alerts"
	// AlertsNotify - Notification
	AlertsNotify = 3
	// AlertsWarning - Warning
	AlertsWarning = 2
	// AlertsAlarm - Alarming
	AlertsAlarm = 1
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
	// ErrResourceNotFound raised if ResourceType is not found
	ErrResourceNotFound = errors.New("Resource not found in Alert")
)

// AlertClient interface for Alerts API
type AlertClient interface {
	//String
	String() string

	// Raise raises an Alert
	Raise(alert Alert) (Alert, error)

	// RaiseWithGenerateId raises an Alert with a custom generateId function for alertIds
	RaiseWithGenerateId(alert Alert, generateId func() (int64, error)) (Alert, error)

	// Retrieve retrieves specific  Alert
	Retrieve(resourceType ResourceType, id uint64) (Alert, error)

	// Enumerate enumerates Alerts
	Enumerate(filter Alert) ([]*Alert, error)

	// EnumerateWithinTimeRange enumerates Alerts between timeStart and timeEnd
	EnumerateWithinTimeRange(timeStart time.Time, timeEnd time.Time, resourceType ResourceType) ([]*Alert, error)

	// Erase erases an Alert
	Erase(resourceType ResourceType, alertID uint64) error

	// Clear an Alert
	Clear(resourceType ResourceType, alertID uint64) error
}
