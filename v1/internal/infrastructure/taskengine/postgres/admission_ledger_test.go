package postgres

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"testing"
	"time"
)

func TestAdmissionLedgerUpsert(t *testing.T) {
	ledger, err := NewAdmissionLedger(newTestDB(t))
	if err != nil {
		t.Fatalf("new admission ledger: %v", err)
	}
	record := taskengine.AdmissionRecord{
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		JobKind:        taskengine.JobKindSCMWorkflow,
		IdempotencyKey: "idem-1",
		QueueTaskID:    "queue-1",
		Queue:          "scm",
		Status:         taskengine.AdmissionStatusQueued,
		EnqueuedAt:     time.Now().UTC(),
	}
	if err := ledger.Upsert(context.Background(), record); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	var count int64
	if err := ledger.db.Model(&admissionLedgerRecord{}).Where("run_id = ? AND task_id = ? AND job_id = ?", "run-1", "task-1", "job-1").Count(&count).Error; err != nil {
		t.Fatalf("count admission records: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one admission record, got %d", count)
	}
}
