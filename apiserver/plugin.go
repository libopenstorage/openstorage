package apiserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
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
		log.Warn("Cannot locate volume driver for %s", d.name)
		return nil, err
	}
	volumes, err := v.Inspect([]types.VolumeID{types.VolumeID(name)})
	if err != nil || len(volumes) == 0 {
		log.Warn("Cannot locate volume %s", name)
		return nil, err
	}
	return &volumeInfo{vol: &volumes[0]}, nil
}

func (d *driver) decode(method string, w http.ResponseWriter, r *http.Request) (*volumeRequest, error) {
	var request volumeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Warn("Cannot decode request.", err)
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
		log.Warnf("Cannot locate volume driver for %+v", d.name)
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
		return
	}

	request, err := d.decode(method, w, r)
	if err != nil {
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
		return
	}

	volInfo, err := d.volFromName(request.Name)
	if err != nil {
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
		return
	}

	// If this is a block driver, first attach the volume.
	if v.Type() == volume.Block {
		_, err = v.Attach(volInfo.vol.ID)
		if err != nil {
			log.Warnf("Cannot attach volume %+v, %+v", volInfo.vol.ID, err)
			json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
			return
		}
		log.Infof("Volume %+v attached", volInfo.vol.ID)
	}

	// Now mount it.
	response.Mountpoint = fmt.Sprintf("/mnt/%s", request.Name)
	os.MkdirAll(response.Mountpoint, 0755)

	err = v.Mount(volInfo.vol.ID, response.Mountpoint)
	if err != nil {
		log.Warnf("Cannot mount volume %+v at %+v, %+v", volInfo.vol.ID, response.Mountpoint, err)
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
		return
	}

	log.Infof("Volume %+v mounted at %+v", volInfo, response.Mountpoint)

	d.logReq(method, request.Name).Debugf("response %v", response.Mountpoint)
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

	response.Mountpoint = volInfo.vol.AttachPath
	d.logReq(method, request.Name).Debugf("response %v", response.Mountpoint)
	json.NewEncoder(w).Encode(&response)
}

func (d *driver) unmount(w http.ResponseWriter, r *http.Request) {
	method := "unmount"

	v, err := volume.Get(d.name)
	if err != nil {
		log.Warn("Cannot locate volume driver for %s", d.name)
		json.NewEncoder(w).Encode(&volumeResponse{Err: err})
		return
	}

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	volInfo, err := d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumeResponse{Err: e})
		return
	}

	if v.Type() == volume.Block {
		err = v.Detach(volInfo.vol.ID)
		if err != nil {
			log.Warnf("Cannot detach volume %+v, %+v", volInfo.vol.ID, err)
			d.logReq(request.Name, method).Warnf("%s", err.Error())
			json.NewEncoder(w).Encode(&volumeResponse{Err: err})
			return
		}
	}

	// XXX TODO unmount
	// log.Infof("Volume %+v mounted at %+v", volInfo, response.Mountpoint)

	d.emptyResponse(w)
}
