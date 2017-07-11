package log

import "github.com/go-kit/kit/log"

type Logger interface {
	Log(keyvals ...interface{}) error
	Event(keyvals ...interface{}) error
}

type logger struct {
	wrapped log.Logger
	sinks   []EventSink
}

func NewLogger(wrapped log.Logger, sinks []EventSink) Logger {
	return &logger{wrapped, sinks}
}

func (l *logger) Log(keyvals ...interface{}) error {
	return l.wrapped.Log(keyvals...)
}

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
