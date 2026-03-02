package worker

import (
	"agentic-orchestrator/internal/application/taskengine"
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	infratracker "agentic-orchestrator/internal/infrastructure/tracker"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func newIngestionTrackerService(t *testing.T, db *gorm.DB) *applicationtracker.Service {
	t.Helper()

	internalProvider, err := infratracker.NewPostgresInternalProvider(db)
	if err != nil {
		t.Fatalf("new postgres internal provider: %v", err)
	}
	registry, err := infratracker.NewProviderRegistry(map[domaintracker.SourceKind]applicationtracker.Provider{
		domaintracker.SourceKindInternal:     internalProvider,
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

func newIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func seedInternalBoardSnapshot(t *testing.T, db *gorm.DB, projectID string, board domaintracker.Board) {
	t.Helper()
	payload, err := json.Marshal(board)
	if err != nil {
		t.Fatalf("marshal board payload: %v", err)
	}
	row := map[string]any{
		"run_id":      projectID,
		"board_id":    board.BoardID,
		"source_kind": "internal",
		"source_ref":  board.BoardID,
		"payload":     payload,
	}
	if err := db.Table("tracker_board_snapshots").Create(row).Error; err != nil {
		t.Fatalf("seed snapshot: %v", err)
	}
}

func TestIngestionWorkflowSyncsInternalBoardThroughCanonicalModel(t *testing.T) {
	db := newIntegrationDB(t)
	service := newIngestionTrackerService(t, db)
	seedInternalBoardSnapshot(t, db, "project-1", domaintracker.Board{
		BoardID: "board-1",
		RunID:   "seed-run",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindInternal,
			Location: "board-1",
			BoardID:  "board-1",
		},
		Status: domaintracker.StatusInProgress,
		Epics: []domaintracker.Epic{{
			WorkItem: domaintracker.WorkItem{ID: "epic-1", BoardID: "board-1", Title: "Implement tracker", Status: domaintracker.StatusInProgress},
			Tasks: []domaintracker.Task{{
				WorkItem: domaintracker.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Build canonical model", Status: domaintracker.StatusInProgress},
			}},
		}},
	})

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
		BoardSources: []IngestionBoardSourcePayload{{
			BoardID:                  "board-1",
			Kind:                     "internal",
			Location:                 "board-1",
			AppliesToAllRepositories: true,
		}},
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
	db := newIntegrationDB(t)
	service := newIngestionTrackerService(t, db)
	seedInternalBoardSnapshot(t, db, "project-1", domaintracker.Board{
		BoardID: "board-invalid",
		RunID:   "seed-run",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindInternal,
			Location: "board-invalid",
			BoardID:  "board-invalid",
		},
		Status: domaintracker.StatusInProgress,
		Epics: []domaintracker.Epic{{
			WorkItem: domaintracker.WorkItem{ID: "epic-1", BoardID: "board-invalid", Title: "Implement tracker", Status: domaintracker.StatusInProgress},
			Tasks: []domaintracker.Task{{
				WorkItem: domaintracker.WorkItem{ID: "task-1", BoardID: "board-invalid", Title: "", Status: domaintracker.StatusInProgress},
			}},
		}},
	})

	handler, err := NewIngestionAgentHandler(service)
	if err != nil {
		t.Fatalf("new ingestion handler: %v", err)
	}

	payload, _ := json.Marshal(IngestionAgentPayload{
		RunID:      "run-1",
		Prompt:     "ingest board",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		BoardSources: []IngestionBoardSourcePayload{{
			BoardID:                  "board-invalid",
			Kind:                     "internal",
			Location:                 "board-invalid",
			AppliesToAllRepositories: true,
		}},
	})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindIngestionAgent, Payload: payload}); err == nil {
		t.Fatalf("expected canonical model validation error")
	}
}
