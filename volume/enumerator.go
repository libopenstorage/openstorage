package volume

import (
	"bytes"
	"fmt"
	// TODO(pedge): what is this for?
	_ "sync"

	"github.com/portworx/kvdb"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/jsonpb"
)

const (
	keyBase = "openstorage"
)

type Store interface {
	// Lock volume specified by volumeID.
	Lock(volumeID string) (interface{}, error)

	// Lock volume with token obtained from call to Lock.
	Unlock(token interface{}) error

	// CreateVol returns error if volume with the same ID already existe.
	CreateVol(vol *api.Volume) error

	// GetVol from volumeID.
	GetVol(volumeID string) (*api.Volume, error)

	// UpdateVol with vol
	UpdateVol(vol *api.Volume) error

	// DeleteVol. Returns error if volume does not exist.
	DeleteVol(volumeID string) error
}

// DefaultEnumerator for volume information. Implements the Enumerator Interface
type DefaultEnumerator struct {
	kvdb   kvdb.Kvdb
	driver string
}

// NewDefaultEnumerator initializes store with specified kvdb.
func NewDefaultEnumerator(driver string, kvdb kvdb.Kvdb) *DefaultEnumerator {
	return &DefaultEnumerator{
		kvdb:   kvdb,
		driver: driver,
	}
}

// Lock volume specified by volumeID.
func (e *DefaultEnumerator) Lock(volumeID string) (interface{}, error) {
	return e.kvdb.Lock(e.lockKey(volumeID), 10)
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
	_, err := e.kvdb.Create(e.volKey(vol.Id), vol, 0)
	return err
}

// GetVol from volumeID.
func (e *DefaultEnumerator) GetVol(volumeID string) (*api.Volume, error) {
	var v api.Volume
	_, err := e.kvdb.GetVal(e.volKey(volumeID), &v)
	return &v, err
}

// UpdateVol with vol
func (e *DefaultEnumerator) UpdateVol(vol *api.Volume) error {
	_, err := e.kvdb.Put(e.volKey(vol.Id), vol, 0)
	return err
}

// DeleteVol. Returns error if volume does not exist.
func (e *DefaultEnumerator) DeleteVol(volumeID string) error {
	_, err := e.kvdb.Delete(e.volKey(volumeID))
	return err
}

// Inspect specified volumes.
// Returns slice of volumes that were found.
func (e *DefaultEnumerator) Inspect(ids []string) ([]*api.Volume, error) {
	volumes := make([]*api.Volume, 0, len(ids))
	for _, id := range ids {
		volume, err := e.GetVol(id)
		// XXX Distinguish between ENOENT and an internal error from KVDB
		if err != nil {
			continue
		}
		volumes = append(volumes, volume)
	}
	return volumes, nil
}

// Enumerate volumes that map to the volumeLocator. Locator fields may be regexp.
// If locator fields are left blank, this will return all volumee.
func (e *DefaultEnumerator) Enumerate(
	locator *api.VolumeLocator,
	labels map[string]string,
) ([]*api.Volume, error) {

	kvp, err := e.kvdb.Enumerate(e.volKeyPrefix())
	if err != nil {
		return nil, err
	}
	volumes := make([]*api.Volume, 0, len(kvp))
	for _, v := range kvp {
		elem := &api.Volume{}
		if err := jsonpb.Unmarshal(bytes.NewReader(v.Value), elem); err != nil {
			return nil, err
		}
		if match(elem, locator, labels) {
			volumes = append(volumes, elem)
		}
	}
	return volumes, nil
}

// SnapEnumerate for specified volume
func (e *DefaultEnumerator) SnapEnumerate(
	volumeIDs []string,
	labels map[string]string,
) ([]*api.Volume, error) {
	kvp, err := e.kvdb.Enumerate(e.volKeyPrefix())
	if err != nil {
		return nil, err
	}
	volumes := make([]*api.Volume, 0, len(kvp))
	for _, v := range kvp {
		elem := &api.Volume{}
		if err := jsonpb.Unmarshal(bytes.NewReader(v.Value), elem); err != nil {
			return nil, err
		}
		if elem.Source == nil ||
			elem.Source.Parent == "" ||
			(volumeIDs != nil && !contains(elem.Source.Parent, volumeIDs)) {
			continue
		}
		if hasSubset(elem.Locator.VolumeLabels, labels) {
			volumes = append(volumes, elem)
		}
	}
	return volumes, nil
}

func (e *DefaultEnumerator) lockKey(volumeID string) string {
	return e.volKeyPrefix() + volumeID + ".lock"
}

func (e *DefaultEnumerator) volKey(volumeID string) string {
	return e.volKeyPrefix() + volumeID
}

// TODO(pedge): not used - bug?
func (d *DefaultEnumerator) lockKeyPrefix() string {
	return fmt.Sprintf("%s/%s/locks/", keyBase, d.driver)
}

func (d *DefaultEnumerator) volKeyPrefix() string {
	return fmt.Sprintf("%s/%s/volumes/", keyBase, d.driver)
}

func hasSubset(set map[string]string, subset map[string]string) bool {
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

func contains(volumeID string, set []string) bool {
	if len(set) == 0 {
		return true
	}
	for _, v := range set {
		if v == volumeID {
			return true
		}
	}
	return false
}

func match(
	v *api.Volume,
	locator *api.VolumeLocator,
	configLabels map[string]string,
) bool {
	if locator.Name != "" && v.Locator.Name != locator.Name {
		return false
	}
	if !hasSubset(v.Locator.VolumeLabels, locator.VolumeLabels) {
		return false
	}
	return hasSubset(v.Spec.ConfigLabels, configLabels)
}
