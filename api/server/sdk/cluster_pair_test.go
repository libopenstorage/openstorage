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

func TestClusterPairServer_CreateSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create response
	remoteClusterID := "id"
	remoteClusterName := "name"
	remoteClusterIP := "127.0.0.1"
	var remoteClusterPort uint32
	remoteClusterPort = uint32(12345)

	req := api.SdkClusterPairCreateRequest{
		Request: &api.ClusterPairCreateRequest{
			RemoteClusterIp:    remoteClusterIP,
			RemoteClusterPort:  remoteClusterPort,
			RemoteClusterToken: "<Auth-Token>",
			SetDefault:         false,
		},
	}
	resp := &api.ClusterPairCreateResponse{
		RemoteClusterId:   remoteClusterID,
		RemoteClusterName: remoteClusterName,
	}

	s.MockCluster().
		EXPECT().
		CreatePair(&api.ClusterPairCreateRequest{
			RemoteClusterIp:    remoteClusterIP,
			RemoteClusterPort:  remoteClusterPort,
			RemoteClusterToken: "<Auth-Token>",
			SetDefault:         false}).
		Return(resp, nil)

	// Setup client
	c := api.NewOpenStorageClusterPairClient(s.Conn())
	// create the pair
	r, err := c.Create(context.Background(), &req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetResult().GetRemoteClusterId())
	assert.Equal(t, remoteClusterID, r.GetResult().GetRemoteClusterId())
	assert.Equal(t, remoteClusterName, r.GetResult().GetRemoteClusterName())
}
func TestClusterPairServer_CreateFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	noreq := &api.SdkClusterPairCreateRequest{
		Request: &api.ClusterPairCreateRequest{},
	}
	s.MockCluster().
		EXPECT().
		CreatePair(&api.ClusterPairCreateRequest{}).
		Return(nil, status.Errorf(codes.InvalidArgument, "Must supply valid request"))
	// Setup client
	c := api.NewOpenStorageClusterPairClient(s.Conn())
	// create the pair
	//noip
	r, err := c.Create(context.Background(), noreq)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Must supply valid request")
}

func TestClusterPairServer_InspectSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkClusterPairInspectRequest{
		Id: "ID",
	}

	info := &api.ClusterPairInfo{
		Id:     "ID",
		Name:   "test",
		Secure: false,
	}
	resp := &api.ClusterPairGetResponse{
		PairInfo: info,
	}
	s.MockCluster().EXPECT().
		GetPair(req.GetId()).Return(resp, nil)
	// Setup client
	c := api.NewOpenStorageClusterPairClient(s.Conn())
	r, err := c.Inspect(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetResult().GetPairInfo())
}

func TestClusterPairServer_InspectFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkClusterPairInspectRequest{}
	// Setup client
	c := api.NewOpenStorageClusterPairClient(s.Conn())
	r, err := c.Inspect(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply cluster ID")
}

func TestClusterPairServer_EnumerateSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkClusterPairEnumerateRequest{}
	info := &api.ClusterPairInfo{
		Id:     "ID",
		Name:   "test",
		Secure: false,
	}

	pair := make(map[string]*api.ClusterPairInfo)
	pair[info.GetId()] = info

	resp := &api.ClusterPairsEnumerateResponse{
		DefaultId: "ID",
		Pairs:     pair,
	}
	s.MockCluster().EXPECT().EnumeratePairs().Return(resp, nil)
	// Setup client
	c := api.NewOpenStorageClusterPairClient(s.Conn())
	r, err := c.Enumerate(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetResult().GetPairs())
}

func TestClusterPairServer_EnumerateFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	// Setup client
	req := &api.SdkClusterPairEnumerateRequest{}

	s.MockCluster().
		EXPECT().
		EnumeratePairs().
		Return(nil, status.Errorf(codes.Internal, "No Pair Found"))

	c := api.NewOpenStorageClusterPairClient(s.Conn())
	r, err := c.Enumerate(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "No Pair Found")
}

func TestClusterPairServer_GetTokenSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	// Setup client
	req := &api.SdkClusterPairGetTokenRequest{}

	resp := &api.ClusterPairTokenGetResponse{
		Token: "<Auth-Token>",
	}
	s.MockCluster().EXPECT().GetPairToken(false).Return(resp, nil)

	c := api.NewOpenStorageClusterPairClient(s.Conn())
	r, err := c.GetToken(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetResult().GetToken())
}

func TestClusterPairServer_GetTokenFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkClusterPairGetTokenRequest{}

	s.MockCluster().
		EXPECT().
		GetPairToken(false).
		Return(nil, status.Errorf(codes.Internal, "Cannot Generate Token"))
	// Setup client
	c := api.NewOpenStorageClusterPairClient(s.Conn())
	r, err := c.GetToken(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Cannot Generate Token")
}

func TestClusterPairServer_ResetTokenSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	// Setup client
	req := &api.SdkClusterPairResetTokenRequest{}

	resp := &api.ClusterPairTokenGetResponse{
		Token: "<Auth-Token>",
	}
	s.MockCluster().EXPECT().GetPairToken(true).Return(resp, nil)

	c := api.NewOpenStorageClusterPairClient(s.Conn())
	r, err := c.ResetToken(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetResult().GetToken())
}

func TestClusterPairServer_ResetTokenFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkClusterPairResetTokenRequest{}

	s.MockCluster().
		EXPECT().
		GetPairToken(true).
		Return(nil, status.Errorf(codes.Internal, "Cannot Generate Token"))
	// Setup client
	c := api.NewOpenStorageClusterPairClient(s.Conn())
	r, err := c.ResetToken(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Cannot Generate Token")
}

func TestClusterPairServer_DeleteSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkClusterPairDeleteRequest{
		ClusterId: "ID",
	}
	s.MockCluster().EXPECT().DeletePair(req.GetClusterId()).Return(nil)
	// Setup client
	c := api.NewOpenStorageClusterPairClient(s.Conn())
	r, err := c.Delete(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}

func TestClusterPairServer_DeleteFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.SdkClusterPairDeleteRequest{}
	// Setup client
	c := api.NewOpenStorageClusterPairClient(s.Conn())
	r, err := c.Delete(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply valid cluster ID")
}
