package postgres

import (
	"agentic-orchestrator/internal/domain/failures"
	domainsupervisor "agentic-orchestrator/internal/domain/supervisor"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSupervisorEventStoreTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func TestSupervisorEventStoreAppendAndListByCorrelation(t *testing.T) {
	store, err := NewEventStore(newSupervisorEventStoreTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}

	now := time.Now().UTC()
	decision := domainsupervisor.Decision{
		CorrelationIDs: domainsupervisor.CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1", ProjectID: "project-1"},
		SignalType:      domainsupervisor.SignalExecutionFailed,
		FromState:       domainsupervisor.StateExecuting,
		ToState:         domainsupervisor.StateRework,
		Action:          domainsupervisor.ActionRequestRework,
		Reason:          domainsupervisor.ReasonExecutionFailedRetry,
		RuleName:        "retry_rule",
		RulePriority:    10,
		OccurredAt:      now,
		Attempt:         1,
		MaxRetry:        3,
		FailureClass:    failures.ClassTransient,
		AttentionZone:   domainsupervisor.AttentionZoneExecution,
		Metadata:        map[string]string{"worker_id": "worker-1"},
	}

	if err := store.Append(context.Background(), decision); err != nil {
		t.Fatalf("append decision: %v", err)
	}

	decisions, err := store.ListByCorrelation(context.Background(), decision.CorrelationIDs)
	if err != nil {
		t.Fatalf("list decisions: %v", err)
	}
	if len(decisions) != 1 {
		t.Fatalf("expected one decision, got %d", len(decisions))
	}
	if decisions[0].Metadata["worker_id"] != "worker-1" {
		t.Fatalf("expected metadata to roundtrip, got %#v", decisions[0].Metadata)
	}
	if decisions[0].FailureClass != failures.ClassTransient {
		t.Fatalf("expected failure class transient, got %q", decisions[0].FailureClass)
	}
}

func TestSupervisorEventStoreProjectFilterRestrictsResults(t *testing.T) {
	store, err := NewEventStore(newSupervisorEventStoreTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}

	now := time.Now().UTC()
	appendDecision := func(projectID string, ruleName string) {
		t.Helper()
		err := store.Append(context.Background(), domainsupervisor.Decision{
			CorrelationIDs: domainsupervisor.CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1", ProjectID: projectID},
			SignalType:      domainsupervisor.SignalExecutionProgressed,
			FromState:       domainsupervisor.StateExecuting,
			ToState:         domainsupervisor.StateExecuting,
			Action:          domainsupervisor.ActionContinue,
			Reason:          domainsupervisor.ReasonExecutionProgressed,
			RuleName:        ruleName,
			RulePriority:    1,
			OccurredAt:      now,
		})
		if err != nil {
			t.Fatalf("append decision for %s: %v", projectID, err)
		}
	}

	appendDecision("project-1", "rule_p1")
	appendDecision("project-2", "rule_p2")

	all, err := store.ListByCorrelation(context.Background(), domainsupervisor.CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1"})
	if err != nil {
		t.Fatalf("list all decisions by correlation: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected two decisions without project filter, got %d", len(all))
	}

	filtered, err := store.ListByCorrelation(context.Background(), domainsupervisor.CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1", ProjectID: "project-2"})
	if err != nil {
		t.Fatalf("list filtered decisions by correlation: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected one decision with project filter, got %d", len(filtered))
	}
	if filtered[0].CorrelationIDs.ProjectID != "project-2" {
		t.Fatalf("expected only project-2 decision, got project_id %q", filtered[0].CorrelationIDs.ProjectID)
	}
}
