package volume

import "github.com/libopenstorage/openstorage/api"

type IoNotSupported struct {
}

func (io *IoNotSupported) Read(volumeID api.VolumeID,
	buf []byte,
	sz uint64,
	offset int64) (int64, error) {
	return 0, ErrNotSupported
}

func (io *IoNotSupported) Write(volumeID api.VolumeID,
	buf []byte,
	sz uint64,
	offset int64) (int64, error) {
	return 0, ErrNotSupported
}
func (io *IoNotSupported) Flush(volumeID api.VolumeID) error {
	return ErrNotSupported
}
