package log

import "github.com/go-kit/kit/log"

//
// Logger interface, supports log messages and "events"
// where an event is an object that should get received
// by the configured EventSinks
//
type Logger interface {
	Log(keyvals ...interface{}) error
	Event(keyvals ...interface{}) error
}

type logger struct {
	wrapped log.Logger
	sinks   []EventSink
}

//
// NewLogger sets up and returns a Logger
//
func NewLogger(wrapped log.Logger, sinks []EventSink) Logger {
	return &logger{wrapped, sinks}
}

func (l *logger) Log(keyvals ...interface{}) error {
	return l.wrapped.Log(keyvals...)
}

//
// Event iterates through the configured EventSinks and
// sends the event to each one
//
func (l *logger) Event(keyvals ...interface{}) error {
	var err error
	if l.sinks != nil {
		for _, sink := range l.sinks {
			if err = sink.Receive(keyvals...); err != nil {
				return err
			}
		}
	}
	return err
}
