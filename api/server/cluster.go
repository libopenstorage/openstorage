package server

import (
	"encoding/json"
	"net/http"

	"github.com/libopenstorage/openstorage/cluster"
)

const (
	clusterApiVersion = "v1"
)

type clusterApi struct {
	restBase
}

func newClusterAPI(name string) restServer {
	return &clusterApi{restBase{version: clusterApiVersion, name: name}}
}

func (c *clusterApi) String() string {
	return c.name
}

func (c *clusterApi) enumerate(w http.ResponseWriter, r *http.Request) {
	method := "enumerate"

	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	cluster, err := inst.Enumerate()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(cluster)
}

func (c *clusterApi) inspect(w http.ResponseWriter, r *http.Request) {
	method := "inspect"

	c.sendError(c.name, method, w, "Not implemented.", http.StatusNotImplemented)
}

func (c *clusterApi) delete(w http.ResponseWriter, r *http.Request) {
	method := "delete"

	c.sendError(c.name, method, w, "Not implemented.", http.StatusNotImplemented)
}

func (c *clusterApi) shutdown(w http.ResponseWriter, r *http.Request) {
	method := "shutdown"

	c.sendError(c.name, method, w, "Not implemented.", http.StatusNotImplemented)
}

func clusterVersion(route string) string {
	return "/" + volApiVersion + "/" + route
}

func clusterPath(route string) string {
	return clusterVersion("cluster" + route)
}

func (c *clusterApi) Routes() []*Route {
	return []*Route{
		&Route{verb: "GET", path: clusterPath("/enumerate"), fn: c.enumerate},
		&Route{verb: "GET", path: clusterPath("/inspect/{id}"), fn: c.inspect},
		&Route{verb: "DELETE", path: clusterPath(""), fn: c.delete},
		&Route{verb: "DELETE", path: clusterPath("/{id}"), fn: c.delete},
		&Route{verb: "PUT", path: snapPath("shutdown"), fn: c.shutdown},
		&Route{verb: "PUT", path: snapPath("shutdown/{id}"), fn: c.shutdown},
	}
}
