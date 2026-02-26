package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
)

type fakeDecomposer struct {
	responses []string
	requests  []appcopilot.DecomposeRequest
}

type failingDecomposer struct {
	err error
}

func (decomposer *fakeDecomposer) Decompose(_ context.Context, request appcopilot.DecomposeRequest) (appcopilot.DecomposeResult, error) {
	decomposer.requests = append(decomposer.requests, request)
	response := `{"board_id":"b1","run_id":"r1","status":"in-progress","epics":[{"id":"e1","board_id":"b1","title":"Epic","status":"in-progress","tasks":[{"id":"t1","board_id":"b1","title":"Task","status":"not-started"}]}],"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`
	if len(decomposer.responses) > 0 {
		response = decomposer.responses[0]
		decomposer.responses = decomposer.responses[1:]
	}
	return appcopilot.DecomposeResult{
		Response: response,
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

func (repo *captureRepo) ListBoardIDs(context.Context) ([]string, error) {
	if repo.saved == nil {
		return []string{}, nil
	}
	return []string{repo.saved.BoardID}, nil
}

func (repo *captureRepo) Save(_ context.Context, board *domaintaskboard.Board) error {
	copiedBoard := *board
	repo.saved = &copiedBoard
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
	decomposer := &fakeDecomposer{responses: []string{
		`{"board_id":"b1","run_id":"r1","status":"in-progress","epics":[{"id":"e1","board_id":"b1","title":"Epic","status":"in-progress","tasks":[{"id":"t1","board_id":"b1","title":"Task one","status":"not-started"},{"id":"t2","board_id":"b1","title":"Task two","status":"not-started","depends_on":["t1"]}]}],"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`,
	}}
	handler := NewCopilotDecomposeHandler(decomposer, repo, repo, nil)

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
	qualityRaw, exists := repo.saved.Metadata["quality_report"]
	if !exists {
		t.Fatalf("expected quality_report metadata to be set on saved board")
	}
	qualityReport, ok := qualityRaw.(apptaskboard.BoardQualityReport)
	if !ok {
		t.Fatalf("expected quality_report metadata to be BoardQualityReport, got %T", qualityRaw)
	}
	if !qualityReport.Passed || qualityReport.Score < apptaskboard.DefaultBoardQualityThreshold {
		t.Fatalf("expected saved board quality report to be passing, got %#v", qualityReport)
	}
	if len(decomposer.requests) != 1 {
		t.Fatalf("expected only generation call when quality already passes, got %d", len(decomposer.requests))
	}
}

func TestCopilotDecomposeHandlerProcessTaskRetriesSupervisorUntilQualityPasses(t *testing.T) {
	repo := &captureRepo{}
	decomposer := &fakeDecomposer{responses: []string{
		`{"board_id":"b1","run_id":"r1","status":"in-progress","epics":[{"id":"e1","board_id":"b1","title":"Epic","status":"in-progress","tasks":[{"id":"t1","board_id":"b1","title":"Wire filesystem adapter as default provider","status":"not-started"},{"id":"t2","board_id":"b1","title":"Wire filesystem adapter as default provider","status":"not-started"}]}],"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`,
		`{"board_id":"b1","run_id":"r1","status":"in-progress","epics":[{"id":"e1","board_id":"b1","title":"Epic","status":"in-progress","tasks":[{"id":"t1","board_id":"b1","title":"Define source contract package","status":"not-started"},{"id":"t2","board_id":"b1","title":"Implement filesystem adapter package","status":"not-started","depends_on":["t1"]}]}],"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`,
	}}
	handler := NewCopilotDecomposeHandler(decomposer, repo, repo, nil)

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
	if len(decomposer.requests) != 2 {
		t.Fatalf("expected generation + one supervisor retry, got %d", len(decomposer.requests))
	}
	if !strings.Contains(decomposer.requests[1].Prompt, "You are a taskboard quality supervisor.") {
		t.Fatalf("expected second prompt to be supervisor prompt")
	}
}

func TestCopilotDecomposeHandlerProcessTaskFailsWhenQualityNeverPasses(t *testing.T) {
	repo := &captureRepo{}
	decomposer := &fakeDecomposer{responses: []string{
		`{"board_id":"b1","run_id":"r1","status":"in-progress","epics":[{"id":"e1","board_id":"b1","title":"Epic","status":"in-progress","tasks":[{"id":"t1","board_id":"b1","title":"Wire filesystem adapter as default provider","status":"not-started"},{"id":"t2","board_id":"b1","title":"Wire filesystem adapter as default provider","status":"not-started"}]}],"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`,
		`{"board_id":"b1","run_id":"r1","status":"in-progress","epics":[{"id":"e1","board_id":"b1","title":"Epic","status":"in-progress","tasks":[{"id":"t1","board_id":"b1","title":"Wire filesystem adapter as default provider","status":"not-started"},{"id":"t2","board_id":"b1","title":"Wire filesystem adapter as default provider","status":"not-started"}]}],"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`,
		`{"board_id":"b1","run_id":"r1","status":"in-progress","epics":[{"id":"e1","board_id":"b1","title":"Epic","status":"in-progress","tasks":[{"id":"t1","board_id":"b1","title":"Wire filesystem adapter as default provider","status":"not-started"},{"id":"t2","board_id":"b1","title":"Wire filesystem adapter as default provider","status":"not-started"}]}],"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`,
	}}
	handler := NewCopilotDecomposeHandler(decomposer, repo, repo, nil)

	taskPayload := tasks.CopilotDecomposePayload{RunID: "run-1", Prompt: "prompt", Model: "gpt-5", WorkingDirectory: "."}
	task, _, err := tasks.NewCopilotDecomposeTask(taskPayload)
	if err != nil {
		t.Fatalf("unexpected task build error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = handler.ProcessTask(ctx, task)
	if err == nil {
		t.Fatalf("expected quality gate failure")
	}
	if repo.workflow == nil || repo.workflow.Status != apptaskboard.WorkflowStatusFailed {
		t.Fatalf("expected failed workflow status, got %#v", repo.workflow)
	}
	if len(decomposer.requests) != 3 {
		t.Fatalf("expected generation + 2 supervisor retries, got %d", len(decomposer.requests))
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

func TestCopilotDecomposeHandlerProcessTaskRetriesOnStartupProbeKilled(t *testing.T) {
	repo := &captureRepo{}
	handler := NewCopilotDecomposeHandler(&failingDecomposer{err: fmt.Errorf("git worktree flow: implement task with agent: copilot preflight failed: copilot cli startup probe failed: signal: killed")}, repo, repo, nil)

	taskPayload := tasks.CopilotDecomposePayload{RunID: "run-3", Prompt: "prompt", Model: "gpt-5", WorkingDirectory: "."}
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
	if errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected retryable error, got skip retry: %v", err)
	}
	if repo.workflow == nil || repo.workflow.Status != apptaskboard.WorkflowStatusFailed {
		t.Fatalf("expected failed workflow status, got %#v", repo.workflow)
	}
}

var _ appcopilot.Decomposer = (*fakeDecomposer)(nil)
var _ appcopilot.Decomposer = (*failingDecomposer)(nil)
var _ apptaskboard.Repository = (*captureRepo)(nil)
var _ apptaskboard.WorkflowRepository = (*captureRepo)(nil)
