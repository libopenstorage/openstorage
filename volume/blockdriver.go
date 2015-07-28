package volume

import (
	"errors"

	"github.com/libopenstorage/openstorage/api"
)

var (
	ErrNotSupported = errors.New("Operation not supported")
)

// DefaultBlockDriver is a default (null) block driver implementation.  This can be
// used by drivers that do not want to (or care about) implement the attach,
// format and detach interfaces.
type DefaultBlockDriver struct {
}

func (d *DefaultBlockDriver) Attach(volumeID api.VolumeID) (string, error) {
	return "", ErrNotSupported
}

func (d *DefaultBlockDriver) Format(volumeID api.VolumeID) error {
	return ErrNotSupported
}

func (d *DefaultBlockDriver) Detach(volumeID api.VolumeID) error {
	return ErrNotSupported
}
