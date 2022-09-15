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
	"strings"
	"testing"
	"time"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/mock/gomock"
	"github.com/libopenstorage/openstorage/api"
	authsecrets "github.com/libopenstorage/openstorage/pkg/auth/secrets"
	"github.com/portworx/kvdb"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetNodeInfo(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	// TestCase; Error in cluster enumerate
	cluster := api.Cluster{NodeId: "node-1"}
	s.MockCluster().EXPECT().
		Enumerate().
		Return(cluster, fmt.Errorf("enumerate error")).
		Times(1)

	_, err := c.NodeGetInfo(context.Background(), &csi.NodeGetInfoRequest{})
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "enumerate error")

	// TestCase: Error in cluster inspect
	node := api.Node{}
	s.MockCluster().EXPECT().
		Enumerate().
		Return(cluster, nil).
		AnyTimes()
	s.MockCluster().EXPECT().
		Inspect("node-1").
		Return(node, fmt.Errorf("inspect error")).
		Times(1)

	_, err = c.NodeGetInfo(context.Background(), &csi.NodeGetInfoRequest{})
	assert.NotNil(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "inspect error")

	// TestCase: Successful node info
	s.MockCluster().EXPECT().
		Inspect("node-1").
		Return(node, nil).
		Times(1)

	res, err := c.NodeGetInfo(context.Background(), &csi.NodeGetInfoRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "node-1", res.NodeId)
	assert.Empty(t, res.AccessibleTopology.Segments)

	// TestCase: Node info with empty topology
	node.SchedulerTopology = &api.SchedulerTopology{}
	s.MockCluster().EXPECT().
		Inspect("node-1").
		Return(node, nil).
		Times(1)

	res, err = c.NodeGetInfo(context.Background(), &csi.NodeGetInfoRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "node-1", res.NodeId)
	assert.Empty(t, res.AccessibleTopology.Segments)

	// TestCase: Node info with empty topology labels
	node.SchedulerTopology.Labels = map[string]string{}
	s.MockCluster().EXPECT().
		Inspect("node-1").
		Return(node, nil).
		Times(1)

	res, err = c.NodeGetInfo(context.Background(), &csi.NodeGetInfoRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "node-1", res.NodeId)
	assert.Empty(t, res.AccessibleTopology.Segments)

	// TestCase: Node info with empty topology labels
	node.SchedulerTopology.Labels["zone"] = "zone-1"
	node.SchedulerTopology.Labels["region"] = "region-1"
	s.MockCluster().EXPECT().
		Inspect("node-1").
		Return(node, nil).
		AnyTimes()

	res, err = c.NodeGetInfo(context.Background(), &csi.NodeGetInfoRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "node-1", res.NodeId)
	assert.NotEmpty(t, res.AccessibleTopology)
	assert.Len(t, res.AccessibleTopology.Segments, 2)
	assert.Equal(t, res.AccessibleTopology.Segments["zone"], "zone-1")
	assert.Equal(t, res.AccessibleTopology.Segments["region"], "region-1")
}

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
			req:                   &csi.NodePublishVolumeRequest{},
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
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{name},
			}, nil).
			Return([]*api.Volume{}, nil).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: "/",
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
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

	os.Mkdir("mypath", 0750)
	defer os.Remove("mypath")
	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: "mypath",
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
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
		AnyTimes()
	req := &csi.NodePublishVolumeRequest{
		VolumeId: name,
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
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
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{name},
			}, nil).
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
			Attach(gomock.Any(), name, gomock.Any()).
			Return("", fmt.Errorf("Unable to attach volume")).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: "/mnt",
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	_, err := c.NodePublishVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unavailable, serverError.Code())
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
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{name},
			}, nil).
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
			Mount(gomock.Any(), name, targetPath, nil).
			Return(fmt.Errorf("Unable to mount volume")).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: targetPath,
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	_, err := c.NodePublishVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unavailable, serverError.Code())
	assert.Contains(t, serverError.Message(), "Unable to mount volume")
}

func TestNodePublishVolumeBlock(t *testing.T) {
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
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{name},
			}, nil).
			Return([]*api.Volume{
				{
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
			Attach(gomock.Any(), name, gomock.Any()).
			Return("", nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{name},
			}, nil).
			Return([]*api.Volume{
				{
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
			Mount(gomock.Any(), name, targetPath, nil).
			Return(nil).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: targetPath,
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: &csi.VolumeCapability_Block{
				Block: &csi.VolumeCapability_BlockVolume{},
			},
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	r, err := c.NodePublishVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)

	// Check that the targetpath exists
	_, err = os.Lstat(targetPath)
	assert.NoError(t, err)
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
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{name},
			}, nil).
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
			Mount(gomock.Any(), name, targetPath, nil).
			Return(nil).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: targetPath,
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	r, err := c.NodePublishVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
}

func TestNodePublishPureVolumeMountOptions(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "myvol"
	size := uint64(10)
	targetPath := "/mnt"
	mountFlags := []string{"nfsvers=4.1", "tcp"}
	options := strings.Join(mountFlags, ",")
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_NONE).
			Times(1),
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{name},
			}, nil).
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
			Mount(gomock.Any(), name, targetPath, map[string]string{api.SpecCSIMountOptions: options}).
			Return(nil).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: targetPath,
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			}, AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{
					MountFlags: mountFlags,
				},
			},
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	r, err := c.NodePublishVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
}

func TestNodePublishVolumeEphemeralEnabled(t *testing.T) {
	// Create server and client connection
	s := newTestServerWithConfig(t, &OsdCsiServerConfig{
		DriverName:          mockDriverName,
		EnableInlineVolumes: true,
	})
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "csi-12345"
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
			Return(name, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{
				VolumeIds: []string{name},
			}, nil).
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
			Mount(gomock.Any(), name, targetPath, nil).
			Return(nil).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: targetPath,
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
		},
		VolumeContext: map[string]string{
			"csi.storage.k8s.io/ephemeral": "true",
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	r, err := c.NodePublishVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
}

func TestNodePublishVolumeEphemeralDisabled(t *testing.T) {
	// Create server and client connection
	s := newTestServerWithConfig(t, &OsdCsiServerConfig{
		DriverName:          mockDriverName,
		EnableInlineVolumes: false,
	})
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "csi-12345"
	targetPath := "/mnt"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_NONE).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: targetPath,
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
		},
		VolumeContext: map[string]string{
			"csi.storage.k8s.io/ephemeral": "true",
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	_, err := c.NodePublishVolume(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CSI ephemeral inline volumes are disabled on this cluster")
}

func TestNodePublishVolumeEphemeralDenyList(t *testing.T) {
	// Create server and client connection
	s := newTestServerWithConfig(t, &OsdCsiServerConfig{
		DriverName:          mockDriverName,
		EnableInlineVolumes: true,
	})
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "csi-12345"
	targetPath := "/mnt"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_NONE).
			Times(1),
	)

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   name,
		TargetPath: targetPath,
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			}},
		VolumeContext: map[string]string{
			"csi.storage.k8s.io/ephemeral": "true",
			api.SpecPriority:               "high",
		},
		Secrets: map[string]string{authsecrets.SecretTokenKey: systemUserToken},
	}

	_, err := c.NodePublishVolume(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ephemeral volume attribute provided")
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
			Enumerate(&api.VolumeLocator{VolumeIds: []string{name}}, nil).
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

func TestNodeUnpublishVolumeFailedToUnmount(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	name := "myvol"
	targetPath := "/tmp/mnttest2"
	size := uint64(10)
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{VolumeIds: []string{name}}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
					AttachPath:    []string{targetPath},
					AttachedOn:    "node1",
					State:         api.VolumeState_VOLUME_STATE_ATTACHED,
					AttachedState: api.AttachState_ATTACH_STATE_EXTERNAL,
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Unmount(gomock.Any(), name, targetPath, nil).
			Return(fmt.Errorf("TEST")).
			Times(1),
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_BLOCK).
			Times(1),
		s.MockDriver().
			EXPECT().
			Detach(gomock.Any(), name, gomock.Any()).
			Return(nil).
			Times(1),
	)

	req := &csi.NodeUnpublishVolumeRequest{
		VolumeId:   name,
		TargetPath: targetPath,
	}

	_, err := c.NodeUnpublishVolume(context.Background(), req)
	assert.Nil(t, err)
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
			Enumerate(&api.VolumeLocator{VolumeIds: []string{name}}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
					AttachPath:    []string{targetPath},
					AttachedOn:    "node1",
					State:         api.VolumeState_VOLUME_STATE_ATTACHED,
					AttachedState: api.AttachState_ATTACH_STATE_EXTERNAL,
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Unmount(gomock.Any(), name, targetPath, nil).
			Return(nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_BLOCK).
			Times(1),
		s.MockDriver().
			EXPECT().
			Detach(gomock.Any(), name, gomock.Any()).
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
	assert.Equal(t, serverError.Code(), codes.Canceled)
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
	targetPath := "/tmp/mnttest"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{VolumeIds: []string{name}}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
					AttachPath:    []string{targetPath},
					AttachedOn:    "node1",
					State:         api.VolumeState_VOLUME_STATE_ATTACHED,
					AttachedState: api.AttachState_ATTACH_STATE_EXTERNAL,
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Unmount(gomock.Any(), name, targetPath, nil).
			Return(nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Type().
			Return(api.DriverType_DRIVER_TYPE_BLOCK).
			Times(1),
		s.MockDriver().
			EXPECT().
			Detach(gomock.Any(), name, gomock.Any()).
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
	assert.Len(t, r.GetCapabilities(), 2)
	assert.Equal(
		t,
		csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
		r.GetCapabilities()[0].GetRpc().GetType())
	assert.Equal(
		t,
		csi.NodeServiceCapability_RPC_VOLUME_CONDITION,
		r.GetCapabilities()[1].GetRpc().GetType())
}

func TestNodeGetVolumeStats(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	size := int64(4 * 1024 * 1024)
	used := int64(1 * 1024 * 1024)
	available := size - used
	id := "myvol123"
	vol := &api.Volume{
		AttachPath: []string{"/test"},
		Id:         id,
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
	c := csi.NewNodeClient(s.Conn())

	// Get VolumeStats - all OK
	resp, err := c.NodeGetVolumeStats(
		context.Background(),
		&csi.NodeGetVolumeStatsRequest{VolumeId: id, VolumePath: "/test"})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(resp.Usage))
	assert.Equal(t, size, resp.Usage[0].Total)
	assert.Equal(t, used, resp.Usage[0].Used)
	assert.Equal(t, available, resp.Usage[0].Available)
	assert.Equal(t, false, resp.VolumeCondition.Abnormal)
	assert.Equal(t, "Volume status is up", resp.VolumeCondition.Message)

	// Get VolumeStats - down
	vol.Status = api.VolumeStatus_VOLUME_STATUS_DOWN
	resp, err = c.NodeGetVolumeStats(
		context.Background(),
		&csi.NodeGetVolumeStatsRequest{VolumeId: id, VolumePath: "/test"})
	assert.NoError(t, err)
	assert.Equal(t, true, resp.VolumeCondition.Abnormal)
	assert.Equal(t, "Volume status is down", resp.VolumeCondition.Message)

	// Get VolumeStats - degraded
	vol.Status = api.VolumeStatus_VOLUME_STATUS_DEGRADED
	resp, err = c.NodeGetVolumeStats(
		context.Background(),
		&csi.NodeGetVolumeStatsRequest{VolumeId: id, VolumePath: "/test"})
	assert.NoError(t, err)
	assert.Equal(t, true, resp.VolumeCondition.Abnormal)
	assert.Equal(t, "Volume status is degraded", resp.VolumeCondition.Message)

	// Get VolumeStats - none
	vol.Status = api.VolumeStatus_VOLUME_STATUS_NONE
	resp, err = c.NodeGetVolumeStats(
		context.Background(),
		&csi.NodeGetVolumeStatsRequest{VolumeId: id, VolumePath: "/test"})
	assert.NoError(t, err)
	assert.Equal(t, true, resp.VolumeCondition.Abnormal)
	assert.Equal(t, "Volume status is unknown", resp.VolumeCondition.Message)

	// Get VolumeStats - not present
	vol.Status = api.VolumeStatus_VOLUME_STATUS_NOT_PRESENT
	resp, err = c.NodeGetVolumeStats(
		context.Background(),
		&csi.NodeGetVolumeStatsRequest{VolumeId: id, VolumePath: "/test"})
	assert.NoError(t, err)
	assert.Equal(t, true, resp.VolumeCondition.Abnormal)
	assert.Equal(t, "Volume status is not present", resp.VolumeCondition.Message)
}

func TestNodeGetVolumeStats_NotFound(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	c := csi.NewNodeClient(s.Conn())

	// Get Capabilities - no volumes found
	id := "myvol123"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{}, nil).
			Times(1),
	)
	_, err := c.NodeGetVolumeStats(
		context.Background(),
		&csi.NodeGetVolumeStatsRequest{VolumeId: id, VolumePath: "/test"})
	assert.Error(t, err)
	statusErr, ok := status.FromError(err)
	assert.Equal(t, true, ok)
	assert.Equal(t, codes.NotFound.String(), statusErr.Code().String())

	// Get Capabilities - err not found
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{}, kvdb.ErrNotFound).
			Times(1),
	)
	_, err = c.NodeGetVolumeStats(
		context.Background(),
		&csi.NodeGetVolumeStatsRequest{VolumeId: id, VolumePath: "/test"})
	assert.Error(t, err)
	statusErr, ok = status.FromError(err)
	assert.Equal(t, true, ok)
	assert.Equal(t, codes.NotFound.String(), statusErr.Code().String())

	// Get Capabilities - attach path does not match VolumePath
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{{
				Id:         id,
				AttachPath: []string{"bad-test", "test-2"},
			}}, nil).
			Times(1),
	)
	_, err = c.NodeGetVolumeStats(
		context.Background(),
		&csi.NodeGetVolumeStatsRequest{VolumeId: id, VolumePath: "/test"})
	assert.Error(t, err)
	statusErr, ok = status.FromError(err)
	assert.Equal(t, true, ok)
	assert.Equal(t, codes.NotFound.String(), statusErr.Code().String())
	assert.Contains(t, statusErr.Err().Error(), "not mounted on path")

}
