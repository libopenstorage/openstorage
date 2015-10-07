package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	types "github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/config"
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

func (d *driver) volNotFound(request string, id string, e error, w http.ResponseWriter) error {
	err := fmt.Errorf("Failed to locate volume: " + e.Error())
	d.logReq(request, id).Warn(http.StatusNotFound, " ", err.Error())
	return err
}

func (d *driver) volNotMounted(request string, id string) error {
	err := fmt.Errorf("volume not mounted")
	d.logReq(request, id).Debug(http.StatusNotFound, " ", err.Error())
	return err
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
	d.logReq(method, request.Name).Debug("")
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
	d.logReq("handshake", "").Debug("Handshake completed")
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

	d.logReq(method, request.Name).Debug("")

	volInfo, err := d.volFromName(request.Name)
	if err != nil {
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
		return
	}

	// If this is a block driver, first attach the volume.
	if v.Type()&volume.Block != 0 {
		attachPath, err := v.Attach(volInfo.vol.ID)
		if err != nil {
			d.logReq(method, request.Name).Warnf("Cannot attach volume: %v", err.Error())
			json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
			return
		}
		d.logReq(method, request.Name).Debugf("response %v", attachPath)
	}

	// Now mount it.
	response.Mountpoint = path.Join(config.MountBase, request.Name)
	os.MkdirAll(response.Mountpoint, 0755)

	err = v.Mount(volInfo.vol.ID, response.Mountpoint)
	if err != nil {
		d.logReq(method, request.Name).Warnf("Cannot mount volume %v, %v",
			response.Mountpoint, err)
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
		return
	}
	response.Mountpoint = path.Join(response.Mountpoint, config.DataDir)
	os.MkdirAll(response.Mountpoint, 0755)

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

	d.logReq(method, request.Name).Debug("")
	response.Mountpoint = volInfo.vol.AttachPath
	if response.Mountpoint == "" {
		e := d.volNotMounted(method, request.Name)
		json.NewEncoder(w).Encode(&volumePathResponse{Err: e})
		return
	}
	response.Mountpoint = path.Join(response.Mountpoint, config.DataDir)
	d.logReq(method, request.Name).Debugf("response %v", response.Mountpoint)
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

	mountpoint := path.Join(config.MountBase, request.Name)
	err = v.Unmount(volInfo.vol.ID, mountpoint)
	if err != nil {
		d.logReq(method, request.Name).Warnf("Cannot unmount volume %v, %v",
			mountpoint, err)
		json.NewEncoder(w).Encode(&volumeResponse{Err: err})
		return
	}

	if v.Type()&volume.Block != 0 {
		_ = v.Detach(volInfo.vol.ID)
	}
	d.emptyResponse(w)
}
