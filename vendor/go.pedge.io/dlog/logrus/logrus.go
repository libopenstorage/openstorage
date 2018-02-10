/*
Package dlog_logrus provides logrus functionality for dlog.

https://github.com/sirupsen/logrus
*/
package dlog_logrus // import "go.pedge.io/dlog/logrus"

import (
	"go.pedge.io/dlog"

	"github.com/sirupsen/logrus"
)

var (
	levelToLogrusLevel = map[dlog.Level]logrus.Level{
		dlog.LevelNone:  logrus.InfoLevel,
		dlog.LevelDebug: logrus.DebugLevel,
		dlog.LevelInfo:  logrus.InfoLevel,
		dlog.LevelWarn:  logrus.WarnLevel,
		dlog.LevelError: logrus.ErrorLevel,
		dlog.LevelFatal: logrus.FatalLevel,
		dlog.LevelPanic: logrus.PanicLevel,
	}
)

// Register registers the default logrus Logger as the dlog Logger.
func Register() {
	dlog.SetLogger(NewLogger(logrus.StandardLogger()))
}

// NewLogger returns a new dlog.Logger that uses the logrus.Logger.
func NewLogger(logrusLogger *logrus.Logger) dlog.Logger {
	return newLogger(&loggerLogrusLogger{logrusLogger})
}

type logrusLogger interface {
	dlog.BaseLogger
	WithField(key string, value interface{}) *logrus.Entry
	WithFields(fields logrus.Fields) *logrus.Entry
	SetLevel(level dlog.Level)
}

type loggerLogrusLogger struct {
	*logrus.Logger
}

func (l *loggerLogrusLogger) SetLevel(level dlog.Level) {
	l.Logger.Level = levelToLogrusLevel[level]
}

type entryLogrusLogger struct {
	*logrus.Entry
}

func (l *entryLogrusLogger) SetLevel(level dlog.Level) {
	l.Entry.Level = levelToLogrusLevel[level]
}

type logger struct {
	dlog.BaseLogger
	l logrusLogger
}

func newLogger(l logrusLogger) *logger {
	return &logger{l, l}
}

func (l *logger) AtLevel(level dlog.Level) dlog.Logger {
	// TODO(pedge): not thread safe, tradeoff here
	// TODO(pedge): neither implementation checks map, even though we expect coverage
	l.l.SetLevel(level)
	return l
}

func (l *logger) WithField(key string, value interface{}) dlog.Logger {
	return newLogger(&entryLogrusLogger{l.l.WithField(key, value)})
}

func (l *logger) WithFields(fields map[string]interface{}) dlog.Logger {
	return newLogger(&entryLogrusLogger{l.l.WithFields(fields)})
}
