package stats

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	FrequencyMin = time.Second * 10
)

type DiskStats struct {
	sync.Mutex
	stats       map[string]*Stats
	frequency   time.Duration
	lastCollect time.Time
	c           chan int
}

type Stats struct {
	Reads        int64
	ReadsMerged  int64
	ReadSectors  int64
	ReadMs       int64
	Writes       int64
	WritesMerged int64
	WriteSectors int64
	WriteMs      int64
	IOProgress   int64
	IOWeightedMS int64
}

func NewDiskStats(frequency time.Duration) *DiskStats {
	if frequency < FrequencyMin {
		frequency = FrequencyMin
	}
	return &DiskStats{
		stats:     make(map[string]*Stats),
		frequency: frequency,
		c:         make(chan int, 100),
	}
}

func (d *DiskStats) Get(dev string) *Stats {
	if time.Since(d.lastCollect) > 2*d.frequency {
		d.collect()
	}

	d.c <- 1
	return d.stats[dev]
}

func (d *DiskStats) Collect() {
	for {
		d.collect()
		d.lastCollect = time.Now()
		time.Sleep(time.Second * d.frequency)
		doCollect := <-d.c
		if doCollect == 0 {
			break
		}
	}
}

func (d *DiskStats) CollectStop() {
	d.c <- 0
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func statCounter(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		n = 0
	}
	return n
}
func (d *DiskStats) collect() {
	filename := "/proc/diskstats"
	s, err := readLines(filename)
	if err != nil && err != io.EOF {
		return
	}
	d.Lock()
	defer d.Unlock()
	for _, v := range s {
		values := strings.Fields(v)
		if len(values) != 14 {
			fmt.Printf("len(values) %v is not 14", len(values))
			continue
		}
		d.stats[values[2]] = &Stats{
			Reads:        statCounter(values[3]),
			ReadsMerged:  statCounter(values[4]),
			ReadSectors:  statCounter(values[5]),
			ReadMs:       statCounter(values[6]),
			Writes:       statCounter(values[7]),
			WritesMerged: statCounter(values[8]),
			WriteSectors: statCounter(values[9]),
			WriteMs:      statCounter(values[10]),
			IOProgress:   statCounter(values[11]),
			IOWeightedMS: statCounter(values[12]),
		}
	}
}
