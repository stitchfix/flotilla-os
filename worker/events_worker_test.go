package worker

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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

func waitFor(t *testing.T, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal(msg)
		}
		time.Sleep(time.Millisecond)
	}
}

func stopWorker(t *testing.T, ew *eventsWorker, done <-chan struct{}) {
	t.Helper()
	ew.t.Kill(nil)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after tomb was killed")
	}
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

	waitFor(t, func() bool { return atomic.LoadInt64(&acked) >= 10 }, "Kubernetes events stalled behind EMR receive")

	close(release)
	stopWorker(t, ew, done)
}

func TestEventsWorker_ProcessesEveryEventInBatch(t *testing.T) {
	var acked int64
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
		receiveEMR: func() ([]state.EmrEvent, error) { return nil, nil },
	}
	ew := &eventsWorker{qm: qm, log: nopLogger{}, pollInterval: time.Millisecond, maxBackoff: 5 * time.Millisecond}
	done := runWorker(ew)

	waitFor(t, func() bool { return atomic.LoadInt64(&acked) >= 10 }, "Expected all 10 events in the batch to be processed")
	stopWorker(t, ew, done)
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
	stopWorker(t, ew, done)
}

func TestEventsWorker_NextBackoff(t *testing.T) {
	ms := time.Millisecond
	ew := &eventsWorker{pollInterval: 100 * ms, maxBackoff: 500 * ms}
	cases := []struct{ current, expected time.Duration }{
		{100 * ms, 200 * ms},
		{200 * ms, 400 * ms},
		{400 * ms, 500 * ms},
		{500 * ms, 500 * ms},
	}
	for _, c := range cases {
		if got := ew.nextBackoff(c.current); got != c.expected {
			t.Errorf("nextBackoff(%v): expected %v, got %v", c.current, c.expected, got)
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

type fakeEventsStateManager struct {
	state.Manager
	snapshot state.Run
	stored   *state.Run
	updates  []state.Run
}

func (f *fakeEventsStateManager) GetRunByEMRJobId(ctx context.Context, emrJobId string) (state.Run, error) {
	return f.snapshot, nil
}

func (f *fakeEventsStateManager) UpdateRun(ctx context.Context, runID string, updates state.Run) (state.Run, error) {
	f.updates = append(f.updates, updates)
	if f.stored == nil {
		return updates, nil
	}
	f.stored.UpdateWith(updates)
	return *f.stored, nil
}

func emrStatusEvent(status string, acked *int64) state.EmrEvent {
	return state.EmrEvent{
		Time: aws.String("2026-11-18T10:00:00Z"),
		Detail: &state.Detail{
			ID:           aws.String("emr-job-1"),
			State:        aws.String(status),
			StateDetails: aws.String("details"),
		},
		Done: func() error {
			atomic.AddInt64(acked, 1)
			return nil
		},
	}
}

func TestEventsWorker_ProcessEventEMRCompletedSendsOnlyStatusFields(t *testing.T) {
	queued := time.Now().Add(-time.Hour)
	cases := []struct {
		name          string
		sparkAppId    *string
		expectMetrics bool
	}{
		{"with spark app id", aws.String("spark-123"), true},
		{"without spark app id", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			podEvents := state.PodEvents{{SourceObject: "exec-1"}}
			sm := &fakeEventsStateManager{snapshot: state.Run{
				RunID:          "run-1",
				Status:         state.StatusRunning,
				QueuedAt:       &queued,
				PodEvents:      &podEvents,
				SparkExtension: &state.SparkExtension{SparkAppId: c.sparkAppId},
			}}
			var acked int64
			ew := &eventsWorker{sm: sm, log: nopLogger{}, emrMetricsServer: "https://metrics"}
			ew.processEventEMR(context.Background(), emrStatusEvent("COMPLETED", &acked))

			if len(sm.updates) != 1 {
				t.Fatalf("Expected 1 UpdateRun call, got %d", len(sm.updates))
			}
			u := sm.updates[0]
			if u.SparkExtension != nil {
				t.Errorf("Expected no SparkExtension in update, got %+v", u.SparkExtension)
			}
			if u.PodEvents != nil {
				t.Errorf("Expected no PodEvents in update, got %+v", u.PodEvents)
			}
			if u.Status != state.StatusStopped {
				t.Errorf("Expected status %s, got %s", state.StatusStopped, u.Status)
			}
			if u.ExitCode == nil || *u.ExitCode != 0 {
				t.Errorf("Expected exit code 0, got %v", u.ExitCode)
			}
			if u.FinishedAt == nil {
				t.Errorf("Expected FinishedAt to be set")
			}
			if u.StartedAt == nil || !u.StartedAt.Equal(queued) {
				t.Errorf("Expected StartedAt to fall back to QueuedAt %v, got %v", queued, u.StartedAt)
			}
			if c.expectMetrics {
				if u.MetricsUri == nil || *u.MetricsUri != "https://metrics?query=%40app_id%3Aspark-123" {
					t.Errorf("Unexpected MetricsUri %v", u.MetricsUri)
				}
			} else if u.MetricsUri != nil {
				t.Errorf("Expected no MetricsUri, got %s", *u.MetricsUri)
			}
			if atomic.LoadInt64(&acked) != 1 {
				t.Errorf("Expected event to be acked")
			}
		})
	}
}

func TestEventsWorker_ProcessEventEMRFailedDriverOOM(t *testing.T) {
	queued := time.Now().Add(-time.Hour)
	sm := &fakeEventsStateManager{snapshot: state.Run{
		RunID:          "run-1",
		Status:         state.StatusRunning,
		QueuedAt:       &queued,
		SparkExtension: &state.SparkExtension{DriverOOM: aws.Bool(true)},
	}}
	var acked int64
	ew := &eventsWorker{sm: sm, log: nopLogger{}}
	ew.processEventEMR(context.Background(), emrStatusEvent("FAILED", &acked))

	if len(sm.updates) != 1 {
		t.Fatalf("Expected 1 UpdateRun call, got %d", len(sm.updates))
	}
	u := sm.updates[0]
	if u.ExitCode == nil || *u.ExitCode != 137 {
		t.Errorf("Expected exit code 137, got %v", u.ExitCode)
	}
	if u.ExitReason == nil || *u.ExitReason != "Driver OOMKilled, retry with more driver memory." {
		t.Errorf("Unexpected exit reason %v", u.ExitReason)
	}
	if u.SparkExtension != nil {
		t.Errorf("Expected no SparkExtension in update, got %+v", u.SparkExtension)
	}
	if u.Status != state.StatusStopped {
		t.Errorf("Expected status %s, got %s", state.StatusStopped, u.Status)
	}
}

func TestEventsWorker_ProcessEventEMRDoesNotEraseConcurrentPodEventWrites(t *testing.T) {
	queued := time.Now().Add(-time.Hour)
	podEvents := state.PodEvents{{SourceObject: "exec-1"}}
	stored := state.Run{
		RunID:     "run-1",
		Status:    state.StatusRunning,
		QueuedAt:  &queued,
		PodEvents: &podEvents,
		SparkExtension: &state.SparkExtension{
			SparkAppId: aws.String("spark-123"),
			HistoryUri: aws.String("https://history/spark-123/jobs/"),
			DriverOOM:  aws.Bool(true),
		},
	}
	sm := &fakeEventsStateManager{
		stored: &stored,
		snapshot: state.Run{
			RunID:          "run-1",
			Status:         state.StatusRunning,
			QueuedAt:       &queued,
			SparkExtension: &state.SparkExtension{},
		},
	}
	var acked int64
	ew := &eventsWorker{sm: sm, log: nopLogger{}}
	ew.processEventEMR(context.Background(), emrStatusEvent("COMPLETED", &acked))

	if stored.SparkExtension == nil || stored.SparkExtension.SparkAppId == nil || *stored.SparkExtension.SparkAppId != "spark-123" {
		t.Errorf("Expected stored SparkAppId to survive the EMR status update, got %+v", stored.SparkExtension)
	}
	if stored.SparkExtension == nil || stored.SparkExtension.DriverOOM == nil || !*stored.SparkExtension.DriverOOM {
		t.Errorf("Expected stored DriverOOM to survive the EMR status update")
	}
	if stored.PodEvents == nil || len(*stored.PodEvents) != 1 {
		t.Errorf("Expected stored PodEvents to survive the EMR status update, got %v", stored.PodEvents)
	}
	if stored.Status != state.StatusStopped {
		t.Errorf("Expected stored status %s, got %s", state.StatusStopped, stored.Status)
	}
}

func TestEventsWorker_EMRPodEventUpdatesSendsOnlyPodFields(t *testing.T) {
	queued := time.Now().Add(-time.Hour)
	finished := time.Now()
	snapshot := state.Run{
		RunID:          "run-1",
		ClusterName:    "cluster-a",
		Status:         state.StatusStopped,
		QueuedAt:       &queued,
		StartedAt:      &queued,
		FinishedAt:     &finished,
		ExitCode:       aws.Int64(0),
		ExitReason:     aws.String("done"),
		SparkExtension: &state.SparkExtension{},
	}
	ew := &eventsWorker{
		log:              nopLogger{},
		emrHistoryServer: "https://history",
		emrMetricsServer: "https://metrics",
		emrAppServer:     map[string]string{"cluster-a": "https://app"},
	}
	event := state.KubernetesEvent{
		FirstTimestamp: "2026-11-18T10:00:00Z",
		Type:           "Normal",
		Reason:         "Scheduled",
		Message:        "scheduled",
		InvolvedObject: state.InvolvedObject{Kind: "Pod", Name: "job-exec-1"},
	}

	u := ew.emrPodEventUpdates(snapshot, event, aws.String("spark-123"), aws.String("driver-svc"), aws.Bool(true), aws.Bool(true))

	if u.Status != "" || u.ExitCode != nil || u.ExitReason != nil || u.StartedAt != nil || u.FinishedAt != nil || u.QueuedAt != nil {
		t.Errorf("Expected no status fields in pod event update, got %+v", u)
	}
	if u.PodEvents == nil || len(*u.PodEvents) != 1 || (*u.PodEvents)[0].SourceObject != "job-exec-1" {
		t.Errorf("Unexpected PodEvents %v", u.PodEvents)
	}
	se := u.SparkExtension
	if se == nil {
		t.Fatal("Expected SparkExtension in update")
	}
	if se.SparkAppId == nil || *se.SparkAppId != "spark-123" {
		t.Errorf("Unexpected SparkAppId %v", se.SparkAppId)
	}
	if se.HistoryUri == nil || *se.HistoryUri != "https://history/spark-123/jobs/" {
		t.Errorf("Unexpected HistoryUri %v", se.HistoryUri)
	}
	if se.AppUri == nil || *se.AppUri != "https://app/job/driver-svc" {
		t.Errorf("Unexpected AppUri %v", se.AppUri)
	}
	if se.DriverOOM == nil || !*se.DriverOOM || se.ExecutorOOM == nil || !*se.ExecutorOOM {
		t.Errorf("Expected OOM flags set, got driver=%v executor=%v", se.DriverOOM, se.ExecutorOOM)
	}
	if u.MetricsUri == nil || *u.MetricsUri != "https://metrics?query=%40app_id%3Aspark-123" {
		t.Errorf("Unexpected MetricsUri %v", u.MetricsUri)
	}
}
