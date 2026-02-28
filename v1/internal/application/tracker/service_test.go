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
	request ProviderSyncRequest
	board   domaintracker.Board
	err     error
}

func (provider *fakeTrackerProvider) SyncBoard(ctx context.Context, request ProviderSyncRequest) (domaintracker.Board, error) {
	_ = ctx
	provider.request = request
	return provider.board, provider.err
}

type fakeProviderResolver struct {
	request  ProviderSyncRequest
	provider Provider
	err      error
}

func (resolver *fakeProviderResolver) Resolve(ctx context.Context, request ProviderSyncRequest) (Provider, error) {
	_ = ctx
	resolver.request = request
	return resolver.provider, resolver.err
}

func TestServiceSyncBoardUsesResolvedProvider(t *testing.T) {
	provider := &fakeTrackerProvider{
		board: domaintracker.Board{
			BoardID: "board-1",
			RunID:   "run-1",
			Source:  domaintracker.SourceRef{Kind: domaintracker.SourceKindLocalJSON, Location: "board-1.json"},
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
	resolver := &fakeProviderResolver{provider: provider}
	service, err := NewService(resolver)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.SyncBoard(context.Background(), SyncBoardRequest{
		RunID:      "run-1",
		Prompt:     "ingest tracker board",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source:     domaintracker.SourceRef{Kind: domaintracker.SourceKindLocalJSON, Location: "board-1.json"},
	})
	if err != nil {
		t.Fatalf("sync board: %v", err)
	}
	if resolver.request.ProjectID != "project-1" || resolver.request.WorkflowID != "workflow-1" {
		t.Fatalf("expected project/workflow boundary selection to propagate, got %+v", resolver.request)
	}
	if provider.request.Source.Kind != domaintracker.SourceKindLocalJSON {
		t.Fatalf("expected local_json source kind, got %q", provider.request.Source.Kind)
	}
}

func TestServiceSyncBoardClassifiesUnknownProviderErrorsAsTransient(t *testing.T) {
	provider := &fakeTrackerProvider{err: errors.New("upstream timeout")}
	service, err := NewService(&fakeProviderResolver{provider: provider})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.SyncBoard(context.Background(), SyncBoardRequest{
		RunID:      "run-1",
		Prompt:     "ingest tracker board",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source:     domaintracker.SourceRef{Kind: domaintracker.SourceKindLocalJSON, Location: "board-1.json"},
	})
	if !failures.IsClass(err, failures.ClassTransient) {
		t.Fatalf("expected transient error classification, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestServiceSyncBoardRejectsInvalidBoundarySelection(t *testing.T) {
	service, err := NewService(&fakeProviderResolver{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.SyncBoard(context.Background(), SyncBoardRequest{
		RunID:      "run-1",
		Prompt:     "ingest tracker board",
		ProjectID:  "",
		WorkflowID: "workflow-1",
		Source:     domaintracker.SourceRef{Kind: domaintracker.SourceKindLocalJSON, Location: "board-1.json"},
	})
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}
