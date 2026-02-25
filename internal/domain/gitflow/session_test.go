package gitflow

import "testing"

func TestTaskExecutionSessionValidateBasics(t *testing.T) {
	session := &TaskExecutionSession{
		RunID:        "run-1",
		TaskID:       "task-1",
		SourceBranch: "revamp",
		TaskBranch:   "task/run-1/task-1",
		WorktreePath: ".worktree/worktrees/run-1-task-1",
	}
	session.Normalize()

	if err := session.ValidateBasics(); err != nil {
		t.Fatalf("expected valid session, got error: %v", err)
	}
}

func TestTaskExecutionSessionValidateBasicsRejectsInvalidWorktreePath(t *testing.T) {
	session := &TaskExecutionSession{
		RunID:        "run-1",
		TaskID:       "task-1",
		SourceBranch: "revamp",
		TaskBranch:   "task/run-1/task-1",
		WorktreePath: "tmp/worktree/run-1-task-1",
	}
	session.Normalize()

	if err := session.ValidateBasics(); err == nil {
		t.Fatalf("expected validation error for invalid worktree path")
	}
}

func TestEnsureMergeTarget(t *testing.T) {
	if err := EnsureMergeTarget("revamp", "revamp"); err != nil {
		t.Fatalf("expected no merge target error, got: %v", err)
	}

	if err := EnsureMergeTarget("revamp", "main"); err == nil {
		t.Fatalf("expected merge target mismatch error")
	}
}
