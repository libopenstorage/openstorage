package schedpolicy

import "errors"

var (
	ErrNotImplemented = errors.New("Not Implemented")
)

type SchedulePolicyProvider interface {
	// SchedPolicyCreate creates a policy with given name and schedule.
	SchedPolicyCreate(name, sched string) error
	// SchedPolicyUpdate updates a policy with given name and schedule.
	SchedPolicyUpdate(name, sched string) error
	// SchedPolicyDelete deletes a policy with given name.
	SchedPolicyDelete(name string) error
	// SchedPolicyEnumerate enumerates all configured policies or the ones specified.
	SchedPolicyEnumerate() ([]*SchedPolicy, error)
	// SchedPolicyGet returns schedule policy matching given name.
	SchedPolicyGet(name string) (*SchedPolicy, error)
}

func NewDefaultSchedulePolicy() SchedulePolicyProvider {
	return &nullSchedMgr{}
}

type nullSchedMgr struct {
}

func (sp *nullSchedMgr) SchedPolicyCreate(name, sched string) error {
	return ErrNotImplemented
}

func (sp *nullSchedMgr) SchedPolicyUpdate(name, sched string) error {
	return ErrNotImplemented
}

func (sp *nullSchedMgr) SchedPolicyDelete(name string) error {
	return ErrNotImplemented
}

func (sp *nullSchedMgr) SchedPolicyEnumerate() ([]*SchedPolicy, error) {
	return nil, ErrNotImplemented
}

func (sp *nullSchedMgr) SchedPolicyGet(name string) (*SchedPolicy, error) {
	return nil, ErrNotImplemented
}
