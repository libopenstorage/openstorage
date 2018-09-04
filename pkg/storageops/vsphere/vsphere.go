package vsphere

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/sirupsen/logrus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/vsphere/vclib"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/vsphere/vclib/diskmanagers"
)

const (
	diskDirectory  = "osd-provisioned-disks"
	dummyDiskName  = "kube-dummyDisk.vmdk"
	diskByIDPath   = "/dev/disk/by-id/"
	diskSCSIPrefix = "wwn-0x"
)

type vsphereOps struct {
	vm   *vclib.VirtualMachine
	conn *vclib.VSphereConnection
	cfg  *VSphereConfig
}

// VirtualDisk encapsulates the existing virtual disk object to add a managed object
// reference to the datastore of the disk
type VirtualDisk struct {
	diskmanagers.VirtualDisk
	// DatastoreRef is the managed object reference of the datastore on which the disk belongs
	DatastoreRef types.ManagedObjectReference
}

// NewClient creates a new vsphere storageops instance
func NewClient(cfg *VSphereConfig) (storageops.Ops, error) {
	vSphereConn := &vclib.VSphereConnection{
		Username:          cfg.User,
		Password:          cfg.Password,
		Hostname:          cfg.VCenterIP,
		Insecure:          cfg.InsecureFlag,
		RoundTripperCount: cfg.RoundTripperCount,
		Port:              cfg.VCenterPort,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	vmObj, err := GetVMObject(ctx, vSphereConn, cfg.VMUUID)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Using following configuration for vsphere:")
	logrus.Debugf("  vCenter: %s:%s", cfg.VCenterIP, cfg.VCenterPort)
	logrus.Debugf("  Datacenter: %s", vmObj.Datacenter.Name())
	logrus.Debugf("  VMUUID: %s", cfg.VMUUID)

	return &vsphereOps{
		cfg:  cfg,
		vm:   vmObj,
		conn: vSphereConn,
	}, nil
}

func (ops *vsphereOps) Name() string { return "vsphere" }

func (ops *vsphereOps) InstanceID() string { return ops.cfg.VMUUID }

func (ops *vsphereOps) Create(opts interface{}, labels map[string]string) (interface{}, error) {
	volumeOptions, ok := opts.(*vclib.VolumeOptions)
	if !ok {
		return nil, fmt.Errorf("invalid volume options specified to create: %v", opts)
	}

	if len(volumeOptions.Tags) == 0 {
		volumeOptions.Tags = labels
	} else {
		for k, v := range labels {
			volumeOptions.Tags[k] = v
		}
	}

	if len(volumeOptions.Datastore) == 0 {
		return nil, fmt.Errorf("datastore is required for the create call")
	}

	datastore := strings.TrimSpace(volumeOptions.Datastore)
	logrus.Infof("Given datastore/datastore cluster: %s for new disk", datastore)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vmObj, err := ops.renewVM(ctx, ops.vm)
	if err != nil {
		return nil, err
	}

	isPod, storagePod, err := IsStoragePod(ctx, vmObj, volumeOptions.Datastore)
	if err != nil {
		return nil, err
	}

	if isPod {
		datastore, err = ops.getDatastoreToUseInStoragePod(ctx, vmObj, volumeOptions, storagePod)
		if err != nil {
			return nil, err
		}
	}

	logrus.Infof("Using datastore: %s for new disk", datastore)

	ds, err := vmObj.Datacenter.GetDatastoreByName(ctx, datastore)
	if err != nil {
		logrus.Errorf("Failed to get datastore: %s due to: %v", datastore, err)
		return nil, err
	}

	volumeOptions.Datastore = datastore

	diskBasePath := filepath.Clean(ds.Path(diskDirectory)) + "/"
	err = ds.CreateDirectory(ctx, diskBasePath, false)
	if err != nil && err != vclib.ErrFileAlreadyExist {
		logrus.Errorf("Cannot create dir %#v. err %s", diskBasePath, err)
		return nil, err
	}

	diskPath := diskBasePath + volumeOptions.Name + ".vmdk"
	disk := diskmanagers.VirtualDisk{
		DiskPath:      diskPath,
		VolumeOptions: volumeOptions,
	}

	diskPath, err = disk.Create(ctx, ds)
	if err != nil {
		logrus.Errorf("Failed to create a vsphere volume with volumeOptions: %+v on "+
			"datastore: %s. err: %+v", volumeOptions, datastore, err)
		return nil, err
	}

	// Get the canonical path for the volume path.
	canonicalVolumePath, err := getCanonicalVolumePath(ctx, vmObj.Datacenter, diskPath)
	if err != nil {
		logrus.Errorf("Failed to get canonical vsphere disk path for: %s with "+
			"volumeOptions: %+v on datastore: %s. err: %+v", diskPath, volumeOptions, datastore, err)
		return nil, err
	}

	disk.DiskPath = canonicalVolumePath

	return &VirtualDisk{
		VirtualDisk:  disk,
		DatastoreRef: ds.Reference(),
	}, nil
}

func (ops *vsphereOps) GetDeviceID(vDisk interface{}) (string, error) {
	disk, ok := vDisk.(*VirtualDisk)
	if !ok {
		return "", fmt.Errorf("invalid input: %v to GetDeviceID", vDisk)
	}

	return disk.DiskPath, nil
}

// Attach takes in the path of the vmdk file and returns where it is attached inside the vm instance
func (ops *vsphereOps) Attach(diskPath string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vmObj, err := ops.renewVM(ctx, ops.vm)
	if err != nil {
		return "", err
	}

	diskUUID, err := vmObj.AttachDisk(ctx, diskPath, &vclib.VolumeOptions{SCSIControllerType: vclib.PVSCSIControllerType})
	if err != nil {
		logrus.Errorf("Failed to attach vsphere disk: %s for VM: %s. err: +%v", diskPath, vmObj.Name(), err)
		return "", err
	}

	return path.Join(diskByIDPath, diskSCSIPrefix+diskUUID), nil
}

func (ops *vsphereOps) Detach(diskPath string) error {
	return ops.detachInternal(diskPath, ops.cfg.VMUUID)
}

func (ops *vsphereOps) DetachFrom(diskPath, instanceID string) error {
	return ops.detachInternal(diskPath, instanceID)
}

func (ops *vsphereOps) detachInternal(diskPath, instanceID string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var vmObj *vclib.VirtualMachine
	var err error
	if instanceID == ops.cfg.VMUUID {
		vmObj, err = ops.renewVM(ctx, ops.vm)
		if err != nil {
			return err
		}
	} else {
		vmObj, err = GetVMObject(ctx, ops.conn, instanceID)
		if err != nil {
			return err
		}
	}

	if err := vmObj.DetachDisk(ctx, diskPath); err != nil {
		logrus.Errorf("Failed to detach vsphere disk: %s for VM: %s. err: +%v", diskPath, vmObj.Name(), err)
		return err
	}

	return nil
}

// Delete virtual disk at given path
func (ops *vsphereOps) Delete(diskPath string) error {
	return ops.deleteInternal(diskPath, ops.cfg.VMUUID)
}

func (ops *vsphereOps) DeleteFrom(diskPath, instanceID string) error {
	return ops.deleteInternal(diskPath, instanceID)
}

func (ops *vsphereOps) deleteInternal(diskPath, instanceID string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var vmObj *vclib.VirtualMachine
	var err error
	if instanceID == ops.cfg.VMUUID {
		vmObj, err = ops.renewVM(ctx, ops.vm)
		if err != nil {
			return err
		}
	} else {
		vmObj, err = GetVMObject(ctx, ops.conn, instanceID)
		if err != nil {
			return err
		}
	}

	disk := diskmanagers.VirtualDisk{
		DiskPath:      diskPath,
		VolumeOptions: &vclib.VolumeOptions{},
		VMOptions:     &vclib.VMOptions{},
	}

	err = disk.Delete(ctx, vmObj.Datacenter)
	if err != nil {
		logrus.Errorf("Failed to delete vsphere disk: %s. err: %+v", diskPath, err)
	}

	return err
}

// Desribe an instance of the virtual machine object to which ops is connected to
func (ops *vsphereOps) Describe() (interface{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return ops.renewVM(ctx, ops.vm)
}

// FreeDevices is not supported by this provider
func (ops *vsphereOps) FreeDevices(blockDeviceMappings []interface{}, rootDeviceName string) ([]string, error) {
	return nil, storageops.ErrNotSupported
}

func (ops *vsphereOps) Inspect(diskPaths []*string) ([]interface{}, error) {
	// TODO find a way to map diskPaths to unattached/attached virtual disks and query info
	// currently returning the disks directly

	return nil, storageops.ErrNotSupported
}

// DeviceMappings returns map[local_attached_volume_path]->volume ID/NAME
func (ops *vsphereOps) DeviceMappings() (map[string]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vmObj, err := ops.renewVM(ctx, ops.vm)
	if err != nil {
		return nil, err
	}

	vmDevices, err := vmObj.Device(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices for vm: %s", vmObj.Name())
	}

	// Go over all the devices attached on this vm and create a map of just the virtual disks and where
	// they are attached on the vm
	m := make(map[string]string)
	for _, device := range vmDevices {
		if vmDevices.TypeName(device) == "VirtualDisk" {
			virtualDevice := device.GetVirtualDevice()
			backing, ok := virtualDevice.Backing.(*types.VirtualDiskFlatVer2BackingInfo)
			if ok {
				devicePath, err := ops.DevicePath(backing.FileName)
				if err == nil && len(devicePath) != 0 { // TODO can ignore errors?
					m[devicePath] = backing.FileName
				}
			}
		}
	}

	return m, nil
}

// DevicePath for the given volume i.e path where it's attached
func (ops *vsphereOps) DevicePath(diskPath string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vmObj, err := ops.renewVM(ctx, ops.vm)
	if err != nil {
		return "", err
	}

	attached, err := vmObj.IsDiskAttached(ctx, diskPath)
	if err != nil {
		return "", fmt.Errorf("failed to check if disk: %s is attached on vm: %s. err: %v",
			diskPath, vmObj.Name(), err)
	}

	if !attached {
		return "", fmt.Errorf("disk: %s is not attached on vm: %s", diskPath, vmObj.Name())
	}

	diskUUID, err := vmObj.Datacenter.GetVirtualDiskPage83Data(ctx, diskPath)
	if err != nil {
		logrus.Errorf("failed to get device path for disk: %s on vm: %s", diskPath, vmObj.Name())
		return "", err
	}

	return path.Join(diskByIDPath, diskSCSIPrefix+diskUUID), nil
}

func (ops *vsphereOps) Enumerate(volumeIds []*string,
	labels map[string]string,
	setIdentifier string,
) (map[string][]interface{}, error) {
	return nil, storageops.ErrNotSupported
}

// Snapshot the volume with given volumeID
func (ops *vsphereOps) Snapshot(volumeID string, readonly bool) (interface{}, error) {
	return nil, storageops.ErrNotSupported
}

// SnapshotDelete deletes the snapshot with given ID
func (ops *vsphereOps) SnapshotDelete(snapID string) error {
	return storageops.ErrNotSupported
}

// ApplyTags will apply given labels/tags on the given volume
func (ops *vsphereOps) ApplyTags(volumeID string, labels map[string]string) error {
	return storageops.ErrNotSupported
}

// RemoveTags removes labels/tags from the given volume
func (ops *vsphereOps) RemoveTags(volumeID string, labels map[string]string) error {
	return storageops.ErrNotSupported
}

// Tags will list the existing labels/tags on the given volume
func (ops *vsphereOps) Tags(volumeID string) (map[string]string, error) {
	return nil, storageops.ErrNotSupported
}

// GetVMObject fetches the VirtualMachine object corresponding to the given virtual machine uuid
func GetVMObject(ctx context.Context, conn *vclib.VSphereConnection, vmUUID string) (*vclib.VirtualMachine, error) {
	// TODO change impl below using multiple goroutines and sync.WaitGroup to make it faster
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := conn.Connect(ctx); err != nil {
		return nil, err
	}

	if len(vmUUID) == 0 {
		return nil, fmt.Errorf("virtual machine uuid is required")
	}

	datacenterObjs, err := vclib.GetAllDatacenter(ctx, conn)
	if err != nil {
		return nil, err
	}

	// Lookup in each vsphere datacenter for this virtual machine
	for _, dc := range datacenterObjs {
		vm, err := dc.GetVMByUUID(ctx, vmUUID)
		if err != nil {
			if err != vclib.ErrNoVMFound {
				logrus.Warnf("failed to find vm with uuid: %s in datacenter: %s due to err: %v", vmUUID, dc.Name(), err)
				// don't let one bad egg fail entire search. keep looking.
			} else {
				logrus.Debugf("did not find vm with uuid: %s in datacenter: %s", vmUUID, dc.Name())
			}
			continue
		}

		if vm != nil {
			return vm, nil
		}
	}

	return nil, fmt.Errorf("failed to find vm with uuid: %s in any datacenter for vc: %s", vmUUID, conn.Hostname)
}

func (ops *vsphereOps) renewVM(ctx context.Context, vm *vclib.VirtualMachine) (*vclib.VirtualMachine, error) {
	var client *govmomi.Client
	err := ops.conn.Connect(ctx)
	if err != nil {
		client, err = ops.conn.NewClient(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		client = ops.conn.GoVmomiClient
	}

	vmObj := vm.RenewVM(client)
	return &vmObj, nil
}

// getDatastoreToUseInStoragePod asks the storage resource manager to recommend a datastore
// in the given storage pod (datastore cluster) for the required disk spec
func (ops *vsphereOps) getDatastoreToUseInStoragePod(
	ctx context.Context, vmObj *vclib.VirtualMachine,
	volumeOptions *vclib.VolumeOptions, storagePod *object.StoragePod) (string, error) {
	logrus.Infof("Using storage pod: %s", storagePod.Name())

	// devices is a list of devices in the virtual machine (disks and disk controllers) that
	// will be part of the request spec to storage resource manager
	var devices object.VirtualDeviceList
	scsi, err := devices.CreateSCSIController("scsi")
	if err != nil {
		return "", err
	}

	devices = append(devices, scsi)

	controller, err := devices.FindDiskController("scsi")
	if err != nil {
		return "", err
	}

	disk := &types.VirtualDisk{
		VirtualDevice: types.VirtualDevice{
			Key: devices.NewKey(),
			Backing: &types.VirtualDiskFlatVer2BackingInfo{
				DiskMode:        string(types.VirtualDiskModePersistent),
				ThinProvisioned: types.NewBool(true),
			},
		},
		CapacityInKB: int64(volumeOptions.CapacityKB),
	}

	devices = append(devices, disk)
	devices.AssignController(disk, controller)
	deviceChange, err := devices.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
	if err != nil {
		return "", err
	}

	spec := &types.VirtualMachineConfigSpec{
		Name: vmObj.Name(),
	}

	spec.DeviceChange = deviceChange
	recommendedDatastore, err := recommendDatastore(ctx, vmObj, storagePod, spec)
	if err != nil {
		return "", err
	}

	return recommendedDatastore.Name(), nil
}

// recommendedDatastore recommends a datastore to use for the given storage pod by
// quering the storage resource manager
// logic borrowwed from recommendDatastore() at https://github.com/vmware/govmomi/blob/master/govc/vm/create.go#L455
func recommendDatastore(
	ctx context.Context,
	vmObj *vclib.VirtualMachine,
	storagePod *object.StoragePod,
	spec *types.VirtualMachineConfigSpec) (*object.Datastore, error) {
	sp := storagePod.Reference()

	// Build pod selection spec from config spec
	podSelectionSpec := types.StorageDrsPodSelectionSpec{
		StoragePod: &sp,
	}

	for _, deviceConfigSpec := range spec.DeviceChange {
		s := deviceConfigSpec.GetVirtualDeviceConfigSpec()
		if s.Operation != types.VirtualDeviceConfigSpecOperationAdd {
			continue
		}

		if s.FileOperation != types.VirtualDeviceConfigSpecFileOperationCreate {
			continue
		}

		d, ok := s.Device.(*types.VirtualDisk)
		if !ok {
			continue
		}

		podConfigForPlacement := types.VmPodConfigForPlacement{
			StoragePod: sp,
			Disk: []types.PodDiskLocator{
				{
					DiskId:          d.Key,
					DiskBackingInfo: d.Backing,
				},
			},
		}

		podSelectionSpec.InitialVmConfig = append(podSelectionSpec.InitialVmConfig, podConfigForPlacement)
	}

	resourcePool, err := vmObj.ResourcePool(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get vm resource pool due to: %v", err)
	}

	if resourcePool == nil {
		return nil, fmt.Errorf("failed to get vm resource pool")
	}

	resourcePoolRef := resourcePool.Reference()

	sps := types.StoragePlacementSpec{
		Type:             string(types.StoragePlacementSpecPlacementTypeCreate),
		PodSelectionSpec: podSelectionSpec,
		ConfigSpec:       spec,
		ResourcePool:     &resourcePoolRef,
	}

	srm := object.NewStorageResourceManager(vmObj.Client())
	result, err := srm.RecommendDatastores(ctx, sps)
	if err != nil {
		logrus.Errorf("failed to get datastore recommendations due to: %v", err)
		return nil, err
	}

	// Use result to pin disks to recommended datastores
	recs := result.Recommendations
	if len(recs) == 0 {
		return nil, fmt.Errorf("no datastores recommendations")
	}

	ds := recs[0].Action[0].(*types.StoragePlacementAction).Destination

	var mds mo.Datastore
	err = property.DefaultCollector(vmObj.Client()).RetrieveOne(ctx, ds, []string{"name"}, &mds)
	if err != nil {
		return nil, err
	}

	datastore := object.NewDatastore(vmObj.Client(), ds)
	datastore.InventoryPath = mds.Name

	return datastore, nil
}

// IsStoragePod checks if the object with given name is a StoragePod (Datastore cluster)
func IsStoragePod(ctx context.Context, vmObj *vclib.VirtualMachine, name string) (bool, *object.StoragePod, error) {
	f := find.NewFinder(vmObj.Client(), true)
	f.SetDatacenter(vmObj.Datacenter.Datacenter)
	sp, err := f.DatastoreCluster(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil, nil
		}

		logrus.Errorf("got error: %v fetching datastore cluster: %s", err, name)
		return false, nil, err
	}

	if sp == nil {
		return false, nil, nil
	}

	return true, sp, nil
}
