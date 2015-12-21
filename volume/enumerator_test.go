package volume

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/stretchr/testify/assert"
)

var (
	e        *DefaultEnumerator
	volName  = "TestVolume"
	snapName = "SnapVolume"
	labels   = map[string]string{"Foo": "DEADBEEF"}
)

func TestInspect(t *testing.T) {
	id := volName
	vol := api.Volume{
		Id:      id,
		Locator: api.VolumeLocator{Name: volName, VolumeLabels: labels},
		State:   api.VolumeAvailable,
		Spec:    &api.VolumeSpec{},
	}
	err := e.CreateVol(&vol)
	assert.NoError(t, err, "Failed in CreateVol")
	vols, err := e.Inspect([]string{id})
	assert.NoError(t, err, "Failed in Inspect")
	assert.Equal(t, len(vols), 1, "Number of volumes returned in inspect should be 1")
	if len(vols) == 1 {
		assert.Equal(t, vols[0].Id, vol.Id, "Invalid volume returned in Inspect")
	}
	err = e.DeleteVol(id)
	assert.NoError(t, err, "Failed in Delete")
	vols, err = e.Inspect([]string{id})
	assert.NotNil(t, vols, "Inspect returned nil vols")
	assert.Equal(t, len(vols), 0, "Number of volumes returned in inspect should be 0")
}

func TestEnumerate(t *testing.T) {
	id := volName
	vol := api.Volume{
		Id:      id,
		Locator: api.VolumeLocator{Name: volName, VolumeLabels: labels},
		State:   api.VolumeAvailable,
		Spec:    &api.VolumeSpec{},
	}
	err := e.CreateVol(&vol)
	assert.NoError(t, err, "Failed in CreateVol")
	vols, err := e.Enumerate(api.VolumeLocator{}, nil)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.Equal(t, 1, len(vols), "Number of volumes returned in enumerate should be 1")

	vols, err = e.Enumerate(api.VolumeLocator{Name: volName}, nil)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.Equal(t, 1, len(vols), "Number of volumes returned in enumerate should be 1")
	if len(vols) == 1 {
		assert.Equal(t, vols[0].Id, vol.Id, "Invalid volume returned in Enumerate")
	}
	vols, err = e.Enumerate(api.VolumeLocator{VolumeLabels: labels}, nil)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.Equal(t, len(vols), 1, "Number of volumes returned in enumerate should be 1")
	if len(vols) == 1 {
		assert.Equal(t, vols[0].Id, vol.Id, "Invalid volume returned in Enumerate")
	}
	err = e.DeleteVol(id)
	assert.NoError(t, err, "Failed in Delete")
	vols, err = e.Enumerate(api.VolumeLocator{Name: volName}, nil)
	assert.Equal(t, len(vols), 0, "Number of volumes returned in enumerate should be 0")
}

func TestSnapEnumerate(t *testing.T) {
	snapID := snapName
	id := volName
	vol := api.Volume{
		Id:      id,
		Locator: api.VolumeLocator{Name: volName, VolumeLabels: labels},
		State:   api.VolumeAvailable,
		Spec:    &api.VolumeSpec{},
	}
	err := e.CreateVol(&vol)
	assert.NoError(t, err, "Failed in CreateVol")
	snap := api.Volume{
		Id:      snapID,
		Locator: api.VolumeLocator{Name: volName, VolumeLabels: labels},
		State:   api.VolumeAvailable,
		Spec:    &api.VolumeSpec{},
		Source:  &api.Source{Parent: id},
	}
	err = e.CreateVol(&snap)
	assert.NoError(t, err, "Failed in CreateSnap")

	snaps, err := e.SnapEnumerate([]string{id}, nil)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.Equal(t, len(snaps), 1, "Number of snaps returned in enumerate should be 1")
	if len(snaps) == 1 {
		assert.Equal(t, snaps[0].Id, snap.Id, "Invalid snap returned in Enumerate")
	}
	snaps, err = e.SnapEnumerate([]string{id}, labels)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.Equal(t, len(snaps), 1, "Number of snaps returned in enumerate should be 1")
	if len(snaps) == 1 {
		assert.Equal(t, snaps[0].Id, snap.Id, "Invalid snap returned in Enumerate")
	}

	snaps, err = e.SnapEnumerate(nil, labels)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.True(t, len(snaps) >= 1, "Number of snaps returned in enumerate should be at least 1")
	if len(snaps) == 1 {
		assert.Equal(t, snaps[0].Id, snap.Id, "Invalid snap returned in Enumerate")
	}

	snaps, err = e.SnapEnumerate(nil, nil)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.True(t, len(snaps) >= 1, "Number of snaps returned in enumerate should be at least 1")
	if len(snaps) == 1 {
		assert.Equal(t, snaps[0].Id, snap.Id, "Invalid snap returned in Enumerate")
	}

	err = e.DeleteVol(snapID)
	assert.NoError(t, err, "Failed in Delete")
	snaps, err = e.SnapEnumerate([]string{id}, labels)
	assert.NotNil(t, snaps, "Inspect returned nil snaps")
	assert.Equal(t, len(snaps), 0, "Number of snaps returned in enumerate should be 0")

	err = e.DeleteVol(id)
	assert.NoError(t, err, "Failed in Delete")
}

func init() {
	kv, err := kvdb.New(mem.Name, "driver_test", []string{}, nil)
	if err != nil {
		logrus.Panicf("Failed to intialize KVDB")
	}
	err = kvdb.SetInstance(kv)
	if err != nil {
		logrus.Panicf("Failed to set KVDB instance")
	}

	e = NewDefaultEnumerator("enumerator_test", kv)
}
