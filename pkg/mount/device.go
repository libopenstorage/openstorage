// +build linux

package mount

import (
	"strings"

	"github.com/docker/docker/pkg/mount"
)

// DeviceMounter implements Ops and keeps track of active mounts for volume drivers.
type DeviceMounter struct {
	Mounter
}

// NewDeviceMounter
func NewDeviceMounter(devPrefix string) (*DeviceMounter, error) {
	m := &DeviceMounter{
		Mounter: Mounter{mounts: make(DeviceMap), paths: make(PathMap)},
	}
	err := m.Load(devPrefix)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Load mount table
func (m *DeviceMounter) Load(devPrefix string) error {
	info, err := mount.GetMounts()
	if err != nil {
		return err
	}
DeviceLoop:
	for _, v := range info {
		if !strings.HasPrefix(v.Source, devPrefix) {
			continue
		}
		mount, ok := m.mounts[v.Source]
		if !ok {
			mount = &Info{
				Device:     v.Source,
				Fs:         v.Fstype,
				Minor:      v.Minor,
				Mountpoint: make([]PathInfo, 0),
			}
			m.mounts[v.Source] = mount
		}
		// Allow Load to be called multiple times.
		for _, p := range mount.Mountpoint {
			if p.Path == v.Mountpoint {
				continue DeviceLoop
			}
		}
		// XXX Reconstruct refs.
		mount.Mountpoint = append(mount.Mountpoint, PathInfo{Path: v.Mountpoint, ref: 1})
		m.paths[v.Mountpoint] = v.Source
	}
	return nil
}
