package resolvers

import (
	applicationsupervisor "agentic-orchestrator/internal/application/supervisor"
	"agentic-orchestrator/internal/application/taskengine"
	domainsupervisor "agentic-orchestrator/internal/domain/supervisor"
	"agentic-orchestrator/internal/interface/graphql/models"
	"context"
	"testing"
)

type supervisorMemoryEventStore struct {
	events []domainsupervisor.Decision
}

func (store *supervisorMemoryEventStore) Append(_ context.Context, decision domainsupervisor.Decision) error {
	store.events = append(store.events, decision)
	return nil
}

func (store *supervisorMemoryEventStore) ListByCorrelation(_ context.Context, correlation domainsupervisor.CorrelationIDs) ([]domainsupervisor.Decision, error) {
	results := make([]domainsupervisor.Decision, 0)
	for _, event := range store.events {
		if event.CorrelationIDs == correlation {
			results = append(results, event)
		}
	}
	return results, nil
}

func newSupervisorServiceForResolverTests(t *testing.T) *applicationsupervisor.Service {
	t.Helper()
	service, err := applicationsupervisor.NewService(&supervisorMemoryEventStore{}, nil)
	if err != nil {
		t.Fatalf("new supervisor service: %v", err)
	}
	correlation := taskengine.CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1"}
	if _, err := service.OnIssueOpened(context.Background(), correlation, "octo/repo", "octo/repo#1"); err != nil {
		t.Fatalf("seed issue opened: %v", err)
	}
	if _, err := service.OnIssueApproved(context.Background(), correlation, "octo/repo", "octo/repo#1", "human"); err != nil {
		t.Fatalf("seed issue approved: %v", err)
	}
	return service
}

func TestSupervisorDecisionHistoryResolver(t *testing.T) {
	service := newSupervisorServiceForResolverTests(t)
	resolver := &Resolver{SupervisorService: service}
	results, err := (&queryResolver{resolver}).SupervisorDecisionHistory(context.Background(), models.SupervisorCorrelationInput{RunID: "run-1", TaskID: "task-1", JobID: "job-1"})
	if err != nil {
		t.Fatalf("SupervisorDecisionHistory() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected two supervisor decisions, got %d", len(results))
	}
	if results[0].Reason != string(domainsupervisor.ReasonIssueAwaitingApproval) {
		t.Fatalf("expected first reason %q, got %q", domainsupervisor.ReasonIssueAwaitingApproval, results[0].Reason)
	}
	if results[1].Reason != string(domainsupervisor.ReasonIssueTaskKickoff) {
		t.Fatalf("expected second reason %q, got %q", domainsupervisor.ReasonIssueTaskKickoff, results[1].Reason)
	}
}

func TestSupervisorDecisionHistoryStreamResolver(t *testing.T) {
	service := newSupervisorServiceForResolverTests(t)
	resolver := &Resolver{SupervisorService: service}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := (&subscriptionResolver{resolver}).SupervisorDecisionHistoryStream(ctx, models.SupervisorCorrelationInput{RunID: "run-1", TaskID: "task-1", JobID: "job-1"}, nil)
	if err != nil {
		t.Fatalf("SupervisorDecisionHistoryStream() error = %v", err)
	}
	first, ok := <-stream
	if !ok {
		t.Fatalf("expected open stream")
	}
	if len(first) != 2 {
		t.Fatalf("expected first stream payload with two decisions, got %d", len(first))
	}
	cancel()
}
