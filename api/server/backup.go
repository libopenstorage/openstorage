package server

import (
	"encoding/json"
	"net/http"

	"github.com/libopenstorage/openstorage/api"
	sdk "github.com/libopenstorage/openstorage/api/server/sdk"
	prototime "github.com/libopenstorage/openstorage/pkg/proto/time"
	"github.com/libopenstorage/openstorage/volume"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (vd *volAPI) cloudBackupCreate(w http.ResponseWriter, r *http.Request) {
	backupReq := &api.CloudBackupCreateRequest{}
	var backupResp api.CloudBackupCreateResponse
	method := "cloudBackupCreate"

	if err := json.NewDecoder(r.Body).Decode(backupReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	volumes := api.NewOpenStorageCloudBackupClient(conn)
	createResp, err := volumes.Create(ctx, &api.SdkCloudBackupCreateRequest{
		VolumeId:            backupReq.VolumeID,
		CredentialId:        backupReq.CredentialUUID,
		Full:                backupReq.Full,
		TaskId:              backupReq.Name,
		Labels:              backupReq.Labels,
		FullBackupFrequency: backupReq.FullBackupFrequency,
	})
	if err != nil {
		if serverError, ok := status.FromError(err); ok {
			if serverError.Code() == codes.AlreadyExists {
				w.WriteHeader(http.StatusConflict)
				return
			}
		}
		vd.sendError(method, backupReq.VolumeID, w, err.Error(), http.StatusInternalServerError)
		return
	}
	backupResp.Name = createResp.TaskId
	json.NewEncoder(w).Encode(&backupResp)
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

	createResp, err := d.CloudBackupGroupCreate(backupGroupReq)
	if err != nil {
		vd.sendError(method, backupGroupReq.GroupID, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(createResp)
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

	/*
		NEED TO ADD TO SDK
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
	*/

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
	var enumerateResp api.CloudBackupEnumerateResponse
	if err := json.NewDecoder(r.Body).Decode(enumerateReq); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	volumes := api.NewOpenStorageCloudBackupClient(conn)
	sdkEnumerateResp, err := volumes.EnumerateWithFilters(ctx, &api.SdkCloudBackupEnumerateWithFiltersRequest{
		SrcVolumeId:       enumerateReq.SrcVolumeID,
		ClusterId:         enumerateReq.ClusterID,
		CredentialId:      enumerateReq.CredentialUUID,
		All:               enumerateReq.All,
		ContinuationToken: enumerateReq.ContinuationToken,
		MaxBackups:        enumerateReq.MaxBackups,
		MetadataFilter:    enumerateReq.MetadataFilter,
		StatusFilter:      api.CloudBackupStatusTypeToSdkCloudBackupStatusType(enumerateReq.StatusFilter),
	})
	if err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusInternalServerError)
		return
	}
	enumerateResp.Backups = make([]api.CloudBackupInfo, 0)
	for _, v := range sdkEnumerateResp.Backups {
		item := api.CloudBackupInfo{
			ID:            v.Id,
			SrcVolumeID:   v.SrcVolumeId,
			SrcVolumeName: v.SrcVolumeName,
			Timestamp:     prototime.TimestampToTime(v.Timestamp),
			Metadata:      v.Metadata,
			Status:        api.SdkCloudBackupStatusTypeToCloudBackupStatusString(v.Status),
		}
		enumerateResp.Backups = append(enumerateResp.Backups, item)
	}
	enumerateResp.ContinuationToken = sdkEnumerateResp.ContinuationToken
	json.NewEncoder(w).Encode(&enumerateResp)
}

func (vd *volAPI) cloudBackupStatus(w http.ResponseWriter, r *http.Request) {
	method := "cloudBackupStatus"
	backupStatus := &api.CloudBackupStatusRequestOld{}

	if err := json.NewDecoder(r.Body).Decode(backupStatus); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}
	if backupStatus.Name != "" {
		backupStatus.CloudBackupStatusRequest.ID = backupStatus.Name
	}

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	cloudBackups := api.NewOpenStorageCloudBackupClient(conn)
	sts, err := cloudBackups.Status(ctx, &api.SdkCloudBackupStatusRequest{
		VolumeId: backupStatus.CloudBackupStatusRequest.SrcVolumeID,
		Local:    backupStatus.CloudBackupStatusRequest.Local,
		TaskId:   backupStatus.CloudBackupStatusRequest.ID,
	})

	if err != nil {
		if serverError, ok := status.FromError(err); ok {
			if serverError.Code() == codes.Unavailable {
				w.WriteHeader(http.StatusConflict)
				return
			}
		}
		vd.sendError(method, backupStatus.CloudBackupStatusRequest.ID, w, err.Error(), http.StatusInternalServerError)
		return
	}
	backupStatusResp := api.CloudBackupStatusResponse{}
	backupStatusResp.Statuses = make(map[string]api.CloudBackupStatus)
	for task, s := range sts.GetStatuses() {
		backupStatusResp.Statuses[task] = api.CloudBackupStatus{
			ID:                 s.GetBackupId(),
			OpType:             api.SdkCloudBackupOpTypeToCloudBackupOpType(s.GetOptype()),
			Status:             api.CloudBackupStatusType(api.SdkCloudBackupStatusTypeToCloudBackupStatusString(s.GetStatus())),
			BytesDone:          s.GetBytesDone(),
			BytesTotal:         s.GetBytesTotal(),
			EtaSeconds:         s.GetEtaSeconds(),
			StartTime:          prototime.TimestampToTime(s.GetStartTime()),
			CompletedTime:      prototime.TimestampToTime(s.GetCompletedTime()),
			NodeID:             s.GetNodeId(),
			SrcVolumeID:        s.GetSrcVolumeId(),
			Info:               s.GetInfo(),
			CredentialUUID:     s.GetCredentialId(),
			GroupCloudBackupID: s.GetGroupId(),
		}
	}
	json.NewEncoder(w).Encode(&backupStatusResp)
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

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	cloudBackups := api.NewOpenStorageCloudBackupClient(conn)
	_, err = cloudBackups.StateChange(ctx, &api.SdkCloudBackupStateChangeRequest{
		TaskId:         stateChangeReq.Name,
		RequestedState: api.CloudBackupRequestedStateToSdkCloudBackupRequestedState(stateChangeReq.RequestedState),
	})

	if err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (vd *volAPI) cloudBackupSchedCreate(w http.ResponseWriter, r *http.Request) {
	method := "cloudBackupSchedCreate"
	backupSchedReq := &api.CloudBackupSchedCreateRequest{}
	var backupSchedResp api.CloudBackupSchedCreateResponse
	if err := json.NewDecoder(r.Body).Decode(backupSchedReq); err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	volumes := api.NewOpenStorageCloudBackupClient(conn)
	resp, err := volumes.SchedCreate(ctx, &api.SdkCloudBackupSchedCreateRequest{
		CloudSchedInfo: sdk.ToSdkCloudBackupdScheduleInfo(api.CloudBackupScheduleInfo{
			SrcVolumeID:    backupSchedReq.SrcVolumeID,
			CredentialUUID: backupSchedReq.CredentialUUID,
			Schedule:       backupSchedReq.Schedule,
			MaxBackups:     backupSchedReq.MaxBackups,
			RetentionDays:  backupSchedReq.RetentionDays,
		}),
	})
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	backupSchedResp.UUID = resp.BackupScheduleId
	json.NewEncoder(w).Encode(&backupSchedResp)
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
