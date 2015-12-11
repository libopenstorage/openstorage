package fuse

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

const (
	virtPath = "/var/lib/openstorage/fuse/virtual"
	physPath = "/var/lib/openstorage/fuse/physical"
)

var (
	fhCache   map[string]*FileHandle
	fileCache map[string]*File
	fuseConn  *fuse.Conn
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
	// log.Infof("Checking directory attributes on %s", d.path)

	fi, err := os.Lstat(d.path)
	if err != nil {
		return err
	}

	a.Inode = (fi.Sys().(*syscall.Stat_t).Ino)
	a.Mode = fi.Mode()
	return nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	fullPath := path.Join(d.path, name)
	// log.Infof("Directory lookup on %s", fullPath)

	fi, err := os.Lstat(fullPath)
	if err != nil {
		return nil, fuse.ENOENT
	}

	if fi.Mode()&os.ModeDir == os.ModeDir {
		return &Dir{path: fullPath}, nil
	} else {
		f := &File{path: fullPath}
		putFile(fullPath, f)
		return f, nil
	}
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	// log.Infof("Readdir on %s", d.path)

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
	// log.Infof("Mkdir on %s", fullPath)

	err := os.MkdirAll(fullPath, req.Mode)
	if err != nil {
		return nil, err
	}

	return &Dir{path: fullPath}, nil
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	file := path.Join(d.path, req.Name)
	// log.Infof("Creating file %s", file)
	f, err := os.Create(file)
	if err != nil {
		return nil, nil, err
	}

	fh := &FileHandle{
		path: file,
		f:    f}

	putFileHandle(file, fh)

	fc := &File{path: file}
	putFile(file, fc)

	// XXX do we need to populate resp.LookupResponse https://godoc.org/bazil.org/fuse#LookupResponse

	return fc, fh, nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	fullPath := path.Join(d.path, req.Name)
	// log.Infof("Remove on %s", fullPath)

	err := os.RemoveAll(fullPath)

	return err
}

func (d *Dir) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	tgtDir := newDir.(*Dir)
	oldpath := path.Join(d.path, req.OldName)
	newpath := path.Join(tgtDir.path, req.NewName)

	if newDir != d {
		log.Warnf("Rename called from incorrect directory: %s.", tgtDir)
		return fuse.Errno(syscall.EXDEV)
	}

	// log.Infof("Renaming %s to %s", oldpath, newpath)

	err := os.Rename(oldpath, newpath)
	if err != nil {
		return err
	}

	f, err := getFile(oldpath)
	if err != nil {
		log.Warnf("Could not find old file in the cache.")
		return err
	}

	f.path = newpath
	putFile(newpath, f)

	return nil
}

// File operations.
type File struct {
	path string
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	// log.Infof("Checking file attributes on %s", f.path)

	fi, err := os.Lstat(f.path)
	if err != nil {
		log.Warnf("%v", err)
		return err
	}

	stat := fi.Sys().(*syscall.Stat_t)

	a.Inode = stat.Ino
	a.Atime = time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
	a.Ctime = time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
	a.Mtime = time.Unix(int64(stat.Mtim.Sec), int64(stat.Mtim.Nsec))
	a.Mode = os.FileMode(stat.Mode)
	a.Nlink = uint32(stat.Nlink)
	a.Blocks = uint64(stat.Blocks)
	a.Size = uint64(stat.Size)
	a.Uid = stat.Uid
	a.Gid = stat.Gid
	a.Rdev = uint32(stat.Rdev)

	return nil
}

func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	// log.Infof("Setting file attributes on %s: %v", f.path, req)

	err := os.Chmod(f.path, req.Mode)
	if err != nil {
		return err
	}

	err = os.Chown(f.path, int(req.Uid), int(req.Gid))
	if err != nil {
		return err
	}

	return nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	// log.Infof("Opening file %s: %v", f.path, req)

	flags := int(req.Flags)
	file, err := os.OpenFile(f.path, flags, 0777)
	if err != nil {
		return nil, err
	}

	fh := &FileHandle{
		path: f.path,
		f:    file}

	putFileHandle(fh.path, fh)

	resp.Flags = fuse.OpenKeepCache

	return fh, nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	// log.Infof("Reading file %s", f.path)

	b, err := ioutil.ReadFile(f.path)
	if err == io.ErrUnexpectedEOF || err == io.EOF {
		err = nil
	}
	return b, err
}

func (f *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	// log.Infof("Syncing file %s", f.path)
	fh, err := getFileHandle(f.path)
	if err != nil {
		log.Warnf("Could not find file %s in the file handle cache.", f.path)
		return err
	}
	err = fh.f.Sync()
	return err
}

// File handle operations.
type FileHandle struct {
	path string
	f    *os.File
}

func (fh *FileHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	// log.Infof("Writing file %s", fh.path)

	sz, err := fh.f.WriteAt(req.Data, req.Offset)
	if err != nil {
		log.Errorf("Error while writing to %s: %v", fh.path, err)
		return err
	}

	resp.Size = sz

	return err
}

func (fh *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	// log.Infof("Reading file %s", fh.path)

	buf := make([]byte, req.Size)
	n, err := fh.f.ReadAt(buf, req.Offset)
	resp.Data = buf[:n]
	if err == io.ErrUnexpectedEOF || err == io.EOF {
		err = nil
	}

	return err
}

func (fh *FileHandle) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	// log.Infof("Syncing file %s", fh.path)

	err := fh.f.Sync()
	return err
}

func (fh *FileHandle) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	// log.Infof("Syncing file %s", fh.path)

	err := fh.f.Sync()
	return err
}

func (fh *FileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	// log.Infof("Releasing file %s", fh.path)

	err := fh.f.Close()
	return err
}

func getFile(path string) (*File, error) {
	f, ok := fileCache[path]

	if !ok {
		return nil, fuse.EIO
	} else {
		return f, nil
	}
}

func putFile(path string, file *File) {
	fileCache[path] = file
}

func getFileHandle(path string) (*FileHandle, error) {
	fh, ok := fhCache[path]

	if !ok {
		return nil, fuse.EIO
	} else {
		return fh, nil
	}
}

func putFileHandle(path string, fh *FileHandle) {
	fhCache[path] = fh
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

	fuseConn, err = fuse.Mount(
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

	defer fuseConn.Close()

	log.Infof("Fuse ready.")

	err = fs.Serve(fuseConn, FS{})
	if err != nil {
		log.Fatal(err)
	}

	// Check if the mount process has an error to report
	<-fuseConn.Ready
	if err := fuseConn.MountError; err != nil {
		log.Fatal(err)
	}
}

func fusePath() string {
	return virtPath
}

func init() {
	fhCache = make(map[string]*FileHandle)
	fileCache = make(map[string]*File)
}
