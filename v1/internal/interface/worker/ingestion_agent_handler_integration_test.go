package worker

import (
	"agentic-orchestrator/internal/application/taskengine"
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	infratracker "agentic-orchestrator/internal/infrastructure/tracker"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type queuedIngestionEngine struct {
	requests []taskengine.EnqueueRequest
}

func (engine *queuedIngestionEngine) Enqueue(ctx context.Context, request taskengine.EnqueueRequest) (taskengine.EnqueueResult, error) {
	_ = ctx
	engine.requests = append(engine.requests, request)
	return taskengine.EnqueueResult{QueueTaskID: fmt.Sprintf("queue-task-%d", len(engine.requests))}, nil
}

func (engine *queuedIngestionEngine) dispatchNext(ctx context.Context, handler taskengine.Handler) error {
	if len(engine.requests) == 0 {
		return fmt.Errorf("no queued requests")
	}
	request := engine.requests[0]
	engine.requests = engine.requests[1:]
	return handler.Handle(ctx, taskengine.Job{
		Kind:        request.Kind,
		QueueTaskID: request.IdempotencyKey,
		Payload:     request.Payload,
	})
}

func newIngestionTrackerService(t *testing.T, baseDirectory string) *applicationtracker.Service {
	t.Helper()

	localProvider, err := infratracker.NewLocalJSONProvider(baseDirectory)
	if err != nil {
		t.Fatalf("new local json provider: %v", err)
	}
	registry, err := infratracker.NewProviderRegistry(map[domaintracker.SourceKind]applicationtracker.Provider{
		domaintracker.SourceKindLocalJSON:    localProvider,
		domaintracker.SourceKindGitHubIssues: infratracker.NewGitHubIssuesProvider(),
	})
	if err != nil {
		t.Fatalf("new provider registry: %v", err)
	}
	service, err := applicationtracker.NewService(registry)
	if err != nil {
		t.Fatalf("new tracker service: %v", err)
	}
	return service
}

func TestIngestionWorkflowSyncsLocalJSONBoardThroughCanonicalModel(t *testing.T) {
	root := t.TempDir()
	boardPath := filepath.Join(root, "board-1.json")
	boardJSON := []byte(`{
  "board_id": "board-1",
  "run_id": "run-1",
  "status": "in-progress",
  "epics": [
    {
      "id": "epic-1",
      "board_id": "board-1",
      "title": "Implement tracker",
      "status": "in-progress",
      "tasks": [
        {
          "id": "task-1",
          "board_id": "board-1",
          "title": "Build canonical model",
          "status": "in-progress"
        }
      ]
    }
  ],
  "created_at": "2026-02-25T03:15:00Z",
  "updated_at": "2026-02-25T03:30:00Z"
}`)
	if err := os.WriteFile(boardPath, boardJSON, 0o644); err != nil {
		t.Fatalf("write board json: %v", err)
	}

	service := newIngestionTrackerService(t, root)
	handler, err := NewIngestionAgentHandler(service)
	if err != nil {
		t.Fatalf("new ingestion handler: %v", err)
	}
	queueEngine := &queuedIngestionEngine{}
	scheduler, err := taskengine.NewScheduler(queueEngine, taskengine.DefaultPolicies())
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}

	payload, err := json.Marshal(IngestionAgentPayload{
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		IdempotencyKey: "id-1",
		Prompt:         "ingest board",
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

	_, err = scheduler.Enqueue(context.Background(), taskengine.EnqueueRequest{
		Kind:           taskengine.JobKindIngestionAgent,
		Payload:        payload,
		IdempotencyKey: "id-1",
	})
	if err != nil {
		t.Fatalf("enqueue ingestion workflow: %v", err)
	}
	if len(queueEngine.requests) != 1 {
		t.Fatalf("expected one queued request, got %d", len(queueEngine.requests))
	}
	if queueEngine.requests[0].Queue != "ingestion" {
		t.Fatalf("expected ingestion queue, got %q", queueEngine.requests[0].Queue)
	}

	if err := queueEngine.dispatchNext(context.Background(), handler); err != nil {
		t.Fatalf("dispatch ingestion workflow: %v", err)
	}
}

func TestIngestionWorkflowRejectsInvalidCanonicalBoardDuringSync(t *testing.T) {
	root := t.TempDir()
	boardPath := filepath.Join(root, "board-invalid.json")
	invalidBoardJSON := []byte(`{
  "board_id": "board-1",
  "run_id": "run-1",
  "status": "in-progress",
  "epics": [
    {
      "id": "epic-1",
      "board_id": "board-1",
      "title": "Implement tracker",
      "status": "in-progress",
      "tasks": [
        {
          "id": "task-1",
          "board_id": "board-1",
          "title": "",
          "status": "in-progress"
        }
      ]
    }
  ],
  "created_at": "2026-02-25T03:15:00Z",
  "updated_at": "2026-02-25T03:30:00Z"
}`)
	if err := os.WriteFile(boardPath, invalidBoardJSON, 0o644); err != nil {
		t.Fatalf("write board json: %v", err)
	}

	service := newIngestionTrackerService(t, root)
	handler, err := NewIngestionAgentHandler(service)
	if err != nil {
		t.Fatalf("new ingestion handler: %v", err)
	}

	payload, _ := json.Marshal(IngestionAgentPayload{
		RunID:      "run-1",
		Prompt:     "ingest board",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		BoardSource: IngestionBoardSourcePayload{
			Kind:     "local_json",
			Location: "board-invalid.json",
		},
	})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindIngestionAgent, Payload: payload}); err == nil {
		t.Fatalf("expected canonical model validation error")
	}
}
