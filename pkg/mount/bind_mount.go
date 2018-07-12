// +build linux

package mount

import (
	"strings"

	"github.com/docker/docker/pkg/mount"
	"github.com/libopenstorage/openstorage/pkg/keylock"
)

const (
	sharedMount = "shared"
)

// bindMounter loads mounts that are bind mounted in the mount table
type bindMounter struct {
	Mounter
}

// NewBindMounter returns a new bindMounter
func NewBindMounter(
	rootSubstrings []string,
	mountImpl MountImpl,
	allowedDirs []string,
	trashLocation string,
) (*bindMounter, error) {
	b := &bindMounter{
		Mounter: Mounter{
			mountImpl:     mountImpl,
			mounts:        make(DeviceMap),
			paths:         make(PathMap),
			allowedDirs:   allowedDirs,
			kl:            keylock.New(),
			trashLocation: trashLocation,
		},
	}
	if err := b.Load(rootSubstrings); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *bindMounter) Reload(rootSubstring string) error {
	newBm, err := NewBindMounter(
		[]string{rootSubstring},
		b.mountImpl,
		b.allowedDirs,
		b.trashLocation,
	)
	if err != nil {
		return err
	}
	return b.reload(rootSubstring, newBm.mounts[rootSubstring])
}

func (b *bindMounter) Load(rootSubstrings []string) error {
	return b.load(rootSubstrings, bindFindMountPoint)
}

func bindFindMountPoint(sInfo *mount.Info, destination string, infos []*mount.Info) (bool, string, string) {
	for _, dInfo := range infos {
		if !strings.Contains(dInfo.Mountpoint, destination) {
			continue
		}
		// Check if the root device is the same for the bind mount
		if dInfo.Root != sInfo.Root {
			continue
		}
		if !strings.Contains(dInfo.Optional, sharedMount) ||
			!strings.Contains(sInfo.Optional, sharedMount) {
			continue
		}
		// Get the mount peer group
		sPeerGroup := strings.Split(sInfo.Optional, sharedMount)[1]
		dPeerGroup := strings.Split(dInfo.Optional, sharedMount)[1]
		if sPeerGroup == dPeerGroup {
			return true, dInfo.Mountpoint, dInfo.Source
		}
	}
	return false, "", ""
}
