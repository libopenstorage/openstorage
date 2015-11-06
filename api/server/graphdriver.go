package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/docker/docker/daemon/graphdriver"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/graph"
)

const (
	// GraphDriver is the string returned in the handshake protocol.
	GraphDriver        = "GraphDriver"
	DefaultGraphDriver = "overlay"
)

// Implementation of the Docker GraphgraphDriver plugin specification.
type graphDriver struct {
	restBase
	gd graphdriver.Driver
}

type GraphResponse struct {
	Err error
}

func newGraphPlugin(name string) restServer {
	return &graphDriver{restBase{name: name, version: "0.3"}, nil}
}

func (d *graphDriver) String() string {
	return d.name
}

func graphDriverPath(method string) string {
	return fmt.Sprintf("/%s.%s", GraphDriver, method)
}

func (d *graphDriver) Routes() []*Route {
	return []*Route{
		&Route{verb: "POST", path: graphDriverPath("Init"), fn: d.init},
		&Route{verb: "POST", path: graphDriverPath("Create"), fn: d.create},
		&Route{verb: "POST", path: graphDriverPath("Remove"), fn: d.remove},
		&Route{verb: "POST", path: graphDriverPath("Get"), fn: d.get},
		&Route{verb: "POST", path: graphDriverPath("Put"), fn: d.put},
		&Route{verb: "POST", path: graphDriverPath("Exists"), fn: d.exists},
		&Route{verb: "POST", path: graphDriverPath("Status"), fn: d.graphStatus},
		&Route{verb: "POST", path: graphDriverPath("GetMetadata"), fn: d.getMetadata},
		&Route{verb: "POST", path: graphDriverPath("Cleanup"), fn: d.cleanup},
		&Route{verb: "POST", path: graphDriverPath("Diff"), fn: d.diff},
		&Route{verb: "POST", path: graphDriverPath("Changes"), fn: d.changes},
		&Route{verb: "POST", path: graphDriverPath("ApplyDiff"), fn: d.applyDiff},
		&Route{verb: "POST", path: graphDriverPath("DiffSize"), fn: d.diffSize},
		&Route{verb: "POST", path: "/Plugin.Activate", fn: d.handshake},
	}
}

func (d *graphDriver) emptyResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(&GraphResponse{})
}

func (d *graphDriver) errResponse(method string, w http.ResponseWriter, err error) {
	d.logReq("ErrReponse", method).Warnf("%v", err)
	json.NewEncoder(w).Encode(&GraphResponse{Err: err})
}

func (d *graphDriver) decode(method string, w http.ResponseWriter, r *http.Request) (*volumeRequest, error) {
	var request volumeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		e := fmt.Errorf("Unable to decode JSON payload")
		d.sendError(method, "", w, e.Error()+":"+err.Error(), http.StatusBadRequest)
		return nil, e
	}
	d.logReq(method, request.Name).Debug("")
	return &request, nil
}

func (d *graphDriver) handshake(w http.ResponseWriter, r *http.Request) {
	h := struct {
		Implements []string
	}{Implements: []string{GraphDriver}}

	err := json.NewEncoder(w).Encode(&h)
	if err != nil {
		d.sendError("handshake", "", w, "encode error", http.StatusInternalServerError)
		return
	}
	d.logReq("handshake", "").Debug("Handshake completed")
}

func (d *graphDriver) status(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, fmt.Sprintln("osd graphgraphDriver", d.version))
}

func (d *graphDriver) decodeError(method string, w http.ResponseWriter, err error) {
	e := fmt.Errorf("Unable to decode JSON payload")
	d.sendError(method, "", w, e.Error()+":"+err.Error(), http.StatusBadRequest)
	return
}

func (d *graphDriver) init(w http.ResponseWriter, r *http.Request) {
	method := "init"
	var request struct {
		Home string
		Opts []string
	}
	var name string
	d.logReq(method, request.Home).Info("")
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		d.decodeError(method, w, err)
		return
	}
	for _, v := range request.Opts {
		opt := strings.Split(v, ":")
		if len(opt) == 2 {
			if opt[0] == "name" {
				name = opt[1]
			}
		}
		if len(name) == 0 {
			name = DefaultGraphDriver
		}
	}
	gd, err := graph.New(name, request.Home, request.Opts)
	if err != nil {
		d.errResponse(method, w, err)
		return
	}
	d.gd = gd
	d.emptyResponse(w)
}

func (d *graphDriver) create(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID     string
		Parent string
	}
	method := "create"
	d.logReq(method, request.ID).Info("")
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		d.decodeError(method, w, err)
		return
	}
	if err := d.gd.Create(request.ID, request.Parent); err != nil {
		d.errResponse(method, w, err)
		return
	}
	d.emptyResponse(w)
}

func (d *graphDriver) remove(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID string
	}
	method := "remove"
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		d.decodeError(method, w, err)
		return
	}
	d.logReq(method, request.ID).Info("")
	if err := d.gd.Remove(request.ID); err != nil {
		d.errResponse(method, w, err)
		return
	}
	d.emptyResponse(w)
}

func (d *graphDriver) get(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID         string
		MountLabel string
	}
	var response struct {
		Dir string
		GraphResponse
	}
	method := "get"
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		d.decodeError(method, w, err)
		return
	}
	d.logReq(method, request.ID).Info("")
	response.Dir, response.Err = d.gd.Get(request.ID, request.MountLabel)
	if response.Err != nil {
		d.logReq("ErrReponse", method).Warnf("%v", response.Err)
	}
	json.NewEncoder(w).Encode(&response)
}

func (d *graphDriver) put(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID string
	}
	method := "put"
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		d.decodeError(method, w, err)
		return
	}
	d.logReq(method, request.ID).Info("")
	err := d.gd.Put(request.ID)
	if err != nil {
		d.errResponse(method, w, err)
		return
	}
	d.emptyResponse(w)
}

func (d *graphDriver) exists(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID string
	}
	var response struct {
		Exists bool
	}
	method := "put"
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		d.decodeError(method, w, err)
		return
	}
	d.logReq(method, request.ID).Info("")
	response.Exists = d.gd.Exists(request.ID)
	json.NewEncoder(w).Encode(&response)
}

func (d *graphDriver) graphStatus(w http.ResponseWriter, r *http.Request) {
	var response struct {
		Status [][2]string
	}
	method := "status"
	d.logReq(method, "").Info("")
	response.Status = d.gd.Status()
	json.NewEncoder(w).Encode(&response)
}

func (d *graphDriver) getMetadata(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID string
	}
	var response struct {
		Metadata map[string]string
		GraphResponse
	}
	method := "getMetadata"
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		d.decodeError(method, w, err)
		return
	}
	d.logReq(method, request.ID).Info("")
	response.Metadata, response.Err = d.gd.GetMetadata(request.ID)
	if response.Err != nil {
		d.logReq("ErrReponse", method).Warnf("%v", response.Err)
	}
	json.NewEncoder(w).Encode(&response)
}

func (d *graphDriver) cleanup(w http.ResponseWriter, r *http.Request) {
	method := "cleanup"
	var response GraphResponse
	d.logReq(method, "").Info("")
	response.Err = d.gd.Cleanup()
	if response.Err != nil {
		d.logReq("ErrReponse", method).Warnf("%v", response.Err)
	}
	json.NewEncoder(w).Encode(&response)
}

func (d *graphDriver) diff(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID     string
		Parent string
	}
	method := "diff"
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		d.decodeError(method, w, err)
		return
	}
	d.logReq(method, request.ID).Info("")
}

func (d *graphDriver) changes(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID     string
		Parent string
	}
	var response struct {
		Changes []api.GraphDriverChanges
		GraphResponse
	}
	method := "changes"
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		d.decodeError(method, w, err)
		return
	}
	d.logReq(method, request.ID).Info("")
	json.NewEncoder(w).Encode(&response)
}

func (d *graphDriver) applyDiff(w http.ResponseWriter, r *http.Request) {
	var response struct {
		Size uint64
		GraphResponse
	}
	method := "applyDirff"
	// XXX Input is a Tar stream.
	d.logReq(method, "").Info("")
	json.NewEncoder(w).Encode(&response)
}

func (d *graphDriver) diffSize(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID     string
		Parent string
	}
	var response struct {
		Size uint64
		GraphResponse
	}
	method := "applyDirff"
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		d.decodeError(method, w, err)
		return
	}
	d.logReq(method, request.ID).Info("")
	json.NewEncoder(w).Encode(&response)
}
