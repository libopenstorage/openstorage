package server

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	"github.com/libopenstorage/openstorage/osdconfig"
)

// swagger:operation GET /config/cluster config getClusterConfig
//
// Get cluster configuration.
//
// This will return the requested cluster configuration object
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: a cluster config
//      schema:
//       $ref: '#/definitions/ClusterConfig'
func (c *clusterApi) getClusterConf(w http.ResponseWriter, r *http.Request) {
	method := "getClusterConf"
	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	config, err := inst.GetClusterConf()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(config)
}

// swagger:operation GET /config/node/{id} config getNodeConfig
//
// Get node configuration.
//
// This will return the requested node configuration object
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get node with
//   required: true
//   type: string
// responses:
//   '200':
//      description: a node
//      schema:
//       $ref: '#/definitions/NodeConfig'
func (c *clusterApi) getNodeConf(w http.ResponseWriter, r *http.Request) {
	method := "getNodeConf"
	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	config, err := inst.GetNodeConf(vars["id"])
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(config)
}

// swagger:operation GET /config/enumerate config enumerate
//
// Get configuration for all nodes.
//
// This will return the node configuration for all nodes
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: node config enumeration
//      schema:
//       $ref: '#/definitions/NodesConfig'
func (c *clusterApi) enumerateConf(w http.ResponseWriter, r *http.Request) {
	method := "enumerateConf"
	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	config, err := inst.EnumerateNodeConf()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(config)
}

// swagger:operation DELETE /config/node/{id} config deleteNodeConfig
//
// Delete node configuration.
//
// This will delete the requested node configuration object
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to reference node
//   required: true
//   type: string
// responses:
//   '200':
//      description: success
func (c *clusterApi) delNodeConf(w http.ResponseWriter, r *http.Request) {
	method := "delNodeConf"
	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	if err := inst.DeleteNodeConf(vars["id"]); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// swagger:operation POST /config/cluster config setClusterConfig
//
// Set cluster configuration.
//
// This will set the requested cluster configuration
//
// ---
// produces:
// - application/json
// parameters:
// - name: config
//   in: body
//   description: cluster config json
//   required: true
//   schema:
//    $ref: '#/definitions/ClusterConfig'
// responses:
//   '200':
//     description: success
//     schema:
//       type: string
func (c *clusterApi) setClusterConf(w http.ResponseWriter, r *http.Request) {
	method := "setClusterConf"
	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(data) > 2 {
		data = data[1 : len(data)-1]
	} else {
		c.sendError(c.name, method, w, "incorrect form input", http.StatusInternalServerError)
		return
	}

	data, err = base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	config := new(osdconfig.ClusterConfig)
	if err := json.Unmarshal(data, config); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := inst.SetClusterConf(config); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(config)
}

// swagger:operation POST /config/node config setNodeConfig
//
// Set node configuration.
//
// This will set the requested node configuration
//
// ---
// produces:
// - application/json
// parameters:
// - name: config
//   in: body
//   description: node config json
//   required: true
//   schema:
//     $ref: '#/definitions/NodeConfig'
// responses:
//   '200':
//      description: success
func (c *clusterApi) setNodeConf(w http.ResponseWriter, r *http.Request) {
	method := "setNodeConf"
	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(data) > 2 {
		data = data[1 : len(data)-1]
	} else {
		c.sendError(c.name, method, w, "incorrect form input", http.StatusInternalServerError)
		return
	}

	data, err = base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	config := new(osdconfig.NodeConfig)
	if err := json.Unmarshal(data, config); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := inst.SetNodeConf(config); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(config)
}
