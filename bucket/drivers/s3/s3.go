package s3

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
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
func (d *S3Driver) NewS3Svc(region string) (*s3.S3, error) {
	// Override the aws config with the region
	s3Config := &aws.Config{
		Credentials: d.config.Credentials,
		Region:      aws.String(region),
	}
	sess, err := session.NewSession(s3Config)
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)
	return svc, nil
}

// Returns a new IAM service client
func (d *S3Driver) NewIamSvc() (*iam.IAM, error) {
	iamConfig := &aws.Config{
		Credentials: d.config.Credentials,
	}
	sess, err := session.NewSession(iamConfig)
	if err != nil {
		return nil, err
	}

	iamSvc := iam.New(sess)
	return iamSvc, nil
}

func (d *S3Driver) CreateBucket(name string, region string) (string, error) {
	// Update driver region config
	svc, err := d.NewS3Svc(region)
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

func (d *S3Driver) DeleteBucket(id string, region string, clearBucket bool) error {
	svc, err := d.NewS3Svc(region)
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

// Creates IAM user based on the given account name
// Returns accounName as accountId as access can be granted
// and revoked based on accounName
func createUser(accountName string, iamSvc *iam.IAM) (string, error) {
	createUserResult, err := iamSvc.CreateUser(&iam.CreateUserInput{
		UserName: aws.String(accountName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == iam.ErrCodeEntityAlreadyExistsException {
			// If the account already exists then we can return success
			logrus.Infof("User %s already exists", accountName)
			return accountName, nil
		}
		return "", fmt.Errorf("unable to create user %s: %v", accountName, err)
	}
	// Arn format : arn:aws:iam::981779513211:user/name
	logrus.Infof("Created arn %s", *createUserResult.User.Arn)
	return accountName, nil
}

// createAccessKey creates access key for the account
func createAccessKey(accountName string, iamSvc *iam.IAM) (string, error) {
	accessKeyResult, err := iamSvc.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(accountName),
	})
	if err != nil {
		return "", fmt.Errorf("unable to retrieve access credentials for user %s: %v", accountName, err)
	}
	accessKeyMap := map[string]interface{}{
		"AccessKeyId":     accessKeyResult.AccessKey.AccessKeyId,
		"SecretAccessKey": accessKeyResult.AccessKey.SecretAccessKey,
	}
	accessKeyMarshalled, err := json.Marshal(accessKeyMap)
	if err != nil {
		return "", fmt.Errorf("unable to decode the credentials for account %s: %v", accountName, err)
	}
	credentials := string(accessKeyMarshalled[:])
	return credentials, nil
}

// getUserPolicyInput gets the policy input that can be applied to the iam account
func getUserPolicyInput(bucketId string, accountName string, inputAccessPolicy string, effect string) (*iam.PutUserPolicyInput, error) {
	accessPolicy := inputAccessPolicy
	// If acces policy is not specified, use default access policy for the bucket
	if len(accessPolicy) == 0 {
		bucketPolicy := map[string]interface{}{
			"Statement": []map[string]interface{}{
				{
					"Effect": effect,
					"Action": []string{
						"s3:*",
					},
					"Resource": []string{
						fmt.Sprintf("arn:aws:s3:::%s/*", bucketId),
						fmt.Sprintf("arn:aws:s3:::%s", bucketId),
					},
				},
			},
		}
		policy, err := json.Marshal(bucketPolicy)
		if err != nil {
			return nil, err
		}
		accessPolicy = string(policy[:])
	}
	input := &iam.PutUserPolicyInput{
		// PolicyDocument: aws.String("{\"Statement\":{\"Effect\":\"Allow\",\"Action\":\"*\",\"Resource\":\"*\"}}"),
		PolicyDocument: aws.String(accessPolicy),
		PolicyName:     aws.String(bucketId + "-AccessPolicy"),
		UserName:       aws.String(accountName),
	}
	return input, nil
}

// grantAccessToBucket allows iam account access to the bucket
func grantAccessToBucket(bucketId string, accountName string, accessPolicy string, iamSvc *iam.IAM) error {
	input, err := getUserPolicyInput(bucketId, accountName, accessPolicy, "Allow")
	if err != nil {
		return err
	}
	_, err = iamSvc.PutUserPolicy(input)
	if err != nil {
		return fmt.Errorf("unable to provide user %s access to bucket %s : %v", accountName, bucketId, err)
	}
	return nil
}

// revokeAccessToBucket revoks iam account access to the bucket
func revokeAccessToBucket(bucketId string, accountName string, iamSvc *iam.IAM) error {
	input, err := getUserPolicyInput(bucketId, accountName, "", "Deny")
	if err != nil {
		return err
	}
	_, err = iamSvc.PutUserPolicy(input)
	if err != nil {
		return fmt.Errorf("unable to revoke user %s access to bucket %s : %v", accountName, bucketId, err)
	}
	return nil
}

// AccessBucket grants access to the S3 bucket
func (d *S3Driver) GrantBucketAccess(id string, accountName string, accessPolicy string) (string, string, error) {
	iamSvc, err := d.NewIamSvc()
	if err != nil {
		return "", "", fmt.Errorf("unable to create iam session: %v ", err)
	}

	accountId, err := createUser(accountName, iamSvc)
	if err != nil {
		return "", "", err
	}

	err = grantAccessToBucket(id, accountName, accessPolicy, iamSvc)
	if err != nil {
		return "", "", err
	}

	credentials, err := createAccessKey(accountName, iamSvc)
	if err != nil {
		return "", "", err
	}

	logrus.Infof("Account %s granted access to bucket %s", accountName, id)
	return accountId, credentials, nil
}

// Actual implementation to be done once we have more clarity on the downstream API
func (d *S3Driver) RevokeBucketAccess(id string, accountId string) error {
	iamSvc, err := d.NewIamSvc()
	if err != nil {
		return fmt.Errorf("unable to create iam session: %v ", err)
	}
	err = revokeAccessToBucket(id, accountId, iamSvc)
	if err != nil {
		return err
	}
	logrus.Infof("Account %s revoked access to bucket %s", accountId, id)
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
