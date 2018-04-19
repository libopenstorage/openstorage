package pwx

import (
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	// Name of the driver
	Name = "pwx"
	// Type of the driver
	Type = api.DriverType_DRIVER_TYPE_BLOCK
	// DefaultUrl where the driver's socket resides
	DefaultUrl = "unix:///" + volume.DriverAPIBase + "pxd.sock"
)

type driver struct {
	volume.VolumeClient
}

// Init initialized the Portworx driver.
// Portworx natively implements the openstorage.org API specification, so
// we can directly point the VolumeDriver to the PWX API server.
func Init(params map[string]string) (volume.VolumeDriver, error) {
	url, ok := params[config.UrlKey]
	if !ok {
		url = DefaultUrl
	}
	version, ok := params[config.VersionKey]
	if !ok {
		version = volume.APIVersion
	}
	c, err := client.NewClient(url, version, "")
	if err != nil {
		return nil, err
	}

	return &driver{VolumeClient: volumeclient.VolumeDriver(c)}, nil
}

func (d *driver) Name() string {
	return Name
}

func (d *driver) Type() api.DriverType {
	return Type
}

func (d *driver) Read(volumeID string, buf []byte, sz uint64, offset int64) (int64, error) {
	return 0, volume.ErrNotSupported
}

func (d *driver) Write(volumeID string, buf []byte, sz uint64, offset int64) (int64, error) {
	return 0, volume.ErrNotSupported
}

func (d *driver) Flush(volumeID string) error {
	return volume.ErrNotSupported
}

func (d *driver) MountedAt(mountPath string) string {
	return ""
}

func (d *driver) Status() [][2]string {
	return [][2]string{}
}

func (d *driver) Shutdown() {}
