package alerts

// Action is something that runs for an alerts manager.
type Action interface {
	Run(manager Manager) error
}

// ActionType defines categories of actions.
type ActionType int

// ActionType constants
const (
	// Deletes alert entries.
	deleteAction ActionType = iota
	// Alerts get marked for deletion.
	clearAction
	// Custom user action.
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

func deleteActionFunc(manager Manager, filters ...Filter) error {
	return manager.Delete(filters...)
}

// clearActionFunc first enumerates, then changes Cleared flag to true,
// then updates it.
// Raise method determines if ttlOption needs to be applied based on clear flag.
func clearActionFunc(manager Manager, filters ...Filter) error {
	myAlerts, err := manager.Enumerate(filters...)
	if err != nil {
		return err
	}

	for _, myAlert := range myAlerts {
		myAlert.Cleared = true
		if err := manager.Raise(myAlert); err != nil {
			return err
		}
	}

	return nil
}
