/*
This tests in-memory fake implementation of the ObjectStore
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

package objectstore

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFakeObjectstore(t *testing.T) {

	f := NewfakeObjectstore()
	volumeId := "test-vol-id"

	// Create objectstore with given volumeID
	objInfo, err := f.ObjectStoreCreate(volumeId)
	assert.NoError(t, err)

	// Inspect
	objResp, err := f.ObjectStoreInspect(objInfo.Uuid)
	assert.NoError(t, err)
	assert.Contains(t, objResp.VolumeId, volumeId)
	assert.Equal(t, objResp.Enabled, false)

	// Update
	err = f.ObjectStoreUpdate(objInfo.Uuid, true)
	assert.NoError(t, err)

	// Inspect to check whether objectstore is enabled
	objResp, err = f.ObjectStoreInspect(objInfo.Uuid)
	assert.NoError(t, err)
	assert.Contains(t, objResp.VolumeId, volumeId)
	assert.Equal(t, objResp.Enabled, true)

	// Delete
	err = f.ObjectStoreDelete(objInfo.Uuid)
	assert.NoError(t, err)
}
