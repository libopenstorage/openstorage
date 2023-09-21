/*
Copyright © 2017 Moby and Kubernetes Authors

Mount parse code originally from:
https://github.com/moby/sys/blob/65f80e71a828ef17e6f573176dc569e55f519937/mountinfo/mountinfo_linux.go

Consistent read code originally from:
https://github.com/kubernetes/utils/blob/a0dff01d8ea5b4e8799d52a7fad727caafc9c685/io/read.go#L33-L35

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

package mount

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/docker/docker/pkg/mount"
)

const (
	/* 36 35 98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
	   (1)(2)(3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)

	   (1) mount ID:  unique identifier of the mount (may be reused after umount)
	   (2) parent ID:  ID of parent (or of self for the top of the mount tree)
	   (3) major:minor:  value of st_dev for files on filesystem
	   (4) root:  root of the mount within the filesystem
	   (5) mount point:  mount point relative to the process's root
	   (6) mount options:  per mount options
	   (7) optional fields:  zero or more fields of the form "tag[:value]"
	   (8) separator:  marks the end of the optional fields
	   (9) filesystem type:  name of filesystem of the form "type[.subtype]"
	   (10) mount source:  filesystem specific information or "none"
	   (11) super options:  per super block options*/
	mountinfoFormat = "%d %d %d:%d %s %s %s %s"
)

// ErrLimitReached means that the read limit is reached.
var ErrLimitReached = errors.New("the read limit is reached")

// Parse /proc/self/mountinfo because comparing Dev and ino does not work from
// bind mounts. function is originally from
// https://github.com/moby/sys/blob/65f80e71a828ef17e6f573176dc569e55f519937/mountinfo/mountinfo_linux.go
func parseMountTable() ([]*mount.Info, error) {
	mountInfoBytes, err := consistentRead("/proc/self/mountinfo", 3)
	if err != nil {
		return nil, err
	}

	return parseInfoFile(bytes.NewReader(mountInfoBytes))
}

// consistentRead repeatedly reads a file until it gets the same content twice.
// This is useful when reading files in /proc that are larger than page size
// and kernel may modify them between individual read() syscalls.
// originally from https://github.com/kubernetes/utils/blob/a0dff01d8ea5b4e8799d52a7fad727caafc9c685/io/read.go#L33-L35
func consistentRead(filename string, attempts int) ([]byte, error) {
	oldContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	for i := 0; i < attempts; i++ {
		newContent, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		if bytes.Compare(oldContent, newContent) == 0 {
			return newContent, nil
		}
		// Files are different, continue reading
		oldContent = newContent
	}
	return nil, fmt.Errorf("could not get consistent content of mount table after %d attempts", attempts)
}

func parseInfoFile(r io.Reader) ([]*mount.Info, error) {
	var (
		s   = bufio.NewScanner(r)
		out = []*mount.Info{}
	)

	for s.Scan() {
		if err := s.Err(); err != nil {
			return nil, err
		}

		var (
			p              = &mount.Info{}
			text           = s.Text()
			optionalFields string
		)

		if _, err := fmt.Sscanf(text, mountinfoFormat,
			&p.ID, &p.Parent, &p.Major, &p.Minor,
			&p.Root, &p.Mountpoint, &p.Opts, &optionalFields); err != nil {
			return nil, fmt.Errorf("Scanning '%s' failed: %s", text, err)
		}
		// Safe as mountinfo encodes mountpoints with spaces as \040.
		index := strings.Index(text, " - ")
		postSeparatorFields := strings.Fields(text[index+3:])
		if len(postSeparatorFields) < 3 {
			return nil, fmt.Errorf("Error found less than 3 fields post '-' in %q", text)
		}

		if optionalFields != "-" {
			p.Optional = optionalFields
		}

		p.Fstype = postSeparatorFields[0]
		p.Source = postSeparatorFields[1]
		p.VfsOpts = strings.Join(postSeparatorFields[2:], " ")
		out = append(out, p)
	}
	return out, nil
}
