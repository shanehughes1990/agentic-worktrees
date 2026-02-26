package asynq

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
)

const (
	ingestionQueueName = "ingestion"
	agentQueueName     = "agent"
)

type RuntimeWorkflowRepository struct {
	inspector *asynq.Inspector
}

type runtimeTaskLoader struct {
	queue string
	state string
	load  func() ([]*asynq.TaskInfo, error)
}

type runtimeTaskSnapshot struct {
	queue string
	state string
	info  *asynq.TaskInfo
}

func NewRuntimeWorkflowRepository(cfg Config) *RuntimeWorkflowRepository {
	return &RuntimeWorkflowRepository{inspector: asynq.NewInspector(cfg.redisConnOpt)}
}

func (repository *RuntimeWorkflowRepository) Close() error {
	if repository == nil || repository.inspector == nil {
		return nil
	}
	return repository.inspector.Close()
}

func (repository *RuntimeWorkflowRepository) ListRuntimeWorkflows(ctx context.Context) ([]apptaskboard.IngestionWorkflow, error) {
	snapshots, err := repository.loadRuntimeTaskSnapshots(ctx)
	if err != nil {
		return nil, err
	}

	workflowsByRunID := map[string]apptaskboard.IngestionWorkflow{}
	for _, snapshot := range snapshots {
		if snapshot.info == nil || !isRuntimeWorkflowTaskType(snapshot.info.Type) {
			continue
		}
		workflow := mapTaskToWorkflow(snapshot.info)
		if workflow.RunID == "" {
			continue
		}
		existing, exists := workflowsByRunID[workflow.RunID]
		if !exists || workflow.UpdatedAt.After(existing.UpdatedAt) {
			workflowsByRunID[workflow.RunID] = workflow
		}
	}

	workflows := make([]apptaskboard.IngestionWorkflow, 0, len(workflowsByRunID))
	for _, workflow := range workflowsByRunID {
		workflows = append(workflows, workflow)
	}
	sort.Slice(workflows, func(i, j int) bool {
		if workflows[i].UpdatedAt.Equal(workflows[j].UpdatedAt) {
			return workflows[i].RunID > workflows[j].RunID
		}
		return workflows[i].UpdatedAt.After(workflows[j].UpdatedAt)
	})
	return workflows, nil
}

func (repository *RuntimeWorkflowRepository) GetRuntimeWorkflow(ctx context.Context, runID string) (*apptaskboard.IngestionWorkflow, error) {
	workflows, err := repository.ListRuntimeWorkflows(ctx)
	if err != nil {
		return nil, err
	}
	cleanRunID := strings.TrimSpace(runID)
	for i := range workflows {
		if workflows[i].RunID == cleanRunID {
			workflow := workflows[i]
			return &workflow, nil
		}
	}
	return nil, nil
}

func (repository *RuntimeWorkflowRepository) CancelRuntimeWorkflow(ctx context.Context, runID string) (apptaskboard.WorkflowCancelResult, error) {
	cleanRunID := strings.TrimSpace(runID)
	if cleanRunID == "" {
		return apptaskboard.WorkflowCancelResult{}, fmt.Errorf("run_id is required")
	}

	snapshots, err := repository.loadRuntimeTaskSnapshots(ctx)
	if err != nil {
		return apptaskboard.WorkflowCancelResult{}, err
	}

	result := apptaskboard.WorkflowCancelResult{RunID: cleanRunID}
	for _, snapshot := range snapshots {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}
		if snapshot.info == nil || !isRuntimeWorkflowTaskType(snapshot.info.Type) {
			continue
		}
		if extractRunID(snapshot.info) != cleanRunID {
			continue
		}
		result.MatchedTasks++

		switch strings.ToLower(strings.TrimSpace(snapshot.state)) {
		case "active":
			if err := repository.inspector.CancelProcessing(snapshot.info.ID); err != nil {
				result.UncancelableTasks++
				continue
			}
			result.SignaledActive++
			result.CanceledTasks++
		case "pending", "scheduled", "retry":
			if err := repository.inspector.ArchiveTask(snapshot.queue, snapshot.info.ID); err != nil {
				result.UncancelableTasks++
				continue
			}
			result.CanceledTasks++
		default:
			result.UncancelableTasks++
		}
	}
	return result, nil
}

func (repository *RuntimeWorkflowRepository) loadRuntimeTaskSnapshots(ctx context.Context) ([]runtimeTaskSnapshot, error) {
	snapshots := make([]runtimeTaskSnapshot, 0, 256)
	for _, loader := range repository.runtimeTaskLoaders() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		tasksInState, err := loader.load()
		if err != nil {
			if isQueueNotFoundError(err) {
				continue
			}
			return nil, fmt.Errorf("list runtime workflows from asynq: %w", err)
		}
		for _, info := range tasksInState {
			snapshots = append(snapshots, runtimeTaskSnapshot{queue: loader.queue, state: loader.state, info: info})
		}
	}
	return snapshots, nil
}

func (repository *RuntimeWorkflowRepository) runtimeTaskLoaders() []runtimeTaskLoader {
	return []runtimeTaskLoader{
		{queue: ingestionQueueName, state: "pending", load: func() ([]*asynq.TaskInfo, error) {
			return repository.inspector.ListPendingTasks(ingestionQueueName, asynq.PageSize(200), asynq.Page(0))
		}},
		{queue: agentQueueName, state: "pending", load: func() ([]*asynq.TaskInfo, error) {
			return repository.inspector.ListPendingTasks(agentQueueName, asynq.PageSize(200), asynq.Page(0))
		}},
		{queue: ingestionQueueName, state: "active", load: func() ([]*asynq.TaskInfo, error) {
			return repository.inspector.ListActiveTasks(ingestionQueueName, asynq.PageSize(200), asynq.Page(0))
		}},
		{queue: agentQueueName, state: "active", load: func() ([]*asynq.TaskInfo, error) {
			return repository.inspector.ListActiveTasks(agentQueueName, asynq.PageSize(200), asynq.Page(0))
		}},
		{queue: ingestionQueueName, state: "scheduled", load: func() ([]*asynq.TaskInfo, error) {
			return repository.inspector.ListScheduledTasks(ingestionQueueName, asynq.PageSize(200), asynq.Page(0))
		}},
		{queue: agentQueueName, state: "scheduled", load: func() ([]*asynq.TaskInfo, error) {
			return repository.inspector.ListScheduledTasks(agentQueueName, asynq.PageSize(200), asynq.Page(0))
		}},
		{queue: ingestionQueueName, state: "retry", load: func() ([]*asynq.TaskInfo, error) {
			return repository.inspector.ListRetryTasks(ingestionQueueName, asynq.PageSize(200), asynq.Page(0))
		}},
		{queue: agentQueueName, state: "retry", load: func() ([]*asynq.TaskInfo, error) {
			return repository.inspector.ListRetryTasks(agentQueueName, asynq.PageSize(200), asynq.Page(0))
		}},
		{queue: ingestionQueueName, state: "completed", load: func() ([]*asynq.TaskInfo, error) {
			return repository.inspector.ListCompletedTasks(ingestionQueueName, asynq.PageSize(200), asynq.Page(0))
		}},
		{queue: agentQueueName, state: "completed", load: func() ([]*asynq.TaskInfo, error) {
			return repository.inspector.ListCompletedTasks(agentQueueName, asynq.PageSize(200), asynq.Page(0))
		}},
		{queue: ingestionQueueName, state: "archived", load: func() ([]*asynq.TaskInfo, error) {
			return repository.inspector.ListArchivedTasks(ingestionQueueName, asynq.PageSize(200), asynq.Page(0))
		}},
		{queue: agentQueueName, state: "archived", load: func() ([]*asynq.TaskInfo, error) {
			return repository.inspector.ListArchivedTasks(agentQueueName, asynq.PageSize(200), asynq.Page(0))
		}},
	}
}

func isRuntimeWorkflowTaskType(taskType string) bool {
	cleanTaskType := strings.TrimSpace(taskType)
	switch cleanTaskType {
	case apptaskboard.WorkflowTaskTypeCopilotDecompose, apptaskboard.WorkflowTaskTypeGitWorktreeFlow, apptaskboard.WorkflowTaskTypeGitConflict, apptaskboard.WorkflowTaskTypeTaskboardExecute:
		return true
	default:
		return false
	}
}

func mapTaskToWorkflow(info *asynq.TaskInfo) apptaskboard.IngestionWorkflow {
	runID := extractRunID(info)
	if runID == "" {
		runID = strings.TrimSpace(info.ID)
	}
	stream := strings.TrimSpace(string(info.Result))
	if stream == "" {
		stream = "(stream details not available yet in Asynq result)"
	}
	return apptaskboard.IngestionWorkflow{
		RunID:      runID,
		TaskID:     info.ID,
		TaskType:   strings.TrimSpace(info.Type),
		Status:     mapTaskState(info.State),
		Message:    mapTaskMessage(info),
		Stream:     stream,
		UpdatedAt:  mapTaskUpdatedAt(info),
		CreatedAt:  time.Time{},
		Cancelable: isCancelableTaskState(info.State),
		Details: map[string]any{
			"asynq_task_id":    strings.TrimSpace(info.ID),
			"asynq_task_type":  strings.TrimSpace(info.Type),
			"asynq_state":      strings.TrimSpace(info.State.String()),
			"asynq_retry_count": info.Retried,
			"asynq_max_retry":   info.MaxRetry,
		},
	}
}

func extractRunID(info *asynq.TaskInfo) string {
	if info == nil {
		return ""
	}
	switch strings.TrimSpace(info.Type) {
	case tasks.TaskTypeCopilotDecompose:
		payload := tasks.CopilotDecomposePayload{}
		if len(info.Payload) > 0 {
			_ = json.Unmarshal(info.Payload, &payload)
		}
		return strings.TrimSpace(payload.RunID)
	case tasks.TaskTypeGitWorktreeFlow:
		payload := tasks.GitWorktreeFlowPayload{}
		if len(info.Payload) > 0 {
			_ = json.Unmarshal(info.Payload, &payload)
		}
		return strings.TrimSpace(payload.RunID)
	case tasks.TaskTypeGitConflictResolve:
		payload := tasks.GitConflictResolvePayload{}
		if len(info.Payload) > 0 {
			_ = json.Unmarshal(info.Payload, &payload)
		}
		return strings.TrimSpace(payload.RunID)
	case tasks.TaskTypeTaskboardExecute:
		payload := tasks.TaskboardExecutePayload{}
		if len(info.Payload) > 0 {
			_ = json.Unmarshal(info.Payload, &payload)
		}
		return strings.TrimSpace(payload.BoardID)
	default:
		return ""
	}
}

func mapTaskState(state asynq.TaskState) apptaskboard.WorkflowStatus {
	switch strings.ToLower(strings.TrimSpace(state.String())) {
	case "active":
		return apptaskboard.WorkflowStatusRunning
	case "pending", "scheduled", "retry":
		return apptaskboard.WorkflowStatusQueued
	case "completed":
		return apptaskboard.WorkflowStatusCompleted
	case "archived":
		return apptaskboard.WorkflowStatusCanceled
	default:
		return apptaskboard.WorkflowStatusQueued
	}
}

func mapTaskMessage(info *asynq.TaskInfo) string {
	lastErr := strings.TrimSpace(info.LastErr)
	if lastErr != "" {
		return lastErr
	}
	stateText := strings.TrimSpace(info.State.String())
	if stateText == "" {
		return "asynq task status available"
	}
	if !info.NextProcessAt.IsZero() {
		return fmt.Sprintf("state=%s next_process_at=%s", stateText, info.NextProcessAt.UTC().Format(time.RFC3339))
	}
	return fmt.Sprintf("state=%s", stateText)
}

func mapTaskUpdatedAt(info *asynq.TaskInfo) time.Time {
	if !info.CompletedAt.IsZero() {
		return info.CompletedAt.UTC()
	}
	if !info.LastFailedAt.IsZero() {
		return info.LastFailedAt.UTC()
	}
	if !info.NextProcessAt.IsZero() {
		return info.NextProcessAt.UTC()
	}
	return time.Now().UTC()
}

func isCancelableTaskState(state asynq.TaskState) bool {
	switch strings.ToLower(strings.TrimSpace(state.String())) {
	case "pending", "active", "scheduled", "retry":
		return true
	default:
		return false
	}
}

func isQueueNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "queue not found")
}
