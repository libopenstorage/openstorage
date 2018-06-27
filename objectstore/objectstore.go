//go:generate mockgen -package=mock -destination=mock/objectstore.mock.go github.com/libopenstorage/openstorage/objectstore ObjectStore
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
