package volume

import (
	"errors"
	"sync"

	"github.com/libopenstorage/openstorage/api"
)

var (
	instances         map[string]VolumeDriver
	drivers           map[string]InitFunc
	mutex             sync.Mutex
	ErrExist          = errors.New("Driver already exists")
	ErrDriverNotFound = errors.New("Driver implementation not found")
	ErrEnoEnt         = errors.New("Volume does not exist.")
	ErrVolDetached    = errors.New("Volume is detached")
	ErrVolAttached    = errors.New("Volume is attached")
	ErrVolHasSnaps    = errors.New("Volume has snapshots associated")
	ErrNotSupported   = errors.New("Operation not supported")
)

type DriverParams map[string]string

type InitFunc func(params DriverParams) (VolumeDriver, error)

type DriverType string

const (
	TypeFileDriver   = DriverType("FilesystemDriver")
	TypeBlockDriver  = DriverType("BlockDriver")
	TypeObjectDriver = DriverType("ObjectDriver")
)

type VolumeDriver interface {
	ProtoDriver
	BlockDriver
	MountDriver
	Enumerator
}

type ProtoDriver interface {
	// String description of this driver.
	String() string

	// Create a new Vol for the specific volume spec.
	// It returns a system generated VolumeID that uniquely identifies the volume
	// If CreateOptions.FailIfExists is set and a volume matching the locator
	// exists then this will fail with ErrEexist. Otherwise if a matching available
	// volume is found then it is returned instead of creating a new volume.
	Create(locator api.VolumeLocator,
		options *api.CreateOptions,
		spec *api.VolumeSpec) (api.VolumeID, error)

	// Inspect specified volumes.
	// Errors ErrEnoEnt may be returned.
	Inspect(volumeIDs []api.VolumeID) ([]api.Volume, error)

	// Delete volume.
	// Errors ErrEnoEnt, ErrVolHasSnaps may be returned.
	Delete(volumeID api.VolumeID) error

	// Snap specified volume. IO to the underlying volume should be quiesced before
	// calling this function.
	// Errors ErrEnoEnt may be returned
	Snapshot(volumeID api.VolumeID, lables api.Labels) (api.SnapID, error)

	// SnapDelete snap specified by snapID.
	// Errors ErrEnoEnt may be returned
	SnapDelete(snapID api.SnapID) error

	// SnapInspect provides details on this snapshot.
	// Errors ErrEnoEnt may be returned
	SnapInspect(snapID []api.SnapID) ([]api.VolumeSnap, error)

	// Stats for specified volume.
	// Errors ErrEnoEnt may be returned
	Stats(volumeID api.VolumeID) (api.VolumeStats, error)

	// Alerts on this volume.
	// Errors ErrEnoEnt may be returned
	Alerts(volumeID api.VolumeID) (api.VolumeAlerts, error)

	// Shutdown and cleanup.
	Shutdown()
}

type Enumerator interface {
	// Enumerate volumes that map to the volumeLocator. Locator fields may be regexp.
	// If locator fields are left blank, this will return all volumes.
	Enumerate(locator api.VolumeLocator, labels api.Labels) ([]api.Volume, error)

	// Enumerate snaps for specified volume
	// Count indicates the number of snaps populated.
	SnapEnumerate(locator api.VolumeLocator, labels api.Labels) ([]api.VolumeSnap, error)
}

type BlockDriver interface {
	// Attach map device to the host.
	// On success the devicePath specifies location where the device is exported
	// Errors ErrEnoEnt, ErrVolAttached may be returned.
	Attach(volumeID api.VolumeID) (string, error)

	// Format volume according to spec provided in Create
	// Errors ErrEnoEnt, ErrVolDetached may be returned.
	Format(volumeID api.VolumeID) error

	// Detach device from the host.
	// Errors ErrEnoEnt, ErrVolDetached may be returned.
	Detach(volumeID api.VolumeID) error
}

type MountDriver interface {
	// Mount volume at specified path
	// Errors ErrEnoEnt, ErrVolDetached may be returned.
	Mount(volumeID api.VolumeID, mountpath string) error

	// Unmount volume at specified path
	// Errors ErrEnoEnt, ErrVolDetached may be returned.
	Unmount(volumeID api.VolumeID, mountpath string) error
}

func Shutdown() {
	mutex.Lock()
	defer mutex.Unlock()
	for _, v := range instances {
		v.Shutdown()
	}
}

func Get(name string) (VolumeDriver, error) {
	if v, ok := instances[name]; ok {
		return v, nil
	}
	return nil, ErrDriverNotFound
}

func New(name string, params DriverParams) (VolumeDriver, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := instances[name]; ok {
		return nil, ErrExist
	}
	if initFunc, exists := drivers[name]; exists {
		driver, err := initFunc(params)
		if err != nil {
			return nil, err
		}
		instances[name] = driver
		return driver, err
	}
	return nil, ErrNotSupported
}

func Register(name string, driverType DriverType, initFunc InitFunc) error {
	mutex.Lock()
	defer mutex.Unlock()
	if _, exists := drivers[name]; exists {
		return ErrExist
	}
	drivers[name] = initFunc
	return nil
}

func init() {
	drivers = make(map[string]InitFunc)
	instances = make(map[string]VolumeDriver)
}
