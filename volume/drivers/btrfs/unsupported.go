// +build !have_btrfs

package btrfs

import (
	"errors"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	Name      = "btrfs"
	Type      = api.File
	RootParam = "home"
	Volumes   = "volumes"
)

var (
	errUnsupported = errors.New("btrfs not supported on this platform")
)

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	return nil, errUnsupported
}
