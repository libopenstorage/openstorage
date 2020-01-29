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
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/auth"
	policy "github.com/libopenstorage/openstorage/pkg/storagepolicy"
	"github.com/libopenstorage/openstorage/pkg/util"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/portworx/kvdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// When create is called for an existing volume, this function is called to make sure
// the SDK only returns that the volume is ready when the status is UP
func (s *VolumeServer) waitForVolumeReady(ctx context.Context, id string) (*api.Volume, error) {
	var v *api.Volume

	minTimeout := 1 * time.Second
	maxTimeout := 60 * time.Minute
	defaultTimeout := 10 * time.Minute

	logrus.Infof("Waiting for volume %s to become available", id)

	e := util.WaitForWithContext(
		ctx,
		minTimeout, maxTimeout, defaultTimeout, // timeouts
		250*time.Millisecond, // period
		func() (bool, error) {
			var err error
			// Get the latest status from the volume
			v, err = util.VolumeFromName(s.driver(ctx), id)
			if err != nil {
				return false, status.Errorf(codes.Internal, err.Error())
			}

			// Check if the volume is ready
			if v.GetStatus() == api.VolumeStatus_VOLUME_STATUS_UP {
				return false, nil
			}

			// Continue waiting
			return true, nil
		})

	return v, e
}

func (s *VolumeServer) waitForVolumeRemoved(ctx context.Context, id string) error {
	minTimeout := 1 * time.Second
	maxTimeout := 10 * time.Minute
	defaultTimeout := 5 * time.Minute

	logrus.Infof("Waiting for volume %s to be removed", id)

	return util.WaitForWithContext(
		ctx,
		minTimeout, maxTimeout, defaultTimeout, // timeouts
		250*time.Millisecond, // period
		func() (bool, error) {
			// Get the latest status from the volume
			if _, err := util.VolumeFromName(s.driver(ctx), id); err != nil {
				// Removed
				return false, nil
			}

			// Continue waiting
			return true, nil
		})
}

func (s *VolumeServer) create(
	ctx context.Context,
	locator *api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec,
) (string, error) {

	// Check if the volume has already been created or is in process of creation
	volName := locator.GetName()
	v, err := util.VolumeFromName(s.driver(ctx), volName)
	// If the volume is still there but it is being delete, then wait until it is removed
	if err == nil && v.GetState() == api.VolumeState_VOLUME_STATE_DELETED {
		if err = s.waitForVolumeRemoved(ctx, volName); err != nil {
			return "", status.Errorf(codes.Internal, "Volume with same name %s is in the process of being deleted. Timed out waiting for deletion to complete: %v", volName, err)
		}

		// If the volume is there but it is not being deleted then just return the current id
	} else if err == nil {
		// Check ownership
		if !v.IsPermitted(ctx, api.Ownership_Admin) {
			return "", status.Errorf(codes.PermissionDenied, "Volume %s already exists and is owned by another user", volName)
		}

		// Wait until ready
		v, err = s.waitForVolumeReady(ctx, volName)
		if err != nil {
			return "", status.Errorf(codes.Internal, "Timed out waiting for volume %s to be in ready state: %v", volName, err)
		}

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
		parent, err := util.VolumeFromName(s.driver(ctx), source.Parent)
		if err != nil {
			return "", status.Errorf(
				codes.NotFound,
				"unable to get parent volume information: %s",
				err.Error())
		}

		// Check ownership
		// Snapshots just need read access
		if !parent.IsPermitted(ctx, api.Ownership_Read) {
			return "", status.Errorf(codes.PermissionDenied, "Access denied to volume %s", parent.GetId())
		}

		// Create a snapshot from the parent
		id, err = s.driver(ctx).Snapshot(parent.GetId(), false, &api.VolumeLocator{
			Name: volName,
		}, false)
		if err != nil {
			return "", status.Errorf(
				codes.Internal,
				"unable to create snapshot: %s",
				err.Error())
		}

		// If this is a different owner, make adjust the clone to this owner
		clone, err := s.Inspect(ctx, &api.SdkVolumeInspectRequest{
			VolumeId: id,
		})
		if err != nil {
			return "", err
		}

		newOwnership, updateNeeded := clone.Volume.Spec.GetCloneCreatorOwnership(ctx)
		if updateNeeded {
			// Set no authentication so that we can override the ownership
			ctxNoAuth := context.Background()

			// New owner for the snapshot, let's make the change
			_, err := s.Update(ctxNoAuth, &api.SdkVolumeUpdateRequest{
				VolumeId: id,
				Spec: &api.VolumeSpecUpdate{
					Ownership: newOwnership,
				},
			})
			if err != nil {
				return "", err
			}
		}
	} else {
		// New volume, set ownership
		spec.Ownership = api.OwnershipSetUsernameFromContext(ctx, spec.Ownership)
		// Create the volume
		id, err = s.driver(ctx).Create(locator, source, spec)
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
	if s.cluster() == nil || s.driver(ctx) == nil {
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

	locator := &api.VolumeLocator{
		Name:         req.GetName(),
		VolumeLabels: req.GetLabels(),
	}
	source := &api.Source{}

	// Validate/Update given spec according to default storage policy set
	// In case policy is not set, should fall back to default way
	// of creating volume
	spec, err := GetDefaultVolSpecs(ctx, req.GetSpec(), false)
	if err != nil {
		return nil, err
	}

	// Copy any labels from the spec to the locator
	locator = locator.MergeVolumeSpecLabels(spec)

	// Convert node IP to ID if necessary for API calls
	if err := s.updateReplicaSpecNodeIPstoIds(spec.GetReplicaSet()); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get replicat set information: %v", err)
	}

	// Create volume
	id, err := s.create(ctx, locator, source, spec)
	if err != nil {
		return nil, err
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
	if s.cluster() == nil || s.driver(ctx) == nil {
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
	// This will also check for Ownership_Read access.
	parentVol, err := s.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: req.GetParentId(),
	})
	if err != nil {
		return nil, err
	}

	// Create the clone
	id, err := s.create(ctx, locator, source, parentVol.GetVolume().GetSpec())
	if err != nil {
		return nil, err
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
	if s.cluster() == nil || s.driver(ctx) == nil {
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
		if IsErrorNotFound(err) {
			return &api.SdkVolumeDeleteResponse{}, nil
		}
		return nil, err
	}
	vol := resp.GetVolume()

	// Only the owner or the admin can delete
	if !vol.IsPermitted(ctx, api.Ownership_Admin) {
		return nil, status.Errorf(codes.PermissionDenied, "Cannot delete volume %v", vol.GetId())
	}

	// Delete the volume
	err = s.driver(ctx).Delete(req.GetVolumeId())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to delete volume %s: %v",
			req.GetVolumeId(),
			err.Error())
	}

	return &api.SdkVolumeDeleteResponse{}, nil
}

// InspectWithFilters is a helper function returning information about volumes which match a filter
func (s *VolumeServer) InspectWithFilters(
	ctx context.Context,
	req *api.SdkVolumeInspectWithFiltersRequest,
) (*api.SdkVolumeInspectWithFiltersResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
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

	enumVols, err := s.driver(ctx).Enumerate(locator, nil)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to enumerate volumes: %v",
			err.Error())
	}

	vols := make([]*api.SdkVolumeInspectResponse, 0, len(enumVols))
	for _, vol := range enumVols {
		// Check access
		if vol.IsPermitted(ctx, api.Ownership_Read) {

			// Check if the caller wants more information
			if req.GetOptions().GetDeep() {
				resp, err := s.Inspect(ctx, &api.SdkVolumeInspectRequest{
					VolumeId: vol.GetId(),
					Options:  req.GetOptions(),
				})
				if IsErrorNotFound(err) {
					continue
				} else if err != nil {
					return nil, err
				}
				vols = append(vols, resp)
			} else {
				// Caller does not require a deep inspect
				// Add the object now
				vols = append(vols, &api.SdkVolumeInspectResponse{
					Volume: vol,
					Name:   vol.GetLocator().GetName(),
					Labels: vol.GetLocator().GetVolumeLabels(),
				})
			}
		}
	}

	return &api.SdkVolumeInspectWithFiltersResponse{
		Volumes: vols,
	}, nil
}

// Inspect returns information about a volume
func (s *VolumeServer) Inspect(
	ctx context.Context,
	req *api.SdkVolumeInspectRequest,
) (*api.SdkVolumeInspectResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	var v *api.Volume
	if !req.GetOptions().GetDeep() {
		vols, err := s.driver(ctx).Enumerate(&api.VolumeLocator{
			VolumeIds: []string{req.GetVolumeId()},
		}, nil)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"Failed to inspect volume %s: %v",
				req.GetVolumeId(), err)
		}
		if len(vols) == 0 {
			return nil, status.Errorf(
				codes.NotFound,
				"Volume id %s not found",
				req.GetVolumeId())
		}
		v = vols[0]
	} else {
		vols, err := s.driver(ctx).Inspect([]string{req.GetVolumeId()})
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
		v = vols[0]
	}

	// Check ownership
	if !v.IsPermitted(ctx, api.Ownership_Read) {
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
	if s.cluster() == nil || s.driver(ctx) == nil {
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
	if s.cluster() == nil || s.driver(ctx) == nil {
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

	vols, err := s.driver(ctx).Enumerate(locator, nil)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to enumerate volumes: %v",
			err.Error())
	}

	ids := make([]string, 0)
	for _, vol := range vols {
		// Check access
		if vol.IsPermitted(ctx, api.Ownership_Read) {
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
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	// Get current state
	// This checks for Read access in ownership
	resp, err := s.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: req.GetVolumeId(),
	})
	if err != nil {
		return nil, err
	}

	// Check if the caller can update the volume
	if !resp.GetVolume().IsPermitted(ctx, api.Ownership_Write) {
		return nil, status.Errorf(codes.PermissionDenied, "Cannot update volume")
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

	// Validate/Update given spec according to default storage policy set
	// to make sure if update does not violates default policy
	updatedSpec, err := GetDefaultVolSpecs(ctx, spec, true)
	if err != nil {
		return nil, err
	}
	// Send to driver
	if err := s.driver(ctx).Set(req.GetVolumeId(), locator, updatedSpec); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update volume: %v", err)
	}

	return &api.SdkVolumeUpdateResponse{}, nil
}

// Stats returns volume statistics
func (s *VolumeServer) Stats(
	ctx context.Context,
	req *api.SdkVolumeStatsRequest,
) (*api.SdkVolumeStatsResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	// Get access rights
	if err := s.checkAccessForVolumeId(ctx, req.GetVolumeId(), api.Ownership_Read); err != nil {
		return nil, err
	}

	stats, err := s.driver(ctx).Stats(req.GetVolumeId(), !req.GetNotCumulative())
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
	if err := s.checkAccessForVolumeId(ctx, req.GetVolumeId(), api.Ownership_Read); err != nil {
		return nil, err
	}

	dResp, err := s.driver(ctx).CapacityUsage(req.GetVolumeId())
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
	spec.Nodiscard = setSpecBool(vol.GetNodiscard(), req.GetNodiscard(), req.GetNodiscardOpt())

	// fastpath extensions
	if req.GetFastpathOpt() != nil {
		spec.FpPreference = req.GetFastpath()
	} else {
		spec.FpPreference = vol.GetFpPreference()
	}

	if req.GetIoStrategy() != nil {
		spec.IoStrategy = req.GetIoStrategy()
	} else {
		spec.IoStrategy = vol.GetIoStrategy()
	}

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

	// ExportSpec
	if req.GetExportSpec() != nil {
		spec.ExportSpec = req.GetExportSpec()
	} else {
		spec.ExportSpec = vol.GetExportSpec()
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

// GetDefaultVolSpecs returns volume spec merged with default storage policy applied if any
func GetDefaultVolSpecs(
	ctx context.Context,
	spec *api.VolumeSpec,
	isUpdate bool,
) (*api.VolumeSpec, error) {
	storPolicy, err := policy.Inst()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to get storage policy instance %v", err)
	}

	var policy *api.SdkStoragePolicy
	// check if custom policy passed with volume
	if spec.GetStoragePolicy() != "" {
		inspReq := &api.SdkOpenStoragePolicyInspectRequest{
			// name of storage policy specified in volSpecs
			Name: spec.GetStoragePolicy(),
		}
		// inspect will make sure user will atleast have read access
		customPolicy, customErr := storPolicy.Inspect(ctx, inspReq)
		if customErr != nil {
			return nil, customErr
		}

		policy = customPolicy.GetStoragePolicy()
	} else {
		// check if default storage policy is set
		defPolicy, err := storPolicy.DefaultInspect(context.Background(), &api.SdkOpenStoragePolicyDefaultInspectRequest{})
		if err != nil {
			// err means there is policy stored, but we are not able to retrive it
			// hence we are not allowing volume create operation
			return nil, status.Errorf(codes.Internal, "Unable to get default policy details %v", err)
		} else if defPolicy.GetStoragePolicy() == nil {
			// no default storage policy found
			return spec, nil
		}
		policy = defPolicy.GetStoragePolicy()
	}

	// track volume created using storage policy
	spec.StoragePolicy = policy.GetName()
	// check if volume update request, if allowupdate is set
	// return spec received as it is
	if isUpdate && policy.GetAllowUpdate() {
		if !policy.IsPermitted(ctx, api.Ownership_Write) {
			return nil, status.Errorf(codes.PermissionDenied, "Cannot use storage policy %v", policy.GetName())
		}
		return spec, nil
	}

	return mergeVolumeSpecsPolicy(spec, policy.GetPolicy(), policy.GetForce())
}

func mergeVolumeSpecsPolicy(vol *api.VolumeSpec, req *api.VolumeSpecPolicy, isValidate bool) (*api.VolumeSpec, error) {
	errMsg := fmt.Errorf("Storage Policy Violation, valid specs are : %v", req.String())
	spec := vol
	// Shared
	if req.GetSharedOpt() != nil {
		if isValidate && vol.GetShared() != req.GetShared() {
			return nil, errMsg
		}
		spec.Shared = req.GetShared()
	}
	//sharedv4
	if req.GetSharedv4Opt() != nil {
		if isValidate && vol.GetSharedv4() != req.GetSharedv4() {
			return vol, errMsg
		}
		spec.Sharedv4 = req.GetSharedv4()
	}
	//sticky
	if req.GetStickyOpt() != nil {
		if isValidate && vol.GetSticky() != req.GetSticky() {
			return vol, errMsg
		}
		spec.Sticky = req.GetSticky()
	}
	//journal
	if req.GetJournalOpt() != nil {
		if isValidate && vol.GetJournal() != req.GetJournal() {
			return vol, errMsg
		}
		spec.Journal = req.GetJournal()
	}
	// encrypt
	if req.GetEncryptedOpt() != nil {
		if isValidate && vol.GetEncrypted() != req.GetEncrypted() {
			return vol, errMsg
		}
		spec.Encrypted = req.GetEncrypted()
	}
	// cos level
	if req.GetCosOpt() != nil {
		if isValidate && vol.GetCos() != req.GetCos() {
			return vol, errMsg
		}
		spec.Cos = req.GetCos()
	}
	// passphrase
	if req.GetPassphraseOpt() != nil {
		if isValidate && vol.GetPassphrase() != req.GetPassphrase() {
			return vol, errMsg
		}
		spec.Passphrase = req.GetPassphrase()
	}
	// IO profile
	if req.GetIoProfileOpt() != nil {
		if isValidate && req.GetIoProfile() != vol.GetIoProfile() {
			return vol, errMsg
		}
		spec.IoProfile = req.GetIoProfile()
	}
	// Group
	if req.GetGroupOpt() != nil {
		if isValidate && req.GetGroup() != vol.GetGroup() {
			return vol, errMsg
		}
		spec.Group = req.GetGroup()
	}
	// Replicaset
	if req.GetReplicaSet() != nil {
		if isValidate && req.GetReplicaSet() != vol.GetReplicaSet() {
			return vol, errMsg
		}
		spec.ReplicaSet = req.GetReplicaSet()
	}
	// QueueDepth
	if req.GetQueueDepthOpt() != nil {
		if isValidate && req.GetQueueDepth() != vol.GetQueueDepth() {
			return vol, errMsg
		}
		spec.QueueDepth = req.GetQueueDepth()
	}
	// SnapshotSchedule
	if req.GetSnapshotScheduleOpt() != nil {
		if isValidate && req.GetSnapshotSchedule() != vol.GetSnapshotSchedule() {
			return vol, errMsg
		}
		spec.SnapshotSchedule = req.GetSnapshotSchedule()
	}
	// aggr level
	if req.GetAggregationLevelOpt() != nil {
		if isValidate && req.GetAggregationLevel() != vol.GetAggregationLevel() {
			return vol, errMsg
		}
		spec.AggregationLevel = req.GetAggregationLevel()
	}

	// Size
	if req.GetSizeOpt() != nil {
		isCorrect := validateMinMaxParams(uint64(req.GetSize()),
			uint64(vol.Size), req.GetSizeOperator())
		if !isCorrect {
			if isValidate {
				return vol, errMsg
			}
			spec.Size = req.GetSize()
		}
	}

	// HA Level
	if req.GetHaLevelOpt() != nil {
		isCorrect := validateMinMaxParams(uint64(req.GetHaLevel()),
			uint64(vol.HaLevel), req.GetHaLevelOperator())
		if !isCorrect {
			if isValidate {
				return vol, errMsg
			}
			spec.HaLevel = req.GetHaLevel()
		}
	}

	// Scale
	if req.GetScaleOpt() != nil {
		isCorrect := validateMinMaxParams(uint64(req.GetScale()),
			uint64(vol.Scale), req.GetScaleOperator())
		if !isCorrect {
			if isValidate {
				return vol, errMsg
			}

			spec.Scale = req.GetScale()
		}
	}

	// Snapshot Interval
	if req.GetSnapshotIntervalOpt() != nil {
		isCorrect := validateMinMaxParams(uint64(req.GetSnapshotInterval()),
			uint64(vol.SnapshotInterval), req.GetSnapshotIntervalOperator())
		if !isCorrect {
			if isValidate {
				return vol, errMsg
			}
			spec.SnapshotInterval = req.GetSnapshotInterval()
		}
	}

	// Nodiscard
	if req.GetNodiscardOpt() != nil {
		if isValidate && vol.GetNodiscard() != req.GetNodiscard() {
			return vol, errMsg
		}
		spec.Nodiscard = req.GetNodiscard()
	}

	// IoStrategy
	if req.GetIoStrategy() != nil {
		if isValidate && vol.GetIoStrategy() != req.GetIoStrategy() {
			return vol, errMsg
		}
		spec.IoStrategy = req.GetIoStrategy()
	}

	// ExportSpec
	if req.GetExportSpec() != nil {
		if isValidate && vol.GetExportSpec() != req.GetExportSpec() {
			return vol, errMsg
		}
		if exportPolicy := vol.GetExportSpec(); exportPolicy == nil {
			spec.ExportSpec = req.GetExportSpec()
		} else {
			// If the spec has an ExportSpec then only modify the fields that came in
			// the request.
			reqExportSpec := req.GetExportSpec()
			if reqExportSpec.ExportProtocol != api.ExportProtocol_INVALID {
				spec.ExportSpec.ExportProtocol = reqExportSpec.ExportProtocol
			}
			if len(reqExportSpec.ExportOptions) != 0 {
				if reqExportSpec.ExportOptions == api.SpecExportOptionsEmpty {
					spec.ExportSpec.ExportOptions = ""
				} else {
					spec.ExportSpec.ExportOptions = reqExportSpec.ExportOptions
				}
			}
		}
	}
	logrus.Debugf("Updated VolumeSpecs %v", spec)
	return spec, nil
}

func validateMinMaxParams(policy uint64, specified uint64, op api.VolumeSpecPolicy_PolicyOp) bool {
	switch op {
	case api.VolumeSpecPolicy_Maximum:
		if specified > policy {
			return false
		}
	case api.VolumeSpecPolicy_Minimum:
		if specified < policy {
			return false
		}
	default:
		if specified != policy {
			return false
		}
	}
	return true
}
