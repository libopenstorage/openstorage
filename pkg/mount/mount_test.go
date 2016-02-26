package mount

import (
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	source = "/tmp/ost/mount_test_src"
	dest   = "/tmp/ost/mount_test_dest"
)

var m Manager

func TestSetup(t *testing.T) {
	var err error
	m, err = New(NFSMount, "")
	if err != nil {
		t.Fatalf("Failed to setup test %v", err)
	}
	cleandir(source)
	cleandir(dest)
}

func cleandir(dir string) {
	syscall.Unmount(dir, 0)
	os.RemoveAll(dir)
	os.MkdirAll(dir, 0755)
}

func TestLoad(t *testing.T) {
	require.NoError(t, m.Load(""), "Failed in load")
}

func TestMount(t *testing.T) {
	err := m.Mount(0, source, dest, "", syscall.MS_BIND, "")
	require.NoError(t, err, "Failed in mount")
	err = m.Unmount(source, dest)
	require.NoError(t, err, "Failed in unmount")
}

func TestInspect(t *testing.T) {
	p := m.Inspect(source)
	require.Equal(t, 0, len(p), "Expect 0 mounts actual %v mounts", len(p))

	err := m.Mount(0, source, dest, "", syscall.MS_BIND, "")
	require.NoError(t, err, "Failed in mount")
	p = m.Inspect(source)
	require.Equal(t, 1, len(p), "Expect 1 mounts actual %v mounts", len(p))
	require.Equal(t, dest, p[0].Path, "Expect %q got %q", dest, p[0].Path)
	err = m.Unmount(source, dest)
	require.NoError(t, err, "Failed in unmount")
}

func TestHasMounts(t *testing.T) {
	count := m.HasMounts(source)
	require.Equal(t, 0, count, "Expect 0 mounts actual %v mounts", count)

	mounts := 0
	for i := 0; i < 10; i++ {
		dir := fmt.Sprintf("%s%d", dest, i)
		cleandir(dir)
		err := m.Mount(0, source, dir, "", syscall.MS_BIND, "")
		require.NoError(t, err, "Failed in mount")
		mounts++
		count = m.HasMounts(source)
		require.Equal(t, mounts, count, "Expect %v mounts actual %v mounts", mounts, count)
	}
	for i := 5; i >= 0; i-- {
		dir := fmt.Sprintf("%s%d", dest, i)
		err := m.Unmount(source, dir)
		require.NoError(t, err, "Failed in unmount")
		mounts--
		count = m.HasMounts(source)
		require.Equal(t, mounts, count, "Expect %v mounts actual %v mounts", mounts, count)
	}

	for i := 9; i > 5; i-- {
		dir := fmt.Sprintf("%s%d", dest, i)
		err := m.Unmount(source, dir)
		require.NoError(t, err, "Failed in mount")
		mounts--
		count = m.HasMounts(source)
		require.Equal(t, mounts, count, "Expect %v mounts actual %v mounts", mounts, count)
	}
	require.Equal(t, mounts, 0, "Expect 0 mounts actual %v mounts", mounts)
}

func TestRefcounts(t *testing.T) {
	require.Equal(t, m.HasMounts(source) == 0, true, "Don't expect mounts in the beginning")
	for i := 0; i < 10; i++ {
		err := m.Mount(0, source, dest, "", syscall.MS_BIND, "")
		require.NoError(t, err, "Failed in mount")
		require.True(t, m.HasMounts(source) > 0, "Refcnt must be greater than zero")
	}
	for i := 9; i > 0; i-- {
		err := m.Unmount(source, dest)
		require.NoError(t, err, "Failed in unmount")
		require.True(t, m.HasMounts(source) > 0, "Refcnt must be greater than zero")
	}
	err := m.Unmount(source, dest)
	require.NoError(t, err, "Failed in unmount")
	require.Equal(t, m.HasMounts(source), 0, "Refcnt must go down to zero")
}

func TestExists(t *testing.T) {
	err := m.Mount(0, source, dest, "", syscall.MS_BIND, "")
	require.NoError(t, err, "Failed in mount")
	exists, _ := m.Exists(source, "foo")
	require.False(t, exists, "%q should not be mapped to foo", source)
	exists, _ = m.Exists(source, dest)
	require.True(t, exists, "%q should  be mapped to %q", source, dest)
	err = m.Unmount(source, dest)
	require.NoError(t, err, "Failed in unmount")
}

func TestShutdown(t *testing.T) {
	os.RemoveAll(dest)
	os.RemoveAll(source)
}
