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
	"github.com/portworx/kvdb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
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
		Return(nil, volume.ErrEnoEnt).
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
	locator := &api.VolumeLocator{
		Name: id,
		VolumeLabels: map[string]string{
			"hello": "world",
		},
	}
	expect := *locator
	s.MockDriver().
		EXPECT().
		Enumerate(&expect, nil).
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
			Locator: locator,
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
		Locator: &api.VolumeLocator{
			VolumeLabels: newlabels,
		},
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
	req.Locator = nil
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
		Locator: &api.VolumeLocator{
			VolumeLabels: newlabels,
		},
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
