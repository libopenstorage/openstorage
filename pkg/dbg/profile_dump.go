package dbg

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// Path defailt location for debug/profile output
	Path     = "/var/cores/"
	fnameFmt = "2006-01-02T15:04:05.999999-0700MST"
)

// GetTimeStamp returns 'readable' timestamp with no spaces 'YYYYMMDDHHMMSS'
func GetTimeStamp() string {
	tnow := time.Now()
	return fmt.Sprintf("%d%02d%02d%02d%02d%02d",
		tnow.Year(), tnow.Month(), tnow.Day(),
		tnow.Hour(), tnow.Minute(), tnow.Second())
}

// GetHostNamePrefix return '<hostname>' to use as prefix.
func GetHostNamePrefix() string {
	hname, _ := os.Hostname()
	return hname
}

// GoProfileDump output goroutines to file.
func DumpGoProfileTo(fname string) error {
	trace := make([]byte, 5120*1024)
	len := runtime.Stack(trace, true)
	return ioutil.WriteFile(fname, trace[:len], 0644)
}

// HeapDump output heap to filename.
func HeapDumpTo(fname string) {
	f, err := os.Create(fname)
	if err != nil {
		logrus.Errorf("could not create memory profile: %v", err)
		return
	}
	defer f.Close()
	if err := pprof.WriteHeapProfile(f); err != nil {
		logrus.Errorf("could not write memory profile: %v", err)
	}
}

// DumpGoMemoryTrace output memory profile to logs.
func DumpGoMemoryTrace() {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	res := fmt.Sprintf("%#v", m)
	logrus.Infof("==== Dumping Memory Profile ===")
	logrus.Infof(res)
}

// DumpGoProfile output goroutines to file.
func DumpGoProfile() error {
	return DumpGoProfileTo(Path + GetHostNamePrefix() + "-" + GetTimeStamp() + ".stack")
}

// DumpHeap output heap to file.
func DumpHeap() {
	HeapDumpTo(Path + GetHostNamePrefix() + "-" + GetTimeStamp() + ".heap")
}
