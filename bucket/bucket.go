package bucket

import (
	"github.com/libopenstorage/openstorage/api"
)

// BucketDriver represents an method for interacting with various object storage backends
type BucketDriver interface {
	// String returns a name for the driver implementation
	String() string

	// Start starts the bucket driver for usage
	Start() error

	// CreateBucket provisions a new bucket
	CreateBucket(name string, region string, endpoint string, anonymousBucketAccessMode api.AnonymousBucketAccessMode) (string, error)

	// DeleteBucket deprovisions the bucket
	DeleteBucket(id string, region string, endpoint string, clearBucket bool) error

	// GrantBucketAccess grants access to the bucket
	// Returns accountId and creadentials for access
	GrantBucketAccess(id string, accountName string, accessPolicy string) (string, *BucketAccessCredentials, error)

	// RevokeBucketAccess revokes access to the bucket
	RevokeBucketAccess(id string, accountId string) error
}

type BucketAccessCredentials struct {
	// Access key id for accessing the bucket
	AccessKeyId string
	// Secret access key for access the bucket
	SecretAccessKey string
}
