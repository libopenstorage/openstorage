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

	"github.com/libopenstorage/openstorage/api"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.pedge.io/dlog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	volumeCapabilityMessageMultinodeVolume    = "Volume is a multinode volume"
	volumeCapabilityMessageNotMultinodeVolume = "Volume is not a multinode volume"
	volumeCapabilityMessageReadOnlyVolume     = "Volume is read only"
	volumeCapabilityMessageNotReadOnlyVolume  = "Volume is not read only"
)

// ControllerGetCapabilities is a CSI API functions which returns to the caller
// the capabilities of the OSD CSI driver.
func (s *OsdCsiServer) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest,
) (*csi.ControllerGetCapabilitiesResponse, error) {

	// Creating and deleting volumes supported
	capCreateDeleteVolume := &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			},
		},
	}

	// ListVolumes supported
	capListVolumes := &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
			},
		},
	}

	return &csi.ControllerGetCapabilitiesResponse{
		Reply: &csi.ControllerGetCapabilitiesResponse_Result_{
			Result: &csi.ControllerGetCapabilitiesResponse_Result{
				Capabilities: []*csi.ControllerServiceCapability{
					capCreateDeleteVolume,
					capListVolumes,
				},
			},
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
// Note: The method used here to return errors is still not part of the spec.
// See: https://github.com/container-storage-interface/spec/pull/115
// Discussion:  https://groups.google.com/forum/#!topic/kubernetes-sig-storage-wg-csi/TpTrNFbRa1I
//
func (s *OsdCsiServer) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest,
) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	// Probably we may use version in the future, but for now, let's just log it
	version := req.GetVersion()
	if version == nil {
		return nil, status.Error(codes.InvalidArgument, "Version must be specified")
	}
	capabilities := req.GetVolumeCapabilities()
	if capabilities == nil || len(capabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume_capabilities must be specified")
	}
	id := req.GetVolumeId()
	if len(id) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume_id must be specified")
	}
	attributes := req.GetVolumeAttributes()

	// Log request
	dlog.Debugf("ValidateVolumeCapabilities of id %s "+
		"capabilities %#v "+
		"version %#v "+
		"attributes %#v ",
		id,
		capabilities,
		version,
		attributes)

	// Check ID is valid with the specified volume capabilities
	volumes, err := s.driver.Inspect([]string{id})
	if err != nil || len(volumes) == 0 {
		return nil, status.Error(codes.NotFound, "ID not found")
	}
	if len(volumes) != 1 {
		errs := fmt.Sprintf(
			"Driver returned an unexpected number of volumes when one was expected: %d",
			len(volumes))
		dlog.Errorln(errs)
		return nil, status.Error(codes.Internal, errs)
	}
	v := volumes[0]
	if v.Id != id {
		errs := fmt.Sprintf(
			"Driver volume id [%s] does not equal requested id of: %s",
			v.Id,
			id)
		dlog.Errorln(errs)
		return nil, status.Error(codes.Internal, errs)
	}

	// Setup uninitialized response object
	result := &csi.ValidateVolumeCapabilitiesResponse_Result{
		Supported: true,
	}
	resp := &csi.ValidateVolumeCapabilitiesResponse{
		Reply: &csi.ValidateVolumeCapabilitiesResponse_Result_{
			Result: result,
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
				result.Supported = false
				result.Message = volumeCapabilityMessageMultinodeVolume
				break
			}
			if v.Readonly {
				result.Supported = false
				result.Message = volumeCapabilityMessageReadOnlyVolume
				break
			}
		case mode.Mode == csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY:
			if v.Spec.Shared {
				result.Supported = false
				result.Message = volumeCapabilityMessageMultinodeVolume
				break
			}
			if !v.Readonly {
				result.Supported = false
				result.Message = volumeCapabilityMessageNotReadOnlyVolume
				break
			}
		case mode.Mode == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY:
			if !v.Spec.Shared {
				result.Supported = false
				result.Message = volumeCapabilityMessageNotMultinodeVolume
				break
			}
			if !v.Readonly {
				result.Supported = false
				result.Message = volumeCapabilityMessageNotReadOnlyVolume
				break
			}
		case mode.Mode == csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER ||
			mode.Mode == csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER:
			if !v.Spec.Shared {
				result.Supported = false
				result.Message = volumeCapabilityMessageNotMultinodeVolume
				break
			}
			if v.Readonly {
				result.Supported = false
				result.Message = volumeCapabilityMessageReadOnlyVolume
				break
			}
		default:
			return nil, status.Errorf(
				codes.InvalidArgument,
				"AccessMode %s is not allowed",
				mode.Mode.String())
		}

		if !result.Supported {
			return resp, nil
		}
	}

	// If we passed all the checks, then it is valid
	result.Message = "Volume is supported"
	return resp, nil
}

// ListVolumes is a CSI API which returns to the caller all volume ids
// on this cluster. This includes ids created by CSI and ids created
// using other interfaces. This is important because the user could
// be requesting to mount a OSD volume created using non-CSI interfaces.
//
// This call does not yet implement tokens to due to the following
// issue: https://github.com/container-storage-interface/spec/issues/138
func (s *OsdCsiServer) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest,
) (*csi.ListVolumesResponse, error) {

	// Future: Once CSI is released, check version
	// for now, just log it.
	dlog.Debugf("ListVolumes req[%#v]", req)

	// Check arguments
	if req.GetVersion() == nil {
		return nil, status.Error(codes.InvalidArgument, "Version must be provided")
	}

	// Until the issue #138 on the CSI spec is resolved we will not support
	// tokenization
	if req.GetMaxEntries() != 0 {
		return nil, status.Error(
			codes.Unimplemented,
			"Driver does not support tokenization. Please see "+
				"https://github.com/container-storage-interface/spec/issues/138")
	}

	volumes, err := s.driver.Enumerate(&api.VolumeLocator{}, nil)
	if err != nil {
		errs := fmt.Sprintf("Unable to get list of volumes: %s", err.Error())
		dlog.Errorln(errs)
		return nil, status.Error(codes.Internal, errs)
	}
	entries := make([]*csi.ListVolumesResponse_Result_Entry, len(volumes))
	for i, v := range volumes {
		// Initialize entry
		entries[i] = &csi.ListVolumesResponse_Result_Entry{
			VolumeInfo: &csi.VolumeInfo{},
		}

		// Required
		entries[i].VolumeInfo.Id = v.Id

		// This entry is optional in the API, but OSD has
		// the information available to provide it
		entries[i].VolumeInfo.CapacityBytes = v.Spec.Size

		// Attributes. We can add or remove as needed since they
		// are optional and opaque to the Container Orchestrator(CO)
		// but could be used for debugging using a csi complient client.
		entries[i].VolumeInfo.Attributes = osdVolumeAttributes(v)
	}

	return &csi.ListVolumesResponse{
		Reply: &csi.ListVolumesResponse_Result_{
			Result: &csi.ListVolumesResponse_Result{
				Entries: entries,
			},
		},
	}, nil
}

// osdVolumeAttributes returns the attributes of a volume as a map
// to be returned to the CSI API caller
func osdVolumeAttributes(v *api.Volume) map[string]string {
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

/*
For next patches what still needs to be worked on in the Conroller server:

	CreateVolume(context.Context, *CreateVolumeRequest) (*CreateVolumeResponse, error)
	DeleteVolume(context.Context, *DeleteVolumeRequest) (*DeleteVolumeResponse, error)
	GetCapacity(context.Context, *GetCapacityRequest) (*GetCapacityResponse, error)
	ControllerGetCapabilities(context.Context, *ControllerGetCapabilitiesRequest) (*ControllerGetCapabilitiesResponse, error)
*/
