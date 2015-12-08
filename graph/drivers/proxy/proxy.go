package proxy

import (
	"github.com/docker/docker/daemon/graphdriver/overlay"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/graph"
)

const (
	Name = "proxy"
	Type = api.Graph
)

func init() {
	graph.Register("proxy", overlay.Init)
}
