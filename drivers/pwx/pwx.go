package pwx

import (
	"github.com/libopenstorage/openstorage/api/client"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	Name           = "pwx"
	Type           = volume.Block
	DefaultUrl     = "unix:///" + config.DriverAPIBase + "pxd.sock"
	DefaultVersion = "v1"
)

type driver struct {
	volume.VolumeDriver
}

// Portworx natively implements the openstorage.org API specification, so
// we can directly point the VolumeDriver to the PWX API server.
func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	url, ok := params[config.UrlKey]
	if !ok {
		url = DefaultUrl
	}
	version, ok := params[config.VersionKey]
	if !ok {
		version = DefaultVersion
	}
	c, err := client.NewClient(url, version)
	if err != nil {
		return nil, err
	}

	return &driver{VolumeDriver: c.VolumeDriver()}, nil
}

func (d *driver) String() string {
	return Name
}

func (d *driver) Type() volume.DriverType {
	return Type
}

func init() {
	volume.Register(Name, Init)
}
