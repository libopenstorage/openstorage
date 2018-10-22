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
	"testing"

	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestVolumeMigrate_StartSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.CloudMigrateStartRequest{
		Operation: api.CloudMigrate_MigrateCluster,
		ClusterId: "Source",
		TargetId:  "Target",
	}
	s.MockDriver().EXPECT().
		CloudMigrateStart(&api.CloudMigrateStartRequest{
			Operation: api.CloudMigrate_MigrateCluster,
			ClusterId: "Source",
			TargetId:  "Target",
		}).
		Return(nil)
	// Setup client
	c := api.NewOpenStorageVolumeMigrateClient(s.Conn())
	r, err := c.Start(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}

func TestVolumeMigrate_StartFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	invalidOp := &api.CloudMigrateStartRequest{
		Operation: api.CloudMigrate_InvalidType,
		ClusterId: "Source",
		TargetId:  "Target",
	}
	noSource := &api.CloudMigrateStartRequest{
		Operation: api.CloudMigrate_MigrateVolume,
		TargetId:  "Target",
	}
	noTarget := &api.CloudMigrateStartRequest{
		Operation: api.CloudMigrate_MigrateVolumeGroup,
		ClusterId: "Source",
	}

	// Setup client
	c := api.NewOpenStorageVolumeMigrateClient(s.Conn())

	r, err := c.Start(context.Background(), invalidOp)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply valid Operation")

	r, err = c.Start(context.Background(), noSource)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply valid Cluster ID")

	r, err = c.Start(context.Background(), noTarget)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply valid Target cluster ID")
}

func TestVolumeMigrate_CancelSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.CloudMigrateCancelRequest{
		Operation: api.CloudMigrate_MigrateCluster,
		ClusterId: "Source",
		TargetId:  "Target",
	}
	s.MockDriver().EXPECT().
		CloudMigrateCancel(&api.CloudMigrateCancelRequest{
			Operation: api.CloudMigrate_MigrateCluster,
			ClusterId: "Source",
			TargetId:  "Target",
		}).
		Return(nil)
	// Setup client
	c := api.NewOpenStorageVolumeMigrateClient(s.Conn())
	r, err := c.Cancel(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r)

}

func TestVolumeMigrate_CancelFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	invalidOp := &api.CloudMigrateCancelRequest{
		Operation: api.CloudMigrate_InvalidType,
		ClusterId: "Source",
		TargetId:  "Target",
	}
	noSource := &api.CloudMigrateCancelRequest{
		Operation: api.CloudMigrate_MigrateVolume,
		TargetId:  "Target",
	}
	noTarget := &api.CloudMigrateCancelRequest{
		Operation: api.CloudMigrate_MigrateVolumeGroup,
		ClusterId: "Source",
	}

	// Setup client
	c := api.NewOpenStorageVolumeMigrateClient(s.Conn())

	r, err := c.Cancel(context.Background(), invalidOp)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply valid Operation")

	r, err = c.Cancel(context.Background(), noSource)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply valid Cluster ID")

	r, err = c.Cancel(context.Background(), noTarget)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply valid Target cluster ID")
}

func TestVolumeMigrate_StatusSucess(t *testing.T) {
	// Create server and cl	ient connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkCloudMigrateStatusRequest{}
	info := &api.CloudMigrateInfo{
		ClusterId:       "Source",
		LocalVolumeId:   "VID",
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
		CloudMigrateStatus().
		Return(resp, nil)
	// Setup client
	c := api.NewOpenStorageVolumeMigrateClient(s.Conn())
	r, err := c.Status(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.NotNil(t, r.GetInfo())
}

func TestVolumeMigrate_StatusFailure(t *testing.T) {
	// Create server and cl	ient connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkCloudMigrateStatusRequest{}
	s.MockDriver().EXPECT().
		CloudMigrateStatus().
		Return(nil, status.Errorf(codes.Internal, "Cannot get status of migration"))
	// Setup client
	c := api.NewOpenStorageVolumeMigrateClient(s.Conn())
	r, err := c.Status(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Cannot get status of migration")
}
