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

	"github.com/golang/protobuf/ptypes"

	"github.com/libopenstorage/openstorage/api"
	authsecrets "github.com/libopenstorage/openstorage/pkg/auth/secrets"
	"github.com/portworx/kvdb"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		Inspect([]string{id}).
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
				Name: "size",
				VolumeCapabilities: []*csi.VolumeCapability{
					&csi.VolumeCapability{},
				},
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
		fmt.Println("err", err)
		assert.Equal(t, codes.AlreadyExists, serverError.Code())
	}
}

func TestControllerCreateVolumeNoCapacity(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

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
			Create(gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(
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
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size:   defaultCSIVolumeSize,
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
	assert.Equal(t, int64(defaultCSIVolumeSize), volumeInfo.GetCapacityBytes())
}

func TestControllerCreateVolumeFoundByVolumeFromName(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

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
					Spec: &api.VolumeSpec{
						Size: uint64(size),
					},
				},
			}, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return([]*api.Volume{&api.Volume{
				Id: name,
				Locator: &api.VolumeLocator{
					Name: name,
				},
				Spec: &api.VolumeSpec{
					Size: uint64(1234),
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
			Inspect([]string{parent}).
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
			Inspect([]string{parent}).
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

func TestControllerCreateVolumeWithSharedVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

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
				Return(nil, fmt.Errorf("not found")).
				Times(1),

			s.MockDriver().
				EXPECT().
				Enumerate(&api.VolumeLocator{Name: name}, nil).
				Return(nil, fmt.Errorf("not found")).
				Times(1),

			s.MockDriver().
				EXPECT().
				Create(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(id, nil).
				Times(1),

			s.MockDriver().
				EXPECT().
				Inspect([]string{id}).
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
			Create(gomock.Any(), gomock.Any(), gomock.Any()).
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
			Create(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
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

func TestControllerCreateVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

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
			Create(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
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
	assert.NotEqual(t, "true", volumeInfo.GetVolumeContext()[api.SpecShared])
}

func TestControllerCreateVolumeFromSnapshot(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

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
	gomock.InOrder(
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

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
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
	assert.NotEqual(t, "true", volumeInfo.GetVolumeContext()[api.SpecShared])
	assert.Equal(t, mockParentID, volumeInfo.GetVolumeContext()[api.SpecParent])
}

func TestControllerCreateVolumeSnapshotThroughParameters(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

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
			Inspect([]string{mockParentID}).
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
			Inspect([]string{id}).
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
			Inspect([]string{id}).
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
			Inspect([]string{id}).
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
	assert.NotEqual(t, "true", volumeInfo.GetVolumeContext()[api.SpecShared])
	assert.Equal(t, mockParentID, volumeInfo.GetVolumeContext()[api.SpecParent])
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
			Inspect([]string{myid}).
			Return([]*api.Volume{
				&api.Volume{
					Id: myid,
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Delete(myid).
			Return(fmt.Errorf("MOCKERRORTEST")).
			Times(1),
	)

	_, err := c.DeleteVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
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

	// Setup mock
	// According to CSI spec, if the ID is not found, it must return OK
	s.MockDriver().
		EXPECT().
		Inspect([]string{myid}).
		Return(nil, kvdb.ErrNotFound).
		Times(1)

	_, err := c.DeleteVolume(context.Background(), req)
	assert.Nil(t, err)

	// According to CSI spec, if the ID is not found, it must return OK
	// Now return no error, but empty list
	s.MockDriver().
		EXPECT().
		Inspect([]string{myid}).
		Return([]*api.Volume{}, nil).
		Times(1)

	_, err = c.DeleteVolume(context.Background(), req)
	assert.Nil(t, err)

	// Setup mock
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{myid}).
			Return([]*api.Volume{
				&api.Volume{
					Id: myid,
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Delete(myid).
			Return(nil).
			Times(1),
	)

	_, err = c.DeleteVolume(context.Background(), req)
	assert.Nil(t, err)
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
	snapInfo := &api.Volume{
		Id: name,
		Source: &api.Source{
			Parent: volume,
		},
		Ctime: ptypes.TimestampNow(),
		Locator: &api.VolumeLocator{
			Name: name,
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
			Inspect([]string{name}).
			Return([]*api.Volume{snapInfo}, nil).
			Times(1),
	)

	r, err := c.CreateSnapshot(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, name, r.GetSnapshot().GetSnapshotId())
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
	snapInfo := &api.Volume{
		Id: name,
		Source: &api.Source{
			Parent: volume,
		},
		Ctime: ptypes.TimestampNow(),
		Locator: &api.VolumeLocator{
			Name: name,
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
			Inspect([]string{volume}).
			Return([]*api.Volume{&api.Volume{Id: volume}}, nil).
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
			Inspect([]string{snapId}).
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
		Inspect([]string{id}).
		Return(nil, kvdb.ErrNotFound).
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
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{},
			}, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Delete(id).
			Return(nil).
			Times(1),
	)

	_, err := c.DeleteSnapshot(context.Background(), &csi.DeleteSnapshotRequest{
		SnapshotId: id,
		Secrets:    map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	})
	assert.NoError(t, err)
}
