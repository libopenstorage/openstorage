package volume

// DefaultBlockDriver is a default (null) block driver implementation.  This can be
// used by drivers that do not want to (or care about) implementing the attach,
// format and detach interfaces.
type DefaultBlockDriver struct {
}

func (d *DefaultBlockDriver) Attach(volumeID string) (path string, err error) {
	return "", ErrNotSupported
}

func (d *DefaultBlockDriver) Detach(volumeID string) error {
	return ErrNotSupported
}
