package controlplane

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

type captureEngine struct {
	last taskengine.EnqueueRequest
}

func (engine *captureEngine) Enqueue(ctx context.Context, request taskengine.EnqueueRequest) (taskengine.EnqueueResult, error) {
	_ = ctx
	engine.last = request
	return taskengine.EnqueueResult{QueueTaskID: "queue-1", Duplicate: false}, nil
}

type ingestionQueryRepoStub struct{}

func (repository *ingestionQueryRepoStub) ListSessions(ctx context.Context, limit int) ([]SessionSummary, error) {
	_ = ctx
	_ = limit
	return nil, nil
}

func (repository *ingestionQueryRepoStub) GetSession(ctx context.Context, runID string) (*SessionSummary, error) {
	_ = ctx
	_ = runID
	return nil, nil
}

func (repository *ingestionQueryRepoStub) ListWorkflowJobs(ctx context.Context, runID string, taskID string, limit int) ([]WorkflowJob, error) {
	_ = ctx
	_ = runID
	_ = taskID
	_ = limit
	return nil, nil
}

func (repository *ingestionQueryRepoStub) ListExecutionHistory(ctx context.Context, filter CorrelationFilter, limit int) ([]ExecutionHistoryRecord, error) {
	_ = ctx
	_ = filter
	_ = limit
	return nil, nil
}

func (repository *ingestionQueryRepoStub) ListDeadLetterHistory(ctx context.Context, queue string, limit int) ([]DeadLetterHistoryRecord, error) {
	_ = ctx
	_ = queue
	_ = limit
	return nil, nil
}

func (repository *ingestionQueryRepoStub) ListLifecycleSessionSnapshots(ctx context.Context, projectID string, pipelineType string, limit int) ([]LifecycleSessionSnapshot, error) {
	_ = ctx
	_ = projectID
	_ = pipelineType
	_ = limit
	return nil, nil
}

func (repository *ingestionQueryRepoStub) ListLifecycleSessionHistory(ctx context.Context, projectID string, sessionID string, fromEventSeq int64, limit int) ([]LifecycleHistoryEvent, error) {
	_ = ctx
	_ = projectID
	_ = sessionID
	_ = fromEventSeq
	_ = limit
	return nil, nil
}

func (repository *ingestionQueryRepoStub) ListLifecycleTreeNodes(ctx context.Context, filter LifecycleTreeFilter, limit int) ([]LifecycleTreeNode, error) {
	_ = ctx
	_ = filter
	_ = limit
	return nil, nil
}

type ingestionProjectRepoStub struct {
	setup *ProjectSetup
}

func (repository *ingestionProjectRepoStub) ListProjectSetups(ctx context.Context, limit int) ([]ProjectSetup, error) {
	_ = ctx
	_ = limit
	if repository.setup == nil {
		return nil, nil
	}
	return []ProjectSetup{*repository.setup}, nil
}

func (repository *ingestionProjectRepoStub) GetProjectSetup(ctx context.Context, projectID string) (*ProjectSetup, error) {
	_ = ctx
	_ = projectID
	return repository.setup, nil
}

func (repository *ingestionProjectRepoStub) UpsertProjectSetup(ctx context.Context, setup ProjectSetup) (*ProjectSetup, error) {
	_ = ctx
	setup.CreatedAt = time.Now().UTC()
	setup.UpdatedAt = setup.CreatedAt
	return &setup, nil
}

func (repository *ingestionProjectRepoStub) DeleteProjectSetup(ctx context.Context, projectID string) error {
	_ = ctx
	_ = projectID
	return nil
}

func newControlPlaneIngestionService(t *testing.T, setup *ProjectSetup, engine *captureEngine) *Service {
	t.Helper()
	scheduler, err := taskengine.NewScheduler(engine, taskengine.DefaultPolicies())
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}
	service, err := NewService(scheduler, &ingestionQueryRepoStub{}, &ingestionProjectRepoStub{setup: setup}, nil)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return service
}

func TestRunIngestionAgentIncludesExplicitBoardAndStreamInPayload(t *testing.T) {
	setup := &ProjectSetup{
		ProjectID: "project-1",
		Boards:    []ProjectBoard{{BoardID: "board-from-setup"}},
	}
	engine := &captureEngine{}
	service := newControlPlaneIngestionService(t, setup, engine)

	_, err := service.RunIngestionAgent(context.Background(), RunIngestionAgentInput{
		ProjectID:           "project-1",
		TaskboardName:       "Board Explicit",
		SelectedDocumentIDs: []string{"doc-1"},
		UserPrompt:          "prompt",
	})
	if err != nil {
		t.Fatalf("RunIngestionAgent() error = %v", err)
	}

	var payload IngestionAgentPayload
	if err := json.Unmarshal(engine.last.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.BoardID == "" {
		t.Fatalf("expected payload board_id to be generated")
	}
	if payload.BoardID == "board-from-setup" {
		t.Fatalf("expected payload board_id to not reuse setup board id")
	}
	if payload.StreamID == "" {
		t.Fatalf("expected non-empty stream_id")
	}
	if payload.StreamID != payload.JobID {
		t.Fatalf("expected stream_id to match job_id, got stream=%q job=%q", payload.StreamID, payload.JobID)
	}
	if !payload.PreferSelectedDocuments {
		t.Fatalf("expected prefer_selected_documents=true when selected documents are present")
	}
}

func TestRunIngestionAgentUsesProjectBoardWhenInputBoardMissing(t *testing.T) {
	setup := &ProjectSetup{
		ProjectID: "project-1",
		Boards:    []ProjectBoard{{BoardID: "board-from-setup"}},
	}
	engine := &captureEngine{}
	service := newControlPlaneIngestionService(t, setup, engine)

	_, err := service.RunIngestionAgent(context.Background(), RunIngestionAgentInput{
		ProjectID:     "project-1",
		TaskboardName: "Board From Prompt",
		UserPrompt:    "prompt-only ingestion",
	})
	if err != nil {
		t.Fatalf("RunIngestionAgent() error = %v", err)
	}

	var payload IngestionAgentPayload
	if err := json.Unmarshal(engine.last.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.BoardID == "" {
		t.Fatalf("expected payload board_id to be generated")
	}
	if payload.BoardID == "board-from-setup" {
		t.Fatalf("expected payload board_id to not reuse setup board id")
	}
	if payload.PreferSelectedDocuments {
		t.Fatalf("expected prefer_selected_documents=false when selected documents are absent")
	}
}

func TestRunIngestionAgentAllowsDocsOnlyWithoutPrompt(t *testing.T) {
	setup := &ProjectSetup{
		ProjectID: "project-1",
		Boards:    []ProjectBoard{{BoardID: "board-from-setup"}},
	}
	engine := &captureEngine{}
	service := newControlPlaneIngestionService(t, setup, engine)

	_, err := service.RunIngestionAgent(context.Background(), RunIngestionAgentInput{
		ProjectID:           "project-1",
		TaskboardName:       "Docs Only Board",
		SelectedDocumentIDs: []string{"doc-1"},
	})
	if err != nil {
		t.Fatalf("RunIngestionAgent() error = %v", err)
	}

	var payload IngestionAgentPayload
	if err := json.Unmarshal(engine.last.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if len(payload.SelectedDocumentLocations) != 1 {
		t.Fatalf("expected one selected document location, got %d", len(payload.SelectedDocumentLocations))
	}
}

func TestRunIngestionAgentRequiresTaskboardName(t *testing.T) {
	setup := &ProjectSetup{ProjectID: "project-1"}
	engine := &captureEngine{}
	service := newControlPlaneIngestionService(t, setup, engine)

	_, err := service.RunIngestionAgent(context.Background(), RunIngestionAgentInput{
		ProjectID:  "project-1",
		UserPrompt: "prompt-only ingestion",
	})
	if err == nil || !strings.Contains(err.Error(), "taskboard_name is required") {
		t.Fatalf("expected taskboard_name validation error, got %v", err)
	}
}
