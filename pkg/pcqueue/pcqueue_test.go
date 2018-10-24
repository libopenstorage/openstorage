package pcqueue

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	for size := uint8(1); size < 10; size++ {
		q := New(size)
		m := &sync.Mutex{}
		count := 0
		doneCh := make(chan struct{}, 1)
		numTasks := 40
		tasks := make([]int, numTasks)
		for i := 0; i < numTasks; i++ {
			j := i
			c := &count
			// Enqueue numTasks unique task and make sure all of them
			// are dequeued and run
			f := func() {
				m.Lock()
				tasks[j] = j
				*c++
				if *c == numTasks {
					close(doneCh)
				}
				m.Unlock()
			}
			q.Enqueue(f)
		}
		select {
		case <-time.After(time.Second * 30):
			break
		case <-doneCh:
			break
		}
		for i := 0; i < numTasks; i++ {
			require.Equal(t, tasks[i], i, "task id")
		}
		require.Equal(t, numTasks, count, "number of tasks")
	}
}

func TestSafeQueue(t *testing.T) {
	q := NewSafeQueue()
	for i := 1; i < 10; i++ {
		for j := 0; j < 10; j++ {
			q.Push(fmt.Sprintf("%d ", j))
		}
		vals := q.PopAll()
		require.Equal(t, 10, len(vals))
		for j := 0; j < 10; j++ {
			require.Equal(t, vals[j].(string), fmt.Sprintf("%d ", j))
		}
	}
}
