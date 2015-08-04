package nfs

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"

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

var (
	devMinor int32
)

// This data is persisted in a DB.
type nfsVolume struct {
	Spec      api.VolumeSpec
	Locator   api.VolumeLocator
	Id        api.VolumeID
	Formatted bool
	Attached  bool
	Mounted   bool
	Device    string
	Mountpath string
}

// Implements the open storage volume interface.
type nfsDriver struct {
	volume.DefaultBlockDriver
	db        kvdb.Kvdb
	nfsServer string
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	uri, ok := params["uri"]
	if !ok {
		return nil, errors.New("No NFS server URI provided")
	}

	log.Println("NFS driver initializing with server: ", uri)

	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		return nil, err
	}
	uuid := string(out)
	uuid = strings.TrimSuffix(uuid, "\n")

	inst := &nfsDriver{
		db:        kvdb.Instance(),
		nfsServer: uri}

	err = os.MkdirAll(nfsMountPath, 0744)
	if err != nil {
		return nil, err
	}

	// Mount the nfs server locally on a unique path.
	err = syscall.Mount(inst.nfsServer, nfsMountPath, "tmpfs", 0, "mode=0700,uid=65534")
	if err != nil {
		return nil, err
	}

	log.Println("NFS initialized and driver mounted at: ", nfsMountPath)
	return inst, nil
}

func (d *nfsDriver) get(volumeID string) (*nfsVolume, error) {
	v := &nfsVolume{}
	key := NfsDBKey + "/" + volumeID
	_, err := d.db.GetVal(key, v)
	return v, err
}

func (d *nfsDriver) enumerate() ([]*nfsVolume, error) {
	key := NfsDBKey
	kvps, err := d.db.Enumerate(key)
	if err != nil {
		return nil, err
	}

	i := 0
	vs := make([]*nfsVolume, len(kvps))
	for _, kvp := range kvps {
		v := &nfsVolume{}
		err = json.Unmarshal(kvp.Value, v)
		if err != nil {
			return nil, err
		}
		vs[i] = v
		i++
	}

	return vs, err
}

func (d *nfsDriver) put(volumeID string, v *nfsVolume) error {
	key := NfsDBKey + "/" + volumeID
	_, err := d.db.Put(key, v, 0)
	return err
}

func (d *nfsDriver) del(volumeID string) {
	key := NfsDBKey + "/" + volumeID
	d.db.Delete(key)
}

func (d *nfsDriver) String() string {
	return Name
}

func (d *nfsDriver) Create(locator api.VolumeLocator, opt *api.CreateOptions, spec *api.VolumeSpec) (api.VolumeID, error) {
	// Validate options.
	if spec.Format != "" && spec.Format != "nfs" {
		return "", errors.New("Unsupported filesystem format: " + string(spec.Format))
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
		return "", err
	}

	// Persist the volume spec.  We use this for all subsequent operations on
	// this volume ID.
	err = d.put(volumeID, &nfsVolume{Id: api.VolumeID(volumeID), Device: nfsMountPath + volumeID, Spec: *spec, Locator: locator})

	return api.VolumeID(volumeID), err
}

func (d *nfsDriver) Inspect(volumeIDs []api.VolumeID) ([]api.Volume, error) {
	volumes := make([]api.Volume, len(volumeIDs))
	for i, id := range volumeIDs {
		v, err := d.get(string(id))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		volumes[i] = api.Volume{
			ID:   id,
			Spec: &v.Spec}
	}

	return volumes, nil
}

func (d *nfsDriver) Delete(volumeID api.VolumeID) error {
	v, err := d.get(string(volumeID))
	if err != nil {
		log.Println(err)
		return err
	}

	// Delete the directory on the nfs server.
	err = os.Remove(v.Device)
	if err != nil {
		log.Println(err)
		return err
	}

	d.del(string(volumeID))

	return nil
}

func (d *nfsDriver) Snapshot(volumeID api.VolumeID, labels api.Labels) (api.SnapID, error) {
	return "", volume.ErrNotSupported
}

func (d *nfsDriver) SnapDelete(snapID api.SnapID) error {
	return volume.ErrNotSupported
}

func (d *nfsDriver) SnapInspect(snapID []api.SnapID) ([]api.VolumeSnap, error) {
	return []api.VolumeSnap{}, volume.ErrNotSupported
}

func (d *nfsDriver) Stats(volumeID api.VolumeID) (api.VolumeStats, error) {
	return api.VolumeStats{}, volume.ErrNotSupported
}

func (d *nfsDriver) Alerts(volumeID api.VolumeID) (api.VolumeAlerts, error) {
	return api.VolumeAlerts{}, volume.ErrNotSupported
}

func (d *nfsDriver) Enumerate(locator api.VolumeLocator, labels api.Labels) ([]api.Volume, error) {
	vs, err := d.enumerate()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	volumes := make([]api.Volume, 0)
	for _, v := range vs {
		if locator.Name != "" {
			if v.Locator.Name != locator.Name {
				continue
			}
		}

		vol := api.Volume{
			ID:   v.Id,
			Spec: &v.Spec}
		volumes = append(volumes, vol)
	}

	return volumes, nil
}

func (d *nfsDriver) SnapEnumerate(locator api.VolumeLocator, labels api.Labels) ([]api.VolumeSnap, error) {
	return nil, volume.ErrNotSupported
}

func (d *nfsDriver) Mount(volumeID api.VolumeID, mountpath string) error {
	v, err := d.get(string(volumeID))
	if err != nil {
		log.Println(err)
		return err
	}

	syscall.Unmount(mountpath, 0)
	err = syscall.Mount(v.Device, mountpath, string(v.Spec.Format), syscall.MS_BIND, "")
	if err != nil {
		log.Printf("Cannot mount %s at %s because %+v", v.Device, mountpath, err)
		return err
	}

	v.Mountpath = mountpath
	v.Mounted = true
	err = d.put(string(volumeID), v)

	return err
}

func (d *nfsDriver) Unmount(volumeID api.VolumeID, mountpath string) error {
	v, err := d.get(string(volumeID))
	if err != nil {
		log.Println(err)
		return err
	}

	err = syscall.Unmount(v.Mountpath, 0)
	if err != nil {
		log.Println(err)
		return err
	}

	v.Mountpath = ""
	v.Mounted = false
	err = d.put(string(volumeID), v)

	return err
}

func (d *nfsDriver) Shutdown() {
	log.Printf("%s Shutting down", Name)
	syscall.Unmount(nfsMountPath, 0)
}

func init() {
	// Register ourselves as an openstorage volume driver.
	volume.Register(Name, volume.File, Init)
}
