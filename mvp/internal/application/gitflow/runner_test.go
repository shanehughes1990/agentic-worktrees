package gitflow

import (
	"context"
	"errors"
	"testing"

	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
)

type fakeGitPort struct {
	mergeAttempt MergeAttempt
	err          error
	cleanupErr   error
	resolved     bool
}

func (port *fakeGitPort) CreateTaskWorktree(context.Context, string, string, string, string) error {
	return port.err
}

func (port *fakeGitPort) MergeTaskBranch(context.Context, string, string, string) (MergeAttempt, error) {
	if port.err != nil {
		return MergeAttempt{}, port.err
	}
	return port.mergeAttempt, nil
}

func (port *fakeGitPort) InspectWorktreeSyncState(context.Context, string, string, string, string) (WorktreeSyncState, error) {
	return WorktreeSyncState{}, nil
}

func (port *fakeGitPort) SyncTaskBranchWithSource(context.Context, string, string, string, string) (MergeAttempt, error) {
	if port.err != nil {
		return MergeAttempt{}, port.err
	}
	return MergeAttempt{NoChanges: true}, nil
}

func (port *fakeGitPort) ResolveConflicts(context.Context, string, []string, string) error {
	if port.err != nil {
		return port.err
	}
	port.resolved = true
	return nil
}

func (port *fakeGitPort) ValidateWorktree(context.Context, string) error {
	return nil
}

func (port *fakeGitPort) Commit(context.Context, string, string) error {
	return nil
}

func (port *fakeGitPort) StageAll(context.Context, string) error {
	return nil
}

func (port *fakeGitPort) CleanupTaskWorktree(context.Context, string, string, string) error {
	return port.cleanupErr
}

func (port *fakeGitPort) CleanupRunArtifacts(context.Context, string, string) error {
	return nil
}

type fakeConflictDispatcher struct {
	lastJob *ConflictResolveJob
	err     error
}

func (dispatcher *fakeConflictDispatcher) EnqueueConflictResolution(_ context.Context, job ConflictResolveJob) (string, error) {
	if dispatcher.err != nil {
		return "", dispatcher.err
	}
	dispatcher.lastJob = &job
	return "conflict-job-1", nil
}

type fakeWorkflowRepository struct {
	lastWorkflow *apptaskboard.IngestionWorkflow
}

func (repository *fakeWorkflowRepository) GetWorkflow(context.Context, string) (*apptaskboard.IngestionWorkflow, error) {
	return nil, nil
}

func (repository *fakeWorkflowRepository) ListWorkflows(context.Context) ([]apptaskboard.IngestionWorkflow, error) {
	return nil, nil
}

func (repository *fakeWorkflowRepository) SaveWorkflow(_ context.Context, workflow *apptaskboard.IngestionWorkflow) error {
	repository.lastWorkflow = workflow
	return nil
}

func TestRunnerRunWorktreeFlowQueuesConflictResolution(t *testing.T) {
	workflowRepository := &fakeWorkflowRepository{}
	conflictDispatcher := &fakeConflictDispatcher{}
	runner := NewRunner(&fakeGitPort{mergeAttempt: MergeAttempt{ConflictFiles: []string{"main.go"}}}, conflictDispatcher, workflowRepository)

	err := runner.RunWorktreeFlow(context.Background(), WorktreeFlowJob{
		RunID:          "run-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   ".worktree/worktrees/run-1-task-1",
	})
	if err != nil {
		t.Fatalf("unexpected run worktree flow error: %v", err)
	}
	if conflictDispatcher.lastJob == nil {
		t.Fatalf("expected conflict resolution job to be queued")
	}
	if workflowRepository.lastWorkflow == nil {
		t.Fatalf("expected workflow to be updated")
	}
}

func TestRunnerRunWorktreeFlowClassifiesFailures(t *testing.T) {
	runner := NewRunner(&fakeGitPort{err: errors.New("bad git")}, &fakeConflictDispatcher{}, &fakeWorkflowRepository{})

	err := runner.RunWorktreeFlow(context.Background(), WorktreeFlowJob{
		RunID:          "run-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   ".worktree/worktrees/run-1-task-1",
	})
	if err == nil {
		t.Fatalf("expected classified failure")
	}
	if !IsTerminalFailure(err) {
		t.Fatalf("expected terminal failure classification")
	}
}

func TestRunnerRunConflictResolution(t *testing.T) {
	gitPort := &fakeGitPort{}
	runner := NewRunner(gitPort, &fakeConflictDispatcher{}, &fakeWorkflowRepository{})

	err := runner.RunConflictResolution(context.Background(), ConflictResolveJob{
		RunID:          "run-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   ".worktree/worktrees/run-1-task-1",
		ConflictFiles:  []string{"main.go"},
	}, "")
	if err != nil {
		t.Fatalf("unexpected conflict resolution error: %v", err)
	}
	if !gitPort.resolved {
		t.Fatalf("expected conflict resolution path to run")
	}
}
