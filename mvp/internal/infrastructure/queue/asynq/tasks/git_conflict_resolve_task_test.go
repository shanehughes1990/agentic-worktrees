package tasks

import (
	"encoding/json"
	"testing"
)

func TestNewGitConflictResolveTaskValidatesInput(t *testing.T) {
	if _, _, err := NewGitConflictResolveTask(GitConflictResolvePayload{}); err == nil {
		t.Fatalf("expected validation error")
	}

	_, _, err := NewGitConflictResolveTask(GitConflictResolvePayload{
		RunID:          "run-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   ".worktree/worktrees/run-1-task-1",
		ConflictFiles:  nil,
	})
	if err == nil {
		t.Fatalf("expected conflict_files validation error")
	}
}

func TestNewGitConflictResolveTaskBuildsTask(t *testing.T) {
	task, options, err := NewGitConflictResolveTask(GitConflictResolvePayload{
		RunID:          "run-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   ".worktree/worktrees/run-1-task-1",
		ConflictFiles:  []string{"main.go"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Type() != TaskTypeGitConflictResolve {
		t.Fatalf("unexpected task type: %s", task.Type())
	}
	if len(options) == 0 {
		t.Fatalf("expected default queue option")
	}
	payload := GitConflictResolvePayload{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("unexpected payload decode error: %v", err)
	}
	if payload.IdempotencyKey == "" {
		t.Fatalf("expected idempotency key fallback to be populated")
	}
}
