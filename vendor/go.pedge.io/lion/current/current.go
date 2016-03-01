/*
Package currentlion implements basic integration with Current using plaintext syslog.

https://current.sh
*/
package currentlion // import "go.pedge.io/lion/current"

import (
	"log/syslog"

	"go.pedge.io/lion"
	"go.pedge.io/lion/syslog"
)

const (
	syslogNetwork = "tcp"
)

// NewPusher returns a new Pusher for current.
func NewPusher(
	appName string,
	syslogAddress string,
	token string,
) (lion.Pusher, error) {
	writer, err := syslog.Dial(
		syslogNetwork,
		syslogAddress,
		syslog.LOG_INFO,
		appName,
	)
	if err != nil {
		return nil, err
	}
	return sysloglion.NewPusher(
		writer,
		sysloglion.PusherWithMarshaller(
			NewMarshaller(
				MarshallerWithToken(token),
				MarshallerDisableNewlines(),
			),
		),
	), nil
}

// MarshallerOption is an option for creating Marshallers.
type MarshallerOption func(*marshaller)

// MarshallerWithToken will add @current:token before every marshalled Entry.
// This is used for syslog output.
func MarshallerWithToken(token string) MarshallerOption {
	return func(marshaller *marshaller) {
		marshaller.token = token
	}
}

// MarshallerDisableNewlines disables newlines after each marshalled Entry.
func MarshallerDisableNewlines() MarshallerOption {
	return func(marshaller *marshaller) {
		marshaller.disableNewlines = true
	}
}

// NewMarshaller returns a new Marshaller that marshals messages into , appropriate
// to send to an io.Writer that can be tailed by the current cli tool.
func NewMarshaller(options ...MarshallerOption) lion.Marshaller {
	return newMarshaller(options...)
}
