package asynq

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

func (platform *APIPlatform) ListDeadLetters(ctx context.Context, queue string, limit int) ([]taskengine.DeadLetterTask, error) {
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
	redisConnOpt, err := asynq.ParseRedisURI(platform.redisURL)
	if err != nil {
		return nil, fmt.Errorf("build redis client options: %w", err)
	}
	inspector := asynq.NewInspector(redisConnOpt)
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

func (platform *APIPlatform) RequeueDeadLetter(ctx context.Context, queue string, taskID string) error {
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

	redisConnOpt, err := asynq.ParseRedisURI(platform.redisURL)
	if err != nil {
		return fmt.Errorf("build redis client options: %w", err)
	}
	inspector := asynq.NewInspector(redisConnOpt)
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

func (platform *APIPlatform) DeleteProjectTasks(ctx context.Context, projectID string) error {
	_ = ctx
	if platform == nil {
		return fmt.Errorf("task engine platform is not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return fmt.Errorf("project_id is required")
	}
	redisConnOpt, err := asynq.ParseRedisURI(platform.redisURL)
	if err != nil {
		return fmt.Errorf("build redis client options: %w", err)
	}
	inspector := asynq.NewInspector(redisConnOpt)
	defer inspector.Close()

	queues, err := inspector.Queues()
	if err != nil {
		return fmt.Errorf("list queues: %w", err)
	}
	for _, queue := range queues {
		if err := deleteProjectTasksInQueue(inspector, strings.TrimSpace(queue), projectID); err != nil {
			return err
		}
	}
	return nil
}

func deleteProjectTasksInQueue(inspector *asynq.Inspector, queue string, projectID string) error {
	if queue == "" {
		return nil
	}
	if err := deleteProjectTasksByState(inspector, queue, projectID, func(page int) ([]*asynq.TaskInfo, error) {
		return inspector.ListPendingTasks(queue, asynq.PageSize(200), asynq.Page(page))
	}); err != nil {
		return fmt.Errorf("delete project tasks from pending queue %q: %w", queue, err)
	}
	if err := deleteProjectTasksByState(inspector, queue, projectID, func(page int) ([]*asynq.TaskInfo, error) {
		return inspector.ListScheduledTasks(queue, asynq.PageSize(200), asynq.Page(page))
	}); err != nil {
		return fmt.Errorf("delete project tasks from scheduled queue %q: %w", queue, err)
	}
	if err := deleteProjectTasksByState(inspector, queue, projectID, func(page int) ([]*asynq.TaskInfo, error) {
		return inspector.ListRetryTasks(queue, asynq.PageSize(200), asynq.Page(page))
	}); err != nil {
		return fmt.Errorf("delete project tasks from retry queue %q: %w", queue, err)
	}
	if err := deleteProjectTasksByState(inspector, queue, projectID, func(page int) ([]*asynq.TaskInfo, error) {
		return inspector.ListArchivedTasks(queue, asynq.PageSize(200), asynq.Page(page))
	}); err != nil {
		return fmt.Errorf("delete project tasks from archived queue %q: %w", queue, err)
	}
	if err := deleteProjectTasksByState(inspector, queue, projectID, func(page int) ([]*asynq.TaskInfo, error) {
		return inspector.ListCompletedTasks(queue, asynq.PageSize(200), asynq.Page(page))
	}); err != nil {
		return fmt.Errorf("delete project tasks from completed queue %q: %w", queue, err)
	}
	if err := deleteProjectTasksByState(inspector, queue, projectID, func(page int) ([]*asynq.TaskInfo, error) {
		return inspector.ListActiveTasks(queue, asynq.PageSize(200), asynq.Page(page))
	}); err != nil {
		return fmt.Errorf("delete project tasks from active queue %q: %w", queue, err)
	}
	return nil
}

func deleteProjectTasksByState(inspector *asynq.Inspector, queue string, projectID string, listPage func(page int) ([]*asynq.TaskInfo, error)) error {
	for page := 0; ; page++ {
		tasks, err := listPage(page)
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			return nil
		}
		for _, task := range tasks {
			if !matchesProjectID(task, projectID) {
				continue
			}
			_ = inspector.CancelProcessing(task.ID)
			_ = inspector.DeleteTask(queue, task.ID)
		}
	}
}

func matchesProjectID(task *asynq.TaskInfo, projectID string) bool {
	if task == nil {
		return false
	}
	type projectScopedPayload struct {
		ProjectID string `json:"project_id"`
	}
	var payload projectScopedPayload
	if err := json.Unmarshal(task.Payload, &payload); err == nil {
		return strings.TrimSpace(payload.ProjectID) == projectID
	}
	return strings.Contains(string(task.Payload), fmt.Sprintf("\"project_id\":\"%s\"", projectID))
}
