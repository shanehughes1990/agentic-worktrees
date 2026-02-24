package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
	"github.com/sirupsen/logrus"
)

type GitWorktreeFlowHandler struct {
	runner *appgitflow.Runner
	logger *logrus.Logger
}

func NewGitWorktreeFlowHandler(runner *appgitflow.Runner, logger *logrus.Logger) *GitWorktreeFlowHandler {
	return &GitWorktreeFlowHandler{runner: runner, logger: logger}
}

func (handler *GitWorktreeFlowHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if handler.runner == nil {
		return fmt.Errorf("runner is required")
	}

	var payload tasks.GitWorktreeFlowPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode git worktree flow payload: %w", err)
	}

	entry := handler.entry().WithFields(logrus.Fields{
		"event":           "worker.git_worktree_flow",
		"run_id":          strings.TrimSpace(payload.RunID),
		"task_id":         strings.TrimSpace(payload.TaskID),
		"task_type":       task.Type(),
		"source_branch":   strings.TrimSpace(payload.SourceBranch),
		"task_branch":     strings.TrimSpace(payload.TaskBranch),
		"worktree_path":   strings.TrimSpace(payload.WorktreePath),
		"repository_root": strings.TrimSpace(payload.RepositoryRoot),
	})
	entry.Info("processing git worktree flow task")

	err := handler.runner.RunWorktreeFlow(ctx, appgitflow.WorktreeFlowJob{
		RunID:          payload.RunID,
		TaskID:         payload.TaskID,
		RepositoryRoot: payload.RepositoryRoot,
		SourceBranch:   payload.SourceBranch,
		TaskBranch:     payload.TaskBranch,
		WorktreePath:   payload.WorktreePath,
	})
	if err == nil {
		entry.Info("git worktree flow task completed")
		return nil
	}
	entry.WithError(err).Error("git worktree flow task failed")
	if appgitflow.IsTerminalFailure(err) {
		return fmt.Errorf("%w: git worktree flow: %v", asynq.SkipRetry, err)
	}
	return fmt.Errorf("git worktree flow: %w", err)
}

func (handler *GitWorktreeFlowHandler) entry() *logrus.Entry {
	if handler.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(handler.logger)
}
