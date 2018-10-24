/*
Package fake provides an in-memory fake driver implementation
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
package fake

import (
	"fmt"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	"github.com/libopenstorage/openstorage/config"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func init() {
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
}

func TestFakeName(t *testing.T) {
	d, err := Init(map[string]string{})
	assert.NoError(t, err)
	assert.Equal(t, Name, d.Name())
}

func TestFakeCredentials(t *testing.T) {
	d, err := Init(map[string]string{})
	assert.NoError(t, err)

	id, err := d.CredsCreate(map[string]string{
		"hello": "world",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	creds, err := d.CredsEnumerate()
	assert.NoError(t, err)
	assert.NotEmpty(t, creds)
	assert.Len(t, creds, 1)

	data := creds[id]
	value, ok := data.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, value)
	assert.Equal(t, value["hello"].(string), "world")

	err = d.CredsDelete(id)
	assert.NoError(t, err)

	creds, err = d.CredsEnumerate()
	assert.NoError(t, err)
	assert.Empty(t, creds)
}

func TestFakeCreateVolume(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	vid, err := d.Create(&api.VolumeLocator{
		Name: "myvol",
	}, &api.Source{}, &api.VolumeSpec{
		Size: 1234,
	})
	assert.Error(t, err)
	assert.Empty(t, vid)

	vid, err = d.Create(&api.VolumeLocator{
		Name: "myvol",
	}, &api.Source{}, &api.VolumeSpec{
		Size:    1234,
		HaLevel: 1,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, vid)

	vols, err := d.Inspect([]string{vid})
	assert.NoError(t, err)
	assert.NotNil(t, vols)
	assert.Len(t, vols, 1)
	vol := vols[0]
	assert.Equal(t, "myvol", vol.GetLocator().GetName())
	assert.Equal(t, uint64(1234), vol.GetSpec().GetSize())
}

func TestFakeInspect(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	v, err := d.Inspect([]string{"asdf"})
	assert.NotNil(t, err)
	assert.Error(t, err)
	assert.Equal(t, err, kvdb.ErrNotFound)
	assert.Nil(t, v)
}

func TestFakeCloudBackupCreate(t *testing.T) {
	d, err := Init(map[string]string{})
	assert.NoError(t, err)

	// No vol id or cred id
	req := &api.CloudBackupCreateRequest{
		VolumeID:       "abc",
		CredentialUUID: "def",
	}
	r, err := d.CloudBackupCreate(req)
	assert.Error(t, err)
	assert.Nil(t, r)

	// Create a vol
	name := "myvol"
	size := uint64(1234)
	volid, err := d.Create(&api.VolumeLocator{Name: name}, &api.Source{}, &api.VolumeSpec{
		Size:    size,
		HaLevel: 1,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, volid)
	req.VolumeID = volid

	// Fail because no cred id
	r, err = d.CloudBackupCreate(req)
	assert.Error(t, err)
	assert.Nil(t, r)
	// Create cred
	credid, err := d.CredsCreate(map[string]string{
		"hello": "world",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, credid)
	req.CredentialUUID = credid

	// Success
	r, err = d.CloudBackupCreate(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, r.Name)
}

func testInitForCloudBackups(t *testing.T, d *driver) (string, string, *api.CloudBackupCreateRequest, *api.Volume) {
	// Create a vol
	name := "myvol"
	size := uint64(1234)
	volid, err := d.Create(&api.VolumeLocator{Name: name}, &api.Source{}, &api.VolumeSpec{
		Size:    size,
		HaLevel: 1,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, volid)

	// Create cred
	credid, err := d.CredsCreate(map[string]string{
		"hello": "world",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, credid)

	req := &api.CloudBackupCreateRequest{
		VolumeID:       volid,
		CredentialUUID: credid,
	}

	name, id, err := d.cloudBackupCreate(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.NotEmpty(t, name)

	origvols, err := d.Inspect([]string{volid})
	assert.NoError(t, err)
	assert.Len(t, origvols, 1)
	origvol := origvols[0]

	return name, id, req, origvol
}

func TestFakeCloudBackupRestore(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	_, backupId, createReq, origvol := testInitForCloudBackups(t, d)
	resp, err := d.CloudBackupRestore(&api.CloudBackupRestoreRequest{
		CredentialUUID:    createReq.CredentialUUID,
		ID:                backupId,
		RestoreVolumeName: "abc",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.RestoreVolumeID)

	vols, err := d.Inspect([]string{resp.RestoreVolumeID})
	assert.NoError(t, err)
	assert.Len(t, vols, 1)
	vol := vols[0]

	assert.Equal(t, vol.GetLocator().GetName(), "abc")
	assert.Equal(t, vol.GetSpec().GetSize(), origvol.GetSpec().GetSize())
}

func TestFakeCloudBackupDelete(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	name, backupId, createReq, _ := testInitForCloudBackups(t, d)
	_, err = d.kv.Get(backupsKeyPrefix + "/" + name)
	assert.NoError(t, err)

	err = d.CloudBackupDelete(&api.CloudBackupDeleteRequest{
		ID:             backupId,
		CredentialUUID: createReq.CredentialUUID,
	})
	assert.NoError(t, err)

	_, err = d.kv.Get(backupsKeyPrefix + "/" + backupId)
	assert.Error(t, err)
}

func TestFakeCloudBackupEnumerateWithoutMatches(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	numbackups := 50
	names := make([]string, 0)
	id := ""
	var credBackupReq *api.CloudBackupCreateRequest
	for i := 0; i < numbackups; i++ {
		id, _, credBackupReq, _ = testInitForCloudBackups(t, d)
		assert.NoError(t, err)
		names = append(names, id)
	}

	resp, err := d.CloudBackupEnumerate(&api.CloudBackupEnumerateRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			CredentialUUID: credBackupReq.CredentialUUID,
		},
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Backups, numbackups)
	for _, id := range names {
		_, err := d.kv.Delete(backupsKeyPrefix + "/" + id)
		assert.NoError(t, err)
	}

}

func TestFakeCloudBackupEnumerateMatchingVolumes(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	numbackups := 50
	var credBackupReq *api.CloudBackupCreateRequest
	var vol *api.Volume
	names := make([]string, 0)
	id := ""
	for i := 0; i < numbackups; i++ {
		id, _, credBackupReq, vol = testInitForCloudBackups(t, d)
		assert.NoError(t, err)
		names = append(names, id)
	}

	resp, err := d.CloudBackupEnumerate(&api.CloudBackupEnumerateRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			CredentialUUID: credBackupReq.CredentialUUID,
			SrcVolumeID:    vol.GetId(),
		},
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Backups, 1)
	for _, id := range names {
		_, err := d.kv.Delete(backupsKeyPrefix + "/" + id)
		assert.NoError(t, err)
	}
	resp, err = d.CloudBackupEnumerate(&api.CloudBackupEnumerateRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			CredentialUUID: credBackupReq.CredentialUUID,
		},
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Backups, 0)

}

func TestFakeCloudBackupDeleteAllWithoutMatches(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	numbackups := 50
	names := make([]string, 0)
	id := ""
	var credBackupReq *api.CloudBackupCreateRequest
	for i := 0; i < numbackups; i++ {
		id, _, credBackupReq, _ = testInitForCloudBackups(t, d)
		assert.NoError(t, err)
		names = append(names, id)
	}

	resp, err := d.CloudBackupEnumerate(&api.CloudBackupEnumerateRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			CredentialUUID: credBackupReq.CredentialUUID,
		},
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Backups, numbackups)

	// Now delete all
	err = d.CloudBackupDeleteAll(&api.CloudBackupDeleteAllRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			CredentialUUID: credBackupReq.CredentialUUID,
		},
	})
	assert.NoError(t, err)

	resp, err = d.CloudBackupEnumerate(&api.CloudBackupEnumerateRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			CredentialUUID: credBackupReq.CredentialUUID,
		},
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Backups, 0)
}

func TestFakeCloudBackupDeleteAllVolumeIdMatch(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	numbackups := 50
	var credBackupReq *api.CloudBackupCreateRequest
	var vol *api.Volume
	for i := 0; i < numbackups; i++ {
		_, _, credBackupReq, vol = testInitForCloudBackups(t, d)
		assert.NoError(t, err)
	}

	resp, err := d.CloudBackupEnumerate(&api.CloudBackupEnumerateRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			CredentialUUID: credBackupReq.CredentialUUID,
		},
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Backups, numbackups)

	// Now delete all
	err = d.CloudBackupDeleteAll(&api.CloudBackupDeleteAllRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			CredentialUUID: credBackupReq.CredentialUUID,
			SrcVolumeID:    vol.GetId(),
		},
	})
	assert.NoError(t, err)

	resp, err = d.CloudBackupEnumerate(&api.CloudBackupEnumerateRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			CredentialUUID: credBackupReq.CredentialUUID,
		},
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Backups, numbackups-1)
	// Now delete all
	err = d.CloudBackupDeleteAll(&api.CloudBackupDeleteAllRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			CredentialUUID: credBackupReq.CredentialUUID,
		},
	})
}

func TestFakeCloudBackupStatusWithoutMatches(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	numbackups := 50
	names := make([]string, 0)
	for i := 0; i < numbackups; i++ {
		// create backups
		id, backupid, credBackupReq, _ := testInitForCloudBackups(t, d)
		names = append(names, id)
		// create restores
		resp, err := d.CloudBackupRestore(&api.CloudBackupRestoreRequest{
			CredentialUUID:    credBackupReq.CredentialUUID,
			ID:                backupid,
			RestoreVolumeName: fmt.Sprintf("restore-%d", i),
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.RestoreVolumeID)
	}

	resp, err := d.CloudBackupStatus(&api.CloudBackupStatusRequest{})
	assert.NoError(t, err)

	// backups and restores
	// restores overwrite backups
	assert.Len(t, resp.Statuses, numbackups*2)

	var nbackups, nrestores int
	for _, status := range resp.Statuses {
		if status.OpType == api.CloudBackupOp {
			nbackups++
		} else {
			nrestores++
		}
	}
	assert.Equal(t, nbackups, 50)
	assert.Equal(t, nrestores, 50)

	resp, err = d.CloudBackupStatus(&api.CloudBackupStatusRequest{
		Local: true, // all where done on this single node fake cluster
	})
	assert.NoError(t, err)

	// backups and restores
	assert.Len(t, resp.Statuses, numbackups*2)

	nbackups = 0
	nrestores = 0
	for _, status := range resp.Statuses {
		if status.OpType == api.CloudBackupOp {
			nbackups++
		} else {
			nrestores++
		}
	}
	assert.Equal(t, 50, nbackups)
	assert.Equal(t, 50, nrestores)

	for _, id := range names {
		_, err := d.kv.Delete(backupsKeyPrefix + "/" + id)
		assert.NoError(t, err)
	}
}

func TestFakeCloudBackupStatusWithMatchingVolume(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	numbackups := 50
	var credBackupReq *api.CloudBackupCreateRequest
	var vol *api.Volume
	var backupId, name string
	for i := 0; i < numbackups; i++ {
		// create backups
		name, backupId, credBackupReq, vol = testInitForCloudBackups(t, d)

		// create restores
		resp, err := d.CloudBackupRestore(&api.CloudBackupRestoreRequest{
			CredentialUUID:    credBackupReq.CredentialUUID,
			ID:                backupId,
			RestoreVolumeName: fmt.Sprintf("restore-%d", i),
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.RestoreVolumeID)
	}

	resp, err := d.CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: vol.GetId(),
	})
	assert.NoError(t, err)

	// backups and restores
	assert.Len(t, resp.Statuses, 1)
	assert.Equal(t, api.CloudBackupOp, resp.Statuses[name].OpType)
}

func TestFakeCloudBackupCatalog(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	_, backupId, createReq, _ := testInitForCloudBackups(t, d)

	resp, err := d.CloudBackupCatalog(&api.CloudBackupCatalogRequest{
		CredentialUUID: createReq.CredentialUUID,
		ID:             backupId + "sdf",
	})
	assert.Error(t, err)

	resp, err = d.CloudBackupCatalog(&api.CloudBackupCatalogRequest{
		CredentialUUID: createReq.CredentialUUID,
		ID:             backupId,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Contents)
}

func TestFakeCloudBackupHistoryWithoutMatches(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	numbackups := 50
	for i := 0; i < numbackups; i++ {
		// create backups
		_, backupid, credBackupReq, _ := testInitForCloudBackups(t, d)

		// create restores
		resp, err := d.CloudBackupRestore(&api.CloudBackupRestoreRequest{
			CredentialUUID:    credBackupReq.CredentialUUID,
			ID:                backupid,
			RestoreVolumeName: fmt.Sprintf("restore-%d", i),
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.RestoreVolumeID)
	}

	resp, err := d.CloudBackupHistory(&api.CloudBackupHistoryRequest{})
	assert.NoError(t, err)

	// backups and restores
	assert.Len(t, resp.HistoryList, numbackups)
}

func TestFakeCloudBackupHistoryWithMatchingVolume(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	numbackups := 50
	var credBackupReq *api.CloudBackupCreateRequest
	var vol *api.Volume
	var id string
	for i := 0; i < numbackups; i++ {
		// create backups
		_, id, credBackupReq, vol = testInitForCloudBackups(t, d)

		// create restores
		resp, err := d.CloudBackupRestore(&api.CloudBackupRestoreRequest{
			CredentialUUID:    credBackupReq.CredentialUUID,
			ID:                id,
			RestoreVolumeName: fmt.Sprintf("restore-%d", i),
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.RestoreVolumeID)
	}

	resp, err := d.CloudBackupHistory(&api.CloudBackupHistoryRequest{
		SrcVolumeID: vol.GetId(),
	})
	assert.NoError(t, err)

	// Only backups
	assert.Len(t, resp.HistoryList, 1)
}

func TestFakeCloudBackupStateChange(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	name, _, _, vol := testInitForCloudBackups(t, d)

	// Update element on db
	var elem *fakeBackups
	_, err = d.kv.GetVal(backupsKeyPrefix+"/"+name, &elem)
	assert.NoError(t, err)
	elem.Status.Status = api.CloudBackupStatusActive
	_, err = d.kv.Update(backupsKeyPrefix+"/"+name, elem, 0)
	assert.NoError(t, err)

	// Confirm db
	statuses, err := d.CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: vol.GetId(),
	})
	assert.NoError(t, err)
	assert.Len(t, statuses.Statuses, 1)
	assert.Equal(t, api.CloudBackupStatusActive, statuses.Statuses[name].Status)

	// No values errors
	err = d.CloudBackupStateChange(&api.CloudBackupStateChangeRequest{})
	assert.Error(t, err)

	// Pause
	err = d.CloudBackupStateChange(&api.CloudBackupStateChangeRequest{
		Name:           name,
		RequestedState: api.CloudBackupRequestedStatePause,
	})
	assert.NoError(t, err)

	// Confirm db
	statuses, err = d.CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: vol.GetId(),
	})
	assert.NoError(t, err)
	assert.Len(t, statuses.Statuses, 1)
	assert.Equal(t, api.CloudBackupStatusPaused, statuses.Statuses[name].Status)

	// Resume
	err = d.CloudBackupStateChange(&api.CloudBackupStateChangeRequest{
		Name:           name,
		RequestedState: api.CloudBackupRequestedStateResume,
	})
	assert.NoError(t, err)

	// Confirm db
	statuses, err = d.CloudBackupStatus(&api.CloudBackupStatusRequest{
		Name: name,
	})
	assert.NoError(t, err)
	assert.Len(t, statuses.Statuses, 1)
	assert.Equal(t, api.CloudBackupStatusActive, statuses.Statuses[name].Status)

	// Stop
	err = d.CloudBackupStateChange(&api.CloudBackupStateChangeRequest{
		Name:           name,
		RequestedState: api.CloudBackupRequestedStateStop,
	})
	assert.NoError(t, err)

	// Confirm db
	statuses, err = d.CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: vol.GetId(),
	})
	assert.NoError(t, err)
	assert.Len(t, statuses.Statuses, 1)
	assert.Equal(t, api.CloudBackupStatusStopped, statuses.Statuses[name].Status)

	// Still stopped
	err = d.CloudBackupStateChange(&api.CloudBackupStateChangeRequest{
		Name:           name,
		RequestedState: api.CloudBackupRequestedStateResume,
	})
	assert.NoError(t, err)

	// Confirm db
	statuses, err = d.CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: vol.GetId(),
	})
	assert.NoError(t, err)
	assert.Len(t, statuses.Statuses, 1)
	assert.Equal(t, api.CloudBackupStatusStopped, statuses.Statuses[name].Status)
}

func TestFakeCloudBackupSchedule(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	_, _, req, vol := testInitForCloudBackups(t, d)

	maxbackups := uint(10)
	id, err := d.CloudBackupSchedCreate(&api.CloudBackupSchedCreateRequest{
		CloudBackupScheduleInfo: api.CloudBackupScheduleInfo{
			SrcVolumeID:    vol.GetId(),
			CredentialUUID: req.CredentialUUID,
			Schedule:       "yaml",
			MaxBackups:     maxbackups,
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	schedules, err := d.CloudBackupSchedEnumerate()
	assert.NoError(t, err)
	assert.NotEmpty(t, schedules.Schedules)
	assert.Equal(t, vol.GetId(), schedules.Schedules[id.UUID].SrcVolumeID)

	err = d.CloudBackupSchedDelete(&api.CloudBackupSchedDeleteRequest{
		UUID: id.UUID,
	})
	assert.NoError(t, err)

	schedules, err = d.CloudBackupSchedEnumerate()
	assert.NoError(t, err)
	assert.Empty(t, schedules.Schedules)
}

func TestFakeSet(t *testing.T) {
	d, err := newFakeDriver(map[string]string{})
	assert.NoError(t, err)

	// Create a vol
	name := "myvol"
	size := uint64(1234)
	volid, err := d.Create(&api.VolumeLocator{Name: name}, &api.Source{}, &api.VolumeSpec{
		Size:    size,
		HaLevel: 1,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, volid)

	// Set values
	err = d.Set(volid, &api.VolumeLocator{
		Name: "newname",
		VolumeLabels: map[string]string{
			"hello": "world",
		},
	}, &api.VolumeSpec{
		Size:    9876,
		HaLevel: 1,
		Journal: true,
	})
	assert.NoError(t, err)

	// Verify
	vols, err := d.Inspect([]string{volid})
	assert.NoError(t, err)
	assert.Len(t, vols, 1)
	assert.NotNil(t, vols[0])

	locator := vols[0].GetLocator()
	assert.NotNil(t, locator)
	assert.Equal(t, locator.GetName(), "newname")
	assert.Equal(t, locator.GetVolumeLabels()["hello"], "world")

	spec := vols[0].GetSpec()
	assert.Equal(t, spec.Size, uint64(9876))
	assert.Equal(t, spec.HaLevel, int64(1))
	assert.Equal(t, spec.Journal, true)
}
