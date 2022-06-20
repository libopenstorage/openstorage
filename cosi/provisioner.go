package cosi

import (
	"context"
	"fmt"

	"github.com/libopenstorage/openstorage/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"
)

// ProvisionerCreateBucket is made to create the bucket in the backend.
// This call is idempotent
//    1. If a bucket that matches both name and parameters already exists, then OK (success) must be returned.
//    2. If a bucket by same name, but different parameters is provided, then the appropriate error code ALREADY_EXISTS must be returned.
func (s *Server) ProvisionerCreateBucket(ctx context.Context, req *cosi.ProvisionerCreateBucketRequest) (*cosi.ProvisionerCreateBucketResponse, error) {
	logrus.Info("cosi.ProvisionerCreateBucket received")
	id, err := s.driver.CreateBucket(req.GetName(), req.GetParameters()["region"], api.AnonymousBucketAccessMode_Private)
	if err != nil {
		return &cosi.ProvisionerCreateBucketResponse{}, status.Error(codes.Internal, fmt.Sprintf("failed to create bucket: %s", err))
	}

	return &cosi.ProvisionerCreateBucketResponse{
		BucketId: id,
	}, nil
}

// ProvisionerDeleteBucket is made to delete the bucket in the backend.
// If the bucket has already been deleted, then no error should be returned.
func (s *Server) ProvisionerDeleteBucket(ctx context.Context, req *cosi.ProvisionerDeleteBucketRequest) (*cosi.ProvisionerDeleteBucketResponse, error) {
	logrus.Info("cosi.ProvisionerDeleteBucket received")
	// Passing clearBucket as true as it is not possible to delete a bucket with objects available.
	// This value is to be made configurable.
	// Region information has to be saved in Bucket object and passed here
	if err := s.driver.DeleteBucket(req.GetBucketId(), "region", true); err != nil {
		return &cosi.ProvisionerDeleteBucketResponse{}, status.Error(codes.Internal, fmt.Sprintf("failed to delete bucket: %s", err))
	}

	return &cosi.ProvisionerDeleteBucketResponse{}, nil
}

// ProvisionerGrantBucketAccess grants access to an account. The account_name in the request shall be used as a unique identifier to create credentials.
// The account_id returned in the response will be used as the unique identifier for deleting this access when calling ProvisionerRevokeBucketAccess.
func (s *Server) ProvisionerGrantBucketAccess(context.Context, *cosi.ProvisionerGrantBucketAccessRequest) (*cosi.ProvisionerGrantBucketAccessResponse, error) {
	logrus.Info("cosi.ProvisionerGrantBucketAccessResponse received")
	return &cosi.ProvisionerGrantBucketAccessResponse{}, nil
}

// ProvisionerRevokeBucketAccess revokes all access to a particular bucket from a principal.
func (s *Server) ProvisionerRevokeBucketAccess(context.Context, *cosi.ProvisionerRevokeBucketAccessRequest) (*cosi.ProvisionerRevokeBucketAccessResponse, error) {
	logrus.Info("cosi.ProvisionerRevokeBucketAccessResponse received")
	return &cosi.ProvisionerRevokeBucketAccessResponse{}, nil
}
