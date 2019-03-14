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
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/portworx/sched-ops/task"
	"github.com/sirupsen/logrus"
)

const (
	envInstanceName      = "AZURE_INSTANCE_NAME"
	envSubscriptionID    = "AZURE_SUBSCRIPTION_ID"
	envResourceGroupName = "AZURE_RESOURCE_GROUP_NAME"
	envAzureEnvironment  = "AZURE_ENVIRONMENT"
	envTenantID          = "AZURE_TENANT_ID"
	envClientID          = "AZURE_CLIENT_ID"
	envClientSecret      = "AZURE_CLIENT_SECRET"
)

const (
	name                    = "azure"
	defaultEnvironment      = "AzurePublicCloud"
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
	subscriptionID    string
	resourceGroupName string
	disksClient       *compute.DisksClient
	vmsClient         *compute.VirtualMachinesClient
	snapshotsClient   *compute.SnapshotsClient
}

func (a *azureOps) Name() string {
	return name
}

func (a *azureOps) InstanceID() string {
	return a.instance
}

func NewEnvClient() (storageops.Ops, error) {
	instance, err := storageops.GetEnvValueStrict(envInstanceName)
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

	return NewClient(instance, subscriptionID, resourceGroupName)
}

func NewClient(
	instance, subscriptionID, resourceGroupName string,
) (storageops.Ops, error) {
	envName := os.Getenv(envAzureEnvironment)
	if len(envName) == 0 {
		envName = defaultEnvironment
	}
	env, err := azure.EnvironmentFromName(envName)
	if err != nil {
		return nil, fmt.Errorf("invalid cloud name '%s' specified: %v", envName, err)
	}

	tenantID, err := storageops.GetEnvValueStrict(envTenantID)
	if err != nil {
		return nil, err
	}

	clientID, err := storageops.GetEnvValueStrict(envClientID)
	if err != nil {
		return nil, err
	}

	clientSecret, err := storageops.GetEnvValueStrict(envClientSecret)
	if err != nil {
		return nil, err
	}

	oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, tenantID)
	if err != nil {
		return nil, err
	}

	token, err := adal.NewServicePrincipalToken(*oauthConfig, clientID,
		clientSecret, env.ResourceManagerEndpoint)
	if err != nil {
		return nil, err
	}

	authorizer := autorest.NewBearerAuthorizer(token)

	disksClient := compute.NewDisksClient(subscriptionID)
	disksClient.Authorizer = authorizer
	disksClient.PollingDelay = clientPollingDelay
	disksClient.RetryAttempts = clientRetryAttempts
	disksClient.AddToUserAgent(userAgentExtension)

	vmsClient := compute.NewVirtualMachinesClient(subscriptionID)
	vmsClient.Authorizer = authorizer
	vmsClient.PollingDelay = clientPollingDelay
	vmsClient.RetryAttempts = clientRetryAttempts
	vmsClient.AddToUserAgent(userAgentExtension)

	snapshotsClient := compute.NewSnapshotsClient(subscriptionID)
	snapshotsClient.Authorizer = authorizer
	snapshotsClient.PollingDelay = clientPollingDelay
	snapshotsClient.RetryAttempts = clientRetryAttempts
	snapshotsClient.AddToUserAgent(userAgentExtension)

	return &azureOps{
		instance:          instance,
		resourceGroupName: resourceGroupName,
		disksClient:       &disksClient,
		vmsClient:         &vmsClient,
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
	vm, err := a.describeInstance()
	if err != nil {
		return "", fmt.Errorf("cannot get vm %v: %v", a.instance, err)
	}

	disk, err := a.disksClient.Get(
		context.Background(),
		a.resourceGroupName,
		diskName,
	)
	if err != nil {
		return "", fmt.Errorf("cannot get disk %v: %v", diskName, err)
	}

	nextLun := nextAvailableLun(*vm.StorageProfile.DataDisks)
	if nextLun < 0 {
		return "", fmt.Errorf("No LUN available to attach the disk. "+
			"%v disks attached to the VM instance", len(*vm.StorageProfile.DataDisks))
	}

	*vm.StorageProfile.DataDisks = append(
		*vm.StorageProfile.DataDisks,
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
	vm.Resources = nil

	ctx := context.Background()
	future, err := a.vmsClient.CreateOrUpdate(
		ctx,
		a.resourceGroupName,
		a.instance,
		vm,
	)
	if err != nil {
		return "", fmt.Errorf("cannot update vm %v: %v", a.instance, err)
	}

	err = future.WaitForCompletionRef(ctx, a.vmsClient.Client)
	if err != nil {
		return "", fmt.Errorf("cannot get the vm create or update future response: %v", err)
	}

	return a.waitForAttach(diskName)
}

func (a *azureOps) Detach(diskName string) error {
	return a.detachInternal(diskName, a.instance)
}

func (a *azureOps) DetachFrom(diskName, instanceName string) error {
	return a.detachInternal(diskName, instanceName)
}

func (a *azureOps) detachInternal(diskName, instanceName string) error {
	vm, err := a.vmsClient.Get(
		context.Background(),
		a.resourceGroupName,
		instanceName,
		compute.InstanceView,
	)
	if err != nil {
		return fmt.Errorf("cannot get vm %v: %v", instanceName, err)
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

	disks := make([]compute.DataDisk, 0)
	for _, d := range *vm.StorageProfile.DataDisks {
		if strings.ToLower(*d.ManagedDisk.ID) == diskToDetach {
			continue
		}
		disks = append(disks, d)
	}

	if len(disks) == 0 {
		vm.StorageProfile.DataDisks = &[]compute.DataDisk{}
	} else {
		vm.StorageProfile.DataDisks = &disks
	}
	vm.Resources = nil

	ctx := context.Background()
	future, err := a.vmsClient.CreateOrUpdate(
		ctx,
		a.resourceGroupName,
		instanceName,
		vm,
	)
	if err != nil {
		return fmt.Errorf("cannot update vm %v: %v", instanceName, err)
	}

	err = future.WaitForCompletionRef(ctx, a.vmsClient.Client)
	if err != nil {
		return fmt.Errorf("cannot get the vm create or update future response: %v", err)
	}

	return a.waitForDetach(diskName)
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
	return a.describeInstance()
}

func (a *azureOps) describeInstance() (compute.VirtualMachine, error) {
	return a.vmsClient.Get(
		context.Background(),
		a.resourceGroupName,
		a.instance,
		compute.InstanceView,
	)
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
	vm, err := a.describeInstance()
	if err != nil {
		return nil, err
	}

	devMap := make(map[string]string)
	for _, d := range *vm.StorageProfile.DataDisks {
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

	vm, err := a.describeInstance()
	if err != nil {
		return "", err
	}

	if vm.StorageProfile == nil || vm.StorageProfile.DataDisks == nil {
		return "", fmt.Errorf("instance does not have any disks attached")
	}

	for _, d := range *vm.StorageProfile.DataDisks {
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

func (a *azureOps) waitForDetach(diskName string) error {
	_, err := task.DoRetryWithTimeout(
		func() (interface{}, bool, error) {
			vm, err := a.describeInstance()
			if err != nil {
				return nil, true, err
			}

			if vm.StorageProfile == nil || vm.StorageProfile.DataDisks == nil {
				return nil, true, fmt.Errorf("vm in invalid state")
			}

			for _, d := range *vm.StorageProfile.DataDisks {
				if *d.Name == diskName {
					return nil, true,
						fmt.Errorf("disk %s is still attached to instance %s",
							diskName, a.instance)
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
