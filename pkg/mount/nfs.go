//go:build linux
// +build linux

package mount

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/libopenstorage/openstorage/pkg/keylock"
	"github.com/moby/sys/mountinfo"
)

const (
	// NFSAllServers is a catch all for all servers.
	NFSAllServers = "NFSAllServers"
)

// nfsMounter implements Manager and keeps track of active mounts for volume drivers.
type nfsMounter struct {
	servers             []*regexp.Regexp
	handleDNSResolution bool
	Mounter
}

// NewnfsMounter returns a Mounter specific to parse NFS mounts. This can work
// VFS also if 'servers' is nil. Use NFSAllServers if the destination server
// is unknown.
func NewNFSMounter(servers []*regexp.Regexp,
	mountImpl MountImpl,
	allowedDirs []string,
	trashLocation string,
	handleDNSResolution bool,
) (Manager, error) {
	m := &nfsMounter{
		servers:             servers,
		handleDNSResolution: handleDNSResolution,
		Mounter: Mounter{
			mountImpl:     mountImpl,
			mounts:        make(DeviceMap),
			paths:         make(PathMap),
			allowedDirs:   allowedDirs,
			kl:            keylock.New(),
			trashLocation: trashLocation,
		},
	}
	err := m.Load([]*regexp.Regexp{}) // Input value is not used, can be anything
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Reload reloads the mount table for the specified source/
func (m *nfsMounter) Reload(inputSource string) error {
	newNFSm, err := NewNFSMounter([]*regexp.Regexp{regexp.MustCompile(NFSAllServers)},
		m.mountImpl,
		m.Mounter.allowedDirs,
		m.trashLocation,
		m.handleDNSResolution,
	)
	if err != nil {
		return err
	}

	newNFSmounter, ok := newNFSm.(*nfsMounter)
	if !ok {
		return fmt.Errorf("Internal error failed to convert %T",
			newNFSmounter)
	}
	newM := newNFSmounter.mounts[inputSource]

	if m.handleDNSResolution && newM == nil {

		// Check if the source is a IP:share combination which maps to an
		// DNS:share combination.
		resolvedInputSourceIPs := resolveToIPs(inputSource)
		inputSourceExportPath := extractSourcePath(inputSource)
		for existingSource, existingMountInfo := range newNFSmounter.mounts {
			if inputSourceExportPath == extractSourcePath(existingSource) {
				// Found a match for the same export path.
				// Now lets check if the input source (IP or DNS) matches
				// with the existing source (IP or DNS).
				resolvedExistingSourceIPs := resolveToIPs(existingSource)
				if areSameIPs(resolvedExistingSourceIPs, resolvedInputSourceIPs) {
					// The input source and existing source are the same.
					// So, we can use the existing mount info even if it was for a
					// a DNS representation of the server and our input source is an IP.
					newM = existingMountInfo
					break
				}
			}
		}

	}
	return m.reload(inputSource, newM)
}

// serverExists utility function to test if a server is part of driver config
func (m *nfsMounter) serverExists(server string) bool {
	for _, v := range m.servers {
		vStr := v.String()
		if vStr == server || vStr == NFSAllServers {
			return true
		}
	}
	return false
}

// normalizeSource - NFS source is returned as IP:share or just :share
// normalize that to always IP:share
func (m *nfsMounter) normalizeSource(info *mountinfo.Info, host string) {
	if info.FSType != "nfs" {
		return
	}

	s := strings.Split(info.Source, ":")
	if len(s) == 2 && len(s[0]) == 0 {
		if net.ParseIP(host) != nil { // Check for IPv6 IP Address
			if strings.Contains(host, ":") && !strings.Contains(host, "[") {
				host = fmt.Sprintf("[%s]", host)
			}
		}
		info.Source = host + info.Source
	}
}

// Load mount table
func (m *nfsMounter) Load(source []*regexp.Regexp) error {
	info, err := GetMounts()
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`,addr=(.*)`)
MountLoop:
	for _, v := range info {
		host := "localhost"
		if len(m.servers) != 0 {
			if !strings.HasPrefix(v.FSType, "nfs") {
				continue
			}
			matches := re.FindStringSubmatch(v.VFSOptions)
			if len(matches) != 2 {
				continue
			}

			if exists := m.serverExists(matches[1]); !exists {
				continue
			}
			host = matches[1]
		}
		m.normalizeSource(v, host)
		mount, ok := m.mounts[v.Source]
		if !ok {
			mount = &Info{
				Device:     v.Source,
				Fs:         v.FSType,
				Minor:      v.Minor,
				Mountpoint: make([]*PathInfo, 0),
			}
			m.mounts[v.Source] = mount
		}
		// Allow Load to be called multiple times.
		for _, p := range mount.Mountpoint {
			if p.Path == v.Mountpoint {
				continue MountLoop
			}
		}
		pi := &PathInfo{
			Path: normalizeMountPath(v.Mountpoint),
		}
		mount.Mountpoint = append(mount.Mountpoint,
			pi,
		)
	}
	return nil
}
