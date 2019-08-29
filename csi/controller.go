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
	"encoding/json"
	"fmt"

	"github.com/portworx/kvdb"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/util"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// These keys are for accessing PVC Metadata added from the external-provisioner
	osdParameterPrefix   = "csi.openstorage.org/"
	osdPvcNameKey        = osdParameterPrefix + "pvc-name"
	osdPvcNamespaceKey   = osdParameterPrefix + "pvc-namespace"
	osdPvcAnnotationsKey = osdParameterPrefix + "pvc-annotations"
	osdPvcLabelsKey      = osdParameterPrefix + "pvc-labels"

	// in-tree keys for name and namespace
	intreePvcNameKey      = "pvc"
	intreePvcNamespaceKey = "namespace"

	volumeCapabilityMessageMultinodeVolume    = "Volume is a multinode volume"
	volumeCapabilityMessageNotMultinodeVolume = "Volume is not a multinode volume"
	volumeCapabilityMessageReadOnlyVolume     = "Volume is read only"
	volumeCapabilityMessageNotReadOnlyVolume  = "Volume is not read only"
	defaultCSIVolumeSize                      = uint64(1024 * 1024 * 1024)
)

// ControllerGetCapabilities is a CSI API functions which returns to the caller
// the capabilities of the OSD CSI driver.
func (s *OsdCsiServer) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest,
) (*csi.ControllerGetCapabilitiesResponse, error) {

	// Cloning: creation of volumes from snapshots, supported
	capClone := &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
			},
		},
	}

	// Creating and deleting volumes supported
	capCreateDeleteVolume := &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			},
		},
	}

	// Resizing volumes supported
	capExpandVolume := &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
			},
		},
	}

	// Creating and deleting snapshots
	capCreateDeleteSnapshot := &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
			},
		},
	}

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: []*csi.ControllerServiceCapability{
			capClone,
			capCreateDeleteVolume,
			capExpandVolume,
			capCreateDeleteSnapshot,
		},
	}, nil

}

// ControllerPublishVolume is a CSI API implements the attachment of a volume
// on to a node.
func (s *OsdCsiServer) ControllerPublishVolume(
	context.Context,
	*csi.ControllerPublishVolumeRequest,
) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "This request is not supported")
}

// ControllerUnpublishVolume is a CSI API which implements the detaching of a volume
// onto a node.
func (s *OsdCsiServer) ControllerUnpublishVolume(
	context.Context,
	*csi.ControllerUnpublishVolumeRequest,
) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "This request is not supported")
}

// ValidateVolumeCapabilities is a CSI API used by container orchestration systems
// to make sure a volume specification is validiated by the CSI driver.
func (s *OsdCsiServer) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest,
) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	capabilities := req.GetVolumeCapabilities()
	if capabilities == nil || len(capabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume_capabilities must be specified")
	}
	id := req.GetVolumeId()
	if len(id) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume_id must be specified")
	}

	// Log request
	logrus.Debugf("ValidateVolumeCapabilities of id %s "+
		"capabilities %#v "+
		id,
		capabilities)

	// Get grpc connection
	conn, err := s.getConn()
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Unable to connect to SDK server: %v", err)
	}

	// Get secret if any was passed
	ctx = s.setupContextWithToken(ctx, req.GetSecrets())

	// Check ID is valid with the specified volume capabilities
	volumes := api.NewOpenStorageVolumeClient(conn)
	resp, err := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: id,
	})
	if err != nil {
		return nil, status.Error(codes.NotFound, "ID not found")
	}
	v := resp.GetVolume()
	if v.Id != id {
		errs := fmt.Sprintf(
			"Driver volume id [%s] does not equal requested id of: %s",
			v.Id,
			id)
		logrus.Errorln(errs)
		return nil, status.Error(codes.Internal, errs)
	}
	// Setup uninitialized response object
	result := &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeContext:      req.GetVolumeContext(),
			VolumeCapabilities: req.GetVolumeCapabilities(),
			Parameters:         req.GetParameters(),
		},
	}

	// Check capability
	for _, capability := range capabilities {
		// Currently the CSI spec defines all storage as "file systems."
		// So we do not need to check this with the volume. All we will check
		// here is the validity of the capability access type.
		if capability.GetMount() == nil && capability.GetBlock() == nil {
			return nil, status.Error(
				codes.InvalidArgument,
				"Cannot have both mount and block be undefined")
		}

		// Check access mode is setup correctly
		mode := capability.GetAccessMode()
		switch {
		case mode.Mode == csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER:
			if v.Spec.Shared {
				result.Confirmed = nil
				result.Message = volumeCapabilityMessageMultinodeVolume
				break
			}
			if v.Readonly {
				result.Confirmed = nil
				result.Message = volumeCapabilityMessageReadOnlyVolume
				break
			}
		case mode.Mode == csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY:
			if v.Spec.Shared {
				result.Confirmed = nil
				result.Message = volumeCapabilityMessageMultinodeVolume
				break
			}
			if !v.Readonly {
				result.Confirmed = nil
				result.Message = volumeCapabilityMessageNotReadOnlyVolume
				break
			}
		case mode.Mode == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY:
			if !v.Spec.Shared {
				result.Confirmed = nil
				result.Message = volumeCapabilityMessageNotMultinodeVolume
				break
			}
			if !v.Readonly {
				result.Confirmed = nil
				result.Message = volumeCapabilityMessageNotReadOnlyVolume
				break
			}
		case mode.Mode == csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER ||
			mode.Mode == csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER:
			if !v.Spec.Shared {
				result.Confirmed = nil
				result.Message = volumeCapabilityMessageNotMultinodeVolume
				break
			}
			if v.Readonly {
				result.Confirmed = nil
				result.Message = volumeCapabilityMessageReadOnlyVolume
				break
			}
		default:
			return nil, status.Errorf(
				codes.InvalidArgument,
				"AccessMode %s is not allowed",
				mode.Mode.String())
		}

		if result.Confirmed == nil {
			return result, nil
		}
	}

	// If we passed all the checks, then it is valid.
	// result.Message needs to be empty on return
	return result, nil
}

// osdVolumeContext returns the attributes of a volume as a map
// to be returned to the CSI API caller
func osdVolumeContext(v *api.Volume) map[string]string {
	return map[string]string{
		api.SpecParent: v.GetSource().GetParent(),
		api.SpecSecure: fmt.Sprintf("%v", v.GetSpec().GetEncrypted()),
		api.SpecShared: fmt.Sprintf("%v", v.GetSpec().GetShared()),
		"readonly":     fmt.Sprintf("%v", v.GetReadonly()),
		"attached":     v.AttachedState.String(),
		"state":        v.State.String(),
		"error":        v.GetError(),
	}
}

func addJsonMapToMetadata(params string, metadata map[string]string) (map[string]string, error) {
	decodedParams := make(map[string]string)
	if len(params) > 0 {
		// Decode and add labels
		err := json.Unmarshal([]byte(params), &decodedParams)
		if err != nil {
			return metadata, err
		}
		for k, v := range decodedParams {
			metadata[k] = v
		}
	}

	return metadata, nil
}

func getPVCMetadata(params map[string]string) (map[string]string, error) {
	metadata := make(map[string]string)

	metadata[intreePvcNameKey] = params[osdPvcNameKey]
	metadata[intreePvcNamespaceKey] = params[osdPvcNamespaceKey]
	metadata, err := addJsonMapToMetadata(params[osdPvcAnnotationsKey], metadata)
	if err != nil {
		return nil, err
	}
	metadata, err = addJsonMapToMetadata(params[osdPvcLabelsKey], metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// CreateVolume is a CSI API which creates a volume on OSD
// This function supports snapshots if the parent volume id is supplied
// in the parameters.
func (s *OsdCsiServer) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest,
) (*csi.CreateVolumeResponse, error) {

	// Log request
	logrus.Debugf("CreateVolume req[%#v]", *req)

	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name must be provided")
	}
	if req.GetVolumeCapabilities() == nil || len(req.GetVolumeCapabilities()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities must be provided")
	}

	// Get parameters
	spec, locator, source, err := s.specHandler.SpecFromOpts(req.GetParameters())
	if err != nil {
		e := fmt.Sprintf("Unable to get parameters: %s\n", err.Error())
		logrus.Errorln(e)
		return nil, status.Error(codes.InvalidArgument, e)
	}

	// Get PVC Metadata and add to locator.VolumeLabels
	pvcMetadata, err := getPVCMetadata(req.GetParameters())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Unable to get PVC Metadata: %v", err)
	}
	for k, v := range pvcMetadata {
		locator.VolumeLabels[k] = v
	}

	// Get parent ID from request: snapshot or volume
	if req.GetVolumeContentSource() != nil {
		if sourceSnap := req.GetVolumeContentSource().GetSnapshot(); sourceSnap != nil {
			source.Parent = sourceSnap.SnapshotId
		}

		if sourceVol := req.GetVolumeContentSource().GetVolume(); sourceVol != nil {
			source.Parent = sourceVol.VolumeId
		}
	}

	// Get Size
	if req.GetCapacityRange() != nil && req.GetCapacityRange().GetRequiredBytes() != 0 {
		spec.Size = uint64(req.GetCapacityRange().GetRequiredBytes())
	} else {
		spec.Size = defaultCSIVolumeSize
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

	// Check ID is valid with the specified volume capabilities
	volumes := api.NewOpenStorageVolumeClient(conn)

	// Create volume
	var newVolumeId string
	if source.Parent == "" {
		// Get Capabilities and Size
		spec.Shared = csiRequestsSharedVolume(req)

		createResp, err := volumes.Create(ctx, &api.SdkVolumeCreateRequest{
			Name:   req.GetName(),
			Spec:   spec,
			Labels: locator.GetVolumeLabels(),
		})
		if err != nil {
			return nil, err
		}
		newVolumeId = createResp.VolumeId
	} else {
		cloneResp, err := volumes.Clone(ctx, &api.SdkVolumeCloneRequest{
			Name:     req.GetName(),
			ParentId: source.Parent,
		})
		if err != nil {
			return nil, err
		}
		newVolumeId = cloneResp.VolumeId
	}

	// Get volume information
	inspectResp, err := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: newVolumeId,
	})
	if err != nil {
		return nil, err
	}

	// Create response
	volume := &csi.Volume{}
	osdToCsiVolumeInfo(volume, inspectResp.GetVolume())
	return &csi.CreateVolumeResponse{
		Volume: volume,
	}, nil
}

// DeleteVolume is a CSI API which deletes a volume
func (s *OsdCsiServer) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest,
) (*csi.DeleteVolumeResponse, error) {

	// Log request
	logrus.Debugf("DeleteVolume req[%#v]", *req)

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume id must be provided")
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

	// Check ID is valid with the specified volume capabilities
	volumes := api.NewOpenStorageVolumeClient(conn)

	// Delete volume
	if _, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: req.GetVolumeId(),
	}); err != nil {
		e := fmt.Sprintf("Unable to delete volume with id %s: %s",
			req.GetVolumeId(),
			err.Error())
		logrus.Errorln(e)
		return nil, status.Error(codes.Internal, e)
	}

	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerExpandVolume is a CSI API which resizes a volume
func (s *OsdCsiServer) ControllerExpandVolume(
	ctx context.Context,
	req *csi.ControllerExpandVolumeRequest,
) (*csi.ControllerExpandVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume id must be provided")
	} else if req.GetCapacityRange() == nil {
		return nil, status.Error(codes.InvalidArgument, "Capacity range must be provided")
	} else if req.GetCapacityRange().GetRequiredBytes() < 0 || req.GetCapacityRange().GetLimitBytes() < 0 {
		return nil, status.Error(codes.InvalidArgument, "Capacity ranges values cannot be negative")
	}

	// Get Size
	spec := &api.VolumeSpecUpdate{}
	newSize := uint64(req.GetCapacityRange().GetRequiredBytes())
	spec.SizeOpt = &api.VolumeSpecUpdate_Size{
		Size: newSize,
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

	// If the new size is greater than the current size, a volume update
	// should be issued. Otherwise, no operation should occur.
	volumes := api.NewOpenStorageVolumeClient(conn)

	// Update volume with new size
	_, err = volumes.Update(ctx, &api.SdkVolumeUpdateRequest{
		VolumeId: req.GetVolumeId(),
		Spec:     spec,
	})
	if err != nil {
		if err == kvdb.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "Volume id %s not found", req.GetVolumeId())
		}
		return nil, status.Errorf(codes.Internal, "Failed to update volume size: %v", err)
	}

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         int64(newSize),
		NodeExpansionRequired: false,
	}, nil
}

func osdToCsiVolumeInfo(dest *csi.Volume, src *api.Volume) {
	dest.VolumeId = src.GetId()
	dest.CapacityBytes = int64(src.Spec.GetSize())
	dest.VolumeContext = osdVolumeContext(src)
}

func csiRequestsSharedVolume(req *csi.CreateVolumeRequest) bool {
	for _, cap := range req.GetVolumeCapabilities() {
		// Check access mode is setup correctly
		mode := cap.GetAccessMode().GetMode()
		if mode == csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER ||
			mode == csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER {
			return true
		}
	}
	return false
}

// CreateSnapshot is a CSI implementation to create a snapshot from the volume
func (s *OsdCsiServer) CreateSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest,
) (*csi.CreateSnapshotResponse, error) {

	if len(req.GetSourceVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume id must be provided")
	} else if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name must be provided")
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

	// Check ID is valid with the specified volume capabilities
	volumes := api.NewOpenStorageVolumeClient(conn)

	// Check if the snapshot with this name already exists
	v, err := util.VolumeFromNameSdk(ctx, volumes, req.GetName())
	if err == nil {
		// Verify the parent is the same
		if req.GetSourceVolumeId() != v.GetSource().GetParent() {
			return nil, status.Error(codes.AlreadyExists, "Requested snapshot already exists for another source volume id")
		}

		return &csi.CreateSnapshotResponse{
			Snapshot: &csi.Snapshot{
				SnapshotId:     v.GetId(),
				SourceVolumeId: v.GetSource().GetParent(),
				CreationTime:   v.GetCtime(),
				ReadyToUse:     true,
			},
		}, nil
	}

	// Get any labels passed in by the CO
	_, locator, _, err := s.specHandler.SpecFromOpts(req.GetParameters())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to get parameters: %v", err)
	}

	// Create snapshot
	snapResp, err := volumes.SnapshotCreate(ctx, &api.SdkVolumeSnapshotCreateRequest{
		VolumeId: req.GetSourceVolumeId(),
		Name:     req.GetName(),
		Labels:   locator.GetVolumeLabels(),
	})
	if err != nil {
		if err == kvdb.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "Volume id %s not found", req.GetSourceVolumeId())
		}
		return nil, status.Errorf(codes.Internal, "Failed to create snapshot: %v", err)
	}
	snapshotID := snapResp.SnapshotId

	snapInfo, err := util.VolumeFromIdSdk(ctx, volumes, snapshotID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get information about the snapshot: %v", err)
	}

	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SnapshotId:     snapshotID,
			SourceVolumeId: req.GetSourceVolumeId(),
			CreationTime:   snapInfo.GetCtime(),
			ReadyToUse:     true,
		},
	}, nil
}

// DeleteSnapshot is a CSI implementation to delete a snapshot
func (s *OsdCsiServer) DeleteSnapshot(
	ctx context.Context,
	req *csi.DeleteSnapshotRequest,
) (*csi.DeleteSnapshotResponse, error) {

	if len(req.GetSnapshotId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Snapshot id must be provided")
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

	// Check ID is valid with the specified volume capabilities
	volumes := api.NewOpenStorageVolumeClient(conn)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: req.GetSnapshotId(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to delete snapshot %s: %v",
			req.GetSnapshotId(),
			err)
	}

	return &csi.DeleteSnapshotResponse{}, nil
}

// ListSnapshots is not supported (we can add this later)
func (s *OsdCsiServer) ListSnapshots(
	ctx context.Context,
	req *csi.ListSnapshotsRequest,
) (*csi.ListSnapshotsResponse, error) {

	// The function ListSnapshots is also not published as
	// supported by this implementation
	return nil, status.Error(codes.Unimplemented, "ListSnapshots is not implemented")
}
