package volume

import "github.com/libopenstorage/openstorage/api"

type SnapshotNotSupported struct {
}

func (s *SnapshotNotSupported) Snapshot(volumeID api.VolumeID, readonly bool, locator api.VolumeLocator) (api.VolumeID, error) {
	return api.BadVolumeID, ErrNotSupported
}
