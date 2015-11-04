package volume

import (
	"github.com/libopenstorage/openstorage/api"
)

// DefaultBlockDriver is a default (null) block driver implementation.  This can be
// used by drivers that do not want to (or care about) implementing the attach,
// format and detach interfaces.
type DefaultBlockDriver struct {
}

func (d *DefaultBlockDriver) Attach(volumeID api.VolumeID) (path string, err error) {
	return "", ErrNotSupported
}

func (d *DefaultBlockDriver) Detach(volumeID api.VolumeID) error {
	return ErrNotSupported
}
