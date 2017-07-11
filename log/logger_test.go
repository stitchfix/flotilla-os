package log

import (
	"testing"
)

type testLogger struct {
	keyvals []interface{}
}

func (tl *testLogger) Log(keyvals ...interface{}) error {
	tl.keyvals = keyvals
	return nil
}

type testSink struct {
	keyvals []interface{}
}

func (ts *testSink) Receive(keyvals ...interface{}) error {
	ts.keyvals = keyvals
	return nil
}

func TestLogger_Log(t *testing.T) {
	tl := &testLogger{}
	l := NewLogger(tl, nil)

	// Verify that the wrapped logger's Log method gets called
	l.Log("message", "value")
	if len(tl.keyvals) != 2 {
		t.Errorf("Expected log message with 2 values, got %v", len(tl.keyvals))
	}

	m1 := tl.keyvals[0]
	m2 := tl.keyvals[1]
	if m1.(string) != "message" || m2.(string) != "value" {
		t.Errorf("Expected [message, value] but got %s", tl.keyvals)
	}
}

func TestLogger_Event(t *testing.T) {
	ts := &testSink{}
	tl := &testLogger{}
	l := NewLogger(tl, []EventSink{ts})

	// Verify that the wrapped logger's Log method gets called
	l.Event("important_event", "act_on_me")
	if len(ts.keyvals) != 2 {
		t.Errorf("Expected to recieve event with 2 values, got %v", len(ts.keyvals))
	}

	m1 := ts.keyvals[0]
	m2 := ts.keyvals[1]
	if m1.(string) != "important_event" || m2.(string) != "act_on_me" {
		t.Errorf("Expected [important_event, act_on_me] but got %s", ts.keyvals)
	}
}
