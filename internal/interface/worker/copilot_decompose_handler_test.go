package worker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
)

type fakeDecomposer struct{}

type failingDecomposer struct {
	err error
}

func (decomposer *fakeDecomposer) Decompose(context.Context, appcopilot.DecomposeRequest) (appcopilot.DecomposeResult, error) {
	return appcopilot.DecomposeResult{
		Response: `{"board_id":"b1","run_id":"r1","status":"in-progress","epics":[{"id":"e1","board_id":"b1","title":"Epic","status":"in-progress","tasks":[{"id":"t1","board_id":"b1","title":"Task","status":"not-started"}]}],"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`,
	}, nil
}

func (decomposer *failingDecomposer) Decompose(context.Context, appcopilot.DecomposeRequest) (appcopilot.DecomposeResult, error) {
	return appcopilot.DecomposeResult{}, decomposer.err
}

type captureRepo struct {
	saved    *domaintaskboard.Board
	workflow *apptaskboard.IngestionWorkflow
}

func (repo *captureRepo) GetByBoardID(context.Context, string) (*domaintaskboard.Board, error) {
	return nil, nil
}

func (repo *captureRepo) Save(context.Context, *domaintaskboard.Board) error {
	repo.saved = &domaintaskboard.Board{BoardID: "saved"}
	return nil
}

func (repo *captureRepo) GetWorkflow(context.Context, string) (*apptaskboard.IngestionWorkflow, error) {
	return repo.workflow, nil
}

func (repo *captureRepo) ListWorkflows(context.Context) ([]apptaskboard.IngestionWorkflow, error) {
	if repo.workflow == nil {
		return []apptaskboard.IngestionWorkflow{}, nil
	}
	return []apptaskboard.IngestionWorkflow{*repo.workflow}, nil
}

func (repo *captureRepo) SaveWorkflow(_ context.Context, workflow *apptaskboard.IngestionWorkflow) error {
	repo.workflow = workflow
	return nil
}

func TestCopilotDecomposeHandlerProcessTask(t *testing.T) {
	repo := &captureRepo{}
	handler := NewCopilotDecomposeHandler(&fakeDecomposer{}, repo, repo, nil)

	taskPayload := tasks.CopilotDecomposePayload{RunID: "run-1", Prompt: "prompt", Model: "gpt-5", WorkingDirectory: "."}
	task, _, err := tasks.NewCopilotDecomposeTask(taskPayload)
	if err != nil {
		t.Fatalf("unexpected task build error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := handler.ProcessTask(ctx, task); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	if repo.saved == nil {
		t.Fatalf("expected board save")
	}
	if repo.workflow == nil || repo.workflow.Status != apptaskboard.WorkflowStatusCompleted {
		t.Fatalf("expected completed workflow status, got %#v", repo.workflow)
	}
}

func TestCopilotDecomposeHandlerProcessTaskSkipsRetryOnTerminalFailure(t *testing.T) {
	repo := &captureRepo{}
	handler := NewCopilotDecomposeHandler(&failingDecomposer{err: fmt.Errorf("start copilot client: CLI process exited: exit status 1")}, repo, repo, nil)

	taskPayload := tasks.CopilotDecomposePayload{RunID: "run-2", Prompt: "prompt", Model: "gpt-5", WorkingDirectory: "."}
	task, _, err := tasks.NewCopilotDecomposeTask(taskPayload)
	if err != nil {
		t.Fatalf("unexpected task build error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = handler.ProcessTask(ctx, task)
	if err == nil {
		t.Fatalf("expected handler error")
	}
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected skip retry error, got: %v", err)
	}
	if repo.workflow == nil || repo.workflow.Status != apptaskboard.WorkflowStatusFailed {
		t.Fatalf("expected failed workflow status, got %#v", repo.workflow)
	}
}

var _ appcopilot.Decomposer = (*fakeDecomposer)(nil)
var _ appcopilot.Decomposer = (*failingDecomposer)(nil)
var _ apptaskboard.Repository = (*captureRepo)(nil)
var _ apptaskboard.WorkflowRepository = (*captureRepo)(nil)
