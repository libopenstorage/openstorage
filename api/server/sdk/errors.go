package sdk

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IsErrorNotFound returns if the given error is due to not found
func IsErrorNotFound(err error) bool {
	if err == nil {
		return false
	}

	s, ok := status.FromError(err)
	return ok && s.Code() == codes.NotFound
}
