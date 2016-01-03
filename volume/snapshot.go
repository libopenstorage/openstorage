package volume

import "github.com/libopenstorage/openstorage/api"

type SnapshotNotSupported struct {}

func (s *SnapshotNotSupported) Snapshot(
	volumeID string,
	readonly bool,
	locator *api.VolumeLocator,
) (string, error) {
	return "", ErrNotSupported
}
