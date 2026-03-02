package taskengine

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type DeadLetterTask struct {
	Queue       string
	TaskID      string
	JobKind     JobKind
	Payload     []byte
	LastError   string
	Retried     int
	MaxRetry    int
	ArchivedAt  time.Time
	FailedAt    time.Time
	CompletedAt time.Time
}

func (task DeadLetterTask) Validate() error {
	if strings.TrimSpace(task.Queue) == "" {
		return fmt.Errorf("%w: queue is required", ErrInvalidDeadLetterRequest)
	}
	if strings.TrimSpace(task.TaskID) == "" {
		return fmt.Errorf("%w: task_id is required", ErrInvalidDeadLetterRequest)
	}
	if strings.TrimSpace(string(task.JobKind)) == "" {
		return fmt.Errorf("%w: job_kind is required", ErrInvalidDeadLetterRequest)
	}
	return nil
}

type DeadLetterAction string

const (
	DeadLetterActionRequeue DeadLetterAction = "requeue"
	DeadLetterActionDiscard DeadLetterAction = "discard"
)

type DeadLetterEvent struct {
	Queue      string
	TaskID     string
	JobKind    JobKind
	Action     DeadLetterAction
	LastError  string
	Reason     string
	Actor      string
	OccurredAt time.Time
}

func (event DeadLetterEvent) Validate() error {
	if strings.TrimSpace(event.Queue) == "" {
		return fmt.Errorf("%w: queue is required", ErrInvalidDeadLetterRequest)
	}
	if strings.TrimSpace(event.TaskID) == "" {
		return fmt.Errorf("%w: task_id is required", ErrInvalidDeadLetterRequest)
	}
	switch event.Action {
	case DeadLetterActionRequeue, DeadLetterActionDiscard:
	default:
		return fmt.Errorf("%w: action is required", ErrInvalidDeadLetterRequest)
	}
	if event.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at is required", ErrInvalidDeadLetterRequest)
	}
	return nil
}

type DeadLetterAudit interface {
	Record(ctx context.Context, event DeadLetterEvent) error
}

type DeadLetterManager interface {
	ListDeadLetters(ctx context.Context, queue string, limit int) ([]DeadLetterTask, error)
	RequeueDeadLetter(ctx context.Context, queue string, taskID string) error
	DeleteProjectTasks(ctx context.Context, projectID string) error
}
