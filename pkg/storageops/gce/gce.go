package gce

import (
	"context"
	"fmt"
	"os"
	"path"
	"regexp"
	"sync"
	"time"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/oauth2/google"

	"github.com/Sirupsen/logrus"
	"github.com/libopenstorage/openstorage/pkg/storageops"
	compute "google.golang.org/api/compute/v1"
)

var notFoundRegex = regexp.MustCompile(`.*notFound`)

const googleDiskPrefix = "/dev/disk/by-id/google-"

type gceOps struct {
	inst    *instance
	service *compute.Service
	mutex   sync.Mutex
}

// instance stores the metadata of the running GCE instance
type instance struct {
	ID         string
	Name       string
	Hostname   string
	Zone       string
	Project    string
	InternalIP string
	ExternalIP string
	LBRequest  string
	ClientIP   string
	Error      string
}

type assigner struct {
	err error
}

func (a *assigner) assign(getVal func() (string, error)) string {
	if a.err != nil {
		return ""
	}

	s, err := getVal()
	if err != nil {
		a.err = err
	}

	return s
}

func getEnvValueStrict(key string) (val string, err error) {
	if val = os.Getenv(key); len(val) != 0 {
		return
	}

	err = fmt.Errorf("env variable %s is not set", key)
	return
}

// gceInfo fetches the GCE instance metadata from the metadata server
func gceInfo(inst *instance) {
	a := &assigner{}
	inst.ID = a.assign(metadata.InstanceID)
	inst.Zone = a.assign(metadata.Zone)
	inst.Name = a.assign(metadata.InstanceName)
	inst.Hostname = a.assign(metadata.Hostname)
	inst.Project = a.assign(metadata.ProjectID)
	inst.InternalIP = a.assign(metadata.InternalIP)
	inst.ExternalIP = a.assign(metadata.ExternalIP)

	if a.err != nil {
		inst.Error = a.err.Error()
	}
}

func gceInfoFromEnv(inst *instance) {
	a := &assigner{}
	inst.Name = a.assign(func() (string, error) { return getEnvValueStrict("GCE_INSTANCE_NAME") })
	inst.Zone = a.assign(func() (string, error) { return getEnvValueStrict("GCE_INSTANCE_ZONE") })
	inst.Project = a.assign(func() (string, error) { return getEnvValueStrict("GCE_INSTANCE_PROJECT") })

	if a.err != nil {
		inst.Error = a.err.Error()
	}
}

// IsDevMode checks if the pkg is invoked in developer mode where GCE credentials
// are set as env variables
func IsDevMode() bool {
	var i = new(instance)
	gceInfoFromEnv(i)
	return len(i.Error) == 0
}

// NewClient creates a new GCE operations client
func NewClient() (storageops.Ops, error) {
	var i = new(instance)
	if metadata.OnGCE() {
		gceInfo(i)
	} else if ok := IsDevMode(); ok {
		gceInfoFromEnv(i)
	} else {
		return nil, fmt.Errorf("instance is not running on GCE")
	}

	if len(i.Error) != 0 {
		return nil, fmt.Errorf("error while fetching instance information. Err: %v", i.Error)
	}

	c, err := google.DefaultClient(context.Background(), compute.ComputeScope)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with google api. Err: %v", err)
	}

	service, err := compute.New(c)
	if err != nil {
		return nil, fmt.Errorf("unable to create Compute service: %v", err)
	}

	return &gceOps{
		inst:    i,
		service: service,
	}, nil
}

func (s *gceOps) waitStatus(id string, desired string) error {
	actual := ""
	for retries, maxRetries := 0, storageops.ProviderOpsMaxRetries; actual != desired && retries < maxRetries; retries++ {
		d, err := s.service.Disks.Get(s.inst.Project, s.inst.Zone, id).Do()
		if err != nil {
			if ignore := notFoundRegex.MatchString(err.Error()); ignore {
				time.Sleep(storageops.ProviderOpsRetryInterval)
				continue
			}

			return err
		}

		if d == nil {
			return fmt.Errorf("expected one disk %v got none", id)
		}

		actual = d.Status
		if len(actual) == 0 {
			return fmt.Errorf("nil volume state for %v", id)
		}

		if actual == desired {
			break
		}

		time.Sleep(storageops.ProviderOpsRetryInterval)
	}

	if actual != desired {
		return fmt.Errorf("disk %v did not transition to %v current state %v",
			id, desired, actual)
	}

	return nil
}

func (s *gceOps) waitForDetach(
	diskURL string,
	timeout time.Duration,
) (err error) {
	var inst *compute.Instance

	interval := 5 * time.Second
	for elapsed := 0 * time.Second; elapsed < timeout; elapsed += interval {
		inst, err = s.describeinstance()
		if err != nil {
			return
		}

		found := false
		for _, d := range inst.Disks {
			if d.Source == diskURL {
				found = true
				break
			}
		}

		if found {
			time.Sleep(interval)
			continue
		} else {
			return
		}
	}

	err = fmt.Errorf("disk: %s is still attached to instance: %s", diskURL, s.inst.Name)
	return
}

// waitForAttach checks if given disk is attached to the local instance
func (s *gceOps) waitForAttach(
	disk *compute.Disk,
	timeout time.Duration,
) (devicePath string, err error) {
	interval := 2 * time.Second
	for elapsed := 0 * time.Second; elapsed < timeout; elapsed += interval {
		devicePath, err = s.DevicePath(disk.Name)
		if err == nil {
			return
		}

		time.Sleep(interval)
	}

	err = fmt.Errorf("disk: %s is not attached to instance: %s", disk.Name, s.inst.Name)
	return
}

func (s *gceOps) Name() string { return "gce" }

func (s *gceOps) describeinstance() (*compute.Instance, error) {
	return s.service.Instances.Get(s.inst.Project, s.inst.Zone, s.inst.Name).Do()
}

func (s *gceOps) rollbackCreate(id string, createErr error) error {
	logrus.Warnf("Rollback create volume %v, Error %v", id, createErr)
	err := s.Delete(id)
	if err != nil {
		logrus.Warnf("Rollback failed volume %v, Error %v", id, err)
	}
	return createErr
}

func (s *gceOps) available(v *compute.Disk) bool {
	return v.Status == "READY"
}

func (s *gceOps) ApplyTags(
	diskName string,
	labels map[string]string) (err error) {
	d, err := s.service.Disks.Get(s.inst.Project, s.inst.Zone, diskName).Do()
	if err != nil {
		return
	}

	var currentLabels map[string]string
	if len(d.Labels) == 0 {
		currentLabels = make(map[string]string)
	} else {
		currentLabels = d.Labels
	}

	for k, v := range labels {
		currentLabels[k] = v
	}

	rb := &compute.ZoneSetLabelsRequest{
		LabelFingerprint: d.LabelFingerprint,
		Labels:           currentLabels,
	}

	_, err = s.service.Disks.SetLabels(s.inst.Project, s.inst.Zone, d.Name, rb).Do()
	return
}

func (s *gceOps) Attach(diskName string) (devicePath string, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var d *compute.Disk
	d, err = s.service.Disks.Get(s.inst.Project, s.inst.Zone, diskName).Do()
	if err != nil {
		return
	}

	if len(d.Users) != 0 {
		err = fmt.Errorf("disk %s is already in use by %s", diskName, d.Users)
		return
	}

	diskURL := d.SelfLink
	rb := &compute.AttachedDisk{
		DeviceName: d.Name,
		Source:     diskURL,
	}

	_, err = s.service.Instances.AttachDisk(
		s.inst.Project,
		s.inst.Zone,
		s.inst.Name,
		rb).Do()
	if err != nil {
		return
	}

	devicePath, err = s.waitForAttach(d, time.Minute)
	if err != nil {
		return
	}

	return
}

func (s *gceOps) Create(
	template interface{},
	labels map[string]string,
) (interface{}, error) {
	v, ok := template.(*compute.Disk)
	if !ok {
		return nil, storageops.NewStorageError(storageops.ErrVolInval,
			"Invalid volume template given", "")
	}

	newDisk := &compute.Disk{
		Description:    "Disk created by openstorage",
		Labels:         labels,
		Name:           v.Name,
		SizeGb:         v.SizeGb,
		SourceImage:    v.SourceImage,
		SourceSnapshot: v.SourceSnapshot,
		Type:           v.Type,
		Zone:           s.inst.Zone,
	}

	resp, err := s.service.Disks.Insert(s.inst.Project, s.inst.Zone, newDisk).Do()
	if err != nil {
		return nil, err
	}

	if err = s.waitStatus(newDisk.Name, "READY"); err != nil {
		return nil, s.rollbackCreate(resp.Name, err)
	}

	d, err := s.service.Disks.Get(s.inst.Project, newDisk.Zone, newDisk.Name).Do()
	if err != nil {
		return nil, err
	}

	return d, err
}

func (s *gceOps) Delete(id string) (err error) {
	_, err = s.service.Disks.Delete(s.inst.Project, s.inst.Zone, id).Do()
	return
}

func (s *gceOps) Detach(devicePath string) (err error) {
	_, err = s.service.Instances.DetachDisk(
		s.inst.Project,
		s.inst.Zone,
		s.inst.Name,
		devicePath).Do()
	if err != nil {
		return
	}

	var d *compute.Disk
	d, err = s.service.Disks.Get(s.inst.Project, s.inst.Zone, devicePath).Do()
	if err != nil {
		return
	}

	err = s.waitForDetach(d.SelfLink, time.Minute)
	if err != nil {
		return
	}

	return
}

func (s *gceOps) DeviceMappings() (map[string]string, error) {
	instance, err := s.describeinstance()
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, d := range instance.Disks {
		if d.Boot {
			continue
		}

		m[fmt.Sprintf("%s%s", googleDiskPrefix, d.DeviceName)] = path.Base(d.Source)
	}

	return m, nil
}

func (s *gceOps) DevicePath(diskName string) (devicePath string, err error) {
	d, err := s.service.Disks.Get(s.inst.Project, s.inst.Zone, diskName).Do()
	if err != nil {
		return
	}

	if len(d.Users) == 0 {
		err = storageops.NewStorageError(storageops.ErrVolDetached,
			fmt.Sprintf("Disk: %s is detached", d.Name), s.inst.Name)
		return
	}

	var inst *compute.Instance
	inst, err = s.describeinstance()
	if err != nil {
		return
	}

	for _, instDisk := range inst.Disks {
		if instDisk.Source == d.SelfLink {
			devicePath = fmt.Sprintf("%s%s", googleDiskPrefix, instDisk.DeviceName)
			return
		}
	}

	err = storageops.NewStorageError(storageops.ErrVolAttachedOnRemoteNode,
		fmt.Sprintf("disk %s is not attached on: %s (Attached on: %v)",
			d.Name, s.inst.Name, d.Users),
		s.inst.Name)
	return
}

func (s *gceOps) Enumerate(
	volumeIds []*string,
	labels map[string]string,
	setIdentifier string,
) (map[string][]interface{}, error) {
	sets := make(map[string][]interface{})
	ctx := context.Background()
	found := false

	req := s.service.Disks.List(s.inst.Project, s.inst.Zone)
	if err := req.Pages(ctx, func(page *compute.DiskList) error {
		for _, disk := range page.Items {
			if len(setIdentifier) == 0 {
				storageops.AddElementToMap(sets, disk, storageops.SetIdentifierNone)
			} else {
				found = false
				for key := range disk.Labels {
					if key == setIdentifier {
						storageops.AddElementToMap(sets, disk, key)
						found = true
						break
					}
				}

				if !found {
					storageops.AddElementToMap(sets, disk, storageops.SetIdentifierNone)
				}
			}
		}

		return nil
	}); err != nil {
		logrus.Errorf("failed to list disks: %v", err)
		return nil, err
	}

	return sets, nil
}

func (s *gceOps) FreeDevices(
	blockDeviceMappings []interface{},
	rootDeviceName string,
) ([]string, error) {
	return nil, fmt.Errorf("function not implemented")
}

func (s *gceOps) GetDeviceID(disk interface{}) string {
	d, ok := disk.(*compute.Disk)
	if !ok {
		logrus.Errorf("Invalid volume given to GetDeviceID API")
		return ""
	}

	return d.Name
}

func (s *gceOps) Inspect(diskNames []*string) (disks []interface{}, err error) {
	for _, id := range diskNames {
		var d *compute.Disk
		d, err = s.service.Disks.Get(s.inst.Project, s.inst.Zone, *id).Do()
		if err != nil {
			return
		}

		disks = append(disks, d)
	}

	return
}

func (s *gceOps) RemoveTags(
	diskName string,
	labels map[string]string,
) (err error) {
	d, err := s.service.Disks.Get(s.inst.Project, s.inst.Zone, diskName).Do()
	if err != nil {
		return
	}

	if len(d.Labels) != 0 {
		currentLabels := d.Labels
		for k := range labels {
			delete(currentLabels, k)
		}

		rb := &compute.ZoneSetLabelsRequest{
			LabelFingerprint: d.LabelFingerprint,
			Labels:           currentLabels,
		}

		_, err = s.service.Disks.SetLabels(s.inst.Project, s.inst.Zone, d.Name, rb).Do()
	}

	return
}

func (s *gceOps) Snapshot(
	disk string,
	readonly bool,
) (interface{}, error) {
	rb := &compute.Snapshot{
		Name: fmt.Sprintf("snap-%d%02d%02d", time.Now().Year(), time.Now().Month(), time.Now().Day()),
	}
	return s.service.Disks.CreateSnapshot(s.inst.Project, s.inst.Zone, disk, rb).Do()
}

func (s *gceOps) Tags(diskName string) (labels map[string]string, err error) {
	d, err := s.service.Disks.Get(s.inst.Project, s.inst.Zone, diskName).Do()
	if err != nil {
		return
	}

	labels = d.Labels
	return
}
