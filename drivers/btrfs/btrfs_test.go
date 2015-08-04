// +build linux

package btrfs

import (
	_ "os"
	_ "os/exec"
	_ "syscall"
	"testing"

	"github.com/libopenstorage/openstorage/drivers/test"
	"github.com/libopenstorage/openstorage/volume"
)

func TestAll(t *testing.T) {

	/*
		cmd := exec.Command("dd", "if=/dev/zero", "of=/tmp/x", "bs=1M", "count=100")
		err := cmd.Run()
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		cmd = exec.Command("/sbin/mkfs.btrfs", "/tmp/x")
		err = cmd.Run()
		if err != nil {
			t.Fatalf("Failed to create btrfs:%v", err)
		}
		err = os.MkdirAll("/tmp/btrfs_test", 0755)
		if err != nil {
			t.Fatalf("Failed to create mkdir: %v", err)
		}
		err = syscall.Mount("/tmp/x", "/tmp/btrfs_test", "btrfs", syscall.MS_NODEV, "")
		if err != nil {
			t.Fatalf("Failed to mount btrfs: %v", err)
		}
	*/
	_, err := volume.New(Name, volume.DriverParams{RootParam: "/tmp/btrfs_test"})
	if err != nil {
		t.Fatalf("Failed to initialize Driver: %v", err)
	}
	d, err := volume.Get(Name)
	if err != nil {
		t.Fatalf("Failed to initialize Volume Driver: %v", err)
	}
	ctx := test.NewContext(d)

	test.Run(t, ctx)
}
