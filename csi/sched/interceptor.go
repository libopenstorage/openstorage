package sched

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// FilterInterceptor is a wrapper for the filter
// to be used an interceptor
type FilterInterceptor struct {
	Filter
}

// SchedUnaryInterceptor calls the filter function based on the req
func (fi *FilterInterceptor) SchedUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var err error

	switch req.(type) {
	case *csi.CreateVolumeRequest:
		csiReq := req.(*csi.CreateVolumeRequest)
		req, err = fi.Filter.PreVolumeCreate(csiReq)
		if err != nil {
			logrus.WithContext(ctx).Warnf("CSI pre-create filter failed: %v", err)

			// Log warning and continue with handler. This is backwards 
			// compatible with the previous behavior of continuing create.
			return handler(ctx, req)
		} else {
			logrus.WithContext(ctx).Tracef("K8s-CSI filter: Filter applied successfully for request %T", req)
		}
	default:
		logrus.WithContext(ctx).Tracef("K8s-CSI filter: Ignoring filter for this request: %T", req)
	}

	return handler(ctx, req)
}
