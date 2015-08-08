package aws

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/chaos"
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
	md        *Metadata
	ec2       *ec2.EC2
	devices   string
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

	creds := credentials.NewEnvCredentials()
	inst := &Driver{
		ec2: ec2.New(&aws.Config{
			Region:      &zone,
			Credentials: creds,
		}),
		md: &Metadata{
			zone:     zone,
			instance: instance,
		},
		devices: "abcdefghijklmnopqrstuvwxyz",
	}
	return inst, nil
}

// mapCos translates a CoS specified in spec to a volume.
func mapCos(cos api.VolumeCos) (int64, string) {
	// AWS provisioned IOPS range is 100 - 20000.
	if cos < 5 {
		// General purpose SSDs don't have provisioned IOPS
		return 0, opsworks.VolumeTypeGp2
	}
	if cos < 7 {
		return 10000, opsworks.VolumeTypeIo1
	}
	return 20000, opsworks.VolumeTypeIo1
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

	return string(body), nil
}

func (d *Driver) assign() (string, error) {
	if len(d.devices) == 0 {
		return "", fmt.Errorf("No free device IDs")
	}
	device := d.devPrefix + d.devices[:1]
	d.devices = d.devices[1:]
	return device, nil
}

func (d *Driver) release(dev string) error {
	if !strings.HasPrefix(dev, d.devPrefix) {
		return fmt.Errorf("Invalid device %s", dev)
	}
	dev = dev[len(d.devPrefix):]
	d.devices += dev
	return nil
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
	opt *api.CreateOptions,
	spec *api.VolumeSpec) (api.VolumeID, error) {

	var snapID *string

	// Spec size is in bytes, translate to GiB.
	sz := int64(spec.Size / (1024 * 1024 * 1024))
	iops, volType := mapCos(spec.Cos)
	if string(opt.CreateFromSnap) != "" {
		id := string(opt.CreateFromSnap)
		snapID = &id
	}
	dryRun := false
	encrypted := false

	req := &ec2.CreateVolumeInput{
		AvailabilityZone: &d.md.zone,
		DryRun:           &dryRun,
		Encrypted:        &encrypted,
		Size:             &sz,
		IOPS:             &iops,
		VolumeType:       &volType,
		SnapshotID:       snapID,
	}

	vol, err := d.ec2.CreateVolume(req)
	if err != nil {
		return api.BadVolumeID, err
	}
	v := &api.Volume{
		ID:       api.VolumeID(*vol.VolumeID),
		Locator:  locator,
		Ctime:    time.Now(),
		Spec:     spec,
		LastScan: time.Now(),
		Format:   "none",
		State:    api.VolumeAvailable,
	}
	err = d.UpdateVol(v)
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
			if aws.Attachments[0].InstanceID != nil {
				v.AttachedOn = api.MachineID(*aws.Attachments[0].InstanceID)
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

func (d *Driver) devicePath(volumeID api.VolumeID) (string, error) {

	awsVolID := string(volumeID)

	request := &ec2.DescribeVolumesInput{VolumeIDs: []*string{&awsVolID}}
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
	if aws.Attachments[0].InstanceID == nil {
		return "", fmt.Errorf("Unable to determine volume instance attachment")
	}
	if d.md.instance != *aws.Attachments[0].InstanceID {
		return "", fmt.Errorf("volume is attched on %q, it must be attached on %q",
			*aws.Attachments[0].InstanceID, d.md.instance)

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
	return *aws.Attachments[0].Device, nil
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
	request := &ec2.DescribeVolumesInput{VolumeIDs: ids}
	awsVols, err := d.ec2.DescribeVolumes(request)
	if err != nil {
		return nil, err
	}
	if awsVols == nil || (len(awsVols.Volumes) != len(vols)) {
		return nil, fmt.Errorf("AwsVols (%v) do not match recorded vols (%v)", awsVols, vols)
	}
	for i, v := range awsVols.Volumes {
		if string(vols[i].ID) != *v.VolumeID {
			d.merge(&vols[i], v)
		}
	}
	return vols, nil
}

func (d *Driver) Delete(volumeID api.VolumeID) error {
	dryRun := false

	id := string(volumeID)
	req := &ec2.DeleteVolumeInput{
		VolumeID: &id,
		DryRun:   &dryRun,
	}
	_, err := d.ec2.DeleteVolume(req)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) Snapshot(volumeID api.VolumeID, labels api.Labels) (api.SnapID, error) {
	dryRun := false
	awsID := string(volumeID)
	request := &ec2.CreateSnapshotInput{
		VolumeID: &awsID,
		DryRun:   &dryRun,
	}
	snap, err := d.ec2.CreateSnapshot(request)
	chaos.Now(koStrayCreate)
	volSnap := &api.VolumeSnap{
		ID:         api.SnapID(*snap.SnapshotID),
		VolumeID:   volumeID,
		SnapLabels: labels,
		Ctime:      time.Now(),
	}
	chaos.Now(koStrayCreate)
	err = d.CreateSnap(volSnap)
	if err != nil {
		return api.BadSnapID, err
	}
	return volSnap.ID, nil
}

func (d *Driver) SnapDelete(snapID api.SnapID) error {
	return volume.ErrNotSupported
}

func (d *Driver) SnapInspect(snapID []api.SnapID) ([]api.VolumeSnap, error) {
	return []api.VolumeSnap{}, volume.ErrNotSupported
}

func (d *Driver) Stats(volumeID api.VolumeID) (api.VolumeStats, error) {
	return api.VolumeStats{}, volume.ErrNotSupported
}

func (d *Driver) Alerts(volumeID api.VolumeID) (api.VolumeAlerts, error) {
	return api.VolumeAlerts{}, volume.ErrNotSupported
}

func (d *Driver) Enumerate(locator api.VolumeLocator, labels api.Labels) ([]api.Volume, error) {
	return nil, volume.ErrNotSupported
}

func (d *Driver) SnapEnumerate(volIds []api.VolumeID, labels api.Labels) ([]api.VolumeSnap, error) {
	return nil, volume.ErrNotSupported
}

func (d *Driver) Attach(volumeID api.VolumeID) (path string, err error) {
	device, err := d.assign()
	if err != nil {
		return "", err
	}
	awsVolID := string(volumeID)
	req := &ec2.AttachVolumeInput{
		Device:     &device,
		InstanceID: &d.md.instance,
		VolumeID:   &awsVolID,
	}
	resp, err := d.ec2.AttachVolume(req)
	if err != nil {
		return "", err
	}
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
	_, err = exec.Command(cmd, devicePath).Output()
	if err != nil {
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
		InstanceID: &d.md.instance,
		VolumeID:   &awsVolID,
		Force:      &force,
	}
	_, err := d.ec2.DetachVolume(req)
	if err != nil {
		return err
	}
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
