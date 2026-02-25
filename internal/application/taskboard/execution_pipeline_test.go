package taskboard

import (
	"context"
	"errors"
	"strings"
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
	outcome         TaskExecutionOutcome
	lastRequest     TaskExecutionRequest
	calls           int
	mu              sync.Mutex
	activeExecutors int
	maxConcurrency  int
	wait            time.Duration
}

func (executor *fakeTaskExecutor) ExecuteTask(_ context.Context, request TaskExecutionRequest) (TaskExecutionOutcome, error) {
	executor.mu.Lock()
	executor.activeExecutors++
	if executor.activeExecutors > executor.maxConcurrency {
		executor.maxConcurrency = executor.activeExecutors
	}
	executor.lastRequest = request
	executor.mu.Unlock()

	if executor.wait > 0 {
		time.Sleep(executor.wait)
	}

	executor.mu.Lock()
	executor.activeExecutors--
	executor.calls++
	executor.mu.Unlock()
	if executor.outcome != (TaskExecutionOutcome{}) {
		return executor.outcome, executor.err
	}
	if executor.err != nil {
		return TaskExecutionOutcome{}, executor.err
	}
	return TaskExecutionOutcome{Status: "merged", Reason: "merged", TaskBranch: "task/test", Worktree: ".worktree/test"}, nil
}

type fakeExecutionDispatcher struct {
	taskID string
	err    error
}

type memoryWorkflowRepo struct {
	workflow *IngestionWorkflow
	err      error
}

func (repository *memoryWorkflowRepo) GetWorkflow(context.Context, string) (*IngestionWorkflow, error) {
	if repository.err != nil {
		return nil, repository.err
	}
	if repository.workflow == nil {
		return nil, nil
	}
	copied := *repository.workflow
	return &copied, nil
}

func (repository *memoryWorkflowRepo) ListWorkflows(context.Context) ([]IngestionWorkflow, error) {
	if repository.err != nil {
		return nil, repository.err
	}
	if repository.workflow == nil {
		return []IngestionWorkflow{}, nil
	}
	return []IngestionWorkflow{*repository.workflow}, nil
}

func (repository *memoryWorkflowRepo) SaveWorkflow(_ context.Context, workflow *IngestionWorkflow) error {
	if repository.err != nil {
		return repository.err
	}
	if workflow == nil {
		return nil
	}
	copied := *workflow
	repository.workflow = &copied
	return nil
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
		MaxTasks:       2,
	})
	if err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}
	if taskID != "pipeline-1" {
		t.Fatalf("unexpected task id: %s", taskID)
	}
}

func TestExecutionCommandServiceStartRejectsNegativeMaxTasks(t *testing.T) {
	service := NewExecutionCommandService(&fakeExecutionDispatcher{taskID: "pipeline-1"})
	_, err := service.Start(context.Background(), StartExecutionRequest{
		BoardID:        "board-1",
		SourceBranch:   "revamp",
		RepositoryRoot: ".",
		MaxTasks:       -1,
	})
	if err == nil {
		t.Fatalf("expected validation error for negative max tasks")
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
	workflowRepo := &memoryWorkflowRepo{}
	pipeline := NewExecutionPipelineService(taskboardService, executor, workflowRepo, 2)

	err := pipeline.ExecuteBoard(context.Background(), "board-1", "revamp", ".", 0)
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
	workflowRepo := &memoryWorkflowRepo{}
	pipeline := NewExecutionPipelineService(taskboardService, &fakeTaskExecutor{err: errors.New("fatal")}, workflowRepo, 2)

	err := pipeline.ExecuteBoard(context.Background(), "board-1", "revamp", ".", 0)
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
	workflowRepo := &memoryWorkflowRepo{}
	pipeline := NewExecutionPipelineService(taskboardService, executor, workflowRepo, 2)

	err := pipeline.ExecuteBoard(context.Background(), "board-1", "revamp", ".", 0)
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

func TestExecutionPipelineServiceExecuteBoardAppendsResumeStreamEvents(t *testing.T) {
	now := time.Now().UTC()
	repository := &memoryRepo{board: &domaintaskboard.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "epic-1", BoardID: "board-1", Title: "Epic", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Task 1", Status: domaintaskboard.StatusInProgress},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}

	taskboardService := NewService(repository)
	executor := &fakeTaskExecutor{}
	workflowRepo := &memoryWorkflowRepo{workflow: &IngestionWorkflow{RunID: "board-1", Stream: "{\"event\":\"prior\"}"}}
	pipeline := NewExecutionPipelineService(taskboardService, executor, workflowRepo, 2)

	err := pipeline.ExecuteBoard(context.Background(), "board-1", "revamp", ".", 0)
	if err != nil {
		t.Fatalf("unexpected execute board error: %v", err)
	}
	if workflowRepo.workflow == nil {
		t.Fatalf("expected workflow to be saved")
	}
	if workflowRepo.workflow.Status != WorkflowStatusCompleted {
		t.Fatalf("expected completed workflow status, got %s", workflowRepo.workflow.Status)
	}
	if !strings.Contains(workflowRepo.workflow.Stream, "\"event\":\"pipeline_start\"") {
		t.Fatalf("expected pipeline_start event in stream")
	}
	if !strings.Contains(workflowRepo.workflow.Stream, "\"event\":\"resume_requeue\"") {
		t.Fatalf("expected resume_requeue event in stream")
	}
	if !strings.Contains(workflowRepo.workflow.Stream, "\"event\":\"task_completed\"") {
		t.Fatalf("expected task_completed event in stream")
	}
}

func TestExecutionPipelineServiceExecuteBoardPropagatesResumeSessionToExecutor(t *testing.T) {
	now := time.Now().UTC()
	repository := &memoryRepo{board: &domaintaskboard.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "epic-1", BoardID: "board-1", Title: "Epic", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Task 1", Status: domaintaskboard.StatusNotStarted},
				Outcome:  &domaintaskboard.TaskOutcome{Status: "interrupted", ResumeSessionID: "session-123"},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}

	taskboardService := NewService(repository)
	executor := &fakeTaskExecutor{}
	pipeline := NewExecutionPipelineService(taskboardService, executor, &memoryWorkflowRepo{}, 1)

	err := pipeline.ExecuteBoard(context.Background(), "board-1", "revamp", ".", 0)
	if err != nil {
		t.Fatalf("unexpected execute board error: %v", err)
	}
	if executor.lastRequest.ResumeSessionID != "session-123" {
		t.Fatalf("expected resume session id session-123, got %q", executor.lastRequest.ResumeSessionID)
	}
}

func TestExecutionPipelineServiceExecuteBoardCancelsTasksAsResumable(t *testing.T) {
	now := time.Now().UTC()
	repository := &memoryRepo{board: &domaintaskboard.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "epic-1", BoardID: "board-1", Title: "Epic", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Task 1", Status: domaintaskboard.StatusNotStarted},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}

	taskboardService := NewService(repository)
	executor := &fakeTaskExecutor{
		err:     context.Canceled,
		outcome: TaskExecutionOutcome{ResumeSessionID: "session-987"},
	}
	pipeline := NewExecutionPipelineService(taskboardService, executor, &memoryWorkflowRepo{}, 1)

	err := pipeline.ExecuteBoard(context.Background(), "board-1", "revamp", ".", 0)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled error, got %v", err)
	}

	task, taskErr := taskboardService.GetTaskByID(context.Background(), "board-1", "task-1")
	if taskErr != nil {
		t.Fatalf("unexpected get task error: %v", taskErr)
	}
	if task == nil {
		t.Fatalf("expected task to exist")
	}
	if task.Status != domaintaskboard.StatusNotStarted {
		t.Fatalf("expected canceled task to be requeued to not-started, got %s", task.Status)
	}
	if task.Outcome == nil {
		t.Fatalf("expected canceled task outcome")
	}
	if task.Outcome.Status != "canceled" {
		t.Fatalf("expected canceled outcome status, got %s", task.Outcome.Status)
	}
	if task.Outcome.ResumeSessionID != "session-987" {
		t.Fatalf("expected resume session id session-987, got %q", task.Outcome.ResumeSessionID)
	}
}

func TestExecutionPipelineServiceExecuteBoardStopsAtMaxTasks(t *testing.T) {
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
	executor := &fakeTaskExecutor{}
	pipeline := NewExecutionPipelineService(taskboardService, executor, &memoryWorkflowRepo{}, 2)

	err := pipeline.ExecuteBoard(context.Background(), "board-1", "revamp", ".", 2)
	if err != nil {
		t.Fatalf("unexpected execute board error: %v", err)
	}
	if executor.calls != 2 {
		t.Fatalf("expected exactly 2 task executions, got %d", executor.calls)
	}

	completedCount := 0
	notStartedCount := 0
	for _, taskID := range []string{"task-1", "task-2", "task-3"} {
		task, taskErr := taskboardService.GetTaskByID(context.Background(), "board-1", taskID)
		if taskErr != nil {
			t.Fatalf("unexpected get %s error: %v", taskID, taskErr)
		}
		switch task.Status {
		case domaintaskboard.StatusCompleted:
			completedCount++
		case domaintaskboard.StatusNotStarted:
			notStartedCount++
		}
	}
	if completedCount != 2 {
		t.Fatalf("expected exactly 2 completed tasks, got %d", completedCount)
	}
	if notStartedCount != 1 {
		t.Fatalf("expected exactly 1 task left not-started, got %d", notStartedCount)
	}
}

func TestExecutionPipelineServiceExecuteBoardStopsAtMaxTasksAcrossMultipleRounds(t *testing.T) {
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
				{WorkItem: domaintaskboard.WorkItem{ID: "task-4", BoardID: "board-1", Title: "Task 4", Status: domaintaskboard.StatusNotStarted}},
				{WorkItem: domaintaskboard.WorkItem{ID: "task-5", BoardID: "board-1", Title: "Task 5", Status: domaintaskboard.StatusNotStarted}},
				{WorkItem: domaintaskboard.WorkItem{ID: "task-6", BoardID: "board-1", Title: "Task 6", Status: domaintaskboard.StatusNotStarted}},
			},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}

	taskboardService := NewService(repository)
	executor := &fakeTaskExecutor{}
	pipeline := NewExecutionPipelineService(taskboardService, executor, &memoryWorkflowRepo{}, 2)

	err := pipeline.ExecuteBoard(context.Background(), "board-1", "revamp", ".", 5)
	if err != nil {
		t.Fatalf("unexpected execute board error: %v", err)
	}
	if executor.calls != 5 {
		t.Fatalf("expected exactly 5 task executions, got %d", executor.calls)
	}

	completedCount := 0
	notStartedCount := 0
	for _, taskID := range []string{"task-1", "task-2", "task-3", "task-4", "task-5", "task-6"} {
		task, taskErr := taskboardService.GetTaskByID(context.Background(), "board-1", taskID)
		if taskErr != nil {
			t.Fatalf("unexpected get %s error: %v", taskID, taskErr)
		}
		switch task.Status {
		case domaintaskboard.StatusCompleted:
			completedCount++
		case domaintaskboard.StatusNotStarted:
			notStartedCount++
		}
	}
	if completedCount != 5 {
		t.Fatalf("expected exactly 5 completed tasks, got %d", completedCount)
	}
	if notStartedCount != 1 {
		t.Fatalf("expected exactly 1 task left not-started, got %d", notStartedCount)
	}
}
