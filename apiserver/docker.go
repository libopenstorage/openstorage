package apiserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	types "github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	// VolumeDriver is the string returned in the handshake protocol.
	VolumeDriver = "VolumeDriver"
)

// Implementation of the Docker volumes plugin specification.
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
	Err error
}
type volumePathResponse struct {
	Mountpoint string
	Err        error
}

type volumeInfo struct {
	vol *types.Volume
}

func newVolumePlugin(name string) restServer {
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
	v, err := volume.Get(d.name)
	if err != nil {
		return nil, fmt.Errorf("Cannot locate volume driver for %s: %s", d.name, err.Error())
	}
	volumes, err := v.Inspect([]types.VolumeID{types.VolumeID(name)})
	if err != nil || len(volumes) == 0 {
		return nil, fmt.Errorf("Cannot locate volume %s", name)
	}
	return &volumeInfo{vol: &volumes[0]}, nil
}

func (d *driver) decode(method string, w http.ResponseWriter, r *http.Request) (*volumeRequest, error) {
	var request volumeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		e := fmt.Errorf("Unable to decode JSON payload")
		d.sendError(method, "", w, e.Error()+":"+err.Error(), http.StatusBadRequest)
		return nil, e
	}
	d.logReq(method, request.Name).Info()
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
	io.WriteString(w, fmt.Sprintln("osd plugin", d.version))
}

func (d *driver) create(w http.ResponseWriter, r *http.Request) {
	method := "create"

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	d.logReq(method, request.Name).Info("")

	// It is an error if the volume doesn't already exist.
	_, err = d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumeResponse{Err: e})
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

	d.logReq(method, request.Name).Info("")

	// It is an error if the volume doesn't exist.
	_, err = d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumeResponse{Err: e})
		return
	}

	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driver) mount(w http.ResponseWriter, r *http.Request) {
	var response volumePathResponse
	method := "mount"

	v, err := volume.Get(d.name)
	if err != nil {
		d.logReq(method, "").Warn("Cannot locate volume driver")
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
		return
	}

	request, err := d.decode(method, w, r)
	if err != nil {
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
		return
	}

	d.logReq(method, request.Name).Info("")

	volInfo, err := d.volFromName(request.Name)
	if err != nil {
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
		return
	}

	// If this is a block driver, first attach the volume.
	if v.Type() == volume.Block {
		attachPath, err := v.Attach(volInfo.vol.ID)
		if err != nil {
			d.logReq(method, request.Name).Warnf("Cannot attach volume: %v", err.Error())
			json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
			return
		}
		d.logReq(method, request.Name).Infof("response %v", attachPath)
	}

	// Now mount it.
	response.Mountpoint = fmt.Sprintf("/mnt/%s", request.Name)
	os.MkdirAll(response.Mountpoint, 0755)

	err = v.Mount(volInfo.vol.ID, response.Mountpoint)
	if err != nil {
		d.logReq(method, request.Name).Warnf("Cannot mount volume %v, %v",
			response.Mountpoint, err)
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
		return
	}

	d.logReq(method, request.Name).Infof("response %v", response.Mountpoint)
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
		json.NewEncoder(w).Encode(&volumePathResponse{Err: e})
		return
	}

	d.logReq(method, request.Name).Info("")

	response.Mountpoint = volInfo.vol.AttachPath
	d.logReq(method, request.Name).Infof("response %v", response.Mountpoint)
	json.NewEncoder(w).Encode(&response)
}

func (d *driver) unmount(w http.ResponseWriter, r *http.Request) {
	method := "unmount"

	v, err := volume.Get(d.name)
	if err != nil {
		d.logReq(method, "").Warnf("Cannot locate volume driver: %v", err.Error())
		json.NewEncoder(w).Encode(&volumeResponse{Err: err})
		return
	}

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	d.logReq(method, request.Name).Info("")

	volInfo, err := d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumeResponse{Err: e})
		return
	}

	if v.Type() == volume.Block {
		err = v.Detach(volInfo.vol.ID)
		if err != nil {
			d.logReq(method, request.Name).Warnf("Cannot detach volume : %s", err.Error())
			json.NewEncoder(w).Encode(&volumeResponse{Err: err})
			return
		}
	}

	// XXX TODO unmount
	// log.Infof("Volume %+v mounted at %+v", volInfo, response.Mountpoint)

	d.emptyResponse(w)
}
