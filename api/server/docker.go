package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/libopenstorage/openstorage/api"
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
	Opts map[string]string
}

type volumeResponse struct {
	Err string
}

type volumePathResponse struct {
	Mountpoint string
	volumeResponse
}

type volumeInfo struct {
	Name       string
	Mountpoint string
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
		&Route{verb: "POST", path: volDriverPath("List"), fn: d.list},
		&Route{verb: "POST", path: volDriverPath("Get"), fn: d.get},
		&Route{verb: "POST", path: volDriverPath("Unmount"), fn: d.unmount},
		&Route{verb: "POST", path: "/Plugin.Activate", fn: d.handshake},
		&Route{verb: "GET", path: "/status", fn: d.status},
	}
}

func (d *driver) emptyResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driver) errorResponse(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(&volumeResponse{Err: err.Error()})
}

func (d *driver) volFromName(name string) (*api.Volume, error) {
	v, err := volume.Get(d.name)
	if err != nil {
		return nil, fmt.Errorf("Cannot locate volume driver for %s: %s", d.name, err.Error())
	}
	vols, err := v.Inspect([]api.VolumeID{api.VolumeID(name)})
	if err == nil && len(vols) == 1 {
		return &vols[0], nil
	}
	vols, err = v.Enumerate(api.VolumeLocator{Name: name}, nil)
	if err == nil && len(vols) == 1 {
		return &vols[0], nil
	}
	return nil, fmt.Errorf("Cannot locate volume %s", name)
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

func (d *driver) specFromOpts(Opts map[string]string) *api.VolumeSpec {
	var spec api.VolumeSpec
	for k, v := range Opts {
		switch k {
		case api.SpecEphemeral:
			spec.Ephemeral, _ = strconv.ParseBool(v)
		case api.SpecSize:
			spec.Size, _ = strconv.ParseUint(v, 10, 64)
		case api.SpecFilesystem:
			spec.Format = api.Filesystem(v)
		case api.SpecBlockSize:
			blockSize, _ := strconv.ParseInt(v, 10, 64)
			spec.BlockSize = int(blockSize)
		case api.SpecHaLevel:
			haLevel, _ := strconv.ParseInt(v, 10, 64)
			spec.HALevel = int(haLevel)
		case api.SpecCos:
			cos, _ := strconv.ParseInt(v, 10, 64)
			spec.Cos = api.VolumeCos(cos)
		case api.SpecDedupe:
			spec.Dedupe, _ = strconv.ParseBool(v)
		case api.SpecSnapshotInterval:
			snapshotInterval, _ := strconv.ParseInt(v, 10, 64)
			spec.SnapshotInterval = int(snapshotInterval)
		}
	}
	return &spec
}

func (d *driver) create(w http.ResponseWriter, r *http.Request) {
	var err error
	method := "create"

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}
	d.logReq(method, request.Name).Info("")

	_, err = d.volFromName(request.Name)
	if err != nil {
		v, err := volume.Get(d.name)
		if err != nil {
			d.errorResponse(w, err)
			return
		}
		spec := d.specFromOpts(request.Opts)
		_, err = v.Create(api.VolumeLocator{Name: request.Name}, nil, spec)
		if err != nil {
			d.errorResponse(w, err)
			return
		}
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
		d.errorResponse(w, e)
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
		d.errorResponse(w, err)
		return
	}

	request, err := d.decode(method, w, r)
	if err != nil {
		d.errorResponse(w, err)
		return
	}

	d.logReq(method, request.Name).Debug("")

	vol, err := d.volFromName(request.Name)
	if err != nil {
		d.errorResponse(w, err)
		return
	}

	// If this is a block driver, first attach the volume.
	if v.Type()&api.Block != 0 {
		attachPath, err := v.Attach(vol.ID)
		if err != nil {
			d.logReq(method, request.Name).Warnf("Cannot attach volume: %v", err.Error())
			d.errorResponse(w, err)
			return
		}
		d.logReq(method, request.Name).Debugf("response %v", attachPath)
	}

	// Now mount it.
	response.Mountpoint = path.Join(config.MountBase, request.Name)
	os.MkdirAll(response.Mountpoint, 0755)

	err = v.Mount(vol.ID, response.Mountpoint)
	if err != nil {
		d.logReq(method, request.Name).Warnf("Cannot mount volume %v, %v",
			response.Mountpoint, err)
		d.errorResponse(w, err)
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

	vol, err := d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		d.errorResponse(w, e)
		return
	}

	d.logReq(method, request.Name).Debug("")
	response.Mountpoint = vol.AttachPath
	if response.Mountpoint == "" {
		e := d.volNotMounted(method, request.Name)
		d.errorResponse(w, e)
		return
	}
	response.Mountpoint = path.Join(response.Mountpoint, config.DataDir)
	d.logReq(method, request.Name).Debugf("response %v", response.Mountpoint)
	json.NewEncoder(w).Encode(&response)
}

func (d *driver) list(w http.ResponseWriter, r *http.Request) {
	method := "list"

	v, err := volume.Get(d.name)
	if err != nil {
		d.logReq(method, "").Warnf("Cannot locate volume driver: %v", err.Error())
		d.errorResponse(w, err)
		return
	}

	vols, err := v.Enumerate(api.VolumeLocator{}, nil)
	if err != nil {
		d.errorResponse(w, err)
		return
	}

	volInfo := make([]volumeInfo, len(vols))
	for i, v := range vols {
		volInfo[i].Name = v.Locator.Name
		if v.AttachPath != "" {
			volInfo[i].Mountpoint = path.Join(v.AttachPath, config.DataDir)
		}
	}
	json.NewEncoder(w).Encode(map[string][]volumeInfo{"Volumes": volInfo})
}

func (d *driver) get(w http.ResponseWriter, r *http.Request) {
	method := "get"

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}
	vol, err := d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		d.errorResponse(w, e)
		return
	}

	volInfo := volumeInfo{Name: request.Name}
	if vol.AttachPath != "" {
		volInfo.Mountpoint = path.Join(vol.AttachPath, config.DataDir)
	}

	json.NewEncoder(w).Encode(map[string]volumeInfo{"Volume": volInfo})
}

func (d *driver) unmount(w http.ResponseWriter, r *http.Request) {
	method := "unmount"

	v, err := volume.Get(d.name)
	if err != nil {
		d.logReq(method, "").Warnf("Cannot locate volume driver: %v", err.Error())
		d.errorResponse(w, err)
		return
	}

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	d.logReq(method, request.Name).Info("")

	vol, err := d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		d.errorResponse(w, e)
		return
	}

	mountpoint := path.Join(config.MountBase, request.Name)
	err = v.Unmount(vol.ID, mountpoint)
	if err != nil {
		d.logReq(method, request.Name).Warnf("Cannot unmount volume %v, %v",
			mountpoint, err)
		d.errorResponse(w, err)
		return
	}

	if v.Type()&api.Block != 0 {
		_ = v.Detach(vol.ID)
	}
	d.emptyResponse(w)
}
