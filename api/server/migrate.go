package server

import (
	"encoding/json"
	"net/http"

	"github.com/libopenstorage/openstorage/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (vd *volAPI) cloudMigrateStart(w http.ResponseWriter, r *http.Request) {
	startReq := &api.CloudMigrateStartRequest{}
	method := "cloudMigrateStart"

	if err := json.NewDecoder(r.Body).Decode(startReq); err != nil {
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

	migrations := api.NewOpenStorageMigrateClient(conn)
	migrateRequest := &api.SdkCloudMigrateStartRequest{
		TaskId:    startReq.TaskId,
		ClusterId: startReq.ClusterId,
	}

	switch startReq.Operation {
	case api.CloudMigrate_MigrateCluster:
		migrateRequest.Opt = &api.SdkCloudMigrateStartRequest_AllVolumes{
			AllVolumes: &api.SdkCloudMigrateStartRequest_MigrateAllVolumes{},
		}
	case api.CloudMigrate_MigrateVolume:
		migrateRequest.Opt = &api.SdkCloudMigrateStartRequest_Volume{
			Volume: &api.SdkCloudMigrateStartRequest_MigrateVolume{
				VolumeId: startReq.TargetId,
			},
		}
	case api.CloudMigrate_MigrateVolumeGroup:
		migrateRequest.Opt = &api.SdkCloudMigrateStartRequest_VolumeGroup{
			VolumeGroup: &api.SdkCloudMigrateStartRequest_MigrateVolumeGroup{
				GroupId: startReq.TargetId,
			},
		}
	}

	resp, err := migrations.Start(ctx, migrateRequest)
	if err != nil {
		if serverError, ok := status.FromError(err); ok {
			if serverError.Code() == codes.AlreadyExists {
				w.WriteHeader(http.StatusConflict)
				return
			}
		}
		vd.sendError(method, startReq.TargetId, w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := &api.CloudMigrateStartResponse{
		TaskId: resp.GetResult().GetTaskId(),
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

	migrations := api.NewOpenStorageMigrateClient(conn)
	migrateRequest := &api.SdkCloudMigrateCancelRequest{
		Request: cancelReq,
	}

	_, err = migrations.Cancel(ctx, migrateRequest)
	if err != nil {
		vd.sendError(method, cancelReq.TaskId, w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (vd *volAPI) cloudMigrateStatus(w http.ResponseWriter, r *http.Request) {
	statusReq := &api.CloudMigrateStatusRequest{}
	method := "cloudMigrateState"

	// Use empty request if nothing was sent
	if r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(statusReq); err != nil {
			vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
			return
		}
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

	migrations := api.NewOpenStorageMigrateClient(conn)
	migrateRequest := &api.SdkCloudMigrateStatusRequest{
		Request: statusReq,
	}

	resp, err := migrations.Status(ctx, migrateRequest)
	if err != nil {
		vd.sendError(method, "", w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusResp := &api.CloudMigrateStatusResponse{
		Info: resp.GetResult().GetInfo(),
	}

	json.NewEncoder(w).Encode(statusResp)
}
