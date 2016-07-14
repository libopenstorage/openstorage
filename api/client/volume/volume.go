package client

import (
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/api/client"	
)

// VolumeDriver returns a REST wrapper for the VolumeDriver interface.
func VolumeDriver(c *client.Client) volume.VolumeDriver {
	return newVolumeClient(c)
}
