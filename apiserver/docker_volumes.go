package apiserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	types "github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	VolumeDriver = "VolumeDriver"
)

type driver struct {
	restBase
}

type handshakeResp struct {
	Implements []string
}

type volumeRequest struct {
	Name string
}

type volumeResponse struct {
	Err string
}
type volumePathResponse struct {
	Mountpoint string
	Err        string
}

type volumeInfo struct {
	metadata string
	vol      *types.Volume
}

func NewVolumePlugin(name string) restServer {
	return &driver{restBase{name: name, version: "0.3"}}
}

func (d *driver) String() string {
	return d.name
}

func volDriverPath(method string) string {
	return fmt.Sprintf("/%s.%s", VolumeDriver, method)
}

func (d *driver) Routes() []*Route {
	return []*Route{
		&Route{verb: "POST", path: volDriverPath("Create"), fn: d.create},
		&Route{verb: "POST", path: volDriverPath("Remove"), fn: d.remove},
		&Route{verb: "POST", path: volDriverPath("Mount"), fn: d.mount},
		&Route{verb: "POST", path: volDriverPath("Path"), fn: d.path},
		&Route{verb: "POST", path: volDriverPath("Unmount"), fn: d.unmount},
		&Route{verb: "POST", path: "/Plugin.Activate", fn: d.handshake},
		&Route{verb: "GET", path: "/status", fn: d.status},
	}
}

func (d *driver) emptyResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driver) volFromName(name string) (*volumeInfo, error) {
	s := strings.Split(name, ":")
	if len(s) != 2 {
		return nil, fmt.Errorf("Invalid d.logReq for name %s", name)
	}
	id, err := strconv.ParseUint(s[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Invalid name %s: %s", name, err.Error())
	}
	volDriver, err := volume.Get(d.name)
	if err != nil {
		return nil, err
	}
	volumes, err := volDriver.Inspect([]types.VolumeID{types.VolumeID(id)})
	if err != nil || len(volumes) == 0 {
		return nil, err
	}
	return &volumeInfo{metadata: s[1], vol: &volumes[0]}, nil
}

func (d *driver) decode(method string, w http.ResponseWriter, r *http.Request) (*volumeRequest, error) {
	var request volumeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		e := fmt.Errorf("Unable to decode JSON payload")
		d.sendError(method, "", w, e.Error()+":"+err.Error(), http.StatusBadRequest)
		return nil, e
	}
	d.logReq(method, request.Name).Debug()
	return &request, nil
}

func (d *driver) handshake(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(&handshakeResp{
		[]string{VolumeDriver},
	})
	if err != nil {
		d.sendError("handshake", "", w, "encode error", http.StatusInternalServerError)
		return
	}
	d.logReq("handshake", "").Info("Handshake completed")
}

func (d *driver) status(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, fmt.Sprintln("pwx plugin", d.version))
}

func (d *driver) create(w http.ResponseWriter, r *http.Request) {
	method := "create"
	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}
	// It is an error if the volume doesn't already exist.
	_, err = d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumeResponse{Err: e.Error()})
		return
	}
	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driver) remove(w http.ResponseWriter, r *http.Request) {
	method := "remove"
	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}
	// It is an error if the volume doesn't exist.
	_, err = d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumeResponse{Err: e.Error()})
		return
	}
	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driver) mount(w http.ResponseWriter, r *http.Request) {
	var response volumePathResponse
	method := "mount"
	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}
	volInfo, err := d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumePathResponse{Err: e.Error()})
		return
	}
	response.Mountpoint = fmt.Sprintf("/mnt/%s", request.Name)
	os.MkdirAll(response.Mountpoint, 0755)

	v, err := volume.Get(d.name)
	if err != nil {
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err.Error()})
		return
	}
	path, err := v.Attach(volInfo.vol.ID)
	if err != nil {
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err.Error()})
		return
	}
	response.Mountpoint = path
	d.logReq(method, request.Name).Debugf("response %v", path)
	json.NewEncoder(w).Encode(&response)
}

func (d *driver) path(w http.ResponseWriter, r *http.Request) {
	method := "path"
	var response volumePathResponse

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}
	volInfo, err := d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumePathResponse{Err: e.Error()})
		return
	}
	response.Mountpoint = volInfo.vol.AttachPath
	d.logReq(method, request.Name).Debugf("response %v", response.Mountpoint)
	json.NewEncoder(w).Encode(&response)
}

func (d *driver) unmount(w http.ResponseWriter, r *http.Request) {
	method := "unmount"
	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}
	volInfo, err := d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumeResponse{Err: e.Error()})
		return
	}
	v, err := volume.Get(d.name)
	if err != nil {
		json.NewEncoder(w).Encode(&volumeResponse{Err: err.Error()})
		return
	}
	err = v.Detach(volInfo.vol.ID)
	if err != nil {
		d.logReq(request.Name, method).Warnf("%s", err.Error())
		json.NewEncoder(w).Encode(&volumeResponse{Err: err.Error()})
		return
	}

	d.emptyResponse(w)
}
