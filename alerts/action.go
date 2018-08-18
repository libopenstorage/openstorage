package alerts

type Action interface {
	Run(manager Manager)
}

type ActionType string

const (
	DeleteAction ActionType = "delete"
	CustomAction ActionType = "custom"
)

type action struct {
	action ActionType
	filter Filter
	f      func(manager Manager, filter Filter)
}

func (a *action) Run(manager Manager) {
	a.f(manager, a.filter)
}

func NewDeleteAction(filter Filter) Action {
	return &action{action: DeleteAction, filter: filter, f: deleteAction}
}

func NewCustomAction(filter Filter, f func(manager Manager, filter Filter)) Action {
	return &action{action: CustomAction, filter: filter, f: f}
}

func deleteAction(manager Manager, filter Filter) {
	manager.Delete(filter)
}
