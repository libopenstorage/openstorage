package volume

import (
	"encoding/json"
	"fmt"
	_ "sync"

	"github.com/portworx/kvdb"

	"github.com/libopenstorage/openstorage/api"
)

const (
	keyBase = "openstorage/"
	locks   = "/locks/"
	volumes = "/volumes/"
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
}

// DefaultEnumerator for volume information. Implements the Enumerator Interface
type DefaultEnumerator struct {
	kvdb          kvdb.Kvdb
	driver        string
	lockKeyPrefix string
	volKeyPrefix  string
}

func (e *DefaultEnumerator) lockKey(volID api.VolumeID) string {
	return e.volKeyPrefix + string(volID) + ".lock"
}

func (e *DefaultEnumerator) volKey(volID api.VolumeID) string {
	return e.volKeyPrefix + string(volID)
}

func hasSubset(set api.Labels, subset api.Labels) bool {
	if subset == nil || len(subset) == 0 {
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
	if len(set) == 0 {
		return true
	}
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

// Inspect specified volumes.
// Returns slice of volumes that were found.
func (e *DefaultEnumerator) Inspect(ids []api.VolumeID) ([]api.Volume, error) {
	var err error
	var vol *api.Volume
	vols := make([]api.Volume, 0, len(ids))

	for _, v := range ids {
		vol, err = e.GetVol(v)
		// XXX Distinguish between ENOENT and an internal error from KVDB
		if err != nil {
			continue
		}
		vols = append(vols, *vol)
	}
	return vols, nil
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

// SnapEnumerate for specified volume
func (e *DefaultEnumerator) SnapEnumerate(
	volIDs []api.VolumeID,
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
		if elem.Source == nil ||
			elem.Source.Parent == api.BadVolumeID ||
			(volIDs != nil && !contains(elem.Source.Parent, volIDs)) {
			continue
		}
		if hasSubset(elem.Locator.VolumeLabels, labels) {
			vols = append(vols, elem)
		}
	}
	return vols, nil
}
