package nfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	losetup "gopkg.in/freddierice/go-losetup.v1"

	"math/rand"
	"strings"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/pkg/mount"
	"github.com/libopenstorage/openstorage/pkg/seed"
	"github.com/libopenstorage/openstorage/pkg/util"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers/common"
	"github.com/pborman/uuid"
	"github.com/portworx/kvdb"
)

const (
	Name         = "nfs"
	NfsDBKey     = "OpenStorageNFSKey"
	nfsMountPath = "/var/lib/openstorage/nfs/"
	nfsBlockFile = ".blockdevice"

	// Set to block, but it will handle size 0 as file based
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
	volume.CloudMigrateDriver
	volume.FilesystemTrimDriver
	volume.FilesystemCheckDriver
	nfsServers []string
	nfsPath    string
	mounter    mount.Manager
}

func Init(params map[string]string) (volume.VolumeDriver, error) {
	path, ok := params["path"]
	if !ok {
		return nil, errors.New("No NFS path provided")
	}
	server, ok := params["server"]
	if !ok {
		logrus.Printf("No NFS server provided, will attempt to bind mount %s", path)
	} else {
		logrus.Printf("NFS driver initializing with %s:%s ", server, path)
	}
	//support more than one server using CSV
	//TB-FIXME: modify driver params flow to support map[string]struct/array
	servers := strings.Split(server, ",")

	// Create a mount manager for this NFS server. Blank sever is OK.
	mounter, err := mount.New(mount.NFSMount, nil, servers, nil, []string{}, "")
	if err != nil {
		logrus.Warnf("Failed to create mount manager for server: %v (%v)", server, err)
		return nil, err
	}

	inst := &driver{
		IODriver:              volume.IONotSupported,
		StoreEnumerator:       common.NewDefaultStoreEnumerator(Name, kvdb.Instance()),
		StatsDriver:           volume.StatsNotSupported,
		QuiesceDriver:         volume.QuiesceNotSupported,
		nfsServers:            servers,
		CredsDriver:           volume.CredsNotSupported,
		nfsPath:               path,
		mounter:               mounter,
		CloudBackupDriver:     volume.CloudBackupNotSupported,
		CloudMigrateDriver:    volume.CloudMigrateNotSupported,
		FilesystemTrimDriver:  volume.FilesystemTrimNotSupported,
		FilesystemCheckDriver: volume.FilesystemCheckNotSupported,
	}

	//make directory for each nfs server
	for _, v := range servers {
		logrus.Infof("Calling mkdirAll: %s", nfsMountPath+v)
		if err := os.MkdirAll(nfsMountPath+v, 0755); err != nil {
			return nil, err
		}
	}

	src := inst.nfsPath
	if server != "" {
		src = ":" + inst.nfsPath
	}

	//mount each nfs server
	for _, v := range inst.nfsServers {
		nfsServer := src
		nfsMountPoint := nfsMountPath + v
		inst.mounter.Reload(nfsServer)

		// ignore the error and let it retry
		mounted, err := inst.mounter.Exists(nfsServer, nfsMountPoint)
		if err != nil && err != mount.ErrEnoent {
			return nil, fmt.Errorf("Unable to determine if mounted nfs exists: %v", err)
		}
		if !mounted {
			if server != "" {
				/*
					err = inst.mounter.Mount(
						0,
						nfsServer,
						nfsMountPoint,
						"nfs",
						0,
						"user_xattr,nolock,addr="+v,
						0,
						nil)
				*/
				err = syscall.Mount(
					src,
					nfsMountPath+v,
					"nfs",
					0,
					"nolock,addr="+v,
				)

			} else {
				err = syscall.Mount(src, nfsMountPath+v, "", syscall.MS_BIND, "")
				/*
					err = inst.mounter.Mount(
						0,
						nfsServer,
						nfsMountPoint,
						"",
						syscall.MS_BIND,
						"user_xattr",
						0,
						nil)
				*/
			}
			if err != nil {
				logrus.Errorf("Unable to mount %s:%s at %s (%+v)",
					v, inst.nfsPath, nfsMountPath+v, err)
				return nil, err
			} else {
				logrus.Infof("NFS: %s mounted", nfsMountPath+v)
			}
		} else {
			logrus.Infof("NFS: %s already mounted", nfsMountPath+v)
		}
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

	logrus.Println("NFS initialized and driver mounted at: ", nfsMountPath)
	return inst, nil
}

func (d *driver) Name() string {
	return Name
}

func (d *driver) Type() api.DriverType {
	return Type
}

func (d *driver) Version() (*api.StorageVersion, error) {
	return &api.StorageVersion{
		Driver:  d.Name(),
		Version: "1.0.0",
	}, nil
}

// Status diagnostic information
func (d *driver) Status() [][2]string {
	return [][2]string{}
}

//
//Utility functions
//
func (d *driver) getNewVolumeServer() (string, error) {
	//randomly select one
	if d.nfsServers != nil && len(d.nfsServers) > 0 {
		return d.nfsServers[rand.Intn(len(d.nfsServers))], nil
	}

	return "", errors.New("No NFS servers found")
}

//get nfsPath for specified volume
func (d *driver) getNFSPath(v *api.Volume) (string, error) {
	locator := v.GetLocator()
	server, ok := locator.VolumeLabels["server"]
	if !ok {
		logrus.Warnf("No server label found on volume")
		return "", fmt.Errorf("No server label found on volume: " + v.Id)
	}

	return path.Join(nfsMountPath, server), nil
}

//get nfsPath for specified volume
func (d *driver) getNFSPathById(volumeID string) (string, error) {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return "", err
	}

	return d.getNFSPath(v)
}

//get nfsPath plus volume name for specified volume
func (d *driver) getNFSVolumePath(v *api.Volume) (string, error) {
	parentPath, err := d.getNFSPath(v)
	if err != nil {
		return "", err
	}

	return path.Join(parentPath, v.Id), nil
}

//get nfsPath plus volume name for specified volume
func (d *driver) getNFSVolumePathById(volumeID string) (string, error) {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return "", err
	}

	return d.getNFSVolumePath(v)
}

//append unix time to volumeID
func (d *driver) getNewSnapVolName(volumeID string) string {
	return volumeID + "-" + strconv.FormatUint(uint64(time.Now().Unix()), 10)
}

//
// These functions below implement the volume driver interface.
//

func (d *driver) Create(
	ctx context.Context,
	locator *api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec) (string, error) {

	if len(locator.Name) == 0 {
		return "", fmt.Errorf("volume name cannot be empty")
	}

	if hasSpaces := strings.Contains(locator.Name, " "); hasSpaces {
		return "", fmt.Errorf("volume name cannot contain space characters")
	}

	volumeID := strings.TrimSuffix(uuid.New(), "\n")

	if _, err := d.GetVol(volumeID); err == nil {
		return "", fmt.Errorf("volume with that id already exists")
	}

	//snapshot passes nil volumelabels
	if locator.VolumeLabels == nil {
		locator.VolumeLabels = make(map[string]string)
	}

	//check if user passed server as option
	labels := locator.GetVolumeLabels()
	_, ok := labels["server"]
	if !ok {
		server, err := d.getNewVolumeServer()
		if err != nil {
			logrus.Infof("no nfs servers found...")
			return "", err
		} else {
			logrus.Infof("Assigning random nfs server: %s to volume: %s", server, volumeID)
		}

		labels["server"] = server
	}
	volPathParent := path.Join(nfsMountPath, labels["server"])
	volPath := path.Join(volPathParent, volumeID)

	// Setup volume object
	if source != nil {
		if len(source.Seed) != 0 {
			seed, err := seed.New(source.Seed, locator.VolumeLabels)
			if err != nil {
				logrus.Warnf("Failed to initailize seed from %q : %v",
					source.Seed, err)
				return "", err
			}
			err = seed.Load(path.Join(volPath, config.DataDir))
			if err != nil {
				logrus.Warnf("Failed to  seed from %q to %q: %v",
					source.Seed, volPathParent, err)
				return "", err
			}
		}
	}

	// Create volume
	var v *api.Volume
	if d.isShared(spec) {
		// File based
		v = common.NewVolume(
			volumeID,
			api.FSType_FS_TYPE_NFS,
			locator,
			source,
			spec,
		)

		// Create a directory on the NFS server with this UUID.
		err := os.MkdirAll(volPath, 0744)
		if err != nil {
			logrus.Println(err)
			return "", err
		}

		// Setup volume
		v.State = api.VolumeState_VOLUME_STATE_PENDING
		if err := d.CreateVol(v); err != nil {
			return "", err
		}

		// Check for cloning
		if source != nil && len(source.GetParent()) != 0 {
			// Need to clone
			if err := d.clone(volumeID, source.GetParent()); err != nil {
				d.Delete(ctx, v.GetId())
				return "", err
			}
		}

		// Set to ready
		v.State = api.VolumeState_VOLUME_STATE_AVAILABLE
		if err := d.UpdateVol(v); err != nil {
			d.Delete(ctx, v.GetId())
			logrus.Errorf("Failed to update volume %s to ready state: %v", volumeID, err)
			return "", err
		}
	} else {
		// Block volume
		if spec.GetSize() == 0 {
			return "", fmt.Errorf("Cannot have size of zero on block volume")
		}
		v = common.NewVolume(
			volumeID,
			spec.GetFormat(),
			locator,
			source,
			spec,
		)

		// Set as pending until the volume is ready
		v.State = api.VolumeState_VOLUME_STATE_PENDING
		if err := d.CreateVol(v); err != nil {
			return "", err
		}

		// Check if this from as a clone
		if source != nil && len(source.GetParent()) != 0 {
			// Need to clone
			if err := d.clone(volumeID, source.GetParent()); err != nil {
				d.Delete(ctx, v.GetId())
				return "", err
			}
		} else {
			// This is a new volume
			blockFile := path.Join(volPathParent, volumeID+nfsBlockFile)
			f, err := os.Create(blockFile)
			if err != nil {
				logrus.Errorf("Unable to create block file %s: %v", blockFile, err)
				return "", err
			}
			defer f.Close()

			// Create sparse file
			if err := f.Truncate(int64(spec.Size)); err != nil {
				logrus.Println(err)
				return "", err
			}

			// Format
			if spec.GetFormat() != api.FSType_FS_TYPE_NONE {
				dev, err := losetup.Attach(blockFile, 0, false)
				if err != nil {
					return "", err
				}
				defer func() {
					if err := dev.Detach(); err != nil {
						logrus.Errorf("Failed to detach %s", dev)
					}
				}()
				logrus.Infof("Formatting %s with %v", dev, spec.Format)

				// Get mkfs
				cmd, err := mkfsFormatTypeCmd(spec.Format.SimpleString())
				if err != nil {
					return "", err
				}
				o, err := exec.Command(cmd, dev.Path()).Output()
				if err != nil {
					logrus.Warnf("Failed to run command %s %s: %v\nOutput: %s", cmd, dev.Path(), err, string(o))
					return "", err
				}
			}
		}

		// Set to ready
		v.State = api.VolumeState_VOLUME_STATE_AVAILABLE
		if err := d.UpdateVol(v); err != nil {
			d.Delete(ctx, v.GetId())
			logrus.Errorf("Failed to update volume %s to ready state: %v", volumeID, err)
			return "", err
		}
	}

	return v.Id, nil
}

func (d *driver) Delete(ctx context.Context, volumeID string) (e error) {
	defer func() {
		if e != nil {
			logrus.Errorf("Delete of %s failed: %v", volumeID, e)
		} else {
			logrus.Infof("Volume %s deleted", volumeID)
		}
	}()

	v, err := d.GetVol(volumeID)
	if err != nil {
		logrus.Println(err)
		return err
	}

	if v.GetState() == api.VolumeState_VOLUME_STATE_ATTACHED {
		return fmt.Errorf("Volume is still attached and cannot be deleted")
	}

	nfsVolPath, err := d.getNFSVolumePath(v)
	if err != nil {
		return err
	}

	// Delete the simulated block volume
	if d.isShared(v.GetSpec()) {
		// Delete the directory on the nfs server.
		if err := os.RemoveAll(nfsVolPath); err != nil {
			return fmt.Errorf("Failed to remove %s: %v", nfsVolPath, err)
		}
	} else {
		// Delete the block device
		if err := os.Remove(nfsVolPath + nfsBlockFile); err != nil {
			return fmt.Errorf("Failed to remove %s: %v", nfsVolPath+nfsBlockFile, err)
		}
	}

	err = d.DeleteVol(volumeID)
	if err != nil {
		logrus.Println(err)
		return err
	}

	return nil
}

func (d *driver) MountedAt(ctx context.Context, mountpath string) string {
	return ""
}

func (d *driver) Mount(ctx context.Context, volumeID string, mountpath string, options map[string]string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		logrus.Println(err)
		return err
	}

	nfsPath, err := d.getNFSPath(v)
	if err != nil {
		logrus.Printf("Could not find server for volume: %s", volumeID)
		return err
	}

	if d.isShared(v.GetSpec()) {
		if v.GetState() != api.VolumeState_VOLUME_STATE_AVAILABLE {
			return fmt.Errorf("Volume is not in an available state")
		}
		// File access
		srcPath := path.Join(":", nfsPath, volumeID)
		mountExists, _ := d.mounter.Exists(srcPath, mountpath)
		if !mountExists {
			/*
				THIS was here and probably needs to be removed
				d.mounter.Unmount(path.Join(nfsPath, volumeID), mountpath,
					syscall.MNT_DETACH, 0, nil)
			*/
			if err := d.mounter.Mount(
				0, path.Join(nfsPath, volumeID),
				mountpath,
				"",
				syscall.MS_BIND,
				"user_xattr",
				0,
				nil,
			); err != nil {
				logrus.Printf("Cannot mount %s at %s because %+v",
					path.Join(nfsPath, volumeID), mountpath, err)
				return err
			}
		}
	} else {
		// Block access
		if d.isRaw(v) {
			return fmt.Errorf("Volume of raw format cannot be mounted")
		}
		if v.GetState() != api.VolumeState_VOLUME_STATE_ATTACHED {
			return fmt.Errorf("Voume %s is not attached", volumeID)
		}
		mountExists, _ := d.mounter.Exists(v.DevicePath, mountpath)
		if !mountExists {
			if err := syscall.Mount(v.DevicePath, mountpath, v.Spec.Format.SimpleString(), 0, ""); err != nil {
				return fmt.Errorf("Failed to mount %v at %v: %v", v.DevicePath, mountpath, err)
			}
		}
	}
	if v.AttachPath == nil {
		v.AttachPath = make([]string, 0)
	}
	v.AttachPath = append(v.AttachPath, mountpath)
	return d.UpdateVol(v)

}

func (d *driver) Unmount(ctx context.Context, volumeID string, mountpath string, options map[string]string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return err
	}
	if len(v.AttachPath) == 0 {
		return fmt.Errorf("Device %v not mounted", volumeID)
	}

	nfsVolPath, err := d.getNFSVolumePath(v)
	if err != nil {
		return err
	}

	if d.isShared(v.GetSpec()) {
		err = d.mounter.Unmount(nfsVolPath, mountpath, syscall.MNT_DETACH, 0, nil)
		if err != nil {
			return err
		}
		v.AttachPath = d.mounter.Mounts(nfsVolPath)
	} else {
		if err := syscall.Unmount(mountpath, 0); err != nil {
			return err
		}
		v.AttachPath = nil
	}
	return d.UpdateVol(v)
}

func (d *driver) clone(newVolumeID, volumeID string) error {
	nfsVolPath, err := d.getNFSVolumePathById(volumeID)
	if err != nil {
		return err
	}

	newNfsVolPath, err := d.getNFSVolumePathById(newVolumeID)
	if err != nil {
		return err
	}

	v, err := d.GetVol(volumeID)
	if err != nil {
		return err
	}

	// NFS does not support snapshots, so just copy the files.

	if d.isShared(v.GetSpec()) {
		// Copy directory
		if err := copyDir(nfsVolPath, newNfsVolPath); err != nil {
			d.Delete(context.Background(), newVolumeID)
			return err
		}
	} else {
		origBlockPath := nfsVolPath + nfsBlockFile
		cloneBlockPath := newNfsVolPath + nfsBlockFile
		// Copy the block file. Could take a while so try to optimize it.
		_, err = exec.Command(
			"/bin/cp",
			"--reflink=always",
			origBlockPath,
			cloneBlockPath,
		).Output()
		if err == nil {
			logrus.Infof("Cloned %s to %s using reflink copy",
				origBlockPath,
				cloneBlockPath)
		} else {
			// Second try sparse copy
			_, err = exec.Command(
				"/bin/cp",
				"--sparse=always",
				origBlockPath,
				cloneBlockPath,
			).Output()
			if err == nil {
				logrus.Infof("Cloned %s to %s using sparse copy",
					origBlockPath,
					cloneBlockPath)
			} else {
				// slow copy
				if err := copyFile(origBlockPath, cloneBlockPath); err == nil {
					logrus.Infof("Cloned %s to %s using slow copy",
						origBlockPath,
						cloneBlockPath)
				} else {
					logrus.Errorf("Failed to clone %s to %s: %v",
						origBlockPath,
						cloneBlockPath,
						err)
					return fmt.Errorf("Failed to clone %s to %s: %v",
						origBlockPath,
						cloneBlockPath,
						err)
				}
			}
		}
	}
	return nil
}

func (d *driver) Snapshot(volumeID string, readonly bool, locator *api.VolumeLocator, noRetry bool) (string, error) {
	volIDs := []string{volumeID}
	vols, err := d.Inspect(volIDs)
	if err != nil {
		return "", nil
	}
	source := &api.Source{Parent: volumeID}
	logrus.Infof("Creating snap vol name: %s", locator.Name)
	return d.Create(context.TODO(), locator, source, vols[0].Spec)
}

func (d *driver) Restore(volumeID string, snapID string) error {
	if _, err := d.Inspect([]string{volumeID, snapID}); err != nil {
		return err
	}

	nfsVolPath, err := d.getNFSVolumePathById(volumeID)
	if err != nil {
		return err
	}

	snapNfsVolPath, err := d.getNFSVolumePathById(snapID)
	if err != nil {
		return err
	}

	// NFS does not support restore, so just copy the files.
	if err := copyDir(snapNfsVolPath, nfsVolPath); err != nil {
		return err
	}
	return nil
}

func (d *driver) SnapshotGroup(groupID string, labels map[string]string, volumeIDs []string, deleteOnFailure bool) (*api.GroupSnapCreateResponse, error) {

	return nil, volume.ErrNotSupported
}

func (d *driver) Attach(ctx context.Context, volumeID string, attachOptions map[string]string) (string, error) {

	nfsPath, err := d.getNFSPathById(volumeID)
	if err != nil {
		return "", err
	}
	blockFile := path.Join(nfsPath, volumeID+nfsBlockFile)

	// Check if it is block
	v, err := util.VolumeFromName(d, volumeID)
	if err != nil {
		return "", err
	}

	// If it has no size, no need to attach
	if d.isShared(v.GetSpec()) {
		return nfsPath, nil
	} else if v.GetState() == api.VolumeState_VOLUME_STATE_ATTACHED {
		// if it already attached.
		return v.GetDevicePath(), nil
	}

	// If it is a block device, create a loop device
	dev, err := losetup.Attach(blockFile, 0, false /* not read only: TODO change this */)
	if err != nil {
		return "", err
	}

	// Update volume info
	v.DevicePath = dev.Path()
	v.State = api.VolumeState_VOLUME_STATE_ATTACHED
	if err := d.UpdateVol(v); err != nil {
		dev.Detach()
		return "", err
	}

	return dev.Path(), nil
}

func (d *driver) Detach(ctx context.Context, volumeID string, options map[string]string) error {

	// Get volume info
	v, err := util.VolumeFromName(d, volumeID)
	if err != nil {
		return err
	}

	// If it has no size, no need to detach
	if d.isShared(v.GetSpec()) {
		return nil
	} else if v.GetState() != api.VolumeState_VOLUME_STATE_ATTACHED {
		// if it is not attached, just return
		return nil
	}

	// Detach -- code from https://github.com/freddierice/go-losetup
	loopFile, err := os.OpenFile(v.GetDevicePath(), os.O_RDONLY, 0660)
	if err != nil {
		return fmt.Errorf("could not open loop device")
	}
	defer loopFile.Close()

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, loopFile.Fd(), losetup.ClrFd, 0)
	if errno != 0 {
		return fmt.Errorf("error clearing loopfile: %v", errno)
	}

	// Update volume info
	v.DevicePath = ""
	v.State = api.VolumeState_VOLUME_STATE_AVAILABLE
	if err := d.UpdateVol(v); err != nil {
		return err
	}

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

func (d *driver) Shutdown() {
	logrus.Printf("%s Shutting down", Name)

	for _, v := range d.nfsServers {
		logrus.Infof("Umounting: %s", nfsMountPath+v)
		syscall.Unmount(path.Join(nfsMountPath, v), 0)
	}
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

func (d *driver) Catalog(volumeID, path, depth string) (api.CatalogResponse, error) {
	return api.CatalogResponse{}, volume.ErrNotSupported
}

func (d *driver) VolService(volumeID string, vtreq *api.VolumeServiceRequest) (*api.VolumeServiceResponse, error) {
	return nil, volume.ErrNotSupported
}

func (d *driver) isRaw(v *api.Volume) bool {
	return d.isRawBySpec(v.GetSpec())
}

func (d *driver) isRawBySpec(spec *api.VolumeSpec) bool {
	return !d.isShared(spec) &&
		spec.GetFormat() == api.FSType_FS_TYPE_NONE
}

func (d *driver) isShared(spec *api.VolumeSpec) bool {
	return spec.GetShared() || spec.GetSharedv4()
}

func mkfsFormatTypeCmd(format string) (string, error) {
	cmd := "mkfs." + format
	dirs := []string{"/usr/sbin", "/sbin", "/usr/bin", "/bin"}
	for _, d := range dirs {
		checkCmd := d + "/" + cmd
		if fileExists(checkCmd) {
			return checkCmd, nil
		}
	}
	return "", fmt.Errorf("File %s not found in %+v", cmd, dirs)
}

// From https://golangcode.com/check-if-a-file-exists/
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
