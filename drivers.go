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
		// AWS driver provisions storage from EBS.
		{driverType: aws.Type, name: aws.Name},
		// NFS driver provisions storage from an NFS server.
		{driverType: nfs.Type, name: nfs.Name},
		// BTRFS driver provisions storage from local btrfs.
		{driverType: btrfs.Type, name: btrfs.Name},
	}
)
