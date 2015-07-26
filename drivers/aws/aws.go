package ebs

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	Name = "aws"
)

var (
	r        *rand.Rand
	devMinor int32
)

// Implements the open storage volume interface.
type awsProvider struct {
	ec2 *ec2.EC2
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	// Initialize the EC2 interface.
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.EC2RoleProvider{},
		})

	inst := &awsProvider{ec2: ec2.New(&aws.Config{
		Region:      "us-west-1",
		Credentials: creds,
	}),
	}

	return inst, nil
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

func (self *awsProvider) String() string {
	return Name
}

func (self *awsProvider) Create(l api.VolumeLocator, opt *api.CreateOptions, spec *api.VolumeSpec) (api.VolumeID, error) {
	availabilityZone := "us-west-1a"
	sz := int64(spec.Size / (1024 * 1024 * 1024))
	iops := mapIops(spec.Cos)
	req := &ec2.CreateVolumeInput{
		AvailabilityZone: &availabilityZone,
		Size:             &sz,
		IOPS:             &iops}
	v, err := self.ec2.CreateVolume(req)

	return api.VolumeID(*v.VolumeID), err
}

func (self *awsProvider) AttachInfo(volInfo *api.VolumeInfo) (int32, string, error) {
	s := fmt.Sprintf("/tmp/gdd_%v", int(devMinor))
	return devMinor, s, nil
}

func (self *awsProvider) Attach(volInfo api.VolumeID, path string) (string, error) {
	devMinor++
	s := fmt.Sprintf("/tmp/gdd_%v", int(devMinor))
	os.Create(s)
	return s, nil
}

func (self *awsProvider) Detach(volID api.VolumeID) error {
	return nil
}

func (self *awsProvider) Delete(volID api.VolumeID) error {
	return nil
}

func (self *awsProvider) Format(id api.VolumeID) error {
	return nil
}

func (self *awsProvider) Inspect(ids []api.VolumeID) ([]api.Volume, error) {
	return nil, nil
}

func (self *awsProvider) Enumerate(locator api.VolumeLocator, labels api.Labels) []api.Volume {
	return nil
}

func (self *awsProvider) Snapshot(volID api.VolumeID, labels api.Labels) (snap api.SnapID, err error) {
	return "", nil
}

func (self *awsProvider) SnapDelete(snapID api.SnapID) (err error) {
	return nil
}

func (self *awsProvider) SnapInspect(snapID api.SnapID) (snap api.VolumeSnap, err error) {
	return api.VolumeSnap{}, nil
}

func (self *awsProvider) SnapEnumerate(locator api.VolumeLocator, labels api.Labels) *[]api.SnapID {
	return nil
}

func (self *awsProvider) Stats(volID api.VolumeID) (stats api.VolumeStats, err error) {
	return api.VolumeStats{}, nil
}

func (self *awsProvider) Alerts(volID api.VolumeID) (stats api.VolumeAlerts, err error) {
	return api.VolumeAlerts{}, nil
}

func (self *awsProvider) Shutdown() {
	fmt.Printf("%s Shutting down", Name)
}

func init() {
	// Register ourselves as an openstorage volume driver.
	volume.Register(Name, Init)
}
