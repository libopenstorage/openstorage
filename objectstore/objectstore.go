//go:generate mockgen -package=mock -destination=mock/objectstore.mock.go github.com/libopenstorage/openstorage/objectstore ObjectStore
package objectstore

import "errors"

var (
	ErrNotImplemented = errors.New("Not Implemented")
)

type ObjectStore interface {
	// ObjectStoreInspect returns status of objectstore
	ObjectStoreInspect() (*ObjectstoreInfo, error)
	// ObjectStoreCreate objectstore on specified volume
	ObjectStoreCreate(volume string) (*ObjectstoreInfo, error)
	// ObjectStoreDelete objectstore from cluster
	ObjectStoreDelete() error
	// ObjectStoreUpdate enable/disable objectstore
	ObjectStoreUpdate(enable bool) error
}

func NewDefaultObjectStore() ObjectStore {
	return &nullObjectStoreMgr{}
}

type nullObjectStoreMgr struct {
}

func (n *nullObjectStoreMgr) ObjectStoreInspect() (*ObjectstoreInfo, error) {
	return nil, ErrNotImplemented
}

func (n *nullObjectStoreMgr) ObjectStoreCreate(volume string) (*ObjectstoreInfo, error) {
	return nil, ErrNotImplemented
}

func (n *nullObjectStoreMgr) ObjectStoreUpdate(enable bool) error {
	return ErrNotImplemented
}

func (n *nullObjectStoreMgr) ObjectStoreDelete() error {
	return ErrNotImplemented
}
