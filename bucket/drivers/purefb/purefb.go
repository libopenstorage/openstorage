package purefb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/libopenstorage/openstorage/bucket"
	"github.com/libopenstorage/openstorage/bucket/drivers/s3"
)

type PureFBDriver struct {
	*s3.S3Driver

	AccessKeyID     string
	SecretAccessKey string
}

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

func (p *PureFBDriver) GrantBucketAccess(id string, accountName string, accessPolicy string) (string, *bucket.BucketAccessCredentials, error) {
	return "", &bucket.BucketAccessCredentials{
		AccessKeyId:     p.AccessKeyID,
		SecretAccessKey: p.SecretAccessKey,
	}, nil
}

// RevokeBucketAccess
func (p *PureFBDriver) RevokeBucketAccess(id string, accountId string) error {
	return nil
}
