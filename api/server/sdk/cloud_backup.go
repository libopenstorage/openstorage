/*
Package sdk is the gRPC implementation of the SDK gRPC server
Copyright 2018 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package sdk

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CloudBackupServer is an implementation of the gRPC OpenStorageCloudBackup interface
type CloudBackupServer struct {
	server serverAccessor
}

func (s *CloudBackupServer) driver(ctx context.Context) volume.VolumeDriver {
	return s.server.driver(ctx)
}

// Create creates a backup for a volume
func (s *CloudBackupServer) Create(
	ctx context.Context,
	req *api.SdkCloudBackupCreateRequest,
) (*api.SdkCloudBackupCreateResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	credId := req.GetCredentialId()
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}
	if len(req.GetCredentialId()) == 0 {
		credId, err = s.defaultCloudBackupCreds(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Check ownership
	if err := checkAccessFromDriverForVolumeId(ctx, s.driver(ctx), req.GetVolumeId(), api.Ownership_Read); err != nil {
		return nil, err
	}

	if len(req.GetCredentialId()) != 0 {
		if err := s.checkAccessToCredential(ctx, credId); err != nil {
			return nil, err
		}
	}

	r, err := s.driver(ctx).CloudBackupCreate(&api.CloudBackupCreateRequest{
		VolumeID:       req.GetVolumeId(),
		CredentialUUID: credId,
		Full:           req.GetFull(),
		Name:           req.GetTaskId(),
		Labels:         req.GetLabels(),
	})
	if err != nil {
		if err == volume.ErrExist {
			return nil, status.Errorf(codes.AlreadyExists, "Backup with this name already exists: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "Failed to create backup: %v", err)
	}

	return &api.SdkCloudBackupCreateResponse{
		TaskId: r.Name,
	}, nil
}

// Restore a backup
func (s *CloudBackupServer) Restore(
	ctx context.Context,
	req *api.SdkCloudBackupRestoreRequest,
) (*api.SdkCloudBackupRestoreResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	credId := req.GetCredentialId()
	var err error
	if len(req.GetBackupId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide backup id")
	} else if len(req.GetCredentialId()) == 0 {
		credId, err = s.defaultCloudBackupCreds(ctx)
		if err != nil {
			return nil, err
		}
	}

	if len(req.GetCredentialId()) != 0 {
		if err := s.checkAccessToCredential(ctx, req.GetCredentialId()); err != nil {
			return nil, err
		}
	}

	r, err := s.driver(ctx).CloudBackupRestore(&api.CloudBackupRestoreRequest{
		ID:                req.GetBackupId(),
		RestoreVolumeName: req.GetRestoreVolumeName(),
		CredentialUUID:    credId,
		NodeID:            req.GetNodeId(),
		Name:              req.GetTaskId(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to restore backup: %v", err)
	}

	return &api.SdkCloudBackupRestoreResponse{
		RestoreVolumeId: r.RestoreVolumeID,
		TaskId:          r.Name,
	}, nil

}

// Delete deletes a backup
func (s *CloudBackupServer) Delete(
	ctx context.Context,
	req *api.SdkCloudBackupDeleteRequest,
) (*api.SdkCloudBackupDeleteResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	credId := req.GetCredentialId()
	var err error
	if len(req.GetBackupId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide backup id")
	} else if len(req.GetCredentialId()) == 0 {
		credId, err = s.defaultCloudBackupCreds(ctx)
		if err != nil {
			return nil, err
		}
	}
	if len(req.GetCredentialId()) != 0 {
		if err := s.checkAccessToCredential(ctx, req.GetCredentialId()); err != nil {
			return nil, err
		}
	}

	if err := s.driver(ctx).CloudBackupDelete(&api.CloudBackupDeleteRequest{
		ID:             req.GetBackupId(),
		CredentialUUID: credId,
		Force:          req.GetForce(),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete backup: %v", err)
	}

	return &api.SdkCloudBackupDeleteResponse{}, nil
}

// DeleteAll deletes all backups for a certain volume
func (s *CloudBackupServer) DeleteAll(
	ctx context.Context,
	req *api.SdkCloudBackupDeleteAllRequest,
) (*api.SdkCloudBackupDeleteAllResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	credId := req.GetCredentialId()
	var err error
	if len(req.GetSrcVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide source volume id")
	} else if len(req.GetCredentialId()) == 0 {
		credId, err = s.defaultCloudBackupCreds(ctx)
		if err != nil {
			return nil, err
		}
	}
	if len(req.GetCredentialId()) != 0 {
		if err := s.checkAccessToCredential(ctx, req.GetCredentialId()); err != nil {
			return nil, err
		}
	}
	if err := s.driver(ctx).CloudBackupDeleteAll(&api.CloudBackupDeleteAllRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			SrcVolumeID:    req.GetSrcVolumeId(),
			CredentialUUID: credId,
		},
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete backup: %v", err)
	}

	return &api.SdkCloudBackupDeleteAllResponse{}, nil
}

// Enumerate returns information about the backups
func (s *CloudBackupServer) EnumerateWithFilters(
	ctx context.Context,
	req *api.SdkCloudBackupEnumerateWithFiltersRequest,
) (*api.SdkCloudBackupEnumerateWithFiltersResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	credId := req.GetCredentialId()
	var err error
	if len(req.GetCredentialId()) == 0 {
		credId, err = s.defaultCloudBackupCreds(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.checkAccessToCredential(ctx, req.GetCredentialId()); err != nil {
			return nil, err
		}
	}

	r, err := s.driver(ctx).CloudBackupEnumerate(&api.CloudBackupEnumerateRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			SrcVolumeID:    req.GetSrcVolumeId(),
			ClusterID:      req.GetClusterId(),
			CredentialUUID: credId,
			All:            req.GetAll(),
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to enumerate backups: %v", err)
	}

	return r.ToSdkCloudBackupEnumerateWithFiltersResponse(), nil
}

// Status provides status on a backup
func (s *CloudBackupServer) Status(
	ctx context.Context,
	req *api.SdkCloudBackupStatusRequest,
) (*api.SdkCloudBackupStatusResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	// Check ownership
	if req.GetVolumeId() != "" {
		if err := checkAccessFromDriverForVolumeId(ctx, s.driver(ctx), req.GetVolumeId(), api.Ownership_Read); err != nil {
			return nil, err
		}
	}

	r, err := s.driver(ctx).CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: req.GetVolumeId(),
		Local:       req.GetLocal(),
		ID:          req.GetTaskId(),
	})
	if err != nil {
		if err == volume.ErrInvalidName {
			return nil, status.Errorf(codes.Unavailable, "No Backup status found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to get status of backup: %v", err)
	}
	// Get volume id from task id
	// remove the volumes that dont belong to caller
	for key, sts := range r.Statuses {
		// Allow failed restores to be seen by all
		if sts.OpType == api.CloudRestoreOp &&
			(sts.Status == api.CloudBackupStatusFailed ||
				sts.Status == api.CloudBackupStatusAborted ||
				sts.Status == api.CloudBackupStatusStopped) {
			continue
		}
		if err := checkAccessFromDriverForVolumeId(ctx, s.driver(ctx), sts.SrcVolumeID, api.Ownership_Read); err != nil {
			delete(r.Statuses, key)
		}
	}
	return r.ToSdkCloudBackupStatusResponse(), nil
}

// Catalog
func (s *CloudBackupServer) Catalog(
	ctx context.Context,
	req *api.SdkCloudBackupCatalogRequest,
) (*api.SdkCloudBackupCatalogResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetBackupId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide backup id")
	}
	credId := req.GetCredentialId()
	var err error
	if len(req.GetCredentialId()) == 0 {
		credId, err = s.defaultCloudBackupCreds(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.checkAccessToCredential(ctx, req.GetCredentialId()); err != nil {
			return nil, err
		}
	}

	r, err := s.driver(ctx).CloudBackupCatalog(&api.CloudBackupCatalogRequest{
		ID:             req.GetBackupId(),
		CredentialUUID: credId,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get catalog: %v", err)
	}

	return &api.SdkCloudBackupCatalogResponse{
		Contents: r.Contents,
	}, nil
}

// History returns ??
func (s *CloudBackupServer) History(
	ctx context.Context,
	req *api.SdkCloudBackupHistoryRequest,
) (*api.SdkCloudBackupHistoryResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetSrcVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide volume id")
	}

	// Check ownership
	if err := checkAccessFromDriverForVolumeId(ctx, s.driver(ctx), req.GetSrcVolumeId(), api.Ownership_Read); err != nil {
		return nil, err
	}

	r, err := s.driver(ctx).CloudBackupHistory(&api.CloudBackupHistoryRequest{
		SrcVolumeID: req.GetSrcVolumeId(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get history: %v", err)
	}

	return r.ToSdkCloudBackupHistoryResponse(), nil
}

// StateChange pauses and resumes backups
func (s *CloudBackupServer) StateChange(
	ctx context.Context,
	req *api.SdkCloudBackupStateChangeRequest,
) (*api.SdkCloudBackupStateChangeResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	// TODO
	// XXX Get vid from tid

	if len(req.GetTaskId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide taskid")
	} else if req.GetRequestedState() == api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStateUnknown {
		return nil, status.Error(codes.InvalidArgument, "Must provide requested state")
	}

	var rs string
	switch req.GetRequestedState() {
	// Not supported yet
	/*
		case api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStatePause:
			rs = api.CloudBackupRequestedStatePause
		case api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStateResume:
			rs = api.CloudBackupRequestedStateResume
	*/
	case api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStateStop:
		rs = api.CloudBackupRequestedStateStop
	default:
		return nil, status.Errorf(codes.InvalidArgument, "Invalid requested state: %v", req.GetRequestedState())
	}

	// Get Status to get the volId
	r, err := s.driver(ctx).CloudBackupStatus(&api.CloudBackupStatusRequest{
		ID: req.GetTaskId(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get status of backup: %v", err)
	}
	// Get volume id from task id
	// remove the volumes that dont belong to caller
	for _, sts := range r.Statuses {
		if err := checkAccessFromDriverForVolumeId(ctx, s.driver(ctx), sts.SrcVolumeID, api.Ownership_Write); err != nil {
			return nil, err
		}
	}
	err = s.driver(ctx).CloudBackupStateChange(&api.CloudBackupStateChangeRequest{
		Name:           req.GetTaskId(),
		RequestedState: rs,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to change state: %v", err)
	}

	return &api.SdkCloudBackupStateChangeResponse{}, nil
}

// SchedCreate new schedule for cloud backup
func (s *CloudBackupServer) SchedCreate(
	ctx context.Context,
	req *api.SdkCloudBackupSchedCreateRequest,
) (*api.SdkCloudBackupSchedCreateResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	credId := req.GetCloudSchedInfo().GetCredentialId()
	var err error
	if req.GetCloudSchedInfo() == nil {
		return nil, status.Error(codes.InvalidArgument, "BackupSchedule object cannot be nil")
	} else if len(req.GetCloudSchedInfo().GetSrcVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply source volume id")
	} else if len(req.GetCloudSchedInfo().GetCredentialId()) == 0 {
		credId, err = s.defaultCloudBackupCreds(ctx)
		if err != nil {
			return nil, err
		}
	} else if req.GetCloudSchedInfo().GetSchedules() == nil ||
		len(req.GetCloudSchedInfo().GetSchedules()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Schedule")
	}

	// Check ownership
	if err := checkAccessFromDriverForVolumeId(ctx, s.driver(ctx), req.GetCloudSchedInfo().GetSrcVolumeId(), api.Ownership_Read); err != nil {
		return nil, err
	}
	if len(req.GetCloudSchedInfo().GetCredentialId()) != 0 {
		if err := s.checkAccessToCredential(ctx, req.GetCloudSchedInfo().GetCredentialId()); err != nil {
			return nil, err
		}
	}

	sched, err := sdkSchedToRetainInternalSpecYamlByte(req.GetCloudSchedInfo().GetSchedules())
	if err != nil {
		return nil, err
	}

	bkpRequest := api.CloudBackupSchedCreateRequest{}
	bkpRequest.SrcVolumeID = req.GetCloudSchedInfo().GetSrcVolumeId()
	bkpRequest.CredentialUUID = credId
	bkpRequest.Schedule = string(sched)
	bkpRequest.MaxBackups = uint(req.GetCloudSchedInfo().GetMaxBackups())
	bkpRequest.Full = req.GetCloudSchedInfo().GetFull()

	// Create the backup
	schedResp, err := s.driver(ctx).CloudBackupSchedCreate(&bkpRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create backup: %v", err)
	}

	return &api.SdkCloudBackupSchedCreateResponse{
		BackupScheduleId: schedResp.UUID,
	}, nil

}

// SchedDelete cloud backup schedule
func (s *CloudBackupServer) SchedDelete(
	ctx context.Context,
	req *api.SdkCloudBackupSchedDeleteRequest,
) (*api.SdkCloudBackupSchedDeleteResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	// TODO
	// XXX inspect from uuid and get volume id

	if len(req.GetBackupScheduleId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide volumeId")
	}

	// Call cloud backup driver function to delete cloud schedule
	if err := s.driver(ctx).CloudBackupSchedDelete(&api.CloudBackupSchedDeleteRequest{
		UUID: req.GetBackupScheduleId(),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete cloud backup schedule: %v", err)
	}

	return &api.SdkCloudBackupSchedDeleteResponse{}, nil
}

// SchedEnumerate cloud backup schedule
func (s *CloudBackupServer) SchedEnumerate(
	ctx context.Context,
	req *api.SdkCloudBackupSchedEnumerateRequest,
) (*api.SdkCloudBackupSchedEnumerateResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	// Pass in ownership and show only valid ones
	r, err := s.driver(ctx).CloudBackupSchedEnumerate()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to enumerate backups: %v", err)
	}
	// since can't import sdk/utils to api because of cyclic import, converting
	// api.CloudBackupScheduleInfo to api.SdkCloudBackupScheduleInfo
	return ToSdkCloudBackupSchedEnumerateResponse(r), nil
}

func (s *CloudBackupServer) checkAccessToCredential(
	ctx context.Context,
	credentialId string,
) error {
	cs := &CredentialServer{server: s.server}
	_, err := cs.Inspect(ctx, &api.SdkCredentialInspectRequest{
		CredentialId: credentialId,
	})
	return err
}

func ToSdkCloudBackupSchedEnumerateResponse(r *api.CloudBackupSchedEnumerateResponse) *api.SdkCloudBackupSchedEnumerateResponse {
	resp := &api.SdkCloudBackupSchedEnumerateResponse{
		CloudSchedList: make(map[string]*api.SdkCloudBackupScheduleInfo),
	}

	for k, v := range r.Schedules {
		resp.CloudSchedList[k] = ToSdkCloudBackupdScheduleInfo(v)
	}

	return resp
}

func ToSdkCloudBackupdScheduleInfo(s api.CloudBackupScheduleInfo) *api.SdkCloudBackupScheduleInfo {

	schedules, err := retainInternalSpecYamlByteToSdkSched([]byte(s.Schedule))
	if err != nil {
		return nil
	}
	cloudSched := &api.SdkCloudBackupScheduleInfo{
		SrcVolumeId:  s.SrcVolumeID,
		CredentialId: s.CredentialUUID,
		Schedules:    schedules,
		// Not sure about go and protobuf type conversion, converting to higher type
		// converting uint to uint64
		MaxBackups: uint64(s.MaxBackups),
		Full:       s.Full,
	}
	return cloudSched
}

func (s *CloudBackupServer) defaultCloudBackupCreds(
	ctx context.Context,
) (string, error) {

	req := &api.SdkCredentialEnumerateRequest{}
	cs := &CredentialServer{server: s.server}
	credList, err := cs.Enumerate(ctx, req)
	if err != nil {
		return "", err
	}

	if len(credList.CredentialIds) > 1 {
		return "", status.Error(codes.InvalidArgument, "More than one credential configured,"+
			"please specify a credential name or uuid to use")
	}
	if len(credList.CredentialIds) == 0 {
		return "", status.Error(codes.InvalidArgument, "No configured credentials found,"+
			"please create a credential")
	}
	return credList.CredentialIds[0], nil
}
