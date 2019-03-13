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
	}

	devPath, err := s.driver(ctx).Attach(req.GetVolumeId(), options)
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
	options[mountattachoptions.OptionsRedirectDetach] = "true"
	if req.GetOptions() != nil {
		options[mountattachoptions.OptionsForceDetach] = fmt.Sprint(req.GetOptions().GetForce())
		options[mountattachoptions.OptionsUnmountBeforeDetach] = fmt.Sprint(req.GetOptions().GetUnmountBeforeDetach())
	}
	err := s.driver(ctx).Detach(req.GetVolumeId(), options)
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

	if s.cluster() == nil || s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}
	if len(req.GetMountPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Mount Path")
	}

	// Get volume information
	resp, err := s.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: req.GetVolumeId(),
	})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}
	vol := resp.GetVolume()
	mountpoint := req.GetMountPath()
	name := vol.GetLocator().GetName()

	// Checks for ownership
	if !vol.IsPermitted(ctx, api.Ownership_Write) {
		return nil, status.Errorf(codes.PermissionDenied, "Access denied to volume %s", vol.GetId())
	}

	if vol.GetSpec().GetScale() > 1 {
		id := s.driver(ctx).MountedAt(mountpoint)
		if len(id) != 0 {
			err = s.driver(ctx).Unmount(id, mountpoint, nil)
			if err != nil {
				return nil, status.Errorf(codes.Internal,
					"Failed to prepare scaled volume by unmounting it: %v. "+
						"Cannot remount scaled volume(%v). "+
						"Volume %v is mounted at %v",
					err,
					name,
					id,
					mountpoint)
			}

			if s.driver(ctx).Type() == api.DriverType_DRIVER_TYPE_BLOCK {
				err = s.driver(ctx).Detach(id, nil)
				if err != nil {
					_ = s.driver(ctx).Mount(id, mountpoint, nil)
					return nil, status.Errorf(codes.Internal,
						"Failed to mount scaled volume: %v. "+
							"Cannot remount scaled volume(%v). "+
							"Volume %v is mounted at %v",
						err,
						name,
						id,
						mountpoint)
				}
			}
		}
	}

	// If this is a block driver, first attach the volume.
	if s.driver(ctx).Type() == api.DriverType_DRIVER_TYPE_BLOCK {
		// If volume is scaled up, a new volume is created and
		// vol will change.
		attachOptions := req.GetDriverOptions()
		if attachOptions == nil {
			attachOptions = make(map[string]string)
		}

		if req.Options != nil {
			attachOptions[mountattachoptions.OptionsSecret] = req.Options.SecretName
			attachOptions[mountattachoptions.OptionsSecretKey] = req.Options.SecretKey
			attachOptions[mountattachoptions.OptionsSecretContext] = req.Options.SecretContext
		}

		if vol.Scaled() {
			vol, err = s.attachScale(ctx, vol, attachOptions)
		} else {
			vol, err = s.attachVol(ctx, vol, attachOptions)
		}
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err = s.driver(ctx).Mount(req.GetVolumeId(), req.GetMountPath(), req.GetDriverOptions())
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

	// Get volume to unmount
	// Checks ownership
	resp, err := s.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: req.GetVolumeId(),
	})
	if err != nil {
		return nil, err
	}
	vol := resp.GetVolume()
	volid := resp.GetVolume().GetId()

	// Checks for ownership
	if !vol.IsPermitted(ctx, api.Ownership_Write) {
		return nil, status.Errorf(codes.PermissionDenied, "Access denied to volume %s", vol.GetId())
	}

	// From old docker server, now it is here in the SDK
	if resp.GetVolume().GetSpec().Scale > 1 {
		volid := s.driver(ctx).MountedAt(req.GetMountPath())
		if len(volid) == 0 {
			return nil, status.Errorf(codes.Internal, "Failed to find volume mapping for %v", req.GetMountPath())
		}
	}

	if err = s.driver(ctx).Unmount(volid, req.GetMountPath(), options); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to unmount volume %s: %v",
			req.GetVolumeId(),
			err.Error())
	}

	return &api.SdkVolumeUnmountResponse{}, nil
}

func (s *VolumeServer) scaleUp(
	ctx context.Context,
	inVol *api.Volume,
	allVols []*api.Volume,
	attachOptions map[string]string,
) (
	outVol *api.Volume,
	err error,
) {
	// Create new volume if existing volumes are not available.
	spec := inVol.Spec.Copy()
	spec.Scale = 1
	spec.ReplicaSet = nil
	volCount := len(allVols)
	for i := len(allVols); volCount < int(inVol.Spec.Scale); i++ {
		name := fmt.Sprintf("%s_%03d", inVol.Locator.Name, i)
		id := ""

		// create, get vol from name, attach
		if id, err = s.driver(ctx).Create(
			&api.VolumeLocator{Name: name},
			nil,
			spec,
		); err != nil {
			// It is possible to get an error on a name conflict
			// either due to concurrent creates or holes punched in
			// from previous deletes.
			if err == volume.ErrExist {
				continue
			}
			return nil, err
		}

		outVol, err := s.volFromId(ctx, id)
		if err != nil {
			return nil, err
		}
		if _, err = s.driver(ctx).Attach(outVol.Id, attachOptions); err == nil {
			return outVol, nil
		}
		// If we fail to attach the volume, continue to look for a
		// free volume.
		volCount++
	}
	return nil, volume.ErrVolAttachedScale
}

func (s *VolumeServer) attachScale(
	ctx context.Context,
	inVol *api.Volume,
	attachOptions map[string]string,
) (
	*api.Volume,
	error,
) {
	// Find a volume that has data local to this node.
	volumes, err := s.driver(ctx).Enumerate(&api.VolumeLocator{
		Name: fmt.Sprintf("%s.*", inVol.Locator.Name),
		VolumeLabels: map[string]string{
			volume.LocationConstraint: volume.LocalNode,
		},
	}, nil)

	// Try to attach local volumes.
	if err == nil {
		for _, vol := range volumes {
			if v, err := s.attachVol(ctx, vol, attachOptions); err == nil {
				return v, nil
			}
		}
	}
	// Create a new local volume if we fail to attach existing local volume
	// or if none exist.
	allVols, err := s.driver(ctx).Enumerate(
		&api.VolumeLocator{
			Name: fmt.Sprintf("%s.*", inVol.Locator.Name),
		},
		nil,
	)

	// Try to attach existing volumes.
	for _, outVol := range allVols {
		if _, err = s.driver(ctx).Attach(outVol.Id, attachOptions); err == nil {
			return outVol, nil
		}
	}

	if len(allVols) < int(inVol.Spec.Scale) {
		name := fmt.Sprintf("%s_%03d", inVol.Locator.Name, len(allVols))
		spec := inVol.Spec.Copy()
		spec.ReplicaSet = &api.ReplicaSet{Nodes: []string{volume.LocalNode}}
		spec.Scale = 1

		// create, vol from name, attach
		id, err := s.driver(ctx).Create(&api.VolumeLocator{Name: name}, nil, spec)
		if err != nil {
			return s.scaleUp(ctx, inVol, allVols, attachOptions)
		}

		outVol, err := s.volFromId(ctx, id)
		if err != nil {
			return nil, err
		}
		if _, err = s.driver(ctx).Attach(outVol.Id, attachOptions); err == nil {
			return outVol, nil
		}

		// We failed to attach, scaleUp.
		allVols = append(allVols, outVol)
	}
	return s.scaleUp(ctx, inVol, allVols, attachOptions)
}

func (s *VolumeServer) attachVol(
	ctx context.Context,
	vol *api.Volume,
	attachOptions map[string]string,
) (
	outVolume *api.Volume,
	err error,
) {
	_, err = s.driver(ctx).Attach(vol.Id, attachOptions)

	switch err {
	case nil:
		return vol, nil
	case volume.ErrVolAttachedOnRemoteNode:
		return vol, status.Errorf(
			codes.Internal,
			"Failed to attach volume %s: %v",
			vol.Id,
			err.Error())
	default:
		return vol, status.Errorf(
			codes.Internal,
			"Failed to attach volume %s: %v",
			vol.Id,
			err.Error())

	}
}

func (s *VolumeServer) volFromName(ctx context.Context, name string) (*api.Volume, error) {
	vols, err := s.driver(ctx).Enumerate(&api.VolumeLocator{Name: name}, nil)
	if err != nil || len(vols) <= 0 {
		return nil, fmt.Errorf("Cannot locate volume with name %s", name)
	}
	return vols[0], nil
}

func (s *VolumeServer) volFromId(ctx context.Context, volId string) (*api.Volume, error) {
	vols, err := s.driver(ctx).Inspect([]string{volId})
	if err != nil || len(vols) <= 0 {
		return nil, fmt.Errorf("Cannot locate volume with id %s", volId)
	}
	return vols[0], nil
}
