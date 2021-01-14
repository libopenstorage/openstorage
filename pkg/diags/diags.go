package diags

import (
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
	"golang.org/x/net/context"
)

// Provider is collection of APIs that implement the SDK OpenStorageDiags service
type Provider interface {
	// Collect collects diagnostics based on the provided request
	Collect(ctx context.Context, in *api.SdkDiagsCollectRequest) (*api.SdkJobResponse, error)
}

// NewDefaultDiagsProvider returns all diags provider APIs as unsupported
func NewDefaultDiagsProvider() Provider {
	return &UnsupportedDiagsProvider{}
}

// UnsupportedDiagsProvider returns unsupported implementation of diags provider
type UnsupportedDiagsProvider struct {
}

func (u *UnsupportedDiagsProvider) Collect(_ context.Context, _ *api.SdkDiagsCollectRequest) (*api.SdkJobResponse, error) {
	return nil, &errors.ErrNotSupported{}
}
