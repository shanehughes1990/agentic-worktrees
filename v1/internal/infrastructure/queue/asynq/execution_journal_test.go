package asynq

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestExecutionJournal(t *testing.T) (*RedisExecutionJournal, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	journal := NewRedisExecutionJournal(client, time.Minute)
	return journal, mr
}

func TestExecutionJournalUpsertAndLoad(t *testing.T) {
	journal, _ := newTestExecutionJournal(t)
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
		t.Fatalf("expected execution record")
	}
	if loaded.Step != record.Step || loaded.Status != record.Status {
		t.Fatalf("unexpected loaded record: %+v", loaded)
	}
}

func TestExecutionJournalLoadMissReturnsNil(t *testing.T) {
	journal, _ := newTestExecutionJournal(t)
	loaded, err := journal.Load(context.Background(), "run-x", "task-x", "job-x")
	if err != nil {
		t.Fatalf("load miss: %v", err)
	}
	if loaded != nil {
		t.Fatalf("expected nil load result, got %+v", loaded)
	}
}
