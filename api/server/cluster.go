package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/config"
)

type clusterApi struct {
	restBase
}

func (c *clusterApi) Routes() []*Route {
	return []*Route{
		&Route{verb: "GET", path: "/versions", fn: c.versions},
		&Route{verb: "GET", path: clusterPath("/enumerate", config.Version), fn: c.enumerate},
		&Route{verb: "GET", path: clusterPath("/status", config.Version), fn: c.status},
		&Route{verb: "GET", path: clusterPath("/inspect/{id}", config.Version), fn: c.inspect},
		&Route{verb: "DELETE", path: clusterPath("", config.Version), fn: c.delete},
		&Route{verb: "DELETE", path: clusterPath("/{id}", config.Version), fn: c.delete},
		&Route{verb: "PUT", path: clusterPath("/enablegossip", config.Version), fn: c.enableGossip},
		&Route{verb: "PUT", path: clusterPath("/disablegossip", config.Version), fn: c.disableGossip},
		&Route{verb: "PUT", path: clusterPath("/shutdown", config.Version), fn: c.shutdown},
		&Route{verb: "PUT", path: clusterPath("/shutdown/{id}", config.Version), fn: c.shutdown},
	}
}
func newClusterAPI() restServer {
	return &clusterApi{restBase{version: config.Version, name: "Cluster API"}}
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

func (c *clusterApi) setSize(w http.ResponseWriter, r *http.Request) {
	method := "set size"
	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := r.URL.Query()

	size := params["size"]
	if size == nil {
		c.sendError(c.name, method, w, "Missing size param", http.StatusBadRequest)
		return
	}

	sz, _ := strconv.Atoi(size[0])

	err = inst.SetSize(sz)

	clusterResponse := &api.ClusterResponse{Error: err.Error()}
	json.NewEncoder(w).Encode(clusterResponse)
}

func (c *clusterApi) inspect(w http.ResponseWriter, r *http.Request) {
	method := "inspect"
	c.sendNotImplemented(w, method)
}

func (c *clusterApi) enableGossip(w http.ResponseWriter, r *http.Request) {
	method := "enablegossip"

	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	inst.EnableUpdates()

	clusterResponse := &api.ClusterResponse{}
	json.NewEncoder(w).Encode(clusterResponse)
}

func (c *clusterApi) disableGossip(w http.ResponseWriter, r *http.Request) {
	method := "disablegossip"

	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	inst.DisableUpdates()

	clusterResponse := &api.ClusterResponse{}
	json.NewEncoder(w).Encode(clusterResponse)
}

func (c *clusterApi) status(w http.ResponseWriter, r *http.Request) {
	method := "status"

	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := inst.GetState()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func (c *clusterApi) delete(w http.ResponseWriter, r *http.Request) {
	method := "delete"
	c.sendNotImplemented(w, method)
}

func (c *clusterApi) shutdown(w http.ResponseWriter, r *http.Request) {
	method := "shutdown"
	c.sendNotImplemented(w, method)
}

func (c *clusterApi) versions(w http.ResponseWriter, r *http.Request) {
	versions := []string{
		config.Version,
		// Update supported versions by adding them here
	}
	json.NewEncoder(w).Encode(versions)
}


func (c *clusterApi) sendNotImplemented(w http.ResponseWriter, method string) {
	c.sendError(c.name, method, w, "Not implemented.", http.StatusNotImplemented)
}

func clusterVersion(route, version string) string {
	return "/" + version + "/" + route
}

func clusterPath(route, version string) string {
	return clusterVersion("cluster" + route, version)
}
