package graph

import (
	"github.com/docker/docker/daemon/graphdriver/overlay"
)

func init() {
	Register("overlay", overlay.Init)
}
