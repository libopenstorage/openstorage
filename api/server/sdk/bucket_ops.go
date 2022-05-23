/*
Package sdk is the gRPC implementation of the SDK gRPC server
Copyright 2022 Portworx

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
	bucket "github.com/libopenstorage/openstorage/bucket"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BucketServer is an implementation of the gRPC OpenStorageBucket interface
type BucketServer struct {
	server serverAccessor
}

func (s *BucketServer) driver(ctx context.Context) bucket.BucketDriver {
	return s.server.bucketDriver(ctx)
}

//  Creates a new bucket
func (s *BucketServer) Create(
	ctx context.Context,
	req *api.BucketCreateRequest,
) (*api.BucketCreateResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	logrus.Info("bucket_driver.Fake delete bucket received")
	name := req.GetName()

	if len(name) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"Must supply a unique name")
	}
	// Create bucket
	id, err := s.driver(ctx).CreateBucket(name)
	if err != nil {
		return nil, err
	}

	return &api.BucketCreateResponse{
		BucketId: id,
	}, nil
}

// Deletes the bucket
func (s *BucketServer) Delete(
	ctx context.Context,
	req *api.BucketDeleteRequest,
) (*api.BucketDeleteResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	id := req.GetBucketId()
	if len(id) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a valid bucket id")
	}
	logrus.Infof("bucket_driver. delete bucket request received for %s", id)

	// Delete the bucket
	err := s.driver(ctx).DeleteBucket(id)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to delete bucket %s: %v",
			req.GetBucketId(),
			err.Error())
	}

	return &api.BucketDeleteResponse{}, nil
}
