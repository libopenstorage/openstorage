package graphdrivers

import (
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/graph/drivers/fuse"
	"github.com/libopenstorage/openstorage/graph/drivers/layer0"
	"github.com/libopenstorage/openstorage/graph/drivers/proxy"
)

// Driver is the description of a supported OST driver. New Drivers are added to
// the drivers array
type Driver struct {
	DriverType api.DriverType
	Name       string
}

var (
	// AllDrivers is a slice of all existing known Drivers.
	AllDrivers = []Driver{
		// FUSE driver implements the Kernel-User Fuse protocol.
		{DriverType: fuse.Type, Name: fuse.Name},
		// Layer0 driver provides persistent storage for the writable layer.
		{DriverType: layer0.Type, Name: layer0.Name},
		// Proxy driver simply uses the Docker overlay driver.
		{DriverType: proxy.Type, Name: proxy.Name},
	}
)
