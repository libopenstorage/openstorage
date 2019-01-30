/*
Package util provides utility functions for OSD servers and drivers.
Copyright 2017 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package util

import (
	"fmt"
	"path"
	"syscall"
)

// IsSameFilesystem takes a group of files/directories, and returns TRUE if all files/dirs belong to the same file-system.
func IsSameFilesystem(filesOrDirs ...string) (bool, error) {
	if len(filesOrDirs) < 2 {
		return false, fmt.Errorf("Need at least 2 arguments")
	}

	var st0, st1 syscall.Stat_t

	err := syscall.Lstat(filesOrDirs[0], &st0)
	if err != nil {
		return false, fmt.Errorf("Could not stat %s: %s", filesOrDirs[1], err)
	}

	for _, m := range filesOrDirs[1:] {
		err = syscall.Lstat(m, &st1)
		if err != nil {
			return false, fmt.Errorf("Could not stat %s: %s", m, err)
		} else if st0.Dev != st1.Dev {
			return false, nil
		}
	}

	return true, nil
}

// IsMountpoint tests if the directory is a mountpoint to a different file-system
func IsMountpoint(dir string) (bool, error) {
	b, err := IsSameFilesystem(dir, path.Dir(dir))
	return !b, err
}
