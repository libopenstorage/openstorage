package sched

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
			logrus.WithContext(ctx).Errorf("CSI pre-create filter failed: %v", err)

			// Return an aborted code to retry from the csi-provisioner.
			// We cannot ignore this error or else a volume will be created w/
			// incorrect locator.VolumeLabels.
			return nil, status.Errorf(codes.Aborted, "pre-create filter failed: %v", err)
		} else {
			logrus.WithContext(ctx).Tracef("K8s-CSI filter: Filter applied successfully for request %T", req)
		}
	default:
		logrus.WithContext(ctx).Tracef("K8s-CSI filter: Ignoring filter for this request: %T", req)
	}

	return handler(ctx, req)
}
