package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
)

type S3Driver struct {
	//svc is IAM S3 service client
	svc    *s3.S3
	config *aws.Config
}

func New(config *aws.Config) *S3Driver {
	sess, _ := session.NewSession(config)
	// Create S3 service client
	svc := s3.New(sess)
	return &S3Driver{
		svc:    svc,
		config: config,
	}
}

func (d *S3Driver) CreateBucket(name string) (string, error) {
	_, err := d.svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(*d.config.Region),
		},
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				logrus.Error("Bucket name already in use!")
				panic(err)
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				logrus.Warn("Bucket exists and is owned by the requester")
			default:
				logrus.Error("Unknown error")
			}
		}
		return "", err
	}
	logrus.Infof("Created Bucket %s in S3 ", name)
	// Bucket name is same as bucket_id in S3
	return name, nil
}

func (d *S3Driver) DeleteBucket(id string) error {
	// DeleteBucket is possible only if the Bucket is empty.
	// Setup BatchDeleteIterator to iterate through a list of objects.
	iter := s3manager.NewDeleteListIterator(d.svc, &s3.ListObjectsInput{
		Bucket: aws.String(id),
	})

	// Traverse iterator deleting each object
	err := s3manager.NewBatchDeleteWithClient(d.svc).Delete(aws.BackgroundContext(), iter)
	if err != nil {
		logrus.Infof("Unable to delete objects from bucket %q  %v", id, err)
		return err
	}
	logrus.Infof("Deleted object(s) from bucket: %v", id)

	_, err = d.svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(id),
	})
	if err != nil {
		return err
	}

	err = d.svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
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
