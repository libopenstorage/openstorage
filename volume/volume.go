package volume

import (
	"errors"
	"fmt"
	api "github.com/libopenstorage/api"
)

var (
	defDriver       VolDriver
	protoDriver     ProtocolDriver
	drivers         map[string]InitFunc
	ErrNotSupported = errors.New("Driver implementation not supported")
	ErrEexist       = errors.New("Default Driver is already set")
	ErrEnomem       = errors.New("Out of memory")
)

type InitFunc func(
	globalVersion uint64,
	// XXX spec []*api.StorageSpec,
	deviceInfo []*api.VolumeInfo) (VolDriver, error)

// notification chan *api.Notification) (VolDriver, error)

type ProtocolDriver interface {
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

type VolDriver interface {
	// String Description of this driver
	String() string

	// Create a new Vol for the specific volume spec.
	// It returns a system generated VolumeID that uniquely identifies the volume
	// If CreateOptions.FailIfExists is set and a volume matching the locator
	// exists then this will fail with ErrEexist. Otherwise if a matching available
	// volume is found then it is returned instead of creating a new volume.
	Create(volumeInfo *api.VolumeInfo) error

	// Attach Maps device to the host.
	// On success the devicePath specifies location where the device is exported
	// An error is returned if the volume is already attached, the volumeID doesn't
	// exist or an internal error occurs.
	Attach(volumeInfo *api.VolumeInfo) (minor int32, devicePath string, err error)

	// AttachInfo provides device minor and device path if the device is attached.
	// Return an error is the device is not attached.
	AttachInfo(volumeInfo *api.VolumeInfo) (minor int32, devicePath string, err error)

	// Detach device from the host.
	// An error is returned if the volume was not attached.
	Detach(volumeID *api.VolumeInfo) error

	// Delete volume.
	// This will return an error if there are any snaps associated with the volume.
	Delete(volumeInfo *api.VolumeInfo) error

	// Snapshot specified volume. IO to the underlying volume should be quiesced before
	// calling this function.
	// On success, system generated label is returned.
	Snapshot(volumeID api.VolumeID, lables *api.Labels) (snapID api.SnapID, err error)

	// SnapDelete snap specified by snapID.
	SnapDelete(snapID api.SnapID) error

	// SnapInspect provides details on this snapshot.
	SnapInspect(snapID api.SnapID) (snap api.VolumeSnap, err error)

	// Stats for specified volume.
	Stats(volumeID api.VolumeID) (stats api.VolumeStats, err error)

	// Alerts on this volume.
	Alerts(volumeID api.VolumeID) (stats api.VolumeAlerts, err error)

	// Shutdown and cleanup.
	Shutdown()
}

func ProtoDriver() ProtocolDriver {
	return protoDriver
}

func Default() VolDriver {
	return defDriver
}

func SetProtoDriver(handler ProtocolDriver) {
	protoDriver = handler
}

// func setDefault(name string, version uint64, s []*api.StorageSpec, d []*api.VolumeInfo) error {
func setDefault(name string, version uint64, d []*api.VolumeInfo) error {
	var err error
	if initFunc, exists := drivers[name]; exists {
		// defDriver, err = initFunc(version, s, d, notification)
		defDriver, err = initFunc(version, d)
		return err
	}
	return ErrNotSupported
}

func Register(name string, initFunc InitFunc) error {
	if _, exists := drivers[name]; exists {
		return fmt.Errorf("Driver provider (%s) is already registered", name)
	}
	drivers[name] = initFunc
	return nil
}

func Shutdown() {
	if defDriver != nil {
		defDriver.Shutdown()
	}
	// close(notification)
}

func init() {
	drivers = make(map[string]InitFunc)
}
