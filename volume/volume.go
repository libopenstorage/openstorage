package volume

import (
	"errors"
	"sync"

	"github.com/libopenstorage/openstorage/api"
)

var (
	instances       map[string]VolumeDriver
	drivers         map[string]InitFunc
	mutex           sync.Mutex
	ErrNotSupported = errors.New("Driver implementation not supported")
	ErrExist        = errors.New("Driver already exists")
	ErrNotFound     = errors.New("Driver implementation not found")
)

type DriverParams map[string]string

type InitFunc func(params DriverParams) (VolumeDriver, error)

type VolumeDriver interface {
	// String description of this driver.
	String() string

	// Create a new Vol for the specific volume spec.
	// It returns a system generated VolumeID that uniquely identifies the volume
	// If CreateOptions.FailIfExists is set and a volume matching the locator
	// exists then this will fail with ErrEexist. Otherwise if a matching available
	// volume is found then it is returned instead of creating a new volume.
	Create(locator api.VolumeLocator,
		options *api.CreateOptions,
		spec *api.VolumeSpec) (volumeID api.VolumeID, err error)

	// Attach map device to the host and mount device on specified path.
	// On success the devicePath specifies location where the device is exported
	// An error is returned if the volume is already attached, the volumeID doesn't
	// exist or an internal error occurs.
	Attach(volumeID api.VolumeID, path string) (devicePath string, err error)

	// Format volume according to spec provided in Create
	// An error is returned if the volume is not detached or an internal error occurs.
	Format(volumeID api.VolumeID) error

	// Detach device from the host.
	// An error is returned if the volume was not attached.
	Detach(volumeID api.VolumeID) (err error)

	// Inspect specified volumes.
	Inspect(ids []api.VolumeID) (volume []api.Volume, err error)

	// Delete volume.
	// Returns an error if there are any snaps associated with the volume.
	Delete(volumeID api.VolumeID) (err error)

	// Enumerate volumes that map to the volumeLocator. Locator fields may be regexp.
	// If locator fields are left blank, this will return all volumes.
	// If err is set to ErrEagain then the call should be retried with empty cursor
	Enumerate(locator api.VolumeLocator, labels api.Labels) []api.Volume

	// Snap specified volume. IO to the underlying volume should be quiesced before
	// calling this function.
	// On success, system generated label is returned.
	Snapshot(volumeID api.VolumeID, lables api.Labels) (snapID api.SnapID, err error)

	// SnapDelete snap specified by snapID.
	SnapDelete(snapID api.SnapID) (err error)

	// SnapInspect provides details on this snapshot.
	SnapInspect(snapID api.SnapID) (snap api.VolumeSnap, err error)

	// Enumerate snaps for specified volume
	// Count indicates the number of snaps populated.
	// If err is set to ErrEagain then the call should be retried with empty cursor
	SnapEnumerate(locator api.VolumeLocator, labels api.Labels) *[]api.SnapID

	// Stats for specified volume.
	Stats(volumeID api.VolumeID) (stats api.VolumeStats, err error)

	// Alerts on this volume.
	Alerts(volumeID api.VolumeID) (stats api.VolumeAlerts, err error)

	// Shutdown and cleanup.
	Shutdown()
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
	return nil, ErrNotFound
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

func Register(name string, initFunc InitFunc) error {
	mutex.Lock()
	defer mutex.Unlock()
	if _, exists := drivers[name]; exists {
		return ErrExist
	}
	drivers[name] = initFunc
	return nil
}
