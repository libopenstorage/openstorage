package k8s

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrK8SApiAccountNotSet is returned when the account used to talk to k8s api is not setup
var ErrK8SApiAccountNotSet = errors.New("k8s api account is not setup")

// ErrFailedToParseYAML error type for objects not found
type ErrFailedToParseYAML struct {
	// Path is the path of the yaml file that was to be parsed
	Path string
	// Cause is the underlying cause of the error
	Cause string
}

func (e *ErrFailedToParseYAML) Error() string {
	return fmt.Sprintf("Failed to parse file: %v due to err: %v", e.Path, e.Cause)
}

// ErrFailedToApplySpec error type for failing to apply a spec file
type ErrFailedToApplySpec struct {
	// Path is the path of the yaml file that was to be applied
	Path string
	// Cause is the underlying cause of the error
	Cause string
}

func (e *ErrFailedToApplySpec) Error() string {
	return fmt.Sprintf("Failed to apply spec file: %v due to err: %v", e.Path, e.Cause)
}

// ErrAppNotReady error type for when an app is not yet ready
type ErrAppNotReady struct {
	// ID is the identifier of the app
	ID string
	// Cause is the underlying cause of the error
	Cause string
}

func (e *ErrAppNotReady) Error() string {
	return fmt.Sprintf("app %v is not ready yet. Cause: %v", e.ID, e.Cause)
}

// ErrAppNotTerminated error type for when an app is not yet terminated
type ErrAppNotTerminated struct {
	// ID is the identifier of the app
	ID string
	// Cause is the underlying cause of the error
	Cause string
}

func (e *ErrAppNotTerminated) Error() string {
	return fmt.Sprintf("app %v is not terminated yet. Cause: %v", e.ID, e.Cause)
}

// ErrPVCNotReady error type for when a PVC is not yet ready/bound
type ErrPVCNotReady struct {
	// ID is the identifier of the app
	ID string
	// Cause is the underlying cause of the error
	Cause string
}

func (e *ErrPVCNotReady) Error() string {
	return fmt.Sprintf("PVC %v is not ready yet. Cause: %v", e.ID, e.Cause)
}

// ErrSnapshotNotReady error type for when a snapshot is not yet ready/bound
type ErrSnapshotNotReady struct {
	// ID is the identifier of the snapshot
	ID string
	// Cause is the underlying cause of the error
	Cause string
}

func (e *ErrSnapshotNotReady) Error() string {
	return fmt.Sprintf("Snapshot %v is not ready yet. Cause: %v", e.ID, e.Cause)
}

// ErrSnapshotDataNotReady error type for when a snapshot data is not yet ready
type ErrSnapshotDataNotReady struct {
	// ID is the identifier of the snapshot data
	ID string
	// Cause is the underlying cause of the error
	Cause string
}

func (e *ErrSnapshotDataNotReady) Error() string {
	return fmt.Sprintf("SnapshotData %v is not ready yet. Cause: %v", e.ID, e.Cause)
}

// ErrSnapshotFailed error type for when a snapshot has failed
type ErrSnapshotFailed struct {
	// ID is the identifier of the snapshot
	ID string
	// Cause is the underlying cause of the error
	Cause string
}

func (e *ErrSnapshotFailed) Error() string {
	return fmt.Sprintf("Snapshot %v has failed. Cause: %v", e.ID, e.Cause)
}

// ErrSnapshotDataFailed error type for when a snapshot data has failed
type ErrSnapshotDataFailed struct {
	// ID is the identifier of the snapshot data
	ID string
	// Cause is the underlying cause of the error
	Cause string
}

func (e *ErrSnapshotDataFailed) Error() string {
	return fmt.Sprintf("SnapshotData %v has failed. Cause: %v", e.ID, e.Cause)
}

// ErrFailedToValidateCustomSpec error type when CRD objects does not applied successfully
type ErrFailedToValidateCustomSpec struct {
	// Name of CRD object
	Name string
	// Cause is the underlying cause of the error
	Cause string
	// Type is the underlying type of CRD objects
	Type interface{}
}

func (e *ErrFailedToValidateCustomSpec) Error() string {
	return fmt.Sprintf("Failed to apply custom spec : %v of type %v due to err: %v", e.Name, reflect.TypeOf(e.Type), e.Cause)
}
