package pcqueue

import (
	"sync"

	"github.com/libopenstorage/openstorage/pkg/dbg"
)

type semaphore struct {
	// max is the number of available resources
	max uint32
	// curr is the currently used resources
	curr uint32
	// m protects curr
	m *sync.Mutex
	// cv coordinates wait/signal
	cv *sync.Cond
}

type Semaphore interface {
	// Wait semaphore, will block until free resource is available
	Wait()
	// TryWait semaphore, will return boolean based on free resource availability
	TryWait() bool
	// Signal semaphore, asserts if resource not used.
	Signal()
}

func NewSemaphore(size uint32) Semaphore {
	dbg.Assert(size > 0, "Size must be greater than zero")
	mtx := &sync.Mutex{}
	s := &semaphore{
		m:    mtx,
		cv:   sync.NewCond(mtx),
		max:  size,
		curr: size}
	return s
}

func (s *semaphore) Wait() {
	s.m.Lock()
	for s.curr == 0 {
		s.cv.Wait()
	}
	dbg.Assert(s.curr > 0, "max error, curr: %d", s.curr)
	s.curr--
	s.m.Unlock()
}

func (s *semaphore) TryWait() bool {
	s.m.Lock()
	defer s.m.Unlock()
	if s.curr == 0 {
		return false
	}
	dbg.Assert(s.curr > 0, "max error, curr: %d", s.curr)
	s.curr--
	return true
}

func (s *semaphore) Signal() {
	s.m.Lock()
	s.curr++
	dbg.Assert(s.curr <= s.max, "max mismatch, max: %d curr: %d", s.max, s.curr)
	s.cv.Signal()
	s.m.Unlock()
}
