package gitflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
)

type resumeTestGitPort struct {
	mergeAttempt MergeAttempt
	syncAttempt  *MergeAttempt
	validateCalls int
	syncState WorktreeSyncState
	inspectSyncErr error
	inspectSyncCalls int
	cleanupCalls int
	cleanupErr error
}

func (port *resumeTestGitPort) CreateTaskWorktree(context.Context, string, string, string, string) error {
	return nil
}

func (port *resumeTestGitPort) MergeTaskBranch(context.Context, string, string, string) (MergeAttempt, error) {
	return port.mergeAttempt, nil
}

func (port *resumeTestGitPort) InspectWorktreeSyncState(context.Context, string, string, string, string) (WorktreeSyncState, error) {
	port.inspectSyncCalls++
	if port.inspectSyncErr != nil {
		return WorktreeSyncState{}, port.inspectSyncErr
	}
	return port.syncState, nil
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
	port.cleanupCalls++
	return port.cleanupErr
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

type failingFirstCallDecomposer struct {
	err error
}

func (decomposer *failingFirstCallDecomposer) Decompose(_ context.Context, _ appcopilot.DecomposeRequest) (appcopilot.DecomposeResult, error) {
	return appcopilot.DecomposeResult{}, decomposer.err
}

type resumeCheckpointRecorder struct {
	boardID   string
	taskID    string
	sessions  []string
}

func (recorder *resumeCheckpointRecorder) CheckpointResumeSession(_ context.Context, boardID string, taskID string, resumeSessionID string) error {
	recorder.boardID = boardID
	recorder.taskID = taskID
	recorder.sessions = append(recorder.sessions, strings.TrimSpace(resumeSessionID))
	return nil
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

func TestTaskExecutorResumedTaskRunsSourceSyncConflictResolutionBeforeImplementation(t *testing.T) {
	decomposer := &resumeTestDecomposer{
		sessionFirst:  "session-preflight",
		sessionSecond: "session-work",
	}
	gitPort := &resumeTestGitPort{
		mergeAttempt: MergeAttempt{NoChanges: true},
		syncAttempt:  &MergeAttempt{ConflictFiles: []string{"README.md"}},
	}
	executor := NewTaskExecutor(gitPort, decomposer)

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
		t.Fatalf("expected decomposer requests")
	}
	if !strings.Contains(decomposer.requests[0].Prompt, "You are synchronizing task") {
		t.Fatalf("expected first decomposer call to handle source sync conflicts, got prompt %q", decomposer.requests[0].Prompt)
	}
	if got := strings.TrimSpace(decomposer.requests[0].ResumeSessionID); got != "session-prev" {
		t.Fatalf("expected first decomposer call to use prior resume session, got %q", got)
	}
}

func TestTaskExecutorExistingWorktreeWithoutResumeSessionRunsPreflightFirst(t *testing.T) {
	tempDirectory := t.TempDir()
	worktreePath := filepath.Join(tempDirectory, ".worktree", "run-1-task-1")
	if err := os.MkdirAll(worktreePath, 0o755); err != nil {
		t.Fatalf("create worktree path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worktreePath, ".git"), []byte("gitdir: /tmp/fake\n"), 0o644); err != nil {
		t.Fatalf("create worktree .git marker: %v", err)
	}

	decomposer := &resumeTestDecomposer{sessionFirst: "session-preflight", sessionSecond: "session-work"}
	gitPort := &resumeTestGitPort{mergeAttempt: MergeAttempt{NoChanges: true}, syncAttempt: &MergeAttempt{ConflictFiles: []string{"README.md"}}}
	executor := NewTaskExecutor(gitPort, decomposer)

	_, err := executor.ExecuteTask(context.Background(), TaskExecutionRequest{
		BoardID:        "board-1",
		RunID:          "run-1",
		TaskID:         "task-1",
		TaskTitle:      "title",
		TaskDetail:     "detail",
		SourceBranch:   "main",
		RepositoryRoot: tempDirectory,
		WorktreePath:   ".worktree/run-1-task-1",
	})
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if len(decomposer.requests) == 0 {
		t.Fatalf("expected decomposer requests")
	}
	if !strings.Contains(decomposer.requests[0].Prompt, "You are synchronizing task") {
		t.Fatalf("expected first decomposer call to handle source sync conflicts, got prompt %q", decomposer.requests[0].Prompt)
	}
	if got := strings.TrimSpace(decomposer.requests[0].ResumeSessionID); got != "" {
		t.Fatalf("expected empty resume session id on first call, got %q", got)
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

func TestTaskExecutorCheckpointsResumeSession(t *testing.T) {
	decomposer := &resumeTestDecomposer{sessionFirst: "session-first", sessionSecond: "session-second"}
	checkpoint := &resumeCheckpointRecorder{}
	executor := NewTaskExecutor(&resumeTestGitPort{mergeAttempt: MergeAttempt{NoChanges: true}}, decomposer, checkpoint)

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
	if checkpoint.boardID != "board-1" || checkpoint.taskID != "task-1" {
		t.Fatalf("unexpected checkpoint target board=%q task=%q", checkpoint.boardID, checkpoint.taskID)
	}
	if len(checkpoint.sessions) == 0 || checkpoint.sessions[0] != "session-first" {
		t.Fatalf("expected first checkpoint session to persist, got %#v", checkpoint.sessions)
	}
}

func TestTaskExecutorClassifiesStartupProbeKilledAsTransient(t *testing.T) {
	decomposer := &failingFirstCallDecomposer{err: fmt.Errorf("copilot preflight failed: copilot cli startup probe failed: signal: killed")}
	executor := NewTaskExecutor(&resumeTestGitPort{mergeAttempt: MergeAttempt{NoChanges: true}}, decomposer)

	_, err := executor.ExecuteTask(context.Background(), TaskExecutionRequest{
		BoardID:        "board-1",
		RunID:          "run-1",
		TaskID:         "task-1",
		TaskTitle:      "title",
		TaskDetail:     "detail",
		SourceBranch:   "main",
		RepositoryRoot: ".",
	})
	if err == nil {
		t.Fatalf("expected execution error")
	}
	if IsTerminalFailure(err) {
		t.Fatalf("expected transient classification for startup probe killed error")
	}
}

func TestTaskExecutorUsesRetryPromptOnSecondAttempt(t *testing.T) {
	decomposer := &resumeTestDecomposer{
		sessionFirst:  "session-updated",
		sessionSecond: "session-updated",
	}
	executor := NewTaskExecutor(&resumeTestGitPort{mergeAttempt: MergeAttempt{NoChanges: true}}, decomposer)

	_, err := executor.ExecuteTask(context.Background(), TaskExecutionRequest{
		BoardID:          "board-1",
		RunID:            "run-1",
		TaskID:           "task-1",
		TaskTitle:        "title",
		TaskDetail:       "detail",
		ExecutionAttempt: 2,
		SourceBranch:     "main",
		RepositoryRoot:   ".",
	})
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if len(decomposer.requests) == 0 {
		t.Fatalf("expected at least one decomposer request")
	}
	if !strings.Contains(decomposer.requests[0].Prompt, "Previous attempt made no forward progress") {
		t.Fatalf("expected retry prompt content, got %q", decomposer.requests[0].Prompt)
	}
}

func TestTaskExecutorReconcileCompletedTaskWorktreeCleansOnlyWhenDriftFree(t *testing.T) {
	tempDirectory := t.TempDir()
	worktreePath := filepath.Join(tempDirectory, ".worktree", "run-1-task-1")
	if err := os.MkdirAll(worktreePath, 0o755); err != nil {
		t.Fatalf("create worktree path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worktreePath, ".git"), []byte("gitdir: /tmp/fake\n"), 0o644); err != nil {
		t.Fatalf("create worktree .git marker: %v", err)
	}

	gitPort := &resumeTestGitPort{syncState: WorktreeSyncState{HasUncommittedChanges: false, AheadFileCount: 0, BehindFileCount: 0}}
	executor := NewTaskExecutor(gitPort, nil)

	err := executor.ReconcileCompletedTaskWorktree(context.Background(), TaskExecutionRequest{
		BoardID:        "board-1",
		RunID:          "run-1",
		TaskID:         "task-1",
		TaskBranch:     "task/run-1/task-1",
		SourceBranch:   "main",
		RepositoryRoot: tempDirectory,
		WorktreePath:   ".worktree/run-1-task-1",
	})
	if err != nil {
		t.Fatalf("unexpected reconcile error: %v", err)
	}
	if gitPort.inspectSyncCalls != 1 {
		t.Fatalf("expected one sync inspection call, got %d", gitPort.inspectSyncCalls)
	}
	if gitPort.cleanupCalls != 1 {
		t.Fatalf("expected one cleanup call, got %d", gitPort.cleanupCalls)
	}
}

func TestTaskExecutorReconcileCompletedTaskWorktreeFailsOnDrift(t *testing.T) {
	tempDirectory := t.TempDir()
	worktreePath := filepath.Join(tempDirectory, ".worktree", "run-1-task-1")
	if err := os.MkdirAll(worktreePath, 0o755); err != nil {
		t.Fatalf("create worktree path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worktreePath, ".git"), []byte("gitdir: /tmp/fake\n"), 0o644); err != nil {
		t.Fatalf("create worktree .git marker: %v", err)
	}

	gitPort := &resumeTestGitPort{syncState: WorktreeSyncState{HasUncommittedChanges: true, AheadFileCount: 1, BehindFileCount: 0}}
	executor := NewTaskExecutor(gitPort, nil)

	err := executor.ReconcileCompletedTaskWorktree(context.Background(), TaskExecutionRequest{
		BoardID:        "board-1",
		RunID:          "run-1",
		TaskID:         "task-1",
		TaskBranch:     "task/run-1/task-1",
		SourceBranch:   "main",
		RepositoryRoot: tempDirectory,
		WorktreePath:   ".worktree/run-1-task-1",
	})
	if err == nil {
		t.Fatalf("expected reconcile drift error")
	}
	if gitPort.cleanupCalls != 0 {
		t.Fatalf("expected no cleanup on drift, got %d", gitPort.cleanupCalls)
	}
}
