// +build !have_fuse

package fuse

import "C"

import (
	"errors"

	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/idtools"
	"github.com/libopenstorage/openstorage/api"
)

const (
	Name = "fuse"
	Type = api.Graph
)

var (
	errUnsupported = errors.New("fuse not supported on this platform")
)

func Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) (graphdriver.Driver, error) {
	return nil, errUnsupported
}
