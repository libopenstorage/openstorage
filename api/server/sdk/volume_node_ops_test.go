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
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
)

func TestSdkVolumeAttachSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myid"
	options := map[string]string{
		"SECRET_NAME": "value1",
	}

	req := &api.SdkVolumeAttachRequest{
		VolumeId: id,
		Options:  options,
	}
	s.MockDriver().
		EXPECT().
		Attach(id, options).
		Return("", nil)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Attach Volume
	_, err := c.Attach(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkVolumeAttachFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "mytestid"
	options := map[string]string{
		"passphrase": "testval",
	}

	req := &api.SdkVolumeAttachRequest{
		VolumeId: id,
		Options:  options,
	}

	s.MockDriver().
		EXPECT().
		Attach(id, options).
		Return("", fmt.Errorf("Failed to Attach device"))

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Attach(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Failed to Attach device")
}

func TestSdkVolumeAttachBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := ""
	options := map[string]string{
		"passphrase": "testval",
	}

	req := &api.SdkVolumeAttachRequest{
		VolumeId: id,
		Options:  options,
	}

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Attach(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply volume id")
}

func TestSdkVolumeDetachSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "dummy-volume-id"
	req := &api.SdkVolumeDetachRequest{
		VolumeId: id,
	}
	s.MockDriver().
		EXPECT().
		Detach(id, nil).
		Return(nil)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Detach(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkVolumeDetachFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "dummy-volume-id"
	req := &api.SdkVolumeDetachRequest{
		VolumeId: id,
	}
	s.MockDriver().
		EXPECT().
		Detach(id, nil).
		Return(fmt.Errorf("Failed to Detach"))

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Detach(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Unknown)
	assert.Contains(t, serverError.Message(), "Failed to Detach")
}

func TestSdkVolumeDetachBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := ""
	req := &api.SdkVolumeDetachRequest{
		VolumeId: id,
	}

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Detach(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply volume id")
}

func TestSdkVolumeMountSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "dummy-volume-id"
	options := map[string]string{
		"option1": "value1",
		"option2": "value2",
	}
	mountPath := "/dev/real/path"

	req := &api.SdkVolumeMountRequest{
		VolumeId:  id,
		MountPath: mountPath,
		Options:   options,
	}
	s.MockDriver().
		EXPECT().
		Mount(id, mountPath, options).
		Return(nil)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Mount(context.Background(), req)
	assert.NoError(t, err)
}
func TestSdkVolumeMountFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "dummy-volume-id"
	options := map[string]string{
		"option1": "value1",
		"option2": "value2",
	}
	mountPath := "/dev/fake/path"

	req := &api.SdkVolumeMountRequest{
		VolumeId:  id,
		MountPath: mountPath,
		Options:   options,
	}
	s.MockDriver().
		EXPECT().
		Mount(id, mountPath, options).
		Return(fmt.Errorf("Invalid Mount Path"))

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Mount(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Unknown)
	assert.Contains(t, serverError.Message(), "Invalid Mount Path")
}

func TestSdkVolumeMountBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "dummy-volume-id"
	options := map[string]string{
		"option1": "value1",
		"option2": "value2",
	}
	mountPath := ""

	req := &api.SdkVolumeMountRequest{
		VolumeId:  id,
		MountPath: mountPath,
		Options:   options,
	}

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Mount(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Invalid Mount Path")
}

func TestSdkVolumeUnmountSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myid"

	options := map[string]string{
		"option1": "value1",
		"option2": "value2",
	}
	mountPath := "/mnt/testmount"
	req := &api.SdkVolumeUnmountRequest{
		VolumeId:  id,
		MountPath: mountPath,
		Options:   options,
	}

	s.MockDriver().
		EXPECT().
		Unmount(id, mountPath, options).
		Return(nil)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Unmount Volume
	_, err := c.Unmount(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkVolumeUnmountFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "testid"
	options := map[string]string{
		"option1": "value1",
		"option2": "value2",
	}
	mountPath := "/dev/fake/path"

	req := &api.SdkVolumeUnmountRequest{
		VolumeId:  id,
		MountPath: mountPath,
		Options:   options,
	}
	s.MockDriver().
		EXPECT().
		Unmount(id, mountPath, options).
		Return(fmt.Errorf("Invalid Mount Path"))

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Unmount(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Invalid Mount Path")
}

func TestSdkVolumeUnountBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := ""
	options := map[string]string{
		"option1": "value1",
		"option2": "value2",
	}
	mountPath := "/mnt/mounttest"

	req := &api.SdkVolumeUnmountRequest{
		VolumeId:  id,
		MountPath: mountPath,
		Options:   options,
	}

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Unmount(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply volume id")
}
