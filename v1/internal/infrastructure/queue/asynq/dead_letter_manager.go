package asynq

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

func (platform *Platform) ListDeadLetters(ctx context.Context, queue string, limit int) ([]taskengine.DeadLetterTask, error) {
	_ = ctx
	if platform == nil {
		return nil, fmt.Errorf("task engine platform is not initialized")
	}
	queue = strings.TrimSpace(queue)
	if queue == "" {
		return nil, fmt.Errorf("%w: queue is required", taskengine.ErrInvalidDeadLetterRequest)
	}
	if limit <= 0 {
		limit = 50
	}

	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     platform.config.RedisAddress,
		Password: platform.config.RedisPassword,
		DB:       platform.config.RedisDatabase,
	})
	defer inspector.Close()
	tasks, err := inspector.ListArchivedTasks(queue, asynq.PageSize(limit), asynq.Page(0))
	if err != nil {
		return nil, fmt.Errorf("list dead letters: %w", err)
	}

	result := make([]taskengine.DeadLetterTask, 0, len(tasks))
	for _, task := range tasks {
		record := taskengine.DeadLetterTask{
			Queue:       queue,
			TaskID:      task.ID,
			JobKind:     taskengine.JobKind(task.Type),
			Payload:     task.Payload,
			LastError:   task.LastErr,
			Retried:     task.Retried,
			MaxRetry:    task.MaxRetry,
			ArchivedAt:  task.LastFailedAt,
			FailedAt:    task.LastFailedAt,
			CompletedAt: task.CompletedAt,
		}
		if err := record.Validate(); err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}

func (platform *Platform) RequeueDeadLetter(ctx context.Context, queue string, taskID string) error {
	if platform == nil {
		return fmt.Errorf("task engine platform is not initialized")
	}
	queue = strings.TrimSpace(queue)
	taskID = strings.TrimSpace(taskID)
	if queue == "" {
		return fmt.Errorf("%w: queue is required", taskengine.ErrInvalidDeadLetterRequest)
	}
	if taskID == "" {
		return fmt.Errorf("%w: task_id is required", taskengine.ErrInvalidDeadLetterRequest)
	}

	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     platform.config.RedisAddress,
		Password: platform.config.RedisPassword,
		DB:       platform.config.RedisDatabase,
	})
	defer inspector.Close()

	archivedTasks, err := inspector.ListArchivedTasks(queue, asynq.PageSize(500), asynq.Page(0))
	if err != nil {
		return fmt.Errorf("list dead letters: %w", err)
	}
	var requeuedTask *asynq.TaskInfo
	for _, archivedTask := range archivedTasks {
		if strings.TrimSpace(archivedTask.ID) == taskID {
			requeuedTask = archivedTask
			break
		}
	}
	if requeuedTask == nil {
		return fmt.Errorf("requeue dead letter: task %q not found in queue %q", taskID, queue)
	}

	if err := inspector.RunTask(queue, taskID); err != nil {
		return fmt.Errorf("requeue dead letter: %w", err)
	}
	if platform.deadLetterAudit != nil {
		auditErr := platform.deadLetterAudit.Record(ctx, taskengine.DeadLetterEvent{
			Queue:      queue,
			TaskID:     taskID,
			JobKind:    taskengine.JobKind(strings.TrimSpace(requeuedTask.Type)),
			Action:     taskengine.DeadLetterActionRequeue,
			LastError:  strings.TrimSpace(requeuedTask.LastErr),
			Reason:     "manual requeue",
			Actor:      "system",
			OccurredAt: time.Now().UTC(),
		})
		if auditErr != nil {
			return fmt.Errorf("record dead-letter requeue audit: %w", auditErr)
		}
	}
	return nil
}
