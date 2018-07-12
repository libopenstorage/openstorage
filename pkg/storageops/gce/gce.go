package gce

import (
	"context"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/portworx/sched-ops/task"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

var notFoundRegex = regexp.MustCompile(`.*notFound`)

const googleDiskPrefix = "/dev/disk/by-id/google-"
const STATUS_READY = "READY"

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
}

// IsDevMode checks if the pkg is invoked in developer mode where GCE credentials
// are set as env variables
func IsDevMode() bool {
	var i = new(instance)
	err := gceInfoFromEnv(i)
	return err == nil
}

// NewClient creates a new GCE operations client
func NewClient() (storageops.Ops, error) {
	var i = new(instance)
	var err error
	if metadata.OnGCE() {
		err = gceInfo(i)
	} else if ok := IsDevMode(); ok {
		err = gceInfoFromEnv(i)
	} else {
		return nil, fmt.Errorf("instance is not running on GCE")
	}

	if err != nil {
		return nil, fmt.Errorf("error fetching instance info. Err: %v", err)
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

func (s *gceOps) Name() string { return "gce" }

func (s *gceOps) ApplyTags(
	diskName string,
	labels map[string]string) error {
	d, err := s.service.Disks.Get(s.inst.Project, s.inst.Zone, diskName).Do()
	if err != nil {
		return err
	}

	var currentLabels map[string]string
	if len(d.Labels) == 0 {
		currentLabels = make(map[string]string)
	} else {
		currentLabels = d.Labels
	}

	for k, v := range formatLabels(labels) {
		currentLabels[k] = v
	}

	rb := &compute.ZoneSetLabelsRequest{
		LabelFingerprint: d.LabelFingerprint,
		Labels:           currentLabels,
	}

	_, err = s.service.Disks.SetLabels(s.inst.Project, s.inst.Zone, d.Name, rb).Do()
	return err
}

func (s *gceOps) Attach(diskName string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var d *compute.Disk
	d, err := s.service.Disks.Get(s.inst.Project, s.inst.Zone, diskName).Do()
	if err != nil {
		return "", err
	}

	if len(d.Users) != 0 {
		return "", fmt.Errorf("disk %s is already in use by %s", diskName, d.Users)
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
		return "", err
	}

	devicePath, err := s.waitForAttach(d, time.Minute)
	if err != nil {
		return "", err
	}

	return devicePath, nil
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
		Labels:         formatLabels(labels),
		Name:           v.Name,
		SizeGb:         v.SizeGb,
		SourceImage:    v.SourceImage,
		SourceSnapshot: v.SourceSnapshot,
		Type:           v.Type,
		Zone:           path.Base(v.Zone),
	}

	resp, err := s.service.Disks.Insert(s.inst.Project, newDisk.Zone, newDisk).Do()
	if err != nil {
		return nil, err
	}

	if err = s.checkDiskStatus(newDisk.Name, newDisk.Zone, STATUS_READY); err != nil {
		return nil, s.rollbackCreate(resp.Name, err)
	}

	d, err := s.service.Disks.Get(s.inst.Project, newDisk.Zone, newDisk.Name).Do()
	if err != nil {
		return nil, err
	}

	return d, err
}

func (s *gceOps) Delete(id string) error {
	ctx := context.Background()
	found := false
	req := s.service.Disks.AggregatedList(s.inst.Project)
	if err := req.Pages(ctx, func(page *compute.DiskAggregatedList) error {
		for _, diskScopedList := range page.Items {
			for _, disk := range diskScopedList.Disks {
				if disk.Name == id {
					found = true
					_, err := s.service.Disks.Delete(s.inst.Project, path.Base(disk.Zone), id).Do()
					return err
				}
			}
		}
		return nil
	}); err != nil {
		logrus.Errorf("failed to list disks: %v", err)
		return err
	}

	if !found {
		return fmt.Errorf("failed to delete disk: %s as it wasn't found", id)
	}

	return nil
}

func (s *gceOps) Detach(devicePath string) error {
	_, err := s.service.Instances.DetachDisk(
		s.inst.Project,
		s.inst.Zone,
		s.inst.Name,
		devicePath).Do()
	if err != nil {
		return err
	}

	var d *compute.Disk
	d, err = s.service.Disks.Get(s.inst.Project, s.inst.Zone, devicePath).Do()
	if err != nil {
		return err
	}

	err = s.waitForDetach(d.SelfLink, time.Minute)
	if err != nil {
		return err
	}

	return err
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

func (s *gceOps) DevicePath(diskName string) (string, error) {
	d, err := s.service.Disks.Get(s.inst.Project, s.inst.Zone, diskName).Do()
	if err != nil {
		return "", err
	}

	if len(d.Users) == 0 {
		err = storageops.NewStorageError(storageops.ErrVolDetached,
			fmt.Sprintf("Disk: %s is detached", d.Name), s.inst.Name)
		return "", err
	}

	var inst *compute.Instance
	inst, err = s.describeinstance()
	if err != nil {
		return "", err
	}

	for _, instDisk := range inst.Disks {
		if instDisk.Source == d.SelfLink {
			return fmt.Sprintf("%s%s", googleDiskPrefix, instDisk.DeviceName), nil
		}
	}

	return "", storageops.NewStorageError(
		storageops.ErrVolAttachedOnRemoteNode,
		fmt.Sprintf("disk %s is not attached on: %s (Attached on: %v)",
			d.Name, s.inst.Name, d.Users),
		s.inst.Name)
}

func (s *gceOps) Enumerate(
	volumeIds []*string,
	labels map[string]string,
	setIdentifier string,
) (map[string][]interface{}, error) {
	sets := make(map[string][]interface{})
	found := false

	allDisks, err := s.getDisksFromAllZones(formatLabels(labels))
	if err != nil {
		return nil, err
	}

	for _, disk := range allDisks {
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

	return sets, nil
}

func (s *gceOps) FreeDevices(
	blockDeviceMappings []interface{},
	rootDeviceName string,
) ([]string, error) {
	return nil, fmt.Errorf("function not implemented")
}

func (s *gceOps) GetDeviceID(disk interface{}) (string, error) {
	if d, ok := disk.(*compute.Disk); ok {
		return d.Name, nil
	} else if d, ok := disk.(*compute.Snapshot); ok {
		return d.Name, nil
	} else {
		return "", fmt.Errorf("invalid type: %v given to GetDeviceID", disk)
	}
}

func (s *gceOps) Inspect(diskNames []*string) ([]interface{}, error) {
	allDisks, err := s.getDisksFromAllZones(nil)
	if err != nil {
		return nil, err
	}

	var disks []interface{}
	for _, id := range diskNames {
		if d, ok := allDisks[*id]; ok {
			disks = append(disks, d)
		} else {
			return nil, fmt.Errorf("disk %s not found", *id)
		}
	}

	return disks, nil
}

func (s *gceOps) RemoveTags(
	diskName string,
	labels map[string]string,
) error {
	d, err := s.service.Disks.Get(s.inst.Project, s.inst.Zone, diskName).Do()
	if err != nil {
		return err
	}

	if len(d.Labels) != 0 {
		currentLabels := d.Labels
		for k := range formatLabels(labels) {
			delete(currentLabels, k)
		}

		rb := &compute.ZoneSetLabelsRequest{
			LabelFingerprint: d.LabelFingerprint,
			Labels:           currentLabels,
		}

		_, err = s.service.Disks.SetLabels(s.inst.Project, s.inst.Zone, d.Name, rb).Do()
	}

	return err
}

func (s *gceOps) Snapshot(
	disk string,
	readonly bool,
) (interface{}, error) {
	rb := &compute.Snapshot{
		Name: fmt.Sprintf("snap-%d%02d%02d", time.Now().Year(), time.Now().Month(), time.Now().Day()),
	}

	_, err := s.service.Disks.CreateSnapshot(s.inst.Project, s.inst.Zone, disk, rb).Do()
	if err != nil {
		return nil, err
	}

	if err = s.checkSnapStatus(rb.Name, STATUS_READY); err != nil {
		return nil, err
	}

	snap, err := s.service.Snapshots.Get(s.inst.Project, rb.Name).Do()
	if err != nil {
		return nil, err
	}

	return snap, err
}

func (s *gceOps) SnapshotDelete(snapID string) error {
	_, err := s.service.Snapshots.Delete(s.inst.Project, snapID).Do()
	return err
}

func (s *gceOps) Tags(diskName string) (map[string]string, error) {
	d, err := s.service.Disks.Get(s.inst.Project, s.inst.Zone, diskName).Do()
	if err != nil {
		return nil, err
	}

	return d.Labels, nil
}

func (s *gceOps) available(v *compute.Disk) bool {
	return v.Status == STATUS_READY
}

func (s *gceOps) checkDiskStatus(id string, zone string, desired string) error {
	_, err := task.DoRetryWithTimeout(
		func() (interface{}, bool, error) {
			d, err := s.service.Disks.Get(s.inst.Project, zone, id).Do()
			if err != nil {
				return nil, true, err
			}

			actual := d.Status
			if len(actual) == 0 {
				return nil, true, fmt.Errorf("nil volume state for %v", id)
			}

			if actual != desired {
				return nil, true,
					fmt.Errorf("invalid status: %s for disk: %s. expected: %s",
						actual, id, desired)
			}

			return nil, false, nil
		},
		storageops.ProviderOpsTimeout,
		storageops.ProviderOpsRetryInterval)

	return err
}

func (s *gceOps) checkSnapStatus(id string, desired string) error {
	_, err := task.DoRetryWithTimeout(
		func() (interface{}, bool, error) {
			snap, err := s.service.Snapshots.Get(s.inst.Project, id).Do()
			if err != nil {
				return nil, true, err
			}

			actual := snap.Status
			if len(actual) == 0 {
				return nil, true, fmt.Errorf("nil snapshot state for %v", id)
			}

			if actual != desired {
				return nil, true,
					fmt.Errorf("invalid status: %s for snapshot: %s. expected: %s",
						actual, id, desired)
			}

			return nil, false, nil
		},
		storageops.ProviderOpsTimeout,
		storageops.ProviderOpsRetryInterval)

	return err
}

// Describe current instance.
func (s *gceOps) Describe() (interface{}, error) {
	return s.describeinstance()
}

func (s *gceOps) describeinstance() (*compute.Instance, error) {
	return s.service.Instances.Get(s.inst.Project, s.inst.Zone, s.inst.Name).Do()
}

// gceInfo fetches the GCE instance metadata from the metadata server
func gceInfo(inst *instance) error {
	var err error
	inst.ID, err = metadata.InstanceID()
	if err != nil {
		return err
	}

	inst.Zone, err = metadata.Zone()
	if err != nil {
		return err
	}

	inst.Name, err = metadata.InstanceName()
	if err != nil {
		return err
	}

	inst.Hostname, err = metadata.Hostname()
	if err != nil {
		return err
	}

	inst.Project, err = metadata.ProjectID()
	if err != nil {
		return err
	}

	inst.InternalIP, err = metadata.InternalIP()
	if err != nil {
		return err
	}

	inst.ExternalIP, err = metadata.ExternalIP()
	if err != nil {
		return err
	}

	return nil
}

func gceInfoFromEnv(inst *instance) error {
	var err error
	inst.Name, err = getEnvValueStrict("GCE_INSTANCE_NAME")
	if err != nil {
		return err
	}

	inst.Zone, err = getEnvValueStrict("GCE_INSTANCE_ZONE")
	if err != nil {
		return err
	}

	inst.Project, err = getEnvValueStrict("GCE_INSTANCE_PROJECT")
	if err != nil {
		return err
	}

	return nil
}

func getEnvValueStrict(key string) (string, error) {
	if val := os.Getenv(key); len(val) != 0 {
		return val, nil
	}

	return "", fmt.Errorf("env variable %s is not set", key)
}

func (s *gceOps) rollbackCreate(id string, createErr error) error {
	logrus.Warnf("Rollback create volume %v, Error %v", id, createErr)
	err := s.Delete(id)
	if err != nil {
		logrus.Warnf("Rollback failed volume %v, Error %v", id, err)
	}
	return createErr
}

// waitForAttach checks if given disk is detached from the local instance
func (s *gceOps) waitForDetach(
	diskURL string,
	timeout time.Duration,
) error {

	_, err := task.DoRetryWithTimeout(
		func() (interface{}, bool, error) {
			inst, err := s.describeinstance()
			if err != nil {
				return nil, true, err
			}

			for _, d := range inst.Disks {
				if d.Source == diskURL {
					return nil, true,
						fmt.Errorf("disk: %s is still attached to instance: %s",
							diskURL, s.inst.Name)
				}
			}

			return nil, false, nil

		},
		storageops.ProviderOpsTimeout,
		storageops.ProviderOpsRetryInterval)

	return err
}

// waitForAttach checks if given disk is attached to the local instance
func (s *gceOps) waitForAttach(
	disk *compute.Disk,
	timeout time.Duration,
) (string, error) {
	devicePath, err := task.DoRetryWithTimeout(
		func() (interface{}, bool, error) {
			devicePath, err := s.DevicePath(disk.Name)
			if err != nil {
				return "", true, err
			}

			return devicePath, false, nil
		},
		storageops.ProviderOpsTimeout,
		storageops.ProviderOpsRetryInterval)
	if err != nil {
		return "", err
	}

	return devicePath.(string), nil
}

// generateListFilterFromLabels create a filter string based off --filter documentation at
// https://cloud.google.com/sdk/gcloud/reference/compute/disks/list
func generateListFilterFromLabels(labels map[string]string) string {
	var filter string
	for k, v := range labels {
		filter = fmt.Sprintf("%s(labels.%s eq %s)", filter, k, v)
	}

	return filter
}

func (s *gceOps) getDisksFromAllZones(labels map[string]string) (map[string]*compute.Disk, error) {
	ctx := context.Background()
	response := make(map[string]*compute.Disk)
	var req *compute.DisksAggregatedListCall

	if len(labels) > 0 {
		filter := generateListFilterFromLabels(labels)
		req = s.service.Disks.AggregatedList(s.inst.Project).Filter(filter)
	} else {
		req = s.service.Disks.AggregatedList(s.inst.Project)
	}

	if err := req.Pages(ctx, func(page *compute.DiskAggregatedList) error {
		for _, diskScopedList := range page.Items {
			for _, disk := range diskScopedList.Disks {
				response[disk.Name] = disk
			}
		}

		return nil
	}); err != nil {
		logrus.Errorf("failed to list disks: %v", err)
		return nil, err
	}

	return response, nil
}

func formatLabels(labels map[string]string) map[string]string {
	newLabels := make(map[string]string)
	for k, v := range labels {
		newLabels[strings.ToLower(k)] = strings.ToLower(v)
	}
	return newLabels
}
