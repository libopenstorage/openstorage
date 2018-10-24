package server

import (
	"encoding/json"
	"net/http"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

func (vd *volAPI) cloudBackupCreate(w http.ResponseWriter, r *http.Request) {
	backupReq := &api.CloudBackupCreateRequest{}
	method := "cloudBackupCreate"

	if err := json.NewDecoder(r.Body).Decode(backupReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	createResp, err := d.CloudBackupCreate(backupReq)
	if err != nil {
		if err == volume.ErrInvalidName {
			w.WriteHeader(http.StatusConflict)
			return
		}
		vd.sendError(method, backupReq.VolumeID, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(createResp)
}

func (vd *volAPI) cloudBackupGroupCreate(w http.ResponseWriter, r *http.Request) {
	backupGroupReq := &api.CloudBackupGroupCreateRequest{}
	method := "cloudBackupGroupCreate"

	if err := json.NewDecoder(r.Body).Decode(backupGroupReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	err = d.CloudBackupGroupCreate(backupGroupReq)
	if err != nil {
		vd.sendError(method, backupGroupReq.GroupID, w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (vd *volAPI) cloudBackupRestore(w http.ResponseWriter, r *http.Request) {
	restoreReq := &api.CloudBackupRestoreRequest{}
	method := "cloudBackupRestore"

	if err := json.NewDecoder(r.Body).Decode(restoreReq); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	if restoreReq.NodeID != "" {
		nodeIds, err := vd.nodeIPtoIds([]string{restoreReq.NodeID})
		if err != nil {
			vd.sendError(method, restoreReq.ID, w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(nodeIds) > 0 {
			if nodeIds[0] != restoreReq.NodeID {
				restoreReq.NodeID = nodeIds[0]
			}
		}
	}

	restoreResp, err := d.CloudBackupRestore(restoreReq)
	if err != nil {
		if err == volume.ErrInvalidName {
			w.WriteHeader(http.StatusConflict)
			return
		}
		vd.sendError(method, restoreReq.ID, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(restoreResp)
}

func (vd *volAPI) cloudBackupDelete(w http.ResponseWriter, r *http.Request) {
	deleteReq := &api.CloudBackupDeleteRequest{}
	method := "cloudBackupDelete"

	if err := json.NewDecoder(r.Body).Decode(deleteReq); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	err = d.CloudBackupDelete(deleteReq)
	if err != nil {
		vd.sendError(method, deleteReq.ID, w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (vd *volAPI) cloudBackupDeleteAll(w http.ResponseWriter, r *http.Request) {
	deleteAllReq := &api.CloudBackupDeleteAllRequest{}
	method := "cloudBackupDeleteAll"

	if err := json.NewDecoder(r.Body).Decode(deleteAllReq); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	err = d.CloudBackupDeleteAll(deleteAllReq)
	if err != nil {
		vd.sendError(method, deleteAllReq.SrcVolumeID, w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (vd *volAPI) cloudBackupEnumerate(w http.ResponseWriter, r *http.Request) {
	method := "cloudBackupEnumerate"
	enumerateReq := &api.CloudBackupEnumerateRequest{}

	if err := json.NewDecoder(r.Body).Decode(enumerateReq); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	enumerateResp, err := d.CloudBackupEnumerate(enumerateReq)
	if err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(enumerateResp)
}

func (vd *volAPI) cloudBackupStatus(w http.ResponseWriter, r *http.Request) {
	method := "cloudBackupStatus"
	backupStatus := &api.CloudBackupStatusRequest{}

	if err := json.NewDecoder(r.Body).Decode(backupStatus); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	backupStatusResp, err := d.CloudBackupStatus(backupStatus)
	if err != nil {
		if err == volume.ErrInvalidName {
			w.WriteHeader(http.StatusConflict)
			return
		}
		vd.sendError(method, "", w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(backupStatusResp)
}

func (vd *volAPI) cloudBackupCatalog(w http.ResponseWriter, r *http.Request) {
	method := "cloudBackupCatalog"
	catalogReq := &api.CloudBackupCatalogRequest{}

	if err := json.NewDecoder(r.Body).Decode(catalogReq); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	catalog, err := d.CloudBackupCatalog(catalogReq)
	if err != nil {
		vd.sendError(method, catalogReq.ID, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(catalog)

}
func (vd *volAPI) cloudBackupHistory(w http.ResponseWriter, r *http.Request) {
	method := "cloudBackupHistory"
	historyReq := &api.CloudBackupHistoryRequest{}

	if err := json.NewDecoder(r.Body).Decode(historyReq); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	history, err := d.CloudBackupHistory(historyReq)
	if err != nil {
		vd.sendError(method, historyReq.SrcVolumeID, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(history)
}

func (vd *volAPI) cloudBackupStateChange(w http.ResponseWriter, r *http.Request) {
	method := "cloudBackupStatusChange"
	stateChangeReq := &api.CloudBackupStateChangeRequest{}
	if err := json.NewDecoder(r.Body).Decode(stateChangeReq); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	err = d.CloudBackupStateChange(stateChangeReq)
	if err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (vd *volAPI) cloudBackupSchedCreate(w http.ResponseWriter, r *http.Request) {
	method := "cloudBackupSchedCreate"
	backupSchedReq := &api.CloudBackupSchedCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(backupSchedReq); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	backupSchedResp, err := d.CloudBackupSchedCreate(backupSchedReq)
	if err != nil {
		vd.sendError(method, backupSchedReq.SrcVolumeID, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(backupSchedResp)
}

func (vd *volAPI) cloudBackupGroupSchedCreate(w http.ResponseWriter, r *http.Request) {
	method := "cloudBackupGroupSchedCreate"
	backupGroupSchedReq := &api.CloudBackupGroupSchedCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(backupGroupSchedReq); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	backupSchedResp, err := d.CloudBackupGroupSchedCreate(backupGroupSchedReq)
	if err != nil {
		vd.sendError(method, backupGroupSchedReq.GroupID, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(backupSchedResp)
}

func (vd *volAPI) cloudBackupSchedDelete(w http.ResponseWriter, r *http.Request) {
	method := "cloudBackupSchedDelete"
	deleteReq := &api.CloudBackupSchedDeleteRequest{}
	if err := json.NewDecoder(r.Body).Decode(deleteReq); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	err = d.CloudBackupSchedDelete(deleteReq)
	if err != nil {
		vd.sendError(method, deleteReq.UUID, w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (vd *volAPI) cloudBackupSchedEnumerate(w http.ResponseWriter, r *http.Request) {
	method := "cloudBackupSchedEnumerate"
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	schedules, err := d.CloudBackupSchedEnumerate()
	if err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(schedules)
}
