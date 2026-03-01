package supervisor

import (
	"agentic-orchestrator/internal/application/taskengine"
	domainsupervisor "agentic-orchestrator/internal/domain/supervisor"
	"context"
	"testing"
	"time"
)

type memoryEventStore struct {
	events []domainsupervisor.Decision
}

func (store *memoryEventStore) Append(_ context.Context, decision domainsupervisor.Decision) error {
	store.events = append(store.events, decision)
	return nil
}

func (store *memoryEventStore) ListByCorrelation(_ context.Context, correlation domainsupervisor.CorrelationIDs) ([]domainsupervisor.Decision, error) {
	results := make([]domainsupervisor.Decision, 0)
	for _, event := range store.events {
		if event.CorrelationIDs == correlation {
			results = append(results, event)
		}
	}
	return results, nil
}

func TestServiceOnAdmission(t *testing.T) {
	store := &memoryEventStore{}
	service, err := NewService(store, nil)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	_, err = service.OnAdmission(context.Background(), taskengine.AdmissionRecord{
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		JobKind:        taskengine.JobKindAgentWorkflow,
		IdempotencyKey: "idem-1",
		QueueTaskID:    "qt-1",
		Queue:          "agent",
		Status:         taskengine.AdmissionStatusQueued,
		EnqueuedAt:     time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("OnAdmission() error = %v", err)
	}
	if len(store.events) != 1 {
		t.Fatalf("expected one decision persisted")
	}
	if store.events[0].Reason != domainsupervisor.ReasonJobAdmitted {
		t.Fatalf("expected reason %q got %q", domainsupervisor.ReasonJobAdmitted, store.events[0].Reason)
	}
}

func TestServiceIssueRequiresApprovalBeforeKickoff(t *testing.T) {
	store := &memoryEventStore{}
	service, err := NewService(store, nil)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	correlation := taskengine.CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1"}

	openedDecision, err := service.OnIssueOpened(context.Background(), correlation, "octo/repo", "octo/repo#1")
	if err != nil {
		t.Fatalf("OnIssueOpened() error = %v", err)
	}
	if openedDecision.Action != domainsupervisor.ActionBlock {
		t.Fatalf("expected action %q got %q", domainsupervisor.ActionBlock, openedDecision.Action)
	}
	if openedDecision.Reason != domainsupervisor.ReasonIssueAwaitingApproval {
		t.Fatalf("expected reason %q got %q", domainsupervisor.ReasonIssueAwaitingApproval, openedDecision.Reason)
	}

	approvedDecision, err := service.OnIssueApproved(context.Background(), correlation, "octo/repo", "octo/repo#1", "human-1")
	if err != nil {
		t.Fatalf("OnIssueApproved() error = %v", err)
	}
	if approvedDecision.Action != domainsupervisor.ActionStartTask {
		t.Fatalf("expected action %q got %q", domainsupervisor.ActionStartTask, approvedDecision.Action)
	}
	if approvedDecision.Reason != domainsupervisor.ReasonIssueTaskKickoff {
		t.Fatalf("expected reason %q got %q", domainsupervisor.ReasonIssueTaskKickoff, approvedDecision.Reason)
	}
}

func TestServiceExecutionFailureEscalatesOnMaxRetries(t *testing.T) {
	store := &memoryEventStore{}
	service, err := NewService(store, nil)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	now := time.Now().UTC()
	_, err = service.OnExecution(context.Background(), taskengine.ExecutionRecord{
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		JobKind:        taskengine.JobKindSCMWorkflow,
		IdempotencyKey: "idem-1",
		Step:           "ensure_worktree",
		Status:         taskengine.ExecutionStatusRunning,
		UpdatedAt:      now,
	}, 1, 3)
	if err != nil {
		t.Fatalf("OnExecution(running) error = %v", err)
	}
	_, err = service.OnExecution(context.Background(), taskengine.ExecutionRecord{
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		JobKind:        taskengine.JobKindSCMWorkflow,
		IdempotencyKey: "idem-1",
		Step:           "ensure_worktree",
		Status:         taskengine.ExecutionStatusFailed,
		ErrorMessage:   "temporary network issue",
		UpdatedAt:      now.Add(time.Second),
	}, 3, 3)
	if err != nil {
		t.Fatalf("OnExecution(failed) error = %v", err)
	}
	if len(store.events) != 2 {
		t.Fatalf("expected two decisions persisted")
	}
	if store.events[1].Action != domainsupervisor.ActionEscalate {
		t.Fatalf("expected escalated action got %q", store.events[1].Action)
	}
	if store.events[1].Reason != domainsupervisor.ReasonExecutionFailedMaxed {
		t.Fatalf("expected max-retries reason got %q", store.events[1].Reason)
	}
}
