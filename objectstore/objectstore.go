package objectstore

import (
	"errors"

	"github.com/libopenstorage/openstorage/api"
)

const (
	Enable        = "enable"
	VolumeName    = "name"
	ObjectStoreID = "id"
)

var (
	ErrNotImplemented = errors.New("Not Implemented")
)

type ObjectStore interface {
	// ObjectStoreInspect returns status of objectstore
	ObjectStoreInspect(objectstoreID string) (*api.ObjectstoreInfo, error)
	// ObjectStoreCreate objectstore on specified volume
	ObjectStoreCreate(volume string) (*api.ObjectstoreInfo, error)
	// ObjectStoreDelete objectstore from cluster
	ObjectStoreDelete(objectstoreID string) error
	// ObjectStoreUpdate enable/disable objectstore
	ObjectStoreUpdate(objectstoreID string, enable bool) error
}

func NewDefaultObjectStore() ObjectStore {
	return &NullObjectStoreMgr{}
}

type NullObjectStoreMgr struct {
}

func (n *NullObjectStoreMgr) ObjectStoreInspect(objectstoreID string) (*api.ObjectstoreInfo, error) {
	return nil, ErrNotImplemented
}

func (n *NullObjectStoreMgr) ObjectStoreCreate(volume string) (*api.ObjectstoreInfo, error) {
	return nil, ErrNotImplemented
}

func (n *NullObjectStoreMgr) ObjectStoreUpdate(objectstoreID string, enable bool) error {
	return ErrNotImplemented
}

func (n *NullObjectStoreMgr) ObjectStoreDelete(objectstoreID string) error {
	return ErrNotImplemented
}
