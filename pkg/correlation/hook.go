package correlation

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type LogHook struct {
	Component       Component
	FunctionContext context.Context
}

var _ logrus.Hook = &LogHook{}

const (
	// LogFieldID represents a logging field for IDs
	LogFieldID = "correlation-id"
	// LogFieldID represents a logging field for the request origin
	LogFieldOrigin = "origin"
	// LogFieldComponent represents a logging field for control plane component.
	// This is typically set per-package and allows you see which component
	// the log originated from.
	LogFieldComponent = "component"
)

// Levels describes which levels this logrus hook
// should run with.
func (lh *LogHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.InfoLevel}
}

// Fire is used to add correlation context info in each log line
func (lh *LogHook) Fire(entry *logrus.Entry) error {

	// Default to tne entry context. This is populated
	// by logrus.WithContext.Infof(...)
	ctx := entry.Context
	if ctx == nil && lh.FunctionContext != nil {
		// If WithContext is not provided, and a function context
		// is provided, use that context.
		ctx = lh.FunctionContext
	}

	// If a context has been found, we will populate the correlation info
	if ctx != nil {
		correlationContext, ok := ctx.Value(ContextKey).(*RequestContext)
		if !ok {
			return fmt.Errorf("failed to get context for correlation logging hook")
		}

		entry.Data[LogFieldID] = correlationContext.ID
		entry.Data[LogFieldOrigin] = correlationContext.Origin
	}

	// Only add component when provided
	if len(lh.Component) > 0 {
		entry.Data[LogFieldComponent] = lh.Component
	}

	return nil
}

// NewPackageLogger creates a package-level logger for a given component
func NewPackageLogger(component Component) *logrus.Logger {
	clogger := logrus.New()
	clogger.AddHook(&LogHook{
		Component: component,
	})

	return clogger
}

// NewFunctionLogger creates a logger for usage at a per-function level
// For example, this logger can be instantiated inside of a function with a given
// context object. As logs are printed, they will automatically include the correlation
// context info.
func NewFunctionLogger(ctx context.Context, component Component) *logrus.Logger {
	clogger := logrus.New()
	clogger.AddHook(&LogHook{
		Component:       component,
		FunctionContext: ctx,
	})

	return clogger
}

// RegisterGlobalHook will register the correlation logging
// hook at a global/multi-package level. Note that this is not
// component-aware and it is better to use NewFunctionLogger
// or NewPackageLogger for logging.
func RegisterGlobalHook() {
	logrus.AddHook(&LogHook{})
}
