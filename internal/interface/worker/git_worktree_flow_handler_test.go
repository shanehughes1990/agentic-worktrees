package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
)

type fakeGitPortForWorker struct{}

func (port *fakeGitPortForWorker) CreateTaskWorktree(context.Context, string, string, string, string) error {
	return nil
}
func (port *fakeGitPortForWorker) MergeTaskBranch(context.Context, string, string, string) (appgitflow.MergeAttempt, error) {
	return appgitflow.MergeAttempt{}, nil
}
func (port *fakeGitPortForWorker) ResolveConflicts(context.Context, string, []string, string) error {
	return nil
}
func (port *fakeGitPortForWorker) Commit(context.Context, string, string) error {
	return nil
}
func (port *fakeGitPortForWorker) CleanupTaskWorktree(context.Context, string, string, string) error {
	return nil
}
func (port *fakeGitPortForWorker) CleanupRunArtifacts(context.Context, string, string) error {
	return nil
}

type fakeConflictDispatcherForWorker struct{}

func (dispatcher *fakeConflictDispatcherForWorker) EnqueueConflictResolution(context.Context, appgitflow.ConflictResolveJob) (string, error) {
	return "conflict-job", nil
}

type failingGitPortForWorker struct{}

func (port *failingGitPortForWorker) CreateTaskWorktree(context.Context, string, string, string, string) error {
	return appgitflow.WrapTerminal(errors.New("fatal git error"))
}
func (port *failingGitPortForWorker) MergeTaskBranch(context.Context, string, string, string) (appgitflow.MergeAttempt, error) {
	return appgitflow.MergeAttempt{}, nil
}
func (port *failingGitPortForWorker) ResolveConflicts(context.Context, string, []string, string) error {
	return nil
}
func (port *failingGitPortForWorker) Commit(context.Context, string, string) error {
	return nil
}
func (port *failingGitPortForWorker) CleanupTaskWorktree(context.Context, string, string, string) error {
	return nil
}
func (port *failingGitPortForWorker) CleanupRunArtifacts(context.Context, string, string) error {
	return nil
}

type fakeWorkflowRepoForWorker struct{}

func (repository *fakeWorkflowRepoForWorker) GetWorkflow(context.Context, string) (*apptaskboard.IngestionWorkflow, error) {
	return nil, nil
}
func (repository *fakeWorkflowRepoForWorker) ListWorkflows(context.Context) ([]apptaskboard.IngestionWorkflow, error) {
	return nil, nil
}
func (repository *fakeWorkflowRepoForWorker) SaveWorkflow(context.Context, *apptaskboard.IngestionWorkflow) error {
	return nil
}

type fakeCopilotDecomposerForWorker struct{}

func (decomposer *fakeCopilotDecomposerForWorker) Decompose(context.Context, appcopilot.DecomposeRequest) (appcopilot.DecomposeResult, error) {
	return appcopilot.DecomposeResult{Response: "resolve with minimal change"}, nil
}

func TestGitWorktreeFlowHandlerProcessTask(t *testing.T) {
	runner := appgitflow.NewRunner(&fakeGitPortForWorker{}, &fakeConflictDispatcherForWorker{}, &fakeWorkflowRepoForWorker{})
	handler := NewGitWorktreeFlowHandler(runner, nil)

	task, _, err := tasks.NewGitWorktreeFlowTask(tasks.GitWorktreeFlowPayload{
		RunID:          "run-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   ".worktree/run-1-task-1",
	})
	if err != nil {
		t.Fatalf("unexpected task build error: %v", err)
	}

	if err := handler.ProcessTask(context.Background(), task); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
}

func TestGitWorktreeFlowHandlerSkipsRetryOnTerminalFailure(t *testing.T) {
	runner := appgitflow.NewRunner(&failingGitPortForWorker{}, &fakeConflictDispatcherForWorker{}, &fakeWorkflowRepoForWorker{})
	handler := NewGitWorktreeFlowHandler(runner, nil)

	task, _, err := tasks.NewGitWorktreeFlowTask(tasks.GitWorktreeFlowPayload{
		RunID:          "run-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   ".worktree/run-1-task-1",
	})
	if err != nil {
		t.Fatalf("unexpected task build error: %v", err)
	}

	err = handler.ProcessTask(context.Background(), task)
	if err == nil {
		t.Fatalf("expected handler error")
	}
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected skip retry error, got: %v", err)
	}
}

func TestGitConflictResolveHandlerProcessTask(t *testing.T) {
	runner := appgitflow.NewRunner(&fakeGitPortForWorker{}, &fakeConflictDispatcherForWorker{}, &fakeWorkflowRepoForWorker{})
	handler := NewGitConflictResolveHandler(runner, &fakeCopilotDecomposerForWorker{}, nil)

	task, _, err := tasks.NewGitConflictResolveTask(tasks.GitConflictResolvePayload{
		RunID:          "run-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   ".worktree/run-1-task-1",
		ConflictFiles:  []string{"main.go"},
	})
	if err != nil {
		t.Fatalf("unexpected task build error: %v", err)
	}

	if err := handler.ProcessTask(context.Background(), task); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
}
