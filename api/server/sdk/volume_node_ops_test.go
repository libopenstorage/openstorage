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

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
	mountattachoptions "github.com/libopenstorage/openstorage/pkg/options"
)

func TestSdkVolumeAttachSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myid"
	devpath := "/my/path"
	options := map[string]string{
		mountattachoptions.OptionsSecret:        "name",
		mountattachoptions.OptionsSecretContext: "context",
		mountattachoptions.OptionsSecretKey:     "key",
	}

	req := &api.SdkVolumeAttachRequest{
		VolumeId: id,
		Options: &api.SdkVolumeAttachRequest_Options{
			SecretName:    "name",
			SecretContext: "context",
			SecretKey:     "key",
		},
	}

	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:    id,
					State: api.VolumeState_VOLUME_STATE_DETACHED,
				},
			}, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Attach(id, options).
			Return(devpath, nil),
	)

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

	// Attach Volume
	res, err := c.Attach(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, res.GetDevicePath(), devpath)
}

func TestSdkVolumeAttachSuccessIdempotent(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myid"
	devpath := "/my/path"
	req := &api.SdkVolumeAttachRequest{
		VolumeId: id,
		Options: &api.SdkVolumeAttachRequest_Options{
			SecretName:    "name",
			SecretContext: "context",
			SecretKey:     "key",
		},
	}

	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{
			&api.Volume{
				Id:            id,
				State:         api.VolumeState_VOLUME_STATE_ATTACHED,
				AttachedState: api.AttachState_ATTACH_STATE_EXTERNAL,
				DevicePath:    devpath,
			},
		}, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

	// Attach Volume
	res, err := c.Attach(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, res.GetDevicePath(), devpath)
}

func TestSdkVolumeAttachFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "mytestid"
	options := map[string]string{
		mountattachoptions.OptionsSecret:        "name",
		mountattachoptions.OptionsSecretContext: "context",
		mountattachoptions.OptionsSecretKey:     "key",
	}

	req := &api.SdkVolumeAttachRequest{
		VolumeId: id,
		Options: &api.SdkVolumeAttachRequest_Options{
			SecretName:    "name",
			SecretContext: "context",
			SecretKey:     "key",
		},
	}
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:    id,
					State: api.VolumeState_VOLUME_STATE_DETACHED,
				},
			}, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Attach(id, options).
			Return("", fmt.Errorf("Failed to Attach device")),
	)

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

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

	req := &api.SdkVolumeAttachRequest{
		VolumeId: "",
	}

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

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
	options := map[string]string{
		mountattachoptions.OptionsRedirectDetach:      "true",
		mountattachoptions.OptionsForceDetach:         "false",
		mountattachoptions.OptionsUnmountBeforeDetach: "true",
	}
	req := &api.SdkVolumeDetachRequest{
		VolumeId: id,
		Options: &api.SdkVolumeDetachRequest_Options{
			Force:               false,
			UnmountBeforeDetach: true,
		},
	}
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:    id,
					State: api.VolumeState_VOLUME_STATE_ATTACHED,
				},
			}, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Detach(id, options).
			Return(nil),
	)

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

	// Get info
	_, err := c.Detach(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkVolumeDetachSuccessIdempotency(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "dummy-volume-id"
	req := &api.SdkVolumeDetachRequest{
		VolumeId: id,
		Options: &api.SdkVolumeDetachRequest_Options{
			Force:               false,
			UnmountBeforeDetach: true,
		},
	}

	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{
			&api.Volume{
				Id:    id,
				State: api.VolumeState_VOLUME_STATE_DETACHED,
			},
		}, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

	// Get info
	_, err := c.Detach(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkVolumeDetachFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "dummy-volume-id"
	options := map[string]string{
		mountattachoptions.OptionsRedirectDetach:      "true",
		mountattachoptions.OptionsForceDetach:         "true",
		mountattachoptions.OptionsUnmountBeforeDetach: "false",
	}
	req := &api.SdkVolumeDetachRequest{
		VolumeId: id,
		Options: &api.SdkVolumeDetachRequest_Options{
			Force:               true,
			UnmountBeforeDetach: false,
		},
	}
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:    id,
					State: api.VolumeState_VOLUME_STATE_ATTACHED,
				},
			}, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Detach(id, options).
			Return(fmt.Errorf("Failed to Detach")),
	)

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

	// Get info
	_, err := c.Detach(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Failed to Detach")
}

func TestSdkVolumeDetachBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkVolumeDetachRequest{}

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

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
	mountPath := "/dev/real/path"

	req := &api.SdkVolumeMountRequest{
		VolumeId:  id,
		MountPath: mountPath,
	}
	s.MockDriver().
		EXPECT().
		Mount(id, mountPath, nil).
		Return(nil)

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

	// Get info
	_, err := c.Mount(context.Background(), req)
	assert.NoError(t, err)
}
func TestSdkVolumeMountFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "dummy-volume-id"
	mountPath := "/dev/fake/path"

	req := &api.SdkVolumeMountRequest{
		VolumeId:  id,
		MountPath: mountPath,
	}
	s.MockDriver().
		EXPECT().
		Mount(id, mountPath, nil).
		Return(fmt.Errorf("Invalid Mount Path"))

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

	// Get info
	_, err := c.Mount(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Invalid Mount Path")
}

func TestSdkVolumeMountBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "dummy-volume-id"
	mountPath := ""

	req := &api.SdkVolumeMountRequest{
		VolumeId:  id,
		MountPath: mountPath,
	}

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

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
		mountattachoptions.OptionsDeleteAfterUnmount: "true",
		mountattachoptions.OptionsWaitBeforeDelete:   "true",
	}
	mountPath := "/mnt/testmount"
	req := &api.SdkVolumeUnmountRequest{
		VolumeId:  id,
		MountPath: mountPath,
		Options: &api.SdkVolumeUnmountRequest_Options{
			DeleteMountPath:                true,
			NoDelayBeforeDeletingMountPath: false,
		},
	}

	s.MockDriver().
		EXPECT().
		Unmount(id, mountPath, options).
		Return(nil)

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

	// Unmount Volume
	_, err := c.Unmount(context.Background(), req)
	assert.NoError(t, err)

	options = map[string]string{
		mountattachoptions.OptionsDeleteAfterUnmount: "true",
		mountattachoptions.OptionsWaitBeforeDelete:   "true",
	}
	req = &api.SdkVolumeUnmountRequest{
		VolumeId:  id,
		MountPath: mountPath,
		Options: &api.SdkVolumeUnmountRequest_Options{
			DeleteMountPath: true,
		},
	}

	s.MockDriver().
		EXPECT().
		Unmount(id, mountPath, options).
		Return(nil)

	_, err = c.Unmount(context.Background(), req)
	assert.NoError(t, err)

	options = map[string]string{
		mountattachoptions.OptionsDeleteAfterUnmount: "true",
		mountattachoptions.OptionsWaitBeforeDelete:   "false",
	}
	req = &api.SdkVolumeUnmountRequest{
		VolumeId:  id,
		MountPath: mountPath,
		Options: &api.SdkVolumeUnmountRequest_Options{
			DeleteMountPath:                true,
			NoDelayBeforeDeletingMountPath: true,
		},
	}

	s.MockDriver().
		EXPECT().
		Unmount(id, mountPath, options).
		Return(nil)

	_, err = c.Unmount(context.Background(), req)
	assert.NoError(t, err)

	options = map[string]string{
		mountattachoptions.OptionsDeleteAfterUnmount: "false",
	}
	req = &api.SdkVolumeUnmountRequest{
		VolumeId:  id,
		MountPath: mountPath,
		Options: &api.SdkVolumeUnmountRequest_Options{
			DeleteMountPath: false,
		},
	}

	s.MockDriver().
		EXPECT().
		Unmount(id, mountPath, options).
		Return(nil)

	_, err = c.Unmount(context.Background(), req)
	assert.NoError(t, err)

	// Check when no values set
	options = map[string]string{
		mountattachoptions.OptionsDeleteAfterUnmount: "false",
	}
	req = &api.SdkVolumeUnmountRequest{
		VolumeId:  id,
		MountPath: mountPath,
		Options:   &api.SdkVolumeUnmountRequest_Options{},
	}

	s.MockDriver().
		EXPECT().
		Unmount(id, mountPath, options).
		Return(nil)

	_, err = c.Unmount(context.Background(), req)
	assert.NoError(t, err)

	// Check when no options are given
	options = map[string]string{}
	req = &api.SdkVolumeUnmountRequest{
		VolumeId:  id,
		MountPath: mountPath,
	}

	s.MockDriver().
		EXPECT().
		Unmount(id, mountPath, options).
		Return(nil)

	_, err = c.Unmount(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkVolumeUnmountFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "testid"
	options := map[string]string{}
	mountPath := "/dev/fake/path"

	req := &api.SdkVolumeUnmountRequest{
		VolumeId:  id,
		MountPath: mountPath,
	}
	s.MockDriver().
		EXPECT().
		Unmount(id, mountPath, options).
		Return(fmt.Errorf("Invalid Mount Path"))

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

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
	mountPath := "/mnt/mounttest"

	req := &api.SdkVolumeUnmountRequest{
		VolumeId:  id,
		MountPath: mountPath,
	}

	// Setup client
	c := api.NewOpenStorageMountAttachClient(s.Conn())

	// Get info
	_, err := c.Unmount(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply volume id")
}
