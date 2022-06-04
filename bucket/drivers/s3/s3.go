package s3

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/sirupsen/logrus"
)

type S3Driver struct {
	config *aws.Config
}

func New(config *aws.Config) (*S3Driver, error) {
	if config == nil {
		return nil, fmt.Errorf("must provide S3 config")
	}

	return &S3Driver{
		config: config,
	}, nil
}

// Returns a new S3 service client
func (d *S3Driver) NewSvc() (*s3.S3, error) {
	// Override the aws config with the region
	sess, err := session.NewSession(d.config)
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)
	return svc, nil
}

func (d *S3Driver) updateRegion(region string) {
	d.config.Region = aws.String(region)
}

func (d *S3Driver) CreateBucket(name string, region string) (string, error) {
	// Update driver region config
	d.updateRegion(region)
	svc, err := d.NewSvc()
	if err != nil {
		return "", fmt.Errorf("unable to create S3 session, %v ", err)
	}
	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(region),
		},
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				logrus.Error("Bucket name is already in use. Provide a new globally unique name")
				return "", fmt.Errorf("bucket %s is not a  globally unique name: %v", name, err)
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				logrus.Info("Bucket exists and is owned by the requester")
				return name, nil
			default:
				logrus.Errorf("An unknown error has occurred while creating bucket %s: %v", name, err)
			}
		}
		return "", err
	}
	logrus.Infof("Created S3 Bucket: %s in region %s ", name, region)
	// Bucket name is same as bucket_id in S3
	return name, nil
}

func deleteObjectsInBucket(id string, svc *s3.S3) error {
	// DeleteBucket is possible only if the Bucket is empty.
	// Setup BatchDeleteIterator to iterate through a list of objects.
	iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
		Bucket: aws.String(id),
	})
	// Traverse iterator deleting each object
	err := s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iter)
	if err != nil {
		logrus.Infof("Unable to delete objects from bucket %q  %v", id, err)
		return err
	}
	logrus.Infof("Deleted object(s) from bucket: %v", id)
	return nil
}

func (d *S3Driver) DeleteBucket(id string, clearBucket bool) error {
	svc, err := d.NewSvc()
	if err != nil {
		return fmt.Errorf("unable to create S3 session: %v ", err)
	}

	if clearBucket {
		err = deleteObjectsInBucket(id, svc)
		if err != nil {
			return fmt.Errorf("can not delete objects in the bucket: %v ", err)
		}
	}

	_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(id),
	})
	if err != nil {
		return err
	}

	err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
		Bucket: aws.String(id),
	})
	if err != nil {
		return err
	}
	logrus.Infof("Deleted bucket: %v from S3", id)
	return nil
}

// AccessBucket grants access to the S3 bucket
// Dummy impplementation
// Actual implementation to be done once we have more clarity on the downstream API
func (d *S3Driver) GrantBucketAccess(id string, accountName string, accessPolicy string) (string, string, error) {
	logrus.Info("bucket_driver.S3 access bucket received")
	return accountName, "", nil
}

// DeleteBucket deprovisions an S3 buceket
// Dummy impplementation
// Actual implementation to be done once we have more clarity on the downstream API
func (d *S3Driver) RevokeBucketAccess(id string, accountId string) error {
	logrus.Info("bucket_driver.S3 revoke bucket received")
	return nil
}

// String name representation of driver
// Not being used currently
func (f *S3Driver) String() string {
	return "S3Driver"
}

// Start starts a new s3 object storage server
// Not being used currently
func (f *S3Driver) Start() error {
	logrus.Infof("Starting s3 bucket driver.")
	return nil
}
