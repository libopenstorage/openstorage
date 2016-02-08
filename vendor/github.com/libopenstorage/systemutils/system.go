package systemutils

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/gosigar"
)

type system struct {
	cpuUsage   float64
	totalTicks float64
	ticks      float64
}

func (s *system) start() {
	go func() {
		for {
			idle0, total0 := getCPUSample()
			time.Sleep(2 * time.Second)
			idle1, total1 := getCPUSample()

			idleTicks := float64(idle1 - idle0)
			s.totalTicks = float64(total1 - total0)
			s.ticks = s.totalTicks - idleTicks
			s.cpuUsage = 100 * s.ticks / s.totalTicks
		}
	}()
}

func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}

func (s *system) CpuUsage() (usage float64, total float64, ticks float64) {
	return s.cpuUsage, s.ticks, s.totalTicks
}

func (s *system) MemUsage() (total, used, free uint64) {
	mem := sigar.Mem{}

	mem.Get()

	return mem.Total, mem.ActualUsed, mem.ActualFree
}

func (s *system) Luns() map[string]Lun {
	luns := make(map[string]Lun)

	/*
		var dev string
		//XXX Temporarily disable scanning devices.
		return luns
		out, _ := exec.Command("/sbin/fdisk", "-l").Output()

		lines := strings.Split(string(out), "\n")

		for _, line := range lines {
			if strings.HasPrefix(line, "Disk /") {
				luns[dev] = Lun{Capacity: 800}
			}
		}
	*/
	return luns
}
