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

	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/chrootarchive"
	"github.com/docker/docker/pkg/directory"
	"github.com/docker/docker/pkg/idtools"

	log "github.com/Sirupsen/logrus"
)

const (
	Name     = "unionfs"
	Type     = api.Graph
	virtPath = "/var/lib/openstorage/fuse/virtual"
	physPath = "/var/lib/openstorage/fuse/physical"
)

type Driver struct {
}

func Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) (graphdriver.Driver, error) {
	log.Infof("Initializing Fuse Graph driver at home:%s and storage: %v...", home, virtPath)

	// In case it is mounted.
	syscall.Unmount(virtPath, 0)

	err := os.MkdirAll(virtPath, 0744)
	if err != nil {
		log.Fatalf("Error while creating FUSE mount path: %v", err)
	}

	err = os.MkdirAll(physPath, 0744)
	if err != nil {
		log.Fatalf("Error while creating FUSE mount path: %v", err)
	}

	cVirtPath := C.CString(virtPath)
	cPhysPath := C.CString(physPath)
	go C.start_unionfs(cPhysPath, cVirtPath)

	d := &Driver{}

	return d, nil
}

func (d *Driver) String() string {
	return "openstorage-fuse"
}

// Cleanup performs necessary tasks to release resources
// held by the driver, e.g., unmounting all layered filesystems
// known to this driver.
func (d *Driver) Cleanup() error {
	log.Infof("Cleaning up fuse %s", virtPath)
	// syscall.Unmount(virtPath, 0)
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

	// log.Infof("Linking layer %s to parent layer %s", child, parent)

	child = child + "/_parent"

	err := os.Symlink(parent, child)
	if err != nil {
		return fmt.Errorf("Error while linking FUSE mount path %v to %v: %v", child, parent, err)
	}

	return nil
}

// Create creates a new, empty, filesystem layer with the
// specified id and parent and mountLabel. Parent and mountLabel may be "".
func (d *Driver) Create(id string, parent string) error {
	path := path.Join(physPath, id)

	// log.Infof("Creating layer %s", path)

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
	// path := path.Join(physPath, id)

	// log.Infof("Removing layer %s", path)

	// XXX FIXME os.RemoveAll(path)

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
		log.Warnf("Error while creating a union FS for %s", id)
		return "", err
	} else {
		log.Infof("Created a union FS for %s", id)
		unionPath := path.Join(virtPath, id)

		return unionPath, err
	}
}

// Put releases the system resources for the specified id,
// e.g, unmounting layered filesystem.
func (d *Driver) Put(id string) error {
	log.Infof("Releasing union FS for %s", id)

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
		log.Warnf("Error while applying diff to %s: %v", id, err)
		os.Exit(-1)
		return 0, err
	}

	// show invalid whiteouts warning.
	files, err := ioutil.ReadDir(path.Join(dir, archive.WhiteoutLinkDir))
	if err == nil && len(files) > 0 {
		log.Warnf("Archive contains aufs hardlink references that are not supported.")
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
	return archive.TarWithOptions(path.Join(physPath, id), &archive.TarOptions{
		Compression:     archive.Uncompressed,
		ExcludePatterns: []string{archive.WhiteoutMetaPrefix + "*", "!" + archive.WhiteoutOpaqueDir},
	})
}

// DiffSize calculates the changes between the specified id
// and its parent and returns the size in bytes of the changes
// relative to its base filesystem directory.
func (d *Driver) DiffSize(id, parent string) (size int64, err error) {
	return directory.Size(path.Join(physPath, id))
}

func init() {
	graph.Register(Name, Init)
}
