package realtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"agentic-orchestrator/internal/domain/failures"
)

type TableChangeEvent struct {
	Topic           string         `json:"topic"`
	Table           string         `json:"table"`
	Operation       string         `json:"operation"`
	ProjectID       string         `json:"project_id"`
	RunID           string         `json:"run_id,omitempty"`
	TaskID          string         `json:"task_id,omitempty"`
	JobID           string         `json:"job_id,omitempty"`
	SessionID       string         `json:"session_id,omitempty"`
	ProjectEventSeq int64          `json:"project_event_seq"`
	SessionEventSeq int64          `json:"session_event_seq"`
	OccurredAt      time.Time      `json:"occurred_at"`
	Payload         map[string]any `json:"payload,omitempty"`
}

func (event TableChangeEvent) Validate() error {
	if strings.TrimSpace(event.Topic) == "" {
		return failures.WrapTerminal(errors.New("topic is required"))
	}
	if strings.TrimSpace(event.Table) == "" {
		return failures.WrapTerminal(errors.New("table is required"))
	}
	if strings.TrimSpace(event.Operation) == "" {
		return failures.WrapTerminal(errors.New("operation is required"))
	}
	if strings.TrimSpace(event.ProjectID) == "" {
		return failures.WrapTerminal(errors.New("project_id is required"))
	}
	if event.ProjectEventSeq < 0 {
		return failures.WrapTerminal(fmt.Errorf("project_event_seq must be >= 0, got %d", event.ProjectEventSeq))
	}
	if event.SessionEventSeq < 0 {
		return failures.WrapTerminal(fmt.Errorf("session_event_seq must be >= 0, got %d", event.SessionEventSeq))
	}
	if event.OccurredAt.IsZero() {
		return failures.WrapTerminal(errors.New("occurred_at is required"))
	}
	return nil
}

type TableChangeWatcher interface {
	Publish(ctx context.Context, event TableChangeEvent) error
	Watch(ctx context.Context, topic string, handler func(TableChangeEvent) error) error
}
