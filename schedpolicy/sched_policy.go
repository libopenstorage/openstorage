//go:generate mockgen -package=mock -destination=mock/schedpolicy.mock.go github.com/libopenstorage/openstorage/schedpolicy SchedulePolicy
package schedpolicy

import "errors"

var (
	ErrNotImplemented = errors.New("Not Implemented")
)

type SchedulePolicy interface {
	// SchedPolicyCreate creates a policy with given name and schedule.
	SchedPolicyCreate(name, sched string) error
	// SchedPolicyUpdate updates a policy with given name and schedule.
	SchedPolicyUpdate(name, sched string) error
	// SchedPolicyDelete deletes a policy with given name.
	SchedPolicyDelete(name string) error
	// SchedPolicyEnumerate enumerates all configured policies or the ones specified.
	SchedPolicyEnumerate([]string) ([]*SchedPolicy, error)
}

type Manager struct {
	SchedulePolicy
}

func NewSchedulePolicyManager(sched SchedulePolicy) *Manager {
	return &Manager{
		SchedulePolicy: sched,
	}
}

type nullSchedMgr struct {
}

func New() *nullSchedMgr {
	return &nullSchedMgr{}
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

func (sp *nullSchedMgr) SchedPolicyEnumerate([]string) ([]*SchedPolicy, error) {
	return nil, ErrNotImplemented
}
