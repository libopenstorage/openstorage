// +build linux

package mount

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/docker/docker/pkg/mount"
	"github.com/libopenstorage/openstorage/pkg/keylock"
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
			traceCache:  []string{},
		},
	}
	err := m.Load([]string{""})
	if err != nil {
		m.LogTraceCache(err)
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

	// Check if source existed in the mounts table prior to reload
	_, sourceExisted := m.mounts[source]
	err = m.reload(source, newNFSmounter.mounts[source])

	m.LogDevices()

	// only log trace cache if source existed and is no longer present
	if sourceExisted && m.mounts[source] == nil {
		newNFSmounter.LogTraceCache(fmt.Errorf("Did not find source in mount table"))
	}

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
	m.traceCache = append(m.traceCache, "Entered nfsMounter.Load()")
	info, err := mount.GetMounts()
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`,addr=(.*)`)
MountLoop:
	for i, v := range info {
		m.traceCache = append(m.traceCache, fmt.Sprintf("Info[%d] = %v", i, *v))
		host := "localhost"
		if len(m.servers) != 0 {
			if !strings.HasPrefix(v.Fstype, "nfs") {
				continue
			}
			matches := re.FindStringSubmatch(v.VfsOpts)
			m.traceCache = append(m.traceCache, fmt.Sprintf("RegEx match[%d] = %v", i, matches))
			if len(matches) != 2 {
				continue
			}

			if exists := m.serverExists(matches[1]); !exists {
				m.traceCache = append(m.traceCache, fmt.Sprintf("Server does not exists [%d]", i))
				continue
			}
			host = matches[1]
		}
		m.normalizeSource(v, host)
		m.traceCache = append(m.traceCache, fmt.Sprintf("Normalized info[%d] = %v", i, *v))
		mount, ok := m.mounts[v.Source]
		if !ok {
			mount = &Info{
				Device:     v.Source,
				Fs:         v.Fstype,
				Minor:      v.Minor,
				Mountpoint: make([]*PathInfo, 0),
			}
			m.mounts[v.Source] = mount
			m.traceCache = append(m.traceCache, fmt.Sprintf("Could not get mount. Assigned: m.mounts[%s] = %v, %v, %v, %v", v.Source, mount.Device, mount.Fs, mount.Minor, mount.Mountpoint))
		}
		// Allow Load to be called multiple times.
		for _, p := range mount.Mountpoint {
			if p.Path == v.Mountpoint {
				m.traceCache = append(m.traceCache, fmt.Sprintf("Continue to MountLoop %s == %s", p.Path, v.Mountpoint))
				continue MountLoop
			}
		}
		pi := &PathInfo{
			Path: normalizeMountPath(v.Mountpoint),
		}
		m.traceCache = append(m.traceCache, fmt.Sprintf("Adding pathInfo to MountPoints: %v", *pi))
		mount.Mountpoint = append(mount.Mountpoint,
			pi,
		)
	}
	return nil
}

func (m *nfsMounter) LogDevices() {
	mnts := make(map[string]string)
	m.traceCache = append(m.traceCache, fmt.Sprintf("Logging NFS device mounts"))
	for _, device := range m.mounts {
		mnts = map[string]string{}
		for _, mntPoint := range device.Mountpoint {
			mnts[mntPoint.Root] = mntPoint.Root
		}
		m.traceCache = append(m.traceCache, fmt.Sprintf("Device: %s MountPoints: [%v]", device.Device, mnts))
	}
}
