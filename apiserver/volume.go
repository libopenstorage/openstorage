package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/libopenstorage/openstorage/api"
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

func newVolumeDriver(name string) restServer {
	return &volDriver{restBase{version: apiVersion, name: name}}
}

func (vd *volDriver) String() string {
	return vd.name
}

func (vd *volDriver) parseVolumeID(r *http.Request) (api.VolumeID, error) {
	vars := mux.Vars(r)
	if id, ok := vars["id"]; ok {
		return api.VolumeID(id), nil
	}
	return api.BadVolumeID, fmt.Errorf("could not parse snap ID")
}

func (vd *volDriver) parseSnapID(r *http.Request) (api.SnapID, error) {
	vars := mux.Vars(r)
	if id, ok := vars["id"]; ok {
		return api.SnapID(id), nil
	}
	return api.BadSnapID, fmt.Errorf("could not parse snap ID")
}

func (vd *volDriver) create(w http.ResponseWriter, r *http.Request) {
	var dcRes api.VolumeCreateResponse
	var dcReq api.VolumeCreateRequest
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
	dcRes.VolumeResponse = api.VolumeResponse{Error: responseStatus(err)}
	dcRes.ID = ID
	json.NewEncoder(w).Encode(&dcRes)
}

func (vd *volDriver) volumeState(w http.ResponseWriter, r *http.Request) {
	var (
		volumeID api.VolumeID
		err      error
		req      api.VolumeStateAction
		resp     api.VolumeStateResponse
	)
	method := "volumeState"

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	if volumeID, err = vd.parseVolumeID(r); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}
	for {
		if req.Format != api.ParamIgnore {
			if req.Format == api.ParamOff {
				err = fmt.Errorf("Invalid request to un-format")
				break
			}
			err = d.Format(volumeID)
			if err != nil {
				break
			}
			resp.Format = api.ParamOn
		}
		if req.Attach != api.ParamIgnore {
			if req.Attach == api.ParamOn {
				resp.DevicePath, err = d.Attach(volumeID)
			} else {
				err = d.Detach(volumeID)
			}
			if err != nil {
				break
			}
			resp.Attach = req.Attach
		}
		if req.Mount != api.ParamIgnore {
			if req.MountPath == "" {
				err = fmt.Errorf("Invalid mount path")
				break
			}
			if req.Mount == api.ParamOn {
				err = d.Mount(volumeID, req.MountPath)
			} else {
				err = d.Unmount(volumeID, req.MountPath)
			}
			if err != nil {
				break
			}
			resp.Mount = req.Mount
			resp.MountPath = req.MountPath
		}
		break
	}

	if err != nil {
		resp.Error = err.Error()
	}
	json.NewEncoder(w).Encode(resp)
}

func (vd *volDriver) inspect(w http.ResponseWriter, r *http.Request) {
	var err error
	var volumeID api.VolumeID

	method := "inspect"
	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}
	if volumeID, err = vd.parseVolumeID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}
	dk, err := d.Inspect([]api.VolumeID{volumeID})
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(dk)
}

func (vd *volDriver) delete(w http.ResponseWriter, r *http.Request) {
	var volumeID api.VolumeID
	var err error

	method := "delete"
	if volumeID, err = vd.parseVolumeID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}

	err = d.Delete(volumeID)
	res := api.ResponseStatusNew(err)
	json.NewEncoder(w).Encode(res)
}

func (vd *volDriver) enumerate(w http.ResponseWriter, r *http.Request) {
	var locator api.VolumeLocator
	var configLabels api.Labels
	var err error
	var vols []api.Volume

	method := "enumerate"

	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}
	params := r.URL.Query()
	v := params[string(api.OptName)]
	if v != nil {
		locator.Name = v[0]
	}
	v = params[string(api.OptLabel)]
	if v != nil {
		if err = json.Unmarshal([]byte(v[0]), &locator.VolumeLabels); err != nil {
			e := fmt.Errorf("Failed to parse parse VolumeLabels: %s", err.Error())
			vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		}
	}
	v = params[string(api.OptConfigLabel)]
	if v != nil {
		if err = json.Unmarshal([]byte(v[0]), &configLabels); err != nil {
			e := fmt.Errorf("Failed to parse parse configLabels: %s", err.Error())
			vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		}
	}
	v = params[string(api.OptVolumeID)]
	if v != nil {
		ids := make([]api.VolumeID, len(v))
		for i, s := range v {
			ids[i] = api.VolumeID(s)
		}
		vols, err = d.Inspect(ids)
		if err != nil {
			e := fmt.Errorf("Failed to inspect volumeID: %s", err.Error())
			vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
			return
		}
	} else {
		vols, _ = d.Enumerate(locator, configLabels)
	}
	json.NewEncoder(w).Encode(vols)
}

func (vd *volDriver) snap(w http.ResponseWriter, r *http.Request) {
	var snapReq api.SnapCreateRequest
	var snapRes api.SnapCreateResponse
	method := "snap"

	if err := json.NewDecoder(r.Body).Decode(&snapReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}
	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}
	ID, err := d.Snapshot(snapReq.ID, snapReq.Labels)
	snapRes.VolumeResponse = api.VolumeResponse{Error: responseStatus(err)}
	snapRes.ID = ID
	json.NewEncoder(w).Encode(&snapRes)
}

func (vd *volDriver) snapDelete(w http.ResponseWriter, r *http.Request) {
	var err error
	var snapID api.SnapID

	method := "snapDelete"
	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}
	if snapID, err = vd.parseSnapID(r); err != nil {
		e := fmt.Errorf("Failed to parse SnapID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}
	err = d.SnapDelete(snapID)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(api.VolumeResponse{Error: responseStatus(err)})
}

func (vd *volDriver) snapInspect(w http.ResponseWriter, r *http.Request) {
	var err error
	var snapID api.SnapID

	method := "snapInspect"
	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}
	if snapID, err = vd.parseSnapID(r); err != nil {
		e := fmt.Errorf("Failed to parse SnapID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}
	dk, err := d.SnapInspect([]api.SnapID{snapID})
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(dk)
}

func (vd *volDriver) snapEnumerate(w http.ResponseWriter, r *http.Request) {
	var err error
	var labels api.Labels
	var ids []api.VolumeID
	var snaps []api.VolumeSnap

	method := "snapEnumerate"
	d, err := volume.Get(vd.name)
	if err != nil {
		vd.notFound(w, r)
		return
	}
	params := r.URL.Query()
	v := params[string(api.OptLabel)]
	if v != nil {
		if err = json.Unmarshal([]byte(v[0]), &labels); err != nil {
			e := fmt.Errorf("Failed to parse parse VolumeLabels: %s", err.Error())
			vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		}
	}

	v = params[string(api.OptVolumeID)]
	if v != nil {
		ids = make([]api.VolumeID, len(params))
		for i, s := range v {
			ids[i] = api.VolumeID(s)
		}
	}

	v = params[string(api.OptSnapID)]
	if v != nil {
		sids := make([]api.SnapID, len(params))
		for i, s := range v {
			sids[i] = api.SnapID(s)
		}
		snaps, err = d.SnapInspect(sids)
		if err != nil {
			e := fmt.Errorf("Failed to inspect snaps: %s", err.Error())
			vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
			return
		}
	} else {
		snaps, err = d.SnapEnumerate(ids, labels)
		if err != nil {
			e := fmt.Errorf("Failed to enumerate snaps: %s", err.Error())
			vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
			return
		}
	}

	json.NewEncoder(w).Encode(snaps)
}

func (vd *volDriver) stats(w http.ResponseWriter, r *http.Request) {
}

func (vd *volDriver) alerts(w http.ResponseWriter, r *http.Request) {
}

func version(route string) string {
	return "/" + apiVersion + "/" + route
}

func volPath(route string) string {
	return version("volumes" + route)
}

func snapPath(route string) string {
	return version("snapshot" + route)
}

func (vd *volDriver) Routes() []*Route {
	return []*Route{
		&Route{verb: "POST", path: volPath(""), fn: vd.create},
		&Route{verb: "PUT", path: volPath("/{id}"), fn: vd.volumeState},
		&Route{verb: "GET", path: volPath(""), fn: vd.enumerate},
		&Route{verb: "GET", path: volPath("/{id}"), fn: vd.inspect},
		&Route{verb: "DELETE", path: volPath("/{id}"), fn: vd.delete},
		&Route{verb: "GET", path: volPath("/stats"), fn: vd.stats},
		&Route{verb: "GET", path: volPath("/stats/{id}"), fn: vd.stats},
		&Route{verb: "GET", path: volPath("/alerts"), fn: vd.alerts},
		&Route{verb: "GET", path: volPath("/alerts/{id}"), fn: vd.alerts},
		&Route{verb: "POST", path: snapPath(""), fn: vd.snap},
		&Route{verb: "GET", path: snapPath(""), fn: vd.snapEnumerate},
		&Route{verb: "GET", path: snapPath("/{id}"), fn: vd.snapInspect},
		&Route{verb: "DELETE", path: snapPath("/{id}"), fn: vd.snapDelete},
	}
}
