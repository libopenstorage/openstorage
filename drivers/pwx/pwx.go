package pwx

import (
	"github.com/libopenstorage/openstorage/client"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	Name           = "pwx"
	Type           = volume.Block
	DefaultUrl     = "/run/pwx/pxd.sock"
	DefaultVersion = "v1"
)

type driver struct {
	volume.VolumeDriver
}

// Portworx natively implements the openstorage.org API specification, so
// we can directly point the VolumeDriver to the PWX API server.
func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	url, ok := params["pwx"]
	if !ok {
		url = DefaultUrl
	}
	version, ok := params["version"]
	if !ok {
		version = DefaultVersion
	}
	c, err := client.NewClient(url, version)
	if err != nil {
		return nil, err
	}
	return c.VolumeDriver(), nil
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
