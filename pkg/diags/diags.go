package diags

import (
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
	"golang.org/x/net/context"
)

// Provider is collection of APIs that implement the SDK OpenStorageDiags service
type Provider interface {
	api.OpenStorageDiagsServer
}

// NewDefaultDiagsProvider returns all diags provider APIs as unsupported
func NewDefaultDiagsProvider() Provider {
	return &UnsupportedDiagsProvider{}
}

// UnsupportedDiagsProvider returns unsupported implementation of diags provider
type UnsupportedDiagsProvider struct {
}

func (u *UnsupportedDiagsProvider) Collect(_ context.Context, _ *api.SdkDiagsCollectRequest) (*api.SdkDiagsCollectResponse, error) {
	return nil, &errors.ErrNotSupported{}
}
