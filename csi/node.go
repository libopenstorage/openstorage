package csi

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.pedge.io/dlog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Get PX Node id for local node.
func (s *OsdCsiServer) GetNodeID(ctx context.Context, req *csi.GetNodeIDRequest) (*csi.GetNodeIDResponse, error) {
	dlog.Debugf("GetNodeID req[%#v]", req)

	// Check arguments
	if req.GetVersion() == nil {
		return nil, status.Error(codes.InvalidArgument, "Version must be provided")
	}

	clus, err := s.cluster.Enumerate()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to Enumerate cluster: %s", err)
	}

	result := &csi.GetNodeIDResponse{
		NodeId: clus.NodeId,
	}

	dlog.Infof("NodeId is %s", result.NodeId)

	return result, nil
}
