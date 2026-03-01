package postgres

import (
	"agentic-orchestrator/internal/application/controlplane"
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"testing"
	"time"
)

func TestControlPlaneQueryRepositoryListsPersistedReadModels(t *testing.T) {
	db := newTestDB(t)
	repository, err := NewControlPlaneQueryRepository(db)
	if err != nil {
		t.Fatalf("new control-plane query repository: %v", err)
	}
	ledger, err := NewAdmissionLedger(db)
	if err != nil {
		t.Fatalf("new admission ledger: %v", err)
	}
	if err := ledger.Upsert(context.Background(), taskengine.AdmissionRecord{
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		JobKind:        taskengine.JobKindIngestionAgent,
		IdempotencyKey: "idem-1",
		QueueTaskID:    "queue-1",
		Queue:          "ingestion",
		Status:         taskengine.AdmissionStatusQueued,
		Duplicate:      false,
		EnqueuedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("upsert admission record: %v", err)
	}
	workerRegistry, err := NewWorkerRegistry(db)
	if err != nil {
		t.Fatalf("new worker registry: %v", err)
	}
	if err := workerRegistry.Upsert(context.Background(), taskengine.WorkerCapabilityAdvertisement{WorkerID: "worker-1", Capabilities: []taskengine.WorkerCapability{{Kind: taskengine.JobKindSCMWorkflow}}}); err != nil {
		t.Fatalf("upsert worker registry: %v", err)
	}
	executionJournal, err := NewPostgresExecutionJournal(db)
	if err != nil {
		t.Fatalf("new execution journal: %v", err)
	}
	if err := executionJournal.Upsert(context.Background(), taskengine.ExecutionRecord{
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		JobKind:        taskengine.JobKindIngestionAgent,
		IdempotencyKey: "idem-1",
		Step:           "sync",
		Status:         taskengine.ExecutionStatusSucceeded,
		UpdatedAt:      time.Now().UTC(),
	}); err != nil {
		t.Fatalf("upsert execution record: %v", err)
	}
	audit, err := NewDeadLetterAudit(db)
	if err != nil {
		t.Fatalf("new dead-letter audit: %v", err)
	}
	if err := audit.Record(context.Background(), taskengine.DeadLetterEvent{Queue: "scm", TaskID: "task-archive-1", JobKind: taskengine.JobKindSCMWorkflow, Action: taskengine.DeadLetterActionRequeue, OccurredAt: time.Now().UTC()}); err != nil {
		t.Fatalf("record dead-letter event: %v", err)
	}

	sessions, err := repository.ListSessions(context.Background(), 10)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 1 || sessions[0].RunID != "run-1" {
		t.Fatalf("unexpected sessions: %+v", sessions)
	}
	jobs, err := repository.ListWorkflowJobs(context.Background(), "run-1", "task-1", 10)
	if err != nil {
		t.Fatalf("list workflow jobs: %v", err)
	}
	if len(jobs) != 1 || jobs[0].QueueTaskID != "queue-1" {
		t.Fatalf("unexpected workflow jobs: %+v", jobs)
	}
	workers, err := repository.ListWorkers(context.Background(), 10)
	if err != nil {
		t.Fatalf("list workers: %v", err)
	}
	if len(workers) != 1 || workers[0].WorkerID != "worker-1" {
		t.Fatalf("unexpected workers: %+v", workers)
	}
	executionHistory, err := repository.ListExecutionHistory(context.Background(), controlplane.CorrelationFilter{RunID: "run-1", TaskID: "task-1", JobID: "job-1"}, 10)
	if err != nil {
		t.Fatalf("list execution history: %v", err)
	}
	if len(executionHistory) != 1 || executionHistory[0].Step != "sync" {
		t.Fatalf("unexpected execution history: %+v", executionHistory)
	}
	deadLetters, err := repository.ListDeadLetterHistory(context.Background(), "scm", 10)
	if err != nil {
		t.Fatalf("list dead-letter history: %v", err)
	}
	if len(deadLetters) != 1 || deadLetters[0].TaskID != "task-archive-1" {
		t.Fatalf("unexpected dead-letter history: %+v", deadLetters)
	}
}
