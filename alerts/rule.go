package alerts

// Rule defines an rule on an Event, for a filter, executing a func.
type Rule interface {
	GetEvent() Event
	GetFilter() Filter
	GetAction() Action
}

// Event contains Event information.
type Event string

// Event constants
const (
	RaiseEvent  Event = "onRaise"
	DeleteEvent Event = "onDelete"
)

// rule implements Rule interface.
type rule struct {
	event  Event
	filter Filter
	action Action
}

func (a *rule) GetEvent() Event {
	return a.event
}

func (a *rule) GetFilter() Filter {
	return a.filter
}

func (a *rule) GetAction() Action {
	return a.action
}

// OnRaise creates a rule that activates when a raised alerts matches filter.
func NewRaiseRule(filter Filter, action Action) Rule {
	return &rule{event: RaiseEvent, filter: filter, action: action}
}

// OnDelete creates a rule that activates when deleted alert matches filter.
func NewDeleteRule(filter Filter, action Action) Rule {
	return &rule{event: DeleteEvent, filter: filter, action: action}
}
