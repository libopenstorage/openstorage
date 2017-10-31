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
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestNewCSIServerGetPluginInfo(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	s.MockDriver().EXPECT().Name().Return("mock").Times(1)

	// Make a call
	c := csi.NewIdentityClient(s.Conn())
	r, err := c.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})
	assert.Nil(t, err)

	// Verify
	name := r.GetResult().GetName()
	version := r.GetResult().GetVendorVersion()
	assert.Equal(t, name, csiDriverName)
	assert.Equal(t, version, csiDriverVersion)
	manifest := r.GetResult().GetManifest()
	assert.Len(t, manifest, 1)
	assert.Equal(t, manifest["driver"], "mock")
}

func TestNewCSIServerGetSupportedVersions(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewIdentityClient(s.Conn())
	r, err := c.GetSupportedVersions(context.Background(), &csi.GetSupportedVersionsRequest{})
	assert.Nil(t, err)

	// Verify
	versions := r.GetResult().GetSupportedVersions()
	assert.Equal(t, len(versions), 1)
	assert.True(t, reflect.DeepEqual(versions[0], csiVersion))
}
