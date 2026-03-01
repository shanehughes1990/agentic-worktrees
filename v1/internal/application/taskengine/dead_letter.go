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

type DeadLetterManager interface {
	ListDeadLetters(ctx context.Context, queue string, limit int) ([]DeadLetterTask, error)
	RequeueDeadLetter(ctx context.Context, queue string, taskID string) error
}
