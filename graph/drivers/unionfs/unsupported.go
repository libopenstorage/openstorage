// +build !have_unionfs

package unionfs

import "C"

import (
	"errors"

	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/idtools"
	"github.com/libopenstorage/openstorage/api"
)

const (
	Name = "unionfs"
	Type = api.Graph
)

var (
	errUnsupported = errors.New("unionfs not supported on this platform")
)

func Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) (graphdriver.Driver, error) {
	return nil, errUnsupported
}
