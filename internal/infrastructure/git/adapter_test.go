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
