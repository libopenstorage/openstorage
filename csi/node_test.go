package csi

import (
	"testing"

	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewCSIServerGetNodeId(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	s.MockCluster().
		EXPECT().
		Enumerate().
		Return(api.Cluster{
			Status: api.Status_STATUS_OK,
			Id:     "pwx-testcluster",
			NodeId: "pwx-testnodeid",
		}, nil).
		Times(1)

	// Setup request
	req := &csi.GetNodeIDRequest{
		Version: &csi.Version{},
	}

	r, err := c.GetNodeID(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)

	// Verify
	nodeid := r.GetNodeId()
	assert.Equal(t, nodeid, "pwx-testnodeid")
}

func TestNewCSIServerGetNodeIdNoVersion(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	// Setup request
	req := &csi.GetNodeIDRequest{}

	// Expect error without version
	_, err := c.GetNodeID(context.Background(), req)

	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Version")
}

func TestNewCSIServerGetNodeIdEnumerateError(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewNodeClient(s.Conn())

	s.MockCluster().
		EXPECT().
		Enumerate().
		Return(api.Cluster{}, fmt.Errorf("TEST")).
		Times(1)

	// Setup request
	req := &csi.GetNodeIDRequest{
		Version: &csi.Version{},
	}

	// Expect error without version
	_, err := c.GetNodeID(context.Background(), req)

	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "TEST")
}
