package errors

import (
	"fmt"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/parser"
)

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

// ErrStoragePoolOperationInProgress error when an operation is already in progress
// on a storage pool
type ErrStoragePoolOperationInProgress struct {
	// Pool is the affected pool
	Pool *api.StoragePool
}

func (e *ErrStoragePoolOperationInProgress) Error() string {
	errMsg := fmt.Sprintf("another operation for pool %s is already in progress.", e.Pool.GetUuid())
	if e.Pool.LastOperation != nil {
		op := e.Pool.LastOperation
		if op.Type == api.SdkStoragePool_OPERATION_RESIZE {
			errMsg = fmt.Sprintf("%s (resize) %s %s", errMsg, op.Msg, parser.LabelsToString(op.Params))
		} else if op.Type == api.SdkStoragePool_OPERATION_UPDATE_IOPS {
			errMsg = fmt.Sprintf("%s (update IOPS) %s %s", errMsg, op.Msg, parser.LabelsToString(op.Param))
		}
	}
	return errMsg
}
