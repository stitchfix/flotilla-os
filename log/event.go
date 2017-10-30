package log

import (
	"os"
	"log"
	"errors"
	"github.com/stitchfix/flotilla-os/clients/httpclient"
	"time"
)

//
// EventSink interface
//
type EventSink interface {
	Receive(keyvals ...interface{}) error
}

//
// LocalEventSink - an implementation of EventSink that 
// simply logs events to os.Stderr.
//
type LocalEventSink struct {
	logger *log.Logger 
}

func NewLocalEventSink() *LocalEventSink {
	logger := log.New(os.Stderr, "[LocalEventSink] ", 
					log.Ldate | log.Ltime | log.Lshortfile)

	return &LocalEventSink{logger}
}

func (localSink *LocalEventSink) Receive(keyvals ...interface{}) error {
	log.Printf("\n%v\n", keyvals)
	return nil
}

//
// HTTPEventSink pushes arbitrary key-value
// events to an external location
//
type HTTPEventSink struct {
	path   string
	method string
	client httpclient.Client
}

//
// HTTPEvent represents an arbitrary key-value
// event
//
type HTTPEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Message   map[string]interface{} `json:"message"`
}

//
// NewHTTPSink initializes and returns an HTTPEventSink
//
func NewHTTPSink(host string, path string, method string) HTTPEventSink {
	return HTTPEventSink{

		path, method, httpclient.Client{Host: host},
	}
}

func (httpsink *HTTPEventSink) headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

func (httpsink *HTTPEventSink) constructMessage(keyvals ...interface{}) (map[string]interface{}, error) {
	n := (len(keyvals) + 1) / 2
	m := make(map[string]interface{}, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		key, ok := k.(string)
		if !ok {
			return m, errors.New("Not all keys are strings")
		}
		var v interface{}
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}
		m[key] = v
	}
	return m, nil
}

//
// Receive consumes an arbitrary set of keys and values (k1,v1,k2,v2,...),
// constructs an HTTPEvent from them, and sends them to the configured
// http endpoint using the configured method
//
func (httpsink *HTTPEventSink) Receive(keyvals ...interface{}) error {
	var err error
	var event HTTPEvent

	m, err := httpsink.constructMessage(keyvals...)
	if err != nil {
		return err
	}
	event.Message = m
	event.Timestamp = time.Now().UTC()

	var response interface{}

	return httpsink.client.Post(
		httpsink.method,
		httpsink.headers(),
		&event, &response)
}
