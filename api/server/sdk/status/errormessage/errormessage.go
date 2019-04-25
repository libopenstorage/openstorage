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

	"github.com/golang/protobuf/proto"
	"github.com/libopenstorage/openstorage/api"
	sdkstatus "github.com/libopenstorage/openstorage/api/server/sdk/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error maintains the initial gRPC error while appending new errors
// to the error details as it goes up the stack. This enables the gRPC to maintain
// the initial error code and condition and separating them from the errors in the
// stack as the server returns from the code.
func Error(err error, c codes.Code, msg string) error {

	// No error was given, we just return a new status
	if err == nil {
		return status.Error(c, msg)
	}

	// Check if a non-gRPC status was provided
	orig, ok := status.FromError(err)
	if !ok {
		// not a grpc style error so use the gRPC error information as the
		// actual error and add err to the details
		return ErrorSet(err, c, msg)
	}

	// Add the current gRPC error to the original
	var derr error
	orig, derr = orig.WithDetails(&api.SdkErrorMessage{
		Status: status.New(c, msg).Proto(),
	})
	if derr != nil {
		panic("Unexpected")
	}
	return orig.Err()
}

// Errorf calls Error() with a formatted string
func Errorf(err error, c codes.Code, format string, a ...interface{}) error {
	return Error(err, c, fmt.Sprintf(format, a...))
}

// ErrorSet places the current gRPC status as the actual error
// pushes all the values in err as details.
func ErrorSet(err error, c codes.Code, msg string) error {

	if err == nil {
		return status.Error(c, msg)
	}

	// Get gRPC status from error. This function also creates a new gRPC status
	// if it is a non-gRPC status error.
	orig := sdkstatus.FromError(err)

	// Create a new error
	s := status.New(c, msg)

	// Add err as the first error to maintain sequence
	origDetails := orig.Details()
	details := make([]proto.Message, 0, 1+len(origDetails))
	details = append(details, &api.SdkErrorMessage{

		// Create a new message so that the orig Details are
		// not copied over. This avoids creating long trees of
		// nested Status->Details->Status->Details->...
		//
		// Instead the details of the orig will be copied to
		// s in order laster.
		Status: status.New(orig.Code(), orig.Message()).Proto(),
	})

	// Have to use a loop to append type to the interface
	for _, detail := range origDetails {
		switch t := detail.(type) {
		case proto.Message:
			details = append(details, t)
		}
	}

	var derr error
	s, derr = s.WithDetails(details...)
	if derr != nil {
		panic("Unexpected")
	}
	return s.Err()
}

// Details enables clients or callers to extract any more
// error messages from the gRPC status details.
func Details(s *status.Status) []*api.SdkErrorMessage {
	details := s.Details()
	messages := make([]*api.SdkErrorMessage, 0, len(details))
	for _, detail := range details {
		switch t := detail.(type) {
		case *api.SdkErrorMessage:
			messages = append(messages, t)
		}
	}
	return messages
}

// Stack returns a string of the full stack of errors and details
// saved in the gRPC status.
func Stack(s *status.Status) string {
	out := fmt.Sprintln(s.Err())
	for _, detail := range Details(s) {
		out += fmt.Sprintln("  -->", detail.Error())
	}
	return out
}
