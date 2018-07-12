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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
)

// SnapshotCreate creates a read-only snapshot of a volume
func (s *VolumeServer) SnapshotCreate(
	ctx context.Context,
	req *api.SdkVolumeSnapshotCreateRequest,
) (*api.SdkVolumeSnapshotCreateResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	readonly := true
	snapshotID, err := s.driver.Snapshot(req.GetVolumeId(), readonly, &api.VolumeLocator{
		VolumeLabels: req.GetLabels(),
	})
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

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	} else if len(req.GetSnapshotId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply snapshot id")
	}

	err := s.driver.Restore(req.GetVolumeId(), req.GetSnapshotId())
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

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	snapshots, err := s.driver.SnapEnumerate([]string{req.GetVolumeId()}, req.GetLabels())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to enumerate snapshots in volume %s: %v",
			req.GetVolumeId(),
			err.Error())
	}

	ids := make([]string, len(snapshots))
	for i, snapshot := range snapshots {
		ids[i] = snapshot.GetId()
	}

	return &api.SdkVolumeSnapshotEnumerateResponse{
		VolumeSnapshotIds: ids,
	}, nil
}
