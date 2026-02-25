package tasks

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

const TaskTypeGitConflictResolve = "git.conflict.resolve"

type GitConflictResolvePayload struct {
	RunID          string   `json:"run_id"`
	TaskID         string   `json:"task_id"`
	RepositoryRoot string   `json:"repository_root"`
	SourceBranch   string   `json:"source_branch"`
	TaskBranch     string   `json:"task_branch"`
	WorktreePath   string   `json:"worktree_path"`
	ConflictFiles  []string `json:"conflict_files"`
	IdempotencyKey string   `json:"idempotency_key"`
}

func NewGitConflictResolveTask(payload GitConflictResolvePayload, options ...asynq.Option) (*asynq.Task, []asynq.Option, error) {
	if strings.TrimSpace(payload.RunID) == "" {
		return nil, nil, fmt.Errorf("run_id is required")
	}
	if strings.TrimSpace(payload.TaskID) == "" {
		return nil, nil, fmt.Errorf("task_id is required")
	}
	if strings.TrimSpace(payload.RepositoryRoot) == "" {
		return nil, nil, fmt.Errorf("repository_root is required")
	}
	if strings.TrimSpace(payload.SourceBranch) == "" {
		return nil, nil, fmt.Errorf("source_branch is required")
	}
	if strings.TrimSpace(payload.TaskBranch) == "" {
		return nil, nil, fmt.Errorf("task_branch is required")
	}
	if strings.TrimSpace(payload.WorktreePath) == "" {
		return nil, nil, fmt.Errorf("worktree_path is required")
	}
	if !isWorktreePathUnderAppRoot(payload.WorktreePath) {
		return nil, nil, fmt.Errorf("worktree_path must be under <app_root>/worktrees")
	}
	if len(payload.ConflictFiles) == 0 {
		return nil, nil, fmt.Errorf("conflict_files is required")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal git conflict resolve payload: %w", err)
	}

	task := asynq.NewTask(TaskTypeGitConflictResolve, body)
	opts := []asynq.Option{asynq.Queue(queueAgent), asynq.Retention(24 * time.Hour)}
	opts = append(opts, options...)
	return task, opts, nil
}
