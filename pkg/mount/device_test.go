package mount

import (
	"os"
	"testing"

	"github.com/docker/docker/pkg/mount"
	"github.com/stretchr/testify/require"
)

const (
	osdDir          = "/dev/osd/"
	osdDevicePrefix = osdDir + "osd"
	mountDir        = "/var/lib/osd/mounts/test/"
	dummyDevice     = osdDevicePrefix + "dummy-device"
	dummyPath       = mountDir + "dummy-device"
)

func addMountEntry(t *testing.T, device string, mountpoint string) int {
	_, err := os.Create(device)
	require.NoError(t, err, "Failed to create test device: ", device)
	index := len(testMounts)
	info := &mount.Info{
		Fstype:     "ext4",
		Minor:      index,
		Mountpoint: mountpoint,
		Source:     device,
	}
	testMounts = append(testMounts, info)
	return index
}

func removeMountEntries(t *testing.T, noOfEntries int) {
	testMounts = testMounts[:len(testMounts)-noOfEntries]
}

func setupDeviceMount(t *testing.T) {
	cleanupDeviceMount(t)
	os.Setenv(testDeviceEnv, "true")
	err := os.MkdirAll(osdDir, 0777)
	require.NoError(t, err, "Unexpected error in setup")
	err = os.MkdirAll(mountDir, 0777)
	require.NoError(t, err, "Unexpected error in setup")
	// Initialize the mounts
	_, err = testGetMounts()
	require.NoError(t, err, "Unexpected error in setup")
}

func cleanupDeviceMount(t *testing.T) {
	os.Unsetenv(testDeviceEnv)
	err := os.RemoveAll(osdDir)
	require.NoError(t, err, "Unexpected error on RemoveAll")
	err = os.RemoveAll(mountDir)
	require.NoError(t, err, "Unexpected error on RemoveAll")
}

func TestBasicDeviceMounter(t *testing.T) {
	// Setup
	setupDeviceMount(t)

	device1 := osdDevicePrefix + "dev1"
	mountPath1 := mountDir + "dev1"
	addMountEntry(t, device1, mountPath1)

	orderedDevices := []string{device1}
	orderedPaths := []string{mountPath1}

	dm, err := New(DeviceMount, nil, []string{osdDevicePrefix}, nil, []string{}, "")
	require.NoError(t, err, "Unexpected error on mount.New")

	// Inspect
	testInspect(t, device1, orderedPaths, dm)

	// Mounts
	testMountsAPI(t, device1, orderedPaths, dm)

	// HasMounts
	testHasMounts(t, device1, 1, dm)

	// HasTarget
	testHasTarget(t, mountPath1, dm)

	// Exists
	testExists(t, orderedDevices, orderedPaths, dm)

	// GetSourcePath
	testGetSourcePath(t, device1, mountPath1, dm)

	removeMountEntries(t, 1)
	cleanupDeviceMount(t)
}

func TestBasicDeviceMounterWithMultipleMounts(t *testing.T) {
	// Setup
	setupDeviceMount(t)

	device1 := osdDevicePrefix + "dev1"
	mountPath1 := mountDir + "dev1"
	addMountEntry(t, device1, mountPath1)

	// Add another mountpoint for the same device
	mountPath2 := mountDir + "dev2"
	addMountEntry(t, device1, mountPath2)

	orderedDevices := []string{device1, device1}
	orderedPaths := []string{mountPath1, mountPath2}

	dm, err := New(DeviceMount, nil, []string{osdDevicePrefix}, nil, []string{}, "")
	require.NoError(t, err, "Unexpected error on mount.New")

	// Inspect
	testInspect(t, device1, orderedPaths, dm)

	// Mounts
	testMountsAPI(t, device1, orderedPaths, dm)

	// HasMounts
	testHasMounts(t, device1, 2, dm)

	// HasTarget
	testHasTarget(t, mountPath1, dm)
	testHasTarget(t, mountPath2, dm)

	// Exists
	testExists(t, orderedDevices, orderedPaths, dm)

	// GetSourcePath
	testGetSourcePath(t, device1, mountPath1, dm)
	testGetSourcePath(t, device1, mountPath2, dm)

	removeMountEntries(t, 2)
	cleanupDeviceMount(t)
}

/*
// This test has been set to skip due to issues in Travis.
func TestBasicDeviceMounterWithSymLinks(t *testing.T) {
	setupDeviceMount(t)

	device1 := osdDevicePrefix + "dev1"
	mountPath1 := mountDir + "dev1"
	// Add the mount entry for the target device -> /dev/dm-0
	addMountEntry(t, device1, mountPath1)

	// Create the symbolic link for the target device -> /dev/mapper/...

	symDevice1 := osdDevicePrefix + "sdev1"
	err := os.Symlink(device1, symDevice1)
	require.NoError(t, err, "Expected no error on Symlink")

	orderedDevices := []string{symDevice1}
	orderedPaths := []string{mountPath1}

	// Pass in the symbolic link to the Device Mounter
	dm, err := New(DeviceMount, nil, []string{symDevice1}, nil, []string{}, "")
	require.NoError(t, err, "Unexpected error on mount.New")

	// Inspect
	testInspect(t, symDevice1, orderedPaths, dm)
	testInspect(t, device1, orderedPaths, dm)

	// Mounts
	testMountsAPI(t, device1, orderedPaths, dm)
	testMountsAPI(t, symDevice1, orderedPaths, dm)

	// HasMounts
	testHasMounts(t, device1, 1, dm)
	testHasMounts(t, symDevice1, 1, dm)

	// HasTarget
	testHasTarget(t, mountPath1, dm)

	// Exists
	testExists(t, orderedDevices, orderedPaths, dm)

	// GetSourcePath
	testGetSourcePath(t, symDevice1, mountPath1, dm)

	removeMountEntries(t, 1)
	cleanupDeviceMount(t)
}
*/

func testExists(t *testing.T, orderedDevices, orderedPaths []string, dm Manager) {
	for i, device := range orderedDevices {
		mountPath := orderedPaths[i]
		exists, err := dm.Exists(device, mountPath)
		require.NoError(t, err, "Unexpected error on Exists")
		require.True(t, exists, "Expected device and mountpath to exist")

		exists, err = dm.Exists(device, "/mnt")
		require.NoError(t, err, "Unexpected error from Exists")
		require.False(t, exists, "Unepxected output from Exists")
	}
	exists, err := dm.Exists(dummyDevice, dummyPath)
	require.EqualError(t, ErrEnoent, err.Error(), "Expected an ErrEnoent from Exists")
	require.False(t, exists, "Unexpected output from Exists")
}

func testHasTarget(t *testing.T, mountPath string, dm Manager) {
	_, hasTarget := dm.HasTarget(mountPath)
	require.True(t, hasTarget, "Expected a target mountpoint")

	_, hasTarget = dm.HasTarget(dummyPath)
	require.False(t, hasTarget, "Unexpected target mountpoint")
}

func testHasMounts(t *testing.T, device string, numberOfMounts int, dm Manager) {
	hasMounts := dm.HasMounts(device)
	require.Equal(t, numberOfMounts, hasMounts, "Unexpected number of mountpoints")

	hasMounts = dm.HasMounts(dummyDevice)
	require.Equal(t, 0, hasMounts, "Unexpected number of mountpoints")

}

func testMountsAPI(t *testing.T, device string, expectedMounts []string, dm Manager) {
	mounts := dm.Mounts(device)
	require.Equal(t, len(expectedMounts), len(mounts), "Unexpected number of mounts")

	for _, actualMount := range mounts {
		foundMount := false
		for _, expectedMount := range expectedMounts {
			if actualMount == expectedMount {
				foundMount = true
				break
			}
		}
		require.True(t, foundMount, "Could not find mountpoint for device: ", device)
	}
}

func testInspect(t *testing.T, device string, expectedMounts []string, dm Manager) {
	paths := dm.Inspect(device)
	require.Equal(t, len(expectedMounts), len(paths), "Unexpected number of mounts")

	for _, pathInfo := range paths {
		foundMount := false
		for _, expectedMount := range expectedMounts {
			if pathInfo.Path == expectedMount {
				foundMount = true
				break
			}
		}
		require.True(t, foundMount, "Could not find mountpoint for device: ", device)
	}
}

func testGetSourcePath(t *testing.T, expectedSourcePath, mountPath string, dm Manager) {
	sourcePath, err := dm.GetSourcePath(mountPath)
	require.NoError(t, err, "Unexpected error on GetSourcePath")
	require.Equal(t, expectedSourcePath, sourcePath, "Unepxected sourcePath from GetSourcePath")

	sourcePath, err = dm.GetSourcePath(dummyPath)
	require.EqualError(t, ErrEnoent, err.Error(), "Expected an ErrEnoent from GetSourcePath")
	require.Equal(t, "", sourcePath, "Unexpected sourcePath from GetSourcePath")
}
