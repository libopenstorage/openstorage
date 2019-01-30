/*
Package util provides utility functions for OSD servers and drivers.
Copyright 2017 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package util

import (
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestIsSameFilesystem(t *testing.T) {

	if u, err := user.Current(); err != nil || u.Uid != "0" {
		t.Skipf("Test reqires ROOT user -- skipping (uid=%v, err=%s)", u.Uid, err)
	}

	testData := []struct {
		input     []string
		expectRes bool
		expectErr string
	}{
		{[]string{"/"}, false, "2 arguments"},
		{[]string{"/", "/"}, true, ""},
		{[]string{"/", "/", "/", "/", "/", "/"}, true, ""},
		{[]string{"/", "/proc/self/mountinfo"}, false, ""},
		{[]string{"/", "/_dont_really_exist"}, false, "Could not stat"},
		{[]string{"/_dont_really_exist", "/"}, false, "Could not stat"},
	}

	for i, td := range testData {
		resB, err := IsSameFilesystem(td.input...)
		if td.expectErr != "" {
			// Test is expecting error here -- let's check if we got a correct error
			require.Error(t, err)
			assert.Contains(t, err.Error(), td.expectErr,
				"was expecting error with %s for test-entry #%d - %q", td.expectErr, i+1, td.input)
		} else {
			// Test is expecting regular result here
			assert.NoError(t, err,
				"was NOT expecting error for test-entry #%d - %q", i+1, td.input)
			assert.Equal(t, td.expectRes, resB,
				"broken expectation for test-entry #%d - %q", i+1, td.input)
		}
	}

	// Let's also prepare our bind-mount
	tdir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	mpath := path.Join(tdir, "anoter", "more", "3rd-subdir", "mount")
	defer func() {
		unix.Unmount(mpath, 0)
		os.RemoveAll(tdir)
	}()

	err = os.MkdirAll(mpath, 0755)
	require.NoError(t, err)
	flags := uintptr(syscall.MS_NOATIME | syscall.MS_SILENT | syscall.MS_NODEV | syscall.MS_NOEXEC | syscall.MS_NOSUID)
	err = unix.Mount("tmpfs", mpath, "tmpfs", flags, "size=5")
	require.NoError(t, err)

	// Validate mountpoint
	resB, err := IsSameFilesystem(mpath, path.Dir(mpath))
	assert.NoError(t, err)
	assert.False(t, resB)

	// .. expect same
	resB, err = IsSameFilesystem(path.Dir(mpath), mpath)
	assert.NoError(t, err)
	assert.False(t, resB)

	// .. expect no mountpoint
	resB, err = IsSameFilesystem(path.Dir(mpath), path.Dir(path.Dir(mpath)))
	assert.NoError(t, err)
	assert.True(t, resB)

	// .. expect false on mixed (applies to _all_ params on the same device)
	resB, err = IsSameFilesystem(mpath, path.Dir(mpath), path.Dir(path.Dir(mpath)), "/")
	assert.NoError(t, err)
	assert.False(t, resB)

	// Test IsMountpoint() while we have a mounted RAMDISK

	resB, err = IsMountpoint(mpath)
	assert.NoError(t, err)
	assert.True(t, resB)

	resB, err = IsMountpoint(path.Dir(mpath))
	assert.NoError(t, err)
	assert.False(t, resB)
}
