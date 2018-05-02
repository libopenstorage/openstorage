package schedpolicy

const (
	SchedName = "name"
)

// SchedPolicy specify name and schedule to create/update/list schedule policy
// swagger:model
type SchedPolicy struct {
	Name     string
	Schedule string
}
