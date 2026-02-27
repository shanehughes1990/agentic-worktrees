package taskengine

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeEngine struct {
	lastRequest EnqueueRequest
	result      EnqueueResult
	err         error
}

func (engine *fakeEngine) Enqueue(_ context.Context, request EnqueueRequest) (EnqueueResult, error) {
	engine.lastRequest = request
	return engine.result, engine.err
}

func TestSchedulerEnqueueRejectsMissingIdempotencyForIngestion(t *testing.T) {
	engine := &fakeEngine{}
	scheduler, err := NewScheduler(engine, DefaultPolicies())
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}

	_, enqueueErr := scheduler.Enqueue(context.Background(), EnqueueRequest{
		Kind:      JobKindIngestionAgent,
		Payload:   []byte(`{"run_id":"run-1","prompt":"hello"}`),
		UniqueFor: 5 * time.Minute,
	})
	if enqueueErr == nil {
		t.Fatalf("expected enqueue error")
	}
	if !errors.Is(enqueueErr, ErrInvalidEnqueueRequest) {
		t.Fatalf("expected ErrInvalidEnqueueRequest, got %v", enqueueErr)
	}
}

func TestSchedulerEnqueueAppliesPolicyDefaults(t *testing.T) {
	engine := &fakeEngine{result: EnqueueResult{QueueTaskID: "queue-task-1"}}
	scheduler, err := NewScheduler(engine, DefaultPolicies())
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}

	result, enqueueErr := scheduler.Enqueue(context.Background(), EnqueueRequest{
		Kind:           JobKindIngestionAgent,
		Payload:        []byte(`{"run_id":"run-1","prompt":"hello"}`),
		IdempotencyKey: "run-1",
	})
	if enqueueErr != nil {
		t.Fatalf("enqueue: %v", enqueueErr)
	}
	if result.QueueTaskID != "queue-task-1" {
		t.Fatalf("expected queue task id queue-task-1, got %s", result.QueueTaskID)
	}
	if engine.lastRequest.Queue != "ingestion" {
		t.Fatalf("expected default queue ingestion, got %s", engine.lastRequest.Queue)
	}
	if engine.lastRequest.UniqueFor <= 0 {
		t.Fatalf("expected default unique_for to be applied")
	}
	if engine.lastRequest.Timeout <= 0 {
		t.Fatalf("expected default timeout to be applied")
	}
	if engine.lastRequest.MaxRetry != 2 {
		t.Fatalf("expected max retry default 2, got %d", engine.lastRequest.MaxRetry)
	}
}
