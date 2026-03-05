package realtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"agentic-orchestrator/internal/domain/failures"
)

type RuntimeActivitySignal struct {
	SignalID     string         `json:"signal_id"`
	ProjectID    string         `json:"project_id"`
	RunID        string         `json:"run_id,omitempty"`
	TaskID       string         `json:"task_id,omitempty"`
	JobID        string         `json:"job_id,omitempty"`
	SessionID    string         `json:"session_id"`
	PipelineType string         `json:"pipeline_type"`
	WorkerID     string         `json:"worker_id,omitempty"`
	EventType    string         `json:"event_type"`
	OccurredAt   time.Time      `json:"occurred_at"`
	Payload      map[string]any `json:"payload,omitempty"`
}

func (signal RuntimeActivitySignal) Validate() error {
	if strings.TrimSpace(signal.SignalID) == "" {
		return failures.WrapTerminal(errors.New("signal_id is required"))
	}
	if strings.TrimSpace(signal.ProjectID) == "" {
		return failures.WrapTerminal(errors.New("project_id is required"))
	}
	if strings.TrimSpace(signal.SessionID) == "" {
		return failures.WrapTerminal(errors.New("session_id is required"))
	}
	if strings.TrimSpace(signal.PipelineType) == "" {
		return failures.WrapTerminal(errors.New("pipeline_type is required"))
	}
	if strings.TrimSpace(signal.EventType) == "" {
		return failures.WrapTerminal(errors.New("event_type is required"))
	}
	if signal.OccurredAt.IsZero() {
		return failures.WrapTerminal(errors.New("occurred_at is required"))
	}
	return nil
}

type RuntimeActivityTransport interface {
	PublishRuntimeActivity(ctx context.Context, signal RuntimeActivitySignal) error
	ListenRuntimeActivity(ctx context.Context, handler func(RuntimeActivitySignal) error) error
}

func RuntimeActivitySignalID(pipelineType string, sessionID string, eventType string, occurredAt time.Time) string {
	resolved := occurredAt
	if resolved.IsZero() {
		resolved = time.Now().UTC()
	}
	return fmt.Sprintf("runtime:%s:%s:%s:%d", strings.TrimSpace(pipelineType), strings.TrimSpace(sessionID), strings.TrimSpace(eventType), resolved.UnixNano())
}
