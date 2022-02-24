package bucket

// BucketDriver represents an method for interacting with various object storage backends
type BucketDriver interface {
	// String returns a name for the driver implementation
	String() string

	// Start starts the bucket driver for usage
	Start() error

	// CreateBucket provisions a new bucket
	CreateBucket(name string) (string, error)

	// DeleteBucket deprovisions a bucket
	DeleteBucket(id string) error
}
