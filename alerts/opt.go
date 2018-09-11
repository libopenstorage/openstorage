package alerts

// OptionType defines type of option for alert manager setup.
type OptionType int

const (
	// ttlOption defines the time to live for alerts that are cleared (default half day)
	// ttlOption is only valid for alerts manager creation.
	ttlOption OptionType = iota
	// timeSpanOption provides a way to tell a filter that it should also apply filtering based
	// on the time span. Such option is useful for creating efficient filters that fetch efficiently
	// from kvdb and apply filtering after fetching.
	timeSpanOption
	// countSpanOption provides a way to tell filter that it should apply filtering based on
	// the count span. Such option is useful for creating efficient filters that fetch efficiently
	// from kvdb and apply filtering after fetching.
	countSpanOption
	// minSeverityOption provides a way to tell filter that it should apply filtering based on
	// the severity. Such option is useful for creating efficient filters that fetch efficiently
	// from kvdb and apply filtering after fetching.
	minSeverityOption
	// flagCheckOption provides a way to tell filter that it should apply filtering based on
	// the clear flag. Such option is useful for creating efficient filters that fetch efficiently
	// from kvdb and apply filtering after fetching.
	flagCheckOption
	// resourceIDOption provides a way to tell filter that it should apply filtering based on
	// the resource id. Such option is useful for creating efficient filters that fetch efficiently
	// from kvdb and apply filtering after fetching.
	resourceIdOption
)

// Option defines what is an option.
type Option interface {
	GetType() OptionType
	GetValue() interface{}
}

// option implements Option interface.
type option struct {
	optionType OptionType
	value      interface{}
}

func (o *option) GetType() OptionType {
	return o.optionType
}

func (o *option) GetValue() interface{} {
	return o.value
}
