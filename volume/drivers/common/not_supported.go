package common

import (
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

type blockNotSupported struct{}

func (b *blockNotSupported) Attach(volumeID string) (string, error) {
	return "", volume.ErrNotSupported
}

func (b *blockNotSupported) Detach(volumeID string) error {
	return volume.ErrNotSupported
}

type snapshotNotSupported struct{}

func (s *snapshotNotSupported) Snapshot(volumeID string, readonly bool, locator *api.VolumeLocator) (string, error) {
	return "", volume.ErrNotSupported
}

type ioNotSupported struct{}

func (i *ioNotSupported) Read(volumeID string, buffer []byte, size uint64, offset int64) (int64, error) {
	return 0, volume.ErrNotSupported
}

func (i *ioNotSupported) Write(volumeID string, buffer []byte, size uint64, offset int64) (int64, error) {
	return 0, volume.ErrNotSupported
}

func (i *ioNotSupported) Flush(volumeID string) error {
	return volume.ErrNotSupported
}
