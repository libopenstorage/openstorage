package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/portworx/kvdb"
	"github.com/sdeoras/openstorage/osdconfig"
	"github.com/sdeoras/openstorage/osdconfig/proto"
	"golang.org/x/net/context"
)

type osdconfigAPI struct {
	name string
	restBase
}

func (p *osdconfigAPI) Routes() []*Route {
	return []*Route{
		{verb: "GET", path: "/cluster", fn: p.getClusterSpec},
		{verb: "PUT", path: "/cluster", fn: p.setClusterSpec},
		{verb: "GET", path: "/node/{id}", fn: p.getNodeSpec},
		{verb: "PUT", path: "/node", fn: p.setNodeSpec},
	}
}

func (p *osdconfigAPI) parseID(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	if id, ok := vars["id"]; ok {
		return string(id), nil
	}
	return "", fmt.Errorf("could not parse node ID")
}

func (p *osdconfigAPI) getConnector() (*osdconfig.OsdConfig, error) {
	options := make(map[string]string)
	options["KvUseInterface"] = ""
	kv, err := kvdb.New("pwx/test", "", nil, options, nil)
	if err != nil {
		return nil, err
	}
	client := osdconfig.NewKVDBConnection(kv)
	return client, nil
}

func (p *osdconfigAPI) setClusterSpec(w http.ResponseWriter, r *http.Request) {
	method := "setClusterSpec"
	client, err := p.getConnector()
	if err != nil {
		p.sendError(p.name, method, w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	config := new(proto.ClusterConfig)
	if err := json.NewDecoder(r.Body).Decode(config); err != nil {
		p.sendError(p.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	if ack, err := client.SetClusterSpec(context.Background(), config); err != nil {
		p.sendError(p.name, method, w, err.Error(), http.StatusServiceUnavailable)
		return
	} else {
		json.NewEncoder(w).Encode(ack)
	}
}

func (p *osdconfigAPI) getClusterSpec(w http.ResponseWriter, r *http.Request) {
	method := "getClusterSpec"
	client, err := p.getConnector()
	if err != nil {
		p.sendError(p.name, method, w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	config, err := client.GetClusterSpec(context.Background())
	if err != nil {
		p.sendError(p.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(config)
}

func (p *osdconfigAPI) getNodeSpec(w http.ResponseWriter, r *http.Request) {
	method := "getNodeSpec"
	client, err := p.getConnector()
	if err != nil {
		p.sendError(p.name, method, w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	nodeID, err := p.parseID(r)
	if err != nil {
		e := fmt.Errorf("failed to parse parse nodeID: %s", err.Error())
		p.sendError(p.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	config, err := client.GetNodeSpec(context.Background(), &proto.NodeID{ID: nodeID})
	if err != nil {
		p.sendError(p.name, method, w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	json.NewEncoder(w).Encode(config)
}

func (p *osdconfigAPI) setNodeSpec(w http.ResponseWriter, r *http.Request) {
	method := "setNodeSpec"
	client, err := p.getConnector()
	if err != nil {
		p.sendError(p.name, method, w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	config := new(proto.NodeConfig)
	if err := json.NewDecoder(r.Body).Decode(config); err != nil {
		p.sendError(p.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	if ack, err := client.SetNodeSpec(context.Background(), config); err != nil {
		p.sendError(p.name, method, w, err.Error(), http.StatusServiceUnavailable)
		return
	} else {
		json.NewEncoder(w).Encode(ack)
	}
}
