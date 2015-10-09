// +build linux

package btrfs

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/libopenstorage/openstorage/drivers/test"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	btrfsFile = "/var/btrfs"
	testPath  = "/var/test_dir"
)

func TestSetup(t *testing.T) {
	umount, err := which("umount")
	if err != nil {
		t.Fatal(err)
	}
	mount, err := which("mount")
	if err != nil {
		t.Fatal(err)
	}
	mkfs, err := which("mkfs")
	if err != nil {
		t.Fatal(err)
	}
	exec.Command(umount, btrfsFile).Output()
	os.Remove(btrfsFile)
	os.MkdirAll(testPath, 0755)

	f, err := os.Create(btrfsFile)
	if err != nil {
		t.Fatalf("Failed to setup btrfs store: %v", err)
	}
	err = f.Truncate(int64(8) << 30)
	if err != nil {
		t.Fatalf("Failed to truncate /var/btrfs 1G  %v", err)
	}
	o, err := exec.Command(mkfs, "-t", "btrfs", "-f", btrfsFile).Output()
	if err != nil {
		t.Fatalf("Failed to format to btrfs: %v: %v", err, o)
	}

	o, err = exec.Command(mount, btrfsFile, testPath).Output()
	if err != nil {
		t.Fatalf("Failed to mount to btrfs: %v: %v", err, o)
	}
}

func TestAll(t *testing.T) {
	_, err := volume.New(Name, volume.DriverParams{RootParam: testPath})
	if err != nil {
		t.Fatalf("Failed to initialize Driver: %v", err)
	}
	d, err := volume.Get(Name)
	if err != nil {
		t.Fatalf("Failed to initialize Volume Driver: %v", err)
	}
	ctx := test.NewContext(d)
	ctx.Filesystem = "btrfs"

	test.Run(t, ctx)
}

func which(executable string) (string, error) {
	output, err := exec.Command("which", executable).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
