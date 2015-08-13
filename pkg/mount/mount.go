// +build linux

package mount

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/mount"
)

// Ops defines the interface for keep track of volume driver mounts.
type Ops interface {
	// String representation of the mount table
	String() string
	// Load mount table for all devices that match the prefix. An empty prefix may
	// be provided.
	Load(devPrefix string) error
	// Inspect mount table for specified device. ErrEnoent may be returned.
	Inspect(device string) (Info, error)
	// HasMounts determines returns the number of mounts for the device.
	HasMounts(devPath string)
	// Exists returns true if the device is mounted at specified path.
	// returned if the device does not exists.
	Exists(device, path string) (bool, error)
	// Mount device at mountpoint or increment refcnt if device is already mounted
	// at specified mountpoint.
	Mount(Minor int32, device, path, fs string, flags uintptr, data string) error
	// Unmount device at mountpoint or decrement refcnt. If device has no
	// mountpoints left after this operation, it is removed from the matrix.
	// ErrEnoent is returned if the device or mountpoint for the device is not found.
	Unmount(device, path string) error
}

// DeviceMap map device name to Info
type DeviceMap map[string]*Info

// PathInfo is a reference counted path
type PathInfo struct {
	Path string
	ref  int
}

// Info per device
type Info struct {
	Device     string
	Minor      int
	Mountpoint []PathInfo
	Fs         string
}

// Matrix implements Ops and keeps track of active mounts for volume drivers.
type Matrix struct {
	sync.Mutex
	mounts DeviceMap
}

var (
	// ErrEnoent is returned for a non existent mount point
	ErrEnoent = errors.New("Mountpath is not mounted")
	// ErrEinval is returned is fields for an entry do no match
	// existing fields
	ErrEinval = errors.New("Invalid arguments for mount entry")
)

// New instance of Matrix
func New(devPrefix string) (*Matrix, error) {
	m := &Matrix{
		mounts: make(DeviceMap),
	}
	err := m.Load(devPrefix)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// HasMounts determines returns the number of mounts for the device.
func (m *Matrix) HasMounts(devPath string) int {
	m.Lock()
	defer m.Unlock()

	v, ok := m.mounts[devPath]
	if !ok {
		return 0
	}
	return len(v.Mountpoint)
}

// Exists scans mountpaths for specified device and returns true if path is one of the
// mountpaths. ErrEnoent may be retuned if the device is not found
func (m *Matrix) Exists(devPath string, path string) (bool, error) {
	m.Lock()
	defer m.Unlock()

	v, ok := m.mounts[devPath]
	if !ok {
		return false, ErrEnoent
	}
	for _, p := range v.Mountpoint {
		if p.Path == path {
			return true, nil
		}
	}
	return false, nil
}

// Mount new mountpoint for specified device.
func (m *Matrix) Mount(minor int, device, path, fs string, flags uintptr, data string) error {
	m.Lock()
	defer m.Unlock()

	info, ok := m.mounts[device]
	if !ok {
		info = &Info{
			Device:     device,
			Mountpoint: make([]PathInfo, 8),
			Minor:      minor,
			Fs:         fs,
		}
		m.mounts[device] = info
	}

	// Validate input params
	if fs != info.Fs {
		log.Warnf("%s Existing mountpoint has fs %q cannot change to %q",
			device, info.Fs, fs)
		return ErrEinval
	}

	// Try to find the mountpoint. If it already exists, then increment refcnt
	for _, p := range info.Mountpoint {
		if p.Path == path {
			p.ref++
			return nil
		}
	}
	// The device is not mounted at path, mount it and add to its mountpoints.
	err := syscall.Mount(device, path, fs, flags, data)
	if err != nil {
		return err
	}
	info.Mountpoint = append(info.Mountpoint, PathInfo{Path: path, ref: 1})
	return nil
}

// Unmount device at mountpoint or decrement refcnt. If device has no
// mountpoints left after this operation, it is removed from the matrix.
// ErrEnoent is returned if the device or mountpoint for the device is not found.
func (m *Matrix) Unmount(device, path string) error {
	m.Lock()
	defer m.Unlock()

	info, ok := m.mounts[device]
	if !ok {
		return ErrEnoent
	}
	for i, p := range info.Mountpoint {
		if p.Path == path {
			p.ref--
			// Unmount only if refcnt is 0
			if p.ref == 0 {
				err := syscall.Unmount(path, 0)
				if err != nil {
					return err
				}
				// Blow away this mountpoint.
				info.Mountpoint[i] = info.Mountpoint[len(info.Mountpoint)-1]
				info.Mountpoint = info.Mountpoint[0 : len(info.Mountpoint)-1]
				// If the device has no more mountpoints, remove it from the map
				if len(info.Mountpoint) == 0 {
					delete(m.mounts, device)
				}
			}
			return nil
		}
	}
	return ErrEnoent
}

// String representation of Matrix
func (m *Matrix) String() string {
	return fmt.Sprintf("%#v", *m)
}

// Load mount table
func (m *Matrix) Load(devPrefix string) error {
	info, err := mount.GetMounts()
	if err != nil {
		return err
	}
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
				continue
			}
		}
		// XXX Reconstruct refs.
		mount.Mountpoint = append(mount.Mountpoint, PathInfo{Path: v.Mountpoint, ref: 1})
	}
	return nil
}
