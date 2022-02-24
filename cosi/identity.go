package cosi

import (
	"context"

	cosi "sigs.k8s.io/container-object-storage-interface-spec"
)

// ProvisionerGetInfo returns any required provisioner info
func (s *Server) ProvisionerGetInfo(ctx context.Context, req *cosi.ProvisionerGetInfoRequest) (*cosi.ProvisionerGetInfoResponse, error) {
	return &cosi.ProvisionerGetInfoResponse{
		Name: "osd.openstorage.org",
	}, nil
}
