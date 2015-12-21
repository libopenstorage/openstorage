package common

import (
	"go.pedge.io/proto/time"

	"github.com/libopenstorage/openstorage/api"
)

// NewVolume returns a new api.Volume for a driver Create call.
func NewVolume(
	volumeID string,
	fsType api.FSType,
	volumeLocator *api.VolumeLocator,
	source *api.Source,
	volumeSpec *api.VolumeSpec,
) *api.Volume {
	return &api.Volume{
		Id:       volumeID,
		Locator:  volumeLocator,
		Ctime:    prototime.Now(),
		Spec:     volumeSpec,
		Source:   source,
		LastScan: prototime.Now(),
		Format:   fsType,
		State:    api.VolumeState_VOLUME_STATE_AVAILABLE,
		Status:   api.VolumeStatus_VOLUME_STATUS_UP,
	}
}
