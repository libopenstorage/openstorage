package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	log "github.com/Sirupsen/logrus"

	types "github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	VolumeDriver = "VolumeDriver"
)

type Driver interface {
	Listen(path string) error
}

type driver struct {
	version string
	name    string
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
	metadata string
	vol      *types.Volume
}

func NewVolumePlugin(name string) Driver {
	return &driver{version: "0.3"}
}

func logReq(request string, id string) *log.Entry {
	return log.WithFields(log.Fields{
		"Driver":  VolumeDriver,
		"Request": request,
		"ID":      id,
	})
}

func (d *driver) Listen(socketPath string) error {
	var (
		listener net.Listener
		err      error
	)
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFound)

	router.Methods("GET").Path("/status").HandlerFunc(d.status)
	router.Methods("POST").Path("/Plugin.Activate").HandlerFunc(d.handshake)

	handleMethod := func(method string, h http.HandlerFunc) {
		router.Methods("POST").Path(fmt.Sprintf("/%s.%s", VolumeDriver, method)).HandlerFunc(h)
	}

	handleMethod("Create", d.create)
	handleMethod("Remove", d.remove)
	handleMethod("Mount", d.mount)
	handleMethod("Path", d.path)
	handleMethod("Unmount", d.unmount)

	socket := path.Join(socketPath, d.name)

	os.Remove(socket)
	os.MkdirAll(path.Dir(socket), 0755)

	log.Printf("Plugin listening on %+v", socket)
	listener, err = net.Listen("unix", socket)
	if err != nil {
		return err
	}

	return http.Serve(listener, router)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	log.Warnf("[%s] Not found: %+v", VolumeDriver, r)
	http.NotFound(w, r)
}

func sendError(request string, id string, w http.ResponseWriter, msg string, code int) {
	logReq(request, id).Warn("%d %s", code, msg)
	http.Error(w, msg, code)
}

func volNotFound(request string, id string, e error, w http.ResponseWriter) error {
	err := fmt.Errorf("Failed to locate volume:" + e.Error())
	logReq(request, id).Warn("%d %s", http.StatusNotFound, err.Error())
	return err
}

func emptyResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driver) volFromName(name string) (*volumeInfo, error) {
	s := strings.Split(name, ":")
	if len(s) != 2 {
		return nil, fmt.Errorf("Invalid logReq for name %s", name)
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
		sendError(method, "", w, e.Error()+":"+err.Error(), http.StatusBadRequest)
		return nil, e
	}
	logReq(method, request.Name).Debug()
	return &request, nil
}

func (d *driver) handshake(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(&handshakeResp{
		[]string{VolumeDriver},
	})
	if err != nil {
		sendError("handshake", "", w, "encode error", http.StatusInternalServerError)
		return
	}
	logReq("handshake", "").Info("Handshake completed")
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
		e := volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumeResponse{Err: e})
		return
	}
	json.NewEncoder(w).Encode(&volumeResponse{Err: nil})
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
		e := volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumeResponse{Err: e})
		return
	}
	json.NewEncoder(w).Encode(&volumeResponse{Err: nil})
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
		e := volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumePathResponse{Err: e})
		return
	}
	response.Mountpoint = fmt.Sprintf("/mnt/%s", request.Name)
	os.MkdirAll(response.Mountpoint, 0755)

	v, err := volume.Get(d.name)
	if err != nil {
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
		return
	}
	path, err := v.Attach(volInfo.vol.ID, response.Mountpoint)
	if err != nil {
		json.NewEncoder(w).Encode(&volumePathResponse{Err: err})
		return
	}
	response.Mountpoint = path
	logReq(method, request.Name).Debugf("response %v", path)
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
		e := volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumePathResponse{Err: e})
		return
	}
	response.Mountpoint = volInfo.vol.AttachPath
	logReq(method, request.Name).Debugf("response %v", response.Mountpoint)
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
		e := volNotFound(method, request.Name, err, w)
		json.NewEncoder(w).Encode(&volumeResponse{Err: e})
		return
	}
	v, err := volume.Get(d.name)
	if err != nil {
		json.NewEncoder(w).Encode(&volumeResponse{Err: err})
		return
	}
	err = v.Detach(volInfo.vol.ID)
	if err != nil {
		logReq(request.Name, method).Warnf("%s", err.Error())
		json.NewEncoder(w).Encode(&volumeResponse{Err: err})
		return
	}

	emptyResponse(w)
}
