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
	"math"
	"reflect"
	"sort"

	"github.com/portworx/kvdb"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/grpcutil"
	"github.com/libopenstorage/openstorage/pkg/units"
	"github.com/libopenstorage/openstorage/pkg/util"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
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

	// These keys are for accessing Snapshot Metadata added from the external-provisioner
	osdSnapshotLabelsTypeKey   = osdParameterPrefix + "snapshot-type"
	osdSnapshotCredentialIDKey = osdParameterPrefix + "credential-id"

	// in-tree keys for name and namespace
	intreePvcNameKey      = "pvc"
	intreePvcNamespaceKey = "namespace"

	// CSI keys for PVC metadata
	csiPVNameKey       = "csi.storage.k8s.io/pv/name"
	csiPVCNameKey      = "csi.storage.k8s.io/pvc/name"
	csiPVCNamespaceKey = "csi.storage.k8s.io/pvc/namespace"

	volumeCapabilityMessageMultinodeVolume    = "Volume is a multinode volume"
	volumeCapabilityMessageNotMultinodeVolume = "Volume is not a multinode volume"
	volumeCapabilityMessageReadOnlyVolume     = "Volume is read only"
	volumeCapabilityMessageNotReadOnlyVolume  = "Volume is not read only"
	defaultCSIVolumeSize                      = uint64(units.GiB * 1)

	// driver type
	DriverTypeLocal = "local"
	DriverTypeCloud = "cloud"
)

// ControllerGetCapabilities is a CSI API functions which returns to the caller
// the capabilities of the OSD CSI driver.
func (s *OsdCsiServer) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest,
) (*csi.ControllerGetCapabilitiesResponse, error) {

	caps := []csi.ControllerServiceCapability_RPC_Type{
		// Creating and deleting volumes supported
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,

		// Cloning: creation of volumes from snapshots, supported
		csi.ControllerServiceCapability_RPC_CLONE_VOLUME,

		// Resizing volumes supported
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,

		// Creating and deleting snapshots
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,

		// Listing snapshots
		csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,

		// Get single volume
		csi.ControllerServiceCapability_RPC_GET_VOLUME,

		// Volume condition
		csi.ControllerServiceCapability_RPC_VOLUME_CONDITION,

		// GetCapacity
		csi.ControllerServiceCapability_RPC_GET_CAPACITY,
	}

	var serviceCapabilities []*csi.ControllerServiceCapability

	for _, cap := range caps {
		serviceCapabilities = append(serviceCapabilities, &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		})
	}

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: serviceCapabilities,
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

// ControllerGetVolume is a CSI API which implements getting a single volume.
// This function skips auth and directly hits the driver as it is read-only
// and only exposed via the CSI unix domain socket. If a secrets field is added
// in csi.ControllerGetVolumeRequest, we can update this to hit the SDK and use auth.
func (s *OsdCsiServer) ControllerGetVolume(
	ctx context.Context,
	req *csi.ControllerGetVolumeRequest,
) (*csi.ControllerGetVolumeResponse, error) {

	clogger.WithContext(ctx).Tracef("ControllerGetVolume request received. VolumeID: %s", req.GetVolumeId())

	vol, err := s.driverGetVolume(ctx, req.GetVolumeId())
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
			return &csi.ControllerGetVolumeResponse{
				Volume: &csi.Volume{
					VolumeId: req.GetVolumeId(),
				},
				Status: &csi.ControllerGetVolumeResponse_VolumeStatus{
					VolumeCondition: &csi.VolumeCondition{
						Abnormal: true,
						Message:  fmt.Sprintf("Volume ID %s not found", req.GetVolumeId()),
					},
				},
			}, nil
		}
		return nil, err
	}

	return &csi.ControllerGetVolumeResponse{
		Volume: &csi.Volume{
			CapacityBytes: int64(vol.Spec.Size),
			VolumeId:      vol.Id,
		},
		Status: &csi.ControllerGetVolumeResponse_VolumeStatus{
			VolumeCondition: getVolumeCondition(vol),
		},
	}, nil
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
	clogger.WithContext(ctx).Infof("csi.ValidateVolumeCapabilities of id %s "+
		"capabilities %#v ", id, capabilities)

	// Get grpc connection
	conn, err := s.getConn()
	if err != nil {
		return nil, status.Errorf(
			codes.Unavailable,
			"Unable to connect to SDK server: %v", err)
	}

	// Get secret if any was passed
	ctx = s.setupContext(ctx, req.GetSecrets())
	ctx, cancel := grpcutil.WithDefaultTimeout(ctx)
	defer cancel()

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
		clogger.WithContext(ctx).Errorln(errs)
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
			if v.Spec.Sharedv4 || v.Spec.Shared {
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
			if v.Spec.Sharedv4 || v.Spec.Shared {
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
			if !v.Spec.Sharedv4 && !v.Spec.Shared {
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
			if !v.Spec.Sharedv4 && !v.Spec.Shared {
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
		api.SpecParent:   v.GetSource().GetParent(),
		api.SpecSecure:   fmt.Sprintf("%v", v.GetSpec().GetEncrypted()),
		api.SpecShared:   fmt.Sprintf("%v", v.GetSpec().GetShared()),
		api.SpecSharedv4: fmt.Sprintf("%v", v.GetSpec().GetSharedv4()),
		"readonly":       fmt.Sprintf("%v", v.GetReadonly()),
		"attached":       v.AttachedState.String(),
		"state":          v.State.String(),
		"error":          v.GetError(),
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

	// Only add PVC name/namespace if present. Do not add empty values.
	// This is useful for non-k8s schedulers.
	if pvcName, ok := params[osdPvcNameKey]; ok {
		metadata[intreePvcNameKey] = pvcName
	}
	if pvcNamespace, ok := params[osdPvcNamespaceKey]; ok {
		metadata[intreePvcNamespaceKey] = pvcNamespace
	}
	metadata, err := addJsonMapToMetadata(params[osdPvcAnnotationsKey], metadata)
	if err != nil {
		return nil, err
	}
	metadata, err = addJsonMapToMetadata(params[osdPvcLabelsKey], metadata)
	if err != nil {
		return nil, err
	}
	if pvcName, ok := params[csiPVCNameKey]; ok {
		metadata[intreePvcNameKey] = pvcName
	}
	if pvcNamespace, ok := params[csiPVCNamespaceKey]; ok {
		metadata[intreePvcNamespaceKey] = pvcNamespace
	}

	return metadata, nil
}

func cleanupVolumeLabels(labels map[string]string) map[string]string {
	delete(labels, osdPvcNameKey)
	delete(labels, osdPvcNamespaceKey)
	delete(labels, osdPvcAnnotationsKey)
	delete(labels, osdPvcLabelsKey)
	delete(labels, csiPVNameKey)
	delete(labels, csiPVCNameKey)
	delete(labels, csiPVCNamespaceKey)

	return labels
}

func validateCreateVolumeCapabilities(caps []*csi.VolumeCapability) error {
	if len(caps) == 0 {
		return status.Error(codes.InvalidArgument, "Volume capabilities must be provided")
	}

	var shared bool
	var block bool
	for _, cap := range caps {
		mode := cap.GetAccessMode().GetMode()
		if mode == csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER ||
			mode == csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER ||
			mode == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY {
			shared = true
		}

		if cap.GetBlock() != nil {
			block = true
		}
	}

	if block && shared {
		return status.Errorf(
			codes.InvalidArgument,
			"Shared raw block volumes are not supported")
	}

	return nil
}

func validateCreateVolumeCapabilitiesPure(caps []*csi.VolumeCapability, proxySpec *api.ProxySpec) error {
	if len(caps) == 0 {
		return status.Error(codes.InvalidArgument, "Volume capabilities must be provided")
	}

	var shared bool
	var mount bool
	var block bool
	for _, cap := range caps {
		mode := cap.GetAccessMode().GetMode()
		if mode == csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER ||
			mode == csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER ||
			mode == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY {
			shared = true
		}

		if cap.GetBlock() != nil {
			block = true
		}

		if cap.GetMount() != nil {
			mount = true
		}
	}

	// Check for shared FA DA volumes. Shared filesystems aren't supported.
	// Shared raw block volumes are temporarily disabled due to PWX-23530.
	// All FA raw block volumes are disabled for now.
	if proxySpec.ProxyProtocol == api.ProxyProtocol_PROXY_PROTOCOL_PURE_BLOCK {
		if mount && shared {
			return status.Errorf(
				codes.InvalidArgument,
				"FlashArray Direct Access shared filesystems are not supported",
			)
		}
	}

	// Check for FB DA volumes: all allowed except raw block
	if proxySpec.ProxyProtocol == api.ProxyProtocol_PROXY_PROTOCOL_PURE_FILE && block {
		return status.Errorf(
			codes.InvalidArgument,
			"FlashBlade Direct Access volumes do not support raw block",
		)
	}

	return nil
}

// CreateVolume is a CSI API which creates a volume on OSD
// This function supports snapshots if the parent volume id is supplied
// in the parameters.
func (s *OsdCsiServer) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest,
) (*csi.CreateVolumeResponse, error) {
	// Log request
	clogger.WithContext(ctx).Infof("csi.CreateVolume request received. Volume: %s", req.GetName())

	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name must be provided")
	}

	// Get parameters
	spec, locator, source, err := s.specHandler.SpecFromOpts(req.GetParameters())
	if err != nil {
		e := fmt.Sprintf("Unable to get parameters: %s\n", err.Error())
		clogger.WithContext(ctx).Errorln(e)
		return nil, status.Error(codes.InvalidArgument, e)
	}

	if spec.IsPureVolume() {
		err = validateCreateVolumeCapabilitiesPure(req.GetVolumeCapabilities(), spec.GetProxySpec())
		if err != nil {
			return nil, err
		}
	} else {
		err = validateCreateVolumeCapabilities(req.GetVolumeCapabilities())
		if err != nil {
			return nil, err
		}
	}

	// Get PVC Metadata and add to locator.VolumeLabels
	// This will override any storage class secrets added above.
	pvcMetadata, err := getPVCMetadata(req.GetParameters())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Unable to get PVC Metadata: %v", err)
	}
	for k, v := range pvcMetadata {
		locator.VolumeLabels[k] = v
	}

	// Copy all SC Parameters (from req.Parameters) to locator.VolumeLabels.
	// This explicit copy matches the equivalent behavior in the in-tree driver
	if len(locator.VolumeLabels) == 0 {
		locator.VolumeLabels = make(map[string]string)
	}
	for k, v := range req.Parameters {
		locator.VolumeLabels[k] = v
	}

	// Add encryption secret information to VolumeLabels
	locator.VolumeLabels = s.addEncryptionInfoToLabels(locator.VolumeLabels, req.GetSecrets())

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
		size, err := roundUpToNearestGiB(req.GetCapacityRange().GetRequiredBytes())
		if err != nil {
			return nil, fmt.Errorf("failed to round volume size up to nearest GiB: %v", err.Error())
		}
		spec.Size = size

	} else {
		spec.Size = defaultCSIVolumeSize
	}

	// cleanup duplicate information after pulling from req.GetParameters
	locator.VolumeLabels = cleanupVolumeLabels(locator.VolumeLabels)

	// Get grpc connection
	conn, err := s.getRemoteConn(ctx)
	if err != nil {
		logrus.Errorf("failed to get remote connection: %v, continuing with local node instead", err)
		conn, err = s.getConn()
		if err != nil {
			return nil, status.Errorf(
				codes.Unavailable,
				"Unable to connect to SDK server: %v", err)
		}
	}

	// Get secret if any was passed
	ctx = s.setupContext(ctx, req.GetSecrets())
	ctx, cancel := grpcutil.WithDefaultTimeout(ctx)
	defer cancel()

	// Check ID is valid with the specified volume capabilities
	volumes := api.NewOpenStorageVolumeClient(conn)

	// Create volume
	var newVolumeId string
	if source.Parent == "" {
		spec, err := resolveSpecFromCSI(spec, req)
		if err != nil {
			return nil, err
		}

		var createResp *api.SdkVolumeCreateResponse
		var createErr error
		topologies := getAllTopologies(req.AccessibilityRequirements)
		if spec.IsPureVolume() && len(topologies) > 0 {
			for _, topo := range topologies {
				spec.TopologyRequirement = &api.TopologyRequirement{
					Labels: topo.Segments,
				}
				createResp, createErr = volumes.Create(ctx, &api.SdkVolumeCreateRequest{
					Name:   req.GetName(),
					Spec:   spec,
					Labels: locator.GetVolumeLabels(),
				})
				if s, ok := status.FromError(createErr); createErr == nil || (ok && s.Code() != codes.ResourceExhausted) {
					break
				}
			}
		} else {
			createResp, createErr = volumes.Create(ctx, &api.SdkVolumeCreateRequest{
				Name:   req.GetName(),
				Spec:   spec,
				Labels: locator.GetVolumeLabels(),
			})
		}
		if createErr != nil {
			return nil, createErr
		}
		newVolumeId = createResp.VolumeId
	} else {
		cloneResp, err := volumes.Clone(ctx, &api.SdkVolumeCloneRequest{
			Name:             req.GetName(),
			ParentId:         source.Parent,
			AdditionalLabels: getClonedPVCMetadata(locator),
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
	osdToCsiVolumeInfo(volume, inspectResp.GetVolume(), req)
	return &csi.CreateVolumeResponse{
		Volume: volume,
	}, nil
}

func getClonedPVCMetadata(locator *api.VolumeLocator) map[string]string {
	metadataLabels := map[string]string{}
	pvcName, ok := locator.VolumeLabels[intreePvcNameKey]
	if ok {
		metadataLabels[intreePvcNameKey] = pvcName
	}
	pvcNamespace, ok := locator.VolumeLabels[intreePvcNamespaceKey]
	if ok {
		metadataLabels[intreePvcNamespaceKey] = pvcNamespace
	}

	return metadataLabels
}

// DeleteVolume is a CSI API which deletes a volume
func (s *OsdCsiServer) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest,
) (*csi.DeleteVolumeResponse, error) {
	// Log request
	clogger.WithContext(ctx).Infof("csi.DeleteVolume request received. VolumeID: %s", req.VolumeId)

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume id must be provided")
	}

	// Get grpc connection
	conn, err := s.getRemoteConn(ctx)
	if err != nil {
		logrus.Errorf("failed to get remote connection: %v, continuing with local node instead", err)
		conn, err = s.getConn()
		if err != nil {
			return nil, status.Errorf(
				codes.Unavailable,
				"Unable to connect to SDK server: %v", err)
		}
	}

	// Get secret if any was passed
	ctx = s.setupContext(ctx, req.GetSecrets())
	ctx, cancel := grpcutil.WithDefaultTimeout(ctx)
	defer cancel()

	// Check ID is valid with the specified volume capabilities
	volumes := api.NewOpenStorageVolumeClient(conn)

	// Delete volume
	if _, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: req.GetVolumeId(),
	}); err != nil {
		e := fmt.Sprintf("Unable to delete volume with id %s: %s",
			req.GetVolumeId(),
			err.Error())
		clogger.WithContext(ctx).Errorln(e)
		return nil, status.Error(codes.Aborted, e)
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
	newSize, err := roundUpToNearestGiB(req.GetCapacityRange().GetRequiredBytes())
	if err != nil {
		return nil, fmt.Errorf("failed to round volume size up to nearest GiB: %v", err.Error())
	}
	spec.SizeOpt = &api.VolumeSpecUpdate_Size{
		Size: newSize,
	}

	// Get grpc connection
	conn, err := s.getConn()
	if err != nil {
		return nil, status.Errorf(
			codes.Unavailable,
			"Unable to connect to SDK server: %v", err)
	}

	// Get secret if any was passed
	ctx = s.setupContext(ctx, req.GetSecrets())
	ctx, cancel := grpcutil.WithDefaultTimeout(ctx)
	defer cancel()

	// If the new size is greater than the current size, a volume update
	// should be issued. Otherwise, no operation should occur.
	volumes := api.NewOpenStorageVolumeClient(conn)

	// Check for idempotency. There are cases where we return an error in the proto
	// driver if the volume update request has no changes, so we return early here instead.
	inspectResp, err := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: req.GetVolumeId(),
	})
	if err != nil {
		if err == kvdb.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "Volume id %s not found", req.GetVolumeId())
		}
		return nil, status.Errorf(codes.Internal, "Failed to get volume %s: %v", req.GetVolumeId(), err)
	}
	if inspectResp.Volume.Spec.Size == spec.GetSize() {
		return &csi.ControllerExpandVolumeResponse{
			CapacityBytes:         int64(newSize),
			NodeExpansionRequired: false,
		}, nil
	}

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

func osdToCsiVolumeInfo(dest *csi.Volume, src *api.Volume, req *csi.CreateVolumeRequest) {
	dest.VolumeId = src.GetId()
	dest.CapacityBytes = int64(src.Spec.GetSize())
	dest.VolumeContext = osdVolumeContext(src)
	dest.ContentSource = req.GetVolumeContentSource()
	if src.Spec.GetTopologyRequirement() != nil && len(src.Spec.TopologyRequirement.Labels) > 0 {
		dest.AccessibleTopology = make([]*csi.Topology, 1)
		segments := map[string]string{}
		for k, v := range src.Spec.TopologyRequirement.Labels {
			segments[k] = v
		}
		dest.AccessibleTopology[0] = &csi.Topology{
			Segments: segments,
		}
	}
}

// isFilesystemSpecSet checks if the fs parameter has been declared
func isFilesystemSpecSet(params map[string]string) bool {
	_, fsSet := params[api.SpecFilesystem]
	return fsSet
}

// resolveSharedSpec makes the following assumptions:
// 1. When a volume is set to RWX or a similar multi-node access mode, we default to Sharedv4
// 2. If a user prefers shared over sharedv4, they may still use it by explicitly declaring "shared": true
func resolveSharedSpec(spec *api.VolumeSpec, req *csi.CreateVolumeRequest) (*api.VolumeSpec, error) {
	// shared or sharedv4 parameter doesn't apply to pure backends so don't set them
	if spec.IsPureVolume() {
		return spec, nil
	}

	var shared bool
	for _, cap := range req.GetVolumeCapabilities() {
		mode := cap.GetAccessMode().GetMode()
		if mode == csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER ||
			mode == csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER ||
			mode == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY {
			shared = true
		}
	}

	// Handle legacy support for shared
	if spec.Shared && !spec.Sharedv4 {
		spec.Shared = shared
	} else {
		// assume sharedv4 by default
		spec.Sharedv4 = shared
	}

	return spec, nil
}

// resolveFSTypeSpec makes the following assumptions:
// 1. If provided, the PX "fstype" parameter should always override corresponding the CSI parameter.
// 2. The default value for spec.Format is determined upstream by SpecFromOpts(req.GetParameters())
func resolveFSTypeSpec(spec *api.VolumeSpec, req *csi.CreateVolumeRequest) (*api.VolumeSpec, error) {
	var csiFsType string
	for _, cap := range req.GetVolumeCapabilities() {
		if cap.GetBlock() != nil {
			spec.Format = api.FSType_FS_TYPE_NONE
			return spec, nil
		}

		// Get FsType according to CSI spec
		if mount := cap.GetMount(); mount != nil {
			csiFsType = mount.FsType
		}
	}

	// if we have the FileSystemSpec option not set, this means the user didn't intend to
	// set the filesystem type based on the "PX way". In this case, we can safely
	// apply the fsType from the CSI request.
	if !isFilesystemSpecSet(req.GetParameters()) && csiFsType != "" {
		format, err := api.FSTypeSimpleValueOf(csiFsType)
		if err != nil {
			return spec, err
		}
		spec.Format = format
	}

	return spec, nil
}

// resolveSpecFromCSI alters the api.VolumeSpec based on any CSI parameters passed in.
// Various volume spec fields have CSI equivalents. This function resolves each one.
func resolveSpecFromCSI(spec *api.VolumeSpec, req *csi.CreateVolumeRequest) (*api.VolumeSpec, error) {
	// Handles shared vs. sharedv4 resolution. We default to Sharedv4
	spec, err := resolveSharedSpec(spec, req)
	if err != nil {
		return nil, err
	}

	// Handle FsType resolution. There are two methods of setting the FsType: PX and CSI.
	// We default to the PX parameter if it is provided.
	// If CSI parameter is provided, but PX is not, use CSI.
	// If neither is provided, use the PX parameter default.
	spec, err = resolveFSTypeSpec(spec, req)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func isSnapshotReady(v *api.Volume) bool {
	return v.GetError() == "" && v.Status == api.VolumeStatus_VOLUME_STATUS_UP
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

	// Get secret if any was passed
	ctx = s.setupContext(ctx, req.GetSecrets())
	ctx, cancel := grpcutil.WithDefaultTimeout(ctx)
	defer cancel()

	// Get any labels passed in by the CO
	_, locator, _, err := s.specHandler.SpecFromOpts(req.GetParameters())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to get parameters: %v", err)
	}
	// Check ID is valid with the specified volume capabilities
	snapshotType, ok := locator.VolumeLabels[osdSnapshotLabelsTypeKey]
	if !ok {
		snapshotType = DriverTypeLocal
	}
	switch snapshotType {
	case DriverTypeCloud:
		return s.createCloudBackup(ctx, req)
	case DriverTypeLocal:
		fallthrough
	default:
		return s.createLocalSnapshot(ctx, req)
	}
}

func (s *OsdCsiServer) createLocalSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest,
) (*csi.CreateSnapshotResponse, error) {
	// Get grpc connection
	conn, err := s.getConn()
	if err != nil {
		return nil, status.Errorf(
			codes.Unavailable,
			"Unable to connect to SDK server: %v", err)
	}

	// Get secret if any was passed
	ctx = s.setupContext(ctx, req.GetSecrets())
	ctx, cancel := grpcutil.WithDefaultTimeout(ctx)
	defer cancel()

	// Check ID is valid with the specified volume capabilities
	volumes := api.NewOpenStorageVolumeClient(conn)

	// Check if the snapshot with this name already exists
	v, err := util.VolumeFromNameSdk(ctx, volumes, req.GetName())
	if err == nil {
		// Verify the parent is the same
		if req.GetSourceVolumeId() != v.GetSource().GetParent() {
			return nil, status.Error(codes.AlreadyExists, "Requested snapshot already exists for another source volume id")
		}
		readyToUse := isSnapshotReady(v)
		if readyToUse {
			return &csi.CreateSnapshotResponse{
				Snapshot: &csi.Snapshot{
					SizeBytes:      int64(v.GetSpec().GetSize()),
					SnapshotId:     v.GetId(),
					SourceVolumeId: v.GetSource().GetParent(),
					CreationTime:   v.GetCtime(),
					ReadyToUse:     readyToUse,
				},
			}, nil
		}
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
		errStatus, ok := status.FromError(err)
		if ok && errStatus.Code() == codes.NotFound {
			// Return a non-final Aborted error, as the PVC may not have been created yet. Otherwise
			// the CSI Snapshotter will not issue retries:
			// https://github.com/kubernetes-csi/external-snapshotter/blob/v6.0.1/pkg/sidecar-controller/snapshot_controller.go#L656-L678
			return nil, status.Errorf(codes.Aborted, "Volume id %s not found: %v", req.GetSourceVolumeId(), err)
		}

		// Check if snapshot has been created but is in error state
		snapInfo, errFindFailed := util.VolumeFromIdSdk(ctx, volumes, req.GetName())
		if errFindFailed == nil && !isSnapshotReady(snapInfo) {

			// If snapshot was created but has errors, cleanup and re-create it on the next iteration.
			_, errCleanupSnap := volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
				VolumeId: snapInfo.Id,
			})
			if errCleanupSnap != nil {
				return nil, status.Errorf(codes.Aborted, "Snapshot create failed and unable to delete failed snapshot %s: %v",
					snapInfo.Id,
					errCleanupSnap)
			}
			logrus.Infof("cleaned up failed snapshot %v in order to retry snapshot create", snapInfo.Id)
		}

		return nil, status.Errorf(codes.Aborted, "Failed to create snapshot: %v", err)
	}
	snapshotID := snapResp.SnapshotId

	snapInfo, err := util.VolumeFromIdSdk(ctx, volumes, snapshotID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "Failed to get information about the snapshot: %v", err)
	}

	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SizeBytes:      int64(snapInfo.GetSpec().GetSize()),
			SnapshotId:     snapshotID,
			SourceVolumeId: req.GetSourceVolumeId(),
			CreationTime:   snapInfo.GetCtime(),
			ReadyToUse:     isSnapshotReady(snapInfo),
		},
	}, nil
}
func (s *OsdCsiServer) getCloudBackupClient(ctx context.Context) (api.OpenStorageCloudBackupClient, error) {
	// Get grpc connection
	conn, err := s.getRemoteConn(ctx)
	if err != nil {
		return nil, status.Errorf(
			codes.Unavailable,
			"Unable to connect to SDK server: %v", err)
	}
	return s.cloudBackupClient(conn), nil
}

func (s *OsdCsiServer) createCloudBackup(
	ctx context.Context,
	req *csi.CreateSnapshotRequest,
) (*csi.CreateSnapshotResponse, error) {
	cloudBackupClient, err := s.getCloudBackupClient(ctx)
	if err != nil {
		return nil, err
	}

	// Get any labels passed in by the CO
	spec, locator, _, err := s.specHandler.SpecFromOpts(req.GetParameters())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to get parameters: %v", err)
	}

	if spec.IsPureVolume() {
		return nil, status.Errorf(codes.InvalidArgument, "cloudsnap feature is not supported for pure volumes")
	}

	credentialID := locator.VolumeLabels[osdSnapshotCredentialIDKey]
	backupID := req.GetName()
	// Create snapshot
	_, err = cloudBackupClient.Create(ctx, &api.SdkCloudBackupCreateRequest{
		VolumeId:     req.GetSourceVolumeId(),
		TaskId:       backupID,
		CredentialId: credentialID,
		Labels:       locator.GetVolumeLabels(),
	})

	if err != nil {
		return nil, status.Errorf(codes.Aborted, "Failed to create cloud snapshot: %v", err)
	}

	var isBackupReady bool
	var backupStatus *api.SdkCloudBackupStatusResponse

	// Check if snapshot has been created but is in error state
	backupStatus, errFindFailed := cloudBackupClient.Status(ctx, &api.SdkCloudBackupStatusRequest{
		VolumeId: req.GetSourceVolumeId(),
		TaskId:   backupID,
	})
	if errFindFailed != nil {
		return nil, status.Errorf(codes.Aborted, "Failed to create cloud snapshot: %v", err)
	}
	isBackupReady = backupStatus.Statuses[backupID].Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone

	snapSize, errSizeFailed := cloudBackupClient.Size(ctx, &api.SdkCloudBackupSizeRequest{
		BackupId: backupID,
	})
	if errSizeFailed != nil {
		return nil, status.Errorf(codes.Aborted, "Failed to get cloud snapshot size: %v", err)
	}

	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SizeBytes:      int64(snapSize.GetTotalDownloadBytes()),
			SnapshotId:     backupID,
			SourceVolumeId: req.GetSourceVolumeId(),
			CreationTime:   backupStatus.Statuses[backupID].StartTime,
			ReadyToUse:     isBackupReady,
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
			codes.Unavailable,
			"Unable to connect to SDK server: %v", err)
	}

	// Get secret if any was passed
	ctx = s.setupContext(ctx, req.GetSecrets())
	ctx, cancel := grpcutil.WithDefaultTimeout(ctx)
	defer cancel()

	// Check ID is valid with the specified volume capabilities
	volumes := api.NewOpenStorageVolumeClient(conn)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: req.GetSnapshotId(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "Unable to delete snapshot %s: %v",
			req.GetSnapshotId(),
			err)
	}

	return &csi.DeleteSnapshotResponse{}, nil
}

// ListSnapshots lists all snapshots in a cluster.
// This is mainly implemented for Nomad, as Kubernetes will not call
// list snapshots for drivers which have synchronous snapshot creation.
// This is because ReadyToUse is set to true immediately upon on our snapshot creation.
func (s *OsdCsiServer) ListSnapshots(
	ctx context.Context,
	req *csi.ListSnapshotsRequest,
) (*csi.ListSnapshotsResponse, error) {

	if len(req.GetSnapshotId()) > 0 {
		return s.listSingleSnapshot(ctx, req)
	}

	return s.listMultipleSnapshots(ctx, req)
}

func (s *OsdCsiServer) listSingleSnapshot(
	ctx context.Context,
	req *csi.ListSnapshotsRequest,
) (*csi.ListSnapshotsResponse, error) {
	snapshotId := req.GetSnapshotId()

	clogger.WithContext(ctx).Infof("ListSnapshots for a single snapshot %s received", req.GetSnapshotId())

	// Get grpc connection
	conn, err := s.getConn()
	if err != nil {
		return nil, status.Errorf(
			codes.Unavailable,
			"Unable to connect to SDK server: %v", err)
	}

	// Get secret if any was passed
	ctx = s.setupContext(ctx, req.GetSecrets())
	ctx, cancel := grpcutil.WithDefaultTimeout(ctx)
	defer cancel()
	volumes := api.NewOpenStorageVolumeClient(conn)

	sdkRequest := &api.SdkVolumeInspectRequest{
		VolumeId: req.GetSnapshotId(),
	}

	// Snapshots are treated as volume, inspect for this volume.
	resp, err := volumes.Inspect(ctx, sdkRequest)
	if err != nil {
		status, ok := status.FromError(err)
		if ok && status.Code() == codes.NotFound {
			return &csi.ListSnapshotsResponse{
				Entries: []*csi.ListSnapshotsResponse_Entry{},
			}, nil
		}

		return nil, err
	}

	snap := csi.Snapshot{
		SizeBytes:      int64(resp.Volume.GetSpec().GetSize()),
		SnapshotId:     snapshotId,
		SourceVolumeId: resp.Volume.GetSource().Parent,
		CreationTime:   resp.Volume.Ctime,
		ReadyToUse:     isSnapshotReady(resp.Volume),
	}

	return &csi.ListSnapshotsResponse{
		Entries: []*csi.ListSnapshotsResponse_Entry{
			{
				Snapshot: &snap,
			},
		},
	}, nil
}

// listMultipleSnapshots gets all snapshots for a given
// volumeId or all snapshots if no volumeId is provided.
// This involves calling both InspectWithFilters and SnapshotEnumerateWithFilters
// as SnapshotEnumerateWithFilters only returns all snapshotIds.
// TODO: optimize this function if a SnapshotInspectWithFilters with pagination is implemented.
func (s *OsdCsiServer) listMultipleSnapshots(
	ctx context.Context,
	req *csi.ListSnapshotsRequest,
) (*csi.ListSnapshotsResponse, error) {
	sourceVolumeId := req.GetSourceVolumeId()
	startingToken := req.GetStartingToken()
	maxEntries := req.GetMaxEntries()

	clogger.WithContext(ctx).Infof("ListSnapshots for multiple snapshots received. sourceVolumeId: %s, startingToken: %s, maxEntries: %v",
		sourceVolumeId,
		startingToken,
		maxEntries,
	)

	// Get grpc connection
	conn, err := s.getConn()
	if err != nil {
		return nil, status.Errorf(
			codes.Unavailable,
			"Unable to connect to SDK server: %v", err)
	}

	// Get secret if any was passed
	ctx = s.setupContext(ctx, req.GetSecrets())
	ctx, cancel := grpcutil.WithDefaultTimeout(ctx)
	defer cancel()
	volumes := api.NewOpenStorageVolumeClient(conn)

	// Get all SnapshotIDs. Filter by source ID if provided.
	snapshotsReq := &api.SdkVolumeSnapshotEnumerateWithFiltersRequest{
		VolumeId: sourceVolumeId,
	}
	snapshotsResp, err := volumes.SnapshotEnumerateWithFilters(ctx, snapshotsReq)
	if err != nil {
		errStatus, ok := status.FromError(err)
		if ok && errStatus.Code() == codes.NotFound {
			return &csi.ListSnapshotsResponse{}, nil
		}
		return nil, status.Errorf(codes.Internal, "Unable to get all snapshots: %v", err)
	}

	// InspectWithFilters for all volumes
	volumesResp, err := volumes.InspectWithFilters(ctx, &api.SdkVolumeInspectWithFiltersRequest{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to get all volumes: %v", err)
	}

	// Sort snapshot IDs for repeatable results
	sortedSnapshotIds := sort.StringSlice(snapshotsResp.VolumeSnapshotIds)
	sort.Sort(sortedSnapshotIds)

	// Keep track of which volumes are snapshots
	volumeForSnapId := make(map[string]*api.Volume)
	for _, volResp := range volumesResp.Volumes {
		for _, snapId := range sortedSnapshotIds {
			if volResp.Volume.Id == snapId {
				volumeForSnapId[snapId] = volResp.Volume
			}
		}
	}

	// Generate response for all snapshots.
	listSnapshotsResp := &csi.ListSnapshotsResponse{}

	// If starting token is provided, start skipping entries
	// until we hit the starting token.
	var skipEntries bool
	if len(startingToken) > 0 {
		skipEntries = true
	}
	for _, snapId := range sortedSnapshotIds {
		// Skip entries until we hit the starting token.
		if skipEntries && startingToken != snapId {
			continue
		}
		skipEntries = false

		// Before adding new object to response, check if we're at the max entries.
		// If we are at max entries, return with current iteration as NextToken.
		// This allows for calls to ListSnapshots to begin where we left off.
		vol := volumeForSnapId[snapId]
		if maxEntries > 0 && len(listSnapshotsResp.Entries) >= int(maxEntries) {
			listSnapshotsResp.NextToken = vol.Id
			return listSnapshotsResp, nil
		}

		// Populate entry with volume info
		entry := &csi.ListSnapshotsResponse_Entry{
			Snapshot: &csi.Snapshot{
				SizeBytes:      int64(vol.GetSpec().GetSize()),
				SnapshotId:     vol.Id,
				SourceVolumeId: vol.GetSource().Parent,
				CreationTime:   vol.Ctime,
				ReadyToUse:     isSnapshotReady(vol),
			},
		}

		listSnapshotsResp.Entries = append(listSnapshotsResp.Entries, entry)
	}

	return listSnapshotsResp, nil
}

func getAllTopologies(req *csi.TopologyRequirement) []*csi.Topology {
	if req == nil {
		return nil
	}
	result := []*csi.Topology{}
	inputTopologies := append([]*csi.Topology{}, req.Preferred...)

	// Dedup topologies between Preferred and Requisite
	for _, requisite := range req.Requisite {
		matched := false
		for _, preferred := range req.Preferred {
			if reflect.DeepEqual(requisite, preferred) {
				matched = true
				break
			}
		}
		if !matched {
			inputTopologies = append(inputTopologies, requisite)
		}
	}

	for _, t := range inputTopologies {
		segments := map[string]string{}
		for key, value := range t.Segments {
			segments[key] = value
		}
		if len(segments) > 0 {
			result = append(result, &csi.Topology{Segments: segments})
		}
	}
	return result
}

// GetCapacity returns the available capacity for this node
func (s *OsdCsiServer) GetCapacity(context.Context, *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	clus, err := s.cluster.Enumerate()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to Enumerate cluster: %s", err)
	}

	node, err := s.cluster.Inspect(clus.NodeId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to Inspect node %s: %s", clus.NodeId, err)
	}

	var totalAvailable uint64 = 0
	for i := range node.Pools {
		available := node.Pools[i].GetTotalSize() - node.Pools[i].GetUsed()
		totalAvailable = available + totalAvailable
	}

	return &csi.GetCapacityResponse{
		AvailableCapacity: int64(totalAvailable),
	}, nil
}

// roundUpToNearestGiB rounds up given quantity upto chunks of GiB
func roundUpToNearestGiB(sizeBytes int64) (uint64, error) {
	sizeGiB, err := roundUpSizeInt64(*resource.NewQuantity(sizeBytes, resource.BinarySI), units.GiB)
	if err != nil {
		return 0, err
	}
	sizeGiBRounded := uint64(sizeGiB) * units.GiB

	return sizeGiBRounded, nil
}

// roundUpSizeInt64 calculates how many allocation units are needed to accommodate
// a volume of a given size. It returns an int64 and an error if there's overflow
// Borrowed from https://github.com/kubernetes/cloud-provider/blob/master/volume/helpers/rounding.go
// until we migrate to Go Modules and can pull this directly.
// TODO: During go modules migration, move to direct import of this function.
func roundUpSizeInt64(size resource.Quantity, allocationUnitBytes int64) (int64, error) {
	// Use CmpInt64() to find out if the value of "size" would overflow an
	// int64 and therefore have Value() return a wrong result. Then, retrieve
	// the value as int64 and perform the rounding.
	// It's not convenient to use AsScale() and related functions as they don't
	// support BinarySI format, nor can we use AsInt64() directly since it's
	// only implemented for int64 scaled numbers (int64Amount).

	// CmpInt64() actually returns 0 when comparing an amount bigger than MaxInt64.
	if size.CmpInt64(math.MaxInt64) >= 0 {
		return 0, fmt.Errorf("quantity %s is too great, overflows int64", size.String())
	}
	volumeSizeBytes := size.Value()

	roundedUp := volumeSizeBytes / allocationUnitBytes
	if volumeSizeBytes%allocationUnitBytes > 0 {
		roundedUp++
	}
	return roundedUp, nil
}
