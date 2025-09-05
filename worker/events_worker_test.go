package worker

import (
	"fmt"
	"testing"
	"time"

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
