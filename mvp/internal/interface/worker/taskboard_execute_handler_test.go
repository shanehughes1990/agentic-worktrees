package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
)

type executeHandlerBoardRepo struct {
	board *domaintaskboard.Board
}

func (repository *executeHandlerBoardRepo) ListBoardIDs(context.Context) ([]string, error) {
	if repository.board == nil {
		return []string{}, nil
	}
	return []string{repository.board.BoardID}, nil
}

func (repository *executeHandlerBoardRepo) GetByBoardID(context.Context, string) (*domaintaskboard.Board, error) {
	if repository.board == nil {
		return nil, nil
	}
	copied := *repository.board
	return &copied, nil
}

func (repository *executeHandlerBoardRepo) Save(_ context.Context, board *domaintaskboard.Board) error {
	if board == nil {
		return nil
	}
	copied := *board
	repository.board = &copied
	return nil
}

type executeHandlerWorkflowRepo struct{}

func (repository *executeHandlerWorkflowRepo) GetWorkflow(context.Context, string) (*apptaskboard.IngestionWorkflow, error) {
	return nil, nil
}

func (repository *executeHandlerWorkflowRepo) ListWorkflows(context.Context) ([]apptaskboard.IngestionWorkflow, error) {
	return []apptaskboard.IngestionWorkflow{}, nil
}

func (repository *executeHandlerWorkflowRepo) SaveWorkflow(context.Context, *apptaskboard.IngestionWorkflow) error {
	return nil
}

type executeHandlerNoopExecutor struct{}

func (executor *executeHandlerNoopExecutor) ExecuteTask(context.Context, apptaskboard.TaskExecutionRequest) (apptaskboard.TaskExecutionOutcome, error) {
	return apptaskboard.TaskExecutionOutcome{}, appgitflow.WrapTerminal(errors.New("not expected in cancellation test"))
}

func TestTaskboardExecuteHandlerReturnsRetryableErrorOnInterruption(t *testing.T) {
	now := time.Now().UTC()
	repo := &executeHandlerBoardRepo{board: &domaintaskboard.Board{
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

	taskboardService := apptaskboard.NewService(repo)
	pipeline := apptaskboard.NewExecutionPipelineService(taskboardService, &executeHandlerNoopExecutor{}, &executeHandlerWorkflowRepo{}, 1)
	handler := NewTaskboardExecuteHandler(pipeline, nil, nil)

	task, _, err := tasks.NewTaskboardExecuteTask(tasks.TaskboardExecutePayload{
		BoardID:        "board-1",
		SourceBranch:   "main",
		RepositoryRoot: ".",
	})
	if err != nil {
		t.Fatalf("unexpected task build error: %v", err)
	}

	interruptedCtx, cancel := context.WithCancel(context.Background())
	cancel()
	err = handler.ProcessTask(interruptedCtx, task)
	if err == nil {
		t.Fatalf("expected interruption error")
	}
	if errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected retryable interruption error, got skip retry: %v", err)
	}
}
