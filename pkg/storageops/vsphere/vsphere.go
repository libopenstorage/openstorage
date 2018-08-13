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

	logrus.Infof("Using following configuration for vsphere:")
	logrus.Infof("  vCenter: %s:%s", cfg.VCenterIP, cfg.VCenterPort)
	logrus.Infof("  Datacenter: %s", vmObj.Datacenter.Name())
	logrus.Infof("  VMUUID: %s", cfg.VMUUID)

	return &vsphereOps{
		cfg:  cfg,
		vm:   vmObj,
		conn: vSphereConn,
	}, nil
}

// Name returns name of the storage operations driver
func (ops *vsphereOps) Name() string {
	return "vsphere"
}

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
	logrus.Infof("Using datastore: %s for new disk", datastore)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vmObj, err := ops.renewVM(ctx, ops.vm)
	if err != nil {
		return nil, err
	}

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

	return canonicalVolumePath, nil

}

func (ops *vsphereOps) GetDeviceID(diskPath interface{}) (string, error) {
	id, ok := diskPath.(string)
	if !ok {
		return "", fmt.Errorf("invalid input: %v to GetDeviceID", diskPath)
	}

	return id, nil
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
			host, err := vm.HostSystem(ctx)
			if err != nil {
				return nil, err
			}

			logrus.Infof("vm: %s uuid: %s is in datacenter: %s running on host: %v", vm.Name(), vmUUID, dc.Name(), host.Reference())
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
