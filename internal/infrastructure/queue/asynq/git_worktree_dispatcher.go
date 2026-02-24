package asynq

import (
	"context"

	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
	"github.com/sirupsen/logrus"
)

type GitWorktreeDispatcher struct {
	client *Client
	logger *logrus.Logger
}

func NewGitWorktreeDispatcher(client *Client, logger *logrus.Logger) *GitWorktreeDispatcher {
	return &GitWorktreeDispatcher{client: client, logger: logger}
}

func (dispatcher *GitWorktreeDispatcher) EnqueueWorktreeFlow(ctx context.Context, job appgitflow.WorktreeFlowJob) (string, error) {
	entry := dispatcher.entry().WithFields(logrus.Fields{
		"event":           "gitflow.enqueue",
		"run_id":          job.RunID,
		"task_id":         job.TaskID,
		"source_branch":   job.SourceBranch,
		"task_branch":     job.TaskBranch,
		"worktree_path":   job.WorktreePath,
		"repository_root": job.RepositoryRoot,
	})
	entry.Info("enqueueing git worktree flow task")

	taskInfo, err := dispatcher.client.EnqueueGitWorktreeFlow(ctx, tasks.GitWorktreeFlowPayload{
		RunID:          job.RunID,
		TaskID:         job.TaskID,
		RepositoryRoot: job.RepositoryRoot,
		SourceBranch:   job.SourceBranch,
		TaskBranch:     job.TaskBranch,
		WorktreePath:   job.WorktreePath,
		IdempotencyKey: job.RunID + ":" + job.TaskID,
	})
	if err != nil {
		entry.WithError(err).Error("failed to enqueue git worktree flow task")
		return "", err
	}
	if taskInfo == nil {
		entry.Warn("enqueue returned nil task info")
		return "", nil
	}
	entry.WithField("task_queue_id", taskInfo.ID).Info("git worktree flow task enqueued")
	return taskInfo.ID, nil
}

func (dispatcher *GitWorktreeDispatcher) EnqueueConflictResolution(ctx context.Context, job appgitflow.ConflictResolveJob) (string, error) {
	entry := dispatcher.entry().WithFields(logrus.Fields{
		"event":           "gitflow.enqueue_conflict",
		"run_id":          job.RunID,
		"task_id":         job.TaskID,
		"source_branch":   job.SourceBranch,
		"task_branch":     job.TaskBranch,
		"worktree_path":   job.WorktreePath,
		"repository_root": job.RepositoryRoot,
		"conflict_count":  len(job.ConflictFiles),
	})
	entry.Info("enqueueing git conflict resolve task")

	taskInfo, err := dispatcher.client.EnqueueGitConflictResolve(ctx, tasks.GitConflictResolvePayload{
		RunID:          job.RunID,
		TaskID:         job.TaskID,
		RepositoryRoot: job.RepositoryRoot,
		SourceBranch:   job.SourceBranch,
		TaskBranch:     job.TaskBranch,
		WorktreePath:   job.WorktreePath,
		ConflictFiles:  job.ConflictFiles,
		IdempotencyKey: job.IdempotencyKey,
	})
	if err != nil {
		entry.WithError(err).Error("failed to enqueue git conflict resolve task")
		return "", err
	}
	if taskInfo == nil {
		entry.Warn("enqueue returned nil task info")
		return "", nil
	}
	entry.WithField("task_queue_id", taskInfo.ID).Info("git conflict resolve task enqueued")
	return taskInfo.ID, nil
}

func (dispatcher *GitWorktreeDispatcher) entry() *logrus.Entry {
	if dispatcher.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(dispatcher.logger)
}
