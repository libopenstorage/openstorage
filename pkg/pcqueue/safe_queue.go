package pcqueue

import (
	"container/list"
	"sync"
)

// SafeQueue is a queue which can be accessed by multiple threads safely
type SafeQueue interface {
	// Push item at the end of the queue
	Push(item interface{})
	// PopAll pops all items from the the queue
	PopAll() []interface{}
}

type safeQueue struct {
	// q is the list of items
	q *list.List
	// qLock protects access to q
	qLock sync.Mutex
}

func (s *safeQueue) Push(item interface{}) {
	s.qLock.Lock()
	s.q.PushBack(item)
	s.qLock.Unlock()
}

func (s *safeQueue) PopAll() []interface{} {
	s.qLock.Lock()
	defer s.qLock.Unlock()
	num := s.q.Len()
	ret := make([]interface{}, num)
	for i := 0; i < num; i++ {
		el := s.q.Front()
		s.q.Remove(el)
		ret[i] = el.Value
	}
	return ret
}

func NewSafeQueue() SafeQueue {
	return &safeQueue{
		q: list.New(),
	}
}
