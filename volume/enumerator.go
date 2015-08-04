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

type StoreUpdate interface {
	// Lock volume specified by volID.
	Lock(volID api.VolumeID) (interface{}, error)

	// Lock volume with token obtained from call to Lock.
	Unlock(token interface{}) error

	// CreateVol returns error if volume with the same ID already exists.
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

// Store for volume information. Implements the Enumerator Interface
type Store struct {
	kvdb          kvdb.Kvdb
	driver        string
	lockKeyPrefix string
	volKeyPrefix  string
	snapKeyPrefix string
}

func (s *Store) lockKey(volID api.VolumeID) string {
	return s.volKeyPrefix + string(volID) + ".lock"
}

func (s *Store) snapKey(snapID api.SnapID) string {
	return s.snapKeyPrefix + string(snapID)
}

func (s *Store) volKey(volID api.VolumeID) string {
	return s.volKeyPrefix + string(volID)
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

func match(v *api.Volume, locator api.VolumeLocator, configLabels api.Labels) bool {
	if locator.Name != "" && v.Locator.Name != locator.Name {
		return false
	}
	if !hasSubset(v.Locator.VolumeLabels, locator.VolumeLabels) {
		return false
	}
	return hasSubset(v.Spec.ConfigLabels, configLabels)
}

func (s *Store) enumerate(
	locator api.VolumeLocator,
	configLabels api.Labels,
	volumeState api.VolumeState) ([]api.Volume, error) {

	kvp, err := s.kvdb.Enumerate(s.volKeyPrefix)
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
		if (volumeState & elem.State) == 0 {
			continue
		}
		if match(&elem, locator, configLabels) {
			vols = append(vols, elem)
		}
	}
	return vols, nil
}

func (s *Store) snapEnumerate(
	locator api.VolumeLocator,
	snapLabels api.Labels) ([]api.VolumeSnap, error) {

	kvp, err := s.kvdb.Enumerate(s.snapKeyPrefix)
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
		vol, err := s.GetVol(elem.VolumeID)
		if err != nil {
			return nil, err
		}
		if match(vol, locator, nil) {
			snaps = append(snaps, elem)
			continue
		}
		if hasSubset(elem.SnapLabels, snapLabels) {
			snaps = append(snaps, elem)
		}
	}
	return snaps, nil
}

// NewStore initializes store with specified kvdb.
func NewStore(driver string, kvdb kvdb.Kvdb) *Store {
	return &Store{
		kvdb:          kvdb,
		driver:        driver,
		lockKeyPrefix: keyBase + driver + locks,
		volKeyPrefix:  keyBase + driver + volumes,
		snapKeyPrefix: keyBase + driver + snapshots,
	}
}

// Lock volume specified by volID.
func (s *Store) Lock(volID api.VolumeID) (interface{}, error) {
	return s.kvdb.Lock(s.lockKey(volID), 10)
}

// Lock volume with token obtained from call to Lock.
func (s *Store) Unlock(token interface{}) error {
	v, ok := token.(*kvdb.KVPair)
	if !ok {
		return fmt.Errorf("Invalid token of type %T", token)
	}
	return s.kvdb.Unlock(v)
}

// CreateVol returns error if volume with the same ID already exists.
func (s *Store) CreateVol(vol *api.Volume) error {
	_, err := s.kvdb.Create(s.volKey(vol.ID), vol, 0)
	return err
}

// GetVol from volID.
func (s *Store) GetVol(volID api.VolumeID) (*api.Volume, error) {
	var v api.Volume
	_, err := s.kvdb.GetVal(s.volKey(volID), &v)

	return &v, err
}

// UpdateVol with vol
func (s *Store) UpdateVol(vol *api.Volume) error {
	_, err := s.kvdb.Put(s.volKey(vol.ID), vol, 0)
	return err
}

// DeleteVol. Returns error if volume does not exist.
func (s *Store) DeleteVol(volID api.VolumeID) error {
	_, err := s.kvdb.Delete(s.volKey(volID))
	return err
}

// GetSnap from snapID
func (s *Store) GetSnap(snapID api.SnapID) (*api.VolumeSnap, error) {
	var snap api.VolumeSnap
	_, err := s.kvdb.GetVal(s.snapKey(snapID), &snap)

	return &snap, err
}

// Update snap with snap
func (s *Store) UpdateSnap(snap *api.VolumeSnap) error {
	_, err := s.kvdb.Put(s.snapKey(snap.ID), snap, 0)
	return err
}

// CreateSnap with new snap
func (s *Store) CreateSnap(snap *api.VolumeSnap) error {
	_, err := s.kvdb.Create(s.snapKey(snap.ID), snap, 0)
	return err
}

// DeleteSnap with new snap
func (s *Store) DeleteSnap(snapID api.SnapID) error {
	_, err := s.kvdb.Delete(s.snapKey(snapID))
	return err
}

// Inspect specified volumes.
// Errors ErrEnoEnt may be returned.
func (s *Store) Inspect(ids []api.VolumeID) ([]api.Volume, error) {
	var err error
	var vol *api.Volume
	vols := make([]api.Volume, 0, len(ids))

	for _, v := range ids {
		vol, err = s.GetVol(v)
		if err != nil {
			break
		}
		vols = append(vols, *vol)
	}
	return vols, err
}

// Enumerate volumes that map to the volumeLocator. Locator fields may be regexp.
// If locator fields are left blank, this will return all volumes.
func (s *Store) Enumerate(locator api.VolumeLocator, labels api.Labels) ([]api.Volume, error) {

	vols, err := s.enumerate(locator, labels, api.VolumeStateAny)
	return vols, err
}

// SnapInspect provides details on this snapshot.
// Errors ErrEnoEnt may be returned
func (s *Store) SnapInspect(ids []api.SnapID) ([]api.VolumeSnap, error) {
	var err error
	var snap *api.VolumeSnap
	snaps := make([]api.VolumeSnap, 0, len(ids))

	for _, v := range ids {
		snap, err = s.GetSnap(v)
		if err != nil {
			break
		}
		snaps = append(snaps, *snap)
	}
	return snaps, err
}

// Enumerate snaps for specified volume
// Count indicates the number of snaps populated.
func (s *Store) SnapEnumerate(locator api.VolumeLocator, labels api.Labels) ([]api.VolumeSnap, error) {
	snaps, err := s.snapEnumerate(locator, labels)
	return snaps, err
}
