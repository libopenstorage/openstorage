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
	"encoding/json"

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
	name := req.GetName()
	if len(name) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"Must supply a unique name")
	}
	region := req.GetRegion()
	if len(region) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"Must supply the region")
	}
	logrus.Infof("Create bucket request received for Bucket: %s, Region: %s", name, region)

	// Create bucket
	id, err := s.driver(ctx).CreateBucket(name, region, req.GetEndpoint(), req.GetAnonymousBucketAccessMode())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to create the Bucket: %v", err)
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
	region := req.GetRegion()
	if len(region) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"Must supply the region")
	}
	logrus.Infof("Delete bucket request received for Bucket: %s", id)

	// Delete the bucket
	err := s.driver(ctx).DeleteBucket(id, region, req.GetEndpoint(), req.GetClearBucket())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to delete bucket %s: %v",
			req.GetBucketId(),
			err.Error())
	}

	return &api.BucketDeleteResponse{}, nil
}

func isJSON(input string) bool {
	var js interface{}
	return json.Unmarshal([]byte(input), &js) == nil
}

func (s *BucketServer) GrantAccess(
	ctx context.Context,
	req *api.BucketGrantAccessRequest,
) (*api.BucketGrantAccessResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	id := req.GetBucketId()
	if len(id) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a valid bucket id")
	}

	accountName := req.GetAccountName()
	if len(accountName) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a valid account name")
	}

	accessPolicy := req.GetAccessPolicy()
	if len(accessPolicy) != 0 && !isJSON(accessPolicy) {
		return nil, status.Error(
			codes.InvalidArgument,
			"Supply a valid access policy or leave it empty to allow account complete access to the bucket")
	}

	// Grant Bucket Access
	id, bucketCredentials, err := s.driver(ctx).GrantBucketAccess(id, accountName, accessPolicy)
	if err != nil {
		return nil, err
	}

	return &api.BucketGrantAccessResponse{
		AccountId: id,
		Credentials: &api.BucketAccessCredentials{
			AccessKeyId:     bucketCredentials.AccessKeyId,
			SecretAccessKey: bucketCredentials.SecretAccessKey,
		},
	}, nil
}

func (s *BucketServer) RevokeAccess(
	ctx context.Context,
	req *api.BucketRevokeAccessRequest,
) (*api.BucketRevokeAccessResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	id := req.GetBucketId()
	if len(id) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a valid bucket id")
	}

	accountId := req.GetAccountId()
	if len(accountId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a valid account id")
	}

	// Revoke Bucket Access
	err := s.driver(ctx).RevokeBucketAccess(id, accountId)
	if err != nil {
		return nil, err
	}

	return &api.BucketRevokeAccessResponse{}, nil
}
