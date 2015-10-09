package mount

import (
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
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
	syscall.Unmount(dest, 0)
	os.RemoveAll(dest)
	os.RemoveAll(source)
	os.MkdirAll(source, 0755)
	os.MkdirAll(dest, 0755)
}

func TestLoad(t *testing.T) {
	assert.NoError(t, m.Load(""), "Failed in load")
}

func TestMount(t *testing.T) {
	err := m.Mount(0, source, dest, "", syscall.MS_BIND, "")
	assert.NoError(t, err, "Failed in mount")
}

func TestInspect(t *testing.T) {
	p := m.Inspect("foo")
	assert.Equal(t, 0, len(p), "Expect 0 mounts actual %v mounts", len(p))
	p = m.Inspect(source)
	require.Equal(t, 1, len(p), "Expect 1 mounts actual %v mounts", len(p))
	assert.Equal(t, dest, p[0].Path, "Expect %q got %q", dest, p[0].Path)
}

func TestHasMounts(t *testing.T) {
	count := m.HasMounts("foo")
	assert.Equal(t, 0, count, "Expect 0 mounts actual %v mounts", count)
	count = m.HasMounts(source)
	assert.Equal(t, 1, count, "Expect 1 mounts actual %v mounts", count)
}

func TestExists(t *testing.T) {
	exists, _ := m.Exists(source, "foo")
	assert.False(t, exists, "%q should not be mapped to foo", source)
	exists, _ = m.Exists(source, dest)
	assert.True(t, exists, "%q should  be mapped to %q", source, dest)
}

func TestShutdown(t *testing.T) {
	os.RemoveAll(dest)
	os.RemoveAll(source)
}
