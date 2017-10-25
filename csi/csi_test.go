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
	"net"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/libopenstorage/openstorage/tests"
)

func TestNewCSIServerListener(t *testing.T) {

	// Listen on port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	tests.Assert(t, err == nil)

	// Setup simple driver
	s := NewOsdCsiServer(&OsdCsiServerConfig{
		Listener: l,
	})
	err = s.Start()
	tests.Assert(t, err == nil)
	defer s.Stop()

	// Setup a connection to the driver
	conn, err := grpc.Dial(s.Address(), grpc.WithInsecure())
	tests.Assert(t, err == nil)
	defer conn.Close()

	// Make a call
	c := csi.NewIdentityClient(conn)
	r, err := c.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})
	tests.Assert(t, err == nil)

	// Verify
	name := r.GetResult().GetName()
	version := r.GetResult().GetVendorVersion()
	tests.Assert(t, name == csiDriverName)
	tests.Assert(t, version == csiDriverVersion)
}
