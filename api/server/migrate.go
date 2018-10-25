package server

import (
	"encoding/json"
	"net/http"

	"github.com/libopenstorage/openstorage/api"
	ost_errors "github.com/libopenstorage/openstorage/api/errors"
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

	response, err := d.CloudMigrateStart(startReq)
	if err != nil {
		if _, ok := err.(*ost_errors.ErrExists); ok {
			w.WriteHeader(http.StatusConflict)
			return
		}
		vd.sendError(method, startReq.TargetId, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(response)
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
		vd.sendError(method, cancelReq.TaskId, w, err.Error(), http.StatusInternalServerError)
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
