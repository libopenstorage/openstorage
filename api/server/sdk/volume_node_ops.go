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

	"github.com/libopenstorage/openstorage/api"
	mountattachoptions "github.com/libopenstorage/openstorage/pkg/options"
	"github.com/libopenstorage/openstorage/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Attach volume to given node
func (s *VolumeServer) Attach(
	ctx context.Context,
	req *api.SdkVolumeAttachRequest,
) (*api.SdkVolumeAttachResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	// Check if already attached
	v, err := util.VolumeFromName(s.driver, req.GetVolumeId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s was not found", req.GetVolumeId())
	}

	// Idempotency
	if v.GetState() == api.VolumeState_VOLUME_STATE_ATTACHED &&
		v.GetAttachedState() != api.AttachState_ATTACH_STATE_INTERNAL {
		return &api.SdkVolumeAttachResponse{DevicePath: v.GetDevicePath()}, nil
	}

	// Check options
	options := make(map[string]string)
	if req.GetOptions() != nil {
		if len(req.GetOptions().GetSecretContext()) != 0 {
			options[mountattachoptions.OptionsSecretContext] = req.GetOptions().GetSecretContext()
		}
		if len(req.GetOptions().GetSecretKey()) != 0 {
			options[mountattachoptions.OptionsSecretKey] = req.GetOptions().GetSecretKey()
		}
		if len(req.GetOptions().GetSecretName()) != 0 {
			options[mountattachoptions.OptionsSecret] = req.GetOptions().GetSecretName()
		}
	}

	devPath, err := s.driver.Attach(req.GetVolumeId(), options)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed  to attach volume: %v",
			err.Error())
	}

	return &api.SdkVolumeAttachResponse{DevicePath: devPath}, nil
}

// Detach function for volume node detach
func (s *VolumeServer) Detach(
	ctx context.Context,
	req *api.SdkVolumeDetachRequest,
) (*api.SdkVolumeDetachResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	// Check if already attached
	v, err := util.VolumeFromName(s.driver, req.GetVolumeId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s was not found", req.GetVolumeId())
	}

	// Idempotency
	if v.GetState() == api.VolumeState_VOLUME_STATE_DETACHED ||
		(v.GetState() == api.VolumeState_VOLUME_STATE_ATTACHED && v.GetAttachedState() == api.AttachState_ATTACH_STATE_INTERNAL) {
		return &api.SdkVolumeDetachResponse{}, nil
	}

	// Check options
	options := make(map[string]string)
	options[mountattachoptions.OptionsRedirectDetach] = "true"
	if req.GetOptions() != nil {
		options[mountattachoptions.OptionsForceDetach] = fmt.Sprint(req.GetOptions().GetForce())
		options[mountattachoptions.OptionsUnmountBeforeDetach] = fmt.Sprint(req.GetOptions().GetUnmountBeforeDetach())
	}
	err = s.driver.Detach(req.GetVolumeId(), options)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to detach volume %s: %v",
			req.GetVolumeId(),
			err)
	}

	return &api.SdkVolumeDetachResponse{}, nil
}

// Mount function for volume node detach
func (s *VolumeServer) Mount(
	ctx context.Context,
	req *api.SdkVolumeMountRequest,
) (*api.SdkVolumeMountResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}
	if len(req.GetMountPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Mount Path")
	}

	err := s.driver.Mount(req.GetVolumeId(), req.GetMountPath(), nil)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to mount volume %s: %v",
			req.GetVolumeId(),
			err.Error())
	}
	return &api.SdkVolumeMountResponse{}, err
}

// Unmount volume from given node
func (s *VolumeServer) Unmount(
	ctx context.Context,
	req *api.SdkVolumeUnmountRequest,
) (*api.SdkVolumeUnmountResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	if len(req.GetMountPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Mount Path")
	}

	options := make(map[string]string)
	if req.GetOptions() != nil {
		options[mountattachoptions.OptionsDeleteAfterUnmount] = fmt.Sprint(req.GetOptions().GetDeleteMountPath())

		// Only set if
		if req.GetOptions().GetDeleteMountPath() {
			options[mountattachoptions.OptionsWaitBeforeDelete] = fmt.Sprint(!req.GetOptions().GetNoDelayBeforeDeletingMountPath())
		}
	}

	err := s.driver.Unmount(req.GetVolumeId(), req.GetMountPath(), options)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to unmount volume %s: %v",
			req.GetVolumeId(),
			err.Error())
	}

	return &api.SdkVolumeUnmountResponse{}, nil
}
