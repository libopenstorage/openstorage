// +build linux

package mount

import (
	"errors"
	"fmt"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

// Mangager defines the interface for keep track of volume driver mounts.
type Manager interface {
	// String representation of the mount table
	String() string
	// Load mount table for all devices that match this identifier
	Load(source string) error
	// Inspect mount table for specified source. ErrEnoent may be returned.
	Inspect(source string) []PathInfo
	// HasMounts determines returns the number of mounts for the source.
	HasMounts(source string) int
	// Exists returns true if the device is mounted at specified path.
	// returned if the device does not exists.
	Exists(source, path string) (bool, error)
	// Mount device at mountpoint or increment refcnt if device is already mounted
	// at specified mountpoint.
	Mount(minor int, device, path, fs string, flags uintptr, data string) error
	// Unmount device at mountpoint or decrement refcnt. If device has no
	// mountpoints left after this operation, it is removed from the matrix.
	// ErrEnoent is returned if the device or mountpoint for the device is not found.
	Unmount(source, path string) error
}

type MountType int

const (
	DeviceMount MountType = 1 << iota
	NFSMount
)

var (
	// ErrEnoent is returned for a non existent mount point
	ErrEnoent = errors.New("Mountpath is not mounted")
	// ErrEinval is returned is fields for an entry do no match
	// existing fields
	ErrEinval = errors.New("Invalid arguments for mount entry")
	// ErrUnsupported is returned for an operation or a mount type not suppored.
	ErrUnsupported = errors.New("Not supported")
)

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

// Mounter implements Ops and keeps track of active mounts for volume drivers.
type Mounter struct {
	sync.Mutex
	mounts DeviceMap
}

// String representation of Mounter
func (m *Mounter) String() string {
	return fmt.Sprintf("%#v", *m)
}

// Inspect mount table for device
func (m *Mounter) Inspect(devPath string) []PathInfo {
	m.Lock()
	defer m.Unlock()

	v, ok := m.mounts[devPath]
	if !ok {
		return []PathInfo{}
	}
	return v.Mountpoint
}

// HasMounts determines returns the number of mounts for the device.
func (m *Mounter) HasMounts(devPath string) int {
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
func (m *Mounter) Exists(devPath string, path string) (bool, error) {
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
func (m *Mounter) Mount(minor int, device, path, fs string, flags uintptr, data string) error {
	m.Lock()
	defer m.Unlock()

	info, ok := m.mounts[device]
	if !ok {
		info = &Info{
			Device:     device,
			Mountpoint: make([]PathInfo, 0),
			Minor:      minor,
			Fs:         fs,
		}
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
	m.mounts[device] = info
	return nil
}

// Unmount device at mountpoint or decrement refcnt. If device has no
// mountpoints left after this operation, it is removed from the matrix.
// ErrEnoent is returned if the device or mountpoint for the device is not found.
func (m *Mounter) Unmount(device, path string) error {
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

func New(mounterType MountType, identifier string) (Manager, error) {
	switch mounterType {
	case DeviceMount:
		return NewDeviceMounter(identifier)
	case NFSMount:
		return NewNFSMounter(identifier)
	}
	return nil, ErrUnsupported
}
