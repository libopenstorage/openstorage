package volume

import (
	"github.com/libopenstorage/openstorage/api"
)

// BlockDriver is a default (null) block driver implementation.  This can be used by drivers
// that do not want to (or care about) implement the attach, format and detach interfaces.
type DefaultBlockDriver struct {
}

func (self *DefaultBlockDriver) Attach(volumeID api.VolumeID) (devicePath string, err error) {
	return "", nil
}

func (self *DefaultBlockDriver) Format(volumeID api.VolumeID) error {
	return nil
}

func (self *DefaultBlockDriver) Detach(volumeID api.VolumeID) error {
	return nil
}
