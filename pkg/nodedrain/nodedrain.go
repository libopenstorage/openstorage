package nodedrain

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
)

// Provider is a collection of APIs for performing different kinds of drain
// operations on a node
type Provider interface {
	// DrainAttachments creates a task to drain volume attachments
	// from the provided node in the cluster.
	DrainAttachments(ctx context.Context, in *api.SdkNodeDrainAttachmentsRequest) (*api.SdkJobResponse, error)
	// CordonAttachments disables any new volume attachments
	// from the provided node in the cluster. Existing volume attachments
	// will stay on the node.
	CordonAttachments(ctx context.Context, in *api.SdkNodeCordonAttachmentsRequest) (*api.SdkNodeCordonAttachmentsResponse, error)
	// UncordonAttachments re-enables volume attachments
	// on the provided node in the cluster.
	UncordonAttachments(ctx context.Context, in *api.SdkNodeUncordonAttachmentsRequest) (*api.SdkNodeUncordonAttachmentsResponse, error)
}

// NewDefaultNodeDrainProvider does not any node drain related operations
func NewDefaultNodeDrainProvider() Provider {
	return &UnsupportedNodeDrainProvider{}
}

// UnsupportedNodeDrainProvider unsupported implementation of drain.
type UnsupportedNodeDrainProvider struct {
}

func (u *UnsupportedNodeDrainProvider) DrainAttachments(
	ctx context.Context,
	in *api.SdkNodeDrainAttachmentsRequest,
) (*api.SdkJobResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (u *UnsupportedNodeDrainProvider) CordonAttachments(
	ctx context.Context,
	in *api.SdkNodeCordonAttachmentsRequest,
) (*api.SdkNodeCordonAttachmentsResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (u *UnsupportedNodeDrainProvider) UncordonAttachments(
	ctx context.Context,
	in *api.SdkNodeUncordonAttachmentsRequest,
) (*api.SdkNodeUncordonAttachmentsResponse, error) {
	return nil, &errors.ErrNotSupported{}
}
