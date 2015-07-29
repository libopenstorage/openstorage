package nfs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/libopenstorage/kvdb"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	Name     = "nfs"
	NfsDBKey = "OpenStorageNFSKey"
)

var (
	devMinor int32
)

// This data is persisted in a DB.
type awsVolume struct {
	spec      api.VolumeSpec
	formatted bool
	attached  bool
	mounted   bool
	device    string
	mountpath string
}

// Implements the open storage volume interface.
type nfsProvider struct {
	volume.DefaultBlockDriver
	db        kvdb.Kvdb
	nfsServer string
	mntPath   string
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	uri, ok := params["uri"]
	if !ok {
		return nil, errors.New("No NFS server URI provided")
	}

	fmt.Println("NFS driver initializing with server:", uri)

	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		return nil, err
	}

	inst := &nfsProvider{
		db:        kvdb.Instance(),
		mntPath:   "/mnt/" + string(out),
		nfsServer: uri}

	// Mount the nfs server locally on a unique path.
	err = syscall.Mount(inst.nfsServer, inst.mntPath, "", 0, "")
	if err != nil {
		return nil, err
	}

	fmt.Printf("NFS initialized and driver mounted at %s.", inst.mntPath)
	return inst, nil
}

func (self *nfsProvider) get(volumeID string) (*awsVolume, error) {
	v := &awsVolume{}
	key := NfsDBKey + "/" + volumeID
	_, err := self.db.GetVal(key, v)
	return v, err
}

func (self *nfsProvider) put(volumeID string, v *awsVolume) error {
	key := NfsDBKey + "/" + volumeID
	_, err := self.db.Put(key, v, 0)
	return err
}

func (self *nfsProvider) del(volumeID string) {
	key := NfsDBKey + "/" + volumeID
	self.db.Delete(key)
}

func (self *nfsProvider) String() string {
	return Name
}

func (self *nfsProvider) Create(l api.VolumeLocator, opt *api.CreateOptions, spec *api.VolumeSpec) (api.VolumeID, error) {
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		return "", err
	}

	volumeID := string(out)

	// Create a directory on the NFS server with this UUID.
	err = os.Mkdir(self.mntPath+volumeID, 0744)
	if err != nil {
		return "", err
	}

	// Persist the volume spec.  We use this for all subsequent operations on
	// this volume ID.
	err = self.put(volumeID, &awsVolume{device: self.mntPath + volumeID, spec: *spec})

	return api.VolumeID(volumeID), err
}

func (self *nfsProvider) Inspect(volumeIDs []api.VolumeID) (volume []api.Volume, err error) {
	return nil, nil
}

func (self *nfsProvider) Delete(volumeID api.VolumeID) error {
	v, err := self.get(string(volumeID))
	if err != nil {
		return err
	}

	// Delete the directory on the nfs server.
	err = os.Remove(v.device)
	if err != nil {
		return err
	}

	self.del(string(volumeID))

	return nil
}

func (self *nfsProvider) Snapshot(volumeID api.VolumeID, labels api.Labels) (snap api.SnapID, err error) {
	return "", errors.New("Unsupported")
}

func (self *nfsProvider) SnapDelete(snapID api.SnapID) (err error) {
	return errors.New("Unsupported")
}

func (self *nfsProvider) SnapInspect(snapID api.SnapID) (snap api.VolumeSnap, err error) {
	return api.VolumeSnap{}, errors.New("Unsupported")
}

func (self *nfsProvider) Stats(volumeID api.VolumeID) (stats api.VolumeStats, err error) {
	return api.VolumeStats{}, errors.New("Unsupported")
}

func (self *nfsProvider) Alerts(volumeID api.VolumeID) (stats api.VolumeAlerts, err error) {
	return api.VolumeAlerts{}, errors.New("Unsupported")
}

func (self *nfsProvider) Enumerate(locator api.VolumeLocator, labels api.Labels) (volumes []api.Volume, err error) {
	return nil, errors.New("Unsupported")
}

func (self *nfsProvider) SnapEnumerate(locator api.VolumeLocator, labels api.Labels) (snaps *[]api.SnapID, err error) {
	return nil, errors.New("Unsupported")
}

func (self *nfsProvider) Mount(volumeID api.VolumeID, mountpath string) error {
	v, err := self.get(string(volumeID))
	if err != nil {
		return err
	}

	err = syscall.Mount(v.device, mountpath, string(v.spec.Format), 0, "")
	if err != nil {
		return err
	}

	v.mountpath = mountpath
	v.mounted = true
	err = self.put(string(volumeID), v)

	return err
}

func (self *nfsProvider) Unmount(volumeID api.VolumeID, mountpath string) error {
	v, err := self.get(string(volumeID))
	if err != nil {
		return err
	}

	err = syscall.Unmount(v.mountpath, 0)
	if err != nil {
		return err
	}

	v.mountpath = ""
	v.mounted = false
	err = self.put(string(volumeID), v)

	return err
}

func (self *nfsProvider) Shutdown() {
	fmt.Printf("%s Shutting down", Name)
}

func init() {
	// Register ourselves as an openstorage volume driver.
	volume.Register(Name, volume.TypeFileDriver, Init)
}
