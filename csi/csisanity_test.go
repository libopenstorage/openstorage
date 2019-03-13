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
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/server/sdk"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/pkg/role"
	"github.com/libopenstorage/openstorage/pkg/storagepolicy"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/kubernetes-csi/csi-test/pkg/sanity"
)

func TestCSISanity(t *testing.T) {

	clustermanager.Init(config.ClusterConfig{
		ClusterId: "fakecluster",
		NodeId:    "fakeNode",
	})
	cm, err := clustermanager.Inst()
	go func() {
		cm.Start(0, false, "9002")
	}()
	defer cm.Shutdown()

	// Start CSI Server
	server, err := NewOsdCsiServer(&OsdCsiServerConfig{
		DriverName: "fake",
		Net:        "tcp",
		Address:    "127.0.0.1:0",
		Cluster:    cm,
		SdkUds:     testSdkSock,
	})
	if err != nil {
		t.Fatalf("Unable to start csi server: %v", err)
	}
	server.Start()
	defer server.Stop()

	// Setup sdk server
	kv, err := kvdb.New(mem.Name, "test", []string{}, nil, logrus.Panicf)
	assert.NoError(t, err)
	stp, err := storagepolicy.Init(kv)
	if err != nil {
		stp, _ = storagepolicy.Inst()
	}
	assert.NotNil(t, stp)
	rm, err := role.NewSdkRoleManager(kv)
	assert.NoError(t, err)

	os.Remove(testSdkSock)
	selfsignedJwt, err := auth.NewJwtAuth(&auth.JwtAuthConfig{
		SharedSecret:  []byte(testSharedSecret),
		UsernameClaim: auth.UsernameClaimTypeName,
	})

	_ = rm
	_ = selfsignedJwt

	// setup sdk server
	sdk, err := sdk.New(&sdk.ServerConfig{
		DriverName:    "fake",
		Net:           "tcp",
		Address:       ":8123",
		RestPort:      "8124",
		Cluster:       cm,
		Socket:        testSdkSock,
		StoragePolicy: stp,
		AccessOutput:  ioutil.Discard,
		AuditOutput:   ioutil.Discard,
		// Auth disabled for now.
		// We're only sanity testing Client -> CSI -> SDK (No Auth)
		/*Security: &sdk.SecurityConfig{
			Role: rm,
			Authenticators: map[string]auth.Authenticator{
				"openstorage.io": selfsignedJwt,
			},
		},*/
	})
	assert.Nil(t, err)

	err = sdk.Start()
	assert.Nil(t, err)
	defer sdk.Stop()

	timeout := time.After(30 * time.Second)
	for {
		select {
		case <-timeout:
			t.Fatal("Timemout waiting for cluster to be ready")
		default:
		}
		cl, err := cm.Enumerate()
		if err != nil {
			t.Fatal("Unable to get cluster status")
		}
		if cl.Status == api.Status_STATUS_OK {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Start CSI Sanity test
	sanity.Test(t, &sanity.Config{
		Address:    server.Address(),
		TargetPath: "/mnt",
	})
}
