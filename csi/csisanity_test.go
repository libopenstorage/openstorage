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
	"time"

	"github.com/libopenstorage/openstorage/api"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/volume/drivers"

	"github.com/kubernetes-csi/csi-test/pkg/sanity"
	"github.com/sirupsen/logrus"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
)

func TestCSISanity(t *testing.T) {
	t.Skip("Flaky")

	kv, err := kvdb.New(mem.Name, "fake_test", []string{}, nil, logrus.Panicf)
	if err != nil {
		logrus.Panicf("Failed to initialize KVDB")
	}
	if err := kvdb.SetInstance(kv); err != nil {
		logrus.Panicf("Failed to set KVDB instance")
	}
	clustermanager.Init(config.ClusterConfig{
		ClusterId: "fakecluster",
		NodeId:    "fakeNode",
	})
	cm, err := clustermanager.Inst()
	go func() {
		cm.Start(0, false, "9002")
	}()
	if err := volumedrivers.Register("fake", map[string]string{}); err != nil {
		t.Fatalf("Unable to start volume driver fake: %v", err)
	}

	// Start CSI Server
	server, err := NewOsdCsiServer(&OsdCsiServerConfig{
		DriverName: "fake",
		Net:        "tcp",
		Address:    "127.0.0.1:0",
		Cluster:    cm,
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
	sanity.Test(t, &sanity.Config{
		Address:    server.Address(),
		TargetPath: "/mnt",
	})
}
