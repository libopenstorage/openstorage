// +build linux,have_unionfs

package unionfs

/*
extern int start_unionfs(char *, char *);
extern int alloc_unionfs(char *, char *);
extern int release_unionfs(char *);
#cgo LDFLAGS: -lfuse -lulockmgr
*/
import "C"

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"syscall"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/graph"
	"github.com/libopenstorage/openstorage/volume"

	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/chrootarchive"
	"github.com/docker/docker/pkg/directory"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/parsers"

	"github.com/Sirupsen/logrus"
)

const (
	Name                = "unionfs"
	Type                = api.Graph
	UnionFSVolumeDriver = "unionfs.volume_driver"
	virtPath            = "/var/lib/openstorage/fuse/virtual"
	physPath            = "/var/lib/openstorage/fuse/physical"
)

type Driver struct {
	volDriver volume.VolumeDriver
}

func Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) (graphdriver.Driver, error) {
	logrus.Infof("Initializing Fuse Graph driver at home:%s and storage: %v...", home, virtPath)

	var volumeDriver string
	for _, option := range options {
		key, val, err := parsers.ParseKeyValueOpt(option)
		if err != nil {
			return nil, err
		}
		switch key {
		case UnionFSVolumeDriver:
			volumeDriver = val
		default:
			return nil, fmt.Errorf("Unknown option %s\n", key)
		}
	}

	if volumeDriver == "" {
		logrus.Warnf("Error - no volume driver specified for UnionFS")
		return nil, fmt.Errorf("No volume driver specified for UnionFS")
	}

	logrus.Infof("UnionFS volume driver: %v", volumeDriver)
	volDriver, err := volume.Get(volumeDriver)
	if err != nil {
		logrus.Warnf("Error while loading volume driver: %s", volumeDriver)
		return nil, err
	}

	// In case it is mounted.
	syscall.Unmount(virtPath, 0)

	err = os.MkdirAll(virtPath, 0744)
	if err != nil {
		volDriver.Shutdown()
		logrus.Warnf("Error while creating FUSE mount path: %v", err)
		return nil, err
	}

	err = os.MkdirAll(physPath, 0744)
	if err != nil {
		volDriver.Shutdown()
		logrus.Warnf("Error while creating FUSE mount path: %v", err)
		return nil, err
	}

	cVirtPath := C.CString(virtPath)
	cPhysPath := C.CString(physPath)
	go C.start_unionfs(cPhysPath, cVirtPath)

	d := &Driver{
		volDriver: volDriver,
	}

	return d, nil
}

func (d *Driver) String() string {
	return "openstorage-fuse"
}

// Cleanup performs necessary tasks to release resources
// held by the driver, e.g., unmounting all layered filesystems
// known to this driver.
func (d *Driver) Cleanup() error {
	logrus.Infof("Cleaning up fuse %s", virtPath)

	d.volDriver.Shutdown()
	syscall.Unmount(virtPath, 0)

	return nil
}

// Status returns a set of key-value pairs which give low
// level diagnostic status about this driver.
func (d *Driver) Status() [][2]string {
	return [][2]string{
		{"OpenStorage FUSE", "OK"},
	}
}

func (d *Driver) linkParent(child, parent string) error {
	parent = path.Join(physPath, parent)

	logrus.Debugf("Linking layer %s to parent layer %s", child, parent)

	child = child + "/.unionfs.parent"

	err := os.Symlink(parent, child)
	if err != nil {
		return fmt.Errorf("Error while linking FUSE mount path %v to %v: %v", child, parent, err)
	}

	return nil
}

// Create creates a new, empty, filesystem layer with the
// specified id and parent and mountLabel. Parent and mountLabel may be "".
func (d *Driver) Create(id string, parent string, mountLabel string) error {
	path := path.Join(physPath, id)

	logrus.Debugf("Creating layer %s", path)

	err := os.MkdirAll(path, 0744)
	if err != nil {
		return fmt.Errorf("Error while creating FUSE mount path %v: %v", path, err)
	}

	if parent != "" {
		return d.linkParent(path, parent)
	}

	return nil
}

// Remove attempts to remove the filesystem layer with this id.
func (d *Driver) Remove(id string) error {
	path := path.Join(physPath, id)

	logrus.Debugf("Removing layer %s", path)

	os.RemoveAll(path)

	return nil
}

// Returns a set of key-value pairs which give low level information
// about the image/container driver is managing.
func (d *Driver) GetMetadata(id string) (map[string]string, error) {
	return nil, nil
}

// Get returns the mountpoint for the layered filesystem referred
// to by this id. You can optionally specify a mountLabel or "".
// Returns the absolute path to the mounted layered filesystem.
func (d *Driver) Get(id, mountLabel string) (string, error) {
	layerPath := path.Join(physPath, id)

	cLayerPath := C.CString(layerPath)
	cID := C.CString(id)

	ret, err := C.alloc_unionfs(cLayerPath, cID)
	if int(ret) != 0 {
		logrus.Warnf("Error while creating a union FS for %s", id)
		return "", err
	} else {
		logrus.Debugf("Created a union FS for %s", id)
		unionPath := path.Join(virtPath, id)

		return unionPath, err
	}
}

// Put releases the system resources for the specified id,
// e.g, unmounting layered filesystem.
func (d *Driver) Put(id string) error {
	logrus.Debugf("Releasing union FS for %s", id)

	cID := C.CString(id)
	_, err := C.release_unionfs(cID)

	return err
}

// Exists returns whether a filesystem layer with the specified
// ID exists on this driver.
// All cache entries exist.
func (d *Driver) Exists(id string) bool {
	path := path.Join(physPath, id)

	_, err := os.Stat(path)

	if err == nil {
		return true
	} else {
		return false
	}
}

// ApplyDiff extracts the changeset from the given diff into the
// layer with the specified id and parent, returning the size of the
// new layer in bytes.
// The archive.Reader must be an uncompressed stream.
func (d *Driver) ApplyDiff(id string, parent string, diff archive.Reader) (size int64, err error) {
	dir := path.Join(physPath, id)

	if err := chrootarchive.UntarUncompressed(diff, dir, nil); err != nil {
		logrus.Warnf("Error while applying diff to %s: %v", id, err)
		return 0, err
	}

	// show invalid whiteouts warning.
	files, err := ioutil.ReadDir(path.Join(dir, archive.WhiteoutLinkDir))
	if err == nil && len(files) > 0 {
		logrus.Warnf("Archive contains aufs hardlink references that are not supported.")
	}

	return d.DiffSize(id, parent)
}

// Changes produces a list of changes between the specified layer
// and its parent layer. If parent is "", then all changes will be ADD changes.
func (d *Driver) Changes(id, parent string) ([]archive.Change, error) {

	return nil, nil
}

// Diff produces an archive of the changes between the specified
// layer and its parent layer which may be "".
func (d *Driver) Diff(id, parent string) (archive.Archive, error) {
	return archive.TarWithOptions(path.Join(virtPath, id), &archive.TarOptions{
		Compression:     archive.Uncompressed,
		ExcludePatterns: []string{archive.WhiteoutMetaPrefix + "*", "!" + archive.WhiteoutOpaqueDir},
	})
}

// DiffSize calculates the changes between the specified id
// and its parent and returns the size in bytes of the changes
// relative to its base filesystem directory.
func (d *Driver) DiffSize(id, parent string) (size int64, err error) {
	return directory.Size(path.Join(virtPath, id))
}

func (d *Driver) Read() (size int64, err error) {

	return 0, nil
}

func init() {
	graph.Register(Name, Init)
}
