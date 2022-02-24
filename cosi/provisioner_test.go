package cosi

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"
)

func TestProvisionerBucketCreate(t *testing.T) {
	// Create test server
	testServer := newCOSITestServer(t)
	defer testServer.Stop()
	time.Sleep(1 * time.Second)

	// Create COSI bucket
	cosiClient := cosi.NewProvisionerClient(testServer.conn)
	bucketName := "newbucket"
	_, err := cosiClient.ProvisionerCreateBucket(context.TODO(), &cosi.ProvisionerCreateBucketRequest{
		Name: bucketName,
	})
	assert.NoError(t, err)

	// Configure s3 client & test bucket
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials("YOUR-ACCESSKEYID", "YOUR-SECRETACCESSKEY", ""),
		Endpoint:         aws.String("127.0.0.1:8085"),
		Region:           aws.String("eu-central-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)
	testString := "TESTSTRING"

	// Put new object
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Body:   strings.NewReader(testString),
		Bucket: aws.String(bucketName),
		Key:    aws.String("test.txt"),
	})
	assert.NoError(t, err)

	// Get object from bucket
	obj, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("newbucket"),
		Key:    aws.String("test.txt"),
	})
	assert.NoError(t, err)

	bs, err := ioutil.ReadAll(obj.Body)
	assert.NoError(t, err)
	assert.Equal(t, testString, string(bs))
}

func TestProvisionerBucketDelete(t *testing.T) {
	// Create test server
	testServer := newCOSITestServer(t)
	defer testServer.Stop()
	time.Sleep(1 * time.Second)

	// Create COSI bucket
	cosiClient := cosi.NewProvisionerClient(testServer.conn)
	bucketName := "newbucket2"
	_, err := cosiClient.ProvisionerCreateBucket(context.TODO(), &cosi.ProvisionerCreateBucketRequest{
		Name: bucketName,
	})
	assert.NoError(t, err)

	// Delete COSI bucket
	_, err = cosiClient.ProvisionerDeleteBucket(context.TODO(), &cosi.ProvisionerDeleteBucketRequest{
		BucketId: bucketName,
	})
	assert.NoError(t, err)
}
