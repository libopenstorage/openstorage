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

To make things simple, packages for glog, logrus, protolog, and lion are given with the ability to blank import:

```go
import (
  _ "go.pedge.io/dlog/glog" // set glog as the global logger
  _ "go.pedge.io/dlog/logrus" // set logrus as the global logger with default settings
  _ "go.pedge.io/dlog/protolog" // set protolog as the global logger with default settings
  _ "go.pedge.io/dlog/lion" // set lion as the global logger with default settings
)
```

Or, do something more custom:

```go
import (
  "os"

  "go.pedge.io/dlog"
  dloglogrus "go.pedge.io/dlog/logrus"

  "github.com/Sirupsen/logrus"
)

func init() { // or anywhere
  logger := logrus.New()
  logger.Out = os.Stdout
  logger.Formatter = &logrus.TextFormatter{
    ForceColors: true,
  }
  dlog.SetLogger(dloglogrus.NewLogger(logger))
}
```

By default, golang's standard logger is used. This is not recommended, however, as the implementation
with the WithFields function is slow. It would be better to choose a different implementation in most cases.
