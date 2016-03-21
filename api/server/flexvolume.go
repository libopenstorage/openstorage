package server

import (
	"fmt"

	"github.com/libopenstorage/openstorage/pkg/mount"
	"github.com/libopenstorage/openstorage/volume"

	"go.pedge.io/dlog"
)

const (
	osdDriverKey = "osdDriver"
	volumeIdKey  = "volumeID"
	pxDriverKey  = "pxd"
)

var (
	deviceDriverMap map[string]string
)

type flexVolumeClient struct {
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
func (c *flexVolumeClient) Init() error {
	return nil
}

func (c *flexVolumeClient) Attach(jsonOptions map[string]string) error {
	driverName, ok := jsonOptions[osdDriverKey]
	if !ok {
		return fmt.Errorf("Invalid kubernetes spec file : No OSD driver specified in flexvolume")
	}

	driver, err := c.getVolumeDriver(driverName)
	if err != nil {
		return err
	}

	mountDevice, ok := jsonOptions[volumeIdKey]
	if !ok {
		return fmt.Errorf("Invalid kubernetes spec file : No volumeID specified in flexvolume")
	}

	_, err = driver.Attach(mountDevice)
	if err != nil {
		return err
	}

	return nil
}

func (c *flexVolumeClient) Detach(mountDevice string) error {
	var driverName string
	driverName, ok := deviceDriverMap[mountDevice]
	if !ok {
		dlog.Infof("Could not find driver for (%v). Defaulting to pxd", mountDevice)
		driverName = pxDriverKey
	}

	driver, err := c.getVolumeDriver(driverName)
	if err != nil {
		return err
	}

	err = driver.Detach(mountDevice)
	if err != nil {
		return err
	}
	return nil
}

func (c *flexVolumeClient) Mount(targetMountDir string, mountDevice string,
	jsonOptions map[string]string) error {
	driverName, ok := jsonOptions[osdDriverKey]
	if !ok {
		return fmt.Errorf("Invalid kubernetes spec file : No OSD driver specified in flexvolume")
	}

	driver, err := c.getVolumeDriver(driverName)
	if err != nil {
		return err
	}

	if targetMountDir == "" {
		return fmt.Errorf("Missing target mount path")
	}

	err = driver.Mount(mountDevice, targetMountDir)
	if err != nil {
		return fmt.Errorf("Failed in mount. Driver returned error : (%v)", err)
	}

	// Update the deviceDriverMap
	mountManager, _ := mount.New(mount.DeviceMount, "")
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
	mountManager, _ := mount.New(mount.DeviceMount, "")
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

	err = driver.Unmount(mountDevice, mountDir)
	if err != nil {
		return err
	}

	return nil
}
