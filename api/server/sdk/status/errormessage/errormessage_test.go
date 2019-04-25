/*
Copyright 2019 Portworx

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

package errormessage

import (
	"fmt"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStatusWithErrorFromNoErrror(t *testing.T) {
	// Brand new gRPC/SDK Error
	err := Error(nil, codes.PermissionDenied, "Denied")
	assert.Error(t, err)

	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.NotNil(t, s)
	assert.Len(t, s.Details(), 0)
	assert.Len(t, Details(s), 0)
	assert.Equal(t, s.Code(), codes.PermissionDenied)
	assert.Equal(t, s.Message(), "Denied")

	// Append new error to error
	err = Error(err, codes.InvalidArgument, "Invalid")
	assert.Error(t, err)

	s, ok = status.FromError(err)
	assert.True(t, ok)
	assert.NotNil(t, s)
	assert.Len(t, s.Details(), 1)
	assert.Len(t, Details(s), 1)
	assert.Equal(t, s.Code(), codes.PermissionDenied)
	assert.Equal(t, s.Message(), "Denied")
	assert.Contains(t, Details(s), &api.SdkErrorMessage{
		Status: status.New(codes.InvalidArgument, "Invalid").Proto(),
	})

	details := Details(s)
	assert.Equal(t, codes.InvalidArgument, details[0].GetGrpcStatus().Code())
	assert.Equal(t, "Invalid", details[0].GetGrpcStatus().Message())

	// Assert that a new value can be placed as the actual error while
	// maintaining all the other errors in the details
	err = ErrorSet(err, codes.DataLoss, "DATALOSS")
	assert.Error(t, err)

	s, ok = status.FromError(err)
	assert.True(t, ok)
	assert.NotNil(t, s)
	assert.Len(t, s.Details(), 2)
	assert.Len(t, Details(s), 2)
	assert.Equal(t, s.Code(), codes.DataLoss)
	assert.Equal(t, s.Message(), "DATALOSS")
	assert.Contains(t, Details(s), &api.SdkErrorMessage{
		Status: status.New(codes.PermissionDenied, "Denied").Proto(),
	})
	assert.Contains(t, Details(s), &api.SdkErrorMessage{
		Status: status.New(codes.InvalidArgument, "Invalid").Proto(),
	})
	fmt.Println(Stack(s))

}

func TestStatusWithErrorFromGolangError(t *testing.T) {
	// Brand new gRPC/SDK Error
	err := Error(fmt.Errorf("ERROR"), codes.PermissionDenied, "Denied")
	assert.NotNil(t, err)

	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.NotNil(t, s)
	assert.Len(t, s.Details(), 1)
	assert.Len(t, Details(s), 1)
	assert.Equal(t, s.Code(), codes.PermissionDenied)
	assert.Equal(t, s.Message(), "Denied")
	assert.Contains(t, Details(s), &api.SdkErrorMessage{
		Status: status.New(codes.Unknown, "ERROR").Proto(),
	})

	// Append new error to error
	err = Error(err, codes.InvalidArgument, "Invalid")
	assert.NotNil(t, s)

	s, ok = status.FromError(err)
	assert.True(t, ok)
	assert.NotNil(t, s)
	assert.Len(t, s.Details(), 2)
	assert.Len(t, Details(s), 2)
	assert.Equal(t, s.Code(), codes.PermissionDenied)
	assert.Equal(t, s.Message(), "Denied")
	assert.Contains(t, Details(s), &api.SdkErrorMessage{
		Status: status.New(codes.Unknown, "ERROR").Proto(),
	})
	assert.Contains(t, Details(s), &api.SdkErrorMessage{
		Status: status.New(codes.InvalidArgument, "Invalid").Proto(),
	})
	fmt.Println(Stack(s))
}
