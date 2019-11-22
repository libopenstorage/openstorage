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

	"github.com/golang/mock/gomock"
	"github.com/libopenstorage/openstorage/api"
	mockapi "github.com/libopenstorage/openstorage/api/mock"
	"github.com/libopenstorage/openstorage/api/server/sdk"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/pkg/role"
	"github.com/libopenstorage/openstorage/pkg/storagepolicy"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/stretchr/testify/assert"

	"github.com/kubernetes-csi/csi-test/pkg/sanity"
	"github.com/kubernetes-csi/csi-test/utils"
)

func TestCSISanity(t *testing.T) {
	tester := &testServer{}
	tester.setPorts()
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})
	tester.s = mockapi.NewMockOpenStoragePoolServer(tester.mc)

	clustermanager.Init(config.ClusterConfig{
		ClusterId: "fakecluster",
		NodeId:    "fakeNode",
	})
	cm, err := clustermanager.Inst()
	go func() {
		cm.Start(false, "9002", "")
	}()
	defer cm.Shutdown()

	// Setup sdk server
	kv, err := kvdb.New(mem.Name, "test", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)
	kvdb.SetInstance(kv)
	stp, err := storagepolicy.Init()
	if err != nil {
		stp, _ = storagepolicy.Inst()
	}
	assert.NotNil(t, stp)
	rm, err := role.NewSdkRoleManager(kv)
	assert.NoError(t, err)

	selfsignedJwt, err := auth.NewJwtAuth(&auth.JwtAuthConfig{
		SharedSecret:  []byte(testSharedSecret),
		UsernameClaim: auth.UsernameClaimTypeName,
	})

	_ = rm
	_ = selfsignedJwt

	// setup sdk server
	sdk, err := sdk.New(&sdk.ServerConfig{
		DriverName:        "fake",
		Net:               "tcp",
		Address:           ":" + tester.port,
		RestPort:          tester.gwport,
		Cluster:           cm,
		Socket:            tester.uds,
		StoragePolicy:     stp,
		StoragePoolServer: tester.s,
		AccessOutput:      ioutil.Discard,
		AuditOutput:       ioutil.Discard,
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

	// Start CSI Server
	server, err := NewOsdCsiServer(&OsdCsiServerConfig{
		DriverName: "fake",
		Net:        "tcp",
		Address:    "127.0.0.1:0",
		Cluster:    cm,
		SdkUds:     tester.uds,
	})
	if err != nil {
		t.Fatalf("Unable to start csi server: %v", err)
	}
	server.Start()
	defer server.Stop()

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
	targetPath := "/tmp/mnt/csi"
	sanity.Test(t, &sanity.Config{
		Address:    server.Address(),
		TargetPath: targetPath,
		CreateTargetDir: func(p string) (string, error) {
			os.MkdirAll(p+"/target", os.FileMode(0755))
			return p, nil
		},
	})
}
