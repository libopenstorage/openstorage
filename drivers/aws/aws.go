package ebs

import (
	_ "encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	Name         = "aws"
	AwsTableName = "AwsOpenStorage"
)

var (
	devMinor int32
)

// This data is persisted in a DB.
type awsVolume struct {
	spec       api.VolumeSpec
	formatted  bool
	attached   bool
	mounted    bool
	device     string
	mountpath  string
	instanceID string
}

// Implements the open storage volume interface.
type awsProvider struct {
	ec2 *ec2.EC2
	db  *dynamodb.DynamoDB
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	// Initialize the EC2 interface.
	creds := credentials.NewEnvCredentials()

	// TODO make the region an env variable.
	config := &aws.Config{Region: "us-west-1", Credentials: creds}
	inst := &awsProvider{ec2: ec2.New(config), db: dynamodb.New(config)}

	err := inst.init()

	return inst, err
}

// AWS provisioned IOPS range is 100 - 20000.
func mapIops(cos api.VolumeCos) int64 {
	if cos < 3 {
		return 1000
	} else if cos < 7 {
		return 10000
	} else {
		return 20000
	}
}

func (self *awsProvider) get(volumeID string) (*awsVolume, error) {
	params := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Key": {
				B:    []byte("PAYLOAD"),
				BOOL: aws.Boolean(true),
				BS: [][]byte{
					[]byte("PAYLOAD"),
				},
				L: []*dynamodb.AttributeValue{
					{},
				},
				M: map[string]*dynamodb.AttributeValue{
					"Key": {},
				},
				N: aws.String("NumberAttributeValue"),
				NS: []*string{
					aws.String("NumberAttributeValue"),
				},
				NULL: aws.Boolean(true),
				S:    aws.String("StringAttributeValue"),
				SS: []*string{
					aws.String("StringAttributeValue"),
				},
			},
		},
		TableName: aws.String("TableName"),
		AttributesToGet: []*string{
			aws.String("AttributeName"),
		},
		ConsistentRead: aws.Boolean(true),
		ExpressionAttributeNames: map[string]*string{
			"Key": aws.String("AttributeName"),
		},
		ProjectionExpression:   aws.String("ProjectionExpression"),
		ReturnConsumedCapacity: aws.String("ReturnConsumedCapacity"),
	}
	resp, err := self.db.GetItem(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}

	v := &awsVolume{}
	// err = json.Unmarshal(b, v)
	return v, nil

	/*
		err := self.db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(AwsBucketName))
			b := bucket.Get([]byte(volumeID))

			if b == nil {
				return errors.New("no such volume ID")
			} else {
			}
		})

		return v, err
	*/
}

func (self *awsProvider) put(volumeID string, v *awsVolume) error {
	/*
		b, _ := json.Marshal(v)

		err := self.db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(AwsBucketName))
			err := bucket.Put([]byte(volumeID), b)
			return err
		})

		return err
	*/
	return nil
}

// Create a DB if one does not exist.  This is where we persist the
// Amazon instance ID, sdevice and volume ID mappings.
func (self *awsProvider) init() error {
	listParams := &dynamodb.ListTablesInput{
		ExclusiveStartTableName: aws.String(AwsTableName),
		Limit: aws.Long(1),
	}

	_, err := self.db.ListTables(listParams)
	if err == nil {
		return nil
	}

	// Assume table does not exist and re-create it.
	createParams := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("KeySchemaAttributeName"),
				AttributeType: aws.String("ScalarAttributeType"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("KeySchemaAttributeName"),
				KeyType:       aws.String("KeyType"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Long(1),
			WriteCapacityUnits: aws.Long(1),
		},
		TableName: aws.String(AwsTableName),
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("IndexName"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("KeySchemaAttributeName"),
						KeyType:       aws.String("KeyType"),
					},
				},
				Projection: &dynamodb.Projection{
					NonKeyAttributes: []*string{
						aws.String("NonKeyAttributeName"),
					},
					ProjectionType: aws.String("ProjectionType"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Long(1),
					WriteCapacityUnits: aws.Long(1),
				},
			},
		},
		LocalSecondaryIndexes: []*dynamodb.LocalSecondaryIndex{
			{
				IndexName: aws.String("IndexName"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("KeySchemaAttributeName"),
						KeyType:       aws.String("KeyType"),
					},
				},
				Projection: &dynamodb.Projection{
					NonKeyAttributes: []*string{
						aws.String("NonKeyAttributeName"),
					},
					ProjectionType: aws.String("ProjectionType"),
				},
			},
		},
		StreamSpecification: &dynamodb.StreamSpecification{
			StreamEnabled:  aws.Boolean(true),
			StreamViewType: aws.String("StreamViewType"),
		},
	}

	_, err = self.db.CreateTable(createParams)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}

		return err
	}

	return nil
}

func (self *awsProvider) String() string {
	return Name
}

func (self *awsProvider) Create(l api.VolumeLocator, opt *api.CreateOptions, spec *api.VolumeSpec) (api.VolumeID, error) {
	// TODO get this via an env variable.
	availabilityZone := "us-west-1a"
	sz := int64(spec.Size / (1024 * 1024 * 1024))
	iops := mapIops(spec.Cos)
	req := &ec2.CreateVolumeInput{
		AvailabilityZone: &availabilityZone,
		Size:             &sz,
		IOPS:             &iops}
	v, err := self.ec2.CreateVolume(req)
	if err != nil {
		fmt.Println(err)
		return api.VolumeID(""), err
	}

	err = self.put(*v.VolumeID, &awsVolume{spec: *spec})

	return api.VolumeID(*v.VolumeID), err
}

func (self *awsProvider) Attach(volumeID api.VolumeID) (string, error) {
	v, err := self.get(string(volumeID))
	if err != nil {
		return "", err
	}

	devMinor++
	device := fmt.Sprintf("/dev/ec2%v", int(devMinor))
	vol := string(volumeID)
	inst := string("")
	req := &ec2.AttachVolumeInput{
		Device:     &device,
		InstanceID: &inst,
		VolumeID:   &vol,
	}

	resp, err := self.ec2.AttachVolume(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	v.instanceID = inst
	v.attached = true
	err = self.put(string(volumeID), v)

	return *resp.Device, err
}

func (self *awsProvider) Mount(volumeID api.VolumeID, mountpath string) error {
	v, err := self.get(string(volumeID))
	if err != nil {
		return err
	}

	err = syscall.Mount(v.device, mountpath, "ext4", 0, "")
	if err != nil {
		fmt.Println(err)
		return err
	}

	v.mountpath = mountpath
	v.mounted = true
	err = self.put(string(volumeID), v)

	return err
}

func (self *awsProvider) Detach(volumeID api.VolumeID) error {
	v, err := self.get(string(volumeID))
	if err != nil {
		return err
	}

	vol := string(volumeID)
	inst := v.instanceID
	force := true
	req := &ec2.DetachVolumeInput{
		InstanceID: &inst,
		VolumeID:   &vol,
		Force:      &force,
	}

	_, err = self.ec2.DetachVolume(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	v.instanceID = inst
	v.attached = false
	err = self.put(string(volumeID), v)

	return err
}

func (self *awsProvider) Unmount(volumeID api.VolumeID, mountpath string) error {
	v, err := self.get(string(volumeID))
	if err != nil {
		return err
	}

	err = syscall.Unmount(v.mountpath, 0)
	if err != nil {
		fmt.Println(err)
		return err
	}

	v.mountpath = ""
	v.mounted = false
	err = self.put(string(volumeID), v)

	return err
}

func (self *awsProvider) Delete(volumeID api.VolumeID) error {
	return nil
}

func (self *awsProvider) Format(volumeID api.VolumeID) error {
	v, err := self.get(string(volumeID))
	if err != nil {
		return err
	}

	if !v.attached {
		return errors.New("volume must be attached")
	}

	if v.mounted {
		return errors.New("volume already mounted")
	}

	if v.formatted {
		return errors.New("volume already formatted")
	}

	cmd := "/sbin/mkfs." + string(v.spec.Format)
	_, err = exec.Command(cmd, v.device).Output()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// TODO validate output

	v.formatted = true
	err = self.put(string(volumeID), v)

	return err
}

func (self *awsProvider) Inspect(volumeIDs []api.VolumeID) (volume []api.Volume, err error) {
	return nil, nil
}

func (self *awsProvider) Enumerate(locator api.VolumeLocator, labels api.Labels) (volumes []api.Volume, err error) {
	return nil, errors.New("Unsupported")
}

func (self *awsProvider) Snapshot(volumeID api.VolumeID, labels api.Labels) (snap api.SnapID, err error) {
	return "", errors.New("Unsupported")
}

func (self *awsProvider) SnapDelete(snapID api.SnapID) (err error) {
	return errors.New("Unsupported")
}

func (self *awsProvider) SnapInspect(snapID api.SnapID) (snap api.VolumeSnap, err error) {
	return api.VolumeSnap{}, errors.New("Unsupported")
}

func (self *awsProvider) SnapEnumerate(locator api.VolumeLocator, labels api.Labels) (snaps *[]api.SnapID, err error) {
	return nil, errors.New("Unsupported")
}

func (self *awsProvider) Stats(volumeID api.VolumeID) (stats api.VolumeStats, err error) {
	return api.VolumeStats{}, errors.New("Unsupported")
}

func (self *awsProvider) Alerts(volumeID api.VolumeID) (stats api.VolumeAlerts, err error) {
	return api.VolumeAlerts{}, errors.New("Unsupported")
}

func (self *awsProvider) Shutdown() {
	fmt.Printf("%s Shutting down", Name)
}

func init() {
	// Register ourselves as an openstorage volume driver.
	volume.Register(Name, Init)
}
