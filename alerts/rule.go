package alerts

// Rule defines an rule on an Event, for a filter, executing a func.
type Rule interface {
	GetName() string
	GetEvent() Event
	GetFilter() Filter
	GetAction() Action
}

// Event contains Event information.
type Event int

// Event constants
const (
	RaiseEvent Event = iota
	DeleteEvent
)

// rule implements Rule interface.
type rule struct {
	name   string
	event  Event
	filter Filter
	action Action
}

func (a *rule) GetName() string {
	return a.name
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
