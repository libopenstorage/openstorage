/*
Package envlion provides simple utilities to setup lion from the environment.
*/
package envlion // import "go.pedge.io/lion/env"

import (
	"log/syslog"
	"os"
	"strings"

	"go.pedge.io/env"
	"go.pedge.io/lion"
	"go.pedge.io/lion/current"
	"go.pedge.io/lion/grpc"
	"go.pedge.io/lion/syslog"
)

// Env defines a struct for environment variables that can be parsed with go.pedge.io/env.
type Env struct {
	// The log app name, will default to app if not set.
	LogAppName string `env:"LOG_APP_NAME,default=app"`
	// The level to log at, must be one of DEBUG, INFO, WARN, ERROR, FATAL, PANIC.
	LogLevel string `env:"LOG_LEVEL"`
	// LogDisableStderr says to disable logging to stderr.
	LogDisableStderr bool `env:"LOG_DISABLE_STDERR"`
	// The syslog network, either udp or tcp.
	// Must be set with SyslogAddress.
	// If not set and LogDir not set, logs will be to stderr.
	SyslogNetwork string `env:"SYSLOG_NETWORK"`
	// The syslog host:port.
	// Must be set with SyslogNetwork.
	// If not set and LogDir not set, logs will be to stderr.
	SyslogAddress string `env:"SYSLOG_ADDRESS"`
	// The current token.
	// Must be set with CurrentSyslogNetwork or CurrentStderr.
	CurrentToken string `env:"CURRENT_TOKEN"`
	// The current syslog host:port.
	// Must be set with CurrentToken.
	CurrentSyslogAddress string `env:"CURRENT_SYSLOG_ADDRESS"`
	// Output logs in current format to stdout.
	CurrentStdout bool `env:"CURRENT_STDOUT"`
}

// Setup gets the Env from the environment, and then calls SetupEnv.
func Setup() error {
	appEnv := Env{}
	if err := env.Populate(&appEnv); err != nil {
		return err
	}
	return SetupEnv(appEnv)
}

// SetupEnv sets up logging for the given Env.
func SetupEnv(env Env) error {
	var pushers []lion.Pusher
	logAppName := env.LogAppName
	if logAppName == "" {
		logAppName = "app"
	}
	if !env.LogDisableStderr {
		pushers = append(pushers, lion.NewTextWritePusher(os.Stderr))
	}
	if env.CurrentStdout {
		pushers = append(pushers, lion.NewWritePusher(os.Stdout, currentlion.NewMarshaller()))
	}
	if env.SyslogNetwork != "" && env.SyslogAddress != "" {
		pusher, err := newSyslogPusher(env.SyslogNetwork, env.SyslogAddress, logAppName)
		if err != nil {
			return err
		}
		pushers = append(pushers, pusher)
	}
	if env.CurrentToken != "" && env.CurrentSyslogAddress != "" {
		pusher, err := currentlion.NewPusher(logAppName, env.CurrentSyslogAddress, env.CurrentToken)
		if err != nil {
			return err
		}
		pushers = append(pushers, pusher)
	}
	switch len(pushers) {
	case 0:
		lion.SetLogger(lion.DiscardLogger)
	case 1:
		lion.SetLogger(lion.NewLogger(pushers[0]))
	default:
		lion.SetLogger(lion.NewLogger(lion.NewMultiPusher(pushers...)))
	}
	lion.RedirectStdLogger()
	grpclion.RedirectGrpclog()
	if env.LogLevel != "" {
		level, err := lion.NameToLevel(strings.ToUpper(env.LogLevel))
		if err != nil {
			return err
		}
		lion.SetLevel(level)
	}
	return nil
}

// Main runs env.Main along with Setup.
func Main(do func(interface{}) error, appEnv interface{}, decoders ...env.Decoder) {
	env.Main(
		func(appEnvObj interface{}) error {
			if err := Setup(); err != nil {
				return err
			}
			return do(appEnvObj)
		},
		appEnv,
		decoders...,
	)
}

func newSyslogPusher(syslogNetwork string, syslogAddress string, logAppName string) (lion.Pusher, error) {
	writer, err := syslog.Dial(
		syslogNetwork,
		syslogAddress,
		syslog.LOG_INFO,
		logAppName,
	)
	if err != nil {
		return nil, err
	}
	return sysloglion.NewPusher(writer), nil
}
