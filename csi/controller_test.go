/*
CSI Interface for OSD
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
	"testing"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/libopenstorage/openstorage/api"
	authsecrets "github.com/libopenstorage/openstorage/pkg/auth/secrets"
	"github.com/libopenstorage/openstorage/pkg/units"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func containsCap(c csi.ControllerServiceCapability_RPC_Type, resp *csi.ControllerGetCapabilitiesResponse) bool {
	for _, capability := range resp.GetCapabilities() {
		if rpc := capability.GetRpc(); rpc != nil {
			if rpc.GetType() == c {
				return true
			}
		}
	}
	return false
}

func getDefaultVolumeCapabilities(t *testing.T) []*csi.VolumeCapability {
	return []*csi.VolumeCapability{
		&csi.VolumeCapability{
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{
					FsType: "ext4",
				},
			},
		},
	}
}

func TestControllerGetCapabilities(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewControllerClient(s.Conn())
	resp, err := c.ControllerGetCapabilities(context.Background(), &csi.ControllerGetCapabilitiesRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Len(t, resp.GetCapabilities(), 7)
	assert.True(t, containsCap(csi.ControllerServiceCapability_RPC_GET_VOLUME, resp))
	assert.True(t, containsCap(csi.ControllerServiceCapability_RPC_CLONE_VOLUME, resp))
	assert.True(t, containsCap(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME, resp))
	assert.True(t, containsCap(csi.ControllerServiceCapability_RPC_EXPAND_VOLUME, resp))
	assert.True(t, containsCap(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT, resp))
	assert.True(t, containsCap(csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS, resp))
	assert.True(t, containsCap(csi.ControllerServiceCapability_RPC_VOLUME_CONDITION, resp))

	assert.False(t, containsCap(csi.ControllerServiceCapability_RPC_UNKNOWN, resp))
}

func TestControllerGetVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	size := int64(4 * 1024 * 1024)
	used := int64(1 * 1024 * 1024)
	id := "myvol123"
	vol := &api.Volume{
		Id: id,
		Locator: &api.VolumeLocator{
			Name: id,
		},
		Spec: &api.VolumeSpec{
			Size: uint64(size),
		},
		Usage:  uint64(used),
		Status: api.VolumeStatus_VOLUME_STATUS_UP,
	}
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				vol,
			}, nil).
			AnyTimes(),
	)

	// Make a call
	c := csi.NewControllerClient(s.Conn())

	// Get Capabilities - all OK
	resp, err := c.ControllerGetVolume(
		context.Background(),
		&csi.ControllerGetVolumeRequest{VolumeId: id})
	assert.NoError(t, err)
	assert.Equal(t, false, resp.Status.VolumeCondition.Abnormal)
	assert.Equal(t, "Volume status is up", resp.Status.VolumeCondition.Message)

	// Get Capabilities - down
	vol.Status = api.VolumeStatus_VOLUME_STATUS_DOWN
	resp, err = c.ControllerGetVolume(
		context.Background(),
		&csi.ControllerGetVolumeRequest{VolumeId: id})
	assert.NoError(t, err)
	assert.Equal(t, true, resp.Status.VolumeCondition.Abnormal)
	assert.Equal(t, "Volume status is down", resp.Status.VolumeCondition.Message)

	// Get Capabilities - degraded
	vol.Status = api.VolumeStatus_VOLUME_STATUS_DEGRADED
	resp, err = c.ControllerGetVolume(
		context.Background(),
		&csi.ControllerGetVolumeRequest{VolumeId: id})
	assert.NoError(t, err)
	assert.Equal(t, true, resp.Status.VolumeCondition.Abnormal)
	assert.Equal(t, "Volume status is degraded", resp.Status.VolumeCondition.Message)

	// Get Capabilities - none
	vol.Status = api.VolumeStatus_VOLUME_STATUS_NONE
	resp, err = c.ControllerGetVolume(
		context.Background(),
		&csi.ControllerGetVolumeRequest{VolumeId: id})
	assert.NoError(t, err)
	assert.Equal(t, true, resp.Status.VolumeCondition.Abnormal)
	assert.Equal(t, "Volume status is unknown", resp.Status.VolumeCondition.Message)

	// Get Capabilities - not present
	vol.Status = api.VolumeStatus_VOLUME_STATUS_NOT_PRESENT
	resp, err = c.ControllerGetVolume(
		context.Background(),
		&csi.ControllerGetVolumeRequest{VolumeId: id})
	assert.NoError(t, err)
	assert.Equal(t, true, resp.Status.VolumeCondition.Abnormal)
	assert.Equal(t, "Volume status is not present", resp.Status.VolumeCondition.Message)
}

func TestControllerPublishVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewControllerClient(s.Conn())
	_, err := c.ControllerPublishVolume(context.Background(), &csi.ControllerPublishVolumeRequest{})
	assert.NotNil(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.Unimplemented)
	assert.Contains(t, serverError.Message(), "not supported")
}

func TestControllerUnPublishVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewControllerClient(s.Conn())
	_, err := c.ControllerUnpublishVolume(context.Background(), &csi.ControllerUnpublishVolumeRequest{})
	assert.NotNil(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.Unimplemented)
	assert.Contains(t, serverError.Message(), "not supported")
}

func TestControllerValidateVolumeCapabilitiesBadArguments(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &csi.ValidateVolumeCapabilitiesRequest{}

	// Miss capabilities and id
	c := csi.NewControllerClient(s.Conn())
	_, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "capabilities")

	// Miss id and capabilities len is 0
	req.VolumeCapabilities = []*csi.VolumeCapability{}
	_, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "capabilities")

	// Miss id
	req.VolumeCapabilities = []*csi.VolumeCapability{
		&csi.VolumeCapability{},
	}
	_, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "volume_id")
}

func TestControllerValidateVolumeInvalidId(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"
	gomock.InOrder(
		// First time called it will say it is not there
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{}, nil),

		// Second time called it will not return an error,
		// but return an empty list
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{}, nil),

		// Third time it is called, it will return
		// a good list with a volume with an id that does
		// not match (even if this probably never could happen)
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: "bad volume id",
				},
			}, nil),
	)

	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		VolumeId: id,
		Secrets:  map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Missing everything
	c := csi.NewControllerClient(s.Conn())
	_, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "ID not found")

	// Send again, same result
	_, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "ID not found")

	// Now it should be an internal id error
	_, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Driver volume id")

	serverError, ok = status.FromError(err)
	assert.True(t, ok)
}

func TestControllerValidateVolumeInvalidCapabilities(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	c := csi.NewControllerClient(s.Conn())
	id := "testvolumeid"
	s.MockDriver().
		EXPECT().
		Enumerate(&api.VolumeLocator{
			VolumeIds: []string{id},
		}, nil).
		Return([]*api.Volume{
			&api.Volume{
				Id: id,
			},
		}, nil).
		Times(1)

	// Setup validate request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		VolumeId: id,
		Secrets:  map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Make request
	_, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Cannot have both")
}

func TestControllerValidateVolumeAccessModeSNWR(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"

	// RO SH
	// x   x
	// *   x
	// x   *
	// *   *
	gomock.InOrder(
		// not-RO and not-SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Sharedv4: false,
					},
				},
			}, nil),

		// RO and not-SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Sharedv4: false,
					},
				},
			}, nil),

		// not-RO and SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Sharedv4: true,
					},
				},
			}, nil),

		// RO and SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Sharedv4: true,
					},
				},
			}, nil),
	)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
			},
		},
		VolumeId: id,
		Secrets:  map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r.GetConfirmed())

	// Expect RO and non-SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.Nil(t, r.GetConfirmed())

	// Expect non-RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.Nil(t, r.GetConfirmed())

	// Expect RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.Nil(t, r.GetConfirmed())
}

func TestControllerValidateVolumeAccessModeSNRO(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"

	// RO SH
	// x   x
	// *   x
	// x   *
	// *   *
	gomock.InOrder(
		// not-RO and not-SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Sharedv4: false,
					},
				},
			}, nil),

		// RO and not-SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Sharedv4: false,
					},
				},
			}, nil),

		// not-RO and SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Sharedv4: true,
					},
				},
			}, nil),

		// RO and SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Sharedv4: true,
					},
				},
			}, nil),
	)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
				},
			},
		},
		VolumeId: id,
		Secrets:  map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.Nil(t, r.GetConfirmed())

	// Expect RO and non-SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r.GetConfirmed())

	// Expect non-RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.Nil(t, r.GetConfirmed())

	// Expect RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.Nil(t, r.GetConfirmed())
}

func TestControllerValidateVolumeAccessModeMNRO(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"

	// RO SH
	// x   x
	// *   x
	// x   *
	// *   *
	gomock.InOrder(
		// not-RO and not-SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Sharedv4: false,
					},
				},
			}, nil),

		// RO and not-SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Sharedv4: false,
					},
				},
			}, nil),

		// not-RO and SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Sharedv4: true,
					},
				},
			}, nil),

		// RO and SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Sharedv4: true,
					},
				},
			}, nil),
	)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
				},
			},
		},
		VolumeId: id,
		Secrets:  map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.Nil(t, r.GetConfirmed())

	// Expect RO and non-SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.Nil(t, r.GetConfirmed())

	// Expect non-RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.Nil(t, r.GetConfirmed())

	// Expect RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r.GetConfirmed())
}

func TestControllerValidateVolumeAccessModeMNWR(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"

	// RO SH
	// x   x
	// *   x
	// x   *
	// *   *
	gomock.InOrder(
		// not-RO and not-SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Sharedv4: false,
					},
				},
			}, nil),

		// RO and not-SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Sharedv4: false,
					},
				},
			}, nil),

		// not-RO and SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Sharedv4: true,
					},
				},
			}, nil),

		// RO and SH
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Sharedv4: true,
					},
				},
			}, nil),
	)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
				},
			},
		},
		VolumeId: id,
		Secrets:  map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.Nil(t, r.GetConfirmed())

	// Expect RO and non-SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.Nil(t, r.GetConfirmed())

	// Expect non-RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r.GetConfirmed())

	// Expect RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.Nil(t, r.GetConfirmed())
}

func TestControllerValidateVolumeAccessModeUnknown(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"
	s.MockDriver().
		EXPECT().
		Enumerate(&api.VolumeLocator{
			VolumeIds: []string{id},
		}, nil).
		Return([]*api.Volume{
			&api.Volume{
				Id:       id,
				Readonly: false,
				Spec: &api.VolumeSpec{
					Sharedv4: false,
				},
			},
		}, nil).
		Times(1)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_UNKNOWN,
				},
			},
		},
		VolumeId: id,
		Secrets:  map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	_, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)
}

func TestControllerCreateVolumeInvalidArguments(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	req := &csi.CreateVolumeRequest{}

	// No version
	_, err := c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Name")

	// No volume capabilities
	name := "myname"
	req.Name = name
	_, err = c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Volume capabilities")

	// Zero volume capabilities
	req.VolumeCapabilities = []*csi.VolumeCapability{}
	_, err = c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Volume capabilities")

}

func TestControllerCreateVolumeFoundByVolumeFromNameConflict(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	// Setup request
	tests := []struct {
		name      string
		req       *csi.CreateVolumeRequest
		ret       *api.Volume
		mockCalls []*gomock.Call
	}{
		{
			name: "size",
			req: &csi.CreateVolumeRequest{
				Name:               "size",
				VolumeCapabilities: getDefaultVolumeCapabilities(t),
				CapacityRange: &csi.CapacityRange{

					// Requested size does not match volume size
					RequiredBytes: 1000,
				},
				Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
			},
			mockCalls: []*gomock.Call{
				s.MockDriver().
					EXPECT().
					Inspect([]string{"size"}).
					Return(nil, fmt.Errorf("not found")).
					Times(1),

				s.MockDriver().
					EXPECT().
					Enumerate(&api.VolumeLocator{Name: "size"}, nil).
					Return([]*api.Volume{&api.Volume{
						Id: "size",
						Locator: &api.VolumeLocator{
							Name: "size",
						},
						Status: api.VolumeStatus_VOLUME_STATUS_UP,
						Spec: &api.VolumeSpec{

							// Size is different
							Size: 10,
						},
					}}, nil).
					Times(1),

				s.MockDriver().
					EXPECT().
					Inspect([]string{"size"}).
					Return(nil, fmt.Errorf("not found")).
					Times(1),

				s.MockDriver().
					EXPECT().
					Enumerate(&api.VolumeLocator{Name: "size"}, nil).
					Return([]*api.Volume{&api.Volume{
						Id: "size",
						Locator: &api.VolumeLocator{
							Name: "size",
						},
						Status: api.VolumeStatus_VOLUME_STATUS_UP,
						Spec: &api.VolumeSpec{

							// Size is different
							Size: 10,
						},
					}}, nil).
					Times(1),
			},
		},
	}

	for _, test := range tests {
		gomock.InOrder(test.mockCalls...)
		_, err := c.CreateVolume(context.Background(), test.req)
		assert.Error(t, err)
		serverError, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, serverError.Code())
	}
}

func TestControllerCreateVolumeNoCapacity(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	// Setup request
	name := "myvol"
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	id := "myid"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(
				ctx context.Context,
				locator *api.VolumeLocator,
				Source *api.Source,
				spec *api.VolumeSpec,
			) (string, error) {
				assert.Equal(t, spec.Size, defaultCSIVolumeSize)
				return id, nil
			}).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size:     defaultCSIVolumeSize,
						Sharedv4: true,
					},
				},
			}, nil).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolume()

	assert.Equal(t, id, volumeInfo.GetVolumeId())
	assert.Equal(t, int64(defaultCSIVolumeSize), volumeInfo.GetCapacityBytes())
}

func TestControllerCreateVolumeFoundByVolumeFromName(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	// Setup request
	name := "myvol"
	size := 2 * int64(units.GiB)
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Volume is already being created and found by calling VolumeFromName
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Status: api.VolumeStatus_VOLUME_STATUS_UP,
					Spec: &api.VolumeSpec{
						Size: uint64(size),
					},
				},
			}, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Status: api.VolumeStatus_VOLUME_STATUS_UP,
					Spec: &api.VolumeSpec{
						Size: uint64(size),
					},
				},
			}, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{name},
			}, nil).
			Return([]*api.Volume{&api.Volume{
				Id: name,
				Locator: &api.VolumeLocator{
					Name: name,
				},
				Status: api.VolumeStatus_VOLUME_STATUS_UP,
				Spec: &api.VolumeSpec{
					Size: uint64(size),
				},
			}}, nil).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolume()

	assert.Equal(t, name, volumeInfo.GetVolumeId())
	assert.Equal(t, size, volumeInfo.GetCapacityBytes())
}

func TestControllerCreateVolumeBadParameters(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	// Setup request
	name := "myvol"
	size := int64(1234)
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Parameters: map[string]string{
			api.SpecFilesystem: "whatkindoffsisthis?",
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	_, err := c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "get parameters")
}

func TestControllerCreateVolumeBadParentId(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	// Setup request
	name := "myvol"
	size := int64(1234)
	parent := "badid"
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Parameters: map[string]string{
			api.SpecParent: parent,
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Volume is already being created and found by calling VolumeFromName
	gomock.InOrder(
		// First check on parent
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{parent},
			}, nil).
			Return([]*api.Volume{&api.Volume{Id: parent}}, nil).
			Times(1),

		// VolFromName (name)
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(gomock.Any(), nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		// VolFromName (parent)
		s.MockDriver().
			EXPECT().
			Inspect([]string{parent}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),
		s.MockDriver().
			EXPECT().
			Enumerate(gomock.Any(), nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),
	)

	_, err := c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "unable to get parent volume information")
}

func TestControllerCreateVolumeBadSnapshot(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	// Setup request
	name := "myvol"
	size := int64(1234)
	parent := "parent"
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Parameters: map[string]string{
			api.SpecParent: parent,
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Volume is already being created and found by calling VolumeFromName
	gomock.InOrder(
		// First check on parent
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{parent},
			}, nil).
			Return([]*api.Volume{&api.Volume{Id: parent}}, nil).
			Times(1),

		// VolFromName (name)
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(gomock.Any(), nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		// VolFromName (parent)
		s.MockDriver().
			EXPECT().
			Inspect([]string{parent}).
			Return([]*api.Volume{&api.Volume{Id: parent}}, nil).
			Times(1),

		// Return an error from snapshot
		s.MockDriver().
			EXPECT().
			Snapshot(parent, false, &api.VolumeLocator{Name: name}, false).
			Return("", fmt.Errorf("snapshoterr")).
			Times(1),
	)

	_, err := c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "snapshoterr")
}

func TestControllerCreateVolumeWithSharedv4Volume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	modes := []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
		csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
	}

	for _, mode := range modes {
		// Setup request
		name := "myvol"
		size := int64(1234)
		req := &csi.CreateVolumeRequest{
			Name: name,
			VolumeCapabilities: []*csi.VolumeCapability{
				&csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: mode,
					},
				},
			},
			CapacityRange: &csi.CapacityRange{
				RequiredBytes: size,
			},
			Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
		}

		// Setup mock functions
		id := "myid"
		gomock.InOrder(
			s.MockDriver().
				EXPECT().
				Inspect([]string{name}).
				Return([]*api.Volume{}, nil).
				Times(1),

			s.MockDriver().
				EXPECT().
				Enumerate(&api.VolumeLocator{Name: name}, nil).
				Return(nil, fmt.Errorf("not found")).
				Times(1),

			s.MockDriver().
				EXPECT().
				Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(id, nil).
				Times(1),

			s.MockDriver().
				EXPECT().
				Enumerate(&api.VolumeLocator{
					VolumeIds: []string{id},
				}, nil).
				Return([]*api.Volume{
					&api.Volume{
						Id: id,
						Locator: &api.VolumeLocator{
							Name: name,
						},
						Spec: &api.VolumeSpec{
							Size:     uint64(size),
							Sharedv4: true,
						},
					},
				}, nil).
				Times(1),
		)

		r, err := c.CreateVolume(context.Background(), req)
		assert.Nil(t, err)
		assert.NotNil(t, r)
		volumeInfo := r.GetVolume()

		assert.Equal(t, id, volumeInfo.GetVolumeId())
		assert.Equal(t, size, volumeInfo.GetCapacityBytes())
		assert.Equal(t, "true", volumeInfo.GetVolumeContext()[api.SpecSharedv4])
	}
}
func TestControllerCreateVolumeWithSharedVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	modes := []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
		csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
	}

	for _, mode := range modes {
		// Setup request
		name := "myvol"
		size := int64(1234)
		req := &csi.CreateVolumeRequest{
			Name: name,
			VolumeCapabilities: []*csi.VolumeCapability{
				&csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: mode,
					},
				},
			},
			CapacityRange: &csi.CapacityRange{
				RequiredBytes: size,
			},
			Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
			Parameters: map[string]string{
				api.SpecShared: "true",
			},
		}

		// Setup mock functions
		id := "myid"
		gomock.InOrder(
			s.MockDriver().
				EXPECT().
				Inspect([]string{name}).
				Return([]*api.Volume{}, nil).
				Times(1),

			s.MockDriver().
				EXPECT().
				Enumerate(&api.VolumeLocator{Name: name}, nil).
				Return(nil, fmt.Errorf("not found")).
				Times(1),

			s.MockDriver().
				EXPECT().
				Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(id, nil).
				Times(1),

			s.MockDriver().
				EXPECT().
				Enumerate(&api.VolumeLocator{
					VolumeIds: []string{id},
				}, nil).
				Return([]*api.Volume{
					&api.Volume{
						Id: id,
						Locator: &api.VolumeLocator{
							Name: name,
						},
						Spec: &api.VolumeSpec{
							Size:   uint64(size),
							Shared: true,
						},
					},
				}, nil).
				Times(1),
		)

		r, err := c.CreateVolume(context.Background(), req)
		assert.Nil(t, err)
		assert.NotNil(t, r)
		volumeInfo := r.GetVolume()

		assert.Equal(t, id, volumeInfo.GetVolumeId())
		assert.Equal(t, size, volumeInfo.GetCapacityBytes())
		assert.Equal(t, "true", volumeInfo.GetVolumeContext()[api.SpecShared])
	}
}

func TestControllerCreateVolumeFails(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	// Setup request
	name := "myvol"
	size := int64(1234)
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Setup mock functions
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return("", fmt.Errorf("createerror")).
			Times(1),
	)

	_, err := c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "createerror")
}

func TestControllerCreateVolumeNoNewVolumeInfo(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	// Setup request
	name := "myvol"
	size := int64(1234)
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Setup mock functions
	id := "myid"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),
	)

	_, err := c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "not found")
}

func TestControllerCreateVolumeFailedRemoteConn(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	secretKeyForLabels := "key123"
	secretValForLabels := "val123"
	s.MockCluster().EXPECT().
		Enumerate().
		Return(api.Cluster{
			NodeId: "node-1",
			Nodes: []*api.Node{{
				Id:     "1",
				MgmtIp: "badip",
			}},
		}, nil).
		AnyTimes()

	// Setup request
	name := "myvol"
	size := int64(1234)
	secretsMap := map[string]string{
		authsecrets.SecretTokenKey: systemUserToken,
		secretKeyForLabels:         secretValForLabels,
	}
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Secrets: secretsMap,
	}

	// Setup mock functions
	id := "myid"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name:         name,
						VolumeLabels: secretsMap,
					},
					Spec: &api.VolumeSpec{
						Size: uint64(size),
					},
				},
			}, nil).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolume()

	assert.Equal(t, id, volumeInfo.GetVolumeId())
	assert.Equal(t, size, volumeInfo.GetCapacityBytes())
	assert.NotEqual(t, "true", volumeInfo.GetVolumeContext()[api.SpecSharedv4])
}

func TestControllerCreateVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	secretKeyForLabels := "key123"
	secretValForLabels := "val123"
	s.mockClusterEnumerateNode(t, "node-1")

	// Setup request
	name := "myvol"
	size := int64(1234)
	secretsMap := map[string]string{
		authsecrets.SecretTokenKey: systemUserToken,
		secretKeyForLabels:         secretValForLabels,
	}
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Secrets: secretsMap,
	}

	// Setup mock functions
	id := "myid"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name:         name,
						VolumeLabels: secretsMap,
					},
					Spec: &api.VolumeSpec{
						Size: uint64(size),
					},
				},
			}, nil).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolume()

	assert.Equal(t, id, volumeInfo.GetVolumeId())
	assert.Equal(t, size, volumeInfo.GetCapacityBytes())
	assert.NotEqual(t, "true", volumeInfo.GetVolumeContext()[api.SpecSharedv4])
}

func TestControllerCreateVolumeRoundUp(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	secretKeyForLabels := "key123"
	secretValForLabels := "val123"
	s.mockClusterEnumerateNode(t, "node-1")
	// Setup request
	name := "myvol"
	size := int64(units.GiB * 1.5)
	secretsMap := map[string]string{
		authsecrets.SecretTokenKey: systemUserToken,
		secretKeyForLabels:         secretValForLabels,
	}
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Secrets: secretsMap,
	}

	// Setup mock functions
	id := "myid"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any(), &api.VolumeSpec{
				Size:      uint64(units.GiB) * 2, // round up to 2GiB from 1.5 GiB
				Format:    api.FSType_FS_TYPE_EXT4,
				HaLevel:   1,
				IoProfile: api.IoProfile_IO_PROFILE_AUTO,
				Ownership: &api.Ownership{
					Owner: "user1",
				},
				Xattr: api.Xattr_COW_ON_DEMAND,
			}).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name:         name,
						VolumeLabels: secretsMap,
					},
					Spec: &api.VolumeSpec{
						Size: uint64(size),
					},
				},
			}, nil).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolume()

	assert.Equal(t, id, volumeInfo.GetVolumeId())
	assert.Equal(t, size, volumeInfo.GetCapacityBytes())
	assert.NotEqual(t, "true", volumeInfo.GetVolumeContext()[api.SpecSharedv4])
}

func TestControllerCreateVolumeFromSnapshot(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	// Setup request
	mockParentID := "parendId"
	name := "myvol"
	size := int64(1234)
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		VolumeContentSource: &csi.VolumeContentSource{
			Type: &csi.VolumeContentSource_Snapshot{
				Snapshot: &csi.VolumeContentSource_SnapshotSource{
					SnapshotId: mockParentID,
				},
			},
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Setup mock functions
	id := "myid"
	snapID := id + "-snap"
	gomock.InOrder(

		// First check on parent
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{mockParentID},
			}, nil).
			Return([]*api.Volume{&api.Volume{Id: mockParentID}}, nil).
			Times(1),

		// VolFromName (name)
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(gomock.Any(), nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		//VolFromName parent
		s.MockDriver().
			EXPECT().
			Inspect(gomock.Any()).
			Return(
				[]*api.Volume{&api.Volume{
					Id: mockParentID,
				}}, nil).
			Times(1),

		// create
		s.MockDriver().
			EXPECT().
			Snapshot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(snapID, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{snapID},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:     id,
					Source: &api.Source{Parent: mockParentID},
				},
			}, nil).
			Times(2),

		s.MockDriver().
			EXPECT().
			Set(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{snapID},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id:     id,
					Source: &api.Source{Parent: mockParentID},
				},
			}, nil).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolume()

	assert.Equal(t, id, volumeInfo.GetVolumeId())
	assert.NotEqual(t, "true", volumeInfo.GetVolumeContext()[api.SpecSharedv4])
	assert.Equal(t, mockParentID, volumeInfo.GetVolumeContext()[api.SpecParent])
}

func TestControllerCreateVolumeSnapshotThroughParameters(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	// Setup request
	mockParentID := "parendId"
	name := "myvol"
	size := int64(1234)
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Parameters: map[string]string{
			api.SpecParent: mockParentID,
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Setup mock functions
	id := "myid"
	gomock.InOrder(
		// first check on parent
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{mockParentID},
			}, nil).
			Return([]*api.Volume{&api.Volume{Id: mockParentID}}, nil).
			Times(1),

		//VolFromName name
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		// VolFromName parent
		s.MockDriver().
			EXPECT().
			Inspect([]string{mockParentID}).
			Return([]*api.Volume{
				&api.Volume{
					Id: mockParentID,
				},
			}, nil).
			Times(1),

		// create snap
		s.MockDriver().
			EXPECT().
			Snapshot(mockParentID, false, &api.VolumeLocator{
				Name: name,
			},
				false).
			Return(id, nil).
			Times(1),

		// check snap
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: uint64(size),
					},
					Source: &api.Source{
						Parent: mockParentID,
					},
				},
			}, nil).
			Times(1),

		// update - inspect and set
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: uint64(size),
					},
					Source: &api.Source{
						Parent: mockParentID,
					},
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Set(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1),
		// final inspect
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: uint64(size),
					},
					Source: &api.Source{
						Parent: mockParentID,
					},
				},
			}, nil).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolume()

	assert.Equal(t, id, volumeInfo.GetVolumeId())
	assert.Equal(t, size, volumeInfo.GetCapacityBytes())
	assert.NotEqual(t, "true", volumeInfo.GetVolumeContext()[api.SpecSharedv4])
	assert.Equal(t, mockParentID, volumeInfo.GetVolumeContext()[api.SpecParent])
}

func TestControllerCreateVolumeBlock(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	secretKeyForLabels := "key123"
	secretValForLabels := "val123"
	s.mockClusterEnumerateNode(t, "node-1")
	// Setup request
	name := "myvol"
	size := int64(1234)
	secretsMap := map[string]string{
		authsecrets.SecretTokenKey: systemUserToken,
		secretKeyForLabels:         secretValForLabels,
	}
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Block{
					Block: &csi.VolumeCapability_BlockVolume{},
				},
			},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Secrets: secretsMap,
	}

	// Setup mock functions
	id := "myid"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name:         name,
						VolumeLabels: secretsMap,
					},
					Spec: &api.VolumeSpec{
						Size: uint64(size),
					},
				},
			}, nil).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolume()

	assert.Equal(t, id, volumeInfo.GetVolumeId())
	assert.Equal(t, size, volumeInfo.GetCapacityBytes())
	assert.NotEqual(t, "true", volumeInfo.GetVolumeContext()[api.SpecSharedv4])
}

func TestControllerCreateVolumeBlockSharedInvalid(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	secretKeyForLabels := "key123"
	secretValForLabels := "val123"
	s.mockClusterEnumerateNode(t, "node-1") // Setup request
	name := "myvol"
	size := int64(1234)
	secretsMap := map[string]string{
		authsecrets.SecretTokenKey: systemUserToken,
		secretKeyForLabels:         secretValForLabels,
	}
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Block{
					Block: &csi.VolumeCapability_BlockVolume{},
				},
			},
			&csi.VolumeCapability{
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
				},
			},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Secrets: secretsMap,
	}

	// Setup mock functions
	_, err := c.CreateVolume(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Shared raw block volumes are not supported")

}

func TestControllerCreateVolumeWithoutTopology(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	name := "myvol"
	size := int64(1234)
	id := "myid"

	// TestCase: Topology requirement present, but not a Pure volume
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		AccessibilityRequirements: &csi.TopologyRequirement{
			Preferred: []*csi.Topology{
				{
					Segments: map[string]string{
						"topology.io/zone": "zone-1",
					},
				},
			},
		},
	}

	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, locator *api.VolumeLocator, source *api.Source, spec *api.VolumeSpec) (string, error) {
				assert.Nil(t, spec.TopologyRequirement)
				return id, nil
			}).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: uint64(size),
					},
				},
			}, nil).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolume()

	assert.Equal(t, id, volumeInfo.GetVolumeId())
	assert.Equal(t, size, volumeInfo.GetCapacityBytes())
	assert.Nil(t, volumeInfo.GetAccessibleTopology())

	// TestCase: Pure volume, but topology requirements absent
	req = &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Parameters: map[string]string{
			api.SpecBackendType: api.SpecBackendPureBlock,
		},
	}

	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, locator *api.VolumeLocator, source *api.Source, spec *api.VolumeSpec) (string, error) {
				assert.Nil(t, spec.TopologyRequirement)
				return id, nil
			}).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: uint64(size),
					},
				},
			}, nil).
			Times(1),
	)

	r, err = c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo = r.GetVolume()

	assert.Equal(t, id, volumeInfo.GetVolumeId())
	assert.Equal(t, size, volumeInfo.GetCapacityBytes())
}

func TestControllerCreateVolumeWithTopology(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())
	s.mockClusterEnumerateNode(t, "node-1")
	name := "myvol"
	size := int64(1234)
	id := "myid"

	// TestCase: Pure volume and topology requirement present, but Create fails.
	// This test validates that we don't retry Create if the failure happens because of
	// any other reason that topology placement.
	req := &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Parameters: map[string]string{
			api.SpecBackendType: api.SpecBackendPureFile,
		},
		AccessibilityRequirements: &csi.TopologyRequirement{
			Preferred: []*csi.Topology{
				{
					Segments: map[string]string{
						"topology.io/zone":   "zone-1",
						"topology.io/region": "region-1",
					},
				},
				{
					Segments: map[string]string{
						"topology.io/zone":   "zone-2",
						"topology.io/region": "region-2",
					},
				},
			},
		},
	}

	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, locator *api.VolumeLocator, source *api.Source, spec *api.VolumeSpec) (string, error) {
				assert.NotNil(t, spec.TopologyRequirement)
				assert.Equal(t, spec.TopologyRequirement.Labels, req.AccessibilityRequirements.Preferred[0].Segments)
				return id, nil
			}).
			Return("", fmt.Errorf("create failed")).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	assert.Nil(t, r)

	// TestCase: Pure volume and topology requirement present.
	// This tests mulitple things -
	// - Multiple topologies are sent by the provisioner in both preferred and requisite sections
	// - Topologies are de-duped from the requirement
	// - Retry volume creation only if the create fails because of topology placement
	req = &csi.CreateVolumeRequest{
		Name: name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Parameters: map[string]string{
			api.SpecBackendType: api.SpecBackendPureFile,
		},
		AccessibilityRequirements: &csi.TopologyRequirement{
			Preferred: []*csi.Topology{
				{
					Segments: map[string]string{
						"topology.io/zone":   "zone-1",
						"topology.io/region": "region-1",
					},
				},
			},
			Requisite: []*csi.Topology{
				{
					Segments: map[string]string{
						"topology.io/zone":   "zone-1",
						"topology.io/region": "region-1",
					},
				},
				{
					Segments: map[string]string{
						"topology.io/zone":   "zone-2",
						"topology.io/region": "region-2",
					},
				},
			},
		},
	}

	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, locator *api.VolumeLocator, source *api.Source, spec *api.VolumeSpec) (string, error) {
				assert.NotNil(t, spec.TopologyRequirement)
				assert.Equal(t, spec.TopologyRequirement.Labels, req.AccessibilityRequirements.Preferred[0].Segments)
				return id, nil
			}).
			Return("", fmt.Errorf("no storage backends were able to meet the request specification")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, locator *api.VolumeLocator, source *api.Source, spec *api.VolumeSpec) (string, error) {
				assert.NotNil(t, spec.TopologyRequirement)
				assert.Equal(t, spec.TopologyRequirement.Labels, req.AccessibilityRequirements.Requisite[1].Segments)
				return id, nil
			}).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: uint64(size),
						TopologyRequirement: &api.TopologyRequirement{
							Labels: req.AccessibilityRequirements.Requisite[1].Segments,
						},
					},
				},
			}, nil).
			Times(1),
	)

	r, err = c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolume()

	assert.Equal(t, id, volumeInfo.GetVolumeId())
	assert.Equal(t, size, volumeInfo.GetCapacityBytes())
	assert.Len(t, volumeInfo.AccessibleTopology, 1)
	assert.Equal(t, req.AccessibilityRequirements.Requisite[1].Segments, volumeInfo.AccessibleTopology[0].Segments)
}

func TestControllerDeleteVolumeInvalidArguments(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// No version
	req := &csi.DeleteVolumeRequest{}
	_, err := c.DeleteVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Volume id")
}

func TestControllerDeleteVolumeError(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// No version
	myid := "myid"
	req := &csi.DeleteVolumeRequest{
		VolumeId: myid,
		Secrets:  map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// Setup mock
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{myid},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: myid,
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Delete(gomock.Any(), myid).
			Return(fmt.Errorf("MOCKERRORTEST")).
			Times(1),
	)

	_, err := c.DeleteVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Aborted)
	assert.Contains(t, serverError.Message(), "Unable to delete")
	assert.Contains(t, serverError.Message(), "MOCKERRORTEST")
}

func TestControllerDeleteVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// No version
	myid := "myid"
	req := &csi.DeleteVolumeRequest{
		VolumeId: myid,
		Secrets:  map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	// According to CSI spec, if the ID is not found, it must return OK
	// Now return no error, but empty list
	s.MockDriver().
		EXPECT().
		Enumerate(&api.VolumeLocator{
			VolumeIds: []string{myid},
		}, nil).
		Return([]*api.Volume{}, nil).
		Times(1)

	_, err := c.DeleteVolume(context.Background(), req)
	assert.Nil(t, err)

	// Setup mock
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{myid},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: myid,
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Delete(gomock.Any(), myid).
			Return(nil).
			Times(1),
	)

	_, err = c.DeleteVolume(context.Background(), req)
	assert.Nil(t, err)
}

func TestControllerExpandVolumeBadParameter(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	_, err := c.ControllerExpandVolume(context.Background(), &csi.ControllerExpandVolumeRequest{})
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "id must be provided")

	_, err = c.ControllerExpandVolume(context.Background(), &csi.ControllerExpandVolumeRequest{
		VolumeId: "id",
	})
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Capacity range must be provided")

	_, err = c.ControllerExpandVolume(context.Background(), &csi.ControllerExpandVolumeRequest{
		VolumeId: "id",
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: int64(-5),
		},
	})
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "cannot be negative")

}

func TestControllerExpandVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	myid := "myid"
	vol := &api.Volume{
		Id: myid,
		Spec: &api.VolumeSpec{
			Size: uint64(units.GiB),
		},
	}
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{myid}).
			Return([]*api.Volume{
				vol,
			}, nil).
			AnyTimes(),
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{myid},
			}, nil).
			Return([]*api.Volume{vol}, nil).
			AnyTimes(),
		s.MockDriver().
			EXPECT().
			Set(gomock.Any(), gomock.Any(), &api.VolumeSpec{
				Size: 46 * units.GiB, // Round up from 45.5 to 46
			}).
			Return(nil).
			Times(1),
	)

	_, err := c.ControllerExpandVolume(context.Background(), &csi.ControllerExpandVolumeRequest{
		VolumeId: myid,
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: int64(45.5 * units.GiB),
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	})
	assert.NoError(t, err)
}

func TestControllerExpandVolumeIdempotent(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	myid := "myid"
	vol := &api.Volume{
		Id: myid,
		Spec: &api.VolumeSpec{
			Size: uint64(5 * units.GiB),
		},
	}
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{myid},
			}, nil).
			Return([]*api.Volume{vol}, nil).
			Times(1),
	)

	_, err := c.ControllerExpandVolume(context.Background(), &csi.ControllerExpandVolumeRequest{
		VolumeId: myid,
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: int64(5 * units.GiB),
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	})
	assert.NoError(t, err)
}
func TestControllerCreateSnapshotBadParameters(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	_, err := c.CreateSnapshot(context.Background(), &csi.CreateSnapshotRequest{})
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "id must be provided")

	_, err = c.CreateSnapshot(context.Background(), &csi.CreateSnapshotRequest{
		SourceVolumeId: "id",
	})
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Name must be provided")
}

func TestControllerCreateSnapshotIdempotent(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	name := "name"
	volume := "id"
	req := &csi.CreateSnapshotRequest{
		Name:           name,
		SourceVolumeId: volume,
		Secrets:        map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}
	size := int64(1024)
	snapInfo := &api.Volume{
		Id: name,
		Source: &api.Source{
			Parent: volume,
		},
		Ctime: ptypes.TimestampNow(),
		Locator: &api.VolumeLocator{
			Name: name,
		},
		Spec: &api.VolumeSpec{
			Size: uint64(size),
		},
	}

	// Snapshot already exists
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return([]*api.Volume{snapInfo}, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{name},
			}, nil).
			Return([]*api.Volume{snapInfo}, nil).
			Times(1),
	)

	r, err := c.CreateSnapshot(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, name, r.GetSnapshot().GetSnapshotId())
	assert.Equal(t, size, r.GetSnapshot().GetSizeBytes())
	assert.Equal(t, snapInfo.Source.Parent, r.GetSnapshot().GetSourceVolumeId())
}

func TestControllerCreateSnapshot(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	name := "name"
	volume := "id"
	req := &csi.CreateSnapshotRequest{
		Name:           name,
		SourceVolumeId: volume,
		Parameters: map[string]string{
			"labels": "hello=world",
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}
	size := uint64(1024)
	snapInfo := &api.Volume{
		Id: name,
		Source: &api.Source{
			Parent: volume,
		},
		Ctime: ptypes.TimestampNow(),
		Locator: &api.VolumeLocator{
			Name: name,
		},
		Spec: &api.VolumeSpec{
			Size: size,
		},
	}

	// Setup mock functions
	snapId := "myid"
	gomock.InOrder(

		// VolumeFromNameSdk
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		// check permissions
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{volume},
			}, nil).
			Return([]*api.Volume{&api.Volume{Id: volume, Spec: &api.VolumeSpec{
				Size: size,
			}}}, nil).
			Times(1),

		// snapshot
		s.MockDriver().
			EXPECT().
			Snapshot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(snapId, nil).
			Times(1),

		// VolumeFromIdSdk
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{snapId},
			}, nil).
			Return([]*api.Volume{snapInfo}, nil).
			Times(1),
	)

	r, err := c.CreateSnapshot(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, snapId, r.GetSnapshot().GetSnapshotId())
	assert.Equal(t, volume, r.GetSnapshot().GetSourceVolumeId())
}

func TestControllerDeleteSnapshotBadParameters(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	_, err := c.DeleteSnapshot(context.Background(), &csi.DeleteSnapshotRequest{
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	})
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "id must be provided")
}

func TestControllerDeleteSnapshotIdempotent(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	id := "id"
	// Snapshot already exists
	s.MockDriver().
		EXPECT().
		Enumerate(&api.VolumeLocator{
			VolumeIds: []string{id},
		}, nil).
		Return([]*api.Volume{}, nil).
		Times(1)

	_, err := c.DeleteSnapshot(context.Background(), &csi.DeleteSnapshotRequest{
		SnapshotId: id,
		Secrets:    map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	})
	assert.NoError(t, err)
}

func TestControllerDeleteSnapshot(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	id := "id"

	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{id},
			}, nil).
			Return([]*api.Volume{
				&api.Volume{},
			}, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Delete(gomock.Any(), id).
			Return(nil).
			Times(1),
	)

	_, err := c.DeleteSnapshot(context.Background(), &csi.DeleteSnapshotRequest{
		SnapshotId: id,
		Secrets:    map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	})
	assert.NoError(t, err)
}

func TestControllerListSnapshots(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	volId := "id"
	volName := "volName"
	ts := &timestamp.Timestamp{
		Seconds: 1620340488,
	}
	vol := &api.Volume{
		Id: volId,
		Locator: &api.VolumeLocator{
			Name: volName,
		},
		Spec: &api.VolumeSpec{
			Size: uint64(units.GiB),
		},
		Ctime: ts,
		Source: &api.Source{
			Parent: "srcVol",
		},
		Status: api.VolumeStatus_VOLUME_STATUS_UP,
	}
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{volId},
			}, nil).
			Return([]*api.Volume{vol}, nil).
			Times(1),
	)

	resp, err := c.ListSnapshots(context.Background(), &csi.ListSnapshotsRequest{
		SnapshotId: volId,
		Secrets:    map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	})
	assert.NoError(t, err)
	assert.Equal(t, true, resp.Entries[0].Snapshot.ReadyToUse)
	assert.Equal(t, int64(units.GiB), resp.Entries[0].Snapshot.SizeBytes)
	assert.Equal(t, ts.Seconds, resp.Entries[0].Snapshot.CreationTime.Seconds)
	assert.Equal(t, volId, resp.Entries[0].Snapshot.SnapshotId)
	assert.Equal(t, "srcVol", resp.Entries[0].Snapshot.SourceVolumeId)

	secondVol := &api.Volume{
		Id: "volId2",
		Locator: &api.VolumeLocator{
			Name: "volName2",
		},
		Spec: &api.VolumeSpec{
			Size: uint64(units.GiB),
		},
		Ctime: ts,
		Source: &api.Source{
			Parent: "srcVol",
		},
		Status: api.VolumeStatus_VOLUME_STATUS_UP,
	}
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			SnapEnumerate(nil, nil).
			Return([]*api.Volume{vol, secondVol}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Enumerate(nil, nil).
			Return([]*api.Volume{vol, secondVol}, nil).
			Times(1),
	)

	_, err = c.ListSnapshots(context.Background(), &csi.ListSnapshotsRequest{
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	})
	assert.NoError(t, err)
}

func TestGetPVCMetadata(t *testing.T) {
	params := make(map[string]string)
	params[osdPvcNameKey] = "mypvc1"
	params[osdPvcNamespaceKey] = "mypvcns1"

	labels := make(map[string]string)
	labels["testkey_labels"] = "testval_1"
	annotations := make(map[string]string)
	annotations["testkey_annotations"] = "testval_2"
	encodedLabels, err := json.Marshal(labels)
	assert.NoError(t, err)
	encodedAnnotations, err := json.Marshal(annotations)
	assert.NoError(t, err)

	params[osdPvcAnnotationsKey] = string(encodedAnnotations)
	params[osdPvcLabelsKey] = string(encodedLabels)
	md, err := getPVCMetadata(params)
	assert.NoError(t, err)

	assert.Equal(t, md[intreePvcNameKey], "mypvc1")
	assert.Equal(t, md[intreePvcNamespaceKey], "mypvcns1")
	assert.Equal(t, md["testkey_labels"], "testval_1")
	assert.Equal(t, md["testkey_annotations"], "testval_2")
}

func TestResolveSpecFromCSI(t *testing.T) {
	tt := []struct {
		name string

		req          *csi.CreateVolumeRequest
		existingSpec *api.VolumeSpec

		expectedSpec  *api.VolumeSpec
		expectedError string
	}{
		{
			name: "Should accept supported non-default FsType and Sharedv4",
			req: &csi.CreateVolumeRequest{
				VolumeCapabilities: []*csi.VolumeCapability{
					&csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{
								FsType: "xfs",
							},
						},
					},
					&csi.VolumeCapability{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
						},
					},
				},
			},
			existingSpec: &api.VolumeSpec{},

			expectedSpec: &api.VolumeSpec{
				Format:   api.FSType_FS_TYPE_XFS,
				Sharedv4: true,
			},
		},
		{
			name: "Should override with the CSI parameter if both are provided",
			req: &csi.CreateVolumeRequest{
				VolumeCapabilities: []*csi.VolumeCapability{
					&csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{
								FsType: "xfs",
							},
						},
					},
					&csi.VolumeCapability{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
						},
					},
				},
			},
			// for cases when the storage class parameter has fsType: EXT4 and this is already parsed out.
			existingSpec: &api.VolumeSpec{
				Format: api.FSType_FS_TYPE_EXT4,
			},

			// We should override with CSI provided parameter.
			expectedSpec: &api.VolumeSpec{
				Format:   api.FSType_FS_TYPE_XFS,
				Sharedv4: true,
			},
		},
		{
			name: "Should accept shared instead of sharedv4 if explicitly provided already",
			req: &csi.CreateVolumeRequest{
				VolumeCapabilities: []*csi.VolumeCapability{
					&csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{
								FsType: "ext4",
							},
						},
					},
					&csi.VolumeCapability{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
						},
					},
				},
			},
			existingSpec: &api.VolumeSpec{
				Shared: true,
			},

			expectedSpec: &api.VolumeSpec{
				Format: api.FSType_FS_TYPE_EXT4,
				Shared: true,
			},
		},
		{
			name: "Should not accept bad FsType",
			req: &csi.CreateVolumeRequest{
				VolumeCapabilities: []*csi.VolumeCapability{
					&csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{
								FsType: "badfs",
							},
						},
					},
					&csi.VolumeCapability{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
						},
					},
				},
			},
			existingSpec: &api.VolumeSpec{},

			expectedError: "no openstorage.FS_TYPE for badfs",
		},
		{
			name: "Should accept block volumes",
			req: &csi.CreateVolumeRequest{
				VolumeCapabilities: []*csi.VolumeCapability{
					&csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Block{
							Block: &csi.VolumeCapability_BlockVolume{},
						},
					},
					&csi.VolumeCapability{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
						},
					},
				},
			},
			existingSpec: &api.VolumeSpec{},
			expectedSpec: &api.VolumeSpec{
				Format:   api.FSType_FS_TYPE_NONE,
				Sharedv4: true,
			},
		},
		{
			name: "Should not set shared flag to true when using pure backends RWX",
			req: &csi.CreateVolumeRequest{
				VolumeCapabilities: []*csi.VolumeCapability{
					&csi.VolumeCapability{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
						},
					},
				},
			},
			existingSpec: &api.VolumeSpec{
				ProxySpec: &api.ProxySpec{
					ProxyProtocol: api.ProxyProtocol_PROXY_PROTOCOL_PURE_FILE,
				},
			},

			expectedSpec: &api.VolumeSpec{
				Shared: false,
				ProxySpec: &api.ProxySpec{
					ProxyProtocol: api.ProxyProtocol_PROXY_PROTOCOL_PURE_FILE,
				},
			},
		},
	}

	for _, tc := range tt {
		actualSpec, err := resolveSpecFromCSI(tc.existingSpec, tc.req)
		if tc.expectedError == "" {
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedSpec, actualSpec)
		} else {
			assert.EqualError(t, err, tc.expectedError)
		}
	}

}
