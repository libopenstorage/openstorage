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
	"strconv"

	"github.com/moby/sys/mountinfo"
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
func parseMountTable() ([]*mountinfo.Info, error) {
	mountInfoBytes, err := consistentRead("/proc/self/mountinfo", 3)
	if err != nil {
		return nil, err
	}

	return ParseInfoFile(bytes.NewReader(mountInfoBytes))
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

func ParseInfoFile(r io.Reader) ([]*mountinfo.Info, error) {
	var (
		s   = bufio.NewScanner(r)
		out = []*mountinfo.Info{}
	)

	for s.Scan() {
		var err error
		if err = s.Err(); err != nil {
			return nil, err
		}
		/*
		   See http://man7.org/linux/man-pages/man5/proc.5.html

		   36 35 98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
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
		   (11) super options:  per super block options

		   In other words, we have:
		    * 6 mandatory fields	(1)..(6)
		    * 0 or more optional fields	(7)
		    * a separator field		(8)
		    * 3 mandatory fields	(9)..(11)
		*/

		text := s.Text()
		fields := strings.Split(text, " ")
		numFields := len(fields)
		if numFields < 10 {
			// should be at least 10 fields
			return nil, fmt.Errorf("parsing '%s' failed: not enough fields (%d)", text, numFields)
		}

		// separator field
		sepIdx := numFields - 4
		// In Linux <= 3.9 mounting a cifs with spaces in a share
		// name (like "//srv/My Docs") _may_ end up having a space
		// in the last field of mountinfo (like "unc=//serv/My Docs").
		// Since kernel 3.10-rc1, cifs option "unc=" is ignored,
		// so spaces should not appear.
		//
		// Check for a separator, and work around the spaces bug
		for fields[sepIdx] != "-" {
			sepIdx--
			if sepIdx == 5 {
				return nil, fmt.Errorf("parsing '%s' failed: missing - separator", text)
			}
		}

		p := &mountinfo.Info{}

		p.Mountpoint, err = unescape(fields[4])
		if err != nil {
			return nil, fmt.Errorf("parsing '%s' failed: mount point: %w", fields[4], err)
		}
		p.FSType, err = unescape(fields[sepIdx+1])
		if err != nil {
			return nil, fmt.Errorf("parsing '%s' failed: fstype: %w", fields[sepIdx+1], err)
		}
		p.Source, err = unescape(fields[sepIdx+2])
		if err != nil {
			return nil, fmt.Errorf("parsing '%s' failed: source: %w", fields[sepIdx+2], err)
		}
		p.VFSOptions = fields[sepIdx+3]

		// ignore any numbers parsing errors, as there should not be any
		p.ID, _ = strconv.Atoi(fields[0])
		p.Parent, _ = strconv.Atoi(fields[1])
		mm := strings.SplitN(fields[2], ":", 3)
		if len(mm) != 2 {
			return nil, fmt.Errorf("parsing '%s' failed: unexpected major:minor pair %s", text, mm)
		}
		p.Major, _ = strconv.Atoi(mm[0])
		p.Minor, _ = strconv.Atoi(mm[1])

		p.Root, err = unescape(fields[3])
		if err != nil {
			return nil, fmt.Errorf("parsing '%s' failed: root: %w", fields[3], err)
		}

		p.Options = fields[5]

		// zero or more optional fields
		p.Optional = strings.Join(fields[6:sepIdx], " ")

		out = append(out, p)
	}
	return out, nil
}


// A few specific characters in mountinfo path entries (root and mountpoint)
// are escaped using a backslash followed by a character's ascii code in octal.
//
//	space              -- as \040
//	tab (aka \t)       -- as \011
//	newline (aka \n)   -- as \012
//	backslash (aka \\) -- as \134
//
// This function converts path from mountinfo back, i.e. it unescapes the above sequences.
func unescape(path string) (string, error) {
	// try to avoid copying
	if strings.IndexByte(path, '\\') == -1 {
		return path, nil
	}

	// The following code is UTF-8 transparent as it only looks for some
	// specific characters (backslash and 0..7) with values < utf8.RuneSelf,
	// and everything else is passed through as is.
	buf := make([]byte, len(path))
	bufLen := 0
	for i := 0; i < len(path); i++ {
		if path[i] != '\\' {
			buf[bufLen] = path[i]
			bufLen++
			continue
		}
		s := path[i:]
		if len(s) < 4 {
			// too short
			return "", fmt.Errorf("bad escape sequence %q: too short", s)
		}
		c := s[1]
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7':
			v := c - '0'
			for j := 2; j < 4; j++ { // one digit already; two more
				if s[j] < '0' || s[j] > '7' {
					return "", fmt.Errorf("bad escape sequence %q: not a digit", s[:3])
				}
				x := s[j] - '0'
				v = (v << 3) | x
			}
			if v > 255 {
				return "", fmt.Errorf("bad escape sequence %q: out of range" + s[:3])
			}
			buf[bufLen] = v
			bufLen++
			i += 3
			continue
		default:
			return "", fmt.Errorf("bad escape sequence %q: not a digit" + s[:3])

		}
	}

	return string(buf[:bufLen]), nil
}
