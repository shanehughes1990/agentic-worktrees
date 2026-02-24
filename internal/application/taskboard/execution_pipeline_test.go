package taskboard

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

type memoryRepo struct {
	board *domaintaskboard.Board
	err   error
}

func (repository *memoryRepo) ListBoardIDs(context.Context) ([]string, error) {
	if repository.err != nil {
		return nil, repository.err
	}
	if repository.board == nil {
		return []string{}, nil
	}
	return []string{repository.board.BoardID}, nil
}

func (repository *memoryRepo) GetByBoardID(context.Context, string) (*domaintaskboard.Board, error) {
	if repository.err != nil {
		return nil, repository.err
	}
	if repository.board == nil {
		return nil, nil
	}
	copied := *repository.board
	return &copied, nil
}

func (repository *memoryRepo) Save(_ context.Context, board *domaintaskboard.Board) error {
	if repository.err != nil {
		return repository.err
	}
	if board == nil {
		return nil
	}
	copied := *board
	repository.board = &copied
	return nil
}

type fakeTaskExecutor struct {
	err             error
	calls           int
	mu              sync.Mutex
	activeExecutors int
	maxConcurrency  int
	wait            time.Duration
}

func (executor *fakeTaskExecutor) ExecuteTask(context.Context, TaskExecutionRequest) error {
	executor.mu.Lock()
	executor.activeExecutors++
	if executor.activeExecutors > executor.maxConcurrency {
		executor.maxConcurrency = executor.activeExecutors
	}
	executor.mu.Unlock()

	if executor.wait > 0 {
		time.Sleep(executor.wait)
	}

	executor.mu.Lock()
	executor.activeExecutors--
	executor.calls++
	executor.mu.Unlock()
	return executor.err
}

type fakeExecutionDispatcher struct {
	taskID string
	err    error
}

func (dispatcher *fakeExecutionDispatcher) EnqueueTaskboardExecution(context.Context, StartExecutionRequest) (string, error) {
	if dispatcher.err != nil {
		return "", dispatcher.err
	}
	if dispatcher.taskID == "" {
		return "task-1", nil
	}
	return dispatcher.taskID, nil
}

func TestExecutionCommandServiceStart(t *testing.T) {
	service := NewExecutionCommandService(&fakeExecutionDispatcher{taskID: "pipeline-1"})
	taskID, err := service.Start(context.Background(), StartExecutionRequest{
		BoardID:        "board-1",
		SourceBranch:   "revamp",
		RepositoryRoot: ".",
	})
	if err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}
	if taskID != "pipeline-1" {
		t.Fatalf("unexpected task id: %s", taskID)
	}
}

func TestExecutionPipelineServiceExecuteBoard(t *testing.T) {
	now := time.Now().UTC()
	repository := &memoryRepo{board: &domaintaskboard.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "epic-1", BoardID: "board-1", Title: "Epic", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Task", Status: domaintaskboard.StatusNotStarted},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}

	taskboardService := NewService(repository)
	executor := &fakeTaskExecutor{}
	pipeline := NewExecutionPipelineService(taskboardService, executor, 2)

	err := pipeline.ExecuteBoard(context.Background(), "board-1", "revamp", ".")
	if err != nil {
		t.Fatalf("unexpected execute board error: %v", err)
	}
	if executor.calls != 1 {
		t.Fatalf("expected one executor call, got %d", executor.calls)
	}
	if repository.board == nil {
		t.Fatalf("expected persisted board")
	}
	if !repository.board.IsCompleted("task-1") {
		t.Fatalf("expected task to be marked completed")
	}
}

func TestExecutionPipelineServiceExecuteBoardMarksBlockedOnFailure(t *testing.T) {
	now := time.Now().UTC()
	repository := &memoryRepo{board: &domaintaskboard.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "epic-1", BoardID: "board-1", Title: "Epic", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Task", Status: domaintaskboard.StatusNotStarted},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}

	taskboardService := NewService(repository)
	pipeline := NewExecutionPipelineService(taskboardService, &fakeTaskExecutor{err: errors.New("fatal")}, 2)

	err := pipeline.ExecuteBoard(context.Background(), "board-1", "revamp", ".")
	if err == nil {
		t.Fatalf("expected pipeline error")
	}
	if repository.board == nil {
		t.Fatalf("expected persisted board")
	}

	task, taskErr := taskboardService.GetTaskByID(context.Background(), "board-1", "task-1")
	if taskErr != nil {
		t.Fatalf("unexpected get task error: %v", taskErr)
	}
	if task == nil || task.Status != domaintaskboard.StatusBlocked {
		t.Fatalf("expected task status blocked, got %#v", task)
	}
}

func TestExecutionPipelineServiceExecuteBoardUsesConcurrencyLimit(t *testing.T) {
	now := time.Now().UTC()
	repository := &memoryRepo{board: &domaintaskboard.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "epic-1", BoardID: "board-1", Title: "Epic", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{
				{WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Task 1", Status: domaintaskboard.StatusNotStarted}},
				{WorkItem: domaintaskboard.WorkItem{ID: "task-2", BoardID: "board-1", Title: "Task 2", Status: domaintaskboard.StatusNotStarted}},
				{WorkItem: domaintaskboard.WorkItem{ID: "task-3", BoardID: "board-1", Title: "Task 3", Status: domaintaskboard.StatusNotStarted}},
			},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}

	taskboardService := NewService(repository)
	executor := &fakeTaskExecutor{wait: 50 * time.Millisecond}
	pipeline := NewExecutionPipelineService(taskboardService, executor, 2)

	err := pipeline.ExecuteBoard(context.Background(), "board-1", "revamp", ".")
	if err != nil {
		t.Fatalf("unexpected execute board error: %v", err)
	}
	if executor.calls != 3 {
		t.Fatalf("expected 3 executor calls, got %d", executor.calls)
	}
	if executor.maxConcurrency > 2 {
		t.Fatalf("expected max concurrency <= 2, got %d", executor.maxConcurrency)
	}
	if executor.maxConcurrency < 2 {
		t.Fatalf("expected pipeline to run tasks concurrently with limit 2, got %d", executor.maxConcurrency)
	}
}
