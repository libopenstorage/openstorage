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

// CreateFromVolumeID creates a new volume from an existing volume
func (s *VolumeServer) CreateFromVolumeId(
	ctx context.Context,
	req *api.SdkVolumeCreateFromVolumeIdRequest,
) (*api.SdkVolumeCreateFromVolumeIdResponse, error) {

	if len(req.GetName()) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"Must supply a uniqe name")
	} else if len(req.GetParentId()) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"Must parent volume id")
	} else if req.GetSpec() == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"Must supply spec object")
	}

	spec := req.GetSpec()
	locator := &api.VolumeLocator{
		Name: req.GetName(),
	}
	source := &api.Source{
		Parent: req.GetParentId(),
	}

	id, err := s.create(locator, source, spec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.SdkVolumeCreateFromVolumeIdResponse{
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
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to inspect volume %s: %v",
			req.GetVolumeId(),
			err.Error())
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

	return &api.SdkVolumeEnumerateResponse{
		Volumes: vols,
	}, nil
}
