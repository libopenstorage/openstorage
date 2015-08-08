package volume

import (
	"encoding/json"
	"fmt"
	_ "sync"

	"github.com/libopenstorage/kvdb"
	"github.com/libopenstorage/openstorage/api"
)

const (
	keyBase   = "openstorage/"
	locks     = "/locks/"
	volumes   = "/volumes/"
	snapshots = "/snapshots/"
)

type Store interface {
	// Lock volume specified by volID.
	Lock(volID api.VolumeID) (interface{}, error)

	// Lock volume with token obtained from call to Lock.
	Unlock(token interface{}) error

	// CreateVol returns error if volume with the same ID already existe.
	CreateVol(vol *api.Volume) error

	// GetVol from volID.
	GetVol(volID api.VolumeID) (*api.Volume, error)

	// UpdateVol with vol
	UpdateVol(vol *api.Volume) error

	// DeleteVol. Returns error if volume does not exist.
	DeleteVol(volID api.VolumeID) error

	// GetSnap from snapID
	GetSnap(snapID api.SnapID) (*api.VolumeSnap, error)

	// Update snap with snap
	UpdateSnap(snap *api.VolumeSnap) error

	// CreateSnap with new snap
	CreateSnap(snap *api.VolumeSnap) error

	// DeleteSnap with new snap
	DeleteSnap(snapID api.SnapID) error
}

// DefaultEnumerator for volume information. Implements the Enumerator Interface
type DefaultEnumerator struct {
	kvdb          kvdb.Kvdb
	driver        string
	lockKeyPrefix string
	volKeyPrefix  string
	snapKeyPrefix string
}

func (e *DefaultEnumerator) lockKey(volID api.VolumeID) string {
	return e.volKeyPrefix + string(volID) + ".lock"
}

func (e *DefaultEnumerator) snapKey(snapID api.SnapID) string {
	return e.snapKeyPrefix + string(snapID)
}

func (e *DefaultEnumerator) volKey(volID api.VolumeID) string {
	return e.volKeyPrefix + string(volID)
}

func hasSubset(set api.Labels, subset api.Labels) bool {
	if subset == nil {
		return true
	}
	if set == nil {
		return false
	}
	for k := range subset {
		if _, ok := set[k]; !ok {
			return false
		}
	}
	return true
}

func contains(volID api.VolumeID, set []api.VolumeID) bool {
	for _, v := range set {
		if v == volID {
			return true
		}
	}
	return false
}

func match(v *api.Volume, locator api.VolumeLocator, configLabels api.Labels) bool {
	if locator.Name != "" && v.Locator.Name != locator.Name {
		return false
	}
	if !hasSubset(v.Locator.VolumeLabels, locator.VolumeLabels) {
		return false
	}
	return hasSubset(v.Spec.ConfigLabels, configLabels)
}

// NewDefaultEnumerator initializes store with specified kvdb.
func NewDefaultEnumerator(driver string, kvdb kvdb.Kvdb) *DefaultEnumerator {
	return &DefaultEnumerator{
		kvdb:          kvdb,
		driver:        driver,
		lockKeyPrefix: keyBase + driver + locks,
		volKeyPrefix:  keyBase + driver + volumes,
		snapKeyPrefix: keyBase + driver + snapshots,
	}
}

// Lock volume specified by volID.
func (e *DefaultEnumerator) Lock(volID api.VolumeID) (interface{}, error) {
	return e.kvdb.Lock(e.lockKey(volID), 10)
}

// Lock volume with token obtained from call to Lock.
func (e *DefaultEnumerator) Unlock(token interface{}) error {
	v, ok := token.(*kvdb.KVPair)
	if !ok {
		return fmt.Errorf("Invalid token of type %T", token)
	}
	return e.kvdb.Unlock(v)
}

// CreateVol returns error if volume with the same ID already existe.
func (e *DefaultEnumerator) CreateVol(vol *api.Volume) error {
	_, err := e.kvdb.Create(e.volKey(vol.ID), vol, 0)
	return err
}

// GetVol from volID.
func (e *DefaultEnumerator) GetVol(volID api.VolumeID) (*api.Volume, error) {
	var v api.Volume
	_, err := e.kvdb.GetVal(e.volKey(volID), &v)

	return &v, err
}

// UpdateVol with vol
func (e *DefaultEnumerator) UpdateVol(vol *api.Volume) error {
	_, err := e.kvdb.Put(e.volKey(vol.ID), vol, 0)
	return err
}

// DeleteVol. Returns error if volume does not exist.
func (e *DefaultEnumerator) DeleteVol(volID api.VolumeID) error {
	_, err := e.kvdb.Delete(e.volKey(volID))
	return err
}

// GetSnap from snapID
func (e *DefaultEnumerator) GetSnap(snapID api.SnapID) (*api.VolumeSnap, error) {
	var snap api.VolumeSnap
	_, err := e.kvdb.GetVal(e.snapKey(snapID), &snap)

	return &snap, err
}

// Update snap with snap
func (e *DefaultEnumerator) UpdateSnap(snap *api.VolumeSnap) error {
	_, err := e.kvdb.Put(e.snapKey(snap.ID), snap, 0)
	return err
}

// CreateSnap with new snap
func (e *DefaultEnumerator) CreateSnap(snap *api.VolumeSnap) error {
	_, err := e.kvdb.Create(e.snapKey(snap.ID), snap, 0)
	return err
}

// DeleteSnap with new snap
func (e *DefaultEnumerator) DeleteSnap(snapID api.SnapID) error {
	_, err := e.kvdb.Delete(e.snapKey(snapID))
	return err
}

// Inspect specified volumee.
// Errors ErrEnoEnt may be returned.
func (e *DefaultEnumerator) Inspect(ids []api.VolumeID) ([]api.Volume, error) {
	var err error
	var vol *api.Volume
	vols := make([]api.Volume, 0, len(ids))

	for _, v := range ids {
		vol, err = e.GetVol(v)
		if err != nil {
			break
		}
		vols = append(vols, *vol)
	}
	return vols, err
}

// Enumerate volumes that map to the volumeLocator. Locator fields may be regexp.
// If locator fields are left blank, this will return all volumee.
func (e *DefaultEnumerator) Enumerate(locator api.VolumeLocator,
	labels api.Labels) ([]api.Volume, error) {

	kvp, err := e.kvdb.Enumerate(e.volKeyPrefix)
	if err != nil {
		return nil, err
	}
	vols := make([]api.Volume, 0, len(kvp))
	for _, v := range kvp {
		var elem api.Volume
		err = json.Unmarshal(v.Value, &elem)
		if err != nil {
			return nil, err
		}
		if match(&elem, locator, labels) {
			vols = append(vols, elem)
		}
	}
	return vols, nil
}

// SnapInspect provides details on this snapshot.
// Errors ErrEnoEnt may be returned
func (e *DefaultEnumerator) SnapInspect(ids []api.SnapID) ([]api.VolumeSnap, error) {
	var err error
	var snap *api.VolumeSnap
	snaps := make([]api.VolumeSnap, 0, len(ids))

	for _, v := range ids {
		snap, err = e.GetSnap(v)
		if err != nil {
			break
		}
		snaps = append(snaps, *snap)
	}
	return snaps, err
}

// Enumerate snaps for specified volume
func (e *DefaultEnumerator) SnapEnumerate(
	volIDs []api.VolumeID,
	snapLabels api.Labels) ([]api.VolumeSnap, error) {

	kvp, err := e.kvdb.Enumerate(e.snapKeyPrefix)
	if err != nil {
		return nil, err
	}
	snaps := make([]api.VolumeSnap, 0, len(kvp))
	for _, v := range kvp {
		var elem api.VolumeSnap
		err = json.Unmarshal(v.Value, &elem)
		if err != nil {
			return nil, err
		}
		if volIDs != nil && !contains(elem.VolumeID, volIDs) {
			continue
		}
		if hasSubset(elem.SnapLabels, snapLabels) {
			snaps = append(snaps, elem)
		}
	}
	return snaps, nil
}
