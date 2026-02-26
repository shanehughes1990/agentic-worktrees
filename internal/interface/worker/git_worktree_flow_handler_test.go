package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
)

type memoryBoardRepoForWorker struct {
	board *domaintaskboard.Board
}

func (repository *memoryBoardRepoForWorker) ListBoardIDs(context.Context) ([]string, error) {
	if repository.board == nil {
		return []string{}, nil
	}
	return []string{repository.board.BoardID}, nil
}

func (repository *memoryBoardRepoForWorker) GetByBoardID(context.Context, string) (*domaintaskboard.Board, error) {
	if repository.board == nil {
		return nil, nil
	}
	copied := *repository.board
	return &copied, nil
}

func (repository *memoryBoardRepoForWorker) Save(_ context.Context, board *domaintaskboard.Board) error {
	if board == nil {
		return nil
	}
	copied := *board
	repository.board = &copied
	return nil
}

type fakeTaskExecutorForWorker struct {
	result      appgitflow.TaskExecutionResult
	err         error
	lastRequest appgitflow.TaskExecutionRequest
}

func (executor *fakeTaskExecutorForWorker) ExecuteTask(_ context.Context, request appgitflow.TaskExecutionRequest) (appgitflow.TaskExecutionResult, error) {
	executor.lastRequest = request
	return executor.result, executor.err
}

type fakeGitPortForWorker struct{}

func (port *fakeGitPortForWorker) CreateTaskWorktree(context.Context, string, string, string, string) error {
	return nil
}
func (port *fakeGitPortForWorker) MergeTaskBranch(context.Context, string, string, string) (appgitflow.MergeAttempt, error) {
	return appgitflow.MergeAttempt{}, nil
}
func (port *fakeGitPortForWorker) SyncTaskBranchWithSource(context.Context, string, string, string, string) (appgitflow.MergeAttempt, error) {
	return appgitflow.MergeAttempt{NoChanges: true}, nil
}
func (port *fakeGitPortForWorker) ResolveConflicts(context.Context, string, []string, string) error {
	return nil
}
func (port *fakeGitPortForWorker) ValidateWorktree(context.Context, string) error {
	return nil
}
func (port *fakeGitPortForWorker) Commit(context.Context, string, string) error {
	return nil
}
func (port *fakeGitPortForWorker) StageAll(context.Context, string) error {
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
func (port *failingGitPortForWorker) SyncTaskBranchWithSource(context.Context, string, string, string, string) (appgitflow.MergeAttempt, error) {
	return appgitflow.MergeAttempt{}, appgitflow.WrapTerminal(errors.New("fatal git error"))
}
func (port *failingGitPortForWorker) ResolveConflicts(context.Context, string, []string, string) error {
	return nil
}
func (port *failingGitPortForWorker) ValidateWorktree(context.Context, string) error {
	return appgitflow.WrapTerminal(errors.New("fatal git error"))
}
func (port *failingGitPortForWorker) Commit(context.Context, string, string) error {
	return nil
}
func (port *failingGitPortForWorker) StageAll(context.Context, string) error {
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
	now := time.Now().UTC()
	repository := &memoryBoardRepoForWorker{board: &domaintaskboard.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "epic-1", BoardID: "board-1", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Status: domaintaskboard.StatusInProgress},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}
	taskboardService := apptaskboard.NewService(repository)
	handler := NewGitWorktreeFlowHandler(&fakeTaskExecutorForWorker{result: appgitflow.TaskExecutionResult{Status: "merged", Reason: "ok", TaskBranch: "task/run-1/task-1", Worktree: ".worktree/worktrees/run-1-task-1"}}, taskboardService, nil)

	task, _, err := tasks.NewGitWorktreeFlowTask(tasks.GitWorktreeFlowPayload{
		RunID:          "run-1",
		BoardID:        "board-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   ".worktree/worktrees/run-1-task-1",
	})
	if err != nil {
		t.Fatalf("unexpected task build error: %v", err)
	}

	if err := handler.ProcessTask(context.Background(), task); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	updatedTask, getErr := taskboardService.GetTaskByID(context.Background(), "board-1", "task-1")
	if getErr != nil {
		t.Fatalf("unexpected get task error: %v", getErr)
	}
	if updatedTask == nil || updatedTask.Status != domaintaskboard.StatusCompleted {
		t.Fatalf("expected completed task, got %#v", updatedTask)
	}
}

func TestGitWorktreeFlowHandlerSkipsRetryOnTerminalFailure(t *testing.T) {
	now := time.Now().UTC()
	repository := &memoryBoardRepoForWorker{board: &domaintaskboard.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "epic-1", BoardID: "board-1", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Status: domaintaskboard.StatusInProgress},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}
	taskboardService := apptaskboard.NewService(repository)
	handler := NewGitWorktreeFlowHandler(&fakeTaskExecutorForWorker{err: appgitflow.WrapTerminal(errors.New("fatal git error"))}, taskboardService, nil)

	task, _, err := tasks.NewGitWorktreeFlowTask(tasks.GitWorktreeFlowPayload{
		RunID:          "run-1",
		BoardID:        "board-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   ".worktree/worktrees/run-1-task-1",
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
	updatedTask, getErr := taskboardService.GetTaskByID(context.Background(), "board-1", "task-1")
	if getErr != nil {
		t.Fatalf("unexpected get task error: %v", getErr)
	}
	if updatedTask == nil || updatedTask.Status != domaintaskboard.StatusBlocked {
		t.Fatalf("expected blocked task, got %#v", updatedTask)
	}
}

func TestGitWorktreeFlowHandlerRequeuesOnTransientFailure(t *testing.T) {
	now := time.Now().UTC()
	repository := &memoryBoardRepoForWorker{board: &domaintaskboard.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "epic-1", BoardID: "board-1", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Status: domaintaskboard.StatusInProgress},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}
	taskboardService := apptaskboard.NewService(repository)
	handler := NewGitWorktreeFlowHandler(&fakeTaskExecutorForWorker{err: appgitflow.WrapTransient(errors.New("worktree missing after cleanup"))}, taskboardService, nil)

	task, _, err := tasks.NewGitWorktreeFlowTask(tasks.GitWorktreeFlowPayload{
		RunID:          "run-1",
		BoardID:        "board-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   ".worktree/worktrees/run-1-task-1",
	})
	if err != nil {
		t.Fatalf("unexpected task build error: %v", err)
	}

	err = handler.ProcessTask(context.Background(), task)
	if err == nil {
		t.Fatalf("expected handler error")
	}
	if errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected retryable transient error, got skip retry: %v", err)
	}
	updatedTask, getErr := taskboardService.GetTaskByID(context.Background(), "board-1", "task-1")
	if getErr != nil {
		t.Fatalf("unexpected get task error: %v", getErr)
	}
	if updatedTask == nil || updatedTask.Status != domaintaskboard.StatusNotStarted {
		t.Fatalf("expected task requeued to not-started, got %#v", updatedTask)
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
		WorktreePath:   ".worktree/worktrees/run-1-task-1",
		ConflictFiles:  []string{"main.go"},
	})
	if err != nil {
		t.Fatalf("unexpected task build error: %v", err)
	}

	if err := handler.ProcessTask(context.Background(), task); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
}

func TestGitWorktreeFlowHandlerReturnsRetryableErrorOnInterruption(t *testing.T) {
	now := time.Now().UTC()
	repository := &memoryBoardRepoForWorker{board: &domaintaskboard.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "epic-1", BoardID: "board-1", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Status: domaintaskboard.StatusInProgress},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}
	taskboardService := apptaskboard.NewService(repository)
	handler := NewGitWorktreeFlowHandler(&fakeTaskExecutorForWorker{err: context.Canceled}, taskboardService, nil)

	task, _, err := tasks.NewGitWorktreeFlowTask(tasks.GitWorktreeFlowPayload{
		RunID:          "run-1",
		BoardID:        "board-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
		TaskBranch:     "task/run-1/task-1",
		WorktreePath:   ".worktree/worktrees/run-1-task-1",
	})
	if err != nil {
		t.Fatalf("unexpected task build error: %v", err)
	}

	err = handler.ProcessTask(context.Background(), task)
	if err == nil {
		t.Fatalf("expected interruption error")
	}
	if errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected retryable interruption error, got skip retry: %v", err)
	}
}

func TestGitWorktreeFlowHandlerPrefersPersistedResumeSessionID(t *testing.T) {
	now := time.Now().UTC()
	repository := &memoryBoardRepoForWorker{board: &domaintaskboard.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "epic-1", BoardID: "board-1", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Status: domaintaskboard.StatusInProgress},
				Outcome:  &domaintaskboard.TaskOutcome{ResumeSessionID: "session-persisted"},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}
	taskboardService := apptaskboard.NewService(repository)
	fakeExecutor := &fakeTaskExecutorForWorker{result: appgitflow.TaskExecutionResult{Status: "merged", Reason: "ok"}}
	handler := NewGitWorktreeFlowHandler(fakeExecutor, taskboardService, nil)

	task, _, err := tasks.NewGitWorktreeFlowTask(tasks.GitWorktreeFlowPayload{
		RunID:           "run-1",
		BoardID:         "board-1",
		TaskID:          "task-1",
		ResumeSessionID: "session-stale",
		RepositoryRoot:  ".",
		SourceBranch:    "revamp",
		TaskBranch:      "task/run-1/task-1",
		WorktreePath:    ".worktree/worktrees/run-1-task-1",
	})
	if err != nil {
		t.Fatalf("unexpected task build error: %v", err)
	}

	if err := handler.ProcessTask(context.Background(), task); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	if got := fakeExecutor.lastRequest.ResumeSessionID; got != "session-persisted" {
		t.Fatalf("expected persisted resume session id, got %q", got)
	}
}
