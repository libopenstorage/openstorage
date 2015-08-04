package aws

import (
	"errors"
	"fmt"
	"os/exec"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/libopenstorage/kvdb"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	Name     = "aws"
	AwsDBKey = "OpenStorageAWSKey"
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
type awsDriver struct {
	db  kvdb.Kvdb
	ec2 *ec2.EC2
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	// Initialize the EC2 interface.
	creds := credentials.NewEnvCredentials()
	inst := &awsDriver{
		ec2: ec2.New(&aws.Config{
			Region:      "us-west-1",
			Credentials: creds,
		}),
	}

	inst.db = kvdb.Instance()

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

func (d *awsDriver) get(volumeID string) (*awsVolume, error) {
	v := &awsVolume{}
	key := AwsDBKey + "/" + volumeID
	_, err := d.db.GetVal(key, v)
	return v, err
}

func (d *awsDriver) put(volumeID string, v *awsVolume) error {
	key := AwsDBKey + "/" + volumeID
	_, err := d.db.Put(key, v, 0)
	return err
}

func (d *awsDriver) del(volumeID string) {
	key := AwsDBKey + "/" + volumeID
	d.db.Delete(key)
}

func (d *awsDriver) String() string {
	return Name
}

func (d *awsDriver) Create(l api.VolumeLocator, opt *api.CreateOptions, spec *api.VolumeSpec) (api.VolumeID, error) {
	availabilityZone := "us-west-1a"
	sz := int64(spec.Size / (1024 * 1024 * 1024))
	iops := mapIops(spec.Cos)
	req := &ec2.CreateVolumeInput{
		AvailabilityZone: &availabilityZone,
		Size:             &sz,
		IOPS:             &iops}
	v, err := d.ec2.CreateVolume(req)
	if err != nil {
		return api.VolumeID(""), err
	}

	// Persist the volume spec.  We use this for all subsequent operations on
	// this volume ID.
	err = d.put(string(*v.VolumeID), &awsVolume{spec: *spec})

	return api.VolumeID(*v.VolumeID), err
}

func (d *awsDriver) Inspect(volumeIDs []api.VolumeID) ([]api.Volume, error) {
	return nil, nil
}

func (d *awsDriver) Delete(volumeID api.VolumeID) error {
	return nil
}

func (d *awsDriver) Snapshot(volumeID api.VolumeID, labels api.Labels) (api.SnapID, error) {
	return "", volume.ErrNotSupported
}

func (d *awsDriver) SnapDelete(snapID api.SnapID) error {
	return volume.ErrNotSupported
}

func (d *awsDriver) SnapInspect(snapID []api.SnapID) ([]api.VolumeSnap, error) {
	return []api.VolumeSnap{}, volume.ErrNotSupported
}

func (d *awsDriver) Stats(volumeID api.VolumeID) (api.VolumeStats, error) {
	return api.VolumeStats{}, volume.ErrNotSupported
}

func (d *awsDriver) Alerts(volumeID api.VolumeID) (api.VolumeAlerts, error) {
	return api.VolumeAlerts{}, volume.ErrNotSupported
}

func (d *awsDriver) Enumerate(locator api.VolumeLocator, labels api.Labels) ([]api.Volume, error) {
	return nil, volume.ErrNotSupported
}

func (d *awsDriver) SnapEnumerate(volIds []api.VolumeID, labels api.Labels) ([]api.VolumeSnap, error) {
	return nil, volume.ErrNotSupported
}

func (d *awsDriver) Attach(volumeID api.VolumeID) (path string, err error) {
	v, err := d.get(string(volumeID))
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

	resp, err := d.ec2.AttachVolume(req)
	if err != nil {
		return "", err
	}

	v.instanceID = inst
	v.attached = true
	err = d.put(string(volumeID), v)

	return *resp.Device, err
}

func (d *awsDriver) Format(volumeID api.VolumeID) error {
	v, err := d.get(string(volumeID))
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
		return err
	}
	// XXX TODO validate output

	v.formatted = true
	err = d.put(string(volumeID), v)

	return err
}

func (d *awsDriver) Detach(volumeID api.VolumeID) error {
	v, err := d.get(string(volumeID))
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

	_, err = d.ec2.DetachVolume(req)
	if err != nil {
		return err
	}

	v.instanceID = inst
	v.attached = false
	err = d.put(string(volumeID), v)

	return err
}

func (d *awsDriver) Mount(volumeID api.VolumeID, mountpath string) error {
	v, err := d.get(string(volumeID))
	if err != nil {
		return err
	}

	err = syscall.Mount(v.device, mountpath, string(v.spec.Format), 0, "")
	if err != nil {
		return err
	}

	v.mountpath = mountpath
	v.mounted = true
	err = d.put(string(volumeID), v)

	return err
}

func (d *awsDriver) Unmount(volumeID api.VolumeID, mountpath string) error {
	v, err := d.get(string(volumeID))
	if err != nil {
		return err
	}

	err = syscall.Unmount(v.mountpath, 0)
	if err != nil {
		return err
	}

	v.mountpath = ""
	v.mounted = false
	err = d.put(string(volumeID), v)

	return err
}

func (d *awsDriver) Shutdown() {
	log.Printf("%s Shutting down", Name)
}

func init() {
	// Register ourselves as an openstorage volume driver.
	volume.Register(Name, volume.Block, Init)
}
