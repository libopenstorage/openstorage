package osdconfig

import (
	"container/heap"
	"math"
	"testing"
	"time"
)

func TestHeap(t *testing.T) {
	d := new(jobHeap)
	heap.Init(d)
	t0 := time.Now().UnixNano()
	for i := 0; i < 5; i++ {
		e := new(dataWrite)
		e.Value = []byte{byte(i)}
		e.Time = time.Now().UnixNano() - t0
		heap.Push(d, e)
	}

	ref := int64(math.MinInt64)
	for d.Len() > 0 {
		e := heap.Pop(d).(*dataWrite)
		if e.Time < ref {
			t.Fatal("heap not working")
		} else {
			ref = e.Time
		}
	}
}
