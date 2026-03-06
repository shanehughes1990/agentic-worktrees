package worker

import (
	applicationlifecycle "agentic-orchestrator/internal/application/lifecycle"
	"agentic-orchestrator/internal/application/taskengine"
	domainlifecycle "agentic-orchestrator/internal/domain/lifecycle"
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
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

type fakeLifecycleStoreWithFailure struct {
	events      []domainlifecycle.Event
	failOnEvent domainlifecycle.EventType
}

func (store *fakeLifecycleStoreWithFailure) Append(_ context.Context, event domainlifecycle.Event) (domainlifecycle.Event, error) {
	if event.EventType == store.failOnEvent {
		return domainlifecycle.Event{}, errors.New("persist failed")
	}
	event.EventSeq = int64(len(store.events) + 1)
	event.ProjectEventSeq = int64(len(store.events) + 1)
	store.events = append(store.events, event)
	return event, nil
}

type fakeJobHandler struct {
	err error
}

type fakeRuntimeTransport struct {
	runtimeSignals []domainrealtime.RuntimeActivitySignal
}

func (transport *fakeRuntimeTransport) PublishRequest(context.Context, domainrealtime.HeartbeatRequest) error {
	return nil
}

func (transport *fakeRuntimeTransport) PublishResponse(context.Context, domainrealtime.HeartbeatResponse) error {
	return nil
}

func (transport *fakeRuntimeTransport) ListenRequests(context.Context, func(domainrealtime.HeartbeatRequest) error) error {
	return nil
}

func (transport *fakeRuntimeTransport) ListenResponses(context.Context, func(domainrealtime.HeartbeatResponse) error) error {
	return nil
}

func (transport *fakeRuntimeTransport) PublishRuntimeActivity(_ context.Context, signal domainrealtime.RuntimeActivitySignal) error {
	transport.runtimeSignals = append(transport.runtimeSignals, signal)
	return nil
}

func (transport *fakeRuntimeTransport) ListenRuntimeActivity(context.Context, func(domainrealtime.RuntimeActivitySignal) error) error {
	return nil
}

func (transport *fakeRuntimeTransport) PublishRegistrationSubmission(context.Context, domainrealtime.RegistrationSubmissionEvent) error {
	return nil
}

func (transport *fakeRuntimeTransport) PublishRegistrationDecision(context.Context, domainrealtime.RegistrationDecisionEvent) error {
	return nil
}

func (transport *fakeRuntimeTransport) ListenRegistrationSubmissions(context.Context, func(domainrealtime.RegistrationSubmissionEvent) error) error {
	return nil
}

func (transport *fakeRuntimeTransport) ListenRegistrationDecisions(context.Context, func(domainrealtime.RegistrationDecisionEvent) error) error {
	return nil
}

func (transport *fakeRuntimeTransport) PublishInvalidationIntent(context.Context, domainrealtime.InvalidationIntent) error {
	return nil
}

func (transport *fakeRuntimeTransport) ListenInvalidationIntents(context.Context, func(domainrealtime.InvalidationIntent) error) error {
	return nil
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

func TestJobLifecycleMiddlewarePublishesRuntimeActivitySignals(t *testing.T) {
	store := &fakeLifecycleStore{}
	service, err := applicationlifecycle.NewService(store)
	if err != nil {
		t.Fatalf("new lifecycle service: %v", err)
	}
	transport := &fakeRuntimeTransport{}

	handler := newJobLifecycleMiddleware("worker-1", service, transport, fakeJobHandler{})
	if err := handler.Handle(context.Background(), taskengine.Job{
		Kind:        taskengine.JobKindAgentWorkflow,
		QueueTaskID: "queue-task-live-1",
		Payload:     []byte(`{"project_id":"project-1","run_id":"run-1","task_id":"task-1","job_id":"job-1","session_id":"session-1","idempotency_key":"idemp-live-1"}`),
	}); err != nil {
		t.Fatalf("handle job: %v", err)
	}

	if len(transport.runtimeSignals) < 2 {
		t.Fatalf("expected at least started and completed runtime signals, got %d", len(transport.runtimeSignals))
	}

	var hasStarted bool
	var hasCompleted bool
	for _, signal := range transport.runtimeSignals {
		if signal.EventType == "started" {
			hasStarted = true
		}
		if signal.EventType == "completed" {
			hasCompleted = true
		}
	}
	if !hasStarted {
		t.Fatalf("expected started runtime signal")
	}
	if !hasCompleted {
		t.Fatalf("expected completed runtime signal")
	}
}

func TestJobLifecycleMiddlewarePublishesRuntimeFailedSignalOnHandlerError(t *testing.T) {
	store := &fakeLifecycleStore{}
	service, err := applicationlifecycle.NewService(store)
	if err != nil {
		t.Fatalf("new lifecycle service: %v", err)
	}
	transport := &fakeRuntimeTransport{}

	handler := newJobLifecycleMiddleware("worker-1", service, transport, fakeJobHandler{err: errors.New("boom")})
	err = handler.Handle(context.Background(), taskengine.Job{
		Kind:        taskengine.JobKindAgentWorkflow,
		QueueTaskID: "queue-task-live-2",
		Payload:     []byte(`{"project_id":"project-1","run_id":"run-1","task_id":"task-1","job_id":"job-2","session_id":"session-2","idempotency_key":"idemp-live-2"}`),
	})
	if err == nil {
		t.Fatalf("expected wrapped handler error")
	}

	var hasFailed bool
	for _, signal := range transport.runtimeSignals {
		if signal.EventType == "failed" {
			hasFailed = true
		}
	}
	if !hasFailed {
		t.Fatalf("expected failed runtime signal")
	}
}

func TestJobLifecycleMiddlewarePersistsHeartbeatRuntimeSignals(t *testing.T) {
	store := &fakeLifecycleStore{}
	service, err := applicationlifecycle.NewService(store)
	if err != nil {
		t.Fatalf("new lifecycle service: %v", err)
	}
	transport := &fakeRuntimeTransport{}
	middleware := &jobLifecycleMiddleware{
		workerID:  "worker-1",
		service:   service,
		transport: transport,
	}

	middleware.publishRuntimeSignal(
		context.Background(),
		lifecycleMetadata{
			ProjectID:    "project-1",
			RunID:        "run-1",
			TaskID:       "task-1",
			JobID:        "job-1",
			SessionID:    "session-1",
			PipelineType: string(taskengine.JobKindAgentWorkflow),
		},
		"heartbeat",
		map[string]any{"runtime_alive": true},
		"",
	)

	if len(transport.runtimeSignals) != 1 {
		t.Fatalf("expected one heartbeat runtime signal, got %d", len(transport.runtimeSignals))
	}
	if len(store.events) != 4 {
		t.Fatalf("expected layered heartbeat lifecycle events to be persisted, got %d", len(store.events))
	}
	if store.events[0].EventType != domainlifecycle.EventHeartbeat {
		t.Fatalf("expected persisted heartbeat event type, got %q", store.events[0].EventType)
	}
	hasRuntime := false
	hasProcess := false
	hasActivity := false
	for _, event := range store.events {
		switch event.EventType {
		case domainlifecycle.EventRuntimeHeartbeat:
			hasRuntime = true
		case domainlifecycle.EventProcessHeartbeat:
			hasProcess = true
		case domainlifecycle.EventActivityHeartbeat:
			hasActivity = true
		}
	}
	if !hasRuntime || !hasProcess || !hasActivity {
		t.Fatalf("expected runtime/process/activity layered heartbeat events, got runtime=%t process=%t activity=%t", hasRuntime, hasProcess, hasActivity)
	}
	if store.events[0].SessionID != "session-1" {
		t.Fatalf("expected persisted session_id session-1, got %q", store.events[0].SessionID)
	}
}

func TestJobLifecycleMiddlewareReturnsErrorWhenStartedPersistenceFails(t *testing.T) {
	store := &fakeLifecycleStoreWithFailure{failOnEvent: domainlifecycle.EventStarted}
	service, err := applicationlifecycle.NewService(store)
	if err != nil {
		t.Fatalf("new lifecycle service: %v", err)
	}
	handler := newJobLifecycleMiddleware("worker-1", service, nil, fakeJobHandler{})
	err = handler.Handle(context.Background(), taskengine.Job{
		Kind:        taskengine.JobKindAgentWorkflow,
		QueueTaskID: "queue-task-start-fail",
		Payload:     []byte(`{"project_id":"project-1","session_id":"session-1"}`),
	})
	if err == nil {
		t.Fatalf("expected started persistence error")
	}
	if !strings.Contains(err.Error(), "append lifecycle started event") {
		t.Fatalf("expected started append context in error, got %v", err)
	}
}

func TestJobLifecycleMiddlewareReturnsJoinedErrorWhenFailedPersistenceFails(t *testing.T) {
	store := &fakeLifecycleStoreWithFailure{failOnEvent: domainlifecycle.EventFailed}
	service, err := applicationlifecycle.NewService(store)
	if err != nil {
		t.Fatalf("new lifecycle service: %v", err)
	}
	handler := newJobLifecycleMiddleware("worker-1", service, nil, fakeJobHandler{err: errors.New("handler boom")})
	err = handler.Handle(context.Background(), taskengine.Job{
		Kind:        taskengine.JobKindAgentWorkflow,
		QueueTaskID: "queue-task-failed-persist",
		Payload:     []byte(`{"project_id":"project-1","session_id":"session-1"}`),
	})
	if err == nil {
		t.Fatalf("expected joined handler/persistence error")
	}
	message := err.Error()
	if !strings.Contains(message, "handler boom") {
		t.Fatalf("expected handler error in joined error, got %v", err)
	}
	if !strings.Contains(message, "append lifecycle failed event") {
		t.Fatalf("expected failed append context in joined error, got %v", err)
	}
}
