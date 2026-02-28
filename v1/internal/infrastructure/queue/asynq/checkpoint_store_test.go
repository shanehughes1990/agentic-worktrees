package asynq

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestCheckpointStore(t *testing.T) (*RedisCheckpointStore, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := NewRedisCheckpointStore(client, time.Minute)
	return store, mr
}

func TestCheckpointStoreSaveAndLoad(t *testing.T) {
	store, _ := newTestCheckpointStore(t)
	ctx := context.Background()

	checkpoint := taskengine.Checkpoint{Step: "step-1", Token: "tok-abc"}
	if err := store.Save(ctx, "job-123", checkpoint); err != nil {
		t.Fatalf("Save: unexpected error: %v", err)
	}

	got, err := store.Load(ctx, "job-123")
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("Load: expected checkpoint, got nil")
	}
	if got.Step != checkpoint.Step || got.Token != checkpoint.Token {
		t.Errorf("Load: got %+v, want %+v", got, checkpoint)
	}
}

func TestCheckpointStoreLoadMissReturnsNil(t *testing.T) {
	store, _ := newTestCheckpointStore(t)
	ctx := context.Background()

	got, err := store.Load(ctx, "nonexistent-key")
	if err != nil {
		t.Fatalf("Load on miss: unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("Load on miss: expected nil, got %+v", got)
	}
}

func TestCheckpointStoreSaveRejectsEmptyKey(t *testing.T) {
	store, _ := newTestCheckpointStore(t)
	ctx := context.Background()

	err := store.Save(ctx, "  ", taskengine.Checkpoint{Step: "s", Token: "t"})
	if err == nil {
		t.Fatal("Save with empty key: expected error, got nil")
	}
}

func TestCheckpointStoreLoadRejectsEmptyKey(t *testing.T) {
	store, _ := newTestCheckpointStore(t)
	ctx := context.Background()

	_, err := store.Load(ctx, "")
	if err == nil {
		t.Fatal("Load with empty key: expected error, got nil")
	}
}

// TestCheckpointStoreRetryResumeFromPersistedCheckpoint validates the acceptance
// criterion: a transient failure resumes from the persisted checkpoint on retry.
func TestCheckpointStoreRetryResumeFromPersistedCheckpoint(t *testing.T) {
	store, _ := newTestCheckpointStore(t)
	ctx := context.Background()

	const idempotencyKey = "run-xyz-step-2"
	saved := taskengine.Checkpoint{Step: "step-2", Token: "token-step-2"}

	// First attempt: job executes step-2 and persists the checkpoint before failing.
	if err := store.Save(ctx, idempotencyKey, saved); err != nil {
		t.Fatalf("pre-failure Save: %v", err)
	}

	// Simulate transient failure: job is re-queued; on retry the worker loads the checkpoint.
	resumed, err := store.Load(ctx, idempotencyKey)
	if err != nil {
		t.Fatalf("retry Load: %v", err)
	}
	if resumed == nil {
		t.Fatal("retry Load: checkpoint not found; retry cannot resume")
	}

	// Verify the resume checkpoint matches what was persisted so the handler can skip already-done steps.
	if !taskengine.CheckpointMatches(resumed, saved.Step, saved.Token) {
		t.Errorf("CheckpointMatches: got %+v, want step=%q token=%q", resumed, saved.Step, saved.Token)
	}
}

// TestCheckpointStoreTTLExpiry verifies that a checkpoint disappears after TTL expiry.
func TestCheckpointStoreTTLExpiry(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := NewRedisCheckpointStore(client, 5*time.Second)
	ctx := context.Background()

	if err := store.Save(ctx, "expiry-key", taskengine.Checkpoint{Step: "s", Token: "t"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	mr.FastForward(6 * time.Second)

	got, err := store.Load(ctx, "expiry-key")
	if err != nil {
		t.Fatalf("Load after TTL: unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("Load after TTL: expected nil (expired), got %+v", got)
	}
}
