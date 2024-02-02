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
	"github.com/libopenstorage/openstorage/pkg/sched"
	"github.com/portworx/kvdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SnapshotCreate creates a read-only snapshot of a volume
func (s *VolumeServer) SnapshotCreate(
	ctx context.Context,
	req *api.SdkVolumeSnapshotCreateRequest,
) (*api.SdkVolumeSnapshotCreateResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	} else if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a name")
	}

	// Get access rights
	if err := s.checkAccessForVolumeId(ctx, req.GetVolumeId(), api.Ownership_Read); err != nil {
		return nil, err
	}

	readonly := true
	snapshotID, err := s.driver(ctx).Snapshot(req.GetVolumeId(), readonly, &api.VolumeLocator{
		Name:         req.GetName(),
		VolumeLabels: req.GetLabels(),
	}, false)
	if err != nil {
		if err == kvdb.ErrNotFound {
			return nil, status.Errorf(
				codes.NotFound,
				"Id %s not found",
				req.GetVolumeId())
		}
		return nil, status.Errorf(codes.Internal, "Failed to create snapshot: %v", err.Error())
	}

	return &api.SdkVolumeSnapshotCreateResponse{
		SnapshotId: snapshotID,
	}, nil
}

// SnapshotRestore restores a volume to the specified snapshot id
func (s *VolumeServer) SnapshotRestore(
	ctx context.Context,
	req *api.SdkVolumeSnapshotRestoreRequest,
) (*api.SdkVolumeSnapshotRestoreResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	} else if len(req.GetSnapshotId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply snapshot id")
	}

	// Get access rights
	if err := s.checkAccessForVolumeId(ctx, req.GetVolumeId(), api.Ownership_Write); err != nil {
		return nil, err
	}

	err := s.driver(ctx).Restore(req.GetVolumeId(), req.GetSnapshotId())
	if err != nil {
		if err == kvdb.ErrNotFound {
			return nil, status.Errorf(
				codes.NotFound,
				"Id %s or %s not found",
				req.GetVolumeId(), req.GetSnapshotId())
		}
		return nil, status.Errorf(
			codes.Internal,
			"Failed to restore volume %s to snapshot %s: %v",
			req.GetVolumeId(),
			req.GetSnapshotId(),
			err.Error())
	}

	return &api.SdkVolumeSnapshotRestoreResponse{}, nil
}

// SnapshotEnumerate returns a list of snapshots for the specified volume
func (s *VolumeServer) SnapshotEnumerate(
	ctx context.Context,
	req *api.SdkVolumeSnapshotEnumerateRequest,
) (*api.SdkVolumeSnapshotEnumerateResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	resp, err := s.SnapshotEnumerateWithFilters(
		ctx,
		&api.SdkVolumeSnapshotEnumerateWithFiltersRequest{
			VolumeId: req.GetVolumeId(),
		},
	)
	if err != nil {
		return nil, err
	}

	return &api.SdkVolumeSnapshotEnumerateResponse{
		VolumeSnapshotIds: resp.GetVolumeSnapshotIds(),
	}, nil

}

// SnapshotEnumerateWithFilters returns a list of snapshots for the specified
// volume and labels
func (s *VolumeServer) SnapshotEnumerateWithFilters(
	ctx context.Context,
	req *api.SdkVolumeSnapshotEnumerateWithFiltersRequest,
) (*api.SdkVolumeSnapshotEnumerateWithFiltersResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	// Get access rights
	var volReq []string
	if len(req.GetVolumeId()) != 0 {
		if err := s.checkAccessForVolumeId(ctx, req.GetVolumeId(), api.Ownership_Read); err != nil {
			return nil, err
		}
		volReq = []string{req.GetVolumeId()}
	} else {
		volReq = nil
	}

	snapshots, err := s.driver(ctx).SnapEnumerate(volReq, req.GetLabels())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to enumerate snapshots in volume %s: %v",
			req.GetVolumeId(),
			err.Error())
	}

	ids := make([]string, 0)
	for _, snapshot := range snapshots {
		// Check access
		if snapshot.IsPermitted(ctx, api.Ownership_Read) {
			ids = append(ids, snapshot.GetId())
		}
	}

	return &api.SdkVolumeSnapshotEnumerateWithFiltersResponse{
		VolumeSnapshotIds: ids,
	}, nil
}

// SnapshotScheduleUpdate updates the snapshot schedule in the volume.
// It only manages the PolicyTags
func (s *VolumeServer) SnapshotScheduleUpdate(
	ctx context.Context,
	req *api.SdkVolumeSnapshotScheduleUpdateRequest,
) (*api.SdkVolumeSnapshotScheduleUpdateResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	// Get volume specification
	// This checks for access also
	resp, err := s.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: req.GetVolumeId(),
	})
	if err != nil {
		return nil, err
	}

	// Check if caller has access to affect volume
	if !resp.GetVolume().IsPermitted(ctx, api.Ownership_Write) {
		return nil, status.Errorf(
			codes.PermissionDenied,
			"Cannot change the snapshot schedule for volume %s",
			req.GetVolumeId())
	}

	// Determine if they exist
	for _, name := range req.GetSnapshotScheduleNames() {
		_, err := s.cluster().SchedPolicyGet(name)
		if err != nil {
			return nil, status.Errorf(
				codes.Aborted,
				"Error accessing schedule policy %s: %v",
				name, err)
		}
	}

	// Apply names to snapshot schedule in the Volume specification
	// merging with any schedule already there in "schedule" format.
	var pt *sched.PolicyTags
	if len(req.GetSnapshotScheduleNames()) != 0 {
		pt, err = sched.NewPolicyTagsFromSlice(req.GetSnapshotScheduleNames())
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"Unable to parse policies: %v", err)
		}
	}
	snapscheds, _, err := sched.ParseScheduleAndPolicies(resp.GetVolume().GetSpec().GetSnapshotSchedule())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Unable to parse snapshot schedule: %v", err)
	}
	snapshotSchedule := sched.ScheduleSummary(snapscheds, pt)

	// Update the volume specification
	_, err = s.Update(ctx, &api.SdkVolumeUpdateRequest{
		VolumeId: req.GetVolumeId(),
		Spec: &api.VolumeSpecUpdate{
			SnapshotScheduleOpt: &api.VolumeSpecUpdate_SnapshotSchedule{
				SnapshotSchedule: snapshotSchedule,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &api.SdkVolumeSnapshotScheduleUpdateResponse{}, nil
}
