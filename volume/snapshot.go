package volume

import "github.com/libopenstorage/openstorage/api"

type SnapshotNotSupported struct {
}

func (s *SnapshotNotSupported) Snapshot(volumeID api.VolumeID, labels api.Labels) (api.SnapID, error) {
	return api.BadSnapID, ErrNotSupported
}

func (s *SnapshotNotSupported) SnapDelete(snapID api.SnapID) error {
	return ErrNotSupported
}

func (s *SnapshotNotSupported) Stats(volumeID api.VolumeID) (api.VolumeStats, error) {
	return api.VolumeStats{}, ErrNotSupported
}
