package mount

import (
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func bindMount(src, dest string) error {
	return syscall.Mount(src, dest, "", syscall.MS_BIND, "")
}

func makeMounts(t *testing.T, srcBase, dstBase string, srcCount, dstCount int) {
	for i := 0; i < srcCount; i++ {
		src := fmt.Sprintf("%s.%d", srcBase, i)
		cleandir(src)
		for j := 0; j < dstCount; j++ {
			dst := fmt.Sprintf("%s.%d.%d", dstBase, i, j)
			os.MkdirAll(dst, 0755)
			require.NoError(t, bindMount(src, dst), "bindMount")
		}
	}

}

func cleanup(srcBase, dstBase string, srcCount, dstCount int) {
	for i := 0; i < srcCount; i++ {
		src := fmt.Sprintf("%s.%d", srcBase, i)
		cleandir(src)
		for j := 0; j < dstCount; j++ {
			dst := fmt.Sprintf("%s.%d.%d", dstBase, i, j)
			cleandir(dst)
		}
	}
}

func testLoadUnmount(t *testing.T, srcBase, dstBase string, srcCount, dstCount int) {
	cleanup(srcBase, dstBase, srcCount, dstCount)
	makeMounts(t, srcBase, dstBase, srcCount, dstCount)
	dm, err := NewDeletedMounter("/mnt", &DefaultMounter{})
	require.NoError(t, err, "NewDeletedMount")
	mounts := len(dm.Mounts(AllDevices))
	require.Equal(t, srcCount*dstCount, mounts,
		fmt.Sprintf("%v", dm.Mounts(AllDevices)))
	testUnmount(t, dm)
	cleanup(srcBase, dstBase, srcCount, dstCount)
}

func failUnmount(t *testing.T, srcBase, dstBase string, srcCount, dstCount int) {
	cleanup(srcBase, dstBase, srcCount, dstCount)
	makeMounts(t, srcBase, dstBase, srcCount, dstCount)
	dm, err := NewDeletedMounter("/mnt/", &DefaultMounter{})
	require.NoError(t, err, "NewDeletedMount")
	mounts := dm.Mounts(AllDevices)
	for i, m := range mounts {
		if (i % 2) == 0 {
			syscall.Unmount(m, 0)
		}
	}
	require.Error(t, dm.Unmount(AllDevices, "", 0, 0, nil), "dm.Unmount")
	mounts = dm.Mounts(AllDevices)
	require.NotEqual(t, srcCount*dstCount, len(mounts),
		fmt.Sprintf("%v", dm.Mounts(AllDevices)))
}

func testUnmount(t *testing.T, dm *deletedMounter) {
	require.NoError(t, dm.Unmount(AllDevices, "", 0, 0, nil), "dm.Unmount")
	mounts := len(dm.Mounts(AllDevices))
	require.Equal(t, mounts, 0, "Non zero mounts after Unmount")
}

func TestLoadUnmount(t *testing.T) {
	srcBase := "/mnt/test_mount_deleted"
	dstBase := "/mnt/dest_bind_mount"

	cleanup(srcBase, dstBase, 10, 10)

	testLoadUnmount(t, srcBase, dstBase, 1, 1)
	testLoadUnmount(t, srcBase, dstBase, 1, 3)
	testLoadUnmount(t, srcBase, dstBase, 5, 1)
	testLoadUnmount(t, srcBase, dstBase, 5, 3)
	testLoadUnmount(t, srcBase, dstBase, 5, 5)

	cleanup(srcBase, dstBase, 10, 10)
}

func TestFailUnmount(t *testing.T) {
	srcBase := "/mnt/test_mount_deleted"
	dstBase := "/mnt/dest_bind_mount"

	cleanup(srcBase, dstBase, 10, 10)

	failUnmount(t, srcBase, dstBase, 1, 3)
	failUnmount(t, srcBase, dstBase, 5, 1)
	failUnmount(t, srcBase, dstBase, 5, 3)
	failUnmount(t, srcBase, dstBase, 5, 5)

	cleanup(srcBase, dstBase, 10, 10)
}
