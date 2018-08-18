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

func deleteAction(manager Manager, filter ...Filter) error {
	return manager.Delete(filter...)
}
