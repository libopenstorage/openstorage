package test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/libopenstorage/kvdb"
	"github.com/libopenstorage/kvdb/mem"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

// Context maintains current device state. It gets passed into tests
// so that tests can build on other tests' work
type Context struct {
	volume.VolumeDriver
	volID      api.VolumeID
	snapID     api.SnapID
	mountPath  string
	tgtPath    string
	devicePath string
}

func NewContext(driverName string) (*Context, error) {
	d, err := volume.Get(driverName)
	if err != nil {
		return nil, err
	}
	return &Context{VolumeDriver: d}, nil
}

func Run(t *testing.T, ctx *Context) {
	create(t, ctx)
	inspect(t, ctx)
	enumerate(t, ctx)
	format(t, ctx)
	attach(t, ctx)
	mount(t, ctx)
	io(t, ctx)
	detach(t, ctx)
	deleteBad(t, ctx)
	unmount(t, ctx)
	detach(t, ctx)
	delete(t, ctx)
	shutdown(t, ctx)
}

func RunSnap(t *testing.T, ctx *Context) {
	snap(t, ctx)
	snapInspect(t, ctx)
	snapEnumerate(t, ctx)
	snapDiff(t, ctx)
	snapDelete(t, ctx)
}

func create(t *testing.T, ctx *Context) {
	fmt.Println("create")

	volID, err := ctx.Create(
		api.VolumeLocator{Name: "foo"},
		&api.CreateOptions{FailIfExists: false},
		&api.VolumeSpec{Size: 10240000,
			HALevel: 1,
			Format:  api.FsExt4,
		})

	assert.NoError(t, err, "Failed in Create")
	ctx.volID = volID
}

func inspect(t *testing.T, ctx *Context) {
	fmt.Println("inspect")

	vols, err := ctx.Inspect([]api.VolumeID{ctx.volID})
	assert.NoError(t, err, "Failed in Inspect")
	assert.NotNil(t, vols, "Nil vols")
	assert.Equal(t, len(vols), 1, "Expect 1 volume actual %v volumes", len(vols))
	assert.Equal(t, vols[0].ID, ctx.volID, "Expect volID %v actual %v", ctx.volID, vols[0].ID)

	vols, err = ctx.Inspect([]api.VolumeID{api.VolumeID("shouldNotExist")})
	assert.Equal(t, len(vols), 0, "Expect 0 volume actual %v volumes", len(vols))
}

func enumerate(t *testing.T, ctx *Context) {
	fmt.Println("enumerate")

	vols, err := ctx.Enumerate(api.VolumeLocator{}, nil)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.NotNil(t, vols, "Nil vols")
	assert.Equal(t, 1, len(vols), "Expect 1 volume actual %v volumes", len(vols))
	assert.Equal(t, vols[0].ID, ctx.volID, "Expect volID %v actual %v", ctx.volID, vols[0].ID)

	vols, err = ctx.Enumerate(api.VolumeLocator{Name: "foo"}, nil)
	assert.NoError(t, err, "Failed in Enumerate")
	assert.NotNil(t, vols, "Nil vols")
	assert.Equal(t, len(vols), 1, "Expect 1 volume actual %v volumes", len(vols))
	assert.Equal(t, vols[0].ID, ctx.volID, "Expect volID %v actual %v", ctx.volID, vols[0].ID)

	vols, err = ctx.Enumerate(api.VolumeLocator{Name: "shouldNotExist"}, nil)
	assert.Equal(t, len(vols), 0, "Expect 0 volume actual %v volumes", len(vols))
}

func format(t *testing.T, ctx *Context) {
	fmt.Println("format")

	err := ctx.Format(ctx.volID)
	if err != nil {
		assert.Equal(t, err, volume.ErrNotSupported, "Error on format %v", err)
	}
}

func attach(t *testing.T, ctx *Context) {
	fmt.Println("attach")
	p, err := ctx.Attach(ctx.volID)
	if err != nil {
		assert.Equal(t, err, volume.ErrNotSupported, "Error on attach %v", err)
	}
	ctx.devicePath = p

	p, err = ctx.Attach(ctx.volID)
	if err == nil {
		assert.Equal(t, p, ctx.devicePath, "Multiple calls to attach if not errored should return the same path")
	}
}

func detach(t *testing.T, ctx *Context) {
	fmt.Println("detach")

	err := ctx.Detach(ctx.volID)
	if err != nil {
		assert.Equal(t, ctx.devicePath, "", "Error on detach %s: %v", ctx.devicePath, err)
	}
	ctx.devicePath = ""

	err = ctx.Detach(ctx.volID)
	assert.Error(t, err, "Detaching an already detached device should fail")
}

func mount(t *testing.T, ctx *Context) {
	fmt.Println("mount")

	mountPath := "/mnt/voltest"
	err := os.MkdirAll(mountPath, 0755)

	tgtPath := "/mnt/foo"
	err = os.MkdirAll(tgtPath, 0755)
	assert.NoError(t, err, "Failed in mkdir")

	err = ctx.Mount(ctx.volID, tgtPath)
	assert.NoError(t, err, "Failed in mount")

	ctx.mountPath = mountPath
	ctx.tgtPath = tgtPath
}

func unmount(t *testing.T, ctx *Context) {
	fmt.Println("unmount")

	assert.NotEqual(t, ctx.mountPath, "", "Device is not mounted")
	err := ctx.Unmount(ctx.volID, ctx.mountPath)

	assert.NoError(t, err, "Failed in unmount")
	err = ctx.Unmount(ctx.volID, ctx.tgtPath)

	assert.NoError(t, err, "Failed in unmount")

	ctx.mountPath = ""
	ctx.tgtPath = ""
}

func shutdown(t *testing.T, ctx *Context) {
	fmt.Println("shutdown")
	ctx.Shutdown()
}

func io(t *testing.T, ctx *Context) {
	fmt.Println("io")
	assert.NotEqual(t, ctx.mountPath, "", "Device is not mounted")

	cmd := exec.Command("dd", "if=/dev/urandom", "of=/tmp/xx", "bs=1M", "count=10")
	err := cmd.Run()
	assert.NoError(t, err, "Failed to run dd")

	cmd = exec.Command("dd", "if=/tmp/xx", fmt.Sprintf("of=%s/xx", ctx.mountPath))
	err = cmd.Run()
	assert.NoError(t, err, "Failed to run dd on mountpoint %s/xx", ctx.mountPath)

	cmd = exec.Command("diff", "if=/tmp/xx", fmt.Sprintf("of=%s/xx", ctx.mountPath))
	assert.NoError(t, err, "data mismatch")
}

func detachBad(t *testing.T, ctx *Context) {
	err := ctx.Detach(ctx.volID)
	assert.True(t, (err == nil || err == volume.ErrNotSupported),
		"Detach on mounted device should fail")
}

func deleteBad(t *testing.T, ctx *Context) {
	fmt.Println("deleteBad")
	assert.NotEqual(t, ctx.mountPath, "", "Device is not mounted")

	err := ctx.Delete(ctx.volID)
	assert.Error(t, err, "Delete on mounted device must fail")
}

func delete(t *testing.T, ctx *Context) {
	fmt.Println("delete")
	err := ctx.Delete(ctx.volID)
	assert.NoError(t, err, "Delete failed")
}

func snap(t *testing.T, ctx *Context) {
}

func snapInspect(t *testing.T, ctx *Context) {
}

func snapEnumerate(t *testing.T, ctx *Context) {
}

func snapDiff(t *testing.T, ctx *Context) {
}

func snapDelete(t *testing.T, ctx *Context) {
}

func init() {
	kv, err := kvdb.New(mem.Name, "driver_test", []string{}, nil)
	if err != nil {
		log.Panicf("Failed to intialize KVDB")
	}
	err = kvdb.SetInstance(kv)
	if err != nil {
		log.Panicf("Failed to set KVDB instance")
	}
}
