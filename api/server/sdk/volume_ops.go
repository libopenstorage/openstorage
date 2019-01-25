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
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/pkg/util"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/portworx/kvdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *VolumeServer) create(
	ctx context.Context,
	locator *api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec,
) (string, error) {

	// Check if the volume has already been created or is in process of creation
	volName := locator.GetName()
	v, err := util.VolumeFromName(s.driver(), volName)
	if err == nil {
		// Check the requested arguments match that of the existing volume
		if v.GetSpec().GetSize() != spec.GetSize() {
			return "", status.Errorf(
				codes.AlreadyExists,
				"Existing volume has a size of %v which differs from requested size of %v",
				v.GetSpec().GetSize(),
				spec.Size)
		}
		if v.GetSpec().GetShared() != spec.GetShared() {
			return "", status.Errorf(
				codes.AlreadyExists,
				"Existing volume has shared=%v while request is asking for shared=%v",
				v.GetSpec().GetShared(),
				spec.GetShared())
		}
		if v.GetSource().GetParent() != source.GetParent() {
			return "", status.Error(codes.AlreadyExists, "Existing volume has conflicting parent value")
		}

		// Return information on existing volume
		return v.GetId(), nil
	}

	// Check if the caller is asking to create a snapshot or for a new volume
	var id string
	if len(source.GetParent()) != 0 {
		// Get parent volume information
		parent, err := util.VolumeFromName(s.driver(), source.Parent)
		if err != nil {
			return "", status.Errorf(
				codes.InvalidArgument,
				"unable to get parent volume information: %s",
				err.Error())
		}

		// Check ownership
		if !parent.IsPermitted(ctx) {
			return "", status.Errorf(codes.PermissionDenied, "Access denied to volume %s", parent.GetId())
		}

		// Create a snapshot from the parent
		id, err = s.driver().Snapshot(parent.GetId(), false, &api.VolumeLocator{
			Name: volName,
		}, false)
		if err != nil {
			return "", status.Errorf(
				codes.Internal,
				"unable to create snapshot: %s",
				err.Error())
		}
	} else {
		// New volume, set ownership
		spec.Ownership = api.OwnershipSetUsernameFromContext(ctx, spec.Ownership)

		// Create the volume
		id, err = s.driver().Create(locator, source, spec)
		if err != nil {
			return "", status.Errorf(
				codes.Internal,
				"Failed to create volume: %v",
				err.Error())
		}
	}

	return id, nil
}

// Create creates a new volume
func (s *VolumeServer) Create(
	ctx context.Context,
	req *api.SdkVolumeCreateRequest,
) (*api.SdkVolumeCreateResponse, error) {
	if s.cluster() == nil || s.driver() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetName()) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"Must supply a unique name")
	} else if req.GetSpec() == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"Must supply spec object")
	}

	spec := req.GetSpec()
	locator := &api.VolumeLocator{
		Name:         req.GetName(),
		VolumeLabels: req.GetLabels(),
	}
	source := &api.Source{}

	// Copy any labels from the spec to the locator
	locator = locator.MergeVolumeSpecLabels(spec)

	// Convert node IP to ID if necessary for API calls
	if err := s.updateReplicaSpecNodeIPstoIds(spec.GetReplicaSet()); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get replicat set information: %v", err)
	}

	// Create volume
	id, err := s.create(ctx, locator, source, spec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.SdkVolumeCreateResponse{
		VolumeId: id,
	}, nil
}

// Clone creates a new volume from an existing volume
func (s *VolumeServer) Clone(
	ctx context.Context,
	req *api.SdkVolumeCloneRequest,
) (*api.SdkVolumeCloneResponse, error) {
	if s.cluster() == nil || s.driver() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetName()) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"Must supply a uniqe name")
	} else if len(req.GetParentId()) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"Must parent volume id")
	}

	locator := &api.VolumeLocator{
		Name: req.GetName(),
	}
	source := &api.Source{
		Parent: req.GetParentId(),
	}

	// Get spec. This also checks if the parend id exists.
	// This will check access rights also
	parentVol, err := s.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: req.GetParentId(),
	})
	if err != nil {
		return nil, err
	}

	// Create the clone
	id, err := s.create(ctx, locator, source, parentVol.GetVolume().GetSpec())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.SdkVolumeCloneResponse{
		VolumeId: id,
	}, nil
}

// Delete deletes a volume
func (s *VolumeServer) Delete(
	ctx context.Context,
	req *api.SdkVolumeDeleteRequest,
) (*api.SdkVolumeDeleteResponse, error) {
	if s.cluster() == nil || s.driver() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	// If the volume is not found, return OK to be idempotent
	// This checks access rights also
	resp, err := s.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: req.GetVolumeId(),
	})
	if err != nil {
		if gErr, ok := status.FromError(err); ok {
			if gErr.Code() == codes.NotFound {
				return &api.SdkVolumeDeleteResponse{}, nil
			}
		}
		return nil, err
	}

	// Only the owner or the admin can delete
	if !resp.GetVolume().GetSpec().IsPermittedToDelete(ctx) {
		return nil, status.Errorf(codes.PermissionDenied, "Cannot delete volume, only the owner can delete the volume")
	}

	// Delete the volume
	err = s.driver().Delete(req.GetVolumeId())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to delete volume %s: %v",
			req.GetVolumeId(),
			err.Error())
	}

	return &api.SdkVolumeDeleteResponse{}, nil
}

// Inspect returns information about a volume
func (s *VolumeServer) Inspect(
	ctx context.Context,
	req *api.SdkVolumeInspectRequest,
) (*api.SdkVolumeInspectResponse, error) {
	if s.cluster() == nil || s.driver() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	vols, err := s.driver().Inspect([]string{req.GetVolumeId()})
	if err == kvdb.ErrNotFound || (err == nil && len(vols) == 0) {
		return nil, status.Errorf(
			codes.NotFound,
			"Volume id %s not found",
			req.GetVolumeId())
	} else if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to inspect volume %s: %v",
			req.GetVolumeId(), err)
	}
	v := vols[0]

	// Check ownership
	if !v.IsPermitted(ctx) {
		return nil, status.Errorf(codes.PermissionDenied, "Access denied to volume %s", v.GetId())
	}

	return &api.SdkVolumeInspectResponse{
		Volume: v,
		Name:   v.GetLocator().GetName(),
		Labels: v.GetLocator().GetVolumeLabels(),
	}, nil
}

// Enumerate returns a list of volumes
func (s *VolumeServer) Enumerate(
	ctx context.Context,
	req *api.SdkVolumeEnumerateRequest,
) (*api.SdkVolumeEnumerateResponse, error) {
	if s.cluster() == nil || s.driver() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	resp, err := s.EnumerateWithFilters(
		ctx,
		&api.SdkVolumeEnumerateWithFiltersRequest{},
	)
	if err != nil {
		return nil, err
	}

	return &api.SdkVolumeEnumerateResponse{
		VolumeIds: resp.GetVolumeIds(),
	}, nil

}

// EnumerateWithFilters returns a list of volumes for the provided filters
func (s *VolumeServer) EnumerateWithFilters(
	ctx context.Context,
	req *api.SdkVolumeEnumerateWithFiltersRequest,
) (*api.SdkVolumeEnumerateWithFiltersResponse, error) {
	if s.cluster() == nil || s.driver() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	var locator *api.VolumeLocator
	if len(req.GetName()) != 0 ||
		len(req.GetLabels()) != 0 ||
		req.GetOwnership() != nil {

		locator = &api.VolumeLocator{
			Name:         req.GetName(),
			VolumeLabels: req.GetLabels(),
			Ownership:    req.GetOwnership(),
		}
	}

	vols, err := s.driver().Enumerate(locator, nil)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to enumerate volumes: %v",
			err.Error())
	}

	ids := make([]string, 0)
	for _, vol := range vols {
		// Check access
		if vol.IsPermitted(ctx) {
			ids = append(ids, vol.GetId())
		}
	}

	return &api.SdkVolumeEnumerateWithFiltersResponse{
		VolumeIds: ids,
	}, nil
}

// Update allows the caller to change values in the volume specification
func (s *VolumeServer) Update(
	ctx context.Context,
	req *api.SdkVolumeUpdateRequest,
) (*api.SdkVolumeUpdateResponse, error) {
	if s.cluster() == nil || s.driver() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	// Get current state
	// This checks for ownership
	resp, err := s.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: req.GetVolumeId(),
	})
	if err != nil {
		return nil, err
	}

	// Merge specs
	spec := s.mergeVolumeSpecs(resp.GetVolume().GetSpec(), req.GetSpec())

	// Update Ownership... carefully
	// First point to the original ownership
	spec.Ownership = resp.GetVolume().GetSpec().GetOwnership()

	// Check if we have been provided an update to the ownership
	if req.GetSpec().GetOwnership() != nil {
		if spec.Ownership == nil {
			spec.Ownership = &api.Ownership{}
		}

		user, _ := auth.NewUserInfoFromContext(ctx)
		if err := spec.Ownership.Update(req.GetSpec().GetOwnership(), user); err != nil {
			return nil, err
		}
	}

	// Check if labels have been updated
	var locator *api.VolumeLocator
	if len(req.GetLabels()) != 0 {
		locator = &api.VolumeLocator{VolumeLabels: req.GetLabels()}
	}

	// Send to driver
	if err := s.driver().Set(req.GetVolumeId(), locator, spec); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update volume: %v", err)
	}

	return &api.SdkVolumeUpdateResponse{}, nil
}

// Stats returns volume statistics
func (s *VolumeServer) Stats(
	ctx context.Context,
	req *api.SdkVolumeStatsRequest,
) (*api.SdkVolumeStatsResponse, error) {
	if s.cluster() == nil || s.driver() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	// Get access rights
	if err := s.checkAccessForVolumeId(ctx, req.GetVolumeId()); err != nil {
		return nil, err
	}

	stats, err := s.driver().Stats(req.GetVolumeId(), !req.GetNotCumulative())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to obtain stats for volume %s: %v",
			req.GetVolumeId(),
			err.Error())
	}

	return &api.SdkVolumeStatsResponse{
		Stats: stats,
	}, nil
}

func (s *VolumeServer) CapacityUsage(
	ctx context.Context,
	req *api.SdkVolumeCapacityUsageRequest,
) (*api.SdkVolumeCapacityUsageResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	// Get access rights
	if err := s.checkAccessForVolumeId(ctx, req.GetVolumeId()); err != nil {
		return nil, err
	}

	dResp, err := s.driver().CapacityUsage(req.GetVolumeId())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to obtain stats for volume %s: %v",
			req.GetVolumeId(),
			err.Error())
	}

	resp := &api.SdkVolumeCapacityUsageResponse{}
	resp.CapacityUsageInfo = &api.CapacityUsageInfo{}
	resp.CapacityUsageInfo.ExclusiveBytes = dResp.CapacityUsageInfo.ExclusiveBytes
	resp.CapacityUsageInfo.SharedBytes = dResp.CapacityUsageInfo.SharedBytes
	resp.CapacityUsageInfo.TotalBytes = dResp.CapacityUsageInfo.TotalBytes
	if dResp.Error != nil {
		if dResp.Error == volume.ErrAborted {
			return resp, status.Errorf(
				codes.Aborted,
				"Failed to obtain stats for volume %s: %v",
				req.GetVolumeId(),
				volume.ErrAborted.Error())
		} else if dResp.Error == volume.ErrNotSupported {
			return resp, status.Errorf(
				codes.Unimplemented,
				"Failed to obtain stats for volume %s: %v",
				req.GetVolumeId(),
				volume.ErrNotSupported.Error())
		}
	}
	return resp, nil
}

func (s *VolumeServer) mergeVolumeSpecs(vol *api.VolumeSpec, req *api.VolumeSpecUpdate) *api.VolumeSpec {

	spec := &api.VolumeSpec{}
	spec.Shared = setSpecBool(vol.GetShared(), req.GetShared(), req.GetSharedOpt())
	spec.Sharedv4 = setSpecBool(vol.GetSharedv4(), req.GetSharedv4(), req.GetSharedv4Opt())
	spec.Sticky = setSpecBool(vol.GetSticky(), req.GetSticky(), req.GetStickyOpt())
	spec.Journal = setSpecBool(vol.GetJournal(), req.GetJournal(), req.GetJournalOpt())

	// Cos
	if req.GetCosOpt() != nil {
		spec.Cos = req.GetCos()
	} else {
		spec.Cos = vol.GetCos()
	}

	// Passphrase
	if req.GetPassphraseOpt() != nil {
		spec.Passphrase = req.GetPassphrase()
	} else {
		spec.Passphrase = vol.GetPassphrase()
	}

	// Snapshot schedule as a string
	if req.GetSnapshotScheduleOpt() != nil {
		spec.SnapshotSchedule = req.GetSnapshotSchedule()
	} else {
		spec.SnapshotSchedule = vol.GetSnapshotSchedule()
	}

	// Scale
	if req.GetScaleOpt() != nil {
		spec.Scale = req.GetScale()
	} else {
		spec.Scale = vol.GetScale()
	}

	// Snapshot Interval
	if req.GetSnapshotIntervalOpt() != nil {
		spec.SnapshotInterval = req.GetSnapshotInterval()
	} else {
		spec.SnapshotInterval = vol.GetSnapshotInterval()
	}

	// Io Profile
	if req.GetIoProfileOpt() != nil {
		spec.IoProfile = req.GetIoProfile()
	} else {
		spec.IoProfile = vol.GetIoProfile()
	}

	// GroupID
	if req.GetGroupOpt() != nil {
		spec.Group = req.GetGroup()
	} else {
		spec.Group = vol.GetGroup()
	}

	// Size
	if req.GetSizeOpt() != nil {
		spec.Size = req.GetSize()
	} else {
		spec.Size = vol.GetSize()
	}

	// ReplicaSet
	if req.GetReplicaSet() != nil {
		spec.ReplicaSet = req.GetReplicaSet()
	} else {
		spec.ReplicaSet = vol.GetReplicaSet()
	}

	// HA Level
	if req.GetHaLevelOpt() != nil {
		spec.HaLevel = req.GetHaLevel()
	} else {
		spec.HaLevel = vol.GetHaLevel()
	}

	// Queue depth
	if req.GetQueueDepthOpt() != nil {
		spec.QueueDepth = req.GetQueueDepth()
	} else {
		spec.QueueDepth = vol.GetQueueDepth()
	}

	return spec
}

func (s *VolumeServer) nodeIPtoIds(nodes []string) ([]string, error) {
	nodeIds := make([]string, 0)

	for _, idIp := range nodes {
		if idIp != "" {
			id, err := s.cluster().GetNodeIdFromIp(idIp)
			if err != nil {
				return nodeIds, err
			}
			nodeIds = append(nodeIds, id)
		}
	}

	return nodeIds, nil
}

// Convert any replica set node values which are IPs to the corresponding Node ID.
// Update the replica set node list.
func (s *VolumeServer) updateReplicaSpecNodeIPstoIds(rspecRef *api.ReplicaSet) error {
	if rspecRef != nil && len(rspecRef.Nodes) > 0 {
		nodeIds, err := s.nodeIPtoIds(rspecRef.Nodes)
		if err != nil {
			return err
		}

		if len(nodeIds) > 0 {
			rspecRef.Nodes = nodeIds
		}
	}

	return nil
}

func setSpecBool(current, req bool, reqSet interface{}) bool {
	if reqSet != nil {
		return req
	}
	return current
}
