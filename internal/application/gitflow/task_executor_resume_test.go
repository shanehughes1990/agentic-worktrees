package gitflow

import (
	"context"
	"fmt"
	"strings"
	"testing"

	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
)

type resumeTestGitPort struct {
	mergeAttempt MergeAttempt
	syncAttempt  *MergeAttempt
	validateCalls int
}

func (port *resumeTestGitPort) CreateTaskWorktree(context.Context, string, string, string, string) error {
	return nil
}

func (port *resumeTestGitPort) MergeTaskBranch(context.Context, string, string, string) (MergeAttempt, error) {
	return port.mergeAttempt, nil
}

func (port *resumeTestGitPort) SyncTaskBranchWithSource(context.Context, string, string, string, string) (MergeAttempt, error) {
	if port.syncAttempt != nil {
		return *port.syncAttempt, nil
	}
	return MergeAttempt{NoChanges: true}, nil
}

func (port *resumeTestGitPort) ResolveConflicts(context.Context, string, []string, string) error {
	return nil
}

func (port *resumeTestGitPort) StageAll(context.Context, string) error {
	return nil
}

func (port *resumeTestGitPort) Commit(context.Context, string, string) error {
	return nil
}

func (port *resumeTestGitPort) CleanupTaskWorktree(context.Context, string, string, string) error {
	return nil
}

func (port *resumeTestGitPort) CleanupRunArtifacts(context.Context, string, string) error {
	return nil
}

func (port *resumeTestGitPort) ValidateWorktree(context.Context, string) error {
	port.validateCalls++
	return nil
}

type resumeTestDecomposer struct {
	calls          int
	requests       []appcopilot.DecomposeRequest
	conflictError  error
	sessionFirst   string
	sessionSecond  string
}

func (decomposer *resumeTestDecomposer) Decompose(_ context.Context, request appcopilot.DecomposeRequest) (appcopilot.DecomposeResult, error) {
	decomposer.calls++
	decomposer.requests = append(decomposer.requests, request)
	if decomposer.calls == 1 {
		return appcopilot.DecomposeResult{SessionID: decomposer.sessionFirst}, nil
	}
	return appcopilot.DecomposeResult{SessionID: decomposer.sessionSecond}, decomposer.conflictError
}

func TestTaskExecutorUsesProvidedResumeSessionForFirstIteration(t *testing.T) {
	decomposer := &resumeTestDecomposer{
		sessionFirst:  "session-updated",
		sessionSecond: "session-updated",
	}
	executor := NewTaskExecutor(&resumeTestGitPort{mergeAttempt: MergeAttempt{NoChanges: true}}, decomposer)

	_, err := executor.ExecuteTask(context.Background(), TaskExecutionRequest{
		BoardID:         "board-1",
		RunID:           "run-1",
		TaskID:          "task-1",
		TaskTitle:       "title",
		TaskDetail:      "detail",
		ResumeSessionID: "session-prev",
		SourceBranch:    "main",
		RepositoryRoot:  ".",
	})
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if len(decomposer.requests) == 0 {
		t.Fatalf("expected at least one decomposer request")
	}
	if got := strings.TrimSpace(decomposer.requests[0].ResumeSessionID); got != "session-prev" {
		t.Fatalf("expected first decomposer call to use prior resume session, got %q", got)
	}
}

func TestTaskExecutorPreservesLatestSessionOnConflictDecomposeFailure(t *testing.T) {
	decomposer := &resumeTestDecomposer{
		conflictError: fmt.Errorf("conflict resolution interrupted"),
		sessionFirst:  "session-first",
		sessionSecond: "session-second",
	}
	executor := NewTaskExecutor(&resumeTestGitPort{mergeAttempt: MergeAttempt{ConflictFiles: []string{"a.go"}}}, decomposer)

	result, err := executor.ExecuteTask(context.Background(), TaskExecutionRequest{
		BoardID:        "board-1",
		RunID:          "run-1",
		TaskID:         "task-1",
		TaskTitle:      "title",
		TaskDetail:     "detail",
		SourceBranch:   "main",
		RepositoryRoot: ".",
	})
	if err == nil {
		t.Fatalf("expected execute error")
	}
	if strings.TrimSpace(result.ResumeSessionID) != "session-second" {
		t.Fatalf("expected resume session to persist latest conflict session, got %q", result.ResumeSessionID)
	}
	if strings.TrimSpace(result.Status) != "failed" {
		t.Fatalf("expected failed status, got %q", result.Status)
	}
}

func TestTaskExecutorRunsValidationWhenSourceSyncIntroducesChanges(t *testing.T) {
	decomposer := &resumeTestDecomposer{sessionFirst: "session-first", sessionSecond: "session-second"}
	gitPort := &resumeTestGitPort{mergeAttempt: MergeAttempt{NoChanges: true}, syncAttempt: &MergeAttempt{NoChanges: false}}
	executor := NewTaskExecutor(gitPort, decomposer)

	_, err := executor.ExecuteTask(context.Background(), TaskExecutionRequest{
		BoardID:        "board-1",
		RunID:          "run-1",
		TaskID:         "task-1",
		TaskTitle:      "title",
		TaskDetail:     "detail",
		SourceBranch:   "main",
		RepositoryRoot: ".",
	})
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if gitPort.validateCalls != 1 {
		t.Fatalf("expected one validation run after source sync merge, got %d", gitPort.validateCalls)
	}
}
