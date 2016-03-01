package sysloglion

import (
	"log/syslog"

	"go.pedge.io/lion"
)

var (
	levelToLogFunc = map[lion.Level]func(*syslog.Writer, string) error{
		lion.LevelDebug: (*syslog.Writer).Debug,
		lion.LevelInfo:  (*syslog.Writer).Info,
		lion.LevelWarn:  (*syslog.Writer).Warning,
		lion.LevelError: (*syslog.Writer).Err,
		lion.LevelFatal: (*syslog.Writer).Crit,
		lion.LevelPanic: (*syslog.Writer).Alert,
		lion.LevelNone:  (*syslog.Writer).Info,
	}
)

type pusher struct {
	writer     *syslog.Writer
	marshaller lion.Marshaller
}

func newPusher(writer *syslog.Writer, options ...PusherOption) *pusher {
	pusher := &pusher{writer, DefaultTextMarshaller}
	for _, option := range options {
		option(pusher)
	}
	return pusher
}

func (p *pusher) Flush() error {
	return nil
}

func (p *pusher) Push(entry *lion.Entry) error {
	data, err := p.marshaller.Marshal(entry)
	if err != nil {
		return err
	}
	return levelToLogFunc[entry.Level](p.writer, string(data))
}
