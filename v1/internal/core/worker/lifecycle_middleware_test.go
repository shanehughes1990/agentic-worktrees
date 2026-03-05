package worker

import (
	applicationlifecycle "agentic-orchestrator/internal/application/lifecycle"
	"agentic-orchestrator/internal/application/taskengine"
	domainlifecycle "agentic-orchestrator/internal/domain/lifecycle"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeLifecycleStore struct {
	events []domainlifecycle.Event
}

func (store *fakeLifecycleStore) Append(_ context.Context, event domainlifecycle.Event) (domainlifecycle.Event, error) {
	event.EventSeq = int64(len(store.events) + 1)
	event.ProjectEventSeq = int64(len(store.events) + 1)
	store.events = append(store.events, event)
	return event, nil
}

type fakeJobHandler struct {
	err error
}

func (handler fakeJobHandler) Handle(ctx context.Context, job taskengine.Job) error {
	_ = ctx
	_ = job
	return handler.err
}

func TestJobLifecycleMiddlewareEmitsStartedAndCompleted(t *testing.T) {
	store := &fakeLifecycleStore{}
	service, err := applicationlifecycle.NewService(store)
	if err != nil {
		t.Fatalf("new lifecycle service: %v", err)
	}
	handler := newJobLifecycleMiddleware("worker-1", service, nil, fakeJobHandler{})
	if err := handler.Handle(context.Background(), taskengine.Job{
		Kind:        taskengine.JobKindAgentWorkflow,
		QueueTaskID: "queue-task-1",
		Payload:     []byte(`{"project_id":"project-1","run_id":"run-1","task_id":"task-1","job_id":"job-1","session_id":"session-1","idempotency_key":"idemp-1"}`),
	}); err != nil {
		t.Fatalf("handle job: %v", err)
	}
	if len(store.events) != 2 {
		t.Fatalf("expected 2 lifecycle events, got %d", len(store.events))
	}
	if store.events[0].EventType != domainlifecycle.EventStarted {
		t.Fatalf("expected first event type started, got %s", store.events[0].EventType)
	}
	if store.events[1].EventType != domainlifecycle.EventCompleted {
		t.Fatalf("expected second event type completed, got %s", store.events[1].EventType)
	}
	if store.events[0].ProjectID != "project-1" {
		t.Fatalf("expected project_id to be propagated")
	}
	if store.events[0].SessionID != "session-1" {
		t.Fatalf("expected session_id to be propagated")
	}
}

func TestJobLifecycleMiddlewareEmitsFailedOnHandlerError(t *testing.T) {
	store := &fakeLifecycleStore{}
	service, err := applicationlifecycle.NewService(store)
	if err != nil {
		t.Fatalf("new lifecycle service: %v", err)
	}
	handler := newJobLifecycleMiddleware("worker-1", service, nil, fakeJobHandler{err: errors.New("boom")})
	err = handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, QueueTaskID: "queue-task-2", Payload: []byte(`{"project_id":"project-2"}`)})
	if err == nil {
		t.Fatalf("expected error from wrapped handler")
	}
	if len(store.events) != 2 {
		t.Fatalf("expected 2 lifecycle events, got %d", len(store.events))
	}
	if store.events[1].EventType != domainlifecycle.EventFailed {
		t.Fatalf("expected failed event type, got %s", store.events[1].EventType)
	}
}

func TestBuildSyntheticSessionIDDeterministic(t *testing.T) {
	first := buildSyntheticSessionID(taskengine.JobKindPromptRefinementAgent, "project-1", "", "", "job-1", "")
	second := buildSyntheticSessionID(taskengine.JobKindPromptRefinementAgent, "project-1", "", "", "job-1", "")
	if first != second {
		t.Fatalf("expected deterministic synthetic session id")
	}
	if first == "" {
		t.Fatalf("expected non-empty synthetic session id")
	}
	_ = time.Now()
}

func TestJobLifecycleMiddlewareUsesAttemptAwareEventIDs(t *testing.T) {
	store := &fakeLifecycleStore{}
	service, err := applicationlifecycle.NewService(store)
	if err != nil {
		t.Fatalf("new lifecycle service: %v", err)
	}
	handler := newJobLifecycleMiddleware("worker-1", service, nil, fakeJobHandler{err: errors.New("timeout")})

	baseJob := taskengine.Job{
		Kind:        taskengine.JobKindIngestionAgent,
		QueueTaskID: "queue-task-42",
		MaxRetry:    2,
		Payload:     []byte(`{"project_id":"project-1","run_id":"run-1","task_id":"task-1","job_id":"job-1","session_id":"session-1","idempotency_key":"idemp-42"}`),
	}
	if err := handler.Handle(context.Background(), baseJob); err == nil {
		t.Fatalf("expected timeout error on attempt 0")
	}
	baseJob.RetryCount = 1
	if err := handler.Handle(context.Background(), baseJob); err == nil {
		t.Fatalf("expected timeout error on attempt 1")
	}

	if len(store.events) != 4 {
		t.Fatalf("expected 4 lifecycle events across two attempts, got %d", len(store.events))
	}
	if store.events[0].EventID == store.events[2].EventID {
		t.Fatalf("expected started event IDs to differ across attempts, got %q", store.events[0].EventID)
	}
	if store.events[1].EventID == store.events[3].EventID {
		t.Fatalf("expected failed event IDs to differ across attempts, got %q", store.events[1].EventID)
	}
	if !strings.Contains(store.events[0].EventID, "attempt-0") || !strings.Contains(store.events[2].EventID, "attempt-1") {
		t.Fatalf("expected attempt suffixes in event IDs, got %q and %q", store.events[0].EventID, store.events[2].EventID)
	}
}
