// +build linux

package mount

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/docker/docker/pkg/mount"
	"github.com/libopenstorage/openstorage/pkg/keylock"
	"github.com/sirupsen/logrus"
)

const (
	// NFSAllServers is a catch all for all servers.
	NFSAllServers = "NFSAllServers"
)

// nfsMounter implements Manager and keeps track of active mounts for volume drivers.
type nfsMounter struct {
	servers []string
	Mounter
}

// NewnfsMounter returns a Mounter specific to parse NFS mounts. This can work
// VFS also if 'servers' is nil. Use NFSAllServers if the destination server
// is unknown.
func NewNFSMounter(servers []string,
	mountImpl MountImpl,
	allowedDirs []string,
) (Manager, error) {
	m := &nfsMounter{
		servers: servers,
		Mounter: Mounter{
			mountImpl:   mountImpl,
			mounts:      make(DeviceMap),
			paths:       make(PathMap),
			allowedDirs: allowedDirs,
			kl:          keylock.New(),
		},
	}
	err := m.Load([]string{""})
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Reload reloads the mount table for the specified source/
func (m *nfsMounter) Reload(source string) error {
	newNFSm, err := NewNFSMounter([]string{NFSAllServers},
		m.mountImpl,
		m.Mounter.allowedDirs,
	)
	if err != nil {
		return err
	}
	m.LogDevices()

	newNFSmounter, ok := newNFSm.(*nfsMounter)
	if !ok {
		return fmt.Errorf("Internal error failed to convert %T",
			newNFSmounter)
	}

	err = m.reload(source, newNFSmounter.mounts[source])
	m.LogDevices()
	return err
}

//serverExists utility function to test if a server is part of driver config
func (m *nfsMounter) serverExists(server string) bool {
	for _, v := range m.servers {
		if v == server || v == NFSAllServers {
			return true
		}
	}
	return false
}

// normalizeSource - NFS source is returned as IP:share or just :share
// normalize that to always IP:share
func (m *nfsMounter) normalizeSource(info *mount.Info, host string) {
	if info.Fstype != "nfs" {
		return
	}
	s := strings.Split(info.Source, ":")
	if len(s) == 2 && len(s[0]) == 0 {
		info.Source = host + info.Source
	}
}

// Load mount table
func (m *nfsMounter) Load(source []string) error {
	logrus.Trace("Entered nfsMounter.Load()")
	info, err := mount.GetMounts()
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`,addr=(.*)`)
MountLoop:
	for i, v := range info {
		logrus.Tracef("Info[%d] = %v", i, *v)
		host := "localhost"
		if len(m.servers) != 0 {
			if !strings.HasPrefix(v.Fstype, "nfs") {
				continue
			}
			matches := re.FindStringSubmatch(v.VfsOpts)
			logrus.Tracef("RegEx match[%d] = %v", i, matches)
			if len(matches) != 2 {
				continue
			}

			if exists := m.serverExists(matches[1]); !exists {
				logrus.Tracef("Server does not exists [%d]", i)
				continue
			}
			host = matches[1]
		}
		m.normalizeSource(v, host)
		logrus.Tracef("Normalized info[%d] = %v", i, *v)
		mount, ok := m.mounts[v.Source]
		if !ok {
			mount = &Info{
				Device:     v.Source,
				Fs:         v.Fstype,
				Minor:      v.Minor,
				Mountpoint: make([]*PathInfo, 0),
			}
			m.mounts[v.Source] = mount
			logrus.Tracef("Could not get mount. Assigned: m.mounts[%s] = %v, %v, %v, %v", v.Source, mount.Device, mount.Fs, mount.Minor, mount.Mountpoint)
		}
		// Allow Load to be called multiple times.
		for _, p := range mount.Mountpoint {
			if p.Path == v.Mountpoint {
				logrus.Tracef("Continue to MountLoop %s == %s", p.Path, v.Mountpoint)
				continue MountLoop
			}
		}
		pi := &PathInfo{
			Path: normalizeMountPath(v.Mountpoint),
		}
		logrus.Tracef("Adding pathInfo to MountPoints: %v", *pi)
		mount.Mountpoint = append(mount.Mountpoint,
			pi,
		)
	}
	return nil
}

func (m *nfsMounter) LogDevices() {
	mnts := make(map[string]string)
	logrus.Tracef("Logging NFS device mounts")
	for _, device := range m.mounts {
		mnts = map[string]string{}
		for _, mntPoint := range device.Mountpoint {
			mnts[mntPoint.Root] = mntPoint.Root
		}
		logrus.Tracef("Device: %s MountPoints: [%v]", device.Device, mnts)
	}
}
