package all

import (
	"github.com/libopenstorage/openstorage/volume/drivers"
	"github.com/libopenstorage/openstorage/volume/drivers/aws"
	"github.com/libopenstorage/openstorage/volume/drivers/btrfs"
	"github.com/libopenstorage/openstorage/volume/drivers/buse"
	"github.com/libopenstorage/openstorage/volume/drivers/coprhd"
	"github.com/libopenstorage/openstorage/volume/drivers/nfs"
	"github.com/libopenstorage/openstorage/volume/drivers/pwx"
	"github.com/libopenstorage/openstorage/volume/drivers/vfs"
)

var (
	// AllDrivers is a slice of all existing known Drivers.
	AllDrivers = []volumedrivers.Driver{
		// AWS driver provisions storage from EBS.
		{DriverType: aws.Type, Name: aws.Name, Init: aws.Init},
		// BTRFS driver provisions storage from local btrfs.
		{DriverType: btrfs.Type, Name: btrfs.Name, Init: btrfs.Init},
		// BUSE driver provisions storage from local volumes and implements block in user space.
		{DriverType: buse.Type, Name: buse.Name, Init: buse.Init},
		// COPRHD driver
		{DriverType: coprhd.Type, Name: coprhd.Name, Init: coprhd.Init},
		// NFS driver provisions storage from an NFS server.
		{DriverType: nfs.Type, Name: nfs.Name, Init: nfs.Init},
		// PWX driver provisions storage from PWX cluster.
		{DriverType: pwx.Type, Name: pwx.Name, Init: pwx.Init},
		// VFS driver provisions storage from local filesystem
		{DriverType: vfs.Type, Name: vfs.Name, Init: vfs.Init},
	}
)

func init() {
	for _, driver := range AllDrivers {
		_ = volumedrivers.Add(driver.Name, driver.Init)
	}
}
