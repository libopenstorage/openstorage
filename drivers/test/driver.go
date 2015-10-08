package test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const testPath = "/tmp/openstorage/mount"

// Context maintains current device state. It gets passed into tests
// so that tests can build on other tests' work
type Context struct {
	volume.VolumeDriver
	volID      api.VolumeID
	snapID     api.VolumeID
	mountPath  string
	devicePath string
	Filesystem string
}

func NewContext(d volume.VolumeDriver) *Context {
	return &Context{
		VolumeDriver: d,
		volID:        api.BadVolumeID,
		snapID:       api.BadVolumeID,
		Filesystem:   string(""),
	}
}

func RunShort(t *testing.T, ctx *Context) {
	create(t, ctx)
	inspect(t, ctx)
	enumerate(t, ctx)
	attach(t, ctx)
	format(t, ctx)
	mount(t, ctx)
	io(t, ctx)
	unmount(t, ctx)
	detach(t, ctx)
	delete(t, ctx)
	runEnd(t, ctx)
}

func Run(t *testing.T, ctx *Context) {
	RunShort(t, ctx)
	RunSnap(t, ctx)
	runEnd(t, ctx)
}

func runEnd(t *testing.T, ctx *Context) {
	time.Sleep(time.Second * 2)
	os.RemoveAll(testPath)
	shutdown(t, ctx)
}

func RunSnap(t *testing.T, ctx *Context) {
	snap(t, ctx)
	snapInspect(t, ctx)
	snapEnumerate(t, ctx)
	snapDiff(t, ctx)
	snapDelete(t, ctx)
	detach(t, ctx)
	delete(t, ctx)
}

func create(t *testing.T, ctx *Context) {
	fmt.Println("create")

	volID, err := ctx.Create(
		api.VolumeLocator{Name: "foo"},
		nil,
		&api.VolumeSpec{
			Size:    1 * 1024 * 1024 * 1024,
			HALevel: 1,
			Format:  api.Filesystem(ctx.Filesystem),
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
	assert.Equal(t, 0, len(vols), "Expect 0 volume actual %v volumes", len(vols))
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

func waitReady(t *testing.T, ctx *Context) error {
	total := time.Minute * 5
	inc := time.Second * 2
	elapsed := time.Second * 0
	vols, err := ctx.Inspect([]api.VolumeID{ctx.volID})
	for err == nil && len(vols) == 1 && vols[0].Status != api.Up && elapsed < total {
		time.Sleep(inc)
		elapsed += inc
		vols, err = ctx.Inspect([]api.VolumeID{ctx.volID})
	}
	if err != nil {
		return err
	}
	if len(vols) != 1 {
		return fmt.Errorf("Expect one volume from inspect got %v", len(vols))
	}
	if vols[0].Status != api.Up {
		return fmt.Errorf("Timed out waiting for volume status %v", vols)
	}
	return err
}

func attach(t *testing.T, ctx *Context) {
	fmt.Println("attach")
	err := waitReady(t, ctx)
	assert.NoError(t, err, "Volume status is not up")
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
}

func mount(t *testing.T, ctx *Context) {
	fmt.Println("mount")

	err := os.MkdirAll(testPath, 0755)

	err = ctx.Mount(ctx.volID, testPath)
	assert.NoError(t, err, "Failed in mount")

	ctx.mountPath = testPath
}

func unmount(t *testing.T, ctx *Context) {
	fmt.Println("unmount")

	assert.NotEqual(t, ctx.mountPath, "", "Device is not mounted")

	err := ctx.Unmount(ctx.volID, ctx.mountPath)
	assert.NoError(t, err, "Failed in unmount")

	ctx.mountPath = ""
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
	ctx.volID = api.BadVolumeID
}

func snap(t *testing.T, ctx *Context) {
	fmt.Println("snap")
	if ctx.volID == api.BadVolumeID {
		create(t, ctx)
	}
	attach(t, ctx)
	labels := api.Labels{"oh": "snap"}
	assert.NotEqual(t, ctx.volID, api.BadVolumeID, "invalid volume ID")
	id, err := ctx.Snapshot(ctx.volID, false,
		api.VolumeLocator{Name: "snappy", VolumeLabels: labels})
	assert.NoError(t, err, "Failed in creating a snapshot")
	ctx.snapID = id
}

func snapInspect(t *testing.T, ctx *Context) {
	fmt.Println("snapInspect")

	snaps, err := ctx.Inspect([]api.VolumeID{ctx.snapID})
	assert.NoError(t, err, "Failed in Inspect")
	assert.NotNil(t, snaps, "Nil snaps")
	assert.Equal(t, len(snaps), 1, "Expect 1 snaps actual %v snaps", len(snaps))
	assert.Equal(t, snaps[0].ID, ctx.snapID, "Expect snapID %v actual %v", ctx.snapID, snaps[0].ID)

	snaps, err = ctx.Inspect([]api.VolumeID{api.VolumeID("shouldNotExist")})
	assert.Equal(t, 0, len(snaps), "Expect 0 snaps actual %v snaps", len(snaps))
}

func snapEnumerate(t *testing.T, ctx *Context) {
	fmt.Println("snapEnumerate")

	snaps, err := ctx.SnapEnumerate(nil, nil)
	assert.NoError(t, err, "Failed in snapEnumerate")
	assert.NotNil(t, snaps, "Nil snaps")
	assert.Equal(t, 1, len(snaps), "Expect 1 snaps actual %v snaps", len(snaps))
	assert.Equal(t, snaps[0].ID, ctx.snapID, "Expect snapID %v actual %v", ctx.snapID, snaps[0].ID)
	labels := snaps[0].Locator.VolumeLabels

	snaps, err = ctx.SnapEnumerate([]api.VolumeID{ctx.volID}, nil)
	assert.NoError(t, err, "Failed in snapEnumerate")
	assert.NotNil(t, snaps, "Nil snaps")
	assert.Equal(t, len(snaps), 1, "Expect 1 snap actual %v snaps", len(snaps))
	assert.Equal(t, snaps[0].ID, ctx.snapID, "Expect snapID %v actual %v", ctx.snapID, snaps[0].ID)

	snaps, err = ctx.SnapEnumerate([]api.VolumeID{api.VolumeID("shouldNotExist")}, nil)
	assert.Equal(t, len(snaps), 0, "Expect 0 snap actual %v snaps", len(snaps))

	snaps, err = ctx.SnapEnumerate(nil, labels)
	assert.NoError(t, err, "Failed in snapEnumerate")
	assert.NotNil(t, snaps, "Nil snaps")
	assert.Equal(t, len(snaps), 1, "Expect 1 snap actual %v snaps", len(snaps))
	assert.Equal(t, snaps[0].ID, ctx.snapID, "Expect snapID %v actual %v", ctx.snapID, snaps[0].ID)
}

func snapDiff(t *testing.T, ctx *Context) {
	fmt.Println("snapDiff")
}

func snapDelete(t *testing.T, ctx *Context) {
	fmt.Println("snapDelete")
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
