package pcqueue

import (
	"github.com/libopenstorage/openstorage/pkg/dbg"
	"sync"
)

// WaitGroupCh is similar to sync.WaitGroup.
// It provides additional WaitCh() so that it can be used to wait in a select.
type WaitGroupCh interface {
	// Done decrements counter for wait group
	Done()
	// WaitCh returns a receive channel which returns when count becomes zero.
	WaitCh() <-chan struct{}
	// Wait blocks until counter is zero
	Wait()
}

type waitGroup struct {
	sync.Mutex
	count  int
	doneCh chan struct{}
}

func NewWaitGroupCh(count int) WaitGroupCh {
	return &waitGroup{count: count, doneCh: make(chan struct{})}
}

func (w *waitGroup) Done() {
	w.Lock()
	defer w.Unlock()
	dbg.Assert(w.count > 0, "mismatched Done")
	w.count--
	if w.count == 0 {
		close(w.doneCh)
	}
}

func (w *waitGroup) WaitCh() <-chan struct{} {
	return w.doneCh
}

func (w *waitGroup) Wait() {
	<-w.doneCh
}
