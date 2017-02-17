package aws

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"go.pedge.io/dlog"
	"go.pedge.io/proto/time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/chaos"
	"github.com/libopenstorage/openstorage/pkg/device"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers/common"
	"github.com/portworx/kvdb"
)

const (
	// Name of the driver
	Name = "aws"
	// Type of the driver
	Type = api.DriverType_DRIVER_TYPE_BLOCK
	// AwsDBKey for openstorage
	AwsDBKey = "OpenStorageAWSKey"
)

var (
	koStrayCreate = chaos.Add("aws", "create", "create in driver before DB")
	koStrayDelete = chaos.Add("aws", "delete", "create in driver before DB")
)

// Metadata for the driver
type Metadata struct {
	zone     string
	instance string
}

// Driver implements VolumeDriver interface
type Driver struct {
	volume.StatsDriver
	volume.StoreEnumerator
	volume.IODriver
	*device.SingleLetter
	md        *Metadata
	ec2       *ec2.EC2
	devPrefix string
}

// Init aws volume driver metadata.
func Init(params map[string]string) (volume.VolumeDriver, error) {
	zone, err := metadata("placement/availability-zone")
	if err != nil {
		return nil, err
	}
	instance, err := metadata("instance-id")
	if err != nil {
		return nil, err
	}
	dlog.Infof("AWS instance %v zone %v", instance, zone)

	accessKey, ok := params["AWS_ACCESS_KEY_ID"]
	if !ok {
		if accessKey = os.Getenv("AWS_ACCESS_KEY_ID"); accessKey == "" {
			return nil, fmt.Errorf("AWS_ACCESS_KEY_ID environment variable must be set")
		}
	}
	secretKey, ok := params["AWS_SECRET_ACCESS_KEY"]
	if !ok {
		if secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY"); secretKey == "" {
			return nil, fmt.Errorf("AWS_SECRET_ACCESS_KEY environment variable must be set")
		}
	}

	creds := credentials.NewStaticCredentials(accessKey, secretKey, "")
	region := zone[:len(zone)-1]
	d := &Driver{
		StatsDriver: volume.StatsNotSupported,
		ec2: ec2.New(
			session.New(
				&aws.Config{
					Region:      &region,
					Credentials: creds,
				},
			),
		),
		md: &Metadata{
			zone:     zone,
			instance: instance,
		},
		IODriver:        volume.IONotSupported,
		StoreEnumerator: common.NewDefaultStoreEnumerator(Name, kvdb.Instance()),
	}
	devPrefix, letters, err := d.freeDevices()
	if err != nil {
		return nil, err
	}
	d.SingleLetter, err = device.NewSingleLetter(devPrefix, letters)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// freeDevices returns list of available device IDs.
func (d *Driver) freeDevices() (string, string, error) {
	initial := []byte("fghijklmnop")
	free := make([]byte, len(initial))
	self, err := d.describe()
	if err != nil {
		return "", "", err
	}
	devPrefix := "/dev/sd"
	for _, dev := range self.BlockDeviceMappings {
		if dev.DeviceName == nil {
			return "", "", fmt.Errorf("Nil device name")
		}
		devName := *dev.DeviceName

		// per AWS docs EC instances have the root mounted at /dev/sda1, this label should be skipped
		if devName == "/dev/sda1" {
			continue
		}
		if !strings.HasPrefix(devName, devPrefix) {
			devPrefix = "/dev/xvd"
			if !strings.HasPrefix(devName, devPrefix) {
				return "", "", fmt.Errorf("bad device name %q", devName)
			}
		}
		letter := devName[len(devPrefix):]
		if len(letter) != 1 {
			return "", "", fmt.Errorf("too many letters %q", devName)
		}
		index := letter[0] - 'f'
		if index > ('p' - 'f') {
			continue
		}
		initial[index] = '0'
	}
	count := 0
	for _, b := range initial {
		if b != '0' {
			free[count] = b
			count++
		}
	}
	return devPrefix, string(free[:count]), nil
}

// mapCos translates a CoS specified in spec to a volume.
func mapCos(cos uint32) (*int64, *string) {
	var iops int64
	var volType string
	switch {
	case cos < 2:
		iops, volType = 0, opsworks.VolumeTypeGp2
	case cos < 7:
		iops, volType = 10000, opsworks.VolumeTypeIo1
	default:
		iops, volType = 20000, opsworks.VolumeTypeIo1
	}
	return &iops, &volType
}

// metadata retrieves instance metadata specified by key.
func metadata(key string) (string, error) {
	client := http.Client{Timeout: time.Second * 10}
	url := "http://169.254.169.254/latest/meta-data/" + key
	res, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		err = fmt.Errorf("Code %d returned for url %s", res.StatusCode, url)
		return "", fmt.Errorf("Error querying AWS metadata for key %s: %v", key, err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Error querying AWS metadata for key %s: %v", key, err)
	}
	if len(body) == 0 {
		return "", fmt.Errorf("Failed to retrieve AWS metadata for key %s: %v", key, err)
	}
	return string(body), nil
}

// describe retrieves running instance desscription.
func (d *Driver) describe() (*ec2.Instance, error) {
	request := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{&d.md.instance},
	}
	out, err := d.ec2.DescribeInstances(request)
	if err != nil {
		return nil, err
	}
	if len(out.Reservations) != 1 {
		return nil, fmt.Errorf("DescribeInstances(%v) returned %v reservations, expect 1",
			d.md.instance, len(out.Reservations))
	}
	if len(out.Reservations[0].Instances) != 1 {
		return nil, fmt.Errorf("DescribeInstances(%v) returned %v Reservations, expect 1",
			d.md.instance, len(out.Reservations[0].Instances))
	}
	return out.Reservations[0].Instances[0], nil
}

// Name returns the name of the driver
func (d *Driver) Name() string {
	return Name
}

// Type returns the type of the driver
func (d *Driver) Type() api.DriverType {
	return Type
}

// Status returns the current status
func (d *Driver) Status() [][2]string {
	return [][2]string{}
}

// Create creates a new volume
func (d *Driver) Create(
	locator *api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec,
) (string, error) {
	var snapID *string
	// Spec size is in bytes, translate to GiB.
	sz := int64(spec.Size / (1024 * 1024 * 1024))
	iops, volType := mapCos(uint32(spec.Cos))
	if source != nil && string(source.Parent) != "" {
		id := string(source.Parent)
		snapID = &id
	}
	dryRun := false
	encrypted := false
	req := &ec2.CreateVolumeInput{
		AvailabilityZone: &d.md.zone,
		DryRun:           &dryRun,
		Encrypted:        &encrypted,
		Size:             &sz,
		VolumeType:       volType,
		SnapshotId:       snapID,
	}
	// Gp2 Volumes don't support the iops parameter
	if *volType != opsworks.VolumeTypeGp2 {
		req.Iops = iops
	}
	vol, err := d.ec2.CreateVolume(req)
	if err != nil {
		dlog.Warnf("Failed in CreateVolumeRequest :%v", err)
		return "", err
	}
	volume := common.NewVolume(
		*vol.VolumeId,
		api.FSType_FS_TYPE_EXT4,
		locator,
		source,
		spec,
	)
	err = d.UpdateVol(volume)
	if err != nil {
		return "", err
	}
	if err = d.waitStatus(volume.Id, ec2.VolumeStateAvailable); err != nil {
		return "", err
	}
	if _, err := d.Attach(volume.Id, nil); err != nil {
		return "", err
	}

	dlog.Infof("aws preparing volume %s...", *vol.VolumeId)

	if err := d.Format(volume.Id); err != nil {
		return "", err
	}
	if err := d.Detach(volume.Id); err != nil {
		return "", err
	}

	return volume.Id, err
}

// merge volume properties from aws into volume.
func (d *Driver) merge(v *api.Volume, aws *ec2.Volume) {
	v.AttachedOn = ""
	v.State = api.VolumeState_VOLUME_STATE_DETACHED
	v.DevicePath = ""

	switch *aws.State {
	case ec2.VolumeStateAvailable:
		v.Status = api.VolumeStatus_VOLUME_STATUS_UP
	case ec2.VolumeStateCreating, ec2.VolumeStateDeleting:
		v.State = api.VolumeState_VOLUME_STATE_PENDING
		v.Status = api.VolumeStatus_VOLUME_STATUS_DOWN
	case ec2.VolumeStateDeleted:
		v.State = api.VolumeState_VOLUME_STATE_DELETED
		v.Status = api.VolumeStatus_VOLUME_STATUS_DOWN
	case ec2.VolumeStateError:
		v.State = api.VolumeState_VOLUME_STATE_ERROR
		v.Status = api.VolumeStatus_VOLUME_STATUS_DOWN
	case ec2.VolumeStateInUse:
		v.Status = api.VolumeStatus_VOLUME_STATUS_UP
		if aws.Attachments != nil && len(aws.Attachments) != 0 {
			if aws.Attachments[0].InstanceId != nil {
				v.AttachedOn = *aws.Attachments[0].InstanceId
			}
			if aws.Attachments[0].State != nil {
				v.State = d.volumeState(aws.Attachments[0].State)
			}
			if aws.Attachments[0].Device != nil {
				v.DevicePath = *aws.Attachments[0].Device
			}
		}
	}
}

func (d *Driver) waitStatus(volumeID string, desired string) error {

	id := volumeID
	request := &ec2.DescribeVolumesInput{VolumeIds: []*string{&id}}
	actual := ""

	for retries, maxRetries := 0, 10; actual != desired && retries < maxRetries; retries++ {
		awsVols, err := d.ec2.DescribeVolumes(request)
		if err != nil {
			return err
		}
		if len(awsVols.Volumes) != 1 {
			return fmt.Errorf("expected one volume %v got %v",
				volumeID, len(awsVols.Volumes))
		}
		if awsVols.Volumes[0].State == nil {
			return fmt.Errorf("Nil volume state for %v", volumeID)
		}
		actual = *awsVols.Volumes[0].State
		if actual == desired {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if actual != desired {
		return fmt.Errorf("Volume %v failed to transition to  %v current state %v",
			volumeID, desired, actual)
	}
	return nil
}

func (d *Driver) waitAttachmentStatus(
	volumeID string,
	desired string,
	timeout time.Duration) error {

	id := volumeID
	request := &ec2.DescribeVolumesInput{VolumeIds: []*string{&id}}
	actual := ""
	interval := 2 * time.Second
	fmt.Printf("Waiting for state transition to %q", desired)
	for elapsed, runs := 0*time.Second, 0; actual != desired && elapsed < timeout; elapsed += interval {
		awsVols, err := d.ec2.DescribeVolumes(request)
		if err != nil {
			return err
		}
		if len(awsVols.Volumes) != 1 {
			return fmt.Errorf("expected one volume %v got %v",
				volumeID, len(awsVols.Volumes))
		}
		awsAttachment := awsVols.Volumes[0].Attachments
		if awsAttachment == nil || len(awsAttachment) == 0 {
			actual = ec2.VolumeAttachmentStateDetached
			if actual == desired {
				break
			}
			return fmt.Errorf("Nil attachment state for %v", volumeID)
		}
		actual = *awsAttachment[0].State
		if actual == desired {
			break
		}
		time.Sleep(interval)
		if (runs % 10) == 0 {
			fmt.Print(".")
		}
	}
	fmt.Printf("\n")
	if actual != desired {
		return fmt.Errorf("Volume %v failed to transition to  %v current state %v",
			volumeID, desired, actual)
	}
	return nil
}

func (d *Driver) devicePath(volumeID string) (string, error) {

	awsVolID := volumeID

	request := &ec2.DescribeVolumesInput{VolumeIds: []*string{&awsVolID}}
	awsVols, err := d.ec2.DescribeVolumes(request)
	if err != nil {
		return "", err
	}
	if awsVols == nil || len(awsVols.Volumes) == 0 {
		return "", fmt.Errorf("Failed to retrieve volume for ID %q", volumeID)

	}
	aws := awsVols.Volumes[0]
	if aws.Attachments == nil || len(aws.Attachments) == 0 {
		return "", fmt.Errorf("Invalid volume state, volume must be attached")
	}
	if aws.Attachments[0].InstanceId == nil {
		return "", fmt.Errorf("Unable to determine volume instance attachment")
	}
	if d.md.instance != *aws.Attachments[0].InstanceId {
		return "", fmt.Errorf("volume is attched on %q, it must be attached on %q",
			*aws.Attachments[0].InstanceId, d.md.instance)

	}
	if aws.Attachments[0].State == nil {
		return "", fmt.Errorf("Unable to determine volume attachment state")
	}
	if *aws.Attachments[0].State != ec2.VolumeAttachmentStateAttached {
		return "", fmt.Errorf("Invalid volume state %q, volume must be attached",
			*aws.Attachments[0].State)
	}
	if aws.Attachments[0].Device == nil {
		return "", fmt.Errorf("Unable to determine volume attachment path")
	}
	dev := strings.TrimPrefix(*aws.Attachments[0].Device, "/dev/sd")
	if dev != *aws.Attachments[0].Device {
		dev = "/dev/xvd" + dev
	}
	return dev, nil
}

// Inspect insepcts a volume
func (d *Driver) Inspect(volumeIDs []string) ([]*api.Volume, error) {
	vols, err := d.StoreEnumerator.Inspect(volumeIDs)
	if err != nil {
		return nil, err
	}
	ids := make([]*string, len(vols))
	for i, v := range vols {
		id := v.Id
		ids[i] = &id
	}
	request := &ec2.DescribeVolumesInput{VolumeIds: ids}
	awsVols, err := d.ec2.DescribeVolumes(request)
	if err != nil {
		return nil, err
	}
	if awsVols == nil || (len(awsVols.Volumes) != len(vols)) {
		return nil, fmt.Errorf("AwsVols (%v) do not match recorded vols (%v)", awsVols, vols)
	}
	for i, v := range awsVols.Volumes {
		if string(vols[i].Id) != *v.VolumeId {
			d.merge(vols[i], v)
		}
	}
	return vols, nil
}

// Delete deletes a volume
func (d *Driver) Delete(volumeID string) error {
	dryRun := false
	id := volumeID
	req := &ec2.DeleteVolumeInput{
		VolumeId: &id,
		DryRun:   &dryRun,
	}
	_, err := d.ec2.DeleteVolume(req)
	if err != nil {
		return err
	}
	err = d.DeleteVol(volumeID)
	return err
}

// Snapshot takes a snapshot of a source volume
func (d *Driver) Snapshot(volumeID string, readonly bool, locator *api.VolumeLocator) (string, error) {
	dryRun := false
	vols, err := d.StoreEnumerator.Inspect([]string{volumeID})
	if err != nil {
		return "", err
	}
	if len(vols) != 1 {
		return "", fmt.Errorf("Failed to inspect %v len %v", volumeID, len(vols))
	}
	awsID := volumeID
	request := &ec2.CreateSnapshotInput{
		VolumeId: &awsID,
		DryRun:   &dryRun,
	}
	snap, err := d.ec2.CreateSnapshot(request)
	chaos.Now(koStrayCreate)
	vols[0].Id = *snap.SnapshotId
	vols[0].Source = &api.Source{Parent: volumeID}
	vols[0].Locator = locator
	vols[0].Ctime = prototime.Now()

	chaos.Now(koStrayCreate)
	if err = d.CreateVol(vols[0]); err != nil {
		return "", err
	}
	return vols[0].Id, nil
}

// Attach attaches a volume
func (d *Driver) Attach(volumeID string, attachOptions map[string]string) (path string, err error) {
	volume, err := d.GetVol(volumeID)
	if err != nil {
		return "", fmt.Errorf("Volume %s could not be located", volumeID)
	}

	dryRun := false
	device, err := d.Assign()
	if err != nil {
		return "", err
	}
	awsVolID := volumeID
	req := &ec2.AttachVolumeInput{
		DryRun:     &dryRun,
		Device:     &device,
		InstanceId: &d.md.instance,
		VolumeId:   &awsVolID,
	}
	resp, err := d.ec2.AttachVolume(req)
	if err != nil {
		return "", err
	}
	if err = d.waitAttachmentStatus(volumeID, ec2.VolumeAttachmentStateAttached, time.Minute*5); err != nil {
		return "", err
	}

	volume.DevicePath = *resp.Device
	if err := d.UpdateVol(volume); err != nil {
		return "", err
	}

	return *resp.Device, nil
}

func (d *Driver) volumeState(ec2VolState *string) api.VolumeState {
	if ec2VolState == nil {
		return api.VolumeState_VOLUME_STATE_DETACHED
	}
	switch *ec2VolState {
	case ec2.VolumeAttachmentStateAttached:
		return api.VolumeState_VOLUME_STATE_ATTACHED
	case ec2.VolumeAttachmentStateDetached:
		return api.VolumeState_VOLUME_STATE_DETACHED
	case ec2.VolumeAttachmentStateAttaching, ec2.VolumeAttachmentStateDetaching:
		return api.VolumeState_VOLUME_STATE_PENDING
	default:
		dlog.Warnf("Failed to translate EC2 volume status %v", ec2VolState)
	}
	return api.VolumeState_VOLUME_STATE_ERROR
}

// Format formats a device
func (d *Driver) Format(volumeID string) error {
	volume, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Failed to locate volume %q", volumeID)
	}

	// XXX: determine mount state
	devicePath, err := d.devicePath(volumeID)
	if err != nil {
		return err
	}
	cmd := "/sbin/mkfs." + volume.Spec.Format.SimpleString()
	o, err := exec.Command(cmd, devicePath).Output()
	if err != nil {
		dlog.Warnf("Failed to run command %v %v: %v", cmd, devicePath, o)
		return err
	}
	volume.Format = volume.Spec.Format
	return d.UpdateVol(volume)
}

// Detach detaches a volume from host
func (d *Driver) Detach(volumeID string) error {
	force := false
	awsVolID := volumeID
	device := ""
	volume, err := d.GetVol(volumeID)
	if err != nil {
		dlog.Warnf("Volume %s could not be located, attempting to detach anyway", volumeID)
	} else {
		device = volume.DevicePath
	}
	req := &ec2.DetachVolumeInput{
		InstanceId: &d.md.instance,
		VolumeId:   &awsVolID,
		Force:      &force,
	}
	if _, err := d.ec2.DetachVolume(req); err != nil {

		return err
	}

	if "" != device {
		if err := d.Release(device); err != nil {
			return err
		}
	}

	return d.waitAttachmentStatus(volumeID, ec2.VolumeAttachmentStateDetached, time.Minute*5)
}

// MountedAt returns the volume mounted at specific path
func (d *Driver) MountedAt(mountpath string) string {
	return ""
}

// Mount mounts a volume at a given path
func (d *Driver) Mount(volumeID string, mountpath string) error {
	volume, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Failed to locate volume %q", volumeID)
	}
	devicePath, err := d.devicePath(volumeID)
	if err != nil {
		return err
	}
	err = syscall.Mount(devicePath, mountpath, volume.Spec.Format.SimpleString(), 0, "")
	if err != nil {
		return err
	}
	return nil
}

// Unmount unmounts a volume
func (d *Driver) Unmount(volumeID string, mountpath string) error {
	// XXX:  determine if valid mount path
	err := syscall.Unmount(mountpath, 0)
	return err
}

// Shutdown stops the driver
func (d *Driver) Shutdown() {
	dlog.Printf("%s Shutting down", Name)
}

// Set updates fields on a volume
func (d *Driver) Set(volumeID string, locator *api.VolumeLocator, spec *api.VolumeSpec) error {
	return volume.ErrNotSupported
}
