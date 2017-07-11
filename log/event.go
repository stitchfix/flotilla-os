package log

import (
	"errors"
	"github.com/stitchfix/httpclient"
	"time"
)

type EventSink interface {
	Receive(keyvals ...interface{}) error
}

type HTTPEventSink struct {
	path   string
	method string
	client httpclient.Client
}

type HTTPEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Message   map[string]interface{} `json:"message"`
}

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
// Receives an event and posts it to external host and path
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
