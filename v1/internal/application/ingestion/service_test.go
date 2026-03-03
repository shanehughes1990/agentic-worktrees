package ingestion

import (
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type fakeBoardStore struct {
	board domaintracker.Board
}

func (store *fakeBoardStore) UpsertBoard(ctx context.Context, board domaintracker.Board) error {
	_ = ctx
	store.board = board
	return nil
}

type fakeArtifactFetcher struct {
	err             error
	fetchedObject   []string
	destinationPath []string
}

func (fetcher *fakeArtifactFetcher) FetchToPath(ctx context.Context, objectPath string, destinationPath string) error {
	_ = ctx
	if fetcher.err != nil {
		return fetcher.err
	}
	fetcher.fetchedObject = append(fetcher.fetchedObject, objectPath)
	fetcher.destinationPath = append(fetcher.destinationPath, destinationPath)
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destinationPath, []byte("doc content for taskboard"), 0o644)
}

func TestServiceExecutePanicsWithAgenticTodo(t *testing.T) {
	boardStore := &fakeBoardStore{}
	artifactFetcher := &fakeArtifactFetcher{}
	service, err := NewService(boardStore, artifactFetcher)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("expected panic placeholder")
		}
		if recovered != "TODO: IMPLEMENT AGENTIC AGENT" {
			t.Fatalf("unexpected panic value: %v", recovered)
		}
		if len(artifactFetcher.fetchedObject) != 1 {
			t.Fatalf("expected one fetched object, got %d", len(artifactFetcher.fetchedObject))
		}
		if !boardStore.board.CreatedAt.IsZero() {
			t.Fatalf("expected board store to remain unused")
		}
	}()

	_, _ = service.Execute(context.Background(), Request{
		RunID:                     "run-1",
		ProjectID:                 "project-1",
		SelectedDocumentLocations: []string{"projects/project-1/documents/doc-1/spec.md"},
		SystemPrompt:              "You are an ingestion planner.",
		UserPrompt:                "Create a delivery taskboard.",
	})
}

func TestServiceExecuteClassifiesFetchFailuresAsTransient(t *testing.T) {
	boardStore := &fakeBoardStore{}
	artifactFetcher := &fakeArtifactFetcher{err: errors.New("gcs timeout")}
	service, err := NewService(boardStore, artifactFetcher)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, executeErr := service.Execute(context.Background(), Request{
		RunID:                     "run-1",
		ProjectID:                 "project-1",
		SelectedDocumentLocations: []string{"projects/project-1/documents/doc-1/spec.md"},
		SystemPrompt:              "You are an ingestion planner.",
		UserPrompt:                "Create a delivery taskboard.",
	})
	if !failures.IsClass(executeErr, failures.ClassTransient) {
		t.Fatalf("expected transient failure class, got %q (%v)", failures.ClassOf(executeErr), executeErr)
	}
}

func TestServiceExecuteCleansTemporaryArtifacts(t *testing.T) {
	boardStore := &fakeBoardStore{}
	artifactFetcher := &fakeArtifactFetcher{}
	service, err := NewService(boardStore, artifactFetcher)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("expected panic placeholder")
		}
	}()

	_, _ = service.Execute(context.Background(), Request{
		RunID:                     "run-1",
		ProjectID:                 "project-1",
		SelectedDocumentLocations: []string{"projects/project-1/documents/doc-1/spec.md"},
		SystemPrompt:              "You are an ingestion planner.",
		UserPrompt:                "Create a delivery taskboard.",
	})
	if len(artifactFetcher.destinationPath) != 1 {
		t.Fatalf("expected one destination path, got %d", len(artifactFetcher.destinationPath))
	}
	if _, statErr := os.Stat(artifactFetcher.destinationPath[0]); !os.IsNotExist(statErr) {
		t.Fatalf("expected temporary artifact to be removed, stat err=%v", statErr)
	}
}
