package api

import (
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
	domainstream "agentic-orchestrator/internal/domain/stream"
	"testing"
	"time"
)

func TestLifecycleStreamEventFromTableChangeMapsLifecyclePayload(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	event, ok := lifecycleStreamEventFromTableChange(domainrealtime.TableChangeEvent{
		ProjectID:       "project-1",
		RunID:           "run-1",
		TaskID:          "task-1",
		JobID:           "job-1",
		SessionID:       "session-1",
		ProjectEventSeq: 42,
		SessionEventSeq: 7,
		OccurredAt:      now,
		Payload: map[string]any{
			"event_id":       "lifecycle-evt-1",
			"event_type":     "started",
			"source_runtime": "worker",
			"occurred_at":    now.Format(time.RFC3339Nano),
			"payload": map[string]any{
				"runtime_alive": true,
			},
		},
	})
	if !ok {
		t.Fatalf("expected lifecycle table-change conversion to succeed")
	}
	if event.EventID != "lifecycle-evt-1" {
		t.Fatalf("expected event_id lifecycle-evt-1, got %q", event.EventID)
	}
	if event.EventType != domainstream.EventSessionStarted {
		t.Fatalf("expected stream.session.started, got %q", event.EventType)
	}
	if event.Source != domainstream.SourceWorker {
		t.Fatalf("expected worker source, got %q", event.Source)
	}
	if event.CorrelationIDs.ProjectID != "project-1" {
		t.Fatalf("expected project correlation id project-1, got %q", event.CorrelationIDs.ProjectID)
	}
	if event.CorrelationIDs.CorrelationID != "session:session-1" {
		t.Fatalf("expected correlation id session:session-1, got %q", event.CorrelationIDs.CorrelationID)
	}
	if event.StreamOffset != 42 {
		t.Fatalf("expected stream offset 42, got %d", event.StreamOffset)
	}
}

func TestLifecycleStreamEventFromTableChangeRejectsMissingPayload(t *testing.T) {
	_, ok := lifecycleStreamEventFromTableChange(domainrealtime.TableChangeEvent{})
	if ok {
		t.Fatalf("expected conversion to fail without payload")
	}
}

func TestLifecycleStreamEventFromTableChangeMapsHeartbeatToSessionHealth(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	event, ok := lifecycleStreamEventFromTableChange(domainrealtime.TableChangeEvent{
		ProjectID:       "project-1",
		SessionID:       "session-1",
		ProjectEventSeq: 9,
		SessionEventSeq: 3,
		OccurredAt:      now,
		Payload: map[string]any{
			"event_id":       "lifecycle-heartbeat-1",
			"event_type":     "heartbeat",
			"source_runtime": "worker",
			"occurred_at":    now.Format(time.RFC3339Nano),
			"payload":        map[string]any{"runtime_alive": true},
		},
	})
	if !ok {
		t.Fatalf("expected lifecycle heartbeat conversion to succeed")
	}
	if event.EventType != domainstream.EventSessionHealth {
		t.Fatalf("expected stream.session.health, got %q", event.EventType)
	}
}

func TestLifecycleStreamEventFromTableChangeFlagsFailedAsErrorEvent(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	event, ok := lifecycleStreamEventFromTableChange(domainrealtime.TableChangeEvent{
		ProjectID:       "project-1",
		SessionID:       "session-1",
		ProjectEventSeq: 10,
		SessionEventSeq: 4,
		OccurredAt:      now,
		Payload: map[string]any{
			"event_id":       "lifecycle-failed-1",
			"event_type":     "failed",
			"source_runtime": "worker",
			"occurred_at":    now.Format(time.RFC3339Nano),
			"payload": map[string]any{
				"error_code": "E_ASYNQ_TIMEOUT",
			},
		},
	})
	if !ok {
		t.Fatalf("expected lifecycle failed conversion to succeed")
	}
	flagged, _ := event.Payload["error_event"].(bool)
	if !flagged {
		t.Fatalf("expected error_event=true for failed lifecycle event")
	}
}

func TestStreamEventFromTableChangePayloadFlagsEndedWithErrorMarker(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	event, ok := streamEventFromTableChangePayload(domainrealtime.TableChangeEvent{
		OccurredAt: now,
		Payload: map[string]any{
			"event_id":      "stream-ended-1",
			"event_type":    string(domainstream.EventSessionEnded),
			"source":        string(domainstream.SourceWorker),
			"occurred_at":   now.Format(time.RFC3339Nano),
			"stream_offset": 123,
			"correlation_id": "session:session-1",
			"payload": map[string]any{
				"error": "asynq: retry exhausted",
			},
		},
	})
	if !ok {
		t.Fatalf("expected stream table-change conversion to succeed")
	}
	flagged, _ := event.Payload["error_event"].(bool)
	if !flagged {
		t.Fatalf("expected error_event=true for stream.session.ended with error marker")
	}
}
