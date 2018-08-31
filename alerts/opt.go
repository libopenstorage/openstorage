package alerts

// OptionType defines type of option for alert manager setup.
type OptionType int

const (
	// TTLOption defines the time to live for alerts that are cleared (default half day)
	TTLOption OptionType = iota
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
