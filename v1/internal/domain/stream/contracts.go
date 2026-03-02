package stream

import (
	"agentic-orchestrator/internal/domain/failures"
	"errors"
	"slices"
	"strings"
	"time"
)

type Source string

const (
	SourceACP         Source = "acp"
	SourceSessionFile Source = "session_file"
	SourceWorker      Source = "worker"
)

type EventType string

type CorrelationIDs struct {
	RunID         string
	TaskID        string
	JobID         string
	ProjectID     string
	SessionID     string
	CorrelationID string
}

func (ids CorrelationIDs) Validate() error {
	if strings.TrimSpace(ids.CorrelationID) == "" {
		return failures.WrapTerminal(errors.New("correlation_id is required"))
	}
	return nil
}

type Event struct {
	EventID       string
	StreamOffset  uint64
	OccurredAt    time.Time
	Source        Source
	EventType     EventType
	CorrelationIDs CorrelationIDs
	Payload       map[string]any
}

func (event Event) Validate() error {
	if strings.TrimSpace(event.EventID) == "" {
		return failures.WrapTerminal(errors.New("event_id is required"))
	}
	if event.OccurredAt.IsZero() {
		return failures.WrapTerminal(errors.New("occurred_at is required"))
	}
	if strings.TrimSpace(string(event.EventType)) == "" {
		return failures.WrapTerminal(errors.New("event_type is required"))
	}
	if !slices.Contains([]Source{SourceACP, SourceSessionFile, SourceWorker}, event.Source) {
		return failures.WrapTerminal(errors.New("source is invalid"))
	}
	if err := event.CorrelationIDs.Validate(); err != nil {
		return err
	}
	if event.Payload == nil {
		return failures.WrapTerminal(errors.New("payload is required"))
	}
	return nil
}
