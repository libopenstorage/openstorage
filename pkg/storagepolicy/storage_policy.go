package storagepolicy

import (
	"github.com/libopenstorage/openstorage/api"
)

// PolicyManager provides an implementation if
// OpenStoragePolicy service methods.
type PolicyManager interface {
	api.OpenStoragePolicyServer
	GetEnforcement() (*api.SdkStoragePolicy, error)
}
