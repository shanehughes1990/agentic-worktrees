package lifecycle

import (
	"agentic-orchestrator/internal/domain/failures"
	"errors"
	"strings"
	"time"
)

type EventType string

const (
	EventEnqueued      EventType = "enqueued"
	EventStarted       EventType = "started"
	EventCompleted     EventType = "completed"
	EventFailed        EventType = "failed"
	EventGapDetected   EventType = "gap_detected"
	EventGapReconciled EventType = "gap_reconciled"
)

type Event struct {
	EventID         string
	SchemaVersion   int
	ProjectID       string
	RunID           string
	TaskID          string
	JobID           string
	SessionID       string
	WorkerID        string
	SourceRuntime   string
	PipelineType    string
	EventType       EventType
	OccurredAt      time.Time
	Payload         map[string]any
	EventSeq        int64
	ProjectEventSeq int64
}

func (event Event) ValidateForAppend() error {
	if strings.TrimSpace(event.EventID) == "" {
		return failures.WrapTerminal(errors.New("event_id is required"))
	}
	if event.SchemaVersion <= 0 {
		return failures.WrapTerminal(errors.New("schema_version is required"))
	}
	if strings.TrimSpace(event.ProjectID) == "" {
		return failures.WrapTerminal(errors.New("project_id is required"))
	}
	if strings.TrimSpace(event.SessionID) == "" {
		return failures.WrapTerminal(errors.New("session_id is required"))
	}
	if strings.TrimSpace(event.SourceRuntime) == "" {
		return failures.WrapTerminal(errors.New("source_runtime is required"))
	}
	if strings.TrimSpace(event.PipelineType) == "" {
		return failures.WrapTerminal(errors.New("pipeline_type is required"))
	}
	if strings.TrimSpace(string(event.EventType)) == "" {
		return failures.WrapTerminal(errors.New("event_type is required"))
	}
	if event.OccurredAt.IsZero() {
		return failures.WrapTerminal(errors.New("occurred_at is required"))
	}
	if event.Payload == nil {
		return failures.WrapTerminal(errors.New("payload is required"))
	}
	return nil
}
