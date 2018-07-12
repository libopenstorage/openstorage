package server

import (
	"encoding/json"
	"net/http"

	"github.com/libopenstorage/openstorage/api"
)

func (vd *volAPI) cloudMigrateStart(w http.ResponseWriter, r *http.Request) {
	startReq := &api.CloudMigrateStartRequest{}
	method := "cloudMigrateStart"

	if err := json.NewDecoder(r.Body).Decode(startReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	err = d.CloudMigrateStart(startReq)
	if err != nil {
		vd.sendError(method, startReq.TargetId, w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (vd *volAPI) cloudMigrateCancel(w http.ResponseWriter, r *http.Request) {
	cancelReq := &api.CloudMigrateCancelRequest{}
	method := "cloudMigrateCancel"

	if err := json.NewDecoder(r.Body).Decode(cancelReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	err = d.CloudMigrateCancel(cancelReq)
	if err != nil {
		vd.sendError(method, cancelReq.TargetId, w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (vd *volAPI) cloudMigrateStatus(w http.ResponseWriter, r *http.Request) {
	method := "cloudMigrateState"

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	statusResp, err := d.CloudMigrateStatus()
	if err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(statusResp)
}
