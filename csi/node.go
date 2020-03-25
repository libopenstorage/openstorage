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

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/pkg/errors"
	"github.com/portworx/kvdb"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/mount"
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
	if req.GetVolumeCapability() == nil || req.GetVolumeCapability().GetAccessMode() == nil ||
		req.GetVolumeCapability().GetAccessMode().Mode == csi.VolumeCapability_AccessMode_UNKNOWN {
		return nil, status.Error(codes.InvalidArgument, "Volume access mode must be provided")
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

	// prepare for mount/attaching
	mountAttachClient := api.NewOpenStorageMountAttachClient(conn)
	var attachedVolDevice string
	if driverType == api.DriverType_DRIVER_TYPE_BLOCK {
		var attachResp *api.SdkVolumeAttachResponse
		// attach is assumed to be idempotent
		// attach is assumed to return the same DevicePath on each call
		if attachResp, err = mountAttachClient.Attach(ctx, &api.SdkVolumeAttachRequest{
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

		if attachResp == nil || attachResp.DevicePath == "" {
			if spec.Ephemeral {
				logrus.Errorf("Failed to attach ephemeral volume, attach resp is empty %s", volumeId)
				s.cleanupEphemeral(ctx, conn, volumeId, false)
			}
			return nil, status.Error(
				codes.Internal,
				"failed to attach device, got empty attach response")
		}
		attachedVolDevice = attachResp.DevicePath
	}

	// implement idempotency for nodePublish calls
	//https://github.com/container-storage-interface/spec/blob/master/spec.md#nodeunpublishvolume
	targetPath := req.GetTargetPath()
	isBlockAccessType := false
	if req.GetVolumeCapability().GetBlock() != nil {
		isBlockAccessType = true
	}

	isSingleNodeAccessMode := false

	accessMode := req.GetVolumeCapability().GetAccessMode()
	if accessMode.Mode == csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER ||
		accessMode.Mode == csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY {
		isSingleNodeAccessMode = true
	}

	isMounted := false
	isMountedToRequestedTargetPath := false

	var mounter mount.Manager
	if isBlockAccessType {
		mounter, err = mount.NewRawBindMounter([]string{""}, nil, []string{}, "")
	} else {
		mounter, err = mount.New(mount.DeviceMount, nil, []string{""}, nil, []string{}, "")
	}
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to instantiate mounter: %s",
			err.Error())
	}

	err = mounter.Reload(attachedVolDevice)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to read mounts for device %s: %s",
			attachedVolDevice,
			err.Error())
	}

	deviceMountsNumber := mounter.HasMounts(attachedVolDevice)
	if deviceMountsNumber > 0 {
		isMounted = true
	}

	isMountedToRequestedTargetPath, err = mounter.Exists(attachedVolDevice, targetPath)
	if err != nil {
		if err == mount.ErrEnoent {
			isMountedToRequestedTargetPath = false
		} else {
			return nil, status.Errorf(
				codes.Internal,
				"failed to check mounts for device %s: %s",
				attachedVolDevice,
				err.Error())
		}
	}

	logrus.Infof(
		"isMounted %v, targetPath: %v, isSingleNodeAccessMode %v, isMountedToRequestedTargetPath %v,"+
			" attachedVolDevice: %v, isBlockAccessType %v",
		isMounted,
		targetPath,
		isSingleNodeAccessMode,
		isMountedToRequestedTargetPath,
		attachedVolDevice,
		isBlockAccessType,
	)

	// this scenario is common for all access modes
	// T1=T2, requested target path and volume mount path are equal
	if isMounted && isMountedToRequestedTargetPath {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	//currently we do not compare req parameters, e. g. cases like T1=T2 && P1!=P2 are not implemented
	// that means for MULTI_NODE cases
	// when requested target path and volume mount path are NOT equal volume shall be mounted one more time

	// NON MULTI_NODE cases
	if isSingleNodeAccessMode {
		// T1!=T2
		if isMounted && !isMountedToRequestedTargetPath {
			return nil, status.Errorf(
				codes.FailedPrecondition,
				"failed to ensure target location %s: %s",
				targetPath,
				err.Error())
		}
	}

	//ensureTargetLocation verifies target location and creates the one if it doesn't exist
	if err = ensureTargetLocation(targetPath, isBlockAccessType); err != nil {
		return nil, status.Errorf(
			codes.Aborted,
			"failed to ensure target location %s: %s",
			targetPath,
			err.Error())
	}

	// for volumes with mount access type just mount volume onto the path
	if _, err := mountAttachClient.Mount(ctx, &api.SdkVolumeMountRequest{
		VolumeId:  volumeId,
		MountPath: targetPath,
		Options:   opts,
	}); err != nil {
		if spec.Ephemeral {
			logrus.Errorf("Failed to mount ephemeral volume %s: %v", volumeId, err.Error())
			s.cleanupEphemeral(ctx, conn, volumeId, true)
		}
		return nil, err
	}

	logrus.Infof("Volume %s mounted on %s", volumeId, targetPath)

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume is a CSI API call which unmounts the volume.
func (s *OsdCsiServer) NodeUnpublishVolume(
	_ context.Context,
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

	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	// Get volume information
	vols, err := s.driver.Inspect([]string{volumeID})
	if err == kvdb.ErrNotFound || len(vols) < 1 {
		return nil, status.Errorf(codes.NotFound, "volume id %s not found: %s",
			volumeID,
			kvdb.ErrNotFound.Error())
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"failed to inspect volume %s: %s",
			volumeID,
			err.Error())
	}

	v := vols[0]
	isBlockAccessType := v.Format == api.FSType_FS_TYPE_NONE
	var mounter mount.Manager

	if isBlockAccessType {
		mounter, err = mount.NewRawBindMounter([]string{""}, nil, []string{}, "")
	} else {
		mounter, err = mount.New(mount.DeviceMount, nil, []string{""}, nil, []string{}, "")
	}
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to instantiate mounter: %s",
			err.Error())
	}

	err = mounter.Reload(v.DevicePath)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to read mounts for device %s: %s",
			v.DevicePath,
			err.Error())
	}

	// it is not mounted to this targetPath
	_, err = mounter.GetSourcePath(targetPath)
	if err != nil {
		if err == mount.ErrEnoent {
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}
		return nil, status.Errorf(codes.Internal,
			"cannot get sourcePath for targetPath %s for volume %s, err: %s",
			targetPath,
			volumeID,
			err.Error())
	}

	if err = s.driver.Unmount(volumeID, targetPath, nil); err != nil {
		return nil, status.Errorf(codes.Internal,
			"unable to unmount volume %s onto %s: %s",
			volumeID,
			targetPath,
			err.Error())
	}

	err = mounter.Reload(v.DevicePath)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to read mounts for device %s: %s",
			v.DevicePath,
			err.Error())
	}

	// Delete the mount point
	err = mounter.RemoveMountPath(targetPath, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"cannot clean targetPath %s for volume %s, err: %s",
			targetPath,
			volumeID,
			err.Error())
	}

	// TODO: should be deattached only in case when there are no mountpoints
	if s.driver.Type() == api.DriverType_DRIVER_TYPE_BLOCK {
		if err = s.driver.Detach(req.GetVolumeId(), nil); err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"Unable to detach volume: %s",
				err.Error())
		}
	}

	logrus.Infof("Volume %s unmounted", req.GetVolumeId())

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

func ensureTargetLocation(targetPath string, isBlock bool) error {
	_, err := os.Lstat(targetPath)
	if os.IsNotExist(err) {
		if isBlock {
			if err = makeFile(targetPath); err != nil {
				return fmt.Errorf("cannot create target location %s for block volume, err: %s", targetPath, err.Error())
			}

			return nil
		}

		if err = makeDir(targetPath); err != nil {
			return fmt.Errorf("cannot create target location %s for mount volume, err: %s", targetPath, err.Error())
		}

		return nil
	}

	return errors.Wrap(err, "unknown error while verifying target location")
}

func makeFile(pathname string) error {
	f, err := os.OpenFile(pathname, os.O_CREATE, os.FileMode(0644))
	defer f.Close()
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	_, err = f.Stat()
	if err != nil {
		return err
	}

	return nil
}

func makeDir(pathname string) error {
	err := os.MkdirAll(pathname, os.FileMode(0755))
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	return nil
}
