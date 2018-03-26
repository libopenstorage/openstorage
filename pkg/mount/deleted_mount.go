// +build linux

package mount

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	// AllDevices designates all devices that show up as source.
	AllDevices = "AllDevices"
)

// deletedMounter loads mounts that show up as deleted in the mount table.
type deletedMounter struct {
	*deviceMounter
}

// NewDeletedMounter returns a new deletedMounter
func NewDeletedMounter(
	rootSubstring string,
	mountImpl MountImpl,
) (*deletedMounter, error) {

	devMounter, err := NewDeviceMounter([]string{"/dev/"}, mountImpl, nil, "")
	if err != nil {
		return nil, err
	}
	deletedMounts := make(DeviceMap)
	for k, v := range devMounter.mounts {
		if matchDeleted(rootSubstring, k) {
			deletedMounts[k] = v
		} else {
			for _, p := range v.Mountpoint {
				if matchDeleted(rootSubstring, p.Root) {
					addMountpoint(deletedMounts, k, p)
				}
			}
		}
	}
	devMounter.mounts = deletedMounts
	return &deletedMounter{deviceMounter: devMounter}, nil
}

func matchDeleted(rootSubstring, source string) bool {
	return strings.Contains(source, "deleted") &&
		(len(rootSubstring) == 0 || strings.Contains(source, rootSubstring))
}

func addMountpoint(dm DeviceMap, root string, mp *PathInfo) {
	info, ok := dm[root]
	if !ok {
		info = &Info{
			Device:     root,
			Mountpoint: make([]*PathInfo, 0),
		}
		dm[root] = info
	}
	info.Mountpoint = append(info.Mountpoint, mp)
}

// Unmount all deleted mounts.
func (m *deletedMounter) Unmount(
	sourcePath string,
	path string,
	flags int,
	timeout int,
	opts map[string]string,
) error {
	if sourcePath != AllDevices {
		return fmt.Errorf("DeletedMounter accepts only %v as sourcePath",
			AllDevices)
	}

	m.Lock()
	defer m.Unlock()

	failedUnmounts := make(DeviceMap)
	for k, v := range m.mounts {
		for _, p := range v.Mountpoint {
			logrus.Warnf("Unmounting deleted mount path %v->%v", k, p)
			if err := m.mountImpl.Unmount(p.Path, flags, timeout); err != nil {
				logrus.Warnf("Failed to unmount mount path %v->%v", k, p)
				addMountpoint(failedUnmounts, k, p)
			}
		}
	}
	m.mounts = failedUnmounts
	if len(m.mounts) > 0 {
		return fmt.Errorf("Not all paths could be unmounted")
	}

	return nil
}

func (m *deletedMounter) Mounts(sourcePath string) []string {
	m.Lock()
	defer m.Unlock()

	if sourcePath != AllDevices {
		logrus.Warnf("DeletedMounter accepts only %v as sourcePath",
			AllDevices)
		return nil
	}
	paths := make([]string, 0)
	for _, v := range m.mounts {
		for _, p := range v.Mountpoint {
			paths = append(paths, p.Path)
		}
	}

	return paths
}

func (m *deletedMounter) Mount(
	minor int,
	device string,
	path string,
	fs string,
	flags uintptr,
	data string,
	timeout int,
	opts map[string]string,
) error {

	return fmt.Errorf("deletedMounter.Mount is not supported")
}
