package mount

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	source        = "/mnt/ost/mount_test_src"
	dest          = "/mnt/ost/mount_test_dest"
	trashLocation = "/tmp/.trash"
	rawSource     = "/mnt/rawtests/test_src_raw"
	rawDest       = "/mnt/rawtests/test_dest_raw"
)

var m Manager

func TestNFSMounter(t *testing.T) {
	setupNFS(t)
	allTests(t, source, dest)
}

func TestBindMounter(t *testing.T) {
	setupBindMounter(t)
	allTests(t, source, dest)
}

func TestRawMounter(t *testing.T) {
	setupRawMounter(t)
	allTests(t, rawSource, rawDest)
}

func allTests(t *testing.T, source, dest string) {
	load(t, source, dest)
	mountTest(t, source, dest)
	mountTestParallel(t, source, dest)
	inspect(t, source, dest)
	reload(t, source, dest)
	hasMounts(t, source, dest)
	refcounts(t, source, dest)
	exists(t, source, dest)
	shutdown(t, source, dest)
}

func setupNFS(t *testing.T) {
	var err error
	m, err = New(NFSMount, nil, []string{""}, nil, []string{}, trashLocation)
	if err != nil {
		t.Fatalf("Failed to setup test %v", err)
	}
	cleandir(source)
	cleandir(dest)
}

func setupBindMounter(t *testing.T) {
	var err error
	m, err = New(BindMount, nil, []string{""}, nil, []string{}, trashLocation)
	if err != nil {
		t.Fatalf("Failed to setup test %v", err)
	}
	cleandir(source)
	cleandir(dest)
}

func setupRawMounter(t *testing.T) {
	var err error
	m, err = New(RawMount, nil, []string{""}, nil, []string{}, trashLocation)
	if err != nil {
		t.Fatalf("Failed to setup test %v", err)
	}
	cleandir(rawSource)
	cleandir(rawDest)
}

func cleandir(dir string) {
	syscall.Unmount(dir, 0)
	os.RemoveAll(dir)
	os.MkdirAll(dir, 0755)
}

func load(t *testing.T, source, dest string) {
	require.NoError(t, m.Load([]string{""}), "Failed in load")
}

func mountTest(t *testing.T, source, dest string) {
	err := m.Mount(0, source, dest, "", syscall.MS_BIND, "", 0, nil)
	require.NoError(t, err, "Failed in mount")
	err = m.Unmount(source, dest, 0, 0, nil)
	require.NoError(t, err, "Failed in unmount")
}

// mountTestParallel runs mount and unmount in parallel with serveral dirs
// in addition, we trigger failed unmount to test race condition in the case
// source directory is not found in the cache
func mountTestParallel(t *testing.T, source, dest string) {
	mountFunc := func(s, d string) {
		err := m.Mount(0, s, d, "", syscall.MS_BIND, "", 0, nil)
		require.NoError(t, err, "Failed in mount")
	}
	unmountFunc := func(s, d string) {
		err := m.Unmount(s, d, 0, 0, nil)
		require.NoError(t, err, "Failed in unmount")
	}
	unmountFailedFunc := func(s, d string) {
		err := m.Unmount(s, d, 0, 0, nil)
		require.Error(t, err, "Failed in unmount; expected an error")
	}
	numRuns := 200
	var wg sync.WaitGroup
	for i := 1; i < numRuns; i++ {
		wg.Add(1)
		s := fmt.Sprintf("%s_%d", source, i)
		d := fmt.Sprintf("%s_%d", dest, i)
		random_s := fmt.Sprintf("%s__%d", source, i)
		cleandir(s)
		cleandir(d)
		go func() {
			mountFunc(s, d)
			unmountFunc(s, d)
			unmountFailedFunc(random_s, d)
			defer wg.Done()
		}()
	}
	wg.Wait()

}

func inspect(t *testing.T, source, dest string) {
	p := m.Inspect(source)
	require.Equal(t, 0, len(p), "Expect 0 mounts actual %v mounts", len(p))

	err := m.Mount(0, source, dest, "", syscall.MS_BIND, "", 0, nil)
	require.NoError(t, err, "Failed in mount")
	p = m.Inspect(source)
	require.Equal(t, 1, len(p), "Expect 1 mounts actual %v mounts", len(p))
	require.Equal(t, dest, p[0].Path, "Expect %q got %q", dest, p[0].Path)
	s := m.GetSourcePaths()
	require.NotZero(t, 1, len(s), "Expect 1 source path, actual %v", s)
	err = m.Unmount(source, dest, 0, 0, nil)
	require.NoError(t, err, "Failed in unmount")
}

func reload(t *testing.T, source, dest string) {
	p := m.Inspect(source)
	require.Equal(t, 0, len(p), "Expect 0 mounts actual %v mounts", len(p))

	err := m.Mount(0, source, dest, "", syscall.MS_BIND, "", 0, nil)
	require.NoError(t, err, "Failed in mount")

	syscall.Unmount(dest, 0)
	p = m.Inspect(source)
	require.Equal(t, 1, len(p), "Expect 1 mounts actual %v mounts", len(p))
	require.Equal(t, dest, p[0].Path, "Expect %q got %q", dest, p[0].Path)
	require.NoError(t, m.Reload(source), "Reload mounts")
	p = m.Inspect(source)
	require.Equal(t, 0, len(p), "Expect 0 mounts actual %v mounts", len(p))
}

func hasMounts(t *testing.T, source, dest string) {
	count := m.HasMounts(source)
	require.Equal(t, 0, count, "Expect 0 mounts actual %v mounts", count)

	mounts := 0
	for i := 0; i < 10; i++ {
		dir := fmt.Sprintf("%s%d", dest, i)
		cleandir(dir)
		err := m.Mount(0, source, dir, "", syscall.MS_BIND, "", 0, nil)
		require.NoError(t, err, "Failed in mount")
		mounts++
		count = m.HasMounts(source)
		require.Equal(t, mounts, count, "Expect %v mounts actual %v mounts", mounts, count)
	}
	for i := 5; i >= 0; i-- {
		dir := fmt.Sprintf("%s%d", dest, i)
		err := m.Unmount(source, dir, 0, 0, nil)
		require.NoError(t, err, "Failed in unmount")
		mounts--
		count = m.HasMounts(source)
		require.Equal(t, mounts, count, "Expect %v mounts actual %v mounts", mounts, count)
	}

	for i := 9; i > 5; i-- {
		dir := fmt.Sprintf("%s%d", dest, i)
		err := m.Unmount(source, dir, 0, 0, nil)
		require.NoError(t, err, "Failed in mount")
		mounts--
		count = m.HasMounts(source)
		require.Equal(t, mounts, count, "Expect %v mounts actual %v mounts", mounts, count)
	}
	require.Equal(t, mounts, 0, "Expect 0 mounts actual %v mounts", mounts)
}

func refcounts(t *testing.T, source, dest string) {
	require.Equal(t, m.HasMounts(source) == 0, true, "Don't expect mounts in the beginning")
	for i := 0; i < 10; i++ {
		err := m.Mount(0, source, dest, "", syscall.MS_BIND, "", 0, nil)
		require.NoError(t, err, "Failed in mount")
		require.Equal(t, m.HasMounts(source), 1, "Refcnt must be one")
	}

	err := m.Unmount(source, dest, 0, 0, nil)
	require.NoError(t, err, "Failed in unmount")
	require.Equal(t, m.HasMounts(source), 0, "Refcnt must go down to zero")

	err = m.Unmount(source, dest, 0, 0, nil)
	require.Error(t, err, "Unmount should fail")
}

func exists(t *testing.T, source, dest string) {
	err := m.Mount(0, source, dest, "", syscall.MS_BIND, "", 0, nil)
	require.NoError(t, err, "Failed in mount")
	exists, _ := m.Exists(source, "foo")
	require.False(t, exists, "%q should not be mapped to foo", source)
	exists, _ = m.Exists(source, dest)
	require.True(t, exists, "%q should  be mapped to %q", source, dest)
	err = m.Unmount(source, dest, 0, 0, nil)
	require.NoError(t, err, "Failed in unmount")
}

func shutdown(t *testing.T, source, dest string) {
	os.RemoveAll(dest)
	os.RemoveAll(source)
}

func makeFile(pathname string) error {
	f, err := os.OpenFile(pathname, os.O_CREATE, os.FileMode(0644))
	defer func() {
		err := f.Close()
		if err != nil {
			logrus.Warnf("failed to close file: %s", err.Error())
		}
	}()
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	return nil
}
