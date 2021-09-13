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
	"github.com/libopenstorage/openstorage/volume"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Attach volume to given node
func (s *VolumeServer) Attach(
	ctx context.Context,
	req *api.SdkVolumeAttachRequest,
) (*api.SdkVolumeAttachResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	// Get access rights
	if err := s.checkAccessForVolumeId(ctx, req.GetVolumeId(), api.Ownership_Write); err != nil {
		return nil, err
	}

	// Check options
	options := req.GetDriverOptions()
	if options == nil {
		options = make(map[string]string)
	}
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
		if len(req.GetOptions().GetFastpath()) != 0 {
			options[mountattachoptions.OptionsFastpath] = req.GetOptions().GetFastpath()
		}
	}

	devPath, err := s.driver(ctx).Attach(ctx, req.GetVolumeId(), options)
	if err == volume.ErrVolAttachedOnRemoteNode {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	} else if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed  to attach volume: %v",
			err.Error())
	}

	s.auditLog(ctx, "mountattach.attach", "Volume %s attached", req.GetVolumeId())
	return &api.SdkVolumeAttachResponse{DevicePath: devPath}, nil
}

// Detach function for volume node detach
func (s *VolumeServer) Detach(
	ctx context.Context,
	req *api.SdkVolumeDetachRequest,
) (*api.SdkVolumeDetachResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	// Get access rights
	err := s.checkAccessForVolumeId(ctx, req.GetVolumeId(), api.Ownership_Write)
	if err != nil && !IsErrorNotFound(err) {
		return nil, err
	}

	// Check options
	options := req.GetDriverOptions()
	if options == nil {
		options = make(map[string]string)
	}
	if req.GetOptions() != nil {
		options[mountattachoptions.OptionsForceDetach] = fmt.Sprint(req.GetOptions().GetForce())
		options[mountattachoptions.OptionsUnmountBeforeDetach] = fmt.Sprint(req.GetOptions().GetUnmountBeforeDetach())
		options[mountattachoptions.OptionsRedirectDetach] = fmt.Sprint(req.GetOptions().GetRedirect())
	}

	err = s.driver(ctx).Detach(context.TODO(), req.GetVolumeId(), options)
	if err != nil && !IsErrorNotFound(err) {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to detach volume %s: %v",
			req.GetVolumeId(),
			err)
	}

	s.auditLog(ctx, "mountattach.detach", "Volume %s detached", req.GetVolumeId())
	return &api.SdkVolumeDetachResponse{}, nil
}

// Mount function for volume node detach
func (s *VolumeServer) Mount(
	ctx context.Context,
	req *api.SdkVolumeMountRequest,
) (*api.SdkVolumeMountResponse, error) {

	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}
	if len(req.GetMountPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Mount Path")
	}

	// Get access rights
	err := s.checkAccessForVolumeId(ctx, req.GetVolumeId(), api.Ownership_Write)
	if err != nil {
		return nil, err
	}

	err = s.driver(ctx).Mount(ctx, req.GetVolumeId(), req.GetMountPath(), req.GetDriverOptions())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to mount volume %s: %v",
			req.GetVolumeId(),
			err.Error())
	}
	s.auditLog(ctx, "mountattach.mount", "Volume %s mounted", req.GetVolumeId())
	return &api.SdkVolumeMountResponse{}, err
}

// Unmount volume from given node
func (s *VolumeServer) Unmount(
	ctx context.Context,
	req *api.SdkVolumeUnmountRequest,
) (*api.SdkVolumeUnmountResponse, error) {
	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	if len(req.GetMountPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Mount Path")
	}

	options := req.GetDriverOptions()
	if options == nil {
		options = make(map[string]string)
	}
	if req.GetOptions() != nil {
		options[mountattachoptions.OptionsDeleteAfterUnmount] = fmt.Sprint(req.GetOptions().GetDeleteMountPath())

		if req.GetOptions().GetDeleteMountPath() {
			options[mountattachoptions.OptionsWaitBeforeDelete] = fmt.Sprint(!req.GetOptions().GetNoDelayBeforeDeletingMountPath())
		}
	}

	// Get access rights
	err := s.checkAccessForVolumeId(ctx, req.GetVolumeId(), api.Ownership_Write)
	if err != nil && !IsErrorNotFound(err) {
		return nil, err
	}

	// Unmount volume
	if err = s.driver(ctx).Unmount(ctx, req.GetVolumeId(), req.GetMountPath(), options); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to unmount volume %s: %v",
			req.GetVolumeId(),
			err.Error())
	}

	s.auditLog(ctx, "mountattach.unmount", "Volume %s mounted", req.GetVolumeId())
	return &api.SdkVolumeUnmountResponse{}, nil
}
