/*
CSI Interface for OSD
Copyright 2017 Portworx

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
package csi

import (
	"testing"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestNewCSIServerGetPluginInfo(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	s.MockDriver().EXPECT().Name().Return("mock").Times(2)

	// Setup client
	c := csi.NewIdentityClient(s.Conn())

	// Get info
	r, err := c.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})
	assert.NoError(t, err)

	// Verify
	name := r.GetName()
	version := r.GetVendorVersion()
	assert.Equal(t, name, csiDriverNamePrefix+"mock")
	assert.Equal(t, version, csiDriverVersion)

	manifest := r.GetManifest()
	assert.Len(t, manifest, 1)
	assert.Equal(t, manifest["driver"], "mock")
}
