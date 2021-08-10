package correlation

import (
	"context"

	"google.golang.org/grpc"
)

// ContextInterceptor represents a correlation interceptor
type ContextInterceptor struct {
	Origin Component
}

// UnaryInterceptor creates a gRPC interceptor for adding
// correlation ID to each request
func (ci *ContextInterceptor) ContextUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	ctx = NewContext(ctx, ci.Origin)

	return handler(ctx, req)
}
