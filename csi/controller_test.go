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
	"fmt"
	"testing"

	"github.com/libopenstorage/openstorage/api"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestControllerGetCapabilities(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ControllerGetCapabilities(context.Background(), &csi.ControllerGetCapabilitiesRequest{})
	assert.Nil(t, err)

	// Verify
	caps := r.GetResult().GetCapabilities()
	assert.Len(t, caps, 1)
	assert.Equal(t,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		caps[0].GetRpc().GetType())
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

	// Missing everything
	c := csi.NewControllerClient(s.Conn())
	_, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Version")

	// Miss capabilities and id
	req.Version = &csi.Version{}
	_, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "capabilities")

	// Miss id and capabilities len is 0
	req.Version = &csi.Version{}
	req.VolumeCapabilities = []*csi.VolumeCapability{}
	_, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "capabilities")

	// Miss id
	req.Version = &csi.Version{}
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
			Inspect([]string{id}).
			Return(nil, fmt.Errorf("Id not found")),

		// Second time called it will not return an error,
		// but return an empty list
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{}, nil),

		// Third time it is called, it will return
		// a good list with a volume with an id that does
		// not match (even if this probably never could happen)
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: "bad volume id",
				},
			}, nil),

		// Fourth time driver will return a list with more than
		// one volume, which should be unexpected since it only
		// asked for one volume.
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: "bad volume id 1",
				},
				&api.Volume{
					Id: "bad volume id 2",
				},
			}, nil),
	)

	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		VolumeId: id,
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

	// Now driver should have returned an unexpected number of volumes
	_, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "unexpected number of volumes")
}

func TestControllerValidateVolumeInvalidCapabilities(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"
	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{
			&api.Volume{
				Id: id,
			},
		}, nil).
		Times(1)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		VolumeId: id,
	}

	// Make request
	c := csi.NewControllerClient(s.Conn())
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
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// RO and not-SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// not-RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),

		// RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),
	)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
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
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.True(t, r.GetResult().Supported)

	// Expect RO and non-SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.GetResult().Supported)

	// Expect non-RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.GetResult().Supported)

	// Expect RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.GetResult().Supported)
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
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// RO and not-SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// not-RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),

		// RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),
	)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
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
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.GetResult().Supported)

	// Expect RO and non-SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.True(t, r.GetResult().Supported)

	// Expect non-RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.GetResult().Supported)

	// Expect RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.GetResult().Supported)
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
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// RO and not-SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// not-RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),

		// RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),
	)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
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
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.GetResult().Supported)

	// Expect RO and non-SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.GetResult().Supported)

	// Expect non-RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.GetResult().Supported)

	// Expect RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.True(t, r.GetResult().Supported)
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
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// RO and not-SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// not-RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),

		// RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),
	)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
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
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.GetResult().Supported)

	// Expect RO and non-SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.GetResult().Supported)

	// Expect non-RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.True(t, r.GetResult().Supported)

	// Expect RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.GetResult().Supported)
}

func TestControllerValidateVolumeAccessModeUnknown(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"
	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{
			&api.Volume{
				Id:       id,
				Readonly: false,
				Spec: &api.VolumeSpec{
					Shared: false,
				},
			},
		}, nil).
		Times(1)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
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
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	_, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)
}
