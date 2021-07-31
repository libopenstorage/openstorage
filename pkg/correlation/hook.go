package correlation

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type LogHook struct {
	Component Component
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
	correlationContext, ok := entry.Context.Value(ContextKey).(*RequestContext)
	if !ok {
		return fmt.Errorf("failed to get context for correlation logging hook")
	}

	entry.Data[LogFieldID] = correlationContext.ID
	entry.Data[LogFieldOrigin] = correlationContext.Origin

	// Only add component when provided
	if len(lh.Component) > 0 {
		entry.Data[LogFieldComponent] = lh.Component
	}
	return nil
}

// NewLogger creates a package-level logger for a given component
func NewLogger(component Component) *logrus.Logger {
	clogger := logrus.New()
	clogger.AddHook(&LogHook{
		Component: component,
	})

	return clogger
}
