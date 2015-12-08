package fuse

import (
	"fmt"
	"os"
	"syscall"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/graph"

	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/idtools"

	log "github.com/Sirupsen/logrus"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

const (
	Name          = "fuse"
	Type          = api.Graph
	fuseMountPath = "/var/lib/openstorage/fuse/"
)

type Driver struct {
	// Driver is an implementation of GraphDriver. Only select methods are overridden
	graphdriver.Driver
}

func startFuse() {
	log.Infof("Initializing Fuse Graph driver at %v...", fuseMountPath)

	// In case it is mounted.
	syscall.Unmount(fuseMountPath, 0)

	err := os.MkdirAll(fuseMountPath, 0744)
	if err != nil {
		log.Fatalf("Error while creating FUSE mount path: %v", err)
	}

	c, err := fuse.Mount(
		fuseMountPath,
		fuse.FSName("openstorage"),
		fuse.Subtype("openstoragefs"),
		fuse.LocalVolume(),
		fuse.VolumeName("Open Storage"),
	)

	if err != nil {
		log.Warnf("Error while loading FUSE.  FUSE will not be available as a Graph driver on this system.  Error: %v", err)
		return
	}

	defer c.Close()

	log.Infof("Fuse ready.")

	err = fs.Serve(c, FS{})
	if err != nil {
		log.Fatal(err)
	}

	// Check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
}

func Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) (graphdriver.Driver, error) {
	d := &Driver{}

	go startFuse()

	return d, nil
}

func (d *Driver) String() string {
	return "fuse"
}

// Cleanup performs necessary tasks to release resources
// held by the driver, e.g., unmounting all layered filesystems
// known to this driver.
func (d *Driver) Cleanup() error {
	return nil
}

// Status returns a set of key-value pairs which give low
// level diagnostic status about this driver.
func (d *Driver) Status() [][2]string {
	return [][2]string{
		{"OpenStorage FUSE", "OK"},
	}
}

// Create creates a new, empty, filesystem layer with the
// specified id and parent and mountLabel. Parent and mountLabel may be "".
func (d *Driver) Create(id string, parent string) error {
	path := fuseMountPath + id
	log.Infof("Creating layer %s/%s", path)

	err := os.MkdirAll(path, 0744)
	if err != nil {
		return fmt.Errorf("Error while creating FUSE mount path %v: %v", path, err)
	}

	return nil
}

// Remove attempts to remove the filesystem layer with this id.
func (d *Driver) Remove(id string) error {
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
	path := fuseMountPath + id
	return path, nil
}

// Put releases the system resources for the specified id,
// e.g, unmounting layered filesystem.
func (d *Driver) Put(id string) error {
	return nil
}

// Exists returns whether a filesystem layer with the specified
// ID exists on this driver.
// All cache entries exist.
func (d *Driver) Exists(id string) bool {
	return true
}

// FS implements the graph file system.
type FS struct{}

func (FS) Root() (fs.Node, error) {
	return Dir{}, nil
}

// Dir implements both Node and Handle for the root directory.
type Dir struct{}

func (Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

func (Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if name == "hello" {
		return File{}, nil
	}
	return nil, fuse.ENOENT
}

var dirDirs = []fuse.Dirent{
	{Inode: 2, Name: "hello", Type: fuse.DT_File},
}

func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return dirDirs, nil
}

// File implements both Node and Handle for the hello file.
type File struct{}

const greeting = "hello, world\n"

func (File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 2
	a.Mode = 0444
	a.Size = uint64(len(greeting))
	return nil
}

func (File) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte(greeting), nil
}

func init() {
	graph.Register("fuse", Init)
}
