package probation

import (
	"fmt"
	"sync"
	"time"

	"github.com/libopenstorage/openstorage/pkg/sched"
)

// Probation is an interface that defines a set of APIs to manage a probation
// list. The probation list is guarded by a configurable timeout. Any clients
// that stay in the probation list after the probation timeout will be passed on
// to the registered ProbationCallback fn
type Probation interface {
	// Add adds a client to the probation list.
	Add(clientID string, clientData interface{}, updateIfExists bool) error
	// Remove removes a client from the probation list.
	Remove(clientID string) error
	// Exists checks if a client is already in the probation list
	Exists(clientID string) bool
	// Start starts monitoring the probationList with the configured
	// probationTimeout
	Start() error
}

type probation struct {
	name             string
	probationTimeout time.Duration
	probationTasks   map[string]sched.TaskID
	mutex            sync.Mutex
	pcf              CallbackFn
	schedInst        sched.Scheduler
}

// CallbackFn will be executed for every client that gets expired in
// the probation list.
type CallbackFn func(clientID string, clientData interface{}) error

// NewProbationManager returns the default implementation of Probation interface
// It takes in a name to be associated with this probation manager and a
// ProbationCallbackFn
func NewProbationManager(
	name string,
	probationTimeout time.Duration,
	pcf CallbackFn,
) Probation {
	if sched.Instance() == nil {
		sched.Init(1 * time.Second)
	}
	p := &probation{
		name:             name,
		probationTimeout: probationTimeout,
		pcf:              pcf,
		probationTasks:   make(map[string]sched.TaskID),
		schedInst:        sched.Instance(),
	}
	return p
}

func (p *probation) Add(clientID string, clientData interface{}, updateIfExists bool) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	existingTaskID, exists := p.probationTasks[clientID]
	if exists {
		if updateIfExists {
			if err := p.schedInst.Cancel(existingTaskID); err != nil {
				return fmt.Errorf("probation manager failed to delete previous task for %s due to: %v", clientID, err)
			}

			delete(p.probationTasks, clientID)
		} else {
			return nil
		}
	}

	taskID, err := p.schedInst.Schedule(
		func(sched.Interval) { p.pcf(clientID, clientData) },
		sched.Periodic(time.Second),
		time.Now().Add(p.probationTimeout), /* run probationTimeout after current time */
		true) /* run only once */
	if err != nil {
		return err
	}

	p.probationTasks[clientID] = taskID
	return nil
}

func (p *probation) Exists(clientID string) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	_, ok := p.probationTasks[clientID]
	return ok
}

func (p *probation) Remove(clientID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	taskID, ok := p.probationTasks[clientID]
	if !ok {
		return nil // nothing to do
	}

	delete(p.probationTasks, clientID)
	p.schedInst.Cancel(taskID) // not checking Cancel err since it could fail if the task is already complete
	return nil
}

// TODO the start shouldn't really be needed. Keeping it to maintain previous behavior
func (p *probation) Start() error {
	p.schedInst.Start()
	return nil
}
