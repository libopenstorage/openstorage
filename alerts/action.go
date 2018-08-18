package alerts

// Action is something that runs for an alerts manager.
type Action interface {
	Run(manager Manager) error
}

// ActionType defines categories of actions.
type ActionType int

// ActionType constants
const (
	DeleteAction ActionType = iota
	CustomAction
)

// action implements Action interface.
type action struct {
	action  ActionType
	filters []Filter
	f       func(manager Manager, filter ...Filter) error
}

func (a *action) Run(manager Manager) error {
	return a.f(manager, a.filters...)
}

func NewDeleteAction(filter ...Filter) Action {
	return &action{action: DeleteAction, filters: filter, f: deleteAction}
}

func NewCustomAction(f func(manager Manager, filter ...Filter) error, filter ...Filter) Action {
	return &action{action: CustomAction, filters: filter, f: f}
}

func deleteAction(manager Manager, filter ...Filter) error {
	return manager.Delete(filter...)
}
