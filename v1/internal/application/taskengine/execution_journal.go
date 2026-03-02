package taskengine

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ExecutionStatus string

const (
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusSucceeded ExecutionStatus = "succeeded"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusSkipped   ExecutionStatus = "skipped"
)

type ExecutionRecord struct {
	RunID          string
	TaskID         string
	JobID          string
	ProjectID      string
	JobKind        JobKind
	IdempotencyKey string
	Step           string
	Status         ExecutionStatus
	ErrorMessage   string
	UpdatedAt      time.Time
}

func (record ExecutionRecord) Validate() error {
	if strings.TrimSpace(record.RunID) == "" {
		return fmt.Errorf("%w: run_id is required", ErrInvalidExecutionRecord)
	}
	if strings.TrimSpace(record.TaskID) == "" {
		return fmt.Errorf("%w: task_id is required", ErrInvalidExecutionRecord)
	}
	if strings.TrimSpace(record.JobID) == "" {
		return fmt.Errorf("%w: job_id is required", ErrInvalidExecutionRecord)
	}
	if strings.TrimSpace(string(record.JobKind)) == "" {
		return fmt.Errorf("%w: job_kind is required", ErrInvalidExecutionRecord)
	}
	if strings.TrimSpace(record.IdempotencyKey) == "" {
		return fmt.Errorf("%w: idempotency_key is required", ErrInvalidExecutionRecord)
	}
	if strings.TrimSpace(record.Step) == "" {
		return fmt.Errorf("%w: step is required", ErrInvalidExecutionRecord)
	}
	switch record.Status {
	case ExecutionStatusRunning, ExecutionStatusSucceeded, ExecutionStatusFailed, ExecutionStatusSkipped:
	default:
		return fmt.Errorf("%w: unsupported status %q", ErrInvalidExecutionRecord, record.Status)
	}
	if record.UpdatedAt.IsZero() {
		return fmt.Errorf("%w: updated_at is required", ErrInvalidExecutionRecord)
	}
	return nil
}

type ExecutionJournal interface {
	Upsert(ctx context.Context, record ExecutionRecord) error
	Load(ctx context.Context, runID string, taskID string, jobID string) (*ExecutionRecord, error)
}
