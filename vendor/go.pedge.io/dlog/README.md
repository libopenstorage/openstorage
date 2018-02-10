[![CircleCI](https://circleci.com/gh/peter-edge/dlog-go/tree/master.png)](https://circleci.com/gh/peter-edge/dlog-go/tree/master)
[![Go Report Card](http://goreportcard.com/badge/peter-edge/dlog-go)](http://goreportcard.com/report/peter-edge/dlog-go)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/go.pedge.io/dlog)
[![MIT License](http://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/peter-edge/dlog-go/blob/master/LICENSE)

dlog (delegating log) wraps common functionality for common golang logging packages.

The `dlog.Logger` interface wraps the common logging functionality. Every method on `dlog.Logger`
is also a global method on the `dlog` package. Given an implementation of `dlog.Logger`, you can
register it as the global logger by calling:

```go
func register(logger dlog.Logger) {
  dlog.SetLogger(logger)
}
```

To make things simple, packages for glog, logrus, log15, and lion are given with the ability to easily register
their implementations as the default logger:

```go
import (
  "go.pedge.io/dlog/glog"
  "go.pedge.io/dlog/lion"
  "go.pedge.io/dlog/log15"
  "go.pedge.io/dlog/logrus"
  "go.pedge.io/dlog/zap"
)

func registrationFunctions() {
  dlog_glog.Register() // set glog as the global logger
  dlog_lion.Register() // set lion as the global logger with default settings
  dlog_log15.Register() // set log15 as the global logger with default settings
  dlog_logrus.Register() // set logrus as the global logger with default settings
  dlog_zap.Register() // set zap as the global logger with default settings
}
```

Or, do something more custom:

```go
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
```

By default, golang's standard logger is used. This is not recommended, however, as the implementation
with the WithFields function is slow. It would be better to choose a different implementation in most cases.
