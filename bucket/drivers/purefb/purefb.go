package purefb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/libopenstorage/openstorage/bucket"
	"github.com/libopenstorage/openstorage/bucket/drivers/s3"
)

// PureFBDriver represents a way to interact with a Pure FlashBlade
// object storage backend via the AWS SDK
type PureFBDriver struct {
	*s3.S3Driver

	AccessKeyID     string
	SecretAccessKey string
}

// New create a new PureFBDriver instance
func New(cfg *aws.Config, accessKeyID, secretAccessKey string) (*PureFBDriver, error) {
	s3Driver, err := s3.New(cfg)
	if err != nil {
		return nil, err
	}

	return &PureFBDriver{
		S3Driver: s3Driver,

		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	}, nil
}

// GrantBucketAccess gives admin access to a bucket
func (p *PureFBDriver) GrantBucketAccess(id string, accountName string, accessPolicy string) (string, *bucket.BucketAccessCredentials, error) {
	return "", &bucket.BucketAccessCredentials{
		AccessKeyId:     p.AccessKeyID,
		SecretAccessKey: p.SecretAccessKey,
	}, nil
}

// RevokeBucketAccess is a no-op for admin access to a bucket
func (p *PureFBDriver) RevokeBucketAccess(id string, accountId string) error {
	return nil
}

// String returns the driver name for Pure FlashBlade
func (p *PureFBDriver) String() string {
	return "PureFBDriver"
}
