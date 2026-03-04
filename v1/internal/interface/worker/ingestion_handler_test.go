package worker

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	applicationingestion "agentic-orchestrator/internal/application/ingestion"
	"agentic-orchestrator/internal/application/taskengine"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"encoding/json"
	"fmt"
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

type fakeIngestionAgentRunner struct{}

func (runner *fakeIngestionAgentRunner) GenerateTaskboard(ctx context.Context, sandboxDir string, prompt string, outputPath string, model string) error {
	_ = ctx
	_ = sandboxDir
	_ = prompt
	_ = model
	content := `{
		"board_id": "board-1",
		"run_id": "run-1",
		"status": "not-started",
		"epics": [{
			"id": "epic-1",
			"board_id": "board-1",
			"title": "Epic",
			"status": "not-started",
			"tasks": [{
				"id": "task-1",
				"board_id": "board-1",
				"title": "Task",
				"status": "not-started"
			}]
		}],
		"created_at": "2026-03-03T10:00:00Z",
		"updated_at": "2026-03-03T10:00:00Z"
	}`
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write generated board: %w", err)
	}
	return nil
}

type fakeIngestionRepositorySynchronizer struct{}

func (synchronizer *fakeIngestionRepositorySynchronizer) Sync(ctx context.Context, projectID string, sandboxDir string, sourceBranch string, sourceRepositories []applicationingestion.SourceRepository) error {
	_ = ctx
	_ = projectID
	_ = sandboxDir
	_ = sourceBranch
	_ = sourceRepositories
	return nil
}

func TestIngestionAgentHandlerHandle(t *testing.T) {
	service, err := applicationingestion.NewService(&fakeTypedIngestionBoardStore{}, &fakeIngestionArtifactFetcher{}, &fakeIngestionAgentRunner{}, &fakeIngestionRepositorySynchronizer{})
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
		SourceRepositories:        []applicationcontrolplane.IngestionSourceRepository{{RepositoryID: "repo-1", RepositoryURL: "https://github.com/acme/source-repo.git"}},
		SourceBranch:              "develop",
		Model:                     "gpt-5.3-codex",
		SystemPrompt:              "System prompt",
		UserPrompt:                "User prompt",
		IdempotencyKey:            "ingestion-agent:key",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindIngestionAgent, Payload: payloadBytes}); err != nil {
		t.Fatalf("handle ingestion job: %v", err)
	}
}
