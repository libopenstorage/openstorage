// Copyright (C) 2014 Andreas Klauer <Andreas.Klauer@metamorpher.de>
// License: MIT

// Package buse uses the Linux NBD layer to emulate a block device in user space
package buse

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"

	"go.pedge.io/dlog"
)

const (
	// Defined in <linux/fs.h>:
	BLKROSET = 4701
	// Defined in <linux/nbd.h>:
	NBD_SET_SOCK        = 43776
	NBD_SET_BLKSIZE     = 43777
	NBD_SET_SIZE        = 43778
	NBD_CONNECT         = 43779
	NBD_CLEAR_SOCK      = 43780
	NBD_CLEAR_QUE       = 43781
	NBD_PRINT_DEBUG     = 43782
	NBD_SET_SIZE_BLOCKS = 43783
	NBD_DISCONNECT      = 43784
	NBD_SET_TIMEOUT     = 43785
	NBD_SET_FLAGS       = 43786
	// enum
	NBD_CMD_READ  = 0
	NBD_CMD_WRITE = 1
	NBD_CMD_DISC  = 2
	NBD_CMD_FLUSH = 3
	NBD_CMD_TRIM  = 4
	// values for flags field
	NBD_FLAG_HAS_FLAGS  = (1 << 0) // nbd-server supports flags
	NBD_FLAG_READ_ONLY  = (1 << 1) // device is read-only
	NBD_FLAG_SEND_FLUSH = (1 << 2) // can flush writeback cache
	NBD_FLAG_SEND_FUA   = (1 << 3) // Send FUA (Force Unit Access)
	NBD_FLAG_ROTATIONAL = (1 << 4) // Use elevator algorithm - rotational media
	NBD_FLAG_SEND_TRIM  = (1 << 5) // Send TRIM (discard)

	// These are sent over the network in the request/reply magic fields
	NBD_REQUEST_MAGIC = 0x25609513
	NBD_REPLY_MAGIC   = 0x67446698
	// Do *not* use magics: 0x12560953 0x96744668.
)

// ioctl() helper function
func ioctl(a1, a2, a3 uintptr) (err error) {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, a1, a2, a3)
	if errno != 0 {
		err = errno
	}
	return err
}

// Device interface is a subset of os.File.
type Device interface {
	ReadAt(b []byte, off int64) (n int, err error)
	WriteAt(b []byte, off int64) (n int, err error)
}

type request struct {
	magic  uint32
	typus  uint32
	handle uint64
	from   uint64
	len    uint32
}

type reply struct {
	magic  uint32
	error  uint32
	handle uint64
}

type NBD struct {
	device     Device
	devicePath string
	deviceFile *os.File
	size       int64
	socket     int
	mutex      *sync.Mutex
}

func Create(device Device, size int64) *NBD {
	if size >= 0 {
		return &NBD{device: device,
			devicePath: "",
			size:       size,
			deviceFile: nil,
			socket:     0,
			mutex:      &sync.Mutex{}}
	}
	return nil
}

// Return true if connected.
func (nbd *NBD) IsConnected() bool {
	return nbd.deviceFile != nil && nbd.socket > 0
}

// Get the size of the NBD.
func (nbd *NBD) GetSize() int64 {
	return nbd.size
}

// Set the size of the NBD.
func (nbd *NBD) Size(size int64) (err error) {
	if err = ioctl(nbd.deviceFile.Fd(), NBD_SET_BLKSIZE, 4096); err != nil {
		err = &os.PathError{nbd.deviceFile.Name(), "ioctl NBD_SET_BLKSIZE", err}
	} else if err = ioctl(nbd.deviceFile.Fd(), NBD_SET_SIZE_BLOCKS, uintptr(size/4096)); err != nil {
		err = &os.PathError{nbd.deviceFile.Name(), "ioctl NBD_SET_SIZE_BLOCKS", err}
	}

	return err
}

// Connect the network block device.
func (nbd *NBD) Connect() (dev string, err error) {
	pair, err := syscall.Socketpair(syscall.SOCK_STREAM, syscall.AF_UNIX, 0)
	if err != nil {
		return "", err
	}

	// Find free NBD device.
	for i := 0; ; i++ {
		dev = fmt.Sprintf("/dev/nbd%d", i)
		if _, err = os.Stat(dev); os.IsNotExist(err) {
			dev = ""
			return "", errors.New("No more NBD devices left.")
		}
		if _, err = os.Stat(fmt.Sprintf("/sys/block/nbd%d/pid", i)); !os.IsNotExist(err) {
			continue // Busy.
		}

		dlog.Infof("Attempting to open device %v", dev)
		if nbd.deviceFile, err = os.Open(dev); err == nil {
			// Possible candidate.
			ioctl(nbd.deviceFile.Fd(), BLKROSET, 0)
			if err := ioctl(nbd.deviceFile.Fd(), NBD_SET_SOCK, uintptr(pair[0])); err == nil {
				nbd.socket = pair[1]
				break // Success.
			}
		}
	}

	// Setup.
	if err = nbd.Size(nbd.size); err != nil {
		// Already set by nbd.Size().
	} else if err = ioctl(nbd.deviceFile.Fd(), NBD_SET_FLAGS, 1); err != nil {
		err = &os.PathError{nbd.deviceFile.Name(), "ioctl NBD_SET_FLAGS", err}
	} else {
		go nbd.connect()
		go nbd.handle()
	}

	nbd.devicePath = dev

	return dev, err
}

func (nbd *NBD) Disconnect() {
	nbd.mutex.Lock()
	if nbd.IsConnected() {
		ioctl(nbd.deviceFile.Fd(), NBD_DISCONNECT, 0)
		ioctl(nbd.deviceFile.Fd(), NBD_CLEAR_QUE, 0)
		ioctl(nbd.deviceFile.Fd(), NBD_CLEAR_SOCK, 0)
		nbd.deviceFile.Close()
		nbd.deviceFile = nil

		dummy := make([]byte, 1)
		syscall.Write(nbd.socket, dummy)
		syscall.Close(nbd.socket)
		nbd.socket = 0
	}
	nbd.mutex.Unlock()
}

func (nbd *NBD) connect() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// NBD_CONNECT does not return until disconnect.
	ioctl(nbd.deviceFile.Fd(), NBD_CONNECT, 0)

	dlog.Infof("Closing device file %s", nbd.devicePath)
}

// Handle block requests.
func (nbd *NBD) handle() {
	buf := make([]byte, 2<<19)
	var x request

	for {
		bytes, err := syscall.Read(nbd.socket, buf[0:28])
		if nbd.deviceFile == nil {
			dlog.Infof("Disconnecting device %s", nbd.devicePath)
			return
		}

		if bytes < 0 || err != nil {
			dlog.Errorf("Error reading from device %s", nbd.devicePath)
			nbd.Disconnect()
			return
		}

		x.magic = binary.BigEndian.Uint32(buf)
		x.typus = binary.BigEndian.Uint32(buf[4:8])
		x.handle = binary.BigEndian.Uint64(buf[8:16])
		x.from = binary.BigEndian.Uint64(buf[16:24])
		x.len = binary.BigEndian.Uint32(buf[24:28])

		switch x.magic {
		case NBD_REPLY_MAGIC:
			fallthrough
		case NBD_REQUEST_MAGIC:
			switch x.typus {
			case NBD_CMD_READ:
				nbd.device.ReadAt(buf[16:16+x.len], int64(x.from))
				binary.BigEndian.PutUint32(buf[0:4], NBD_REPLY_MAGIC)
				binary.BigEndian.PutUint32(buf[4:8], 0)
				syscall.Write(nbd.socket, buf[0:16+x.len])
			case NBD_CMD_WRITE:
				n, _ := syscall.Read(nbd.socket, buf[28:28+x.len])
				for uint32(n) < x.len {
					m, _ := syscall.Read(nbd.socket, buf[28+n:28+x.len])
					n += m
				}
				nbd.device.WriteAt(buf[28:28+x.len], int64(x.from))
				binary.BigEndian.PutUint32(buf[0:4], NBD_REPLY_MAGIC)
				binary.BigEndian.PutUint32(buf[4:8], 0)
				syscall.Write(nbd.socket, buf[0:16])
			case NBD_CMD_DISC:
				dlog.Infof("Disconnecting device %s", nbd.devicePath)
				nbd.Disconnect()
				return
			case NBD_CMD_FLUSH:
				fallthrough
			case NBD_CMD_TRIM:
				binary.BigEndian.PutUint32(buf[0:4], NBD_REPLY_MAGIC)
				binary.BigEndian.PutUint32(buf[4:8], 1)
				syscall.Write(nbd.socket, buf[0:16])
			default:
				dlog.Errorf("Unknown command recieved on device %s", nbd.devicePath)
				nbd.Disconnect()
				return
			}
		default:
			dlog.Errorf("Invalid packet command recieved on device %s", nbd.devicePath)
			nbd.Disconnect()
			return
		}
	}
}
