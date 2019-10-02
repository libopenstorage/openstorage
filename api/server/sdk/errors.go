/*
Package sdk is the gRPC implementation of the SDK gRPC server
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
package sdk

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IsErrorNotFound returns if the given error is due to not found
func IsErrorNotFound(err error) bool {
	return FromError(err).Code() == codes.NotFound
}

func FromError(err error) *status.Status {
	// From github.com/grpc-ecosystem/grpc-gateway/runtime/errors.go
	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}

	return s
}

func HTTPStatusFromSdkError(err error) int {
	return runtime.HTTPStatusFromCode(FromError(err).Code())
}
