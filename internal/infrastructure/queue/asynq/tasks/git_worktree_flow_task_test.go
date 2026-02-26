package tasks

import (
	"encoding/json"
	"testing"
)

func TestNewGitWorktreeFlowTaskValidatesInput(t *testing.T) {
	if _, _, err := NewGitWorktreeFlowTask(GitWorktreeFlowPayload{}); err == nil {
		t.Fatalf("expected validation error")
	}

	_, _, err := NewGitWorktreeFlowTask(GitWorktreeFlowPayload{
		RunID:          "run-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   "tmp/worktree/run-1-task-1",
	})
	if err == nil {
		t.Fatalf("expected invalid worktree path validation error")
	}
}

func TestNewGitWorktreeFlowTaskBuildsTask(t *testing.T) {
	task, options, err := NewGitWorktreeFlowTask(GitWorktreeFlowPayload{
		RunID:          "run-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   ".worktree/worktrees/run-1-task-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Type() != TaskTypeGitWorktreeFlow {
		t.Fatalf("unexpected task type: %s", task.Type())
	}
	if len(options) == 0 {
		t.Fatalf("expected default queue option")
	}
	payload := GitWorktreeFlowPayload{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("unexpected payload decode error: %v", err)
	}
	if payload.IdempotencyKey == "" {
		t.Fatalf("expected idempotency key fallback to be populated")
	}
}
