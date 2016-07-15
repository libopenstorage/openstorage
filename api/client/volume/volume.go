package volume

import (
	"github.com/libopenstorage/openstorage/api/client"
	"github.com/libopenstorage/openstorage/volume"
)

// VolumeDriver returns a REST wrapper for the VolumeDriver interface.
func VolumeDriver(c *client.Client) volume.VolumeDriver {
	return newVolumeClient(c)
}
