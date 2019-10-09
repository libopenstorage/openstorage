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
	ost_errors "github.com/libopenstorage/openstorage/api/errors"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Start a volume migration
func (s *VolumeServer) Start(
	ctx context.Context,
	req *api.SdkCloudMigrateStartRequest,
) (*api.SdkCloudMigrateStartResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if volume := req.GetVolume(); volume != nil {
		// Check ownership
		if err := checkAccessFromDriverForVolumeIds(ctx, s.driver(ctx), []string{volume.GetVolumeId()}, api.Ownership_Read); err != nil {
			return nil, err
		}

		//migrate volume
		return s.volumeMigrate(ctx, req, volume)
	} else if volumeGroup := req.GetVolumeGroup(); volumeGroup != nil {
		if !s.haveOwnership(ctx, nil, &api.VolumeLocator{
			Group: &api.Group{
				Id: volumeGroup.GetGroupId(),
			},
		}) {
			return nil, status.Error(codes.PermissionDenied, "Volume Operation not Permitted.")
		}
		//migrate volume groups
		return s.volumeGroupMigrate(ctx, req, volumeGroup)
	} else if allVolumes := req.GetAllVolumes(); allVolumes != nil {
		// migrate all volumes
		if !s.haveOwnership(ctx, nil, nil) {
			return nil, status.Error(codes.PermissionDenied, "Volume Operation not Permitted.")
		}
		return s.allVolumesMigrate(ctx, req, allVolumes)
	}
	return nil, status.Error(codes.InvalidArgument, "Unknown operation request")
}

func (s *VolumeServer) haveOwnership(ctx context.Context, labels map[string]string, locator *api.VolumeLocator) bool {
	// checking if driver is initialized happens in Start
	vols, err := s.driver(ctx).Enumerate(locator, labels)
	if err != nil {
		return false
	}
	for _, vol := range vols {
		// Check ownership
		if !vol.IsPermitted(ctx, api.Ownership_Read) {
			return false
		}
	}

	return true
}

func (s *VolumeServer) volumeGroupMigrate(
	ctx context.Context,
	req *api.SdkCloudMigrateStartRequest,
	volumeGroup *api.SdkCloudMigrateStartRequest_MigrateVolumeGroup,
) (*api.SdkCloudMigrateStartResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	//Create a request object with operation as Migrate volume
	request := &api.CloudMigrateStartRequest{
		Operation: api.CloudMigrate_MigrateVolumeGroup,
		ClusterId: req.GetClusterId(),
		TargetId:  volumeGroup.GetGroupId(),
		TaskId:    req.GetTaskId(), // optional will be "" if not passed
	}
	resp, err := s.driver(ctx).CloudMigrateStart(request)
	if err != nil {
		if _, ok := err.(*ost_errors.ErrExists); ok {
			return nil, status.Errorf(codes.AlreadyExists, "Cannot start migration for %s : %v", req.GetClusterId(), err)
		}
		// if errExist return codes.
		return nil, status.Errorf(codes.Internal, "Cannot start migration for %s : %v", req.GetClusterId(), err)
	}
	return &api.SdkCloudMigrateStartResponse{
		Result: resp,
	}, nil
}

func (s *VolumeServer) allVolumesMigrate(
	ctx context.Context,
	req *api.SdkCloudMigrateStartRequest,
	allVolume *api.SdkCloudMigrateStartRequest_MigrateAllVolumes,
) (*api.SdkCloudMigrateStartResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	//Create a request object with operation as Migrate volume
	request := &api.CloudMigrateStartRequest{
		Operation: api.CloudMigrate_MigrateCluster,
		ClusterId: req.GetClusterId(),
		TaskId:    req.GetTaskId(),
	}
	resp, err := s.driver(ctx).CloudMigrateStart(request)
	if err != nil {
		if _, ok := err.(*ost_errors.ErrExists); ok {
			return nil, status.Errorf(codes.AlreadyExists, "Cannot start migration for %s : %v", req.GetClusterId(), err)
		}
		return nil, status.Errorf(codes.Internal, "Cannot start migration for %s : %v", req.GetClusterId(), err)
	}
	return &api.SdkCloudMigrateStartResponse{
		Result: resp,
	}, nil
}

func (s *VolumeServer) volumeMigrate(
	ctx context.Context,
	req *api.SdkCloudMigrateStartRequest,
	volume *api.SdkCloudMigrateStartRequest_MigrateVolume,
) (*api.SdkCloudMigrateStartResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	//Create a request object with operation as Migrate volume
	request := &api.CloudMigrateStartRequest{
		Operation: api.CloudMigrate_MigrateVolume,
		ClusterId: req.GetClusterId(),
		TargetId:  volume.GetVolumeId(),
		TaskId:    req.GetTaskId(),
	}
	resp, err := s.driver(ctx).CloudMigrateStart(request)
	if err != nil {
		if _, ok := err.(*ost_errors.ErrExists); ok {
			return nil, status.Errorf(codes.AlreadyExists, "Cannot start migration for %s : %v", req.GetClusterId(), err)
		}
		return nil, status.Errorf(codes.Internal, "Cannot start migration for %s : %v", req.GetClusterId(), err)
	}
	return &api.SdkCloudMigrateStartResponse{
		Result: resp,
	}, nil
}

func (s *VolumeServer) checkMigrationPermissions(ctx context.Context, taskId string) error {
	// Inspect migration to get VolumeIds
	resp, err := s.driver(ctx).CloudMigrateStatus(&api.CloudMigrateStatusRequest{
		TaskId: taskId,
	})
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to get migration information : %v", err)
	}

	// Check that a user has access to all volumes being migrated
	for _, cluster := range resp.Info {
		for _, migrateInfo := range cluster.List {
			if err := checkAccessFromDriverForVolumeIds(ctx, s.driver(ctx),
				[]string{migrateInfo.GetLocalVolumeId()}, api.Ownership_Read); err != nil {
				return err
			}
		}
	}

	return nil
}

// Cancel or stop a ongoing migration
func (s *VolumeServer) Cancel(
	ctx context.Context,
	req *api.SdkCloudMigrateCancelRequest,
) (*api.SdkCloudMigrateCancelResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	if req.GetRequest() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Must supply valid request")
	} else if len(req.GetRequest().GetTaskId()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Must supply valid Task ID")
	}

	// Check if the user has access to all volumes associated with the TaskID
	if err := s.checkMigrationPermissions(ctx, req.GetRequest().GetTaskId()); err != nil {
		return nil, err
	}
	err := s.driver(ctx).CloudMigrateCancel(req.GetRequest())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot stop migration for %s : %v",
			req.GetRequest().GetTaskId(), err)
	}
	return &api.SdkCloudMigrateCancelResponse{}, nil
}

// filterStatusResponseForPermissions alters the response object to only return objects
// that we have access to. While it seems too complicated, it minimizes the number of driver calls.
func (s *VolumeServer) filterStatusResponseForPermissions(
	ctx context.Context,
	resp *api.CloudMigrateStatusResponse) (*api.CloudMigrateStatusResponse, error) {

	// Generate new response with permitted migrate info based
	// on which volume ids we have access to
	var filteredResp api.CloudMigrateStatusResponse
	filteredResp.Info = make(map[string]*api.CloudMigrateInfoList)
	for clusterId, cluster := range resp.Info {
		filteredCluster := api.CloudMigrateInfoList{}
		filteredCluster.List = make([]*api.CloudMigrateInfo, 0)

		for _, migrateInfo := range cluster.List {
			err := checkAccessFromDriverForLocator(ctx, s.driver(ctx), &api.VolumeLocator{
				VolumeIds: []string{migrateInfo.GetLocalVolumeId()},
			}, api.Ownership_Read)
			if err == nil {
				filteredCluster.List = append(filteredCluster.List, migrateInfo)
			}
		}

		// Do not include cluster if we don't have access to any of the volumes.
		if len(cluster.List) > 0 {
			filteredResp.Info[clusterId] = &filteredCluster
		}
	}

	return &filteredResp, nil
}

// Status of ongoing migration
func (s *VolumeServer) Status(
	ctx context.Context,
	req *api.SdkCloudMigrateStatusRequest,
) (*api.SdkCloudMigrateStatusResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	csReq := req.GetRequest()
	if csReq == nil {
		csReq = &api.CloudMigrateStatusRequest{}
	}

	resp, err := s.driver(ctx).CloudMigrateStatus(csReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot get status of migration : %v", err)
	}

	// Filter out volumes we don't have access to
	if auth.Enabled() {
		resp, err = s.filterStatusResponseForPermissions(ctx, resp)
		if err != nil {
			return nil, err
		}
	}

	return &api.SdkCloudMigrateStatusResponse{
		Result: resp,
	}, nil

}
