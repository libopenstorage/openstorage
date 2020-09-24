/*
Package csi is CSI driver interface for OSD
Copyright 2017 Portworx

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
package csi

import (
	"fmt"
	"os"

	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *OsdCsiServer) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest,
) (*csi.NodeGetInfoResponse, error) {

	clus, err := s.cluster.Enumerate()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to Enumerate cluster: %s", err)
	}

	result := &csi.NodeGetInfoResponse{
		NodeId: clus.NodeId,
	}

	return result, nil
}

// NodePublishVolume is a CSI API call which mounts the volume on the specified
// target path on the node.
//
// TODO: Support READ ONLY Mounts
//
func (s *OsdCsiServer) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {

	logrus.Debugf("NodePublishVolume req[%#v]", req)

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume id must be provided")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path must be provided")
	}
	if req.GetVolumeCapability() == nil || req.GetVolumeCapability().GetAccessMode() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume access mode must be provided")
	}
	if req.GetVolumeCapability().GetBlock() != nil {
		return nil, status.Errorf(codes.Unimplemented, "CSI raw block is not supported")
	}

	// Verify target location is an existing directory
	if err := verifyTargetLocation(req.GetTargetPath()); err != nil {
		return nil, status.Errorf(
			codes.Aborted,
			"Failed to use target location %s: %s",
			req.GetTargetPath(),
			err.Error())
	}

	// Get grpc connection
	conn, err := s.getConn()
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Unable to connect to SDK server: %v", err)
	}

	// Get secret if any was passed
	ctx = s.setupContextWithToken(ctx, req.GetSecrets())

	// Check if block device
	driverType := s.driver.Type()
	if driverType != api.DriverType_DRIVER_TYPE_BLOCK &&
		req.GetVolumeCapability().GetBlock() != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Trying to attach as block a non block device")
	}

	// Gather volume attributes
	spec, locator, _, err := s.specHandler.SpecFromOpts(req.GetVolumeContext())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Invalid volume attributes: %#v",
			req.GetVolumeContext())
	}

	// Get volume encryption info from req.Secrets
	driverOpts := s.addEncryptionInfoToLabels(make(map[string]string), req.GetSecrets())

	// prepare for mount/attaching
	mounts := api.NewOpenStorageMountAttachClient(conn)
	opts := &api.SdkVolumeAttachOptions{
		SecretName: spec.GetPassphrase(),
	}

	// can use either spec.Ephemeral or VolumeContext label
	volumeId := req.GetVolumeId()
	if req.GetVolumeContext()["csi.storage.k8s.io/ephemeral"] == "true" || spec.Ephemeral {
		spec.Ephemeral = true
		volumes := api.NewOpenStorageVolumeClient(conn)
		resp, err := volumes.Create(ctx, &api.SdkVolumeCreateRequest{
			Name:   req.GetVolumeId(),
			Spec:   spec,
			Labels: locator.GetVolumeLabels(),
		})
		if err != nil {
			return nil, err
		}
		volumeId = resp.VolumeId
	}

	if driverType == api.DriverType_DRIVER_TYPE_BLOCK {
		if _, err = mounts.Attach(ctx, &api.SdkVolumeAttachRequest{
			VolumeId:      volumeId,
			Options:       opts,
			DriverOptions: driverOpts,
		}); err != nil {
			if spec.Ephemeral {
				logrus.Errorf("Failed to attach ephemeral volume %s: %v", volumeId, err.Error())
				s.cleanupEphemeral(ctx, conn, volumeId, false)
			}
			return nil, err
		}
	}

	// Mount volume onto the path
	if _, err := mounts.Mount(ctx, &api.SdkVolumeMountRequest{
		VolumeId:  volumeId,
		MountPath: req.GetTargetPath(),
		Options:   opts,
	}); err != nil {
		if spec.Ephemeral {
			logrus.Errorf("Failed to mount ephemeral volume %s: %v", volumeId, err.Error())
			s.cleanupEphemeral(ctx, conn, volumeId, true)
		}
		return nil, err
	}

	logrus.Infof("Volume %s mounted on %s",
		volumeId,
		req.GetTargetPath())

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume is a CSI API call which unmounts the volume.
func (s *OsdCsiServer) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {

	logrus.Debugf("NodeUnPublishVolume req[%#v]", req)

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume id must be provided")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path must be provided")
	}

	// Get volume information
	vols, err := s.driver.Inspect([]string{req.GetVolumeId()})
	if err != nil || len(vols) < 1 {
		if err == kvdb.ErrNotFound {
			logrus.Infof("Volume %s was deleted or cannot be found: %s", req.GetVolumeId(), err.Error())
			return &csi.NodeUnpublishVolumeResponse{}, nil
		} else if err != nil {
			return nil, status.Errorf(codes.NotFound, "Volume id %s not found: %s",
				req.GetVolumeId(),
				err.Error())
		} else {
			logrus.Infof("Volume %s was deleted or cannot be found", req.GetVolumeId())
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}
	}

	// Mount volume onto the path
	if err = s.driver.Unmount(req.GetVolumeId(), req.GetTargetPath(), nil); err != nil {
		logrus.Infof("Unable to unmount volume %s onto %s: %s",
			req.GetVolumeId(),
			req.GetTargetPath(),
			err.Error())
	}

	if s.driver.Type() == api.DriverType_DRIVER_TYPE_BLOCK {
		if err = s.driver.Detach(req.GetVolumeId(), nil); err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"Unable to detach volume: %s",
				err.Error())
		}
	}

	logrus.Infof("Volume %s unmounted from path %s", req.GetVolumeId(), req.GetTargetPath())

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetCapabilities is a CSI API function which seems to be setup for
// future patches
func (s *OsdCsiServer) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest,
) (*csi.NodeGetCapabilitiesResponse, error) {

	logrus.Debugf("NodeGetCapabilities req[%#v]", req)

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_UNKNOWN,
					},
				},
			},
		},
	}, nil
}

// cleanupEphemeral detaches and deletes an ephemeral volume if either attach or mount fails
func (s *OsdCsiServer) cleanupEphemeral(ctx context.Context, conn *grpc.ClientConn, volumeId string, detach bool) {
	if detach {
		mounts := api.NewOpenStorageMountAttachClient(conn)
		if _, err := mounts.Detach(ctx, &api.SdkVolumeDetachRequest{
			VolumeId: volumeId,
		}); err != nil {
			logrus.Errorf("Failed to detach ephemeral volume %s during cleanup: %v", volumeId, err.Error())
			return
		}
	}
	volumes := api.NewOpenStorageVolumeClient(conn)
	if _, err := volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: volumeId,
	}); err != nil {
		logrus.Errorf("Failed to delete ephemeral volume %s during cleanup: %v", volumeId, err.Error())
	}
}

func verifyTargetLocation(targetPath string) error {
	fileInfo, err := os.Lstat(targetPath)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("Target location %s does not exist", targetPath)
	} else if err != nil {
		return fmt.Errorf(
			"Unknown error while verifying target location %s: %s",
			targetPath,
			err.Error())
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("Target location %s is not a directory", targetPath)
	}

	return nil
}
