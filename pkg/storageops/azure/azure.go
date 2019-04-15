package azure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-06-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/portworx/sched-ops/task"
	"github.com/sirupsen/logrus"
)

const (
	envInstanceID        = "AZURE_INSTANCE_ID"
	envScaleSetName      = "AZURE_SCALE_SET_NAME"
	envSubscriptionID    = "AZURE_SUBSCRIPTION_ID"
	envResourceGroupName = "AZURE_RESOURCE_GROUP_NAME"
)

const (
	name                    = "azure"
	userAgentExtension      = "osd"
	azureDiskPrefix         = "/dev/disk/azure/scsi1/lun"
	snapNameFormat          = "2006-01-02_15.04.05.999999"
	clientPollingDelay      = 5 * time.Second
	clientRetryAttempts     = 10
	devicePathMaxRetryCount = 3
	devicePathRetryInterval = 2 * time.Second
)

type azureOps struct {
	instance          string
	resourceGroupName string
	disksClient       *compute.DisksClient
	vmsClient         vmsClient
	snapshotsClient   *compute.SnapshotsClient
}

func (a *azureOps) Name() string {
	return name
}

func (a *azureOps) InstanceID() string {
	return a.instance
}

func NewEnvClient() (storageops.Ops, error) {
	instance, err := storageops.GetEnvValueStrict(envInstanceID)
	if err != nil {
		return nil, err
	}
	subscriptionID, err := storageops.GetEnvValueStrict(envSubscriptionID)
	if err != nil {
		return nil, err
	}
	resourceGroupName, err := storageops.GetEnvValueStrict(envResourceGroupName)
	if err != nil {
		return nil, err
	}
	scaleSetName := os.Getenv(envScaleSetName)
	return NewClient(instance, scaleSetName, subscriptionID, resourceGroupName)
}

func NewClient(
	instance, scaleSetName, subscriptionID, resourceGroupName string,
) (storageops.Ops, error) {
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil, err
	}

	disksClient := compute.NewDisksClient(subscriptionID)
	disksClient.Authorizer = authorizer
	disksClient.PollingDelay = clientPollingDelay
	disksClient.RetryAttempts = clientRetryAttempts
	disksClient.AddToUserAgent(userAgentExtension)

	vmsClient := NewVMsClient(scaleSetName, subscriptionID, resourceGroupName, authorizer)

	snapshotsClient := compute.NewSnapshotsClient(subscriptionID)
	snapshotsClient.Authorizer = authorizer
	snapshotsClient.PollingDelay = clientPollingDelay
	snapshotsClient.RetryAttempts = clientRetryAttempts
	snapshotsClient.AddToUserAgent(userAgentExtension)

	return &azureOps{
		instance:          instance,
		resourceGroupName: resourceGroupName,
		disksClient:       &disksClient,
		vmsClient:         vmsClient,
		snapshotsClient:   &snapshotsClient,
	}, nil
}

func (a *azureOps) Create(
	template interface{},
	labels map[string]string,
) (interface{}, error) {
	d, ok := template.(*compute.Disk)
	if !ok {
		return nil, storageops.NewStorageError(
			storageops.ErrVolInval,
			"Invalid volume template given",
			a.instance,
		)
	}

	// Check if the disk already exists; return err if it does
	_, err := a.disksClient.Get(
		context.Background(),
		a.resourceGroupName,
		*d.Name,
	)
	if err == nil {
		return "", fmt.Errorf("disk with id %v already exists", *d.Name)
	} else {
		derr, ok := err.(autorest.DetailedError)
		if !ok {
			return "", err
		}
		code, ok := derr.StatusCode.(int)
		if !ok || code != 404 {
			return "", err
		}
	}

	ctx := context.Background()
	future, err := a.disksClient.CreateOrUpdate(
		ctx,
		a.resourceGroupName,
		*d.Name,
		compute.Disk{
			Location: d.Location,
			Type:     d.Type,
			Zones:    d.Zones,
			Tags:     formatTags(labels),
			Sku:      d.Sku,
			DiskProperties: &compute.DiskProperties{
				CreationData: &compute.CreationData{
					CreateOption: compute.Empty,
				},
				DiskSizeGB:        d.DiskProperties.DiskSizeGB,
				DiskIOPSReadWrite: d.DiskProperties.DiskIOPSReadWrite,
				DiskMBpsReadWrite: d.DiskProperties.DiskMBpsReadWrite,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create disk: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, a.disksClient.Client)
	if err != nil {
		return nil, fmt.Errorf("cannot get the disk create or update future response: %v", err)
	}

	dd, err := future.Result(*a.disksClient)
	return &dd, err
}

func (a *azureOps) GetDeviceID(disk interface{}) (string, error) {
	if d, ok := disk.(*compute.Disk); ok {
		return *d.Name, nil
	} else if s, ok := disk.(*compute.Snapshot); ok {
		return *s.Name, nil
	}
	return "", storageops.NewStorageError(
		storageops.ErrVolInval,
		"Invalid volume given",
		a.instance,
	)
}

func (a *azureOps) Attach(diskName string) (string, error) {
	dataDisks, err := a.vmsClient.getDataDisks(a.instance)
	if err != nil {
		return "", err
	}

	disk, err := a.disksClient.Get(
		context.Background(),
		a.resourceGroupName,
		diskName,
	)
	if err != nil {
		return "", fmt.Errorf("cannot get disk %v: %v", diskName, err)
	}

	nextLun := nextAvailableLun(dataDisks)
	if nextLun < 0 {
		return "", fmt.Errorf("No LUN available to attach the disk. "+
			"%v disks attached to the VM instance", len(dataDisks))
	}

	newDataDisks := append(
		dataDisks,
		compute.DataDisk{
			Lun:          &nextLun,
			Name:         to.StringPtr(diskName),
			DiskSizeGB:   disk.DiskSizeGB,
			CreateOption: compute.DiskCreateOptionTypesAttach,
			ManagedDisk: &compute.ManagedDiskParameters{
				ID: disk.ID,
			},
		},
	)
	if err := a.vmsClient.updateDataDisks(a.instance, newDataDisks); err != nil {
		return "", err
	}

	return a.waitForAttach(diskName)
}

func (a *azureOps) Detach(diskName string) error {
	return a.detachInternal(diskName, a.instance)
}

func (a *azureOps) DetachFrom(diskName, instance string) error {
	return a.detachInternal(diskName, instance)
}

func (a *azureOps) detachInternal(diskName, instance string) error {
	dataDisks, err := a.vmsClient.getDataDisks(instance)
	if err != nil {
		return err
	}

	disk, err := a.disksClient.Get(
		context.Background(),
		a.resourceGroupName,
		diskName,
	)
	if err != nil {
		return fmt.Errorf("cannot get disk %v: %v", diskName, err)
	}

	diskToDetach := strings.ToLower(*disk.ID)

	newDataDisks := make([]compute.DataDisk, 0)
	for _, d := range dataDisks {
		if strings.ToLower(*d.ManagedDisk.ID) == diskToDetach {
			continue
		}
		newDataDisks = append(newDataDisks, d)
	}

	if err := a.vmsClient.updateDataDisks(instance, newDataDisks); err != nil {
		return err
	}

	return a.waitForDetach(diskName, instance)
}

func (a *azureOps) Delete(diskName string) error {
	ctx := context.Background()
	future, err := a.disksClient.Delete(ctx, a.resourceGroupName, diskName)
	if err != nil {
		return fmt.Errorf("cannot delete disk %s: %v", diskName, err)
	}

	err = future.WaitForCompletionRef(ctx, a.disksClient.Client)
	if err != nil {
		return fmt.Errorf("cannot delete the disk %s or update future response: %v", diskName, err)
	}

	_, err = future.Result(*a.disksClient)
	return err
}

func (a *azureOps) DeleteFrom(diskName, _ string) error {
	return a.Delete(diskName)
}

func (a *azureOps) Describe() (interface{}, error) {
	return a.vmsClient.describe(a.instance)
}

func (a *azureOps) FreeDevices(
	blockDeviceMappings []interface{},
	rootDeviceName string,
) ([]string, error) {
	return nil, storageops.ErrNotSupported
}

func (a *azureOps) Inspect(diskNames []*string) ([]interface{}, error) {
	allDisks, err := a.getDisks(nil)
	if err != nil {
		return nil, err
	}

	var disks []interface{}
	for _, id := range diskNames {
		if d, ok := allDisks[*id]; ok {
			disks = append(disks, d)
		} else {
			return nil, storageops.NewStorageError(
				storageops.ErrVolNotFound,
				fmt.Sprintf("disk %s not found", *id),
				a.instance,
			)
		}
	}

	return disks, nil
}

func (a *azureOps) DeviceMappings() (map[string]string, error) {
	dataDisks, err := a.vmsClient.getDataDisks(a.instance)
	if err != nil {
		return nil, err
	}

	devMap := make(map[string]string)
	for _, d := range dataDisks {
		devPath, err := lunToBlockDevPath(*d.Lun)
		if err != nil {
			return nil, storageops.NewStorageError(
				storageops.ErrInvalidDevicePath,
				fmt.Sprintf("unable to find block dev path for lun%v: %v", *d.Lun, err),
				a.instance,
			)
		}
		devMap[devPath] = *d.Name
	}

	return devMap, nil
}

func (a *azureOps) Enumerate(
	diskNames []*string,
	labels map[string]string,
	setIdentifier string,
) (map[string][]interface{}, error) {
	allDisks, err := a.getDisks(labels)
	if err != nil {
		return nil, err
	}

	sets := make(map[string][]interface{})
	for _, disk := range allDisks {
		if len(setIdentifier) == 0 {
			storageops.AddElementToMap(sets, disk, storageops.SetIdentifierNone)
		} else {
			found := false
			for key, value := range disk.Tags {
				if key == setIdentifier && value != nil {
					storageops.AddElementToMap(sets, disk, *value)
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

func (a *azureOps) DevicePath(diskName string) (string, error) {
	disk, err := a.disksClient.Get(
		context.Background(),
		a.resourceGroupName,
		diskName,
	)
	if derr, ok := err.(autorest.DetailedError); ok {
		code, ok := derr.StatusCode.(int)
		if ok && code == 404 {
			return "", storageops.NewStorageError(
				storageops.ErrVolNotFound,
				fmt.Sprintf("disk: %s not found", diskName),
				a.instance,
			)
		}
		return "", err
	} else if err != nil {
		return "", err
	}

	if disk.ManagedBy == nil || len(*disk.ManagedBy) == 0 {
		return "", storageops.NewStorageError(
			storageops.ErrVolDetached,
			fmt.Sprintf("Disk: %s is detached", diskName),
			a.instance,
		)
	}

	dataDisks, err := a.vmsClient.getDataDisks(a.instance)
	if err != nil {
		return "", err
	}

	for _, d := range dataDisks {
		if *d.Name == diskName {
			// Retry to get the block dev path as it may take few seconds for the path
			// to be created even after the disk shows attached.
			devPath, err := lunToBlockDevPathWithRetry(*d.Lun)
			if err == nil {
				return devPath, nil
			}
			return "", storageops.NewStorageError(
				storageops.ErrInvalidDevicePath,
				fmt.Sprintf("unable to find block dev path for lun%v: %v", *d.Lun, err),
				a.instance,
			)
		}
	}

	return "", storageops.NewStorageError(
		storageops.ErrVolAttachedOnRemoteNode,
		fmt.Sprintf("disk %s is not attached on: %s (Attached on: %v)",
			diskName, a.instance, disk.ManagedBy),
		a.instance,
	)
}

func (a *azureOps) Snapshot(diskName string, readonly bool) (interface{}, error) {
	if !readonly {
		return nil, fmt.Errorf("read-write snapshots are not supported in Azure")
	}

	disk, err := a.disksClient.Get(context.Background(), a.resourceGroupName, diskName)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	future, err := a.snapshotsClient.CreateOrUpdate(
		ctx,
		a.resourceGroupName,
		fmt.Sprint("snap-", time.Now().Format(snapNameFormat)),
		compute.Snapshot{
			Location: disk.Location,
			SnapshotProperties: &compute.SnapshotProperties{
				CreationData: &compute.CreationData{
					CreateOption:     compute.Copy,
					SourceResourceID: disk.ID,
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create snapshot: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, a.snapshotsClient.Client)
	if err != nil {
		return nil, fmt.Errorf("cannot get the snapshot create or update future response: %v", err)
	}

	snap, err := future.Result(*a.snapshotsClient)
	return &snap, err
}

func (a *azureOps) SnapshotDelete(snapName string) error {
	ctx := context.Background()
	future, err := a.snapshotsClient.Delete(ctx, a.resourceGroupName, snapName)
	if err != nil {
		return fmt.Errorf("cannot delete snapshot %s: %v", snapName, err)
	}

	err = future.WaitForCompletionRef(ctx, a.snapshotsClient.Client)
	if err != nil {
		return fmt.Errorf("cannot delete the snapshot %s or update future response: %v", snapName, err)
	}

	_, err = future.Result(*a.snapshotsClient)
	return err
}

func (a *azureOps) ApplyTags(diskName string, labels map[string]string) error {
	if len(labels) == 0 {
		return nil
	}

	disk, err := a.disksClient.Get(
		context.Background(),
		a.resourceGroupName,
		diskName,
	)
	if err != nil {
		return err
	}

	if len(disk.Tags) == 0 {
		disk.Tags = make(map[string]*string)
	}

	for k, v := range labels {
		disk.Tags[k] = to.StringPtr(v)
	}

	ctx := context.Background()
	future, err := a.disksClient.Update(
		ctx,
		a.resourceGroupName,
		diskName,
		compute.DiskUpdate{
			Tags: disk.Tags,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot update disk: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, a.disksClient.Client)
	if err != nil {
		return fmt.Errorf("cannot get the disk create or update future response: %v", err)
	}

	_, err = future.Result(*a.disksClient)
	return err
}

func (a *azureOps) RemoveTags(diskName string, labels map[string]string) error {
	if len(labels) == 0 {
		return nil
	}

	disk, err := a.disksClient.Get(
		context.Background(),
		a.resourceGroupName,
		diskName,
	)
	if err != nil {
		return err
	}

	if len(disk.Tags) == 0 {
		return nil
	}

	for k := range labels {
		delete(disk.Tags, k)
	}

	ctx := context.Background()
	future, err := a.disksClient.Update(
		ctx,
		a.resourceGroupName,
		diskName,
		compute.DiskUpdate{
			Tags: disk.Tags,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot update disk: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, a.disksClient.Client)
	if err != nil {
		return fmt.Errorf("cannot get the disk create or update future response: %v", err)
	}

	_, err = future.Result(*a.disksClient)
	return err
}

func (a *azureOps) Tags(diskName string) (map[string]string, error) {
	disk, err := a.disksClient.Get(context.Background(), a.resourceGroupName, diskName)
	if err != nil {
		return nil, err
	}

	tags := make(map[string]string)
	for k, v := range disk.Tags {
		if v == nil {
			tags[k] = ""
		} else {
			tags[k] = *v
		}
	}
	return tags, nil
}

func (a *azureOps) getDisks(labels map[string]string) (map[string]*compute.Disk, error) {
	response := make(map[string]*compute.Disk)

	for it, err := a.disksClient.ListComplete(context.Background()); it.NotDone(); err = it.Next() {
		if err != nil {
			return nil, err
		}

		disk := it.Value()
		if labelsMatch(&disk, labels) {
			response[*it.Value().Name] = &disk
		}
	}

	return response, nil
}

func (a *azureOps) waitForAttach(diskName string) (string, error) {
	devicePath, err := task.DoRetryWithTimeout(
		func() (interface{}, bool, error) {
			devicePath, err := a.DevicePath(diskName)
			if se, ok := err.(*storageops.StorageError); ok &&
				se.Code == storageops.ErrVolAttachedOnRemoteNode {
				return "", false, err
			} else if err != nil {
				return "", true, err
			}

			return devicePath, false, nil
		},
		storageops.ProviderOpsTimeout,
		storageops.ProviderOpsRetryInterval,
	)
	if err != nil {
		return "", err
	}

	return devicePath.(string), nil
}

func (a *azureOps) waitForDetach(diskName, instance string) error {
	_, err := task.DoRetryWithTimeout(
		func() (interface{}, bool, error) {
			dataDisks, err := a.vmsClient.getDataDisks(instance)
			if err != nil {
				return nil, true, err
			}

			for _, d := range dataDisks {
				if *d.Name == diskName {
					return nil, true,
						fmt.Errorf("disk %s is still attached to instance %s",
							diskName, instance)
				}
			}

			return nil, false, nil
		},
		storageops.ProviderOpsTimeout,
		storageops.ProviderOpsRetryInterval,
	)

	return err
}

func labelsMatch(disk *compute.Disk, labels map[string]string) bool {
	for key, expected := range labels {
		if actual, exists := disk.Tags[key]; exists {
			// Nil values are not allowed in tags, just safety check
			if actual == nil && expected != "" {
				return false
			} else if actual != nil && *actual != expected {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func formatTags(labels map[string]string) map[string]*string {
	tags := make(map[string]*string)
	for k, v := range labels {
		value := v
		tags[k] = &value
	}
	return tags
}

func nextAvailableLun(dataDisks []compute.DataDisk) int32 {
	usedLuns := make(map[int32]struct{})
	for _, d := range dataDisks {
		if d.Lun != nil {
			usedLuns[*d.Lun] = struct{}{}
		}
	}
	nextAvailableLun := int32(-1)
	for i := int32(0); i < 64; i++ {
		if _, ok := usedLuns[i]; !ok {
			nextAvailableLun = i
			break
		}
	}
	return nextAvailableLun
}

func lunToBlockDevPathWithRetry(lun int32) (string, error) {
	var (
		retryCount int
		path       string
		err        error
	)

	for {
		if path, err = lunToBlockDevPath(lun); err == nil {
			return path, nil
		}
		logrus.Warnf(err.Error())
		retryCount++
		if retryCount >= devicePathMaxRetryCount {
			break
		}
		time.Sleep(devicePathRetryInterval)
	}
	return "", err
}

func lunToBlockDevPath(lun int32) (string, error) {
	devPath := azureDiskPrefix + strconv.Itoa(int(lun))
	// check if path is a sym link. If yes, return pointee
	fi, err := os.Lstat(devPath)
	if err != nil {
		return "", err
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		output, err := filepath.EvalSymlinks(devPath)
		if err != nil {
			return "", fmt.Errorf("failed to read symlink %s due to: %v", devPath, err)
		}

		devPath = strings.TrimSpace(string(output))
	} else {
		return "", fmt.Errorf("%s was expected to be a symlink to actual "+
			"device path", devPath)
	}

	return devPath, nil
}
