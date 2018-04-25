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

	"github.com/sirupsen/logrus"

	"strings"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers/common"
	"github.com/pborman/uuid"
	"github.com/portworx/kvdb"
)

const (
	Name = "fake"
	Type = api.DriverType_DRIVER_TYPE_BLOCK
)

// Implements the open storage volume interface.
type driver struct {
	volume.IODriver
	volume.StoreEnumerator
	volume.StatsDriver
	volume.QuiesceDriver
	volume.CredsDriver
	volume.CloudBackupDriver
}

func Init(params map[string]string) (volume.VolumeDriver, error) {
	inst := &driver{
		IODriver:          volume.IONotSupported,
		StoreEnumerator:   common.NewDefaultStoreEnumerator(Name, kvdb.Instance()),
		StatsDriver:       volume.StatsNotSupported,
		QuiesceDriver:     volume.QuiesceNotSupported,
		CredsDriver:       volume.CredsNotSupported,
		CloudBackupDriver: volume.CloudBackupNotSupported,
	}

	volumeInfo, err := inst.StoreEnumerator.Enumerate(&api.VolumeLocator{}, nil)
	if err == nil {
		for _, info := range volumeInfo {
			if info.Status == api.VolumeStatus_VOLUME_STATUS_NONE {
				info.Status = api.VolumeStatus_VOLUME_STATUS_UP
				inst.UpdateVol(info)
			}
		}
	}

	logrus.Println("Fake driver initialized")
	return inst, nil
}

func (d *driver) Name() string {
	return Name
}

func (d *driver) Type() api.DriverType {
	return Type
}

// Status diagnostic information
func (d *driver) Status() [][2]string {
	return [][2]string{}
}

//
// These functions below implement the volume driver interface.
//

func (d *driver) Create(
	locator *api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec) (string, error) {

	if spec.Size == 0 {
		return "", fmt.Errorf("Volume size cannot be zero")
	}

	volumeID := strings.TrimSuffix(uuid.New(), "\n")

	if _, err := d.GetVol(volumeID); err == nil {
		return "", fmt.Errorf("volume with that id already exists")
	}

	// snapshot passes nil volumelabels
	if locator.VolumeLabels == nil {
		locator.VolumeLabels = make(map[string]string)
	}

	v := common.NewVolume(
		volumeID,
		api.FSType_FS_TYPE_XFS,
		locator,
		source,
		spec,
	)

	if err := d.CreateVol(v); err != nil {
		return "", err
	}
	return v.Id, nil
}

func (d *driver) Delete(volumeID string) error {
	_, err := d.GetVol(volumeID)
	if err != nil {
		logrus.Println(err)
		return err
	}

	err = d.DeleteVol(volumeID)
	if err != nil {
		logrus.Println(err)
		return err
	}

	return nil
}

func (d *driver) MountedAt(mountpath string) string {
	return ""
}

func (d *driver) Mount(volumeID string, mountpath string, options map[string]string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		logrus.Println(err)
		return err
	}

	v.AttachPath = append(v.AttachPath, mountpath)
	return d.UpdateVol(v)
}

func (d *driver) Unmount(volumeID string, mountpath string, options map[string]string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return err
	}
	if len(v.AttachPath) == 0 {
		return fmt.Errorf("Device %v not mounted", volumeID)
	}

	v.AttachPath = nil
	return d.UpdateVol(v)
}

func (d *driver) Snapshot(volumeID string, readonly bool, locator *api.VolumeLocator) (string, error) {
	volIDs := []string{volumeID}
	vols, err := d.Inspect(volIDs)
	if err != nil {
		return "", nil
	}
	source := &api.Source{Parent: volumeID}
	logrus.Infof("Creating snap vol name: %s", locator.Name)
	newVolumeID, err := d.Create(locator, source, vols[0].Spec)
	if err != nil {
		return "", nil
	}

	return newVolumeID, nil
}

func (d *driver) Restore(volumeID string, snapID string) error {
	if _, err := d.Inspect([]string{volumeID, snapID}); err != nil {
		return err
	}

	return nil
}

func (d *driver) SnapshotGroup(groupID string, labels map[string]string) (*api.GroupSnapCreateResponse, error) {

	// We can return something here.
	return nil, volume.ErrNotSupported
}

func (d *driver) Attach(volumeID string, attachOptions map[string]string) (string, error) {
	return "/dev/fake/" + volumeID, nil
}

func (d *driver) Detach(volumeID string, options map[string]string) error {
	return nil
}

func (d *driver) Set(volumeID string, locator *api.VolumeLocator, spec *api.VolumeSpec) error {
	if spec != nil {
		return volume.ErrNotSupported
	}
	v, err := d.GetVol(volumeID)
	if err != nil {
		return err
	}
	if locator != nil {
		v.Locator = locator
	}
	return d.UpdateVol(v)
}

func (d *driver) Shutdown() {}
