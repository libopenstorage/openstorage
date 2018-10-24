package pcqueue

import (
	"container/list"
	"sync"

	"github.com/libopenstorage/openstorage/pkg/dbg"
	"github.com/sirupsen/logrus"
)

type Task func()

type taskQueue struct {
	// taskQ is the list of tasks
	taskQ *list.List
	// m is the mutex to protect taskQ
	m *sync.Mutex
	// cv is used to coordinate the producer-consumer threads
	cv *sync.Cond
}

// PCQueue is a producer consumer queue.
type PCQueue interface {
	// Enqueue will enqueue an update. It is non-blocking.
	Enqueue(t Task)
	// Dump will dump information about the producer consumer queue.
	Dump()
}

// NewPCQueue returns PCQueue
func New(numWorkers uint8) PCQueue {
	dbg.Assert(numWorkers > 0, "Number of workers must be greater than zero")
	mtx := &sync.Mutex{}
	q := &taskQueue{
		m:     mtx,
		cv:    sync.NewCond(mtx),
		taskQ: list.New()}
	workerFn := func() { q.dequeue() }
	for i := uint8(0); i < numWorkers; i++ {
		go workerFn()
	}
	return q
}

func (w *taskQueue) dequeue() {
	for {
		w.m.Lock()
		for w.taskQ.Len() == 0 {
			w.cv.Wait()
		}
		el := w.taskQ.Front()
		w.taskQ.Remove(el)
		w.m.Unlock()
		el.Value.(Task)()
	}
}

func (w *taskQueue) Dump() {
	logrus.Infof("Queue size: %d", w.taskQ.Len())
}

// Enqueue enqueues and never blocks
func (w *taskQueue) Enqueue(t Task) {
	w.m.Lock()
	w.taskQ.PushBack(t)
	w.cv.Signal()
	w.m.Unlock()
}
