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

	"github.com/Sirupsen/logrus"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/pborman/uuid"
	"github.com/portworx/kvdb"
)

const (
	Name          = "buse"
	Type          = api.Block
	BuseDBKey     = "OpenStorageBuseKey"
	BuseMountPath = "/var/lib/openstorage/buse/"
)

// Implements the open storage volume interface.
type driver struct {
	*volume.IoNotSupported
	*volume.DefaultEnumerator
	buseDevices map[string]*buseDev
}

// Implements the Device interface.
type buseDev struct {
	file string
	f    *os.File
	nbd  *NBD
}

func (d *buseDev) ReadAt(b []byte, off int64) (n int, err error) {
	return d.f.ReadAt(b, off)
}

func (d *buseDev) WriteAt(b []byte, off int64) (n int, err error) {
	return d.f.WriteAt(b, off)
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
		IoNotSupported:    &volume.IoNotSupported{},
		DefaultEnumerator: volume.NewDefaultEnumerator(Name, kvdb.Instance()),
	}

	inst.buseDevices = make(map[string]*buseDev)

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
		logrus.Println("Could not enumerate Volumes, ", err)
	}

	c, err := cluster.Inst()
	if err != nil {
		logrus.Println("BUSE initializing in single node mode")
	} else {
		logrus.Println("BUSE initializing in clustered mode")
		c.AddEventListener(inst)
	}

	logrus.Println("BUSE initialized and driver mounted at: ", BuseMountPath)
	return inst, nil
}

//
// These functions below implement the volume driver interface.
//

func (d *driver) String() string {
	return Name
}

func (d *driver) Type() api.DriverType {
	return Type
}

// Status diagnostic information
func (d *driver) Status() [][2]string {
	return [][2]string{}
}

func (d *driver) Create(locator api.VolumeLocator, source *api.Source, spec *api.VolumeSpec) (api.VolumeID, error) {
	volumeID := uuid.New()
	volumeID = strings.TrimSuffix(volumeID, "\n")

	if spec.Size == 0 {
		return api.BadVolumeID, fmt.Errorf("Volume size cannot be zero", "buse")
	}

	if spec.Format == "" {
		return api.BadVolumeID, fmt.Errorf("Missing volume format", "buse")
	}

	// Create a file on the local buse path with this UUID.
	buseFile := path.Join(BuseMountPath, string(volumeID))
	f, err := os.Create(buseFile)
	if err != nil {
		logrus.Println(err)
		return api.BadVolumeID, err
	}

	err = f.Truncate(int64(spec.Size))
	if err != nil {
		logrus.Println(err)
		return api.BadVolumeID, err
	}

	bd := &buseDev{
		file: buseFile,
		f:    f}

	nbd := Create(bd, int64(spec.Size))
	bd.nbd = nbd

	logrus.Infof("Connecting to NBD...")
	dev, err := bd.nbd.Connect()
	if err != nil {
		logrus.Println(err)
		return api.BadVolumeID, err
	}

	logrus.Infof("Formatting %s with %v", dev, spec.Format)
	cmd := "/sbin/mkfs." + string(spec.Format)
	o, err := exec.Command(cmd, dev).Output()
	if err != nil {
		logrus.Warnf("Failed to run command %v %v: %v", cmd, dev, o)
		return api.BadVolumeID, err
	}

	logrus.Infof("BUSE mapped NBD device %s (size=%v) to block file %s", dev, spec.Size, buseFile)

	v := &api.Volume{
		ID:         api.VolumeID(volumeID),
		Source:     source,
		Locator:    locator,
		Ctime:      time.Now(),
		Spec:       spec,
		LastScan:   time.Now(),
		Format:     spec.Format,
		State:      api.VolumeAvailable,
		Status:     api.Up,
		DevicePath: dev,
	}

	d.buseDevices[dev] = bd

	err = d.CreateVol(v)
	if err != nil {
		return api.BadVolumeID, err
	}
	return v.ID, err
}

func (d *driver) Delete(volumeID api.VolumeID) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		logrus.Println(err)
		return err
	}

	bd, ok := d.buseDevices[v.DevicePath]
	if !ok {
		err = fmt.Errorf("Cannot locate a BUSE device for %s", v.DevicePath)
		logrus.Println(err)
		return err
	}

	// Clean up buse block file and close the NBD connection.
	os.Remove(bd.file)
	bd.f.Close()
	bd.nbd.Disconnect()

	logrus.Infof("BUSE deleted volume %v at NBD device %s", volumeID, v.DevicePath)

	err = d.DeleteVol(volumeID)
	if err != nil {
		logrus.Println(err)
		return err
	}

	return nil
}

func (d *driver) Mount(volumeID api.VolumeID, mountpath string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Failed to locate volume %q", string(volumeID))
	}
	err = syscall.Mount(v.DevicePath, mountpath, string(v.Spec.Format), 0, "")
	if err != nil {
		logrus.Errorf("Mounting %s on %s failed because of %v", v.DevicePath, mountpath, err)
		return fmt.Errorf("Failed to mount %v at %v: %v", v.DevicePath, mountpath, err)
	}

	logrus.Infof("BUSE mounted NBD device %s at %s", v.DevicePath, mountpath)

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

func (d *driver) Set(volumeID api.VolumeID, locator *api.VolumeLocator, spec *api.VolumeSpec) error {
	if spec != nil {
		return volume.ErrNotSupported
	}
	v, err := d.GetVol(volumeID)
	if err != nil {
		return err
	}
	if locator != nil {
		v.Locator = *locator
	}
	err = d.UpdateVol(v)
	return err
}

func (d *driver) Attach(volumeID api.VolumeID) (string, error) {
	// Nothing to do on attach.
	return path.Join(BuseMountPath, string(volumeID)), nil
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
	logrus.Printf("%s Shutting down", Name)
	syscall.Unmount(BuseMountPath, 0)
}

func (d *driver) ClusterInit(self *api.Node, db *cluster.Database) error {
	return nil
}

func (d *driver) Init(self *api.Node, db *cluster.Database) error {
	return nil
}

func (d *driver) CleanupInit(self *api.Node, db *cluster.Database) error {
	return nil
}

func (d *driver) Join(self *api.Node, db *cluster.Database) error {
	return nil
}

func (d *driver) Add(self *api.Node) error {
	return nil
}

func (d *driver) Remove(self *api.Node) error {
	return nil
}

func (d *driver) Update(self *api.Node) error {
	return nil
}

func (d *driver) Leave(self *api.Node) error {
	return nil
}

func init() {

	// Register ourselves as an openstorage volume driver.
	volume.Register(Name, Init)
}
