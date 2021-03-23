/*
Copyright 2021 Openstorage.org

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
