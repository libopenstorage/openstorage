package fuse

import (
	"os"
	"syscall"

	"github.com/docker/docker/daemon/graphdriver/overlay"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/graph"

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

func startFuse() {
	log.Infof("Initializing Fuse Graph driver...")

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

func init() {

	graph.Register("fuse", overlay.Init)

	// XXX move this to Init()
	go startFuse()
}
