// +build linux

package mount

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"syscall"

	"go.pedge.io/dlog"
)

// Manager defines the interface for keep track of volume driver mounts.
type Manager interface {
	// String representation of the mount table
	String() string
	// Reload mount table for specified device.
	Reload(source string) error
	// Load mount table for all devices that match this identifier
	Load(source string) error
	// Inspect mount table for specified source. ErrEnoent may be returned.
	Inspect(source string) []*PathInfo
	// Mounts returns paths for specified source.
	Mounts(source string) []string
	// HasMounts determines returns the number of mounts for the source.
	HasMounts(source string) int
	// HasTarget determines returns the number of mounts for the target.
	HasTarget(target string) (string, bool)
	// Exists returns true if the device is mounted at specified path.
	// returned if the device does not exists.
	Exists(source, path string) (bool, error)
	// GetSourcePath scans mount for a specified mountPath and returns the
	// sourcePath if found or returnes an ErrEnoent
	GetSourcePath(mountPath string) (string, error)
	// GetSourcePaths returns all source paths from the mount table
	GetSourcePaths() []string
	// Mount device at mountpoint or increment refcnt if device is already
	//  mounted at specified mountpoint.
	Mount(minor int, device, path, fs string, flags uintptr, data string, timeout int) error
	// Unmount device at mountpoint or decrement refcnt. If device has no
	// mountpoints left after this operation, it is removed from the matrix.
	// ErrEnoent is returned if the device or mountpoint for the device
	// is not found.
	Unmount(source, path string, timeout int) error
}

// MountImpl backend implementation for Mount/Unmount calls
type MountImpl interface {
	Mount(source, target, fstype string, flags uintptr, data string, timeout int) error
	Unmount(target string, flags int, timeout int) error
}

// MountType indicates different mount types supported
type MountType int

const (
	// DeviceMount indicates a device mount type
	DeviceMount MountType = 1 << iota
	// NFSMount indicates a NFS mount point
	NFSMount
)

var (
	// ErrExist is returned if path is already mounted to a different device.
	ErrExist = errors.New("Mountpath already exists")
	// ErrEnoent is returned for a non existent mount point
	ErrEnoent = errors.New("Mountpath is not mounted")
	// ErrEinval is returned is fields for an entry do no match
	// existing fields
	ErrEinval = errors.New("Invalid arguments for mount entry")
	// ErrUnsupported is returned for an unsupported operation or a mount type.
	ErrUnsupported = errors.New("Not supported")
)

// DeviceMap map device name to Info
type DeviceMap map[string]*Info

// PathMap map path name to device
type PathMap map[string]string

// PathInfo is a reference counted path
type PathInfo struct {
	Path string
	ref  int
}

// Info per device
type Info struct {
	sync.Mutex
	Device     string
	Minor      int
	Mountpoint []*PathInfo
	Fs         string
}

// Mounter implements Ops and keeps track of active mounts for volume drivers.
type Mounter struct {
	sync.Mutex
	mountImpl MountImpl
	mounts    DeviceMap
	paths     PathMap
}

// DefaultMounter defaults to syscall implementation.
type DefaultMounter struct {
}

// Mount default mount implementation is syscall.
func (m *DefaultMounter) Mount(
	source string,
	target string,
	fstype string,
	flags uintptr,
	data string,
	timeout int,
) error {
	return syscall.Mount(source, target, fstype, flags, data)
}

// Unmount default unmount implementation is syscall.
func (m *DefaultMounter) Unmount(target string, flags int, timeout int) error {
	return syscall.Unmount(target, flags)
}

// Refcnt for PathInfo
func (p *PathInfo) Refcnt() int {
	return p.ref
}

// String representation of Mounter
func (m *Mounter) String() string {
	return fmt.Sprintf("%#v", *m)
}

// Inspect mount table for device
func (m *Mounter) Inspect(sourcePath string) []*PathInfo {
	m.Lock()
	defer m.Unlock()

	v, ok := m.mounts[sourcePath]
	if !ok {
		return nil
	}
	return v.Mountpoint
}

// Mounts returns  mount table for device
func (m *Mounter) Mounts(sourcePath string) []string {
	m.Lock()
	defer m.Unlock()

	v, ok := m.mounts[sourcePath]
	if !ok {
		return nil
	}

	mounts := make([]string, len(v.Mountpoint))
	for i, v := range v.Mountpoint {
		mounts[i] = v.Path
	}

	return mounts
}

// GetSourcePaths returns all source paths from the mount table
func (m *Mounter) GetSourcePaths() []string {
	m.Lock()
	defer m.Unlock()

	sourcePaths := make([]string, len(m.mounts))
	i := 0
	for path := range m.mounts {
		sourcePaths[i] = path
		i++
	}
	return sourcePaths
}

// HasMounts determines returns the number of mounts for the device.
func (m *Mounter) HasMounts(sourcePath string) int {
	m.Lock()
	defer m.Unlock()

	v, ok := m.mounts[sourcePath]
	if !ok {
		return 0
	}
	return len(v.Mountpoint)
}

// HasTarget returns true/false based on the target provided
func (m *Mounter) HasTarget(targetPath string) (string, bool) {
	m.Lock()
	defer m.Unlock()

	for k, v := range m.mounts {
		for _, p := range v.Mountpoint {
			if p.Path == targetPath {
				return k, true
			}
		}
	}
	return "", false
}

// Exists scans mountpaths for specified device and returns true if path is one of the
// mountpaths. ErrEnoent may be retuned if the device is not found
func (m *Mounter) Exists(sourcePath string, path string) (bool, error) {
	m.Lock()
	defer m.Unlock()

	v, ok := m.mounts[sourcePath]
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

// GetSourcePath scans mount for a specified mountPath and returns the sourcePath
// if found or returnes an ErrEnoent
func (m *Mounter) GetSourcePath(mountPath string) (string, error) {
	m.Lock()
	defer m.Unlock()

	for k, v := range m.mounts {
		for _, p := range v.Mountpoint {
			if p.Path == mountPath {
				return k, nil
			}
		}
	}
	return "", ErrEnoent
}

func normalizeMountPath(mountPath string) string {
	if len(mountPath) > 1 && strings.HasSuffix(mountPath, "/") {
		return mountPath[:len(mountPath)-1]
	}
	return mountPath
}

func (m *Mounter) maybeRemoveDevice(device string) {
	m.Lock()
	defer m.Unlock()
	if info, ok := m.mounts[device]; ok {
		// If the device has no more mountpoints, remove it from the map
		if len(info.Mountpoint) == 0 {
			delete(m.mounts, device)
		}
	}
}

func (m *Mounter) hasPath(path string) (string, bool) {
	m.Lock()
	defer m.Unlock()
	p, ok := m.paths[path]
	return p, ok
}

func (m *Mounter) addPath(path, device string) {
	m.Lock()
	defer m.Unlock()
	m.paths[path] = device
}

func (m *Mounter) deletePath(path string) bool {
	m.Lock()
	defer m.Unlock()
	if _, pathExists := m.paths[path]; pathExists {
		delete(m.paths, path)
		return true
	}
	return false
}

// Mount new mountpoint for specified device.
func (m *Mounter) Mount(
	minor int,
	device, path, fs string,
	flags uintptr,
	data string,
	timeout int,
) error {

	path = normalizeMountPath(path)
	dev, ok := m.hasPath(path)
	if ok && dev != device {
		dlog.Warnf("cannot mount %q,  device %q is mounted at %q", device, dev, path)
		return ErrExist
	}
	m.Lock()
	info, ok := m.mounts[device]
	if !ok {
		info = &Info{
			Device:     device,
			Mountpoint: make([]*PathInfo, 0),
			Minor:      minor,
			Fs:         fs,
		}
	}
	m.mounts[device] = info
	m.Unlock()
	info.Lock()
	defer info.Unlock()

	// Validate input params
	if fs != info.Fs {
		dlog.Warnf("%s Existing mountpoint has fs %q cannot change to %q",
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
	err := m.mountImpl.Mount(device, path, fs, flags, data, timeout)
	if err != nil {
		return err
	}
	info.Mountpoint = append(info.Mountpoint, &PathInfo{Path: path, ref: 1})
	m.addPath(path, device)
	return nil
}

// Unmount device at mountpoint or decrement refcnt. If device has no
// mountpoints left after this operation, it is removed from the matrix.
// ErrEnoent is returned if the device or mountpoint for the device is not found.
func (m *Mounter) Unmount(device, path string, timeout int) error {
	m.Lock()

	path = normalizeMountPath(path)
	info, ok := m.mounts[device]
	if !ok {
		m.Unlock()
		return ErrEnoent
	}
	m.Unlock()
	info.Lock()
	defer info.Unlock()
	for i, p := range info.Mountpoint {
		if p.Path != path {
			continue
		}
		p.ref--
		// Unmount only if refcnt is 0
		if p.ref != 0 {
			return nil
		}
		err := m.mountImpl.Unmount(path, syscall.MNT_DETACH, timeout)
		if err != nil {
			p.ref++
			return err
		}
		if pathExists := m.deletePath(path); !pathExists {
			dlog.Warnf("Path %q for device %q does not exist in pathMap",
				path, device)
		}
		// Blow away this mountpoint.
		info.Mountpoint[i] = info.Mountpoint[len(info.Mountpoint)-1]
		info.Mountpoint = info.Mountpoint[0 : len(info.Mountpoint)-1]
		m.maybeRemoveDevice(device)
		return nil
	}
	return ErrEnoent
}

// New returns a new Mount Manager
func New(mounterType MountType,
	mountImpl MountImpl,
	identifier string,
) (Manager, error) {

	if mountImpl == nil {
		mountImpl = &DefaultMounter{}
	}

	switch mounterType {
	case DeviceMount:
		return NewDeviceMounter(identifier, mountImpl)
	case NFSMount:
		return NewNFSMounter(identifier, mountImpl)
	}
	return nil, ErrUnsupported
}
