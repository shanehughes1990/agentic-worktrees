package tracker

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"testing"
)

func TestJiraProviderDefinesBoundaryAndReturnsNotImplemented(t *testing.T) {
	provider := NewJiraProvider()
	_, err := provider.SyncBoard(context.Background(), applicationtracker.ProviderSyncRequest{
		RunID:      "run-1",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source: domaintracker.SourceRef{
			Kind:    domaintracker.SourceKindJira,
			BoardID: "TEAM-1",
		},
	})
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal not-implemented error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestLinearProviderDefinesBoundaryAndReturnsNotImplemented(t *testing.T) {
	provider := NewLinearProvider()
	_, err := provider.SyncBoard(context.Background(), applicationtracker.ProviderSyncRequest{
		RunID:      "run-1",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source: domaintracker.SourceRef{
			Kind:    domaintracker.SourceKindLinear,
			BoardID: "TEAM-1",
		},
	})
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal not-implemented error, got %q (%v)", failures.ClassOf(err), err)
	}
}
