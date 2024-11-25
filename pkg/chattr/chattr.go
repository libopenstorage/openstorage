package chattr

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	chattrCmd = "chattr"
	lsattrCmd = "lsattr"
)

func isInappropriateIoctl(err string) bool {
	lowerErr := strings.ToLower(err)
	// "inappropriate ioctl for device" returned inside PX container
	// "invalid argument while setting flags" returned on host
	return strings.Contains(lowerErr, "inappropriate ioctl for device") || strings.Contains(lowerErr, "invalid argument while setting flags")
}

func runChattr(path string, arg string, force bool) error {
	chattrBin := which(chattrCmd)
	if _, err := os.Stat(path); err == nil {
		args := []string{arg}
		if force {
			args = append(args, "-f")
		}
		args = append(args, path)
		cmd := exec.Command(chattrBin, args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err = cmd.Run(); err != nil {
			// chattr exits with status 1 for files that are not permitted without force flag
			// this will happen for files that throws the error
			//  "Operation not supported while reading flags"
			if force && cmd.ProcessState.ExitCode() == 1 {
				return nil
			}

			stderrStr := stderr.String()
			if isInappropriateIoctl(stderrStr) { // Returned in case of filesystem that does not support chattr
				return nil
			}
			return fmt.Errorf("%s %s failed: %s. Err: %v", chattrBin, strings.Join(args, " "), stderrStr, err)
		}
	}

	return nil
}

func AddImmutable(path string) error {
	return ForceAddImmutable(path, false)
}

func ForceAddImmutable(path string, force bool) error {
	return runChattr(path, "+i", force)
}

func RemoveImmutable(path string) error {
	return ForceRemoveImmutable(path, false)
}
func ForceRemoveImmutable(path string, force bool) error {
	return runChattr(path, "-i", force)
}

func IsImmutable(path string) bool {
	lsattrBin := which(lsattrCmd)
	if _, err := os.Stat(path); err != nil {
		logrus.Errorf("Failed to stat mount path:%v", err)
		return true
	}
	op, err := exec.Command(lsattrBin, "-d", path).CombinedOutput()
	if err != nil {
		if isInappropriateIoctl(string(op)) { // Returned in case of filesystem that does not support chattr
			return false
		}
		// Cannot get path status, return true so that immutable bit is not reverted
		logrus.Errorf("Error listing attrs for %v err:%v", path, string(op))
		return true
	}
	// 'lsattr -d' output is a single line with 2 fields separated by space; 1st one
	// is list of applicable attrs and 2nd field is the path itself.Sample output below.
	// lsattr -d /mnt/vol2
	// ----i--------e-- /mnt/vol2
	attrs := strings.Split(string(op), " ")
	if len(attrs) != 2 {
		// Cannot get path status, return true so that immutable bit is not reverted
		logrus.Errorf("Invalid lsattr output %v", string(op))
		return true
	}
	if strings.Contains(attrs[0], "i") {
		logrus.Warnf("Path %v already set to immutable", path)
		return true
	}

	return false
}

func which(bin string) string {
	pathList := []string{"/usr/bin", "/sbin", "/usr/sbin", "/usr/local/bin"}
	for _, p := range pathList {
		if _, err := os.Stat(path.Join(p, bin)); err == nil {
			return path.Join(p, bin)
		}
	}
	return bin
}
