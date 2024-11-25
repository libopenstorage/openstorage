package chattr

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testFile = "/tmp/osd-test"
)

func TestImmutable(t *testing.T) {
	// create a test file
	_, err := os.Create(testFile)
	require.NoError(t, err, "Unexpected error on create test file")

	// chattr +i
	err = AddImmutable(testFile)
	require.NoError(t, err, "Unexpected error on AddImmutable")

	// check if +i is added on the file
	isImmutable := IsImmutable(testFile)
	require.True(t, isImmutable, "Unexpected: Path is not immutable")

	// remove should fail
	err = os.RemoveAll(testFile)
	require.Error(t, err, "Expected an error on remove")

	// check if file still exists
	_, err = os.Stat(testFile)
	require.NoError(t, err, "Expected the file to be present")

	// chattr -i
	err = RemoveImmutable(testFile)
	require.NoError(t, err, "Unexpected error on RemoveImmutable")

	// check if +i is removed on the file
	isImmutable = IsImmutable(testFile)
	require.False(t, isImmutable, "Unexpected: Path is not mutable")

	// remove should succeed
	err = os.RemoveAll(testFile)
	require.NoError(t, err, "Unexpected an error on remove")

	// check if file still exists
	_, err = os.Stat(testFile)
	require.Error(t, err, "Expected the file to be removed")
	require.True(t, os.IsNotExist(err), "Unexpected error on remove")
}

func TestImmutable_UnsupportedFilesystem(t *testing.T) {

	const (
		imgFile   = "/tmp/chattr-test.img"
		ntfsMount = "/tmp/chattr-test-mount"
		fsType    = "ntfs"
	)
	mkfsCommand := "mkfs." + fsType
	if which(mkfsCommand) == mkfsCommand {
		t.Skip("Skipping test: " + mkfsCommand + " not found")
	}

	// Create a file to serve as a disk image
	err := exec.Command("dd", "if=/dev/zero", "of="+imgFile, "bs=1M", "count=100").Run()
	require.NoError(t, err, "Failed to create NTFS image file")

	defer func() {
		err = os.Remove(imgFile)
		require.NoError(t, err, "Failed to remove NTFS image file")
	}()

	// Format the file
	err = exec.Command(mkfsCommand, "-F", imgFile).Run()
	require.NoError(t, err, fmt.Sprintf("Failed to format %s image", fsType))

	// Create mount point
	err = os.MkdirAll(ntfsMount, 0755)
	require.NoError(t, err, "Failed to create mount point")

	defer func() {
		err = os.RemoveAll(ntfsMount)
		require.NoError(t, err, "Failed to remove mount point")
	}()

	// Mount the image
	err = exec.Command("mount", "-o", "loop", "-t", fsType, imgFile, ntfsMount).Run()
	require.NoError(t, err, "Failed to mount image")

	defer func() {
		err = exec.Command("umount", ntfsMount).Run()
		require.NoError(t, err, "Failed to unmount image")
	}()

	// Attempt to add immutable attribute to the mount point: should not error even though filesystem does not support it
	err = AddImmutable(ntfsMount)
	require.NoError(t, err, "Unexpected error on AddImmutable")

	// Check if the mount point is immutable
	isImmutable := IsImmutable(ntfsMount)
	require.False(t, isImmutable, "Mount point should not be immutable")
}
