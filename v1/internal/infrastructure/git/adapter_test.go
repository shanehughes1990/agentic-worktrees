package git

import "testing"

func TestIsMissingWorktreePathError(t *testing.T) {
	message := "fatal: cannot change to '/Volumes/External/workspace/projects/personal/agentic-worktrees/.worktree/worktrees/run-1-task-1': No such file or directory"
	if !isMissingWorktreePathError(message) {
		t.Fatalf("expected missing worktree path error to be detected")
	}
}

func TestIsMissingWorktreePathErrorIgnoresGenericMissingPath(t *testing.T) {
	message := "fatal: cannot change to '/tmp/somewhere': No such file or directory"
	if isMissingWorktreePathError(message) {
		t.Fatalf("expected non-worktree missing path not to be treated as transient worktree cleanup case")
	}
}

func TestIsRetryableIndexConflictError(t *testing.T) {
	message := "error: you need to resolve your current index first\ninternal/application/taskboard/ingestion_test.go: needs merge"
	if !isRetryableIndexConflictError(message) {
		t.Fatalf("expected index conflict message to be retryable")
	}
}

func TestIsRetryableIndexConflictErrorIgnoresOtherGitErrors(t *testing.T) {
	message := "fatal: not a git repository (or any of the parent directories): .git"
	if isRetryableIndexConflictError(message) {
		t.Fatalf("expected generic fatal git error not to be treated as retryable index conflict")
	}
}
