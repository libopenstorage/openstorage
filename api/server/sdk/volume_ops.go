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
	"github.com/libopenstorage/openstorage/pkg/util"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/portworx/kvdb"
)

func (s *VolumeServer) create(
	locator *api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec,
) (string, error) {

	// Check if the volume has already been created or is in process of creation
	volName := locator.GetName()
	v, err := util.VolumeFromName(s.driver, volName)
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
		parent, err := util.VolumeFromName(s.driver, source.Parent)
		if err != nil {
			return "", status.Errorf(
				codes.InvalidArgument,
				"unable to get parent volume information: %s",
				err.Error())
		}

		// Create a snapshot from the parent
		id, err = s.driver.Snapshot(parent.GetId(), false, &api.VolumeLocator{
			Name: volName,
		})
		if err != nil {
			return "", status.Errorf(
				codes.Internal,
				"unable to create snapshot: %s\n",
				err.Error())
		}
	} else {
		// Create the volume
		id, err = s.driver.Create(locator, source, spec)
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

	if len(req.GetName()) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"Must supply a uniqe name")
	} else if req.GetSpec() == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"Must supply spec object")
	}

	spec := req.GetSpec()
	locator := &api.VolumeLocator{
		Name: req.GetName(),
	}
	source := &api.Source{}

	id, err := s.create(locator, source, spec)
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
	parentVol, err := s.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: req.GetParentId(),
	})
	if err != nil {
		return nil, err
	}

	// Create the clone
	id, err := s.create(locator, source, parentVol.GetVolume().GetSpec())
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

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	// If the volume is not found, return OK to be idempotent
	volumes, err := s.driver.Inspect([]string{req.GetVolumeId()})
	if (err == nil && len(volumes) == 0) ||
		(err != nil && err == volume.ErrEnoEnt) {
		return &api.SdkVolumeDeleteResponse{}, nil
	} else if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to determine if volume id %s exists: %v",
			req.GetVolumeId(),
			err.Error())
	}

	err = s.driver.Delete(req.GetVolumeId())
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

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	vols, err := s.driver.Inspect([]string{req.GetVolumeId()})
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

	return &api.SdkVolumeInspectResponse{
		Volume: vols[0],
	}, nil
}

// Enumerate returns a list of volumes
func (s *VolumeServer) Enumerate(
	ctx context.Context,
	req *api.SdkVolumeEnumerateRequest,
) (*api.SdkVolumeEnumerateResponse, error) {

	vols, err := s.driver.Enumerate(req.GetLocator(), nil)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to enumerate volumes: %v",
			err.Error())
	}

	ids := make([]string, len(vols))
	for i, vol := range vols {
		ids[i] = vol.GetId()
	}

	return &api.SdkVolumeEnumerateResponse{
		VolumeIds: ids,
	}, nil
}

// Update allows the caller to change values in the volume specification
func (s *VolumeServer) Update(
	ctx context.Context,
	req *api.SdkVolumeUpdateRequest,
) (*api.SdkVolumeUpdateResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	// Get current state
	resp, err := s.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: req.GetVolumeId(),
	})
	if err != nil {
		return nil, err
	}
	spec := s.mergeVolumeSpecs(resp.GetVolume().GetSpec(), req.GetSpec())

	// Send to driver
	if err := s.driver.Set(req.GetVolumeId(), req.GetLocator(), spec); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update volume: %v", err)
	}

	return &api.SdkVolumeUpdateResponse{}, nil
}

func (s *VolumeServer) mergeVolumeSpecs(vol *api.VolumeSpec, req *api.VolumeSpecUpdate) *api.VolumeSpec {

	spec := &api.VolumeSpec{}
	spec.Shared = setSpecBool(vol.GetShared(), req.GetShared(), req.GetSharedOpt())
	spec.Sharedv4 = setSpecBool(vol.GetSharedv4(), req.GetSharedv4(), req.GetSharedv4Opt())
	spec.Sticky = setSpecBool(vol.GetSticky(), req.GetSticky(), req.GetStickyOpt())
	spec.Compressed = setSpecBool(vol.GetCompressed(), req.GetCompressed(), req.GetCompressedOpt())
	spec.GroupEnforced = setSpecBool(vol.GetGroupEnforced(), req.GetGroupEnforced(), req.GetGroupEnforcedOpt())
	spec.Ephemeral = setSpecBool(vol.GetEphemeral(), req.GetEphemeral(), req.GetEphemeralOpt())
	spec.Encrypted = setSpecBool(vol.GetEncrypted(), req.GetEncrypted(), req.GetEncryptedOpt())
	spec.Cascaded = setSpecBool(vol.GetCascaded(), req.GetCascaded(), req.GetCascadedOpt())
	spec.Journal = setSpecBool(vol.GetJournal(), req.GetJournal(), req.GetJournalOpt())

	// Cos
	if req.GetCosOpt() != nil {
		spec.Cos = req.GetCos()
	} else {
		spec.Cos = vol.GetCos()
	}

	// Volume configuration labels
	// If none are provided, send `nil` to the driver
	spec.VolumeLabels = req.GetVolumeLabels()

	// Aggregation Level
	if req.GetAggregationLevelOpt() != nil {
		spec.AggregationLevel = req.GetAggregationLevel()
	} else {
		spec.AggregationLevel = vol.GetAggregationLevel()
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

	// FSType / format
	if req.GetFormatOpt() != nil {
		spec.Format = req.GetFormat()
	} else {
		spec.Format = vol.GetFormat()
	}

	// GroupID
	if req.GetGroupOpt() != nil {
		spec.Group = req.GetGroup()
	} else {
		spec.Group = vol.GetGroup()
	}

	// Group Enforced
	if req.GetGroupEnforcedOpt() != nil {
		spec.GroupEnforced = req.GetGroupEnforced()
	} else {
		spec.GroupEnforced = vol.GetGroupEnforced()
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

	return spec
}

func setSpecBool(current, req bool, reqSet interface{}) bool {
	if reqSet != nil {
		return req
	}
	return current
}
