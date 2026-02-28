package worker

import (
	"agentic-orchestrator/internal/application/taskengine"
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type fakeTrackerService struct {
	request applicationtracker.SyncBoardRequest
	called  bool
	err     error
}

func (service *fakeTrackerService) SyncBoard(ctx context.Context, request applicationtracker.SyncBoardRequest) (domaintracker.Board, error) {
	_ = ctx
	service.request = request
	service.called = true
	return domaintracker.Board{}, service.err
}

func TestIngestionAgentHandlerDispatchesTrackerSync(t *testing.T) {
	service := &fakeTrackerService{}
	handler, err := NewIngestionAgentHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	payload, err := json.Marshal(IngestionAgentPayload{
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		IdempotencyKey: "id-1",
		Prompt:         "ingest tracker board",
		ProjectID:      "project-1",
		WorkflowID:     "workflow-1",
		BoardSource: IngestionBoardSourcePayload{
			Kind:     "local_json",
			Location: "board-1.json",
		},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindIngestionAgent, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !service.called {
		t.Fatalf("expected tracker service to be called")
	}
	if service.request.ProjectID != "project-1" || service.request.WorkflowID != "workflow-1" {
		t.Fatalf("expected project/workflow boundary propagation, got %+v", service.request)
	}
	if service.request.Source.Kind != domaintracker.SourceKindLocalJSON {
		t.Fatalf("expected source kind local_json, got %q", service.request.Source.Kind)
	}
}

func TestIngestionAgentHandlerReturnsServiceError(t *testing.T) {
	service := &fakeTrackerService{err: errors.New("boom")}
	handler, err := NewIngestionAgentHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	payload, _ := json.Marshal(IngestionAgentPayload{
		RunID:      "run-1",
		Prompt:     "ingest tracker board",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		BoardSource: IngestionBoardSourcePayload{
			Kind:     "local_json",
			Location: "board-1.json",
		},
	})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindIngestionAgent, Payload: payload}); err == nil {
		t.Fatalf("expected service error")
	}
}
