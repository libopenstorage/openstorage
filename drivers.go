// To add a driver to openstorage, declare the driver here.
package main

import (
	"github.com/libopenstorage/openstorage/drivers/aws"
	"github.com/libopenstorage/openstorage/drivers/btrfs"
	"github.com/libopenstorage/openstorage/drivers/nfs"
)

var (
	drivers = []string{
		// AWS driver. This provisions storage from EBS.
		aws.Name,
		// NFS driver. This provisions storage from an NFS server.
		nfs.Name,
		// BTRFS driver. This provisions storage from local btrfs fs.
		btrfs.Name,
	}
)
