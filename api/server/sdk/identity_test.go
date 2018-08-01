/*
Package sdk is the gRPC implementation of the SDK gRPC server
Copyright 2018 Portworx

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
package sdk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/libopenstorage/openstorage/api"
)

func expectCapability(
	t *testing.T,
	expected api.SdkServiceCapability_OpenStorageService_Type,
	capabilities []*api.SdkServiceCapability,
) {

	for _, capOneOf := range capabilities {
		cap := capOneOf.GetService().GetType()
		if cap == expected {
			return
		}
	}

	t.Errorf("Capability %s not found in %+v", expected, capabilities)
}

func TestIdentityCapabilities(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	c := api.NewOpenStorageIdentityClient(s.Conn())

	// Get identities
	r, err := c.Capabilities(context.Background(), &api.SdkIdentityCapabilitiesRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, r)

	expectedCapabilities := []api.SdkServiceCapability_OpenStorageService_Type{
		api.SdkServiceCapability_OpenStorageService_CLUSTER,
		api.SdkServiceCapability_OpenStorageService_CLOUD_BACKUP,
		api.SdkServiceCapability_OpenStorageService_CREDENTIALS,
		api.SdkServiceCapability_OpenStorageService_NODE,
		api.SdkServiceCapability_OpenStorageService_OBJECT_STORAGE,
		api.SdkServiceCapability_OpenStorageService_SCHEDULE_POLICY,
		api.SdkServiceCapability_OpenStorageService_VOLUME,
	}

	for _, cap := range expectedCapabilities {
		expectCapability(t, cap, r.GetCapabilities())
	}
}
