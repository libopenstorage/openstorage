/*
Package sysloglion defines functionality for integration with syslog.
*/
package sysloglion // import "go.pedge.io/lion/syslog"

import (
	"log/syslog"

	"go.pedge.io/lion"
)

var (
	// DefaultTextMarshaller is the default text Marshaller for syslog.
	DefaultTextMarshaller = lion.NewTextMarshaller(
		lion.TextMarshallerDisableTime(),
		lion.TextMarshallerDisableLevel(),
	)
)

// PusherOption is an option for constructing a new Pusher.
type PusherOption func(*pusher)

// PusherWithMarshaller uses the Marshaller for the Pusher.
//
// By default, DefaultTextMarshaller is used.
func PusherWithMarshaller(marshaller lion.Marshaller) PusherOption {
	return func(pusher *pusher) {
		pusher.marshaller = marshaller
	}
}

// NewPusher creates a new lion.Pusher that logs using syslog.
func NewPusher(writer *syslog.Writer, options ...PusherOption) lion.Pusher {
	return newPusher(writer, options...)
}
