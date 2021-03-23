/*
Copyright 2021 Openstorage.org

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

func (s *DiagsServer) Collect(ctx context.Context, in *api.SdkDiagsCollectRequest) (*api.SdkDiagsCollectResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return s.cluster().Collect(ctx, in)
}
