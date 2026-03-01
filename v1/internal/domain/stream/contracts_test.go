package stream

import (
	"testing"
	"time"
)

func TestEventValidate(t *testing.T) {
	event := Event{
		EventID:    "evt-1",
		OccurredAt: time.Now().UTC(),
		Source:     SourceACP,
		EventType:  EventType("stream.agent.chunk"),
		CorrelationIDs: CorrelationIDs{
			RunID:         "run-1",
			TaskID:        "task-1",
			JobID:         "job-1",
			SessionID:     "session-1",
			CorrelationID: "corr-1",
		},
		Payload: map[string]any{"text": "hello"},
	}
	if err := event.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestEventValidateRequiresCorrelationID(t *testing.T) {
	event := Event{
		EventID:    "evt-1",
		OccurredAt: time.Now().UTC(),
		Source:     SourceACP,
		EventType:  EventType("stream.agent.chunk"),
		CorrelationIDs: CorrelationIDs{
			RunID:     "run-1",
			TaskID:    "task-1",
			JobID:     "job-1",
			SessionID: "session-1",
		},
		Payload: map[string]any{"text": "hello"},
	}
	if err := event.Validate(); err == nil {
		t.Fatalf("expected correlation_id error")
	}
}
