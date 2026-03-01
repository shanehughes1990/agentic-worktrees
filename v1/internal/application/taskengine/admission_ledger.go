package taskengine

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type AdmissionStatus string

const (
	AdmissionStatusQueued AdmissionStatus = "queued"
)

type AdmissionRecord struct {
	RunID          string
	TaskID         string
	JobID          string
	JobKind        JobKind
	IdempotencyKey string
	QueueTaskID    string
	Queue          string
	Status         AdmissionStatus
	Duplicate      bool
	EnqueuedAt     time.Time
}

func (record AdmissionRecord) Validate() error {
	if strings.TrimSpace(record.RunID) == "" {
		return fmt.Errorf("%w: run_id is required", ErrInvalidEnqueueRequest)
	}
	if strings.TrimSpace(record.TaskID) == "" {
		return fmt.Errorf("%w: task_id is required", ErrInvalidEnqueueRequest)
	}
	if strings.TrimSpace(record.JobID) == "" {
		return fmt.Errorf("%w: job_id is required", ErrInvalidEnqueueRequest)
	}
	if strings.TrimSpace(string(record.JobKind)) == "" {
		return fmt.Errorf("%w: job_kind is required", ErrInvalidEnqueueRequest)
	}
	if strings.TrimSpace(record.IdempotencyKey) == "" {
		return fmt.Errorf("%w: idempotency_key is required", ErrInvalidEnqueueRequest)
	}
	if strings.TrimSpace(record.QueueTaskID) == "" {
		return fmt.Errorf("%w: queue_task_id is required", ErrInvalidEnqueueRequest)
	}
	if strings.TrimSpace(record.Queue) == "" {
		return fmt.Errorf("%w: queue is required", ErrInvalidEnqueueRequest)
	}
	if record.Status != AdmissionStatusQueued {
		return fmt.Errorf("%w: unsupported admission status %q", ErrInvalidEnqueueRequest, record.Status)
	}
	if record.EnqueuedAt.IsZero() {
		return fmt.Errorf("%w: enqueued_at is required", ErrInvalidEnqueueRequest)
	}
	return nil
}

type AdmissionLedger interface {
	Upsert(ctx context.Context, record AdmissionRecord) error
}
