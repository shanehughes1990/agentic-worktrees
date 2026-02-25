package tasks

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

const TaskTypeGitWorktreeFlow = "git.worktree.flow"

type GitWorktreeFlowPayload struct {
	RunID          string `json:"run_id"`
	BoardID        string `json:"board_id,omitempty"`
	TaskID         string `json:"task_id"`
	TaskTitle      string `json:"task_title,omitempty"`
	TaskDetail     string `json:"task_detail,omitempty"`
	ResumeSessionID string `json:"resume_session_id,omitempty"`
	RepositoryRoot string `json:"repository_root"`
	SourceBranch   string `json:"source_branch"`
	TaskBranch     string `json:"task_branch"`
	WorktreePath   string `json:"worktree_path"`
	IdempotencyKey string `json:"idempotency_key"`
}

func NewGitWorktreeFlowTask(payload GitWorktreeFlowPayload, options ...asynq.Option) (*asynq.Task, []asynq.Option, error) {
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
	if !strings.HasPrefix(strings.TrimSpace(payload.WorktreePath), ".worktree/") {
		return nil, nil, fmt.Errorf("worktree_path must be under .worktree")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal git worktree flow payload: %w", err)
	}

	task := asynq.NewTask(TaskTypeGitWorktreeFlow, body)
	opts := []asynq.Option{asynq.Queue(queueAgent), asynq.Retention(24 * time.Hour)}
	opts = append(opts, options...)
	return task, opts, nil
}
