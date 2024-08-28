/*
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
package manager

import (
	"fmt"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/pkg/auth/systemtoken"
	"github.com/libopenstorage/openstorage/pkg/dbg"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/stretchr/testify/assert"
)

const (
	testClusterId   = "test-cluster-id"
	testClusterUuid = "test-cluster-uuid"
)

var (
	kv kvdb.Kvdb
)

func init() {
	var err error
	kv, err = kvdb.New(mem.Name, "manager_test/"+testClusterId, []string{}, nil, kvdb.LogFatalErrorCB)
	if err != nil {
		dbg.LogErrorAndPanicf(fmt.Errorf("failed to initialize KVDB"), "init kvdb")
	}
	if err := kvdb.SetInstance(kv); err != nil {
		dbg.LogErrorAndPanicf(fmt.Errorf("failed to set KVDB instance"), "set kvdb instance")
	}
}

func cleanup() {
	inst.Shutdown()
	time.Sleep(100 * time.Millisecond)
	kvdb.Instance().Delete(ClusterDBKey)
	inst = nil
}

func TestClusterManagerUuid(t *testing.T) {
	oldInst := inst
	defer func() {
		inst = oldInst
	}()

	uuid := "uuid"
	id := "id"
	err := Init(config.ClusterConfig{
		ClusterId:   id,
		ClusterUuid: uuid,
	})
	assert.NoError(t, err)
	assert.Equal(t, uuid, inst.Uuid())
	cleanup()
}

func TestUpdateSchedulerNodeName(t *testing.T) {

	nodeID := "node-alpha"
	Init(config.ClusterConfig{
		ClusterId:         testClusterId,
		ClusterUuid:       testClusterUuid,
		NodeId:            nodeID,
		SchedulerNodeName: "old-sched-name",
	})
	manager, err := systemtoken.NewManager(&systemtoken.Config{
		ClusterId:    testClusterId,
		NodeId:       nodeID,
		SharedSecret: "mysecret",
	})
	assert.NoError(t, err)
	auth.InitSystemTokenManager(manager)

	err = inst.StartWithConfiguration(false, "1001", []string{}, "", &cluster.ClusterServerConfiguration{
		ConfigSystemTokenManager: manager,
	}, "gobRegisterName/path")
	assert.NoError(t, err)

	node, err := inst.Inspect(nodeID)
	assert.NoError(t, err)
	assert.Equal(t, "old-sched-name", node.SchedulerNodeName)
	assert.Equal(t, node.GossipPort, "1001", "Expected gossip port to be updated in cluster database")

	err = inst.UpdateSchedulerNodeName("new-sched-name")
	assert.NoError(t, err)

	node, err = inst.Inspect(nodeID)
	assert.NoError(t, err)
	assert.Equal(t, "new-sched-name", node.SchedulerNodeName)

	tokenResp, err := inst.GetPairToken(false)
	assert.NoError(t, err)
	assert.True(t, auth.IsJwtToken(tokenResp.Token))

	cleanup()
}
