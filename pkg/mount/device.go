// +build linux

package mount

import (
	"os"
	"strings"

	"github.com/docker/docker/pkg/mount"
	"github.com/libopenstorage/openstorage/pkg/keylock"
)

// deviceMounter implements Ops and tracks active mounts for volume drivers.
type deviceMounter struct {
	Mounter
}

// NewDeviceMounter returns a new deviceMounter
func NewDeviceMounter(
	devPrefixes []string,
	mountImpl MountImpl,
	allowedDirs []string,
	trashLocation string,
) (*deviceMounter, error) {

	m := &deviceMounter{
		Mounter: Mounter{
			mountImpl:     mountImpl,
			mounts:        make(DeviceMap),
			paths:         make(PathMap),
			allowedDirs:   allowedDirs,
			kl:            keylock.New(),
			trashLocation: trashLocation,
		},
	}
	err := m.Load(devPrefixes)
	if err != nil {
		return nil, err
	}

	if len(m.trashLocation) > 0 {
		if err := os.MkdirAll(m.trashLocation, 0755); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// Reload reloads the mount table
func (m *deviceMounter) Reload(device string) error {
	newDm, err := NewDeviceMounter([]string{device},
		m.mountImpl,
		m.Mounter.allowedDirs,
		m.trashLocation,
	)
	if err != nil {
		return err
	}
	return m.reload(device, newDm.mounts[device])
}

// Load mount table
func (m *deviceMounter) Load(devPrefixes []string) error {
	return m.load(devPrefixes, deviceFindMountPoint)
}

func deviceFindMountPoint(info *mount.Info, destination string, infos []*mount.Info) (bool, string, string) {
	if strings.HasPrefix(info.Source, destination) {
		return true, info.Source, info.Source
	}
	return false, "", ""
}
