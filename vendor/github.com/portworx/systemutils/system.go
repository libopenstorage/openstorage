package systemutils

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type system struct {
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

func (s system) CpuUsage() (usage float64, total float64, ticks float64) {
	idle0, total0 := getCPUSample()
	time.Sleep(3 * time.Second)
	idle1, total1 := getCPUSample()

	idleTicks := float64(idle1 - idle0)
	totalTicks := float64(total1 - total0)
	cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks

	return cpuUsage, totalTicks - idleTicks, totalTicks
}

func (s system) MemUsage() (available float64) {
	out, _ := exec.Command("/bin/free", "-m").Output()
	r := regexp.MustCompile("(^|\\s)([0-9]+)($|\\s)")
	str := r.FindString(string(out))
	f, _ := strconv.ParseFloat(str, 64)
	return f
}

func (s system) Luns() map[string]Lun {
    var dev string
    // var sz float64
    // var bytes, sectors uint64
    luns := make(map[string]Lun)

    out, _ := exec.Command("/sbin/fdisk", "-l").Output()

    lines := strings.Split(string(out), "\n")

    for _, line := range lines {
        if strings.HasPrefix(line, "Disk /") {
            // _, err := fmt.Sscanf(line, "Disk %s: %f GB, %d bytes, %d sectors\n", &dev, &sz, &bytes, &sectors)
            luns[dev] = Lun{Capacity: 800}
            // fmt.Printf("%s\nDevice: %s, Size: %f, %d %d\n", line, dev, sz, bytes, sectors)
        }
    }

    return luns
}

