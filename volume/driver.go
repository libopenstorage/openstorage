package volume

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/libopenstorage/api"
)

const (
	MethodReceiver = "VolumeDriver"
)

type Driver interface {
	Listen(string) error
}

type driver struct {
	version string
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

func NewVolumePlugin() Driver {
	return &driver{version: "0.3"}
}

func (driver *driver) Listen(socket string) error {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFound)

	router.Methods("GET").Path("/status").HandlerFunc(driver.status)
	router.Methods("POST").Path("/Plugin.Activate").HandlerFunc(driver.handshake)

	handleMethod := func(method string, h http.HandlerFunc) {
		router.Methods("POST").Path(fmt.Sprintf("/%s.%s", MethodReceiver, method)).HandlerFunc(h)
	}

	handleMethod("Create", driver.create)
	handleMethod("Remove", driver.remove)
	handleMethod("Mount", driver.mount)
	handleMethod("Path", driver.path)
	handleMethod("Unmount", driver.unmount)

	var (
		listener net.Listener
		err      error
	)

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
	log.Printf("[plugin] Not found: %+v", r)
	http.NotFound(w, r)
}

func sendError(w http.ResponseWriter, msg string, code int) {
	log.Warn("%d %s", code, msg)
	http.Error(w, msg, code)
}

func emptyResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(&volumeResponse{})
}

func volumeFromName(name string) *api.Volume {
	id, err := strconv.ParseUint(name, 10, 64)
	if err != nil {
		return nil
	}
	v := Default()
	vols, err := v.Inspect([]api.VolumeID{api.VolumeID(id)})
	if err != nil || len(vols) == 0 {
		return nil
	}
	return &vols[0]
}

func (driver *driver) handshake(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(&handshakeResp{
		[]string{"VolumeDriver"},
	})
	if err != nil {
		log.Warn("handshake encode:", err)
		sendError(w, "encode error", http.StatusInternalServerError)
		return
	}
	log.Printf("Handshake completed")
}

func (driver *driver) status(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, fmt.Sprintln("libOpenStorage plugin", driver.version))
}

func (driver *driver) create(w http.ResponseWriter, r *http.Request) {
	var request volumeRequest
	var vcReq api.VolumeCreateRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Debugf("[CREATE VOL] %+v", &request)

	vol := volumeFromName(request.Name)
	if vol != nil {
		json.NewEncoder(w).Encode(&volumeResponse{Err: nil})
		return
	}

	// XXX Volume Spec needs to be passed in via the request.  Currently, it is hardcoded
	// to 10Gb and ext4.
	vcReq.Locator.Name = request.Name
	vcReq.Spec = &api.VolumeSpec{Size: 10 * 1000 * 1000, Format: api.FsExt4, HALevel: 1}
	v := Default()
	ID, err := v.Create(vcReq.Locator, vcReq.Options, vcReq.Spec)
	if err == nil {
		err = v.Format(ID)
	}

	json.NewEncoder(w).Encode(&volumeResponse{Err: err})
}

func (driver *driver) remove(w http.ResponseWriter, r *http.Request) {
	var request volumeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Debugf("[REMOVE VOL] %+v", &request)

	emptyResponse(w)
}

func (driver *driver) mount(w http.ResponseWriter, r *http.Request) {
	var request volumeRequest
	var response volumePathResponse
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Debugf("[MOUNT VOL] %+v", &request)

	vol := volumeFromName(request.Name)
	if vol == nil {
		sendError(w, "locate volume error", http.StatusInternalServerError)
		return
	}
	response.Mountpoint = fmt.Sprintf("/mnt/%s", request.Name)
	os.MkdirAll(response.Mountpoint, 0755)

	v := Default()
	path, err := v.Attach(vol.ID, response.Mountpoint)
	if err != nil {
		response.Err = err
	}
	response.Mountpoint = path

	log.Debugf("Mount response %+v", response)
	json.NewEncoder(w).Encode(&response)
}

func (driver *driver) path(w http.ResponseWriter, r *http.Request) {
	var request volumeRequest
	var response volumePathResponse
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Debugf("[PATH VOL] %+v", &request)

	vol := volumeFromName(request.Name)
	if vol == nil {
		sendError(w, "locate volume error", http.StatusInternalServerError)
		return
	}
	response.Mountpoint = vol.AttachPath

	json.NewEncoder(w).Encode(&response)
}

func (driver *driver) unmount(w http.ResponseWriter, r *http.Request) {
	var request volumeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Debugf("[UNMOUNT VOL] %+v", &request)

	vol := volumeFromName(request.Name)
	if vol == nil {
		sendError(w, "locate volume error", http.StatusInternalServerError)
		return
	}
	v := Default()
	err = v.Detach(vol.ID)
	if err != nil {
		sendError(w, "Unable to unmount: "+err.Error(), http.StatusBadRequest)
		return
	}

	emptyResponse(w)
}
