// To add a driver to openstorage, declare the driver here.
package main

import (
	"github.com/libopenstorage/openstorage/drivers/aws"
	"github.com/libopenstorage/openstorage/drivers/btrfs"
	"github.com/libopenstorage/openstorage/drivers/nfs"
	"github.com/libopenstorage/openstorage/volume"
)

type Driver struct {
	driverType volume.DriverType
	name       string
}

var (
	drivers = []Driver{
		// AWS driver. This provisions storage from EBS.
		{driverType: volume.Block,
			name: aws.Name},
		// NFS driver. This provisions storage from an NFS server.
		{driverType: volume.File,
			name: nfs.Name},
		// BTRFS driver. This provisions storage from local btrfs fs.
		{driverType: volume.File,
			name: btrfs.Name},
	}
)
