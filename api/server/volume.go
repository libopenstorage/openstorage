package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	volApiVersion = "v1"
)

type volApi struct {
	restBase
}

func responseStatus(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func newVolumeAPI(name string) restServer {
	return &volApi{restBase{version: volApiVersion, name: name}}
}

func (vd *volApi) String() string {
	return vd.name
}

func (vd *volApi) parseVolumeID(r *http.Request) (api.VolumeID, error) {
	vars := mux.Vars(r)
	if id, ok := vars["id"]; ok {
		return api.VolumeID(id), nil
	}
	return api.BadVolumeID, fmt.Errorf("could not parse snap ID")
}

func (vd *volApi) create(w http.ResponseWriter, r *http.Request) {
	var dcRes api.VolumeCreateResponse
	var dcReq api.VolumeCreateRequest
	method := "create"

	if err := json.NewDecoder(r.Body).Decode(&dcReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}
	d, err := volume.Get(vd.name)
	if err != nil {
		notFound(w, r)
		return
	}
	ID, err := d.Create(dcReq.Locator, dcReq.Source, dcReq.Spec)
	dcRes.VolumeResponse = api.VolumeResponse{Error: responseStatus(err)}
	dcRes.ID = ID
	json.NewEncoder(w).Encode(&dcRes)
}

func (vd *volApi) volumeState(w http.ResponseWriter, r *http.Request) {
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
		notFound(w, r)
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
			if req.Mount == api.ParamOn {
				if req.MountPath == "" {
					err = fmt.Errorf("Invalid mount path")
					break
				}
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

func (vd *volApi) inspect(w http.ResponseWriter, r *http.Request) {
	var err error
	var volumeID api.VolumeID

	method := "inspect"
	d, err := volume.Get(vd.name)
	if err != nil {
		notFound(w, r)
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

func (vd *volApi) delete(w http.ResponseWriter, r *http.Request) {
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
		notFound(w, r)
		return
	}

	err = d.Delete(volumeID)
	res := api.ResponseStatusNew(err)
	json.NewEncoder(w).Encode(res)
}

func (vd *volApi) enumerate(w http.ResponseWriter, r *http.Request) {
	var locator api.VolumeLocator
	var configLabels api.Labels
	var err error
	var vols []api.Volume

	method := "enumerate"

	d, err := volume.Get(vd.name)
	if err != nil {
		notFound(w, r)
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

func (vd *volApi) snap(w http.ResponseWriter, r *http.Request) {
	var snapReq api.SnapCreateRequest
	var snapRes api.SnapCreateResponse
	method := "snap"

	if err := json.NewDecoder(r.Body).Decode(&snapReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}
	d, err := volume.Get(vd.name)
	if err != nil {
		notFound(w, r)
		return
	}
	ID, err := d.Snapshot(snapReq.ID, snapReq.Readonly, snapReq.Locator)
	snapRes.VolumeCreateResponse.VolumeResponse = api.VolumeResponse{Error: responseStatus(err)}
	snapRes.VolumeCreateResponse.ID = ID
	json.NewEncoder(w).Encode(&snapRes)
}

func (vd *volApi) snapEnumerate(w http.ResponseWriter, r *http.Request) {
	var err error
	var labels api.Labels
	var ids []api.VolumeID

	method := "snapEnumerate"
	d, err := volume.Get(vd.name)
	if err != nil {
		notFound(w, r)
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

	v, ok := params[string(api.OptVolumeID)]
	if v != nil && ok {
		ids = make([]api.VolumeID, len(params))
		for i, s := range v {
			ids[i] = api.VolumeID(s)
		}
	}

	snaps, err := d.SnapEnumerate(ids, labels)
	if err != nil {
		e := fmt.Errorf("Failed to enumerate snaps: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(snaps)
}

func (vd *volApi) stats(w http.ResponseWriter, r *http.Request) {
	var volumeID api.VolumeID
	var err error

	method := "stats"
	if volumeID, err = vd.parseVolumeID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	d, err := volume.Get(vd.name)
	if err != nil {
		notFound(w, r)
		return
	}

	stats, err := d.Stats(volumeID)
	if err != nil {
		e := fmt.Errorf("Failed to get stats: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(stats)
}

func (vd *volApi) alerts(w http.ResponseWriter, r *http.Request) {
	var volumeID api.VolumeID
	var err error

	method := "alerts"
	if volumeID, err = vd.parseVolumeID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	d, err := volume.Get(vd.name)
	if err != nil {
		notFound(w, r)
		return
	}

	alerts, err := d.Alerts(volumeID)
	if err != nil {
		e := fmt.Errorf("Failed to get alerts: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(alerts)
}

func volVersion(route string) string {
	return "/" + volApiVersion + "/" + route
}

func volPath(route string) string {
	return volVersion("volumes" + route)
}

func snapPath(route string) string {
	return volVersion("snapshot" + route)
}

func (vd *volApi) Routes() []*Route {
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
	}
}
