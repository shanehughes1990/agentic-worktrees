package tracker

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"testing"
)

type fakeInfraTrackerProvider struct{}

func (provider *fakeInfraTrackerProvider) SyncBoard(ctx context.Context, request applicationtracker.ProviderSyncRequest) (domaintracker.Board, error) {
	_ = ctx
	_ = request
	return domaintracker.Board{}, nil
}

func TestProviderRegistryResolveReturnsRegisteredProvider(t *testing.T) {
	registeredProvider := &fakeInfraTrackerProvider{}
	registry, err := NewProviderRegistry(map[domaintracker.SourceKind]applicationtracker.Provider{
		domaintracker.SourceKindLocalJSON: registeredProvider,
	})
	if err != nil {
		t.Fatalf("new provider registry: %v", err)
	}

	resolved, err := registry.Resolve(context.Background(), applicationtracker.ProviderSyncRequest{
		RunID:      "run-1",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source:     domaintracker.SourceRef{Kind: domaintracker.SourceKindLocalJSON, Location: "board-1.json"},
	})
	if err != nil {
		t.Fatalf("resolve provider: %v", err)
	}
	if resolved != registeredProvider {
		t.Fatalf("expected resolved provider to match registered provider")
	}
}

func TestProviderRegistryResolveRejectsUnregisteredKind(t *testing.T) {
	registry, err := NewProviderRegistry(map[domaintracker.SourceKind]applicationtracker.Provider{
		domaintracker.SourceKindLocalJSON: &fakeInfraTrackerProvider{},
	})
	if err != nil {
		t.Fatalf("new provider registry: %v", err)
	}

	_, err = registry.Resolve(context.Background(), applicationtracker.ProviderSyncRequest{
		RunID:      "run-1",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source:     domaintracker.SourceRef{Kind: domaintracker.SourceKind("unsupported"), BoardID: "TEAM-1"},
	})
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal error classification, got %q (%v)", failures.ClassOf(err), err)
	}
}
