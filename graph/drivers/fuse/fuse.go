package fuse

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

const (
	virtPath = "/var/lib/openstorage/fuse/virtual"
	physPath = "/var/lib/openstorage/fuse/physical"
)

// FS implements the graph file system.
type FS struct{}

func (FS) Root() (fs.Node, error) {
	return &Dir{path: physPath}, nil
}

// Directory operations.
type Dir struct {
	path string
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Infof("Checking directory attributes on %s", d.path)

	fi, err := os.Stat(d.path)
	if err != nil {
		return err
	}

	a.Inode = (fi.Sys().(*syscall.Stat_t).Ino)
	a.Mode = fi.Mode()
	return nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	fullPath := path.Join(d.path, name)
	log.Infof("Directory lookup on %s", fullPath)

	fi, err := os.Stat(fullPath)
	if err != nil {
		log.Infof("Error is %v", err)
		return nil, fuse.ENOENT
	}

	if fi.Mode()&os.ModeDir == os.ModeDir {
		return &Dir{path: fullPath}, nil
	} else {
		return &File{path: fullPath}, nil
	}
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Infof("Readdir on %s", d.path)

	fi, err := ioutil.ReadDir(d.path)
	if err != nil {
		return nil, err
	}

	var res []fuse.Dirent
	for _, f := range fi {
		var de fuse.Dirent
		de.Name = f.Name()
		de.Inode = (f.Sys().(*syscall.Stat_t).Ino)
		if f.IsDir() {
			de.Type = fuse.DT_Dir
		} else {
			de.Type = fuse.DT_File
		}

		res = append(res, de)
	}

	return res, nil
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	fullPath := path.Join(d.path, req.Name)
	log.Infof("Mkdir on %s", fullPath)

	err := os.MkdirAll(fullPath, req.Mode)
	if err != nil {
		return nil, err
	}

	return &Dir{path: fullPath}, nil
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	file := path.Join(d.path, req.Name)
	log.Infof("Creating file %s", file)
	f, err := os.Create(file)
	if err != nil {
		return nil, nil, err
	}

	fh := &FileHandle{
		path: file,
		f:    f}

	return &File{path: file}, fh, nil
}

// File operations.
type File struct {
	path string
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Infof("Checking file attributes on %s", f.path)

	fi, err := os.Stat(f.path)
	if err != nil {
		return err
	}

	a.Inode = (fi.Sys().(*syscall.Stat_t).Ino)
	a.Mode = fi.Mode()
	a.Size = uint64(fi.Size())
	a.Mtime = fi.ModTime()

	return nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Infof("Opening file %s: %v", f.path, req)

	file, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}

	fh := &FileHandle{
		path: f.path,
		f:    file}

	return fh, nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	log.Infof("Reading file %s", f.path)

	b, err := ioutil.ReadFile(f.path)
	if err == io.ErrUnexpectedEOF || err == io.EOF {
		err = nil
	}
	return b, err
}

// File handle operations.
type FileHandle struct {
	path string
	f    *os.File
}

func (fh *FileHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Infof("Writing file %s", fh.path)

	sz, err := fh.f.WriteAt(req.Data, req.Offset)

	resp.Size = sz

	return err
}

func (fh *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	log.Infof("Reading file %s", fh.path)

	buf := make([]byte, req.Size)
	n, err := fh.f.ReadAt(buf, req.Offset)
	resp.Data = buf[:n]
	if err == io.ErrUnexpectedEOF || err == io.EOF {
		err = nil
	}

	return err
}

func startFuse() {
	log.Infof("Initializing Fuse Graph driver at %v...", virtPath)

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

	c, err := fuse.Mount(
		virtPath,
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

func fusePath() string {
	return virtPath
}
