package server

import (
	"errors"
	"fmt"

	"github.com/libopenstorage/openstorage/pkg/mount"
	"github.com/libopenstorage/openstorage/volume"

	"go.pedge.io/dlog"
)

const (
	osdDriverKey = "osdDriver"
	volumeIDKey  = "volumeID"
	pxDriverKey  = "pxd"
)

var (
	deviceDriverMap map[string]string
	// ErrInvalidSpecDriver is returned when incorrect or no OSD driver is specified in spec file
	ErrInvalidSpecDriver = errors.New("Invalid kubernetes spec file : No OSD driver specified in flexvolume")
	// ErrInvalidSpecVolumeID is returned when incorrect or no volumeID is specified in spec file
	ErrInvalidSpecVolumeID = errors.New("Invalid kubernetes spec file : No volumeID specified in flexvolume")
	// ErrNoMountInfo is returned when flexvolume is unable to read from proc/self/mountinfo
	ErrNoMountInfo = errors.New("Could not read mountpoints from /proc/self/mountinfo. ")
	// ErrMissingMountPath is returned when no mountPath specified in mount call
	ErrMissingMountPath = errors.New("Missing target mount path")
)

type flexVolumeClient struct{}

func (c *flexVolumeClient) Init() error {
	return nil
}

func (c *flexVolumeClient) Attach(jsonOptions map[string]string) error {
	driverName, ok := jsonOptions[osdDriverKey]
	if !ok {
		return ErrInvalidSpecDriver
	}

	driver, err := c.getVolumeDriver(driverName)
	if err != nil {
		return err
	}

	mountDevice, ok := jsonOptions[volumeIDKey]
	if !ok {
		return ErrInvalidSpecVolumeID
	}

	if _, err = driver.Attach(mountDevice); err != nil {
		return err
	}
	return nil
}

func (c *flexVolumeClient) Detach(mountDevice string) error {
	driverName, ok := deviceDriverMap[mountDevice]
	if !ok {
		dlog.Infof("Could not find driver for (%v). Defaulting to pxd", mountDevice)
		driverName = pxDriverKey
	}

	driver, err := c.getVolumeDriver(driverName)
	if err != nil {
		return err
	}

	if err = driver.Detach(mountDevice); err != nil {
		return err
	}
	return nil
}

func (c *flexVolumeClient) Mount(targetMountDir string, mountDevice string,
	jsonOptions map[string]string) error {

	driverName, ok := jsonOptions[osdDriverKey]
	if !ok {
		return ErrInvalidSpecDriver
	}

	driver, err := c.getVolumeDriver(driverName)
	if err != nil {
		return err
	}

	if targetMountDir == "" {
		return ErrMissingMountPath
	}

	if err = driver.Mount(mountDevice, targetMountDir); err != nil {
		return err
	}

	// Update the deviceDriverMap
	mountManager, err := mount.New(mount.DeviceMount, "")
	if err != nil {
		dlog.Infof("Could not read mountpoints from /proc/self/mountinfo. Device - Driver mapping not saved!")
		return nil
	}

	devPath, err := mountManager.GetDevPath(targetMountDir)
	if err != nil {
		return fmt.Errorf("Mount Failed. Could not find mountpoint in /proc/self/mountinfo")
	}
	deviceDriverMap[devPath] = driverName
	return nil
}

func (c *flexVolumeClient) Unmount(mountDir string) error {
	var driverName string

	// Get the mountDevice from mount manager
	mountManager, err := mount.New(mount.DeviceMount, "")
	if err != nil {
		return ErrNoMountInfo
	}

	mountDevice, err := mountManager.GetDevPath(mountDir)
	if err != nil {
		return fmt.Errorf("Invalid mountDir (%v). No device mounted on this path", mountDir)
	}

	driverName, ok := deviceDriverMap[mountDevice]
	if !ok {
		dlog.Infof("Could not find driver for (%v). Defaulting to pxd", mountDevice)
		driverName = pxDriverKey
	}

	driver, err := c.getVolumeDriver(driverName)
	if err != nil {
		return err
	}

	if err = driver.Unmount(mountDevice, mountDir); err != nil {
		return err
	}
	return nil
}

func newFlexVolumeClient() *flexVolumeClient {
	deviceDriverMap = make(map[string]string)
	return &flexVolumeClient{}
}

func (c *flexVolumeClient) getVolumeDriver(driverName string) (volume.VolumeDriver, error) {
	driver, err := volume.Get(driverName)
	if err != nil {
		return nil, fmt.Errorf("%v driver not initialized in OSD", driverName)
	}
	return driver, nil
}
