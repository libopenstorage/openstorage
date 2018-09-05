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
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"

	"github.com/libopenstorage/openstorage/config"
)

func init() {
	kv, err := kvdb.New(mem.Name, "manager_test", []string{}, nil, logrus.Panicf)
	if err != nil {
		logrus.Panicf("Failed to initialize KVDB")
	}
	if err := kvdb.SetInstance(kv); err != nil {
		logrus.Panicf("Failed to set KVDB instance")
	}
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
}
