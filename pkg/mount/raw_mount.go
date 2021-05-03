// +build linux

package mount

import (
	"fmt"
	"strings"

	"github.com/docker/docker/pkg/mount"
	"github.com/libopenstorage/openstorage/pkg/keylock"
)

var (
	skippedLocations = []string{"/proc", "/null"}
)

// rawMounter loads mounts for raw volumes that are bind mounted in the mount table
type rawMounter struct {
	Mounter
}

// NewRawBindMounter returns a new rawBindMounter
func NewRawBindMounter(
	rootSubstrings []string,
	mountImpl MountImpl,
	allowedDirs []string,
	trashLocation string,
) (*rawMounter, error) {
	rm := &rawMounter{
		Mounter: Mounter{
			mountImpl:     mountImpl,
			mounts:        make(DeviceMap),
			paths:         make(PathMap),
			allowedDirs:   allowedDirs,
			kl:            keylock.New(),
			trashLocation: trashLocation,
		},
	}
	if err := rm.Load(rootSubstrings); err != nil {
		return nil, err
	}
	return rm, nil
}

func (rm *rawMounter) Reload(rootSubstring string) error {
	newRBM, err := NewRawBindMounter(
		[]string{rootSubstring},
		rm.mountImpl,
		rm.allowedDirs,
		rm.trashLocation,
	)
	if err != nil {
		return err
	}
	return rm.reload(rootSubstring, newRBM.mounts[rootSubstring])
}

func shouldSkipMountPoint(mountPoint string) bool {
	for _, skip := range skippedLocations {
		if strings.HasPrefix(mountPoint, skip) {
			return true
		}
	}

	return false
}

// this mount filtering implementation is done based on logic implemented in findmnt + libmount
func (rm *rawMounter) Load(rawVolumeDevicesPaths []string) error {
	mountPoints, err := GetMounts()
	if err != nil {
		return err
	}

	// try to find all bind mounts of raw volumes
	if len(rawVolumeDevicesPaths) == 0 || rawVolumeDevicesPaths[0] == "" {
		mountPointsByMajMin := make(map[string]*[]mount.Info)
		mountPointsByTarget := make(map[string]*mount.Info)

		for _, mp := range mountPoints {
			// skip proc mount points
			if shouldSkipMountPoint(mp.Mountpoint) || mp.Root == "/null" {
				continue
			}
			majMin := fmt.Sprintf("%v:%v", mp.Major, mp.Minor)

			mountPointsForNumber, exists := mountPointsByMajMin[majMin]
			if !exists {
				mountPointsForNumber = &[]mount.Info{}
				mountPointsByMajMin[majMin] = mountPointsForNumber
			}

			*mountPointsForNumber = append(*mountPointsForNumber, *mp)

			mountPointsByTarget[mp.Mountpoint] = mp
		}

		devicesMP, exists := mountPointsByTarget["/dev"]
		if !exists {
			return fmt.Errorf("cannot find /dev mount point while loading mount points")
		}

		devMajMin := fmt.Sprintf("%v:%v", devicesMP.Major, devicesMP.Minor)

		mountPointsForNumber := mountPointsByMajMin[devMajMin]

		mps := *mountPointsForNumber
		filteredMPs := []mount.Info{}
		for i := range mps {
			if mps[i].Mountpoint != "/dev" {
				filteredMPs = append(filteredMPs, mps[i])
			}
		}

		for _, mp := range filteredMPs {
			devicePath := "/dev" + mp.Root

			// source for raw volumes is equal to rawVolumeDevicePath
			mount, ok := rm.mounts[devicePath]
			if !ok {
				mount = &Info{
					Device: devicePath,
					Fs:     "",
					Minor:  mp.Minor,
					Mountpoint: []*PathInfo{&PathInfo{
						Root: normalizeMountPath(devicePath),
						Path: normalizeMountPath(mp.Mountpoint),
					}},
				}
				rm.mounts[devicePath] = mount
			}
		}
	}

	// find raw volume bind mounts
	for _, rawVolumeDevicePath := range rawVolumeDevicesPaths {
		var mountPointForRoot *mount.Info
		for _, mp := range mountPoints {
			if strings.HasSuffix(rawVolumeDevicePath, mp.Root) {
				mountPointForRoot = mp
				break
			}
		}

		if mountPointForRoot == nil {
			continue
		}

		devicePath := "/dev" + mountPointForRoot.Root

		// source for raw volumes is equal to rawVolumeDevicePath
		mount, ok := rm.mounts[devicePath]
		if !ok {
			mount = &Info{
				Device: devicePath,
				Fs:     "",
				Minor:  mountPointForRoot.Minor,
				Mountpoint: []*PathInfo{&PathInfo{
					Root: normalizeMountPath(devicePath),
					Path: normalizeMountPath(mountPointForRoot.Mountpoint),
				}},
			}
			rm.mounts[devicePath] = mount
		}
	}

	return nil
}
