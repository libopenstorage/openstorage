package volumedrivers

import (
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

var (
	volumeDriverRegistry = volume.NewVolumeDriverRegistry(nil)
)

// Driver is the description of a supported OST driver. New Drivers are added to
// the drivers array
type Driver struct {
	DriverType api.DriverType
	Name       string
	Init       func(map[string]string) (volume.VolumeDriver, error)
}

func Get(name string) (volume.VolumeDriver, error) {
	return volumeDriverRegistry.Get(name)
}

func Register(name string, params map[string]string) error {
	return volumeDriverRegistry.Register(name, params)
}

func Add(name string, init func(map[string]string) (volume.VolumeDriver, error)) error {
	return volumeDriverRegistry.Add(name, init)
}

func Shutdown() error {
	return volumeDriverRegistry.Shutdown()
}
