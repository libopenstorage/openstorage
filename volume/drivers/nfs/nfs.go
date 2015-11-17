package nfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"

	"github.com/portworx/kvdb"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/pkg/mount"
	"github.com/libopenstorage/openstorage/pkg/seed"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	Name         = "nfs"
	Type         = api.File
	NfsDBKey     = "OpenStorageNFSKey"
	nfsMountPath = "/var/lib/openstorage/nfs/"
	nfsBlockFile = ".blockdevice"
)

// Implements the open storage volume interface.
type driver struct {
	*volume.DefaultEnumerator
	nfsServer string
	nfsPath   string
	mounter   mount.Manager
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

func copyDir(source string, dest string) (err error) {
	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir

	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		sourcefilepointer := source + "/" + obj.Name()

		destinationfilepointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			// create sub-directories - recursively
			err = copyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy
			err = copyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	return
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	path, ok := params["path"]
	if !ok {
		return nil, errors.New("No NFS path provided")
	}

	server, ok := params["server"]
	if !ok {
		log.Printf("No NFS server provided, will attempt to bind mount %s", path)
	} else {
		log.Printf("NFS driver initializing with %s:%s ", server, path)
	}

	// Create a mount manager for this NFS server. Blank sever is OK.
	mounter, err := mount.New(mount.NFSMount, server)
	if err != nil {
		log.Warnf("Failed to create mount manager for server: %v (%v)", server, err)
		return nil, err
	}

	inst := &driver{
		DefaultEnumerator: volume.NewDefaultEnumerator(Name, kvdb.Instance()),
		nfsServer:         server,
		nfsPath:           path,
		mounter:           mounter,
	}

	err = os.MkdirAll(nfsMountPath, 0744)
	if err != nil {
		return nil, err
	}
	src := inst.nfsPath
	if server != "" {
		src = ":" + inst.nfsPath
	}

	// If src is already mounted at dest, leave it be.
	mountExists, err := mounter.Exists(src, nfsMountPath)
	if !mountExists {
		// Mount the nfs server locally on a unique path.
		syscall.Unmount(nfsMountPath, 0)
		if server != "" {
			err = syscall.Mount(src, nfsMountPath, "nfs", 0, "nolock,addr="+inst.nfsServer)
		} else {
			err = syscall.Mount(src, nfsMountPath, "", syscall.MS_BIND, "")
		}
		if err != nil {
			log.Printf("Unable to mount %s:%s at %s (%+v)", inst.nfsServer, inst.nfsPath, nfsMountPath, err)
			return nil, err
		}
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

	log.Println("NFS initialized and driver mounted at: ", nfsMountPath)
	return inst, nil
}

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

//
// These functions below implement the graph driver interface.
//

// Create a new, empty, filesystem layer with the specified ID and Parent. Parent may be an empty string,
// which would indicate that there is no parent layer.
func (d *driver) GraphDriverCreate(id, parent string) error {
	return nil
}

// Remove the filesystem layer with this given ID.
func (d *driver) GraphDriverRemove(id string) error {
	return nil
}

// Get the mountpoint for the layered filesystem referred to by the given ID.
func (d *driver) GraphDriverGet(id, mountLabel string) (string, error) {
	return "", nil
}

// Release the system resources for the specified ID,
// such as unmounting the filesystem layer.
func (d *driver) GraphDriverRelease(id string) error {
	return nil
}

// Determine if a filesystem layer with the specified ID exists.
func (d *driver) GraphDriverExists(id string) bool {
	return false
}

// Get an archive of the changes between the filesystem layers specified by the ID
// and Parent. Parent may be an empty string, in which case there is no parent.
func (d *driver) GraphDriverDiff(id, parent string) io.Writer {
	return nil
}

// Get a list of changes between the filesystem layers specified by the ID and Parent.
// Parent may be an empty string, in which case there is no parent.
func (d *driver) GraphDriverChanges(id, parent string) ([]api.GraphDriverChanges, error) {
	changes := make([]api.GraphDriverChanges, 0)
	return changes, nil
}

// Extract the changeset from the given diff into the layer with the specified ID and Parent
func (d *driver) GraphDriverApplyDiff(id, parent string, diff io.Reader) (int, error) {
	return 0, nil
}

// Calculate the changes between the specified ID
func (d *driver) GraphDriverDiffSize(id, parent string) (int, error) {
	return 0, nil
}

//
// These functions below implement the volume driver interface.
//

func (d *driver) Create(locator api.VolumeLocator, source *api.Source, spec *api.VolumeSpec) (api.VolumeID, error) {
	volumeID := uuid.New()
	volumeID = strings.TrimSuffix(volumeID, "\n")

	// Create a directory on the NFS server with this UUID.
	volPath := path.Join(nfsMountPath, volumeID)
	err := os.MkdirAll(volPath, 0744)
	if err != nil {
		log.Println(err)
		return api.BadVolumeID, err
	}
	if source != nil {
		if len(source.Seed) != 0 {
			seed, err := seed.New(source.Seed, spec.ConfigLabels)
			if err != nil {
				log.Warnf("Failed to initailize seed from %q : %v",
					source.Seed, err)
				return api.BadVolumeID, err
			}
			err = seed.Load(path.Join(volPath, config.DataDir))
			if err != nil {
				log.Warnf("Failed to  seed from %q to %q: %v",
					source.Seed, nfsMountPath, err)
				return api.BadVolumeID, err
			}
		}
	}

	f, err := os.Create(path.Join(nfsMountPath, string(volumeID)+nfsBlockFile))
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

	v := &api.Volume{
		ID:         api.VolumeID(volumeID),
		Source:     source,
		Locator:    locator,
		Ctime:      time.Now(),
		Spec:       spec,
		LastScan:   time.Now(),
		Format:     "nfs",
		State:      api.VolumeAvailable,
		Status:     api.Up,
		DevicePath: path.Join(nfsMountPath, string(volumeID)+nfsBlockFile),
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

	// Delete the simulated block volume
	os.Remove(v.DevicePath)

	// Delete the directory on the nfs server.
	os.RemoveAll(path.Join(nfsMountPath, string(volumeID)))

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

	srcPath := path.Join(":", d.nfsPath, string(volumeID))
	mountExists, err := d.mounter.Exists(srcPath, mountpath)
	if !mountExists {
		syscall.Unmount(mountpath, 0)
		err = syscall.Mount(path.Join(nfsMountPath, string(volumeID)), mountpath, string(v.Spec.Format), syscall.MS_BIND, "")
		if err != nil {
			log.Printf("Cannot mount %s at %s because %+v",
				path.Join(nfsMountPath, string(volumeID)), mountpath, err)
			return err
		}
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

	// NFS does not support snapshots, so just copy the files.
	err = copyDir(nfsMountPath+string(volumeID), nfsMountPath+string(newVolumeID))
	if err != nil {
		d.Delete(newVolumeID)
		return api.BadVolumeID, nil
	}

	return newVolumeID, nil
}

func (d *driver) Attach(volumeID api.VolumeID) (string, error) {
	return path.Join(nfsMountPath, string(volumeID)+nfsBlockFile), nil
}

func (d *driver) Detach(volumeID api.VolumeID) error {
	return nil
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

func (d *driver) Stats(volumeID api.VolumeID) (api.Stats, error) {
	return api.Stats{}, volume.ErrNotSupported
}

func (d *driver) Alerts(volumeID api.VolumeID) (api.Alerts, error) {
	return api.Alerts{}, volume.ErrNotSupported
}

func (d *driver) Shutdown() {
	log.Printf("%s Shutting down", Name)
	syscall.Unmount(nfsMountPath, 0)
}

func init() {
	// Register ourselves as an openstorage volume driver.
	volume.Register(Name, Init)
}
