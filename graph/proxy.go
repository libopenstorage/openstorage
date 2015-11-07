package graph

import (
	"github.com/docker/docker/daemon/graphdriver/overlay"
)

func init() {
	Register("proxy", overlay.Init)
}
