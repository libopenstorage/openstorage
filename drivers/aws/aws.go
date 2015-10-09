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

	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"

	"github.com/portworx/kvdb"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/chaos"
	"github.com/libopenstorage/openstorage/pkg/device"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	Name     = "aws"
	Type     = volume.Block
	AwsDBKey = "OpenStorageAWSKey"
)

type Metadata struct {
	zone     string
	instance string
}

var (
	koStrayCreate chaos.ID
	koStrayDelete chaos.ID
)

// Driver implements VolumeDriver interface
type Driver struct {
	*volume.DefaultEnumerator
	*device.SingleLetter
	md        *Metadata
	ec2       *ec2.EC2
	devPrefix string
}

// Init aws volume driver metadata.
func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	zone, err := metadata("placement/availability-zone")
	if err != nil {
		return nil, err
	}
	instance, err := metadata("instance-id")
	if err != nil {
		return nil, err
	}
	log.Infof("AWS instance %v zone %v", instance, zone)
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
		ec2: ec2.New(&aws.Config{
			Region:      &region,
			Credentials: creds,
		}),
		md: &Metadata{
			zone:     zone,
			instance: instance,
		},
		DefaultEnumerator: volume.NewDefaultEnumerator(Name, kvdb.Instance()),
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
func mapCos(cos api.VolumeCos) (*int64, *string) {
	volType := opsworks.VolumeTypeIo1
	if cos < 2 {
		// General purpose SSDs don't have provisioned IOPS
		volType = opsworks.VolumeTypeGp2
		return nil, &volType
	}
	// AWS provisioned IOPS range is 100 - 20000.
	var iops int64
	if cos < 7 {
		iops = 10000
	} else {
		iops = 20000
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

// String is a description of this driver.
func (d *Driver) String() string {
	return Name
}

// Type returns aws as a Block driver.
func (d *Driver) Type() volume.DriverType {
	return Type
}

// Status diagnostic information
func (v *Driver) Status() [][2]string {
	return [][2]string{}
}

// Create aws volume from spec.
func (d *Driver) Create(
	locator api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec) (api.VolumeID, error) {

	var snapID *string

	// Spec size is in bytes, translate to GiB.
	sz := int64(spec.Size / (1024 * 1024 * 1024))
	iops, volType := mapCos(spec.Cos)
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
		Iops:             iops,
		VolumeType:       volType,
		SnapshotId:       snapID,
	}

	vol, err := d.ec2.CreateVolume(req)
	if err != nil {
		log.Warnf("Failed in CreateVolumeRequest :%v", err)
		return api.BadVolumeID, err
	}
	v := &api.Volume{
		ID:       api.VolumeID(*vol.VolumeId),
		Locator:  locator,
		Ctime:    time.Now(),
		Spec:     spec,
		Source:   source,
		LastScan: time.Now(),
		Format:   "none",
		State:    api.VolumeAvailable,
		Status:   api.Up,
	}
	err = d.UpdateVol(v)
	err = d.waitStatus(v.ID, ec2.VolumeStateAvailable)
	return v.ID, err
}

// merge volume properties from aws into volume.
func (d *Driver) merge(v *api.Volume, aws *ec2.Volume) {
	v.AttachedOn = api.MachineID("")
	v.State = api.VolumeDetached
	v.DevicePath = ""

	switch *aws.State {
	case ec2.VolumeStateAvailable:
		v.Status = api.Up
	case ec2.VolumeStateCreating, ec2.VolumeStateDeleting:
		v.State = api.VolumePending
		v.Status = api.Down
	case ec2.VolumeStateDeleted:
		v.State = api.VolumeDeleted
		v.Status = api.Down
	case ec2.VolumeStateError:
		v.State = api.VolumeError
		v.Status = api.Down
	case ec2.VolumeStateInUse:
		v.Status = api.Up
		if aws.Attachments != nil && len(aws.Attachments) != 0 {
			if aws.Attachments[0].InstanceId != nil {
				v.AttachedOn = api.MachineID(*aws.Attachments[0].InstanceId)
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

func (d *Driver) waitStatus(volumeID api.VolumeID, desired string) error {

	id := string(volumeID)
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
	volumeID api.VolumeID,
	desired string,
	timeout time.Duration) error {

	id := string(volumeID)
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

func (d *Driver) devicePath(volumeID api.VolumeID) (string, error) {

	awsVolID := string(volumeID)

	request := &ec2.DescribeVolumesInput{VolumeIds: []*string{&awsVolID}}
	awsVols, err := d.ec2.DescribeVolumes(request)
	if err != nil {
		return "", err
	}
	if awsVols == nil || len(awsVols.Volumes) == 0 {
		return "", fmt.Errorf("Failed to retrieve volume for ID %q", string(volumeID))

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

func (d *Driver) Inspect(volumeIDs []api.VolumeID) ([]api.Volume, error) {
	vols, err := d.DefaultEnumerator.Inspect(volumeIDs)
	if err != nil {
		return nil, err
	}
	var ids []*string = make([]*string, len(vols))
	for i, v := range vols {
		id := string(v.ID)
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
		if string(vols[i].ID) != *v.VolumeId {
			d.merge(&vols[i], v)
		}
	}
	return vols, nil
}

func (d *Driver) Delete(volumeID api.VolumeID) error {
	dryRun := false
	id := string(volumeID)
	req := &ec2.DeleteVolumeInput{
		VolumeId: &id,
		DryRun:   &dryRun,
	}
	_, err := d.ec2.DeleteVolume(req)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) Snapshot(volumeID api.VolumeID, readonly bool, locator api.VolumeLocator) (api.VolumeID, error) {
	dryRun := false
	vols, err := d.DefaultEnumerator.Inspect([]api.VolumeID{volumeID})
	if err != nil {
		return api.BadVolumeID, err
	}
	if len(vols) != 1 {
		return api.BadVolumeID, fmt.Errorf("Failed to inspect %v len %v", volumeID, len(vols))
	}
	awsID := string(volumeID)
	request := &ec2.CreateSnapshotInput{
		VolumeId: &awsID,
		DryRun:   &dryRun,
	}
	snap, err := d.ec2.CreateSnapshot(request)
	chaos.Now(koStrayCreate)
	vols[0].ID = api.VolumeID(*snap.SnapshotId)
	vols[0].Source = &api.Source{Parent: volumeID}
	vols[0].Locator = locator
	vols[0].Ctime = time.Now()

	chaos.Now(koStrayCreate)
	err = d.CreateVol(&vols[0])
	if err != nil {
		return api.BadVolumeID, err
	}
	return vols[0].ID, nil
}

func (d *Driver) Stats(volumeID api.VolumeID) (api.Stats, error) {
	return api.Stats{}, volume.ErrNotSupported
}

func (d *Driver) Alerts(volumeID api.VolumeID) (api.Alerts, error) {
	return api.Alerts{}, volume.ErrNotSupported
}

func (d *Driver) Attach(volumeID api.VolumeID) (path string, err error) {
	dryRun := false
	device, err := d.Assign()
	if err != nil {
		return "", err
	}
	awsVolID := string(volumeID)
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
	err = d.waitAttachmentStatus(volumeID, ec2.VolumeAttachmentStateAttached, time.Minute*5)
	return *resp.Device, err
}

func (d *Driver) volumeState(ec2VolState *string) api.VolumeState {
	if ec2VolState == nil {
		return api.VolumeDetached
	}
	switch *ec2VolState {
	case ec2.VolumeAttachmentStateAttached:
		return api.VolumeAttached
	case ec2.VolumeAttachmentStateDetached:
		return api.VolumeDetached
	case ec2.VolumeAttachmentStateAttaching, ec2.VolumeAttachmentStateDetaching:
		return api.VolumePending
	default:
		log.Warnf("Failed to translate EC2 volume status %v", ec2VolState)
	}
	return api.VolumeError
}

func (d *Driver) Format(volumeID api.VolumeID) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Failed to locate volume %q", string(volumeID))
	}

	// XXX: determine mount state
	devicePath, err := d.devicePath(volumeID)
	if err != nil {
		return err
	}
	cmd := "/sbin/mkfs." + string(v.Spec.Format)
	o, err := exec.Command(cmd, devicePath).Output()
	if err != nil {
		log.Warnf("Failed to run command %v %v: %v", cmd, devicePath, o)
		return err
	}
	v.Format = v.Spec.Format
	err = d.UpdateVol(v)
	return err
}

func (d *Driver) Detach(volumeID api.VolumeID) error {
	force := false
	awsVolID := string(volumeID)
	req := &ec2.DetachVolumeInput{
		InstanceId: &d.md.instance,
		VolumeId:   &awsVolID,
		Force:      &force,
	}
	_, err := d.ec2.DetachVolume(req)
	if err != nil {
		return err
	}
	err = d.waitAttachmentStatus(volumeID, ec2.VolumeAttachmentStateDetached, time.Minute*5)
	return err
}

func (d *Driver) Mount(volumeID api.VolumeID, mountpath string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Failed to locate volume %q", string(volumeID))
	}
	devicePath, err := d.devicePath(volumeID)
	if err != nil {
		return err
	}
	err = syscall.Mount(devicePath, mountpath, string(v.Spec.Format), 0, "")
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) Unmount(volumeID api.VolumeID, mountpath string) error {
	// XXX:  determine if valid mount path
	err := syscall.Unmount(mountpath, 0)
	return err
}

func (d *Driver) Shutdown() {
	log.Printf("%s Shutting down", Name)
}

func init() {
	// Register ourselves as an openstorage volume driver.
	volume.Register(Name, Init)
	koStrayCreate = chaos.Add("aws", "create", "create in driver before DB")
	koStrayDelete = chaos.Add("aws", "delete", "create in driver before DB")
}
