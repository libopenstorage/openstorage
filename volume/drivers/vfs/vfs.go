package vfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"go.pedge.io/dlog"
	
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers/common"
	"github.com/pborman/uuid"
	"github.com/portworx/kvdb"
)

const (
	Name       = "vfs"
	Type       = api.DriverType_DRIVER_TYPE_FILE
)

func init() {
	volume.Register(Name, Init)
}

type driver struct {
	*volume.IoNotSupported
	*volume.DefaultBlockDriver
	*volume.DefaultEnumerator
	*volume.SnapshotNotSupported
}

// Init Driver intialization.
func Init(params map[string]string) (volume.VolumeDriver, error) {
	return &driver{
		IoNotSupported:    &volume.IoNotSupported{},
		DefaultEnumerator: volume.NewDefaultEnumerator(Name, kvdb.Instance()),
	}, nil
}

func (d *driver) String() string {
	return Name
}

func (d *driver) Type() api.DriverType {
	return Type
}

func (d *driver) Create(locator *api.VolumeLocator, source *api.Source, spec *api.VolumeSpec) (string, error) {
	volumeID := uuid.New()
	volumeID = strings.TrimSuffix(volumeID, "\n")
	// Create a directory on the Local machine with this UUID.
	if err := os.MkdirAll(filepath.Join(config.VolumeBase, string(volumeID)), 0744); err != nil {
		dlog.Println(err)
		return "", err
	}

	v := common.NewVolume(
		volumeID,
		api.FSType_FS_TYPE_VFS,
		locator,
		source,
		spec,
	)
	v.DevicePath = filepath.Join(config.VolumeBase, volumeID)

	if err := d.CreateVol(v); err != nil {
		return "", err
	}
	return v.Id, d.UpdateVol(v)
}

func (d *driver) Delete(volumeID string) error {

	// Check if volume exists
	_, err := d.GetVol(volumeID)
	if err != nil {
		dlog.Println("Volume not found ", err)
		return err
	}

	// Delete the directory
	os.RemoveAll(filepath.Join(config.VolumeBase, string(volumeID)))

	err = d.DeleteVol(volumeID)
	if err != nil {
		dlog.Println(err)
		return err
	}

	return nil

}

// Mount volume at specified path
// Errors ErrEnoEnt, ErrVolDetached may be returned.
func (d *driver) Mount(volumeID string, mountpath string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		dlog.Println(err)
		return err
	}
	syscall.Unmount(mountpath, 0)
	if err := syscall.Mount(filepath.Join(config.VolumeBase, string(volumeID)), mountpath, string(v.Spec.Format), syscall.MS_BIND, ""); err != nil {
		dlog.Printf("Cannot mount %s at %s because %+v", filepath.Join(config.VolumeBase, string(volumeID)), mountpath, err)
		return err
	}
	v.AttachPath = mountpath
	// TODO(pedge): why ignoring error?
	err = d.UpdateVol(v)
	return nil
}

// Unmount volume at specified path
// Errors ErrEnoEnt, ErrVolDetached may be returned.
func (d *driver) Unmount(volumeID string, mountpath string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return err
	}
	if v.AttachPath == "" {
		return fmt.Errorf("Device %v not mounted", volumeID)
	}
	if err := syscall.Unmount(v.AttachPath, 0); err != nil {
		return err
	}
	v.AttachPath = ""
	// TODO(pedge): why ignoring error?
	err = d.UpdateVol(v)
	return nil
}

// Set update volume with specified parameters.
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

// Stats Not Supported.
func (d *driver) Stats(volumeID string) (*api.Stats, error) {
	return nil, volume.ErrNotSupported
}

// Alerts Not Supported.
func (d *driver) Alerts(volumeID string) (*api.Alerts, error) {
	return nil, volume.ErrNotSupported
}

// Status returns a set of key-value pairs which give low
// level diagnostic status about this driver.
func (d *driver) Status() [][2]string {
	return [][2]string{}
}

// Shutdown and cleanup.
func (d *driver) Shutdown() {
	dlog.Debugf("%s Shutting down", Name)
}
