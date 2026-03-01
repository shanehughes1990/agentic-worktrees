package postgres

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"testing"
	"time"
)

func TestPostgresExecutionJournalUpsertAndLoad(t *testing.T) {
	journal, err := NewPostgresExecutionJournal(newTestDB(t))
	if err != nil {
		t.Fatalf("new journal: %v", err)
	}
	record := taskengine.ExecutionRecord{
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		JobKind:        taskengine.JobKindSCMWorkflow,
		IdempotencyKey: "id-1",
		Step:           "source_state",
		Status:         taskengine.ExecutionStatusRunning,
		UpdatedAt:      time.Now().UTC(),
	}
	if err := journal.Upsert(context.Background(), record); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	loaded, err := journal.Load(context.Background(), "run-1", "task-1", "job-1")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded == nil {
		t.Fatalf("expected record")
	}
	if loaded.Step != record.Step || loaded.Status != record.Status {
		t.Fatalf("unexpected loaded record: %+v", loaded)
	}
}

func TestPostgresExecutionJournalLoadMissReturnsNil(t *testing.T) {
	journal, err := NewPostgresExecutionJournal(newTestDB(t))
	if err != nil {
		t.Fatalf("new journal: %v", err)
	}
	loaded, err := journal.Load(context.Background(), "run-x", "task-x", "job-x")
	if err != nil {
		t.Fatalf("load miss: %v", err)
	}
	if loaded != nil {
		t.Fatalf("expected nil result, got %+v", loaded)
	}
}
