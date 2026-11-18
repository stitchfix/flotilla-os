package worker

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
)

func TestEventsWorker_applySlidingWindow(t *testing.T) {
	ew := &eventsWorker{}

	now := time.Now()

	// Create test events with different timestamps
	event1 := state.PodEvent{
		Message:   "Event 1",
		Timestamp: &now,
	}

	event2Time := now.Add(1 * time.Minute)
	event2 := state.PodEvent{
		Message:   "Event 2",
		Timestamp: &event2Time,
	}

	event3Time := now.Add(2 * time.Minute)
	event3 := state.PodEvent{
		Message:   "Event 3",
		Timestamp: &event3Time,
	}

	event4Time := now.Add(3 * time.Minute)
	event4 := state.PodEvent{
		Message:   "Event 4",
		Timestamp: &event4Time,
	}

	// Test case 1: Under the limit
	t.Run("UnderLimit", func(t *testing.T) {
		var events state.PodEvents
		result := ew.applySlidingWindow(events, event1, 3)

		if len(result) != 1 {
			t.Errorf("Expected 1 event, got %d", len(result))
		}
		if result[0].Message != "Event 1" {
			t.Errorf("Expected 'Event 1', got %s", result[0].Message)
		}
	})

	// Test case 2: At the limit
	t.Run("AtLimit", func(t *testing.T) {
		events := state.PodEvents{event1, event2}
		result := ew.applySlidingWindow(events, event3, 3)

		if len(result) != 3 {
			t.Errorf("Expected 3 events, got %d", len(result))
		}
	})

	// Test case 3: Over the limit - should keep only the most recent
	t.Run("OverLimit", func(t *testing.T) {
		events := state.PodEvents{event1, event2, event3}
		result := ew.applySlidingWindow(events, event4, 3)

		if len(result) != 3 {
			t.Errorf("Expected 3 events, got %d", len(result))
		}

		// Should keep the 3 most recent: event4, event3, event2 (newest first)
		if result[0].Message != "Event 4" {
			t.Errorf("Expected newest event 'Event 4' first, got %s", result[0].Message)
		}
		if result[1].Message != "Event 3" {
			t.Errorf("Expected second newest 'Event 3', got %s", result[1].Message)
		}
		if result[2].Message != "Event 2" {
			t.Errorf("Expected third newest 'Event 2', got %s", result[2].Message)
		}
	})

	// Test case 4: EKS default limit (20) - realistic scenario
	t.Run("EKSDefaultLimit", func(t *testing.T) {
		var events state.PodEvents
		// Add 21 events to test the sliding window at default EKS limit
		for i := 1; i <= 21; i++ {
			eventTime := now.Add(time.Duration(i) * time.Minute)
			newEvent := state.PodEvent{
				Message:   fmt.Sprintf("Event %d", i),
				Timestamp: &eventTime,
			}
			events = ew.applySlidingWindow(events, newEvent, 20)
		}

		if len(events) != 20 {
			t.Errorf("Expected 20 events for EKS limit, got %d", len(events))
		}

		// Should have events 21, 20, 19, ... 2 (newest first)
		if events[0].Message != "Event 21" {
			t.Errorf("Expected newest event 'Event 21' first, got %s", events[0].Message)
		}
		if events[19].Message != "Event 2" {
			t.Errorf("Expected oldest kept event 'Event 2', got %s", events[19].Message)
		}
	})
}

type nopLogger struct{}

func (nopLogger) Log(keyvals ...interface{}) error   { return nil }
func (nopLogger) Event(keyvals ...interface{}) error { return nil }

type fakeEventsQueueManager struct {
	queue.Manager
	receiveKubernetes func() ([]state.KubernetesEvent, error)
	receiveEMR        func() ([]state.EmrEvent, error)
}

func (f *fakeEventsQueueManager) QurlFor(name string, prefixed bool) (string, error) {
	return name, nil
}

func (f *fakeEventsQueueManager) Initialize(conf config.Config, engine string) error {
	return nil
}

func (f *fakeEventsQueueManager) ReceiveKubernetesEvents(qURL string) ([]state.KubernetesEvent, error) {
	return f.receiveKubernetes()
}

func (f *fakeEventsQueueManager) ReceiveEMREvents(qURL string) ([]state.EmrEvent, error) {
	return f.receiveEMR()
}

func ackablePodEvent(acked *int64) state.KubernetesEvent {
	return state.KubernetesEvent{
		InvolvedObject: state.InvolvedObject{Kind: "Pod"},
		Done: func() error {
			atomic.AddInt64(acked, 1)
			return nil
		},
	}
}

func runWorker(ew *eventsWorker) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		_ = ew.Run(context.Background())
		close(done)
	}()
	return done
}

func TestEventsWorker_SlowEMRReceiveDoesNotStallKubernetesEvents(t *testing.T) {
	var acked int64
	release := make(chan struct{})
	var calls int64
	qm := &fakeEventsQueueManager{
		receiveKubernetes: func() ([]state.KubernetesEvent, error) {
			if atomic.AddInt64(&calls, 1) > 5 {
				return nil, nil
			}
			return []state.KubernetesEvent{ackablePodEvent(&acked), ackablePodEvent(&acked)}, nil
		},
		receiveEMR: func() ([]state.EmrEvent, error) {
			<-release
			return nil, nil
		},
	}
	ew := &eventsWorker{qm: qm, log: nopLogger{}, pollInterval: time.Millisecond, maxBackoff: 5 * time.Millisecond}
	done := runWorker(ew)

	deadline := time.After(2 * time.Second)
	for atomic.LoadInt64(&acked) < 10 {
		select {
		case <-deadline:
			t.Fatalf("Kubernetes events stalled behind EMR receive: acked %d of 10", atomic.LoadInt64(&acked))
		case <-time.After(time.Millisecond):
		}
	}

	ew.t.Kill(nil)
	close(release)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("Run did not exit after tomb was killed")
	}
}

func TestEventsWorker_ProcessesEveryEventInBatch(t *testing.T) {
	var acked int64
	var emrCalls int64
	var k8sCalls int64
	qm := &fakeEventsQueueManager{
		receiveKubernetes: func() ([]state.KubernetesEvent, error) {
			if atomic.AddInt64(&k8sCalls, 1) == 1 {
				batch := make([]state.KubernetesEvent, 10)
				for i := range batch {
					batch[i] = ackablePodEvent(&acked)
				}
				return batch, nil
			}
			return nil, nil
		},
		receiveEMR: func() ([]state.EmrEvent, error) {
			atomic.AddInt64(&emrCalls, 1)
			return []state.EmrEvent{{}, {}}, nil
		},
	}
	ew := &eventsWorker{qm: qm, log: nopLogger{}, pollInterval: time.Millisecond, maxBackoff: 5 * time.Millisecond}
	done := runWorker(ew)

	deadline := time.After(2 * time.Second)
	for atomic.LoadInt64(&acked) < 10 || atomic.LoadInt64(&emrCalls) < 2 {
		select {
		case <-deadline:
			t.Fatalf("Expected 10 acked kubernetes events and repeated EMR receives, got %d and %d",
				atomic.LoadInt64(&acked), atomic.LoadInt64(&emrCalls))
		case <-time.After(time.Millisecond):
		}
	}
	ew.t.Kill(nil)
	<-done
	if atomic.LoadInt64(&acked) != 10 {
		t.Errorf("Expected exactly 10 acked events, got %d", atomic.LoadInt64(&acked))
	}
}

func TestEventsWorker_RunExitsWhenTombKilled(t *testing.T) {
	qm := &fakeEventsQueueManager{
		receiveKubernetes: func() ([]state.KubernetesEvent, error) { return nil, nil },
		receiveEMR:        func() ([]state.EmrEvent, error) { return nil, nil },
	}
	ew := &eventsWorker{qm: qm, log: nopLogger{}, pollInterval: 50 * time.Millisecond, maxBackoff: time.Hour}
	done := runWorker(ew)

	time.Sleep(20 * time.Millisecond)
	ew.t.Kill(nil)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("Run did not exit after tomb was killed")
	}
}

func TestEventsWorker_PollLoopBackoff(t *testing.T) {
	ew := &eventsWorker{log: nopLogger{}, pollInterval: 100 * time.Millisecond, maxBackoff: 500 * time.Millisecond}

	results := []struct {
		n   int
		err error
	}{
		{0, nil}, {0, nil}, {0, nil}, {0, nil}, {0, nil},
		{3, nil},
		{0, nil}, {0, nil},
		{0, errors.New("boom")},
		{0, nil},
	}
	call := 0
	receive := func(ctx context.Context) (int, error) {
		r := results[call]
		call++
		return r.n, r.err
	}
	var slept []time.Duration
	sleep := func(d time.Duration) bool {
		slept = append(slept, d)
		return len(slept) < 9
	}

	ew.pollLoop(context.Background(), receive, sleep)

	ms := time.Millisecond
	expected := []time.Duration{100 * ms, 200 * ms, 400 * ms, 500 * ms, 500 * ms, 100 * ms, 200 * ms, 400 * ms, 400 * ms}
	if call != 10 {
		t.Errorf("Expected 10 receives, got %d", call)
	}
	if len(slept) != len(expected) {
		t.Fatalf("Expected sleeps %v, got %v", expected, slept)
	}
	for i := range expected {
		if slept[i] != expected[i] {
			t.Errorf("Expected sleeps %v, got %v", expected, slept)
			break
		}
	}
}

func TestEventsWorker_InitializeMaxBackoff(t *testing.T) {
	confDir := "../conf"
	c, _ := config.NewConfig(&confDir)
	ew := &eventsWorker{}
	_ = ew.Initialize(c, nil, nil, nil, nopLogger{}, 100*time.Millisecond, &fakeEventsQueueManager{}, nil)
	if ew.maxBackoff != 5*time.Second {
		t.Errorf("Expected default max backoff 5s, got %v", ew.maxBackoff)
	}
}
