package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/cluster"
)

type clusterApi struct {
	restBase
}

func (c *clusterApi) Routes() []*Route {
	return []*Route{
		{verb: "GET", path: "/cluster/versions", fn: c.versions},
		{verb: "GET", path: clusterPath("/enumerate", cluster.APIVersion), fn: c.enumerate},
		{verb: "GET", path: clusterPath("/gossipstate", cluster.APIVersion), fn: c.gossipState},
		{verb: "GET", path: clusterPath("/status", cluster.APIVersion), fn: c.status},
		{verb: "GET", path: clusterPath("/peerstatus", cluster.APIVersion), fn: c.peerStatus},
		{verb: "GET", path: clusterPath("/inspect/{id}", cluster.APIVersion), fn: c.inspect},
		{verb: "DELETE", path: clusterPath("", cluster.APIVersion), fn: c.delete},
		{verb: "DELETE", path: clusterPath("/{id}", cluster.APIVersion), fn: c.delete},
		{verb: "PUT", path: clusterPath("/enablegossip", cluster.APIVersion), fn: c.enableGossip},
		{verb: "PUT", path: clusterPath("/disablegossip", cluster.APIVersion), fn: c.disableGossip},
		{verb: "PUT", path: clusterPath("/shutdown", cluster.APIVersion), fn: c.shutdown},
		{verb: "PUT", path: clusterPath("/shutdown/{id}", cluster.APIVersion), fn: c.shutdown},
	}
}
func newClusterAPI() restServer {
	return &clusterApi{restBase{version: cluster.APIVersion, name: "Cluster API"}}
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

func (c *clusterApi) gossipState(w http.ResponseWriter, r *http.Request) {
	method := "gossipState"

	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := inst.GetGossipState()
	json.NewEncoder(w).Encode(resp)
}

func (c *clusterApi) status(w http.ResponseWriter, r *http.Request) {
	method := "status"

	params := r.URL.Query()
	listenerName := params["name"]
	if listenerName[0] == "" {
		c.sendError(c.name, method, w, "Missing id param", http.StatusBadRequest)
		return
	}
	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := inst.NodeStatus(listenerName[0])
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)

}

func (c *clusterApi) peerStatus(w http.ResponseWriter, r *http.Request) {
	method := "peerStatus"

	params := r.URL.Query()
	listenerName := params["name"]
	if listenerName[0] == "" {
		c.sendError(c.name, method, w, "Missing id param", http.StatusBadRequest)
		return
	}
	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := inst.PeerStatus(listenerName[0])
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)

}

func (c *clusterApi) delete(w http.ResponseWriter, r *http.Request) {
	method := "delete"

	params := r.URL.Query()

	nodeID := params["id"]
	if nodeID == nil {
		c.sendError(c.name, method, w, "Missing id param", http.StatusBadRequest)
		return
	}

	forceRemoveParam := params["forceRemove"]
	forceRemove := false
	if forceRemoveParam != nil {
		var err error
		forceRemove, err = strconv.ParseBool(forceRemoveParam[0])
		if err != nil {
			c.sendError(c.name, method, w, "Invalid forceRemove Option: "+
				forceRemoveParam[0], http.StatusBadRequest)
			return
		}
	}

	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	nodes := make([]api.Node, 0)
	for _, id := range nodeID {
		nodes = append(nodes, api.Node{Id: id})
	}

	clusterResponse := &api.ClusterResponse{}

	err = inst.Remove(nodes, forceRemove)
	if err != nil {
		clusterResponse.Error = fmt.Errorf("Node Remove: %s", err).Error()
	}
	json.NewEncoder(w).Encode(clusterResponse)
}

func (c *clusterApi) shutdown(w http.ResponseWriter, r *http.Request) {
	method := "shutdown"
	c.sendNotImplemented(w, method)
}

func (c *clusterApi) versions(w http.ResponseWriter, r *http.Request) {
	versions := []string{
		cluster.APIVersion,
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
	return clusterVersion("cluster"+route, version)
}
