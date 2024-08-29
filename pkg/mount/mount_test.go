package mount

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/pkg/options"
	"github.com/libopenstorage/openstorage/pkg/sched"
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

func setLogger(fn string, t *testing.T) {
	// The mount tests log a lot of messages, so we route the logs
	// to a tmp location to avoid Travis CI log limits.
	logFile, err := os.Create("/tmp/" + fn + ".log")
	require.NoError(t, err, "unable to create log file")
	logrus.SetOutput(logFile)
}
func TestNFSMounterHandleDNSResolution(t *testing.T) {
	setLogger("TestNFSMounterHandleDNSResolution", t)
	setupNFS(t, true)
	allTests(t, source, dest)
}

func TestNFSMounter(t *testing.T) {
	setLogger("TestNFSMounter", t)
	setupNFS(t, false)
	allTests(t, source, dest)
}

func TestBindMounter(t *testing.T) {
	setLogger("TestBindMounter", t)
	setupBindMounter(t)
	allTests(t, source, dest)
}

func TestRawMounter(t *testing.T) {
	setLogger("TestRawMounter", t)
	setupRawMounter(t)
	allTests(t, rawSource, rawDest)
}

func allTests(t *testing.T, source, dest string) {
	load(t, source, dest)
	mountTest(t, source, dest)
	enoentUnmountTest(t, source, dest)
	doubleUnmountTest(t, source, dest)
	enoentUnmountTestWithoutOptions(t, source, dest)
	doubleMountTest(t, source, dest)
	mountTestHostMismatchFailure(t, source, dest)
	mountTestHostMismatchSuccessWithOptions(t, source, dest)
	mountTestPathMismatchFailureWithOptions(t, source, dest)
	mountTestParallel(t, source, dest)
	inspect(t, source, dest)
	reload(t, source, dest)
	hasMounts(t, source, dest)
	refcounts(t, source, dest)
	exists(t, source, dest)
	shutdown(t, source, dest)
}

func setupNFS(t *testing.T, handleDNSResolution bool) {
	var err error
	m, err = New(NFSMount, nil, []*regexp.Regexp{regexp.MustCompile("")}, nil, []string{}, trashLocation, handleDNSResolution)
	if err != nil {
		t.Fatalf("Failed to setup test %v", err)
	}
	cleandir(source)
	cleandir(dest)
}

func setupBindMounter(t *testing.T) {
	var err error
	m, err = New(BindMount, nil, []*regexp.Regexp{regexp.MustCompile("")}, nil, []string{}, trashLocation, false)
	if err != nil {
		t.Fatalf("Failed to setup test %v", err)
	}
	cleandir(source)
	cleandir(dest)
}

func setupRawMounter(t *testing.T) {
	var err error
	m, err = New(RawMount, nil, []*regexp.Regexp{regexp.MustCompile("")}, nil, []string{}, trashLocation, false)
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
	require.NoError(t, m.Load([]*regexp.Regexp{regexp.MustCompile("")}), "Failed in load")
}

func mountTest(t *testing.T, source, dest string) {
	err := m.Mount(0, source, dest, "", syscall.MS_BIND, "", 0, nil)
	require.NoError(t, err, "Failed in mount")
	err = m.Unmount(source, dest, 0, 0, nil)
	require.NoError(t, err, "Failed in unmount")
}

func doubleMountTest(t *testing.T, source, dest string) {
	err := m.Mount(0, source, dest, "", syscall.MS_BIND, "", 0, nil)
	require.NoError(t, err, "Failed in mount")

	// Mount point is already created and new request lands on the same mount point
	err = m.Mount(0, source, dest, "", syscall.MS_BIND, "", 0, nil)
	require.NoError(t, err, "Unexpected error in mount")

	err = m.Unmount(source, dest, 0, 0, nil)
	require.NoError(t, err, "Failed in unmount")
}

func mountTestHostMismatchFailure(t *testing.T, source, dest string) {
	cleandir("localhost:" + source)
	cleandir("127.0.0.1:" + source)
	err := m.Mount(0, "localhost:"+source, dest, "", syscall.MS_BIND, "", 0, nil)
	require.NoError(t, err, "Failed in mount")

	// Mount point is already created and new request lands on the same mount point
	// but source paths are different
	err = m.Mount(0, "127.0.0.1:"+source, dest, "", syscall.MS_BIND, "", 0, nil)
	// Expected error as source paths are different
	require.Error(t, err, "Expected error in mount")
	require.Equal(t, err.Error(), "Mountpath already exists", "Expected \"Mountpath already exists\"")

	err = m.Unmount("localhost:"+source, dest, 0, 0, nil)
	require.NoError(t, err, "Failed in unmount")
	shutdown(t, "localhost:"+source, dest)
	shutdown(t, "127.0.0.1:"+source, dest)
}

func mountTestHostMismatchSuccessWithOptions(t *testing.T, source, dest string) {
	opts := make(map[string]string)
	opts[options.OptionsResolveDNSOnMount] = "true"
	cleandir("localhost:" + source)
	cleandir("127.0.0.1:" + source)
	err := m.Mount(0, "localhost:"+source, dest, "", syscall.MS_BIND, "", 0, opts)
	require.NoError(t, err, "Failed in mount")

	// Mount point is already created and new request lands on the same mount point
	// and source paths resolve to same IP with OptionsResolveDNSOnMount
	err = m.Mount(0, "127.0.0.1:"+source, dest, "", syscall.MS_BIND, "", 0, opts)
	// Expected success as source paths are different but resolve to same IP
	require.NoError(t, err, "Failed in mount")

	err = m.Unmount("localhost:"+source, dest, 0, 0, nil)
	require.NoError(t, err, "Failed in unmount")
	shutdown(t, "localhost:"+source, dest)
	shutdown(t, "127.0.0.1:"+source, dest)
}

func mountTestPathMismatchFailureWithOptions(t *testing.T, source, dest string) {
	opts := make(map[string]string)
	opts[options.OptionsResolveDNSOnMount] = "true"
	cleandir("localhost:" + source + "/path1")
	cleandir("localhost:" + source + "/path2")
	err := m.Mount(0, "localhost:"+source+"/path1", dest, "", syscall.MS_BIND, "", 0, nil)
	require.NoError(t, err, "Failed in mount")

	// Mount point is already created and new request lands on the same mount point
	// but source paths are different even when NFS server is same.
	// Unlikely in practice.
	err = m.Mount(0, "localhost:"+source+"/path2", dest, "", syscall.MS_BIND, "", 0, nil)
	// Expected error as source paths are different
	require.Error(t, err, "Expected error in mount")
	require.Equal(t, err.Error(), "Mountpath already exists", "Expected \"Mountpath already exists\"")

	err = m.Unmount("localhost:"+source+"/path1", dest, 0, 0, nil)
	require.NoError(t, err, "Failed in unmount")
	shutdown(t, "localhost:"+source+"/path1", dest)
	shutdown(t, "localhost:"+source+"/path2", dest)
}

func enoentUnmountTest(t *testing.T, source, dest string) {
	opts := make(map[string]string)
	opts[options.OptionsUnmountOnEnoent] = "true"
	syscall.Mount(source, dest, "", syscall.MS_BIND, "")
	err := m.Unmount(source, dest, 0, 0, opts)
	require.NoError(t, err, "Failed in unmount")
}

func doubleUnmountTest(t *testing.T, source, dest string) {
	err := m.Mount(0, source, dest, "", syscall.MS_BIND, "", 0, nil)
	require.NoError(t, err, "Failed in mount")
	err = m.Unmount(source, dest, 0, 0, nil)
	require.NoError(t, err, "Failed in unmount")
	err = m.Unmount(source, dest, 0, 0, nil)
	require.Error(t, err, "Failed in second unmount, expected an error")
}

func enoentUnmountTestWithoutOptions(t *testing.T, source, dest string) {
	syscall.Mount(source, dest, "", syscall.MS_BIND, "")
	err := m.Unmount(source, dest, 0, 0, nil)
	require.Error(t, err, "Failed in unmount, expected an error")
	syscall.Unmount(dest, 0)
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

func TestResolveToIPs(t *testing.T) {
	tests := []struct {
		hostPath string
		expected []string
	}{
		// Case: Valid hostname with path
		{"localhost:/path", []string{"127.0.0.1"}},

		// Case: Valid hostname without path
		{"localhost", []string{"127.0.0.1"}},

		// Case: Invalid hostname
		{"invalidhost", []string{"invalidhost"}},

		// Case: IP address with path
		{"192.168.1.1:/path", []string{"192.168.1.1"}},

		// Case: IP address without path
		{"192.168.1.1", []string{"192.168.1.1"}},

		// Case: Empty string
		{"", []string{""}},
	}
	for _, test := range tests {
		result := resolveToIPs(test.hostPath)
		if !areSameIPs(result, test.expected) {
			t.Errorf("resolveToIPs(%v) = %v; expected %v", test.hostPath, result, test.expected)
		}
	}
}
func TestAreSameIPs(t *testing.T) {
	tests := []struct {
		name     string
		ips1     []string
		ips2     []string
		expected bool
	}{
		{
			name:     "One matching IP",
			ips1:     []string{"192.168.1.1", "192.168.1.2"},
			ips2:     []string{"10.0.0.1", "192.168.1.2"},
			expected: true,
		},
		{
			name:     "No matching IPs",
			ips1:     []string{"192.168.1.1", "192.168.1.2"},
			ips2:     []string{"10.0.0.1", "10.0.0.2"},
			expected: false,
		},
		{
			name:     "Empty second slice",
			ips1:     []string{"192.168.1.1", "192.168.1.2"},
			ips2:     []string{},
			expected: false,
		},
		{
			name:     "Empty first slice",
			ips1:     []string{},
			ips2:     []string{"192.168.1.1", "192.168.1.2"},
			expected: false,
		},
		{
			name:     "Both slices empty",
			ips1:     []string{},
			ips2:     []string{},
			expected: false,
		},
		{
			name:     "Identical IPs in both slices",
			ips1:     []string{"192.168.1.1", "192.168.1.2"},
			ips2:     []string{"192.168.1.1", "192.168.1.2"},
			expected: true,
		},
		{
			name:     "Multiple matches",
			ips1:     []string{"192.168.1.1", "10.0.0.1"},
			ips2:     []string{"192.168.1.1", "10.0.0.1"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := areSameIPs(tt.ips1, tt.ips2)
			if result != tt.expected {
				t.Errorf("areSameIPs(%v, %v) = %v; expected %v", tt.ips1, tt.ips2, result, tt.expected)
			}
		})
	}
}

func TestExtractSourcePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single colon with valid suffix",
			input:    "a:/b",
			expected: "/b",
		},
		{
			name:     "Colon at the beginning",
			input:    ":/path",
			expected: "/path",
		},
		{
			name:     "Multiple colons in string",
			input:    "path:/to:/resource",
			expected: "/resource",
		},
		{
			name:     "Colon at the end",
			input:    "path:/",
			expected: "/",
		},
		{
			name:     "No colon in string",
			input:    "noColonHere",
			expected: "noColonHere",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Colon only",
			input:    ":",
			expected: ":",
		},
		{
			name:     "Colon followed by space",
			input:    "path: /to/resource",
			expected: " /to/resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSourcePath(tt.input)
			if result != tt.expected {
				t.Errorf("extractSourcePath(%q) = %q; expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSafeEmptyTrashDir(t *testing.T) {
	sched.Init(time.Second)
	m, err := New(NFSMount, nil, []*regexp.Regexp{regexp.MustCompile("")}, nil, []string{}, "", true)
	require.NoError(t, err, "Failed to setup test %v", err)

	err = os.MkdirAll("/tmp/safe-empty-trash-dir-tests", 0755)
	require.NoError(t, err)

	defer func() {
		err = os.RemoveAll("/tmp/safe-empty-trash-dir-tests")
		require.NoError(t, err, "Failed to cleanup after test")
	}()

	// Create files that should not be removed
	file, err := os.Create("/tmp/safe-empty-trash-dir-tests/should-not-remove.txt")
	require.NoError(t, err, "Failed to create file: %v", err)
	file.Close()

	// Create a symbolic link that should not be removed
	err = os.Symlink("/tmp/safe-empty-trash-dir-tests/should-not-remove.txt", "/tmp/safe-empty-trash-dir-tests/should-not-remove-symlink.txt")
	require.NoError(t, err, "Failed to create symlink: %v", err)

	// Create a file that should be removed
	file, err = os.Create("/tmp/safe-empty-trash-dir-tests/should-remove-file.txt")
	require.NoError(t, err, "Failed to create file: %v", err)

	file.Close()

	// Create a symbolic link
	err = os.Symlink("/tmp/safe-empty-trash-dir-tests/should-remove-file.txt", "/tmp/safe-empty-trash-dir-tests/should-remove-symlink.txt")
	require.NoError(t, err, "Failed to create symlink: %v", err)

	err = m.SafeEmptyTrashDir("/tmp/safe-empty-trash-dir-tests/should-remove", "/tmp/safe-empty-trash-dir-tests")
	require.NoError(t, err, "Failed to empty trash dir %v", err)

	time.Sleep(mountPathRemoveDelay + 5*time.Second)

	_, err = os.Stat("/tmp/safe-empty-trash-dir-tests/should-remove-file.txt")
	require.True(t, os.IsNotExist(err), "File should be removed")
	_, err = os.Stat("/tmp/safe-empty-trash-dir-tests/should-remove-symlink.txt")
	require.True(t, os.IsNotExist(err), "File should be removed")
	_, err = os.Stat("/tmp/safe-empty-trash-dir-tests/should-not-remove.txt")
	require.NoError(t, err, "File should not be removed")
	_, err = os.Stat("/tmp/safe-empty-trash-dir-tests/should-not-remove-symlink.txt")
	require.NoError(t, err, "File should not be removed")
}
