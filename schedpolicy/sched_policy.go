//go:generate mockgen -package=mock -destination=mock/schedpolicy.mock.go github.com/libopenstorage/openstorage/schedpolicy SchedulePolicy
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
