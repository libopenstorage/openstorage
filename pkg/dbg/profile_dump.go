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
	path     = "/var/cores/"
	fnameFmt = "2006-01-02T15:04:05.999999-0700MST"
)

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
	trace := make([]byte, 5120*1024)
	len := runtime.Stack(trace, true)
	return ioutil.WriteFile(path+time.Now().Format(fnameFmt)+".stack", trace[:len], 0644)
}

func DumpHeap() {
	f, err := os.Create(path + time.Now().Format(fnameFmt) + ".heap")
	if err != nil {
		logrus.Errorf("could not create memory profile: %v", err)
		return
	}
	defer f.Close()
	if err := pprof.WriteHeapProfile(f); err != nil {
		logrus.Errorf("could not write memory profile: %v", err)
	}
}
