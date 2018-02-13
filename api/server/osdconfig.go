package server

import (
	"context"

	"github.com/libopenstorage/openstorage/osdconfig"
	"github.com/portworx/kvdb"
)

type osdconfigAPI struct {
	restBase
	cc osdconfig.ConfigManager
}

func (g *osdconfigAPI) String() string {
	return g.name
}

func (g *osdconfigAPI) Routes() []*Route {
	routes := g.cc.GetRoutes()
	Routes := make([]*Route, len(routes))
	for i, route := range routes {
		Routes[i].verb = route.Method
		Routes[i].path = route.Path
		Routes[i].fn = route.Fn
	}
	return Routes
}

func newOsdconfigAPI(name, ver string, ctx context.Context, kv kvdb.Kvdb) (*osdconfigAPI, error) {
	var err error
	api := new(osdconfigAPI)
	api.cc, err = osdconfig.NewManager(ctx, kv)
	if err != nil {
		return nil, err
	}
	api.name, api.version = name, ver
	return api, nil
}

func (g *osdconfigAPI) close() {
	g.cc.Close()
}
