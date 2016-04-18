package coprhd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	diskById = "/dev/disk/by-id"
)

// GetScaleIoDevice returns the local device associated with the volume
func (d *driver) GetScaleIoDevice(volumeID string) (string, error) {
	if _, err := d.GetVol(volumeID); err != nil {
		return "", fmt.Errorf("Volume could not be located")
	}

	wwn := volumeID

	devices, err := ioutil.ReadDir(diskById)
	if err != nil {
		return "", fmt.Errorf("Failed to read devices directory")
	}

	for _, deviceInfo := range devices {
		name := strings.Replace(deviceInfo.Name(), "-", "", -1)
		name = strings.TrimPrefix(name, "emcvol")
		if name == wwn {
			path := filepath.Join(diskById, deviceInfo.Name())

			if device, err := filepath.EvalSymlinks(path); err != nil {
				return "", err
			} else if device != "" {
				return device, nil
			}
		}
	}

	return "", fmt.Errorf("Device path %s not found", wwn)
}
