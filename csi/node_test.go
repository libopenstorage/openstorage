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
	"testing"
	"time"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/mock/gomock"
	"github.com/libopenstorage/openstorage/api"
	authsecrets "github.com/libopenstorage/openstorage/pkg/auth/secrets"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNodePublishVolumeBadArguments(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	testargs := []struct {
		expectedErrorContains string
		req                   *csi.NodePublishVolumeRequest
	}{
		{
			expectedErrorContains: "Volume id",
			req: &csi.NodePublishVolumeRequest{},
		},
		{
			expectedErrorContains: "Target path",
			req: &csi.NodePublishVolumeRequest{
				VolumeId: "abc",
			},
		},
		{
			expectedErrorContains: "Volume access mode",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:   "abc",
				TargetPath: "mypath",
			},
		},
		{
			expectedErrorContains: "Volume access mode",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:         "abc",
				TargetPath:       "mypath",
				VolumeCapability: &csi.VolumeCapability{},
			},
		},
	}

	for _, testarg := range testargs {
		_, err := c.NodePublishVolume(context.Background(), testarg.req)
		assert.NotNil(t, err)
		serverError, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, serverError.Code(), codes.InvalidArgument)
		assert.Contains(t, serverError.Message(), testarg.expectedErrorContains)
	}
}

func TestNodePublishVolumeVolumeNotFound(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "myvol"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_NONE).
			Times(1),

		// Getting volume information
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: "/",
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{},
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	_, err := c.NodePublishVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, serverError.Code())
	assert.Contains(t, serverError.Message(), "not found")
}

func TestNodePublishVolumeBadAttribute(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "myvol"
	s.MockDriver().
		EXPECT().
		Type().
		Return(api.DriverType_DRIVER_TYPE_BLOCK).
		Times(1)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: "mypath",
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{},
		},

		// This will cause an error
		VolumeContext: map[string]string{
			api.SpecFilesystem: "whatkindoffsisthis?",
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	_, err := c.NodePublishVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Invalid volume attributes")
}

func TestNodePublishVolumeInvalidTargetLocation(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testargs := []struct {
		expectedErrorContains string
		targetPath            string
	}{
		{
			expectedErrorContains: "does not exist",
			targetPath:            "////a/sdf//fd/asdf/as/f/asdfasf/fds",
		},
		{
			expectedErrorContains: "not a directory",
			targetPath:            "/etc/hosts",
		},
	}

	c := csi.NewNodeClient(s.Conn())
	name := "myvol"
	s.MockDriver().
		EXPECT().
		Type().
		Return(api.DriverType_DRIVER_TYPE_NONE).
		Times(len(testargs))
	req := &csi.NodePublishVolumeRequest{
		VolumeId: name,
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{},
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	for _, testarg := range testargs {
		req.TargetPath = testarg.targetPath
		_, err := c.NodePublishVolume(context.Background(), req)
		assert.NotNil(t, err)
		serverError, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, serverError.Code(), codes.Aborted)
		assert.Contains(t, serverError.Message(), testarg.expectedErrorContains)
	}
}

func TestNodePublishVolumeFailedToAttach(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "myvol"
	size := uint64(10)
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_BLOCK).
			Times(1),

		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Attach(name, gomock.Any()).
			Return("", fmt.Errorf("Unable to attach volume")).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: "/mnt",
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{},
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	_, err := c.NodePublishVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	fmt.Println("err", err, serverError)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Unable to attach volume")
}

func TestNodePublishVolumeFailedMount(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "myvol"
	size := uint64(10)
	targetPath := "/mnt"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_NONE).
			Times(1),
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_NONE).
			Times(1),
		s.MockDriver().
			EXPECT().
			Mount(name, targetPath, nil).
			Return(fmt.Errorf("Unable to mount volume")).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: targetPath,
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{},
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	_, err := c.NodePublishVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Unable to mount volume")
}

func TestNodePublishVolumeBlock(t *testing.T) {
	// Skipping for now - issues with symlink
	t.Skip()

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "myvol"
	size := uint64(10)

	// create devicepath/targetpath
	devicePath := fmt.Sprintf("/tmp/csi-devicePath.%d", time.Now().Unix())
	targetPath := fmt.Sprintf("/tmp/csi-targetPath.%d", time.Now().Unix())

	// Create the devicePath
	f, err := os.Create(devicePath)
	assert.NoError(t, err)
	f.Close()

	// cleanup devicepath and targetpath
	defer os.Remove(devicePath)
	defer os.Remove(targetPath)

	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_BLOCK).
			Times(1),
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_BLOCK).
			Times(1),
		s.MockDriver().
			EXPECT().
			Attach(name, gomock.Any()).
			Return(devicePath, nil).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: targetPath,
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{},
			AccessType: &csi.VolumeCapability_Block{
				Block: &csi.VolumeCapability_BlockVolume{},
			},
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	r, err := c.NodePublishVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)

	// Check that the symlink was created
	fileInfo, err := os.Lstat(targetPath)
	assert.NoError(t, err)
	assert.Equal(t, fileInfo.Mode()&os.ModeSymlink, os.ModeSymlink)
}

func TestNodePublishVolumeMount(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "myvol"
	size := uint64(10)
	targetPath := "/mnt"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_NONE).
			Times(1),
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_NONE).
			Times(1),

		s.MockDriver().
			EXPECT().
			Mount(name, targetPath, nil).
			Return(nil).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: targetPath,
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{},
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	r, err := c.NodePublishVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
}

func TestNodeUnpublishVolumeVolumeNotFound(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "myvol"
	gomock.InOrder(
		// Getting volume information
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
	)

	req := &csi.NodeUnpublishVolumeRequest{
		VolumeId:   name,
		TargetPath: "mypath",
	}

	_, err := c.NodeUnpublishVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "not found")
}

func TestNodeUnpublishVolumeInvalidTargetLocation(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testargs := []struct {
		expectedErrorContains string
		targetPath            string
	}{
		{
			expectedErrorContains: "not a directory",
			targetPath:            "/etc/hosts",
		},
	}

	c := csi.NewNodeClient(s.Conn())
	name := "myvol"
	s.MockDriver().
		EXPECT().
		Inspect([]string{name}).
		Return([]*api.Volume{
			&api.Volume{
				Id: name,
			},
		}, nil).
		Times(len(testargs))

	req := &csi.NodeUnpublishVolumeRequest{
		VolumeId: name,
	}

	for _, testarg := range testargs {
		req.TargetPath = testarg.targetPath
		_, err := c.NodeUnpublishVolume(context.Background(), req)
		assert.NotNil(t, err)
		serverError, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, serverError.Code(), codes.NotFound)
		assert.Contains(t, serverError.Message(), testarg.expectedErrorContains)
	}
}

func TestNodeUnpublishVolumeFailedToUnmount(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "myvol"
	targetPath := "/mnt"
	size := uint64(10)
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Unmount(name, targetPath, nil).
			Return(fmt.Errorf("TEST")).
			Times(1),
	)

	req := &csi.NodeUnpublishVolumeRequest{
		VolumeId:   name,
		TargetPath: "/mnt",
	}

	_, err := c.NodeUnpublishVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Unable to unmount volume")
	assert.Contains(t, serverError.Message(), "TEST")
}

func TestNodeUnpublishVolumeFailedDetach(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "myvol"
	size := uint64(10)
	targetPath := "/mnt"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Unmount(name, targetPath, nil).
			Return(nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_BLOCK).
			Times(1),
		s.MockDriver().
			EXPECT().
			Detach(name, gomock.Any()).
			Return(fmt.Errorf("DETACH ERROR")).
			Times(1),
	)

	req := &csi.NodeUnpublishVolumeRequest{
		VolumeId:   name,
		TargetPath: targetPath,
	}

	_, err := c.NodeUnpublishVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Unable to detach volume")
	assert.Contains(t, serverError.Message(), "DETACH ERROR")
}

func TestNodeUnpublishVolumeUnmount(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "myvol"
	size := uint64(10)
	targetPath := "/mnt"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Unmount(name, targetPath, nil).
			Return(nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_BLOCK).
			Times(1),
		s.MockDriver().
			EXPECT().
			Detach(name, gomock.Any()).
			Return(nil).
			Times(1),
	)

	req := &csi.NodeUnpublishVolumeRequest{
		VolumeId:   name,
		TargetPath: targetPath,
	}

	r, err := c.NodeUnpublishVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
}

func TestNodeGetCapabilities(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	// Get Capabilities
	r, err := c.NodeGetCapabilities(
		context.Background(),
		&csi.NodeGetCapabilitiesRequest{})
	assert.NoError(t, err)
	assert.Len(t, r.GetCapabilities(), 1)
	assert.Equal(
		t,
		csi.NodeServiceCapability_RPC_UNKNOWN,
		r.GetCapabilities()[0].GetRpc().GetType())
}
