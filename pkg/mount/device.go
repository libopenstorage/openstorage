// +build linux

package mount

import (
	"os"
	"path/filepath"
	"regexp"
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
	devRegexes []*regexp.Regexp,
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
	err := m.Load(devRegexes)
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
	newDm, err := NewDeviceMounter([]*regexp.Regexp{regexp.MustCompile(regexp.QuoteMeta(device))},
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
func (m *deviceMounter) Load(devRegexes []*regexp.Regexp) error {
	return m.load(devRegexes, deviceFindMountPoint)
}

func deviceFindMountPoint(info *mount.Info, destination *regexp.Regexp, infos []*mount.Info) (bool, string, string) {
	if destination.MatchString(info.Source) {
		return true, info.Source, info.Source
	}
	return false, "", ""
}

// getTargetDevice follows the symbolic link for the provided devPath and gets the target device
// For ex. /dev/mapper/vg-lv1 -> /dev/dm-1
func getTargetDevice(devPath string) string {
	fi, err := os.Lstat(devPath)
	if err == nil &&
		(fi.Mode()&os.ModeSymlink != 0) {
		// Found a symbolic link reference
		out, err := filepath.EvalSymlinks(devPath)
		if err != nil {
			// We do a best effort to find the target
			return ""
		}
		return strings.TrimSpace(string(out))

	} // else we are given a prefix
	return ""
}
