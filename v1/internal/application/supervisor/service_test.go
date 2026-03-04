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

func TestServiceDeterministicTransitionFixtures(t *testing.T) {
	store := &memoryEventStore{}
	service, err := NewService(store, nil)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	t.Run("issue opened blocks", func(t *testing.T) {
		decision, fixtureErr := service.OnIssueOpened(context.Background(), taskengine.CorrelationIDs{RunID: "run-issue", TaskID: "task-issue", JobID: "job-issue"}, "octo/repo", "octo/repo#12")
		if fixtureErr != nil {
			t.Fatalf("OnIssueOpened() error = %v", fixtureErr)
		}
		if decision.Action != domainsupervisor.ActionBlock || decision.Reason != domainsupervisor.ReasonIssueAwaitingApproval {
			t.Fatalf("expected issue-opened block/await-approval, got %q/%q", decision.Action, decision.Reason)
		}
	})

	t.Run("checks failed requests rework", func(t *testing.T) {
		correlation := taskengine.CorrelationIDs{RunID: "run-check-fail", TaskID: "task-check-fail", JobID: "job-check-fail"}
		if _, fixtureErr := service.OnIssueOpened(context.Background(), correlation, "octo/repo", "octo/repo#22"); fixtureErr != nil {
			t.Fatalf("seed blocked state error = %v", fixtureErr)
		}
		decision, fixtureErr := service.OnPRChecksEvaluated(context.Background(), correlation, "github", "octo", "repo", 34, false, "checks_failed")
		if fixtureErr != nil {
			t.Fatalf("OnPRChecksEvaluated(false) error = %v", fixtureErr)
		}
		if decision.Action != domainsupervisor.ActionRequestRework || decision.Reason != domainsupervisor.ReasonPRChecksFailed {
			t.Fatalf("expected checks-failed rework, got %q/%q", decision.Action, decision.Reason)
		}
	})

	t.Run("merge request refused when not ready", func(t *testing.T) {
		correlation := taskengine.CorrelationIDs{RunID: "run-refuse", TaskID: "task-refuse", JobID: "job-refuse"}
		if _, fixtureErr := service.OnIssueOpened(context.Background(), correlation, "octo/repo", "octo/repo#35"); fixtureErr != nil {
			t.Fatalf("seed blocked state error = %v", fixtureErr)
		}
		decision, fixtureErr := service.OnPRMergeRequested(context.Background(), correlation, "github", "octo", "repo", 35, "squash")
		if fixtureErr != nil {
			t.Fatalf("OnPRMergeRequested(not-ready) error = %v", fixtureErr)
		}
		if decision.Action != domainsupervisor.ActionRefuse || decision.Reason != domainsupervisor.ReasonPRMergeRefused {
			t.Fatalf("expected merge-refused, got %q/%q", decision.Action, decision.Reason)
		}
	})

	t.Run("checks passed then merge request merges", func(t *testing.T) {
		correlation := taskengine.CorrelationIDs{RunID: "run-merge", TaskID: "task-merge", JobID: "job-merge"}
		now := time.Now().UTC()
		if _, fixtureErr := service.OnAdmission(context.Background(), taskengine.AdmissionRecord{
			RunID:          correlation.RunID,
			TaskID:         correlation.TaskID,
			JobID:          correlation.JobID,
			JobKind:        taskengine.JobKindSCMWorkflow,
			IdempotencyKey: "idem-merge",
			QueueTaskID:    "qt-merge",
			Queue:          "scm",
			Status:         taskengine.AdmissionStatusQueued,
			EnqueuedAt:     now,
		}); fixtureErr != nil {
			t.Fatalf("seed admitted state error = %v", fixtureErr)
		}
		if _, fixtureErr := service.OnExecution(context.Background(), taskengine.ExecutionRecord{
			RunID:          correlation.RunID,
			TaskID:         correlation.TaskID,
			JobID:          correlation.JobID,
			JobKind:        taskengine.JobKindSCMWorkflow,
			IdempotencyKey: "idem-merge",
			Step:           "checks",
			Status:         taskengine.ExecutionStatusSucceeded,
			UpdatedAt:      now.Add(time.Second),
		}, 1, 3); fixtureErr != nil {
			t.Fatalf("seed completed state error = %v", fixtureErr)
		}
		checksDecision, fixtureErr := service.OnPRChecksEvaluated(context.Background(), correlation, "github", "octo", "repo", 36, true, "merge_ready")
		if fixtureErr != nil {
			t.Fatalf("OnPRChecksEvaluated(true) error = %v", fixtureErr)
		}
		if checksDecision.Action != domainsupervisor.ActionContinue || checksDecision.ToState != domainsupervisor.StateMergeReady {
			t.Fatalf("expected checks-passed continue to merge-ready, got %q/%q", checksDecision.Action, checksDecision.ToState)
		}
		mergeDecision, mergeErr := service.OnPRMergeRequested(context.Background(), correlation, "github", "octo", "repo", 36, "squash")
		if mergeErr != nil {
			t.Fatalf("OnPRMergeRequested(ready) error = %v", mergeErr)
		}
		if mergeDecision.Action != domainsupervisor.ActionMerge || mergeDecision.Reason != domainsupervisor.ReasonPRMergeApproved {
			t.Fatalf("expected merge-approved merge action, got %q/%q", mergeDecision.Action, mergeDecision.Reason)
		}
	})
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
		Step:           "ensure_repository",
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
		Step:           "ensure_repository",
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
