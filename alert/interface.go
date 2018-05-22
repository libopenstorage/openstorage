package alert

import (
	"io"
	"time"
)

const (
	NoAlertFound Error = "no alert found for the input key"
)

// Error implements Error interface
type Error string

// Error reports underlying error as a string
func (e Error) Error() string {
	return string(e)
}

// AlertType uniquely identifies type of an alert using a number
type AlertType int64

// Altertable defines an interface that can be managed using Manager interface.
// An instance of Alertable is an alert object
type Alertable interface {
	// AlertType fetches alert type
	AlertType() AlertType
	// ResourceID fetches resource ID of alert generator
	ResourceID() string
	// Reason reports a brief comment on why alert was raised
	Reason() string
	// FirstSeen is the time stamp of first alert of this type and ID
	FirstSeen() time.Time
	// LastSeen is the time stamp of last alert of this type and ID
	LastSeen() time.Time
	// Marshal defines serialization of implementing object
	Marshal() ([]byte, error)
	// Unmarshal defines deserialization of implementing object
	Unmarshal(b []byte) error
	// Print prints the alert info to input writer
	Print(w io.Writer) error
}

// Manager defines an interface to manage alerts
type Manager interface {
	// Raise takes an input of Alertable type, serializes it and stores it in a backend service such as etcd
	Raise(alertable Alertable) error
	// Retrieve takes alertType and resourceID as inputs
	// and returns an instance of Alertable object. It returns NoAlertFound if no entry is found
	Retrieve(alertType AlertType, resourceID string) (Alertable, error)
	// Enumerate lists all alerts that have been raised so far
	Enumerate() ([]Alertable, error)
	// Count retrieves number of alerts raised for a particular key
	Count() int64
	// Delete removes an alert from the backend service such as etcd
	Delete(key string) error
}
