/*
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

	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestVolumeMigrate_StartVolumeSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkCloudMigrateStartRequest{
		ClusterId: "Source",
		Opt: &api.SdkCloudMigrateStartRequest_Volume{
			Volume: &api.SdkCloudMigrateStartRequest_MigrateVolume{
				VolumeId: "Target",
			},
		},
	}

	resp := &api.Volume{
		Locator: &api.VolumeLocator{
			Name: "Target",
		},
	}

	s.MockDriver().EXPECT().
		Inspect([]string{"Target"}).
		Return([]*api.Volume{resp}, nil)

	resp2 := &api.CloudMigrateStartResponse{
		TaskId: "1",
	}

	s.MockDriver().EXPECT().
		CloudMigrateStart(&api.CloudMigrateStartRequest{
			Operation: api.CloudMigrate_MigrateVolume,
			ClusterId: "Source",
			TargetId:  "Target",
		}).
		Return(resp2, nil)
	// Setup client
	c := api.NewOpenStorageMigrateClient(s.Conn())
	r, err := c.Start(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}
func TestVolumeMigrate_StartVolumeGroupSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkCloudMigrateStartRequest{
		ClusterId: "Source",
		Opt: &api.SdkCloudMigrateStartRequest_VolumeGroup{
			VolumeGroup: &api.SdkCloudMigrateStartRequest_MigrateVolumeGroup{
				GroupId: "Target",
			},
		},
	}
	labels := make(map[string]string, 0)

	resp := &api.Volume{
		Id: "Target",
		Locator: &api.VolumeLocator{
			Name:         "Target",
			VolumeLabels: labels,
			Group: &api.Group{
				Id: "Target",
			},
		},
	}

	// Enumerate all volumes that have desired group
	s.MockDriver().EXPECT().
		Enumerate(&api.VolumeLocator{
			Group: &api.Group{
				Id: "Target",
			},
		}, nil).
		Return([]*api.Volume{resp}, nil)

	resp2 := &api.CloudMigrateStartResponse{
		TaskId: "1",
	}

	s.MockDriver().EXPECT().
		CloudMigrateStart(&api.CloudMigrateStartRequest{
			Operation: api.CloudMigrate_MigrateVolumeGroup,
			ClusterId: "Source",
			TargetId:  "Target",
		}).
		Return(resp2, nil)

	// Setup client
	c := api.NewOpenStorageMigrateClient(s.Conn())
	r, err := c.Start(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}
func TestVolumeMigrate_StartAllVolumeFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkCloudMigrateStartRequest{
		ClusterId: "Source",
		TaskId:    "1",
		Opt: &api.SdkCloudMigrateStartRequest_AllVolumes{
			AllVolumes: &api.SdkCloudMigrateStartRequest_MigrateAllVolumes{},
		},
	}

	resp := &api.Volume{
		Id: "Target",
		Locator: &api.VolumeLocator{
			Name: "Target",
		},
	}

	// Enumerate all volumes that have desired group
	s.MockDriver().EXPECT().
		Enumerate(nil, nil).
		Return([]*api.Volume{resp}, nil)

	s.MockDriver().EXPECT().
		CloudMigrateStart(&api.CloudMigrateStartRequest{
			Operation: api.CloudMigrate_MigrateCluster,
			ClusterId: "Source",
			TaskId:    "1",
		}).
		Return(nil, fmt.Errorf("Cannot start migration"))

	// Setup client
	c := api.NewOpenStorageMigrateClient(s.Conn())
	_, err := c.Start(context.Background(), req)
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Cannot start migration")
}

func TestVolumeMigrate_StartVolumeGroupFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkCloudMigrateStartRequest{
		ClusterId: "Source",
		TaskId:    "1",
		Opt: &api.SdkCloudMigrateStartRequest_VolumeGroup{
			VolumeGroup: &api.SdkCloudMigrateStartRequest_MigrateVolumeGroup{
				GroupId: "Target",
			},
		},
	}

	labels := make(map[string]string, 0)
	labels["group"] = "Target"

	resp := &api.Volume{
		Id: "Target",
		Locator: &api.VolumeLocator{
			Name:         "Target",
			VolumeLabels: labels,
		},
	}

	// Enumerate all volumes that have desired group
	s.MockDriver().EXPECT().
		Enumerate(&api.VolumeLocator{
			Group: &api.Group{
				Id: "Target",
			},
		}, nil).
		Return([]*api.Volume{resp}, nil)

	s.MockDriver().EXPECT().
		CloudMigrateStart(&api.CloudMigrateStartRequest{
			Operation: api.CloudMigrate_MigrateVolumeGroup,
			ClusterId: "Source",
			TargetId:  "Target",
			TaskId:    "1",
		}).
		Return(nil, fmt.Errorf("Cannot start migration"))

	// Setup client
	c := api.NewOpenStorageMigrateClient(s.Conn())
	_, err := c.Start(context.Background(), req)
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Cannot start migration")
}
func TestVolumeMigrate_StartVolumeFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkCloudMigrateStartRequest{
		ClusterId: "Source",
		TaskId:    "1",
		Opt: &api.SdkCloudMigrateStartRequest_Volume{
			Volume: &api.SdkCloudMigrateStartRequest_MigrateVolume{
				VolumeId: "Target",
			},
		},
	}

	resp := &api.Volume{
		Id: "Target",
		Locator: &api.VolumeLocator{
			Name: "Target",
		},
	}

	// Inspect volumes to get their ownership
	s.MockDriver().EXPECT().
		Inspect([]string{"Target"}).
		Return([]*api.Volume{resp}, nil)

	s.MockDriver().EXPECT().
		CloudMigrateStart(&api.CloudMigrateStartRequest{
			Operation: api.CloudMigrate_MigrateVolume,
			ClusterId: "Source",
			TargetId:  "Target",
			TaskId:    "1",
		}).
		Return(nil, fmt.Errorf("Cannot start migration"))

	// Setup client
	c := api.NewOpenStorageMigrateClient(s.Conn())
	_, err := c.Start(context.Background(), req)
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Cannot start migration")
}
func TestVolumeMigrate_StartAllVolumeSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkCloudMigrateStartRequest{
		ClusterId: "Source",
		Opt: &api.SdkCloudMigrateStartRequest_AllVolumes{
			AllVolumes: &api.SdkCloudMigrateStartRequest_MigrateAllVolumes{},
		},
	}

	resp := &api.Volume{
		Id: "Target",
		Locator: &api.VolumeLocator{
			Name: "Target",
		},
	}

	// Enumerate all volumes that have desired group
	s.MockDriver().EXPECT().
		Enumerate(nil, nil).
		Return([]*api.Volume{resp}, nil)

	resp2 := &api.CloudMigrateStartResponse{
		TaskId: "1",
	}

	s.MockDriver().EXPECT().
		CloudMigrateStart(&api.CloudMigrateStartRequest{
			Operation: api.CloudMigrate_MigrateCluster,
			ClusterId: "Source",
		}).
		Return(resp2, nil)

	// Setup client
	c := api.NewOpenStorageMigrateClient(s.Conn())
	r, err := c.Start(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}
func TestVolumeMigrate_CancelSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	taskId := "1"
	req := &api.SdkCloudMigrateCancelRequest{
		Request: &api.CloudMigrateCancelRequest{
			TaskId: taskId,
		},
	}

	resp := &api.CloudMigrateStatusResponse{}
	s.MockDriver().EXPECT().
		CloudMigrateStatus(&api.CloudMigrateStatusRequest{
			TaskId: taskId,
		}).
		Return(resp, nil)

	s.MockDriver().EXPECT().
		CloudMigrateCancel(&api.CloudMigrateCancelRequest{
			TaskId: taskId,
		}).
		Return(nil)
	// Setup client
	c := api.NewOpenStorageMigrateClient(s.Conn())
	r, err := c.Cancel(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r)

}

func TestVolumeMigrate_CancelFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	invalidOp := &api.SdkCloudMigrateCancelRequest{
		Request: &api.CloudMigrateCancelRequest{},
	}

	noSource := &api.SdkCloudMigrateCancelRequest{
		Request: &api.CloudMigrateCancelRequest{
			TaskId: "",
		},
	}

	// Setup client
	c := api.NewOpenStorageMigrateClient(s.Conn())

	r, err := c.Cancel(context.Background(), invalidOp)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply valid Task ID")

	r, err = c.Cancel(context.Background(), noSource)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply valid Task ID")

}

func TestVolumeMigrate_StatusSucess(t *testing.T) {
	// Create server and cl	ient connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkCloudMigrateStatusRequest{
		Request: &api.CloudMigrateStatusRequest{},
	}
	vId := "VID"
	info := &api.CloudMigrateInfo{
		ClusterId:       "Source",
		LocalVolumeId:   vId,
		LocalVolumeName: "VNAME",
		RemoteVolumeId:  "RID",
		CloudbackupId:   "CBKUPID",
		CurrentStage:    api.CloudMigrate_Backup,
		Status:          api.CloudMigrate_Queued,
	}
	ll := make([]*api.CloudMigrateInfo, 0)
	ll = append(ll, info)
	l := api.CloudMigrateInfoList{
		List: ll,
	}
	infoList := make(map[string]*api.CloudMigrateInfoList)
	infoList["Source"] = &l
	resp := &api.CloudMigrateStatusResponse{
		Info: infoList,
	}
	s.MockDriver().EXPECT().
		CloudMigrateStatus(&api.CloudMigrateStatusRequest{}).
		Return(resp, nil)

	inspectResp := &api.Volume{
		Id: vId,
	}
	s.MockDriver().EXPECT().
		Inspect([]string{vId}).
		Return([]*api.Volume{inspectResp}, nil)

	// Setup client
	c := api.NewOpenStorageMigrateClient(s.Conn())
	r, err := c.Status(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.NotNil(t, r.GetResult().GetInfo())
}

func TestVolumeMigrate_StatusFailure(t *testing.T) {
	// Create server and cl	ient connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkCloudMigrateStatusRequest{
		Request: &api.CloudMigrateStatusRequest{},
	}
	s.MockDriver().EXPECT().
		CloudMigrateStatus(&api.CloudMigrateStatusRequest{}).
		Return(nil, status.Errorf(codes.Internal, "Cannot get status of migration"))
	// Setup client
	c := api.NewOpenStorageMigrateClient(s.Conn())
	r, err := c.Status(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Cannot get status of migration")
}
