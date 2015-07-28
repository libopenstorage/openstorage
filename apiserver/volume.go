package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	types "github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	apiVersion = "v1"
)

type volDriver struct {
	restBase
}

func responseStatus(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func NewVolumeDriver(name string) restServer {
	return &volDriver{restBase{version: apiVersion, name: name}}
}

func (d *volDriver) String() string {
	return d.name
}

func (vd *volDriver) parseVolumeID(r *http.Request) (types.VolumeID, error) {
	vars := mux.Vars(r)
	if id, ok := vars["id"]; ok {
		return types.VolumeID(id), nil
	}
	return types.VolumeID(""), fmt.Errorf("could not vd.parse volume ID")
}

func (vd *volDriver) create(w http.ResponseWriter, r *http.Request) {
	var dcRes types.VolumeCreateResponse
	var dcReq types.VolumeCreateRequest
	method := "create"

	if err := json.NewDecoder(r.Body).Decode(&dcReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}
	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}
	ID, err := d.Create(dcReq.Locator, dcReq.Options, dcReq.Spec)
	dcRes.Status = responseStatus(err)
	dcRes.ID = ID
	json.NewEncoder(w).Encode(&dcRes)
}

func (vd *volDriver) attach(w http.ResponseWriter, r *http.Request) {

	var attachReq types.VolumeAttachRequest
	method := "attach"

	err := json.NewDecoder(r.Body).Decode(&attachReq)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}

	path, err := d.Attach(attachReq.ID)
	res := types.VolumeAttachResponse{
		Status: responseStatus(err),
		Path:   path,
	}
	json.NewEncoder(w).Encode(res)
}

func (vd *volDriver) detach(w http.ResponseWriter, r *http.Request) {

	var volumeID types.VolumeID
	var err error

	method := "detach"
	if volumeID, err = vd.parseVolumeID(r); err != nil {
		log.Printf("Server: couldn't vd.parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Server: detaching %v\n", volumeID)
	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}
	err = d.Detach(volumeID)
	res := types.ResponseStatusNew(err)
	json.NewEncoder(w).Encode(res)
}

func (vd *volDriver) mount(w http.ResponseWriter, r *http.Request) {
}

func (vd *volDriver) unmount(w http.ResponseWriter, r *http.Request) {
}

func (vd *volDriver) inspect(w http.ResponseWriter, r *http.Request) {
	var ids []types.VolumeID
	var err error

	method := "inspect"
	params := r.URL.Query()
	v := params[string(types.OptID)]
	if v != nil {
		if err = json.Unmarshal([]byte(v[0]), &ids); err != nil {
			log.Printf("Server: couldn't vd.parse VolumeIDs: %s", err.Error())
			vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		}
	}

	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}
	dk, err := d.Inspect(ids)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(dk)
}

func (vd *volDriver) delete(w http.ResponseWriter, r *http.Request) {
	var volumeID types.VolumeID
	var err error

	method := "delete"
	if volumeID, err = vd.parseVolumeID(r); err != nil {
		log.Printf("Server: couldn't vd.parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Server: deleting %v\n", volumeID)
	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}

	err = d.Delete(volumeID)
	res := types.ResponseStatusNew(err)
	json.NewEncoder(w).Encode(res)
}

func (vd *volDriver) format(w http.ResponseWriter, r *http.Request) {
	var volumeID types.VolumeID
	var err error

	method := "format"
	if volumeID, err = vd.parseVolumeID(r); err != nil {
		log.Printf("Server: couldn't vd.parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Server: formatting %v\n", volumeID)
	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}

	err = d.Format(volumeID)
	res := types.ResponseStatusNew(err)
	json.NewEncoder(w).Encode(res)
}

func (vd *volDriver) enumerate(w http.ResponseWriter, r *http.Request) {
	var locator types.VolumeLocator
	var configLabels types.Labels
	var err error

	method := "enumerate"
	params := r.URL.Query()
	v := params[string(types.OptName)]
	if v != nil {
		locator.Name = v[0]
	}
	v = params[string(types.OptVolumeLabel)]
	if v != nil {
		if err = json.Unmarshal([]byte(v[0]), &locator.VolumeLabels); err != nil {
			log.Printf("Server: couldn't vd.parse VolumeLabels: %s", err.Error())
			vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		}
	}
	v = params[string(types.OptConfigLabel)]
	if v != nil {
		if err = json.Unmarshal([]byte(v[0]), &configLabels); err != nil {
			log.Printf("Server: couldn't vd.parse configLabels: %s", err.Error())
			vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		}
	}

	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}
	res, _ := d.Enumerate(locator, configLabels)
	json.NewEncoder(w).Encode(res)
}

func (vd *volDriver) snap(w http.ResponseWriter, r *http.Request) {
}

func (vd *volDriver) snapDelete(w http.ResponseWriter, r *http.Request) {
}

func (vd *volDriver) snapInspect(w http.ResponseWriter, r *http.Request) {
}

func (vd *volDriver) snapEnumerate(w http.ResponseWriter, r *http.Request) {
}

func (vd *volDriver) stats(w http.ResponseWriter, r *http.Request) {
}

func (vd *volDriver) alerts(w http.ResponseWriter, r *http.Request) {
}

func version(route string) string {
	return "/" + apiVersion + "/" + route
}

func volPath(route string) string {
	return version("volumes/" + route)
}

func snapPath(route string) string {
	return version("snap/" + route)
}

func (vd *volDriver) Routes() []*Route {
	return []*Route{
		&Route{verb: "POST", path: volPath(""), fn: vd.create},
		&Route{verb: "PUT", path: volPath("attach/{id}"), fn: vd.attach},
		&Route{verb: "PUT", path: volPath("detach/{id}"), fn: vd.detach},
		&Route{verb: "PUT", path: volPath("mount/{id}"), fn: vd.mount},
		&Route{verb: "PUT", path: volPath("unmount/{id}"), fn: vd.unmount},
		&Route{verb: "GET", path: volPath(""), fn: vd.enumerate},
		&Route{verb: "GET", path: volPath("{id}"), fn: vd.inspect},
		&Route{verb: "DELETE", path: volPath("{id}"), fn: vd.delete},
		&Route{verb: "GET", path: volPath("stats/"), fn: vd.stats},
		&Route{verb: "GET", path: volPath("stats/{id}"), fn: vd.stats},
		&Route{verb: "GET", path: volPath("alerts"), fn: vd.alerts},
		&Route{verb: "GET", path: volPath("alerts/{id}"), fn: vd.alerts},
		&Route{verb: "GET", path: snapPath("{id}"), fn: vd.snapInspect},
		&Route{verb: "GET", path: snapPath(""), fn: vd.snapEnumerate},
		&Route{verb: "POST", path: snapPath("{id}"), fn: vd.snap},
		&Route{verb: "DELETE", path: snapPath("{id}"), fn: vd.snapDelete},
	}
}
