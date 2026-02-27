package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
	"github.com/sirupsen/logrus"
)

type TaskboardExecuteHandler struct {
	pipeline *apptaskboard.ExecutionPipelineService
	registry *apptaskboard.ExecutionRegistry
	logger   *logrus.Logger
}

func NewTaskboardExecuteHandler(pipeline *apptaskboard.ExecutionPipelineService, registry *apptaskboard.ExecutionRegistry, logger *logrus.Logger) *TaskboardExecuteHandler {
	return &TaskboardExecuteHandler{pipeline: pipeline, registry: registry, logger: logger}
}

func (handler *TaskboardExecuteHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if handler.pipeline == nil {
		return fmt.Errorf("pipeline service is required")
	}

	var payload tasks.TaskboardExecutePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode taskboard execute payload: %w", err)
	}

	entry := handler.entry().WithFields(logrus.Fields{
		"event":           "worker.taskboard_execute",
		"board_id":        strings.TrimSpace(payload.BoardID),
		"source_branch":   strings.TrimSpace(payload.SourceBranch),
		"repository_root": strings.TrimSpace(payload.RepositoryRoot),
		"task_type":       task.Type(),
	})
	entry.Info("processing taskboard execution pipeline")

	executionCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	cleanBoardID := strings.TrimSpace(payload.BoardID)
	if handler.registry != nil {
		handler.registry.Register(cleanBoardID, cancel)
		defer handler.registry.Unregister(cleanBoardID)
	}

	err := handler.pipeline.ExecuteBoard(executionCtx, payload.BoardID, payload.SourceBranch, payload.RepositoryRoot, payload.MaxTasks)
	if err == nil {
		entry.Info("taskboard execution pipeline completed")
		return nil
	}
	if errors.Is(err, context.Canceled) {
		entry.WithError(err).Warn("taskboard execution pipeline interrupted; retrying for automatic resume")
		return fmt.Errorf("taskboard execution interrupted: %w", err)
	}

	entry.WithError(err).Error("taskboard execution pipeline failed")
	if appgitflow.IsTerminalFailure(err) {
		return fmt.Errorf("%w: taskboard execution: %v", asynq.SkipRetry, err)
	}
	return fmt.Errorf("taskboard execution: %w", err)
}

func (handler *TaskboardExecuteHandler) entry() *logrus.Entry {
	if handler.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(handler.logger)
}
