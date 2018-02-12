package server

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/libopenstorage/openstorage/osdconfig"
	"github.com/portworx/kvdb"
)

type osdconfigAPI struct {
	//restBase
	cc osdconfig.ConfigManager
}

func newOsdconfigAPI(name, ver string, ctx context.Context, kv kvdb.Kvdb) (*osdconfigAPI, error) {
	var err error
	api := new(osdconfigAPI)
	api.cc, err = osdconfig.NewManager(ctx, kv)
	if err != nil {
		return nil, err
	}
	//api.name, api.version = name, ver
	return api, nil
}

func (g *osdconfigAPI) close() {
	g.cc.Close()
}

func (g *osdconfigAPI) String() string {
	//return g.name
	return ""
}

func (g *osdconfigAPI) getHTTPFunc(name string) func(w http.ResponseWriter, r *http.Request) {
	switch name {
	case filepath.Join(get, cluster):
		fn, _ := g.cc.GetHTTPFunc(nil, g.cc.GetClusterConf)
		return fn
	case filepath.Join(get, node):
		fn, _ := g.cc.GetHTTPFunc(id, g.cc.GetNodeConf)
		return fn
	case filepath.Join(post, cluster):
		fn, _ := g.cc.GetHTTPFunc(nil, g.cc.SetClusterConf)
		return fn
	case filepath.Join(post, node):
		fn, _ := g.cc.GetHTTPFunc(nil, g.cc.SetNodeConf)
		return fn
	}
	return nil
}

func (g *osdconfigAPI) Routes() []*Route {
	return []*Route{
		{verb: "GET", path: filepath.Join("/", cluster), fn: g.getHTTPFunc(filepath.Join(get, cluster))},
		{verb: "GET", path: filepath.Join("/", node, "{"+id+"}"), fn: g.getHTTPFunc(filepath.Join(get, node))},
		{verb: "POST", path: filepath.Join("/", cluster), fn: g.getHTTPFunc(filepath.Join(post, cluster))},
		{verb: "POST", path: filepath.Join("/", node), fn: g.getHTTPFunc(filepath.Join(post, node))},
	}
}
