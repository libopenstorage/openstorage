package sdk

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
	"github.com/libopenstorage/openstorage/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DiagsServer is an implementation of the OpenStorageDiags SDK
type DiagsServer struct {
	server serverAccessor
}

func (s *DiagsServer) cluster() cluster.Cluster {
	return s.server.cluster()
}

func (s *DiagsServer) Collect(ctx context.Context, in *api.SdkDiagsCollectRequest) (*api.SdkJobResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return s.cluster().Collect(ctx, in)
}
