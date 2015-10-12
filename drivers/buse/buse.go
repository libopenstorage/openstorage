package buse

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"

	"github.com/portworx/kvdb"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	Name          = "buse"
	Type          = volume.Block
	BuseDBKey     = "OpenStorageBuseKey"
	BuseMountPath = "/var/lib/openstorage/buse/"
	NbdMax        = 16
)

// Implements the open storage volume interface.
type driver struct {
	*volume.DefaultEnumerator
	nbdSlots [NbdMax]bool
}

func allocNBD(sz uint) (string, error) {
	return "/dev/nbd0", nil
}

func freeNBD(dev string) {
	// XXX FIXME dellocate NBD device.
}

func copyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}

	}

	return
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	inst := &driver{
		DefaultEnumerator: volume.NewDefaultEnumerator(Name, kvdb.Instance()),
	}

	err := os.MkdirAll(BuseMountPath, 0744)
	if err != nil {
		return nil, err
	}

	volumeInfo, err := inst.DefaultEnumerator.Enumerate(
		api.VolumeLocator{},
		nil)
	if err == nil {
		for _, info := range volumeInfo {
			if info.Status == "" {
				info.Status = api.Up
				inst.UpdateVol(&info)
			}
		}
	} else {
		log.Println("Could not enumerate Volumes, ", err)
	}

	log.Println("BUSE initialized and driver mounted at: ", BuseMountPath)
	return inst, nil
}

func (d *driver) String() string {
	return Name
}

func (d *driver) Type() volume.DriverType {
	return Type
}

// Status diagnostic information
func (d *driver) Status() [][2]string {
	return [][2]string{}
}

func (d *driver) Create(locator api.VolumeLocator, source *api.Source, spec *api.VolumeSpec) (api.VolumeID, error) {
	volumeID := uuid.New()
	volumeID = strings.TrimSuffix(volumeID, "\n")

	// Create a file on the local buse path with this UUID.
	volPath := path.Join(BuseMountPath, volumeID)
	err := os.MkdirAll(volPath, 0744)
	if err != nil {
		log.Println(err)
		return api.BadVolumeID, err
	}

	f, err := os.Create(path.Join(BuseMountPath, string(volumeID)))
	if err != nil {
		log.Println(err)
		return api.BadVolumeID, err
	}
	defer f.Close()

	err = f.Truncate(int64(spec.Size))
	if err != nil {
		log.Println(err)
		return api.BadVolumeID, err
	}

	dev, err := allocNBD(uint(spec.Size))
	if err != nil {
		log.Println(err)
		return api.BadVolumeID, err
	}

	v := &api.Volume{
		ID:         api.VolumeID(volumeID),
		Source:     source,
		Locator:    locator,
		Ctime:      time.Now(),
		Spec:       spec,
		LastScan:   time.Now(),
		Format:     "buse",
		State:      api.VolumeAvailable,
		Status:     api.Up,
		DevicePath: dev,
	}

	err = d.CreateVol(v)
	if err != nil {
		return api.BadVolumeID, err
	}
	return v.ID, err
}

func (d *driver) Delete(volumeID api.VolumeID) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		log.Println(err)
		return err
	}

	// Delete the block file on the local buse path.
	os.Remove(path.Join(BuseMountPath, string(volumeID)))

	freeNBD(v.DevicePath)

	err = d.DeleteVol(volumeID)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (d *driver) Mount(volumeID api.VolumeID, mountpath string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Failed to locate volume %q", string(volumeID))
	}
	err = syscall.Mount(v.DevicePath, mountpath, string(v.Spec.Format), syscall.MS_BIND, "")
	if err != nil {
		return fmt.Errorf("Faield to mount %v at %v: %v", v.DevicePath, mountpath, err)
	}

	v.AttachPath = mountpath
	err = d.UpdateVol(v)

	return nil
}

func (d *driver) Unmount(volumeID api.VolumeID, mountpath string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return err
	}
	if v.AttachPath == "" {
		return fmt.Errorf("Device %v not mounted", volumeID)
	}
	err = syscall.Unmount(v.AttachPath, 0)
	if err != nil {
		return err
	}
	v.AttachPath = ""
	err = d.UpdateVol(v)
	return err
}

func (d *driver) Snapshot(volumeID api.VolumeID, readonly bool, locator api.VolumeLocator) (api.VolumeID, error) {
	volIDs := make([]api.VolumeID, 1)
	volIDs[0] = volumeID
	vols, err := d.Inspect(volIDs)
	if err != nil {
		return api.BadVolumeID, nil
	}

	source := &api.Source{Parent: volumeID}
	newVolumeID, err := d.Create(locator, source, vols[0].Spec)
	if err != nil {
		return api.BadVolumeID, nil
	}

	// BUSE does not support snapshots, so just copy the block files.
	err = copyFile(BuseMountPath+string(volumeID), BuseMountPath+string(newVolumeID))
	if err != nil {
		d.Delete(newVolumeID)
		return api.BadVolumeID, nil
	}

	return newVolumeID, nil
}

func (d *driver) Attach(volumeID api.VolumeID) (string, error) {
	// Nothing to do on attach.
	return path.Join(BuseMountPath, string(volumeID)), nil
}

func (d *driver) Format(volumeID api.VolumeID) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Failed to locate volume %q", string(volumeID))
	}

	cmd := "/sbin/mkfs." + string(v.Spec.Format)
	o, err := exec.Command(cmd, v.DevicePath).Output()
	if err != nil {
		log.Warnf("Failed to run command %v %v: %v", cmd, v.DevicePath, o)
		return err
	}
	v.Format = v.Spec.Format
	err = d.UpdateVol(v)

	return nil
}

func (d *driver) Detach(volumeID api.VolumeID) error {
	// Nothing to do on detach.
	return nil
}

func (d *driver) Stats(volumeID api.VolumeID) (api.Stats, error) {
	return api.Stats{}, volume.ErrNotSupported
}

func (d *driver) Alerts(volumeID api.VolumeID) (api.Alerts, error) {
	return api.Alerts{}, volume.ErrNotSupported
}

func (d *driver) Shutdown() {
	log.Printf("%s Shutting down", Name)
	syscall.Unmount(BuseMountPath, 0)
}

func init() {
	// Register ourselves as an openstorage volume driver.
	volume.Register(Name, Init)
}
