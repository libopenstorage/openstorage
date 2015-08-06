package nfs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/libopenstorage/kvdb"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	Name         = "nfs"
	NfsDBKey     = "OpenStorageNFSKey"
	nfsMountPath = "/var/lib/openstorage/nfs/"
)

// Implements the open storage volume interface.
type driver struct {
	*volume.DefaultBlockDriver
	*volume.DefaultEnumerator
	nfsServer string
	nfsPath   string
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	server, ok := params["server"]
	if !ok {
		return nil, errors.New("No NFS server provided")
	}

	path, ok := params["path"]
	if !ok {
		return nil, errors.New("No NFS path provided")
	}

	log.Printf("NFS driver initializing with %s:%s ", server, path)

	inst := &driver{
		DefaultEnumerator: volume.NewDefaultEnumerator(Name, kvdb.Instance()),
		nfsServer:         server,
		nfsPath:           path}

	err := os.MkdirAll(nfsMountPath, 0744)
	if err != nil {
		return nil, err
	}

	// Mount the nfs server locally on a unique path.
	syscall.Unmount(nfsMountPath, 0)
	err = syscall.Mount(":"+inst.nfsPath, nfsMountPath, "nfs", 0, "nolock,addr="+inst.nfsServer)
	if err != nil {
		log.Printf("Unable to mount %s at %s.\n", inst.nfsServer, nfsMountPath)
		return nil, err
	}

	log.Println("NFS initialized and driver mounted at: ", nfsMountPath)
	return inst, nil
}

func (d *driver) String() string {
	return Name
}

// Status diagnostic information
func (d *driver) Status() [][2]string {
	return [][2]string{}
}

func (d *driver) Create(locator api.VolumeLocator, opt *api.CreateOptions, spec *api.VolumeSpec) (api.VolumeID, error) {
	// Validate options.
	if spec.Format != "nfs" && spec.Format != "" {
		return api.BadVolumeID, errors.New("Unsupported filesystem format: " + string(spec.Format))
	}

	if spec.BlockSize != 0 {
		log.Println("NFS driver will ignore the blocksize option.")
	}

	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Println(err)
		return "", err
	}
	volumeID := string(out)
	volumeID = strings.TrimSuffix(volumeID, "\n")

	// Create a directory on the NFS server with this UUID.
	err = os.MkdirAll(nfsMountPath+volumeID, 0744)
	if err != nil {
		log.Println(err)
		return api.BadVolumeID, err
	}

	v := &api.Volume{
		ID:         api.VolumeID(volumeID),
		Locator:    locator,
		Ctime:      time.Now(),
		Spec:       spec,
		LastScan:   time.Now(),
		Format:     "nfs",
		State:      api.VolumeAvailable,
		DevicePath: nfsMountPath + volumeID,
	}

	err = d.CreateVol(v)
	if err != nil {
		return api.BadVolumeID, err
	}

	err = d.UpdateVol(v)

	return v.ID, err
}

func (d *driver) Delete(volumeID api.VolumeID) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		log.Println(err)
		return err
	}

	// Delete the directory on the nfs server.
	os.Remove(v.DevicePath)

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
		log.Println(err)
		return err
	}

	syscall.Unmount(mountpath, 0)
	err = syscall.Mount(v.DevicePath, mountpath, string(v.Spec.Format), syscall.MS_BIND, "")
	if err != nil {
		log.Printf("Cannot mount %s at %s because %+v", v.DevicePath, mountpath, err)
		return err
	}

	v.AttachPath = mountpath
	err = d.UpdateVol(v)

	return err
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

func (d *driver) Snapshot(volumeID api.VolumeID, labels api.Labels) (api.SnapID, error) {
	return "", volume.ErrNotSupported
}

func (d *driver) SnapDelete(snapID api.SnapID) error {
	return volume.ErrNotSupported
}

func (d *driver) Stats(volumeID api.VolumeID) (api.VolumeStats, error) {
	return api.VolumeStats{}, volume.ErrNotSupported
}

func (d *driver) Alerts(volumeID api.VolumeID) (api.VolumeAlerts, error) {
	return api.VolumeAlerts{}, volume.ErrNotSupported
}

func (d *driver) Shutdown() {
	log.Printf("%s Shutting down", Name)
	syscall.Unmount(nfsMountPath, 0)
}

func init() {
	// Register ourselves as an openstorage volume driver.
	volume.Register(Name, volume.File, Init)
}
