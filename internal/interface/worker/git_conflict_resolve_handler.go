package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
	"github.com/sirupsen/logrus"
)

type GitConflictResolveHandler struct {
	runner     *appgitflow.Runner
	decomposer appcopilot.Decomposer
	logger     *logrus.Logger
}

func NewGitConflictResolveHandler(runner *appgitflow.Runner, decomposer appcopilot.Decomposer, logger *logrus.Logger) *GitConflictResolveHandler {
	return &GitConflictResolveHandler{runner: runner, decomposer: decomposer, logger: logger}
}

func (handler *GitConflictResolveHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if handler.runner == nil {
		return fmt.Errorf("runner is required")
	}

	var payload tasks.GitConflictResolvePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode git conflict resolve payload: %w", err)
	}

	entry := handler.entry().WithFields(logrus.Fields{
		"event":           "worker.git_conflict_resolve",
		"run_id":          strings.TrimSpace(payload.RunID),
		"task_id":         strings.TrimSpace(payload.TaskID),
		"task_type":       task.Type(),
		"source_branch":   strings.TrimSpace(payload.SourceBranch),
		"task_branch":     strings.TrimSpace(payload.TaskBranch),
		"worktree_path":   strings.TrimSpace(payload.WorktreePath),
		"repository_root": strings.TrimSpace(payload.RepositoryRoot),
		"conflict_count":  len(payload.ConflictFiles),
	})
	entry.Info("processing git conflict resolve task")

	copilotAdvice := ""
	if handler.decomposer != nil {
		result, err := handler.decomposer.Decompose(ctx, appcopilot.DecomposeRequest{
			RunID:            payload.RunID,
			Prompt:           buildConflictResolutionPrompt(payload),
			WorkingDirectory: payload.RepositoryRoot,
		})
		if err != nil {
			entry.WithError(err).Warn("copilot conflict advice unavailable; continuing with deterministic resolution")
		} else {
			copilotAdvice = result.Response
		}
	}

	err := handler.runner.RunConflictResolution(ctx, appgitflow.ConflictResolveJob{
		RunID:          payload.RunID,
		TaskID:         payload.TaskID,
		RepositoryRoot: payload.RepositoryRoot,
		SourceBranch:   payload.SourceBranch,
		TaskBranch:     payload.TaskBranch,
		WorktreePath:   payload.WorktreePath,
		ConflictFiles:  payload.ConflictFiles,
		IdempotencyKey: payload.IdempotencyKey,
	}, copilotAdvice)
	if err == nil {
		entry.Info("git conflict resolve task completed")
		return nil
	}
	entry.WithError(err).Error("git conflict resolve task failed")
	if appgitflow.IsTerminalFailure(err) {
		return fmt.Errorf("%w: git conflict resolve: %v", asynq.SkipRetry, err)
	}
	return fmt.Errorf("git conflict resolve: %w", err)
}

func buildConflictResolutionPrompt(payload tasks.GitConflictResolvePayload) string {
	files := strings.Join(payload.ConflictFiles, ", ")
	return fmt.Sprintf("Resolve merge conflicts for run %s task %s on files: %s. Keep changes task-scoped and compile-safe.", strings.TrimSpace(payload.RunID), strings.TrimSpace(payload.TaskID), files)
}

func (handler *GitConflictResolveHandler) entry() *logrus.Entry {
	if handler.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(handler.logger)
}
