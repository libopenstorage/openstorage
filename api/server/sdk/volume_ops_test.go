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
	"reflect"
	"testing"

	"github.com/kubernetes-csi/csi-test/utils"
	"github.com/sirupsen/logrus"

	"github.com/golang/mock/gomock"
	"github.com/libopenstorage/openstorage/api"
	mockcluster "github.com/libopenstorage/openstorage/cluster/mock"
	"github.com/libopenstorage/openstorage/pkg/auth"
	policy "github.com/libopenstorage/openstorage/pkg/storagepolicy"
	"github.com/libopenstorage/openstorage/volume"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSdkVolumeCreateCheckIdempotency(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	name := "myvol"
	size := uint64(1234)
	req := &api.SdkVolumeCreateRequest{
		Name: name,
		Spec: &api.VolumeSpec{
			Size: size,
		},
	}

	// Create response
	id := "myid"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil).
			Times(1),
	)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.Create(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, r.GetVolumeId(), "myid")
}

func TestSdkVolumeCreate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	name := "myvol"
	size := uint64(1234)
	req := &api.SdkVolumeCreateRequest{
		Name: name,
		Spec: &api.VolumeSpec{
			Size: size,
		},
	}

	// Create response
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
			Create(&api.VolumeLocator{
				Name: name,
			}, &api.Source{}, &api.VolumeSpec{Size: size}).
			Return(id, nil).
			Times(1),
	)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.Create(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, r.GetVolumeId(), "myid")
}

func TestSdkVolumeClone(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	name := "myvol"
	parentid := "myparent"
	parentVol := &api.Volume{
		Id: parentid,
		Spec: &api.VolumeSpec{
			Size: 1234,
		},
		Locator: &api.VolumeLocator{
			Name: parentid,
		},
	}
	cloneVol := parentVol
	req := &api.SdkVolumeCloneRequest{
		Name:     name,
		ParentId: parentid,
	}

	// Create response
	id := "myid"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{parentid}).
			Return([]*api.Volume{parentVol}, nil).
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
			Inspect([]string{parentid}).
			Return([]*api.Volume{parentVol}, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Snapshot(parentid, false, &api.VolumeLocator{Name: name}, false).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{cloneVol}, nil).
			Times(1),
	)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.Clone(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, r.GetVolumeId(), "myid")
}
func TestSdkVolumeDelete(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myvol"
	req := &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	}

	// Create response
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

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Delete(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkVolumeDeleteReturnOkWhenVolumeNotFound(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myvol"
	req := &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return(nil, kvdb.ErrNotFound).
		Times(1)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Delete(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkVolumeDeleteBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkVolumeDeleteRequest{}

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Delete(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "volume id")
}

func TestSdkVolumeInspect(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myid"
	req := &api.SdkVolumeInspectRequest{
		VolumeId: id,
	}

	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{
			&api.Volume{
				Id: id,
			},
		}, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.Inspect(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetVolume())
	assert.Equal(t, r.GetVolume().GetId(), id)
}

func TestSdkVolumeInspectKeyNotFound(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myid"
	req := &api.SdkVolumeInspectRequest{
		VolumeId: id,
	}

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Returns key not found
	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{}, kvdb.ErrNotFound).
		Times(1)

	// Get info
	_, err := c.Inspect(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "not found")

	// Key not found, err is nil but empty list returned
	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{}, nil).
		Times(1)

	// Get info
	_, err = c.Inspect(context.Background(), req)
	assert.Error(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "not found")

	// Other error
	expectedErr := fmt.Errorf("WEIRD ERROR")
	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{}, expectedErr).
		Times(1)

	// Get info
	_, err = c.Inspect(context.Background(), req)
	assert.Error(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "WEIRD ERROR")
}

func TestSdkVolumeEnumerate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myid"
	s.MockDriver().
		EXPECT().
		Enumerate(nil, nil).
		Return([]*api.Volume{
			&api.Volume{
				Id: id,
			},
		}, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.Enumerate(context.Background(), &api.SdkVolumeEnumerateRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetVolumeIds())
	assert.Len(t, r.GetVolumeIds(), 1)
	assert.Equal(t, r.GetVolumeIds()[0], id)
}

func TestSdkVolumeEnumerateWithFilters(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myid"
	labels := map[string]string{
		"hello": "world",
	}
	locator := &api.VolumeLocator{
		Name:         id,
		VolumeLabels: labels,
	}

	s.MockDriver().
		EXPECT().
		Enumerate(locator, nil).
		Return([]*api.Volume{
			&api.Volume{
				Id: id,
			},
		}, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.EnumerateWithFilters(
		context.Background(),
		&api.SdkVolumeEnumerateWithFiltersRequest{
			Name:   id,
			Labels: labels,
		})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetVolumeIds())
	assert.Len(t, r.GetVolumeIds(), 1)
	assert.Equal(t, r.GetVolumeIds()[0], id)
}

func TestSdkVolumeUpdate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myid"
	newlabels := map[string]string{
		"hello": "world",
	}
	req := &api.SdkVolumeUpdateRequest{
		VolumeId: id,
		Labels:   newlabels,
	}

	// Check Locator
	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{&api.Volume{Spec: &api.VolumeSpec{}}}, nil).
		AnyTimes()
	s.MockDriver().
		EXPECT().
		Set(id, &api.VolumeLocator{VolumeLabels: newlabels}, &api.VolumeSpec{}).
		Return(nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Update(context.Background(), req)
	assert.NoError(t, err)

	// Now check spec
	req.Labels = nil
	req.Spec = &api.VolumeSpecUpdate{
		SizeOpt: &api.VolumeSpecUpdate_Size{
			Size: 1234,
		},
	}

	s.MockDriver().
		EXPECT().
		Set(id, nil, &api.VolumeSpec{Size: 1234}).
		Return(nil).
		Times(1)
	_, err = c.Update(context.Background(), req)
	assert.NoError(t, err)

	// Check both locator and spec
	req = &api.SdkVolumeUpdateRequest{
		VolumeId: id,
		Labels:   newlabels,
		Spec: &api.VolumeSpecUpdate{
			SizeOpt: &api.VolumeSpecUpdate_Size{
				Size: 1234,
			},
		},
	}

	s.MockDriver().
		EXPECT().
		Set(
			id,
			&api.VolumeLocator{VolumeLabels: newlabels},
			&api.VolumeSpec{Size: 1234},
		).
		Return(nil).
		Times(1)
	_, err = c.Update(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkVolumeStats(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myid"
	cumulative := true

	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{
			&api.Volume{
				Id: id,
			},
		}, nil).
		Times(1)
	s.MockDriver().
		EXPECT().
		Stats(id, cumulative).
		Return(&api.Stats{
			Reads: 12345,
		}, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.Stats(
		context.Background(),
		&api.SdkVolumeStatsRequest{
			VolumeId: id,
		})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetStats())
	assert.Equal(t, uint64(12345), r.GetStats().GetReads())
}

func TestSdkVolumeStatsBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkVolumeStatsRequest{}

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.Stats(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "volume id")
}

func TestSdkVolumeCapacityUsage(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myid"
	resp := &api.CapacityUsageResponse{}
	resp.CapacityUsageInfo = &api.CapacityUsageInfo{}
	resp.CapacityUsageInfo.ExclusiveBytes = 12000
	resp.CapacityUsageInfo.SharedBytes = 345
	resp.CapacityUsageInfo.TotalBytes = 12345

	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{
			&api.Volume{
				Id: id,
			},
		}, nil).
		Times(1)
	s.MockDriver().
		EXPECT().
		CapacityUsage(id).
		Return(resp, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.CapacityUsage(
		context.Background(),
		&api.SdkVolumeCapacityUsageRequest{
			VolumeId: id,
		})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetCapacityUsageInfo)
}

func TestSdkVolumeCapacityUsageAbortedResult(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	id := "myid"
	resp := &api.CapacityUsageResponse{}
	resp.CapacityUsageInfo = &api.CapacityUsageInfo{}
	resp.CapacityUsageInfo.ExclusiveBytes = 0
	resp.CapacityUsageInfo.SharedBytes = 0
	resp.CapacityUsageInfo.TotalBytes = 12345
	resp.Error = volume.ErrAborted

	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{
			&api.Volume{
				Id: id,
			},
		}, nil).
		Times(1)
	s.MockDriver().
		EXPECT().
		CapacityUsage(id).
		Return(resp, nil).
		Times(1)
	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.CapacityUsage(
		context.Background(),
		&api.SdkVolumeCapacityUsageRequest{
			VolumeId: id,
		})
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Aborted)
	assert.NotNil(t, r.GetCapacityUsageInfo)
}

func TestSdkVolumeCapacityUsageUnimplementedResult(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	id := "myid"
	resp := &api.CapacityUsageResponse{}
	resp.CapacityUsageInfo = &api.CapacityUsageInfo{}
	resp.CapacityUsageInfo.ExclusiveBytes = 0
	resp.CapacityUsageInfo.SharedBytes = 0
	resp.CapacityUsageInfo.TotalBytes = 12345
	resp.Error = volume.ErrNotSupported

	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{
			&api.Volume{
				Id: id,
			},
		}, nil).
		Times(1)
	s.MockDriver().
		EXPECT().
		CapacityUsage(id).
		Return(resp, nil).
		Times(1)
	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.CapacityUsage(
		context.Background(),
		&api.SdkVolumeCapacityUsageRequest{
			VolumeId: id,
		})
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Unimplemented)
	assert.NotNil(t, r.GetCapacityUsageInfo)
}

func TestSdkDeleteOnlyByOwner(t *testing.T) {
	// This test does not use the gRPC server
	mc := gomock.NewController(&utils.SafeGoroutineTester{})
	mv := mockdriver.NewMockVolumeDriver(mc)
	mcluster := mockcluster.NewMockCluster(mc)
	s := VolumeServer{
		server: &sdkGrpcServer{
			driverHandlers: map[string]volume.VolumeDriver{
				"mock":            mv,
				DefaultDriverName: mv,
			},
			clusterHandler: mcluster,
		},
	}

	// Create volumes
	vauth := &api.Volume{
		Spec: &api.VolumeSpec{
			Ownership: &api.Ownership{
				Owner: "testowner",
				Acls: &api.Ownership_AccessControl{
					Collaborators: map[string]api.Ownership_AccessType{
						"notmyname": api.Ownership_Write,
						"trusted":   api.Ownership_Admin,
					},
				},
			},
		},
	}
	vpublic := &api.Volume{Spec: &api.VolumeSpec{}}

	// Create contexts
	ctxNoAuth := context.Background()
	ctxWithNotOwner := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "notmyname",
	})
	ctxWithTrusted := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "trusted",
	})

	// -- test: should work on no auth and public vol
	id := "volid"
	mv.
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{vpublic}, nil)
	mv.
		EXPECT().
		Delete(id).
		Return(nil)
	_, err := s.Delete(ctxNoAuth, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

	// -- test: should work, with auth public vol
	mv.
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{vpublic}, nil)
	mv.
		EXPECT().
		Delete(id).
		Return(nil)
	_, err = s.Delete(ctxWithNotOwner, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

	// -- test: should work, no auth and owned vol
	mv.
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{vauth}, nil)
	mv.
		EXPECT().
		Delete(id).
		Return(nil)
	_, err = s.Delete(ctxNoAuth, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

	// -- test: should not work, auth with non-public vol for collaborator
	// with non-api.Ownship_Admin rights
	mv.
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{vauth}, nil)
	_, err = s.Delete(ctxWithNotOwner, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.Error(t, err)

	// -- test: should work, auth with non-public vol for trusted admin collaborator
	mv.
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{vauth}, nil)
	mv.
		EXPECT().
		Delete(id).
		Return(nil)
	_, err = s.Delete(ctxWithTrusted, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

}

func TestSdkCloneOwnership(t *testing.T) {
	// This test does not use the gRPC server
	mc := gomock.NewController(&utils.SafeGoroutineTester{})
	mv := mockdriver.NewMockVolumeDriver(mc)
	mcluster := mockcluster.NewMockCluster(mc)
	s := VolumeServer{
		server: &sdkGrpcServer{
			driverHandlers: map[string]volume.VolumeDriver{
				"mock":            mv,
				DefaultDriverName: mv,
			},
			clusterHandler: mcluster,
		},
	}

	// Create volumes
	name := "myvol"
	parentid := "myparent"
	parentVol := &api.Volume{
		Id: parentid,
		Spec: &api.VolumeSpec{
			Size: 1234,
		},
		Locator: &api.VolumeLocator{
			Name: parentid,
		},
	}
	cloneVol := parentVol
	req := &api.SdkVolumeCloneRequest{
		Name:     name,
		ParentId: parentid,
	}

	// Create contexts
	user1 := "user1"
	user2 := "user2"
	ctxNoAuth := context.Background()
	ctxWithNotOwner := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: user2,
	})

	// -- test: no auth, Public volume no ownership transferred
	// Update will not be called
	id := "myid"
	gomock.InOrder(
		mv.
			EXPECT().
			Inspect([]string{parentid}).
			Return([]*api.Volume{parentVol}, nil).
			Times(1),

		mv.
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.
			EXPECT().
			Inspect([]string{parentid}).
			Return([]*api.Volume{parentVol}, nil).
			Times(1),

		mv.
			EXPECT().
			Snapshot(parentid, false, &api.VolumeLocator{Name: name}, false).
			Return(id, nil).
			Times(1),

		mv.
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{cloneVol}, nil).
			Times(1),
	)
	cloneId, err := s.Clone(ctxNoAuth, req)
	assert.NoError(t, err)
	assert.NotEmpty(t, cloneId)

	// -- test: auth, Public volume new owner
	// Update will be called
	gomock.InOrder(
		mv.
			EXPECT().
			Inspect([]string{parentid}).
			Return([]*api.Volume{parentVol}, nil).
			Times(1),

		mv.
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.
			EXPECT().
			Inspect([]string{parentid}).
			Return([]*api.Volume{parentVol}, nil).
			Times(1),

		mv.
			EXPECT().
			Snapshot(parentid, false, &api.VolumeLocator{Name: name}, false).
			Return(id, nil).
			Times(1),

		mv.
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{cloneVol}, nil).
			Times(1),

		// By Update since it is a new owner
		mv.
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{cloneVol}, nil).
			Times(1),
		mv.
			EXPECT().
			Set(id, nil, &api.VolumeSpec{
				Size: 1234,
				Ownership: &api.Ownership{
					Owner: user2,
				},
			}).
			Return(nil).
			Times(1),
	)
	cloneId, err = s.Clone(ctxWithNotOwner, req)
	assert.NoError(t, err)
	assert.NotEmpty(t, cloneId)

	// -- test: auth, owned volume same owner
	// Update will not be called
	parentVol.Spec.Ownership = &api.Ownership{
		Owner: user2,
	}

	gomock.InOrder(
		mv.
			EXPECT().
			Inspect([]string{parentid}).
			Return([]*api.Volume{parentVol}, nil).
			Times(1),

		mv.
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.
			EXPECT().
			Inspect([]string{parentid}).
			Return([]*api.Volume{parentVol}, nil).
			Times(1),

		mv.
			EXPECT().
			Snapshot(parentid, false, &api.VolumeLocator{Name: name}, false).
			Return(id, nil).
			Times(1),

		mv.
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{cloneVol}, nil).
			Times(1),
	)
	cloneId, err = s.Clone(ctxWithNotOwner, req)
	assert.NoError(t, err)
	assert.NotEmpty(t, cloneId)

	// -- test: auth, Public volume new owner
	// Update will be called
	parentVol.Spec.Ownership = &api.Ownership{
		Owner: user1,
		Acls: &api.Ownership_AccessControl{
			Collaborators: map[string]api.Ownership_AccessType{
				user2: api.Ownership_Read,
			},
		},
	}

	gomock.InOrder(
		mv.
			EXPECT().
			Inspect([]string{parentid}).
			Return([]*api.Volume{parentVol}, nil).
			Times(1),

		mv.
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.
			EXPECT().
			Inspect([]string{parentid}).
			Return([]*api.Volume{parentVol}, nil).
			Times(1),

		mv.
			EXPECT().
			Snapshot(parentid, false, &api.VolumeLocator{Name: name}, false).
			Return(id, nil).
			Times(1),

		mv.
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{cloneVol}, nil).
			Times(1),

		// By Update since it is a new owner
		mv.
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{cloneVol}, nil).
			Times(1),
		mv.
			EXPECT().
			Set(id, nil, &api.VolumeSpec{
				Size: 1234,
				Ownership: &api.Ownership{
					Owner: user2,
				},
			}).
			Return(nil).
			Times(1),
	)
	cloneId, err = s.Clone(ctxWithNotOwner, req)
	assert.NoError(t, err)
	assert.NotEmpty(t, cloneId)
}

// check volume create after storage policy is set as default
func TestSdkVolumeCreateEnforced(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create storage policy and set it as default Enforcement
	storePolicy, err := policy.Inst()
	assert.NoError(t, err)
	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 8123,
		},
		SizeOperator: api.VolumeSpecPolicy_Maximum,
		SharedOpt: &api.VolumeSpecPolicy_Shared{
			Shared: false,
		},
		Sharedv4Opt: &api.VolumeSpecPolicy_Sharedv4{
			Sharedv4: false,
		},
		JournalOpt: &api.VolumeSpecPolicy_Journal{
			Journal: true,
		},
		HaLevelOpt: &api.VolumeSpecPolicy_HaLevel{
			HaLevel: 2,
		},
		HaLevelOperator: api.VolumeSpecPolicy_Minimum,
	}

	req := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testrestvolcreate",
			Policy: volSpec,
			Force:  false,
		},
	}

	_, err = storePolicy.Create(context.Background(), req)
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{
		Name: "testrestvolcreate",
	}

	resp, err := storePolicy.Inspect(context.Background(), inspReq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StoragePolicy.GetName(), inspReq.GetName())
	assert.True(t, reflect.DeepEqual(resp.StoragePolicy.GetPolicy(), req.StoragePolicy.GetPolicy()))

	defaultReq := &api.SdkOpenStoragePolicySetDefaultRequest{
		Name: inspReq.GetName(),
	}
	_, err = storePolicy.SetDefault(context.Background(), defaultReq)
	assert.NoError(t, err)

	policy, err := storePolicy.DefaultInspect(context.Background(), &api.SdkOpenStoragePolicyDefaultInspectRequest{})
	assert.NoError(t, err)
	assert.Equal(t, policy.GetStoragePolicy().GetName(), inspReq.GetName())

	// create volume with policy enabled
	name := "myvol"
	size := uint64(1123234)
	volReq := &api.SdkVolumeCreateRequest{
		Name: name,
		Spec: &api.VolumeSpec{
			Size:    size,
			HaLevel: 3,
		},
	}

	// Ideal spec should be passed to volume create after applying
	// policy specs
	updatedSpec := &api.VolumeSpec{
		Size:     volSpec.GetSize(),
		Shared:   volSpec.GetShared(),
		Sharedv4: volSpec.GetSharedv4(),
		Journal:  volSpec.GetJournal(),
		HaLevel:  volReq.GetSpec().GetHaLevel(),
		// since voluem is created as per default policy
		StoragePolicy: "testrestvolcreate",
	}

	// Create response
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
			Create(&api.VolumeLocator{
				Name: name,
			}, &api.Source{}, updatedSpec).
			Return(id, nil).
			Times(1),
	)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.Create(context.Background(), volReq)
	assert.NoError(t, err)
	assert.Equal(t, r.GetVolumeId(), "myid")

	releaseReq := &api.SdkOpenStoragePolicyReleaseRequest{}
	_, err = storePolicy.Release(context.Background(), releaseReq)
	assert.NoError(t, err)
}

// check volume create with ownership after storage policy is set as default
func TestSdkVolumeCreateDefaultPolicyOwnership(t *testing.T) {
	mc := gomock.NewController(&utils.SafeGoroutineTester{})
	mv := mockdriver.NewMockVolumeDriver(mc)
	mcluster := mockcluster.NewMockCluster(mc)
	kv, err := kvdb.New(mem.Name, "policy", []string{}, nil, logrus.Panicf)
	assert.NoError(t, err)

	// Init storage policy manager
	_, err = policy.Init(kv)
	storePolicy, err := policy.Inst()
	assert.NoError(t, err)

	s := VolumeServer{
		server: &sdkGrpcServer{
			driverHandlers: map[string]volume.VolumeDriver{
				"mock":            mv,
				DefaultDriverName: mv,
			},
			clusterHandler: mcluster,
			policyServer:   storePolicy,
		},
	}

	// Create storage policy and set it as default Enforcement
	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 8123,
		},
		SizeOperator: api.VolumeSpecPolicy_Maximum,
		SharedOpt: &api.VolumeSpecPolicy_Shared{
			Shared: false,
		},
		HaLevelOpt: &api.VolumeSpecPolicy_HaLevel{
			HaLevel: 2,
		},
		HaLevelOperator: api.VolumeSpecPolicy_Minimum,
	}

	spAreq := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testpolicyA",
			Policy: volSpec,
			Force:  false,
			Ownership: &api.Ownership{
				Owner: "testowner",
				Acls: &api.Ownership_AccessControl{
					Collaborators: map[string]api.Ownership_AccessType{
						"notmyname": api.Ownership_Write,
						"trusted":   api.Ownership_Admin,
					},
				},
			},
		},
	}

	spBreq := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testpolicyB",
			Policy: volSpec,
			Force:  false,
			Ownership: &api.Ownership{
				Owner: "testownerB",
				Acls: &api.Ownership_AccessControl{
					Collaborators: map[string]api.Ownership_AccessType{
						"userR":     api.Ownership_Read,
						"userW":     api.Ownership_Write,
						"useradmin": api.Ownership_Admin,
					},
				},
			},
		},
	}

	// Create contexts
	// ctxNoAuth := context.Background()
	ctxWithNotOwner := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "notuser",
	})
	ctxWithTrusted := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "trusted",
		Claims: auth.Claims{
			Groups: []string{"*"},
		},
	})
	ctxWithPolicyB := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "userW",
	})

	// create spA
	_, err = storePolicy.Create(ctxWithTrusted, spAreq)
	assert.NoError(t, err)
	// create spB
	_, err = storePolicy.Create(ctxWithPolicyB, spBreq)
	assert.NoError(t, err)

	defaultReq := &api.SdkOpenStoragePolicySetDefaultRequest{
		Name: "testpolicyA",
	}
	_, err = storePolicy.SetDefault(ctxWithTrusted, defaultReq)
	assert.NoError(t, err)

	policy, err := storePolicy.DefaultInspect(ctxWithTrusted, &api.SdkOpenStoragePolicyDefaultInspectRequest{})
	assert.NoError(t, err)
	assert.Equal(t, policy.GetStoragePolicy().GetName(), "testpolicyA")

	// create volume with policy enabled
	name := "myvol"
	size := uint64(1123234)
	volInvalidReq := &api.SdkVolumeCreateRequest{
		Name: name,
		Spec: &api.VolumeSpec{
			Size:          size,
			HaLevel:       3,
			StoragePolicy: "testpolicyB",
		},
	}
	volValidReq := &api.SdkVolumeCreateRequest{
		Name: name,
		Spec: &api.VolumeSpec{
			Size:    size,
			HaLevel: 3,
		},
	}

	owner := &api.Ownership{
		Owner: "notuser",
	}
	// Ideal spec should be passed to volume create after applying
	// policy specs
	updatedSpec := &api.VolumeSpec{
		Size:    volSpec.GetSize(),
		Shared:  volSpec.GetShared(),
		HaLevel: volValidReq.GetSpec().GetHaLevel(),
		// since voluem is created as per default policy
		StoragePolicy: "testpolicyA",
		Ownership:     owner,
	}

	// Create response
	id := "myid"
	gomock.InOrder(
		mv.EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.EXPECT().
			Create(&api.VolumeLocator{
				Name: name,
			}, &api.Source{}, updatedSpec).
			Return(id, nil).
			Times(1),
	)

	// Use case 1
	// user dont; have access to storagePolicyB try to create volume with it
	_, err = s.Create(ctxWithNotOwner, volInvalidReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Access denied to storage policy ")

	// create with valid volume req, since testpolicyA is set to default
	// vol request should follow spA
	_, err = s.Create(ctxWithNotOwner, volValidReq)
	assert.NoError(t, err)

	// Use case 2
	// no default policy
	releaseReq := &api.SdkOpenStoragePolicyReleaseRequest{}
	_, err = storePolicy.Release(ctxWithTrusted, releaseReq)
	assert.NoError(t, err)

	nopolVol := &api.SdkVolumeCreateRequest{
		Name: name,
		Spec: &api.VolumeSpec{
			Size:    size,
			HaLevel: 3,
		},
	}
	// Ideal spec should be passed to volume create after applying
	// policy specs
	upSpec := &api.VolumeSpec{
		Size:      nopolVol.GetSpec().GetSize(),
		Shared:    nopolVol.GetSpec().GetShared(),
		HaLevel:   nopolVol.GetSpec().GetHaLevel(),
		Ownership: owner,
	}

	// Create response
	gomock.InOrder(
		mv.EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.EXPECT().
			Create(&api.VolumeLocator{
				Name: name,
			}, &api.Source{}, upSpec).
			Return(id, nil).
			Times(1),
	)

	// create with valid volume req, since no policy is set
	// vol req will use their own specs
	_, err = s.Create(ctxWithNotOwner, nopolVol)
	assert.NoError(t, err)

	_, err = storePolicy.Delete(ctxWithTrusted, &api.SdkOpenStoragePolicyDeleteRequest{
		Name: "testpolicyA",
	})

	assert.NoError(t, err)

	_, err = storePolicy.Delete(ctxWithTrusted, &api.SdkOpenStoragePolicyDeleteRequest{
		Name: "testpolicyB",
	})
	assert.NoError(t, err)
}

func TestSdkVolumeUpdatePolicyOwnership(t *testing.T) {
	mc := gomock.NewController(&utils.SafeGoroutineTester{})
	mv := mockdriver.NewMockVolumeDriver(mc)
	mcluster := mockcluster.NewMockCluster(mc)

	kv, err := kvdb.New(mem.Name, "policy", []string{}, nil, logrus.Panicf)
	assert.NoError(t, err)

	// Init storage policy manager
	_, err = policy.Init(kv)
	storePolicy, err := policy.Inst()
	assert.NoError(t, err)

	s := VolumeServer{
		server: &sdkGrpcServer{
			driverHandlers: map[string]volume.VolumeDriver{
				"mock":            mv,
				DefaultDriverName: mv,
			},
			clusterHandler: mcluster,
			policyServer:   storePolicy,
		},
	}

	// Create storage policy and set it as default storage policy
	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 8123,
		},
		SizeOperator: api.VolumeSpecPolicy_Maximum,
		SharedOpt: &api.VolumeSpecPolicy_Shared{
			Shared: false,
		},
		HaLevelOpt: &api.VolumeSpecPolicy_HaLevel{
			HaLevel: 2,
		},
		HaLevelOperator: api.VolumeSpecPolicy_Minimum,
	}

	spAreq := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testpolicyA",
			Policy: volSpec,
			// we will overrride given vol specs
			Force: false,
			Ownership: &api.Ownership{
				Owner: "testowner",
				Acls: &api.Ownership_AccessControl{
					Collaborators: map[string]api.Ownership_AccessType{
						"notmyname": api.Ownership_Write,
						"trusted":   api.Ownership_Admin,
					},
				},
			},
		},
	}

	// Create contexts
	ctxWithTrusted := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "trusted",
		Claims: auth.Claims{
			Groups: []string{"*"},
		},
	})
	ctxWithNoUser := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "notspuser",
	})

	// create spA & set it as default
	_, err = storePolicy.Create(ctxWithTrusted, spAreq)
	assert.NoError(t, err)

	defaultReq := &api.SdkOpenStoragePolicySetDefaultRequest{
		Name: "testpolicyA",
	}
	_, err = storePolicy.SetDefault(ctxWithTrusted, defaultReq)
	assert.NoError(t, err)

	policy, err := storePolicy.DefaultInspect(ctxWithTrusted, &api.SdkOpenStoragePolicyDefaultInspectRequest{})
	assert.NoError(t, err)
	assert.Equal(t, policy.GetStoragePolicy().GetName(), "testpolicyA")

	// create volume with policy enabled
	name := "myvol"
	size := uint64(1123234)
	volReq := &api.SdkVolumeCreateRequest{
		Name: name,
		Spec: &api.VolumeSpec{
			Size:    size,
			HaLevel: 1,
		},
	}

	owner := &api.Ownership{
		Owner: "notspuser",
	}
	// Ideal spec should be passed to volume create after applying
	// policy specs
	volPolSpec := &api.VolumeSpec{
		Size:    volSpec.GetSize(),
		Shared:  volSpec.GetShared(),
		HaLevel: volSpec.GetHaLevel(),
		// since volume is created as per default policy
		StoragePolicy: "testpolicyA",
		Ownership:     owner,
	}

	// Create response
	id := "myid"
	gomock.InOrder(
		mv.EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		mv.EXPECT().
			Create(&api.VolumeLocator{
				Name: name,
			}, &api.Source{}, volPolSpec).
			Return(id, nil).
			Times(1),
	)

	_, err = s.Create(ctxWithNoUser, volReq)
	assert.NoError(t, err)

	// violates rule
	updatVolReq := &api.SdkVolumeUpdateRequest{
		VolumeId: id,
		Spec: &api.VolumeSpecUpdate{
			HaLevelOpt: &api.VolumeSpecUpdate_HaLevel{
				HaLevel: 1,
			},
		},
	}

	// Check Locator
	mv.
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{&api.Volume{Spec: volPolSpec}}, nil).
		AnyTimes()
	mv.
		EXPECT().
		Set(id, nil, volPolSpec).
		Return(nil).
		Times(1)

	_, err = s.Update(ctxWithNoUser, updatVolReq)
	assert.NoError(t, err)

	_, err = storePolicy.Delete(ctxWithTrusted, &api.SdkOpenStoragePolicyDeleteRequest{Name: "testpolicyA"})
	assert.NoError(t, err)
}
