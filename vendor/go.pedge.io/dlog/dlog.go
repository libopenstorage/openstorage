/*
Package dlog (delegating log) wraps common functionality for common golang logging packages.

The Logger interface wraps the common logging functionality. Every method on Logger
is also a global method on the dlog package. Given an implementation of Logger, you can
register it as the global logger by calling:

	func register(logger dlog.Logger) {
	  dlog.SetLogger(logger)
	}

To make things simple, packages for glog, logrus, log15, and lion are given with the ability to easily register
their implementations as the default logger:

	import (
	  "go.pedge.io/dlog/glog"
	  "go.pedge.io/dlog/lion"
	  "go.pedge.io/dlog/log15"
	  "go.pedge.io/dlog/logrus"
	)

	func registrationFunctions() {
	  dlog_glog.Register() // set glog as the global logger
	  dlog_lion.Register() // set lion as the global logger with default settings
	  dlog_log15.Register() // set log15 as the global logger with default settings
	  dlog_logrus.Register() // set logrus as the global logger with default settings
	}

Or, do something more custom:

	import (
	  "os"

	  "go.pedge.io/dlog"
	  "go.pedge.io/dlog/logrus"

	  "github.com/sirupsen/logrus"
	)

	func init() { // or anywhere
	  logger := logrus.New()
	  logger.Out = os.Stdout
	  logger.Formatter = &logrus.TextFormatter{
		ForceColors: true,
	  }
	  dlog.SetLogger(dlog_logrus.NewLogger(logger))
	}

By default, golang's standard logger is used. This is not recommended, however, as the implementation
with the WithFields function is slow. It would be better to choose a different implementation in most cases.
*/
package dlog // import "go.pedge.io/dlog"

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"unicode"
)

var (
	// DefaultLogger is the default Logger.
	DefaultLogger = NewStdLogger(log.New(os.Stderr, "", log.LstdFlags))
	// DefaultLevel is the default Level.
	DefaultLevel = LevelInfo

	globalLogger   = DefaultLogger
	globalLevel    = DefaultLevel
	globalLevelSet = false
	globalLock     = &sync.Mutex{}
)

// BaseLogger is the Logger's log functionality, split from WithField/WithFields for easier wrapping of other libraries.
type BaseLogger interface {
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Panicf(format string, args ...interface{})
	Panicln(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
}

// Logger is an interface that all logging implementations must implement.
type Logger interface {
	BaseLogger
	AtLevel(level Level) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// Register re-registers the default Logger as the dlog global Logger.
func Register() {
	SetLogger(DefaultLogger)
}

// SetLogger sets the global logger used by dlog.
func SetLogger(logger Logger) {
	globalLock.Lock()
	defer globalLock.Unlock()
	if globalLevelSet {
		logger = logger.AtLevel(globalLevel)
	}
	globalLogger = logger
}

// SetLevel sets the global Level.
func SetLevel(level Level) {
	globalLock.Lock()
	defer globalLock.Unlock()
	if globalLevel != level {
		globalLogger = globalLogger.AtLevel(level)
	}
	globalLevel = level
	globalLevelSet = true
}

// NewLogger creates a new Logger using a print function, and optionally
// specific Level to print functions (levelToPrintFunc can be nil).
//
// printFunc is used if a Level is not represented.
// LevelNone overrides printFunc.
//
// printFunc is required.
func NewLogger(printFunc func(...interface{}), levelToPrintFunc map[Level]func(...interface{})) Logger {
	return newLogger(globalLevel, printFunc, levelToPrintFunc)
}

// NewStdLogger creates a new Logger using a standard golang Logger.
func NewStdLogger(l *log.Logger) Logger {
	return newLogger(globalLevel, l.Println, nil)
}

// WithField calls WithField on the global Logger.
func WithField(key string, value interface{}) Logger {
	return globalLogger.WithField(key, value)
}

// WithFields calls WithFields on the global Logger.
func WithFields(fields map[string]interface{}) Logger {
	return globalLogger.WithFields(fields)
}

// Debugf logs at the debug level with the semantics of fmt.Printf.
func Debugf(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

// Debugln logs at the debug level with the semantics of fmt.Println.
func Debugln(args ...interface{}) {
	globalLogger.Debugln(args...)
}

// Infof logs at the info level with the semantics of fmt.Printf.
func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// Infoln logs at the info level with the semantics of fmt.Println.
func Infoln(args ...interface{}) {
	globalLogger.Infoln(args...)
}

// Warnf logs at the warn level with the semantics of fmt.Printf.
func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

// Warnln logs at the warn level with the semantics of fmt.Println.
func Warnln(args ...interface{}) {
	globalLogger.Warnln(args...)
}

// Errorf logs at the error level with the semantics of fmt.Printf.
func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

// Errorln logs at the error level with the semantics of fmt.Println.
func Errorln(args ...interface{}) {
	globalLogger.Errorln(args...)
}

// Fatalf logs at the fatal level with the semantics of fmt.Printf and exits with os.Exit(1).
func Fatalf(format string, args ...interface{}) {
	globalLogger.Fatalf(format, args...)
}

// Fatalln logs at the fatal level with the semantics of fmt.Println and exits with os.Exit(1).
func Fatalln(args ...interface{}) {
	globalLogger.Fatalln(args...)
}

// Panicf logs at the panic level with the semantics of fmt.Printf and panics.
func Panicf(format string, args ...interface{}) {
	globalLogger.Panicf(format, args...)
}

// Panicln logs at the panic level with the semantics of fmt.Println and panics.
func Panicln(args ...interface{}) {
	globalLogger.Panicln(args...)
}

// Printf logs at the info level with the semantics of fmt.Printf.
func Printf(format string, args ...interface{}) {
	globalLogger.Printf(format, args...)
}

// Println logs at the info level with the semantics of fmt.Println.
func Println(args ...interface{}) {
	globalLogger.Println(args...)
}

type logger struct {
	level            Level
	levelToPrintFunc map[Level]func(...interface{})
	fields           map[string]interface{}
}

func newLogger(initialLevel Level, printFunc func(...interface{}), levelToPrintFunc map[Level]func(...interface{})) *logger {
	if printFunc == nil {
		// really not a fan of this, but since this is generally called at initialization, just makes things
		// easier for now
		panic("dlog: printFunc is nil")
	}
	return &logger{initialLevel, getLevelToPrintFunc(printFunc, levelToPrintFunc), make(map[string]interface{}, 0)}
}

func (l *logger) AtLevel(level Level) Logger {
	return &logger{level, l.levelToPrintFunc, l.fields}
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return l.WithFields(map[string]interface{}{key: value})
}

func (l *logger) WithFields(fields map[string]interface{}) Logger {
	newFields := make(map[string]interface{}, len(l.fields)+len(fields))
	for key, value := range l.fields {
		newFields[key] = value
	}
	for key, value := range fields {
		newFields[key] = value
	}
	return &logger{l.level, l.levelToPrintFunc, newFields}
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.print(LevelDebug, fmt.Sprintf(format, args...))
}

func (l *logger) Debugln(args ...interface{}) {
	l.print(LevelDebug, fmt.Sprint(args...))
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.print(LevelInfo, fmt.Sprintf(format, args...))
}

func (l *logger) Infoln(args ...interface{}) {
	l.print(LevelInfo, fmt.Sprint(args...))
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.print(LevelWarn, fmt.Sprintf(format, args...))
}

func (l *logger) Warnln(args ...interface{}) {
	l.print(LevelWarn, fmt.Sprint(args...))
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.print(LevelError, fmt.Sprintf(format, args...))
}

func (l *logger) Errorln(args ...interface{}) {
	l.print(LevelError, fmt.Sprint(args...))
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.print(LevelFatal, fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (l *logger) Fatalln(args ...interface{}) {
	l.print(LevelFatal, fmt.Sprint(args...))
	os.Exit(1)
}

func (l *logger) Panicf(format string, args ...interface{}) {
	l.print(LevelPanic, fmt.Sprintf(format, args...))
	panic(fmt.Sprintf(format, args...))
}

func (l *logger) Panicln(args ...interface{}) {
	l.print(LevelPanic, fmt.Sprint(args...))
	panic(fmt.Sprint(args...))
}

func (l *logger) Printf(format string, args ...interface{}) {
	l.print(LevelNone, fmt.Sprintf(format, args...))
}

func (l *logger) Println(args ...interface{}) {
	l.print(LevelNone, fmt.Sprint(args...))
}

func (l *logger) print(level Level, value string) {
	if level < l.level && l.level != LevelNone {
		return
	}
	// expected to be ok since we covered this internally
	printFunc, ok := l.levelToPrintFunc[level]
	if !ok {
		printFunc, ok = l.levelToPrintFunc[LevelNone]
		if !ok {
			panic("dlog: cannot find any printFunc")
		}
	}
	fieldsString := l.getFieldsString()
	if fieldsString == "" {
		printFunc(value)
	} else {
		printFunc(fmt.Sprintf("%s %s", strings.TrimRightFunc(value, unicode.IsSpace), fieldsString))
	}
}

func (l *logger) getFieldsString() string {
	if len(l.fields) == 0 {
		return ""
	}
	values := make([]string, len(l.fields))
	i := 0
	for key, value := range l.fields {
		values[i] = fmt.Sprintf("%s=%v", key, value)
		i++
	}
	return strings.Join(values, " ")
}

func getLevelToPrintFunc(printFunc func(...interface{}), inputLevelToPrintFunc map[Level]func(...interface{})) map[Level]func(...interface{}) {
	levelToPrintFunc := make(map[Level]func(...interface{}))
	if inputLevelToPrintFunc != nil {
		for level, inputPrintFunc := range inputLevelToPrintFunc {
			levelToPrintFunc[level] = inputPrintFunc
		}
	}
	if _, ok := levelToPrintFunc[LevelNone]; !ok {
		levelToPrintFunc[LevelNone] = printFunc
	}
	for level := range levelToName {
		if _, ok := levelToPrintFunc[level]; !ok {
			levelToPrintFunc[level] = printFunc
		}
	}
	return levelToPrintFunc
}
