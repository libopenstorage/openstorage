package main

import (
	"github.com/libopenstorage/openstorage/drivers/aws"
	"github.com/libopenstorage/openstorage/drivers/btrfs"
	"github.com/libopenstorage/openstorage/drivers/vfs"
	"github.com/libopenstorage/openstorage/drivers/nfs"
	"github.com/libopenstorage/openstorage/drivers/pwx"
	"github.com/libopenstorage/openstorage/volume"
)

// Driver is the description of a supported OST driver. New Drivers are added to
// the drivers array
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
		// PWX driver provisions storage from PWX cluster.
		{driverType: pwx.Type, name: pwx.Name},
		// VFS driver provisions storage from local filesystem
		{driverType: vfs.Type, name: vfs.Name},
	}
)
