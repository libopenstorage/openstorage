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

// getScaleIoDevice returns the local device associated with the volume
func (d *driver) getScaleIoDevice(volumeID string) (string, error) {
	_, err := d.GetVol(volumeID)
	if err != nil {
		return "", fmt.Errorf("Volume could not be located")
	}

	wwn := volumeID

	devs, _ := ioutil.ReadDir(diskById)

	for _, p := range devs {
		name := strings.Replace(p.Name(), "-", "", -1)
		if strings.Contains(name, wwn) {
			path := filepath.Join(diskById, p.Name())
			dev, err := filepath.EvalSymlinks(path)
			if err != nil {
				return "", err
			}
			if dev != "" {
				return dev, nil
			}
		}
	}

	return "", fmt.Errorf("Device path %s not found", wwn)
}
