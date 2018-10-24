package errors

import "fmt"

// ErrNotFound error type for objects not found
type ErrNotFound struct {
	// ID unique object identifier.
	ID string
	// Type of the object which wasn't found
	Type string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%v with ID: %v not found", e.Type, e.ID)
}

// ErrExists type for objects already present
type ErrExists struct {
	// ID unique object identifier.
	ID string
	// Type of the object which already exists
	Type string
}

func (e *ErrExists) Error() string {
	return fmt.Sprintf("%v with ID: %v already exists", e.Type, e.ID)
}

// ErrNotSupported error type for APIs that are not supported
type ErrNotSupported struct{}

func (e *ErrNotSupported) Error() string {
	return fmt.Sprintf("Not Supported")
}
