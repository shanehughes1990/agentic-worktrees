package tasks

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

const TaskTypeGitWorktreeFlow = "git.worktree.flow"

const defaultGitWorktreeFlowTimeout = 8 * time.Minute

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
	if strings.TrimSpace(payload.IdempotencyKey) == "" {
		payload.IdempotencyKey = strings.TrimSpace(payload.RunID) + ":" + strings.TrimSpace(payload.TaskID)
	}
	if !isWorktreePathUnderAppRoot(payload.WorktreePath) {
		return nil, nil, fmt.Errorf("worktree_path must be under <app_root>/worktrees")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal git worktree flow payload: %w", err)
	}

	task := asynq.NewTask(TaskTypeGitWorktreeFlow, body)
	opts := []asynq.Option{
		asynq.Queue(queueAgent),
		asynq.Retention(24 * time.Hour),
		asynq.TaskID(payload.IdempotencyKey),
		asynq.Unique(6 * time.Hour),
		asynq.Timeout(defaultGitWorktreeFlowTimeout),
	}
	opts = append(opts, options...)
	return task, opts, nil
}

func isWorktreePathUnderAppRoot(worktreePath string) bool {
	cleanPath := strings.TrimSpace(worktreePath)
	if cleanPath == "" || cleanPath == "." || cleanPath == ".." || strings.HasPrefix(cleanPath, "/") || strings.HasPrefix(cleanPath, "../") {
		return false
	}
	marker := "/worktrees/"
	index := strings.Index(cleanPath, marker)
	if index <= 0 {
		return false
	}
	return index+len(marker) < len(cleanPath)
}
