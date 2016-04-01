/*
Package grpclion provides a logger for grpclog.
*/
package grpclion // import "go.pedge.io/lion/grpc"

import (
	"go.pedge.io/lion"
	"google.golang.org/grpc/grpclog"
)

// RedirectGrpclog will redirect grpclog to lion.
func RedirectGrpclog() {
	lion.AddGlobalHook(
		func(l lion.Logger) {
			grpclog.SetLogger(NewLogger(l))
		},
	)
}

// NewLogger returns a new grpclog.Logger for the BaseLogger.
func NewLogger(baseLogger lion.BaseLogger) grpclog.Logger {
	return &logger{baseLogger}
}

type logger struct {
	lion.BaseLogger
}

func (l *logger) Fatal(args ...interface{}) {
	l.BaseLogger.Fatalln(args...)
}

func (l *logger) Print(args ...interface{}) {
	l.BaseLogger.Println(args...)
}
