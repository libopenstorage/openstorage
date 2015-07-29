// To add a driver to openstorage, declare the driver here.
package main

import (
	"github.com/libopenstorage/openstorage/drivers/aws"
	"github.com/libopenstorage/openstorage/drivers/nfs"
)

var (
	drivers = []string{aws.Name, nfs.Name}
)
