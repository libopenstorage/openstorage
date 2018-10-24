package pcqueue

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSemaphore(t *testing.T) {
	for size := uint32(1); size < 10; size++ {
		s := NewSemaphore(size)
		m := &sync.Mutex{}
		count := uint32(0)
		maxEntries := 100
		doneCh := make(chan struct{})
		testf := func(c *uint32, max *uint32, totalDone *int) {
			s.Wait()
			m.Lock()
			(*c)++
			require.True(t, *c <= *max, "number of simultaneous entries")
			m.Unlock()

			m.Lock()
			require.True(t, *c > 0, "number of tasks")
			(*c)--
			(*totalDone)++
			if *totalDone == maxEntries {
				close(doneCh)
			}
			s.Signal()
			m.Unlock()
		}
		totalEntries := 0
		for i := 0; i < maxEntries; i++ {
			go testf(&count, &size, &totalEntries)
		}
		select {
		case <-time.After(time.Second * 60):
			require.Fail(t, "timeout in test")
		case <-doneCh:
			break
		}
	}
}
