// +build !linux

package btrfs

import (
	"errors"

	"github.com/libopenstorage/openstorage/volume"
)

var (
	errUnsupported = errors.New("btrfs not supported on this platform")
)

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	return nil, errUnsupported
}
