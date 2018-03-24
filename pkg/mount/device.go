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
	info, err := mount.GetMounts()
	if err != nil {
		return err
	}
DeviceLoop:
	for _, v := range info {
		foundPrefix := false
		for _, devPrefix := range devPrefixes {
			if strings.HasPrefix(v.Source, devPrefix) {
				foundPrefix = true
				break
			}
		}
		if !foundPrefix {
			continue
		}
		mount, ok := m.mounts[v.Source]
		if !ok {
			mount = &Info{
				Device:     v.Source,
				Fs:         v.Fstype,
				Minor:      v.Minor,
				Mountpoint: make([]*PathInfo, 0),
			}
			m.mounts[v.Source] = mount
		}
		// Allow Load to be called multiple times.
		for _, p := range mount.Mountpoint {
			if p.Path == v.Mountpoint {
				continue DeviceLoop
			}
		}
		mount.Mountpoint = append(
			mount.Mountpoint,
			&PathInfo{
				Root: normalizeMountPath(v.Root),
				Path: normalizeMountPath(v.Mountpoint),
			},
		)
		m.paths[v.Mountpoint] = v.Source
	}
	return nil
}
