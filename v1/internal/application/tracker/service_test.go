package tracker

import (
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"errors"
	"testing"
	"time"
)

type fakeTrackerProvider struct {
	request IngestionSourceRequest
	board   domaintracker.Board
	err     error
}

func (provider *fakeTrackerProvider) SyncBoard(ctx context.Context, request IngestionSourceRequest) (domaintracker.Board, error) {
	_ = ctx
	provider.request = request
	return provider.board, provider.err
}

type fakeSourceResolver struct {
	request  IngestionSourceRequest
	source   IngestionSource
	err      error
}

type fakeBoardStore struct {
	board domaintracker.Board
	err   error
}

func (store *fakeBoardStore) UpsertBoard(ctx context.Context, board domaintracker.Board) error {
	_ = ctx
	store.board = board
	return store.err
}

func (resolver *fakeSourceResolver) Resolve(ctx context.Context, request IngestionSourceRequest) (IngestionSource, error) {
	_ = ctx
	resolver.request = request
	return resolver.source, resolver.err
}

func TestServiceSyncBoardUsesResolvedProvider(t *testing.T) {
	provider := &fakeTrackerProvider{
		board: domaintracker.Board{
			BoardID: "board-1",
			RunID:   "run-1",
			Source:  domaintracker.SourceRef{Kind: domaintracker.SourceKindInternal, Location: "board-1"},
			Status:  domaintracker.StatusInProgress,
			Epics: []domaintracker.Epic{
				{
					WorkItem: domaintracker.WorkItem{ID: "epic-1", BoardID: "board-1", Title: "Epic", Status: domaintracker.StatusInProgress},
					Tasks: []domaintracker.Task{
						{WorkItem: domaintracker.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Task", Status: domaintracker.StatusInProgress}},
					},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}
	resolver := &fakeSourceResolver{source: provider}
	store := &fakeBoardStore{}
	service, err := NewService(resolver, store)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.SyncBoard(context.Background(), SyncBoardRequest{
		RunID:      "run-1",
		Prompt:     "ingest tracker board",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source:     domaintracker.SourceRef{Kind: domaintracker.SourceKindInternal, Location: "board-1"},
	})
	if err != nil {
		t.Fatalf("sync board: %v", err)
	}
	if resolver.request.ProjectID != "project-1" || resolver.request.WorkflowID != "workflow-1" {
		t.Fatalf("expected project/workflow boundary selection to propagate, got %+v", resolver.request)
	}
	if provider.request.Source.Kind != domaintracker.SourceKindInternal {
		t.Fatalf("expected internal source kind, got %q", provider.request.Source.Kind)
	}
	if store.board.BoardID != "board-1" {
		t.Fatalf("expected persisted board, got %+v", store.board)
	}
}

func TestServiceSyncBoardClassifiesUnknownProviderErrorsAsTransient(t *testing.T) {
	provider := &fakeTrackerProvider{err: errors.New("upstream timeout")}
	service, err := NewService(&fakeSourceResolver{source: provider}, &fakeBoardStore{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.SyncBoard(context.Background(), SyncBoardRequest{
		RunID:      "run-1",
		Prompt:     "ingest tracker board",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source:     domaintracker.SourceRef{Kind: domaintracker.SourceKindInternal, Location: "board-1"},
	})
	if !failures.IsClass(err, failures.ClassTransient) {
		t.Fatalf("expected transient error classification, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestServiceSyncBoardRejectsInvalidBoundarySelection(t *testing.T) {
	service, err := NewService(&fakeSourceResolver{}, &fakeBoardStore{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.SyncBoard(context.Background(), SyncBoardRequest{
		RunID:      "run-1",
		Prompt:     "ingest tracker board",
		ProjectID:  "",
		WorkflowID: "workflow-1",
		Source:     domaintracker.SourceRef{Kind: domaintracker.SourceKindInternal, Location: "board-1"},
	})
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestServiceSyncBoardClassifiesBoardStoreErrorsAsTransient(t *testing.T) {
	provider := &fakeTrackerProvider{
		board: domaintracker.Board{
			BoardID: "board-1",
			RunID:   "run-1",
			Source:  domaintracker.SourceRef{Kind: domaintracker.SourceKindInternal, Location: "board-1"},
			Status:  domaintracker.StatusInProgress,
			Epics: []domaintracker.Epic{{
				WorkItem: domaintracker.WorkItem{ID: "epic-1", BoardID: "board-1", Title: "Epic", Status: domaintracker.StatusInProgress},
				Tasks: []domaintracker.Task{{
					WorkItem: domaintracker.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Task", Status: domaintracker.StatusInProgress},
				}},
			}},
		},
	}
	service, err := NewService(&fakeSourceResolver{source: provider}, &fakeBoardStore{err: errors.New("db unavailable")})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.SyncBoard(context.Background(), SyncBoardRequest{
		RunID:      "run-1",
		Prompt:     "ingest tracker board",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source:     domaintracker.SourceRef{Kind: domaintracker.SourceKindInternal, Location: "board-1"},
	})
	if !failures.IsClass(err, failures.ClassTransient) {
		t.Fatalf("expected transient error classification, got %q (%v)", failures.ClassOf(err), err)
	}
}
