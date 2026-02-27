package gitflow

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type MergeState string

const (
	MergeStatePending    MergeState = "pending"
	MergeStateConflicted MergeState = "conflicted"
	MergeStateMerged     MergeState = "merged"
	MergeStateFailed     MergeState = "failed"
)

type ConflictState string

const (
	ConflictStateNone     ConflictState = "none"
	ConflictStatePending  ConflictState = "pending"
	ConflictStateResolved ConflictState = "resolved"
	ConflictStateFailed   ConflictState = "failed"
)

type TaskExecutionSession struct {
	RunID        string        `json:"run_id"`
	TaskID       string        `json:"task_id"`
	SourceBranch string        `json:"source_branch"`
	TaskBranch   string        `json:"task_branch"`
	WorktreePath string        `json:"worktree_path"`
	MergeState   MergeState    `json:"merge_state"`
	Conflict     ConflictState `json:"conflict_state"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

func (session *TaskExecutionSession) Normalize() {
	now := time.Now().UTC()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now

	session.RunID = strings.TrimSpace(session.RunID)
	session.TaskID = strings.TrimSpace(session.TaskID)
	session.SourceBranch = strings.TrimSpace(session.SourceBranch)
	session.TaskBranch = strings.TrimSpace(session.TaskBranch)
	session.WorktreePath = strings.TrimSpace(session.WorktreePath)

	if session.MergeState == "" {
		session.MergeState = MergeStatePending
	}
	if session.Conflict == "" {
		session.Conflict = ConflictStateNone
	}
}

func (session *TaskExecutionSession) ValidateBasics() error {
	if session.RunID == "" {
		return fmt.Errorf("run_id is required")
	}
	if session.TaskID == "" {
		return fmt.Errorf("task_id is required")
	}
	if session.SourceBranch == "" {
		return fmt.Errorf("source_branch is required")
	}
	if session.TaskBranch == "" {
		return fmt.Errorf("task_branch is required")
	}
	if session.TaskBranch == session.SourceBranch {
		return fmt.Errorf("task_branch must differ from source_branch")
	}
	if err := ValidateWorktreePath(session.WorktreePath); err != nil {
		return err
	}
	return nil
}

func EnsureMergeTarget(capturedSourceBranch, mergeTargetBranch string) error {
	cleanSource := strings.TrimSpace(capturedSourceBranch)
	cleanTarget := strings.TrimSpace(mergeTargetBranch)
	if cleanSource == "" {
		return fmt.Errorf("captured source branch is required")
	}
	if cleanTarget == "" {
		return fmt.Errorf("merge target branch is required")
	}
	if cleanSource != cleanTarget {
		return fmt.Errorf("invalid merge target: expected %s, got %s", cleanSource, cleanTarget)
	}
	return nil
}

func ValidateWorktreePath(path string) error {
	cleanPath := filepath.ToSlash(strings.TrimSpace(path))
	if cleanPath == "" {
		return fmt.Errorf("worktree_path is required")
	}
	if strings.HasPrefix(cleanPath, "/") {
		return fmt.Errorf("worktree_path must be repo-relative under <app_root>/worktrees")
	}
	if isWorktreePathUnderAppRoot(cleanPath) {
		return nil
	}
	return fmt.Errorf("worktree_path must be under <app_root>/worktrees: %s", cleanPath)
}

func isWorktreePathUnderAppRoot(worktreePath string) bool {
	cleanPath := filepath.ToSlash(filepath.Clean(strings.TrimSpace(worktreePath)))
	if cleanPath == "" || cleanPath == "." || cleanPath == ".." || strings.HasPrefix(cleanPath, "../") {
		return false
	}
	marker := "/worktrees/"
	index := strings.Index(cleanPath, marker)
	if index <= 0 {
		return false
	}
	return index+len(marker) < len(cleanPath)
}
