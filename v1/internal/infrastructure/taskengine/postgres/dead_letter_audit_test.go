package postgres

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"testing"
	"time"
)

func TestDeadLetterAuditRecord(t *testing.T) {
	audit, err := NewDeadLetterAudit(newTestDB(t))
	if err != nil {
		t.Fatalf("new dead-letter audit: %v", err)
	}
	event := taskengine.DeadLetterEvent{
		Queue:      "scm",
		TaskID:     "task-1",
		JobKind:    taskengine.JobKindSCMWorkflow,
		Action:     taskengine.DeadLetterActionRequeue,
		LastError:  "retry exhausted",
		Reason:     "manual triage",
		Actor:      "system",
		OccurredAt: time.Now().UTC(),
	}
	if err := audit.Record(context.Background(), event); err != nil {
		t.Fatalf("record: %v", err)
	}
	var count int64
	if err := audit.db.Model(&deadLetterEventRecord{}).Where("queue = ? AND task_id = ?", "scm", "task-1").Count(&count).Error; err != nil {
		t.Fatalf("count audit records: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one dead-letter event, got %d", count)
	}
}
