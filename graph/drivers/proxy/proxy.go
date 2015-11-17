package graph

import (
	"github.com/docker/docker/daemon/graphdriver/overlay"
	"github.com/libopenstorage/openstorage/graph"
)

func init() {
	graph.Register("proxy", overlay.Init)
}
