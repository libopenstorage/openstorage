package volume

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/libopenstorage/kvdb"
	"github.com/libopenstorage/kvdb/mem"
	"github.com/libopenstorage/openstorage/api"
)

var (
	store    *Store
	volName  = "TestVolume"
	snapName = "SnapVolume"
	labels   = api.Labels{"Foo": "DEADBEEF"}
)

func TestInspect(t *testing.T) {
	id := api.VolumeID(volName)
	vol := api.Volume{
		ID:      id,
		Locator: api.VolumeLocator{Name: volName, VolumeLabels: labels},
		State:   api.VolumeAvailable,
		Spec:    &api.VolumeSpec{},
	}
	err := store.CreateVol(&vol)
	assert.NoError(t, err, "Failed in CreateVol")
	vols, err := store.Inspect([]api.VolumeID{id})
	assert.NoError(t, err, "Failed in Inspect")
	assert.Equal(t, len(vols), 1, "Number of volumes returned in inspect should be 1")
	if len(vols) == 1 {
		assert.Equal(t, vols[0].ID, vol.ID, "Invalid volume returned in Inspect")
	}
	err = store.DeleteVol(id)
	assert.NoError(t, err, "Failed in Delete")
	vols, err = store.Inspect([]api.VolumeID{id})
	assert.Error(t, err, "Inspect should return error for deleted volume")
	assert.NotNil(t, vols, "Inspect returned nil vols")
	assert.Equal(t, len(vols), 0, "Number of volumes returned in inspect should be 0")
}

func TestEnumerate(t *testing.T) {
	id := api.VolumeID(volName)
	vol := api.Volume{
		ID:      id,
		Locator: api.VolumeLocator{Name: volName, VolumeLabels: labels},
		State:   api.VolumeAvailable,
		Spec:    &api.VolumeSpec{},
	}
	err := store.CreateVol(&vol)
	assert.NoError(t, err, "Failed in CreateVol")
	vols, err := store.Enumerate(api.VolumeLocator{Name: volName}, nil)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.Equal(t, 1, len(vols), "Number of volumes returned in enumerate should be 1")
	if len(vols) == 1 {
		assert.Equal(t, vols[0].ID, vol.ID, "Invalid volume returned in Enumerate")
	}
	vols, err = store.Enumerate(api.VolumeLocator{VolumeLabels: labels}, nil)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.Equal(t, len(vols), 1, "Number of volumes returned in enumerate should be 1")
	if len(vols) == 1 {
		assert.Equal(t, vols[0].ID, vol.ID, "Invalid volume returned in Enumerate")
	}
	err = store.DeleteVol(id)
	assert.NoError(t, err, "Failed in Delete")
	vols, err = store.Enumerate(api.VolumeLocator{Name: volName}, nil)
	assert.Equal(t, len(vols), 0, "Number of volumes returned in enumerate should be 0")
}

func TestSnapInspect(t *testing.T) {
	snapID := api.SnapID(snapName)
	id := api.VolumeID(volName)
	snap := api.VolumeSnap{
		ID:         snapID,
		VolumeID:   id,
		SnapLabels: labels,
	}
	err := store.CreateSnap(&snap)
	assert.NoError(t, err, "Failed in CreateSnap")

	snaps, err := store.SnapInspect([]api.SnapID{snapID})
	assert.NoError(t, err, "Failed in Inspect")
	assert.Equal(t, len(snaps), 1, "Number of snaps returned in Inspect should be 1")
	if len(snaps) == 1 {
		assert.Equal(t, snaps[0].ID, snap.ID, "Invalid snap returned in Enumerate")
	}

	err = store.DeleteSnap(snapID)
	assert.NoError(t, err, "Failed in Delete")
	snaps, err = store.SnapEnumerate(api.VolumeLocator{Name: volName}, nil)
	assert.Equal(t, len(snaps), 0, "Number of snaps returned in enumerate should be 1")
}

func TestSnapEnumerate(t *testing.T) {
	snapID := api.SnapID(snapName)
	id := api.VolumeID(volName)
	vol := api.Volume{
		ID:      id,
		Locator: api.VolumeLocator{Name: volName, VolumeLabels: labels},
		State:   api.VolumeAvailable,
		Spec:    &api.VolumeSpec{},
	}
	err := store.CreateVol(&vol)
	assert.NoError(t, err, "Failed in CreateVol")
	snap := api.VolumeSnap{
		ID:         snapID,
		VolumeID:   id,
		SnapLabels: labels,
	}
	err = store.CreateSnap(&snap)
	assert.NoError(t, err, "Failed in CreateSnap")

	snaps, err := store.SnapEnumerate(api.VolumeLocator{Name: volName}, nil)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.Equal(t, len(snaps), 1, "Number of snaps returned in enumerate should be 1")
	if len(snaps) == 1 {
		assert.Equal(t, snaps[0].ID, snap.ID, "Invalid snap returned in Enumerate")
	}
	snaps, err = store.SnapEnumerate(api.VolumeLocator{VolumeLabels: labels}, nil)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.Equal(t, len(snaps), 1, "Number of snaps returned in enumerate should be 1")
	if len(snaps) == 1 {
		assert.Equal(t, snaps[0].ID, snap.ID, "Invalid snap returned in Enumerate")
	}

	snaps, err = store.SnapEnumerate(api.VolumeLocator{}, labels)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.True(t, len(snaps) >= 1, "Number of snaps returned in enumerate should be at least 1")
	if len(snaps) == 1 {
		assert.Equal(t, snaps[0].ID, snap.ID, "Invalid snap returned in Enumerate")
	}

	snaps, err = store.SnapEnumerate(api.VolumeLocator{}, nil)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.True(t, len(snaps) >= 1, "Number of snaps returned in enumerate should be at least 1")
	if len(snaps) == 1 {
		assert.Equal(t, snaps[0].ID, snap.ID, "Invalid snap returned in Enumerate")
	}

	err = store.DeleteSnap(snapID)
	assert.NoError(t, err, "Failed in Delete")
	snaps, err = store.SnapEnumerate(api.VolumeLocator{Name: volName}, nil)
	assert.NotNil(t, snaps, "Inspect returned nil snaps")
	assert.Equal(t, len(snaps), 0, "Number of snaps returned in enumerate should be 0")

	err = store.DeleteVol(id)
	assert.NoError(t, err, "Failed in Delete")
}

func init() {
	kv, err := kvdb.New(mem.Name, "driver_test", []string{}, nil)
	if err != nil {
		log.Panicf("Failed to intialize KVDB")
	}
	err = kvdb.SetInstance(kv)
	if err != nil {
		log.Panicf("Failed to set KVDB instance")
	}

	store = NewStore("enumerator_test", kv)
}
