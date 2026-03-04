package worker

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	applicationingestion "agentic-orchestrator/internal/application/ingestion"
	"agentic-orchestrator/internal/application/taskengine"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type fakeTypedIngestionBoardStore struct{}

func (store *fakeTypedIngestionBoardStore) UpsertBoard(ctx context.Context, board domaintracker.Board) error {
	_ = ctx
	_ = board
	return nil
}

type fakeIngestionArtifactFetcher struct{}

func (fetcher *fakeIngestionArtifactFetcher) FetchToPath(ctx context.Context, objectPath string, destinationPath string) error {
	_ = ctx
	_ = objectPath
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destinationPath, []byte("doc content"), 0o644)
}

func TestIngestionAgentHandlerHandle(t *testing.T) {
	service, err := applicationingestion.NewService(&fakeTypedIngestionBoardStore{}, &fakeIngestionArtifactFetcher{})
	if err != nil {
		t.Fatalf("new ingestion service: %v", err)
	}
	handler, err := NewIngestionAgentHandler(service)
	if err != nil {
		t.Fatalf("new ingestion handler: %v", err)
	}
	payloadBytes, err := json.Marshal(applicationcontrolplane.IngestionAgentPayload{
		RunID:                     "run-1",
		TaskID:                    "ingestion",
		JobID:                     "job-1",
		ProjectID:                 "project-1",
		SelectedDocumentLocations: []string{"projects/project-1/documents/doc-1/doc.md"},
		SystemPrompt:              "System prompt",
		UserPrompt:                "User prompt",
		IdempotencyKey:            "ingestion-agent:key",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("expected panic placeholder")
		}
		if recovered != "TODO: IMPLEMENT AGENTIC AGENT" {
			t.Fatalf("unexpected panic value: %v", recovered)
		}
	}()
	_ = handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindIngestionAgent, Payload: payloadBytes})
}
