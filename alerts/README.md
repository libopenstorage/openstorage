# alerts
pkg `alerts` provides a way to manage alerts by allowing you to express various querying and 
filtering conditions.

## Intro
This package works using following high level objects:
* Filters: These are conditions used to match an alert. A filter can match on alert attributes
such as alert type, resource type, resource id, count, time and clear flag.
* Actions: An action is performed on alerts that match a set of filters. Typically, action is taken
during an event, such as an alert raise event or alert delete event.
* Rules: A rule specifically identifies an action to be taken on an event.

To further define these objects, options are required, which allow configuration. Options can be passed during creation
of such objects.

Before we look at details of individual objects, let's review the high level API to manage raising, deleting and
enumerating alerts.

## API
The package provides a way to create an instance of an alerts manager. An alerts manager is an interface and provides
following methods to work with alerts
```go
// Manager manages alerts.
type Manager interface {
	// Raise raises an alert.
	Raise(alert *api.Alert) error
	// Enumerate lists all alerts filtered by a variadic list of filters.
	// It will fetch a superset such that every alert is matched by at least one filter.
	Enumerate(filters ...Filter) ([]*api.Alert, error)
	// Filter filters given list of alerts successively through each filter.
	Filter(alerts []*api.Alert, filters ...Filter) ([]*api.Alert, error)
	// Delete deletes alerts filtered by a chain of filters.
	Delete(filters ...Filter) error
	// SetRules sets a set of rules to be performed on alert events.
	SetRules(rules ...Rule)
	// DeleteRules deletes rules
	DeleteRules(rules ...Rule)
}
```

The package also provides an implementation of this interface via `kvdb`. So an instance of alerts manager could be 
created as follows:
```go
// NewManager obtains instance of Manager for alerts management.
func NewManager(kv kvdb.Kvdb, options ...Option) (Manager, error) {...}

// NewTTLOption provides an option to be used in manager creation.
func NewTTLOption(ttl uint64) Option {...}
```

Currently, it accepts only one type of option, which is TTL option. A `TTL` value indicates how long a cleared
alert should live in kvdb. Please refer to `kvdb` doc for more details.

As shown in the definition of alerts manager interface, the API makes use of filters and rules. So what are these objects?

# Filters
A `Filter` is an interface used to define a matching condition on an alert. 
You can obtain an instance of this interface using following API:
```go
// Filter API

// NewResourceTypeFilter creates a filter that matches on <resourceType>
func NewResourceTypeFilter(resourceType api.ResourceType, options ...Option) Filter {...}

// NewAlertTypeFilter creates a filter that matches on <resourceType>/<alertType>
func NewAlertTypeFilter(alertType int64, resourceType api.ResourceType, options ...Option) Filter {...}

// NewResourceIDFilter creates a filter that matches on <resourceType>/<alertType>/<resourceID>
func NewResourceIDFilter(resourceID string, alertType int64, resourceType api.ResourceType, options ...Option) Filter {...}

// NewTimeSpanFilter creates a filter that matches on alert raised in a given time window.
func NewTimeSpanFilter(start, stop time.Time) Filter {...}

// NewMatchResourceIDFilter provides a filter that matches on resource id.
func NewMatchResourceIDFilter(resourceID string) Filter {...}

// NewCountSpanFilter provides a filter that matches on alert count.
func NewCountSpanFilter(minCount, maxCount int64) Filter {...}

// NewMinSeverityFilter provides a filter that matches on alert when severity is greater than
// or equal to the minSev value.
func NewMinSeverityFilter(minSev api.SeverityType) Filter {...}

// NewFlagCheckFilter provides a filter that matches on alert clear flag.
func NewFlagCheckFilter(flag bool) Filter {...}

// NewCustomFilter creates a filter that matches on UDF (user defined function)
func NewCustomFilter(f func(alert *api.Alert) (bool, error)) Filter {...}
```

As you can see three of these filters take options for further configuration. These options provide a way to filter
out alerts. An option is also an interface and following options can be created.

```go
// NewTimeSpanOption provides an option to be used in filter definition.
// Filters that take options, apply options only during matching alerts.
func NewTimeSpanOption(start, stop time.Time) Option {...}

// NewCountSpanOption provides an option to be used in filter definition that
// accept options. Only filters that are efficient in querying kvdb accept options
// and apply these options during matching alerts.
func NewCountSpanOption(minCount, maxCount int64) Option {...}

// NewMinSeverityOption provides an option to be used during filter creation that
// accept such options. Only filters that are efficient in querying kvdb accept options
// and apply these options during matching alerts.
func NewMinSeverityOption(minSev api.SeverityType) Option {...}

// NewFlagCheckOptions provides an option to be used during filter creation that
// accept such options. Only filters that are efficient in querying kvdb accept options
// and apply these options during matching alerts.
func NewFlagCheckOption(flag bool) Option {...}

// NewresourceIdOption provides an option to be used during filter creation that
// accept such options. Only filters that are efficient in querying kvdb accept options
// and apply these options during matching alerts.
func NewResourceIdOption(resourceId string) Option {...}
```

A filter with an option works in a way that _all_ conditions need to be satisfied. It is an AND operation.

## Rules
A rule is something that is defined by a client but managed by the alerts manager bookkeeping. Every instance
of alerts manager has its own list of rules that it applies on every qualifying event.

A rule takes a single filter that it matches against an alert triggering an event. For instance, it would match
this input filter against the alert being raised if the qualifying event is a raise event.

Different rules for different qualifying events can be obtained as follows:
```go
// NewRaiseRule creates a rule that runs action when a raised alerts matche filter.
// Action happens before incoming alert is raised.
func NewRaiseRule(name string, filter Filter, action Action) Rule {...}

// NewDeleteRule creates a rule that runs action when deleted alerts matche filter.
// Action happens after matching alerts are deleted.
func NewDeleteRule(name string, filter Filter, action Action) Rule {...}
```

Each rule also takes an action. So what is an action?

## Action
An action is an interface that will trigger on matching alerts. For instance, an action could be deleting a set of
alerts by using a criteria. Therefore, an action is essentially a set of filter and a function that runs on each of
of the matched alert.

An instance of an action can be obtained as follows:

```go
// NewDeleteAction deletes alert entries based on filters.
func NewDeleteAction(filters ...Filter) Action {...}

// NewClearAction marks alert entries as cleared that get deleted after half a day of life in kvdb.
func NewClearAction(filters ...Filter) Action {...}

// NewCustomAction takes custom action using user defined function.
func NewCustomAction(f func(manager Manager, filters ...Filter) error, filters ...Filter) Action {...}
```

The list of filters here work together in such a way that an action will be performed on an alert as long as
at least one of the filter matches with the alert. It is an OR operation.
