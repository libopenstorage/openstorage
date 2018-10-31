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
	assert.NotNil(t, r.GetRemoteClusterId())
	assert.Equal(t, remoteClusterID, r.GetResult().GetRemoteClusterId())
	assert.Equal(t, remoteClusterName, r.GetResult().GetRemoteClusterName())
}
func TestClusterPairServer_CreateFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create response
	remoteClusterIP := "127.0.0.1"
	var remoteClusterPort uint32
	remoteClusterPort = uint32(12345)

	noip := &api.ClusterPairCreateRequest{
		RemoteClusterPort:  remoteClusterPort,
		RemoteClusterToken: "<Auth-Token>",
		SetDefault:         false,
	}

	noport := &api.ClusterPairCreateRequest{
		RemoteClusterIp:    remoteClusterIP,
		RemoteClusterToken: "<Auth-Token>",
		SetDefault:         false,
	}

	notoken := &api.ClusterPairCreateRequest{
		RemoteClusterIp:   remoteClusterIP,
		RemoteClusterPort: remoteClusterPort,
		SetDefault:        false,
	}

	// Setup client
	c := api.NewOpenStorageClusterPairClient(s.Conn())
	// create the pair
	//noip
	r, err := c.Create(context.Background(), noip)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Remote cluster IP")

	r, err = c.Create(context.Background(), noport)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Remote cluster Port")

	r, err = c.Create(context.Background(), notoken)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Authentication Token")

}

func TestClusterPairServer_ProcessRequestSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.ClusterPairProcessRequest{
		SourceClusterId:    "source",
		RemoteClusterToken: "<Auth-Token>",
	}
	var endpoints = make([]string, 0)

	endpoints = append(endpoints, "ep1:123")
	endpoints = append(endpoints, "ep2:456")
	endpoints = append(endpoints, "ep3:789")

	resp := &api.ClusterPairProcessResponse{
		RemoteClusterId:        "remote",
		RemoteClusterName:      "remote",
		RemoteClusterEndpoints: endpoints,
		Options:                make(map[string]string, 0),
	}
	s.MockCluster().
		EXPECT().ProcessPairRequest(&api.ClusterPairProcessRequest{
		SourceClusterId:    "source",
		RemoteClusterToken: "<Auth-Token>",
	}).Return(resp, nil)
	// Setup client
	c := api.NewOpenStorageClusterPairClient(s.Conn())
	// create the pair
	r, err := c.ProcessRequest(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetRemoteClusterId())
	assert.Equal(t, "remote", r.GetRemoteClusterId())
	assert.Equal(t, "remote", r.GetRemoteClusterName())
	assert.Equal(t, endpoints, r.GetRemoteClusterEndpoints())
}

func TestClusterPairServer_ProcessRequestFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	noSource := &api.ClusterPairProcessRequest{
		RemoteClusterToken: "<Auth-Token>",
	}

	noToken := &api.ClusterPairProcessRequest{
		SourceClusterId: "source",
	}
	// Setup client
	c := api.NewOpenStorageClusterPairClient(s.Conn())

	r, err := c.ProcessRequest(context.Background(), noSource)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Source cluster ID")

	r, err = c.ProcessRequest(context.Background(), noToken)
	assert.Error(t, err)
	assert.Nil(t, r)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Remote cluster Token")
}

func TestClusterPairServer_GetSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.ClusterPairGetRequest{
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
	r, err := c.Get(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetPairInfo())
}

func TestClusterPairServer_GetFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.ClusterPairGetRequest{}
	// Setup client
	c := api.NewOpenStorageClusterPairClient(s.Conn())
	r, err := c.Get(context.Background(), req)
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
	req := &api.SdkClusterPairsEnumerateRequest{}
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
	assert.NotNil(t, r.GetPairs())
}

func TestClusterPairServer_EnumerateFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	// Setup client
	req := &api.SdkClusterPairsEnumerateRequest{}

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

func TestClusterPairServer_TokenGetSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	// Setup client
	req := &api.ClusterPairTokenGetRequest{
		ResetToken: true,
	}
	resp := &api.ClusterPairTokenGetResponse{
		Token: "<Auth-Token>",
	}
	s.MockCluster().EXPECT().GetPairToken(req.GetResetToken()).Return(resp, nil)

	c := api.NewOpenStorageClusterPairClient(s.Conn())
	r, err := c.GetToken(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetToken())
}

func TestClusterPairServer_TokenGetFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.ClusterPairTokenGetRequest{
		ResetToken: true,
	}
	s.MockCluster().
		EXPECT().
		GetPairToken(req.GetResetToken()).
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

func TestClusterPairServer_DeleteSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	req := &api.ClusterPairDeleteRequest{
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
	req := &api.ClusterPairDeleteRequest{}
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
