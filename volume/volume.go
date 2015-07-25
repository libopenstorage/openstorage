package volume

import (
	"errors"
	api "github.com/libopenstorage/api"
)

var (
	defaultDriver   VolumeDriver
	ErrNotSupported = errors.New("Driver implementation not supported")
	ErrEexist       = errors.New("Default Driver is already set")
	ErrEnomem       = errors.New("Out of memory")
)

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

func Default() VolumeDriver {
	if defaultDriver == nil {
		panic("No default volume driver.")
	}

	return defaultDriver
}

func SetDefaultDriver(driver VolumeDriver) {
	defaultDriver = driver
}

func Shutdown() {
	if defaultDriver != nil {
		defaultDriver.Shutdown()
	}
}

func init() {
}
