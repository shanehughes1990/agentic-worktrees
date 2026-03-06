package postgres

import (
	"agentic-orchestrator/internal/domain/failures"
	domainlifecycle "agentic-orchestrator/internal/domain/lifecycle"
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newLifecycleTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func TestEventStoreAppendAssignsSessionAndProjectSequences(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}

	first, err := store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "evt-1",
		SchemaVersion: 1,
		ProjectID:     "project-1",
		RunID:         "run-1",
		TaskID:        "task-1",
		JobID:         "job-1",
		SessionID:     "session-1",
		WorkerID:      "worker-1",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventStarted,
		OccurredAt:    time.Now().UTC(),
		Payload:       map[string]any{"state": "started"},
	})
	if err != nil {
		t.Fatalf("append first event: %v", err)
	}
	if first.EventSeq != 1 {
		t.Fatalf("expected session event_seq=1, got %d", first.EventSeq)
	}
	if first.ProjectEventSeq != 1 {
		t.Fatalf("expected project_event_seq=1, got %d", first.ProjectEventSeq)
	}

	second, err := store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "evt-2",
		SchemaVersion: 1,
		ProjectID:     "project-1",
		RunID:         "run-1",
		TaskID:        "task-1",
		JobID:         "job-1",
		SessionID:     "session-1",
		WorkerID:      "worker-1",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventCompleted,
		OccurredAt:    time.Now().UTC(),
		Payload:       map[string]any{"state": "completed"},
	})
	if err != nil {
		t.Fatalf("append second event: %v", err)
	}
	if second.EventSeq != 2 {
		t.Fatalf("expected session event_seq=2, got %d", second.EventSeq)
	}
	if second.ProjectEventSeq != 2 {
		t.Fatalf("expected project_event_seq=2, got %d", second.ProjectEventSeq)
	}

	snapshot, err := loadSnapshotBySessionID(store, "session-1")
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	if snapshot == nil {
		t.Fatalf("expected snapshot to exist")
	}
	if snapshot.LastEventSeq != 2 {
		t.Fatalf("expected snapshot last_event_seq=2, got %d", snapshot.LastEventSeq)
	}
	if snapshot.LastProjectEventSeq != 2 {
		t.Fatalf("expected snapshot last_project_event_seq=2, got %d", snapshot.LastProjectEventSeq)
	}
	if snapshot.CurrentState != "completed" {
		t.Fatalf("expected snapshot state completed, got %q", snapshot.CurrentState)
	}
	if snapshot.EndedAt == nil {
		t.Fatalf("expected ended_at to be set on completed event")
	}
}

func TestEventStoreAppendIsIdempotentByEventID(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}
	occurredAt := time.Now().UTC()
	first, err := store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "evt-idempotent",
		SchemaVersion: 1,
		ProjectID:     "project-2",
		RunID:         "run-2",
		TaskID:        "task-2",
		JobID:         "job-2",
		SessionID:     "session-2",
		WorkerID:      "worker-2",
		SourceRuntime: "worker",
		PipelineType:  "scm.workflow.run",
		EventType:     domainlifecycle.EventStarted,
		OccurredAt:    occurredAt,
		Payload:       map[string]any{"state": "started"},
	})
	if err != nil {
		t.Fatalf("append first event: %v", err)
	}
	second, err := store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "evt-idempotent",
		SchemaVersion: 1,
		ProjectID:     "project-2",
		RunID:         "run-2",
		TaskID:        "task-2",
		JobID:         "job-2",
		SessionID:     "session-2",
		WorkerID:      "worker-2",
		SourceRuntime: "worker",
		PipelineType:  "scm.workflow.run",
		EventType:     domainlifecycle.EventStarted,
		OccurredAt:    occurredAt,
		Payload:       map[string]any{"state": "started again"},
	})
	if err != nil {
		t.Fatalf("append duplicate event: %v", err)
	}
	if first.EventSeq != second.EventSeq || first.ProjectEventSeq != second.ProjectEventSeq {
		t.Fatalf("expected duplicate event to return same sequences: first(%d,%d) second(%d,%d)", first.EventSeq, first.ProjectEventSeq, second.EventSeq, second.ProjectEventSeq)
	}

	snapshot, err := loadSnapshotBySessionID(store, "session-2")
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	if snapshot == nil {
		t.Fatalf("expected snapshot to exist")
	}
	if snapshot.LastEventSeq != 1 {
		t.Fatalf("expected snapshot last_event_seq=1 for idempotent duplicate, got %d", snapshot.LastEventSeq)
	}
}

func TestEventStoreAppendRedactsSensitivePayloadValues(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}
	appended, err := store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "evt-redaction-lifecycle",
		SchemaVersion: 1,
		ProjectID:     "project-redaction",
		RunID:         "run-redaction",
		TaskID:        "task-redaction",
		JobID:         "job-redaction",
		SessionID:     "session-redaction",
		WorkerID:      "worker-redaction",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventStarted,
		OccurredAt:    time.Now().UTC(),
		Payload: map[string]any{
			"api_key": "plain-api-key",
			"nested": map[string]any{
				"password": "plain-password",
			},
		},
	})
	if err != nil {
		t.Fatalf("append lifecycle event: %v", err)
	}
	if appended.Payload["api_key"] == "plain-api-key" {
		t.Fatalf("expected api_key to be redacted")
	}
	nested, ok := appended.Payload["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested payload map")
	}
	if nested["password"] == "plain-password" {
		t.Fatalf("expected nested password to be redacted")
	}
	if appended.Payload["_redaction_policy_version"] != "v1" {
		t.Fatalf("expected redaction policy version v1")
	}
}

func TestEventStoreAppendEmitsGapDetectedAndGapReconciled(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}
	now := time.Now().UTC()

	_, err = store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "gap-seed",
		SchemaVersion: 1,
		ProjectID:     "project-gap",
		RunID:         "run-gap",
		TaskID:        "task-gap",
		JobID:         "job-gap",
		SessionID:     "session-gap",
		WorkerID:      "worker-gap",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventStarted,
		OccurredAt:    now,
		EventSeq:      1,
		Payload:       map[string]any{"state": "seed"},
	})
	if err != nil {
		t.Fatalf("append seed event: %v", err)
	}

	gapDetectedEvent, err := store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "gap-detected-source",
		SchemaVersion: 1,
		ProjectID:     "project-gap",
		RunID:         "run-gap",
		TaskID:        "task-gap",
		JobID:         "job-gap",
		SessionID:     "session-gap",
		WorkerID:      "worker-gap",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventStarted,
		OccurredAt:    now.Add(1 * time.Second),
		EventSeq:      4,
		Payload:       map[string]any{"state": "future"},
	})
	if err != nil {
		t.Fatalf("append gap-detected source event: %v", err)
	}
	if !payloadBool(gapDetectedEvent.Payload, "gap_detected") {
		t.Fatalf("expected gap_detected payload flag")
	}
	if payloadBool(gapDetectedEvent.Payload, "gap_reconciled") {
		t.Fatalf("did not expect gap_reconciled on detected source event")
	}

	gapReconciledEvent, err := store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "gap-reconcile-source",
		SchemaVersion: 1,
		ProjectID:     "project-gap",
		RunID:         "run-gap",
		TaskID:        "task-gap",
		JobID:         "job-gap",
		SessionID:     "session-gap",
		WorkerID:      "worker-gap",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventCompleted,
		OccurredAt:    now.Add(2 * time.Second),
		EventSeq:      2,
		Payload:       map[string]any{"state": "late"},
	})
	if err != nil {
		t.Fatalf("append gap-reconciled source event: %v", err)
	}
	if payloadBool(gapReconciledEvent.Payload, "gap_detected") {
		t.Fatalf("did not expect gap_detected on reconciled source event")
	}
	if !payloadBool(gapReconciledEvent.Payload, "gap_reconciled") {
		t.Fatalf("expected gap_reconciled payload flag")
	}

	history, err := loadHistoryBySessionID(store, "session-gap")
	if err != nil {
		t.Fatalf("load history: %v", err)
	}
	if len(history) < 5 {
		t.Fatalf("expected at least 5 history records with synthetic gap events, got %d", len(history))
	}

	var hasGapDetected bool
	var hasGapReconciled bool
	for _, record := range history {
		if strings.TrimSpace(record.EventType) == string(domainlifecycle.EventGapDetected) {
			hasGapDetected = true
		}
		if strings.TrimSpace(record.EventType) == string(domainlifecycle.EventGapReconciled) {
			hasGapReconciled = true
		}
	}
	if !hasGapDetected {
		t.Fatalf("expected synthetic gap_detected history event")
	}
	if !hasGapReconciled {
		t.Fatalf("expected synthetic gap_reconciled history event")
	}
}

func TestEventStoreAppendEmitsHeartbeatQuorumTransitions(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}
	now := time.Now().UTC()

	_, err = store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "heartbeat-quorum-seed",
		SchemaVersion: 1,
		ProjectID:     "project-heartbeat-quorum",
		RunID:         "run-heartbeat-quorum",
		TaskID:        "task-heartbeat-quorum",
		JobID:         "job-heartbeat-quorum",
		SessionID:     "session-heartbeat-quorum",
		WorkerID:      "worker-heartbeat-quorum",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventHeartbeat,
		OccurredAt:    now,
		Payload: map[string]any{
			"heartbeat_quorum_state":     "running_ok",
			"heartbeat_confidence_score": 90,
		},
	})
	if err != nil {
		t.Fatalf("append quorum seed: %v", err)
	}

	_, err = store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "heartbeat-quorum-degrade",
		SchemaVersion: 1,
		ProjectID:     "project-heartbeat-quorum",
		RunID:         "run-heartbeat-quorum",
		TaskID:        "task-heartbeat-quorum",
		JobID:         "job-heartbeat-quorum",
		SessionID:     "session-heartbeat-quorum",
		WorkerID:      "worker-heartbeat-quorum",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventActivityHeartbeat,
		OccurredAt:    now.Add(10 * time.Second),
		Payload: map[string]any{
			"heartbeat_quorum_state":     "running_degraded",
			"heartbeat_confidence_score": 55,
		},
	})
	if err != nil {
		t.Fatalf("append quorum degrade: %v", err)
	}

	_, err = store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "heartbeat-quorum-recover",
		SchemaVersion: 1,
		ProjectID:     "project-heartbeat-quorum",
		RunID:         "run-heartbeat-quorum",
		TaskID:        "task-heartbeat-quorum",
		JobID:         "job-heartbeat-quorum",
		SessionID:     "session-heartbeat-quorum",
		WorkerID:      "worker-heartbeat-quorum",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventHeartbeat,
		OccurredAt:    now.Add(20 * time.Second),
		Payload: map[string]any{
			"heartbeat_quorum_state":     "running_ok",
			"heartbeat_confidence_score": 88,
		},
	})
	if err != nil {
		t.Fatalf("append quorum recover: %v", err)
	}

	history, err := loadHistoryBySessionID(store, "session-heartbeat-quorum")
	if err != nil {
		t.Fatalf("load history: %v", err)
	}

	var hasDegraded bool
	var hasRecovered bool
	for _, record := range history {
		if strings.TrimSpace(record.EventType) == string(domainlifecycle.EventHeartbeatQuorumDegraded) {
			hasDegraded = true
		}
		if strings.TrimSpace(record.EventType) == string(domainlifecycle.EventHeartbeatQuorumRecovered) {
			hasRecovered = true
		}
	}
	if !hasDegraded {
		t.Fatalf("expected synthetic heartbeat_quorum_degraded history event")
	}
	if !hasRecovered {
		t.Fatalf("expected synthetic heartbeat_quorum_recovered history event")
	}
}

func TestEventStoreAppendDebouncesHeartbeatQuorumTransitions(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}
	now := time.Now().UTC()

	_, err = store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "heartbeat-quorum-debounce-seed",
		SchemaVersion: 1,
		ProjectID:     "project-heartbeat-debounce",
		RunID:         "run-heartbeat-debounce",
		TaskID:        "task-heartbeat-debounce",
		JobID:         "job-heartbeat-debounce",
		SessionID:     "session-heartbeat-debounce",
		WorkerID:      "worker-heartbeat-debounce",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventHeartbeat,
		OccurredAt:    now,
		Payload: map[string]any{
			"heartbeat_quorum_state":     "running_ok",
			"heartbeat_confidence_score": 90,
		},
	})
	if err != nil {
		t.Fatalf("append debounce seed: %v", err)
	}

	_, err = store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "heartbeat-quorum-debounce-fast",
		SchemaVersion: 1,
		ProjectID:     "project-heartbeat-debounce",
		RunID:         "run-heartbeat-debounce",
		TaskID:        "task-heartbeat-debounce",
		JobID:         "job-heartbeat-debounce",
		SessionID:     "session-heartbeat-debounce",
		WorkerID:      "worker-heartbeat-debounce",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventActivityHeartbeat,
		OccurredAt:    now.Add(1 * time.Second),
		Payload: map[string]any{
			"heartbeat_quorum_state":     "running_degraded",
			"heartbeat_confidence_score": 55,
		},
	})
	if err != nil {
		t.Fatalf("append debounce fast transition: %v", err)
	}

	history, err := loadHistoryBySessionID(store, "session-heartbeat-debounce")
	if err != nil {
		t.Fatalf("load history: %v", err)
	}

	for _, record := range history {
		if strings.TrimSpace(record.EventType) == string(domainlifecycle.EventHeartbeatQuorumDegraded) {
			t.Fatalf("did not expect heartbeat_quorum_degraded transition within debounce window")
		}
		if strings.TrimSpace(record.EventType) == string(domainlifecycle.EventHeartbeatQuorumRecovered) {
			t.Fatalf("did not expect heartbeat_quorum_recovered transition within debounce window")
		}
	}
}

func TestEventStoreAppendPublishesSyntheticQuorumTransitionSignals(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}
	watcher := &fakeTableChangeWatcher{}
	store.SetTableChangeWatcher(watcher)
	now := time.Now().UTC()

	_, err = store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "heartbeat-watch-seed",
		SchemaVersion: 1,
		ProjectID:     "project-heartbeat-watch",
		RunID:         "run-heartbeat-watch",
		TaskID:        "task-heartbeat-watch",
		JobID:         "job-heartbeat-watch",
		SessionID:     "session-heartbeat-watch",
		WorkerID:      "worker-heartbeat-watch",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventHeartbeat,
		OccurredAt:    now,
		Payload: map[string]any{
			"heartbeat_quorum_state":     "running_ok",
			"heartbeat_confidence_score": 90,
		},
	})
	if err != nil {
		t.Fatalf("append seed event: %v", err)
	}

	_, err = store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "heartbeat-watch-degrade",
		SchemaVersion: 1,
		ProjectID:     "project-heartbeat-watch",
		RunID:         "run-heartbeat-watch",
		TaskID:        "task-heartbeat-watch",
		JobID:         "job-heartbeat-watch",
		SessionID:     "session-heartbeat-watch",
		WorkerID:      "worker-heartbeat-watch",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventActivityHeartbeat,
		OccurredAt:    now.Add(10 * time.Second),
		Payload: map[string]any{
			"heartbeat_quorum_state":     "running_degraded",
			"heartbeat_confidence_score": 55,
		},
	})
	if err != nil {
		t.Fatalf("append degraded event: %v", err)
	}

	var hasPrimaryActivityHistorySignal bool
	var hasSyntheticQuorumHistorySignal bool
	for _, published := range watcher.published {
		if published.Topic != "lifecycle_project_session_history" || published.Payload == nil {
			continue
		}
		typeName, _ := published.Payload["event_type"].(string)
		if typeName == string(domainlifecycle.EventActivityHeartbeat) {
			hasPrimaryActivityHistorySignal = true
		}
		if typeName == string(domainlifecycle.EventHeartbeatQuorumDegraded) {
			hasSyntheticQuorumHistorySignal = true
		}
	}
	if !hasPrimaryActivityHistorySignal {
		t.Fatalf("expected history watcher signal for primary activity heartbeat")
	}
	if !hasSyntheticQuorumHistorySignal {
		t.Fatalf("expected history watcher signal for synthetic heartbeat_quorum_degraded event")
	}
}

func TestEventStoreAppendPublishesTableChangeSignalsWhenWatcherConfigured(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}
	watcher := &fakeTableChangeWatcher{}
	store.SetTableChangeWatcher(watcher)

	_, err = store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "evt-table-change",
		SchemaVersion: 1,
		ProjectID:     "project-watch",
		RunID:         "run-watch",
		TaskID:        "task-watch",
		JobID:         "job-watch",
		SessionID:     "session-watch",
		WorkerID:      "worker-watch",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventStarted,
		OccurredAt:    time.Now().UTC(),
		Payload:       map[string]any{"state": "started"},
	})
	if err != nil {
		t.Fatalf("append event: %v", err)
	}
	if len(watcher.published) != 2 {
		t.Fatalf("expected two table-change publish calls, got %d", len(watcher.published))
	}
	if watcher.published[0].Topic != "lifecycle_project_session_history" {
		t.Fatalf("expected first topic lifecycle_project_session_history, got %q", watcher.published[0].Topic)
	}
	if watcher.published[1].Topic != "lifecycle_project_sessions" {
		t.Fatalf("expected second topic lifecycle_project_sessions, got %q", watcher.published[1].Topic)
	}
	if watcher.published[0].Payload == nil {
		t.Fatalf("expected history table-change event payload")
	}
	if watcher.published[0].Payload["event_type"] != string(domainlifecycle.EventStarted) {
		t.Fatalf("expected payload event_type=%q, got %+v", domainlifecycle.EventStarted, watcher.published[0].Payload["event_type"])
	}
	if watcher.published[1].Payload != nil {
		t.Fatalf("expected project_sessions upsert notification without payload")
	}
}

func TestEventStoreAppendDoesNotFailWhenTableChangePublishFails(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}
	store.SetTableChangeWatcher(&fakeFailingTableChangeWatcher{})

	_, err = store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "evt-table-change-failing-watcher",
		SchemaVersion: 1,
		ProjectID:     "project-watch-fail",
		RunID:         "run-watch-fail",
		TaskID:        "task-watch-fail",
		JobID:         "job-watch-fail",
		SessionID:     "session-watch-fail",
		WorkerID:      "worker-watch-fail",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventStarted,
		OccurredAt:    time.Now().UTC(),
		Payload:       map[string]any{"state": "started"},
	})
	if err != nil {
		t.Fatalf("append should not fail when watcher publish fails, got %v", err)
	}
}

func TestEventStoreAppendPersistsDefaultListenerDeliveries(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}

	appended, err := store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "evt-delivery-rows",
		SchemaVersion: 1,
		ProjectID:     "project-delivery",
		RunID:         "run-delivery",
		TaskID:        "task-delivery",
		JobID:         "job-delivery",
		SessionID:     "session-delivery",
		WorkerID:      "worker-delivery",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventStarted,
		OccurredAt:    time.Now().UTC(),
		Payload:       map[string]any{"state": "started"},
	})
	if err != nil {
		t.Fatalf("append event: %v", err)
	}

	deliveries, err := loadDeliveriesByEventID(store, appended.EventID)
	if err != nil {
		t.Fatalf("load delivery rows: %v", err)
	}
	if len(deliveries) != 2 {
		t.Fatalf("expected two default listener deliveries, got %d", len(deliveries))
	}
	for _, delivery := range deliveries {
		if delivery.DeliveryStatus != "pending" {
			t.Fatalf("expected pending delivery status, got %q", delivery.DeliveryStatus)
		}
	}
}

func TestEventStoreMarkDeliveryAttemptUpdatesRetryAndTerminalStatuses(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}

	appended, err := store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "evt-delivery-status",
		SchemaVersion: 1,
		ProjectID:     "project-delivery-status",
		RunID:         "run-delivery-status",
		TaskID:        "task-delivery-status",
		JobID:         "job-delivery-status",
		SessionID:     "session-delivery-status",
		WorkerID:      "worker-delivery-status",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventStarted,
		OccurredAt:    time.Now().UTC(),
		Payload:       map[string]any{"state": "started"},
	})
	if err != nil {
		t.Fatalf("append event: %v", err)
	}

	if err := store.MarkDeliveryAttempt(context.Background(), appended.EventID, "graphql_default", failures.WrapTransient(errors.New("temporary"))); err != nil {
		t.Fatalf("mark transient attempt: %v", err)
	}
	if err := store.MarkDeliveryAttempt(context.Background(), appended.EventID, "internal_default", failures.WrapTerminal(errors.New("fatal"))); err != nil {
		t.Fatalf("mark terminal attempt: %v", err)
	}

	deliveries, err := loadDeliveriesByEventID(store, appended.EventID)
	if err != nil {
		t.Fatalf("load deliveries: %v", err)
	}
	if len(deliveries) != 2 {
		t.Fatalf("expected two delivery rows, got %d", len(deliveries))
	}

	for _, delivery := range deliveries {
		switch delivery.ListenerID {
		case "graphql_default":
			if delivery.DeliveryStatus != "retrying" {
				t.Fatalf("expected graphql_default retrying, got %q", delivery.DeliveryStatus)
			}
			if delivery.NextAttemptAt == nil {
				t.Fatalf("expected retrying delivery to have next_attempt_at")
			}
		case "internal_default":
			if delivery.DeliveryStatus != "failed_terminal" {
				t.Fatalf("expected internal_default failed_terminal, got %q", delivery.DeliveryStatus)
			}
			if delivery.NextAttemptAt != nil {
				t.Fatalf("expected terminal failure to clear next_attempt_at")
			}
		}
	}
}

func TestEventStoreMarkDeliveryAttemptOpensCircuitAfterRepeatedTransientFailures(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}

	appended, err := store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "evt-circuit-open",
		SchemaVersion: 1,
		ProjectID:     "project-circuit-open",
		RunID:         "run-circuit-open",
		TaskID:        "task-circuit-open",
		JobID:         "job-circuit-open",
		SessionID:     "session-circuit-open",
		WorkerID:      "worker-circuit-open",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventStarted,
		OccurredAt:    time.Now().UTC(),
		Payload:       map[string]any{"state": "started"},
	})
	if err != nil {
		t.Fatalf("append event: %v", err)
	}

	for attempt := 0; attempt < listenerCircuitOpenThreshold; attempt++ {
		if err := store.MarkDeliveryAttempt(context.Background(), appended.EventID, "graphql_default", failures.WrapTransient(errors.New("temporary"))); err != nil {
			t.Fatalf("mark transient attempt %d: %v", attempt+1, err)
		}
	}

	deliveries, err := loadDeliveriesByEventID(store, appended.EventID)
	if err != nil {
		t.Fatalf("load deliveries: %v", err)
	}
	if len(deliveries) != 2 {
		t.Fatalf("expected two delivery rows, got %d", len(deliveries))
	}

	for _, delivery := range deliveries {
		if delivery.ListenerID != "graphql_default" {
			continue
		}
		if delivery.DeliveryStatus != "circuit_open" {
			t.Fatalf("expected circuit_open status, got %q", delivery.DeliveryStatus)
		}
		if delivery.NextAttemptAt != nil {
			t.Fatalf("expected circuit_open to clear next_attempt_at")
		}
		if !strings.Contains(delivery.LastError, "circuit breaker open") {
			t.Fatalf("expected last_error to mention circuit breaker, got %q", delivery.LastError)
		}
	}
}

func TestEventStoreRegisterListenersPersistsExternalDeliveryRows(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}
	store.RegisterListeners([]domainlifecycle.ListenerTarget{
		{ListenerID: "webhook_ops", ListenerType: domainlifecycle.ListenerTypeWebhook},
		{ListenerID: "slack_ops", ListenerType: domainlifecycle.ListenerTypeSlack},
		{ListenerID: "bus_analytics", ListenerType: domainlifecycle.ListenerTypeBus},
	})

	appended, err := store.Append(context.Background(), domainlifecycle.Event{
		EventID:       "evt-external-listeners",
		SchemaVersion: 1,
		ProjectID:     "project-ext-listeners",
		RunID:         "run-ext-listeners",
		TaskID:        "task-ext-listeners",
		JobID:         "job-ext-listeners",
		SessionID:     "session-ext-listeners",
		WorkerID:      "worker-ext-listeners",
		SourceRuntime: "worker",
		PipelineType:  "agent.workflow.run",
		EventType:     domainlifecycle.EventStarted,
		OccurredAt:    time.Now().UTC(),
		Payload:       map[string]any{"state": "started"},
	})
	if err != nil {
		t.Fatalf("append event: %v", err)
	}

	deliveries, err := loadDeliveriesByEventID(store, appended.EventID)
	if err != nil {
		t.Fatalf("load deliveries: %v", err)
	}
	if len(deliveries) != 5 {
		t.Fatalf("expected 5 listener delivery rows (2 internal + 3 external), got %d", len(deliveries))
	}

	found := map[string]bool{}
	for _, delivery := range deliveries {
		found[delivery.ListenerID] = true
	}
	if !found["webhook_ops"] || !found["slack_ops"] || !found["bus_analytics"] {
		t.Fatalf("expected external listener rows to be persisted, got %+v", found)
	}
}

func TestEventStoreRegisterListenersSkipsInvalidAndDuplicateTargets(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}
	store.RegisterListeners([]domainlifecycle.ListenerTarget{
		{ListenerID: "graphql_default", ListenerType: domainlifecycle.ListenerTypeGraphQL},
		{ListenerID: "", ListenerType: domainlifecycle.ListenerTypeWebhook},
		{ListenerID: "webhook_ops", ListenerType: domainlifecycle.ListenerTypeWebhook},
	})
	if len(store.listeners) != 3 {
		t.Fatalf("expected one additional valid listener, got %d", len(store.listeners))
	}
}

func TestEventStoreChaosWorkerCrashAndRestartMaintainsDeterministicRecovery(t *testing.T) {
	store, err := NewEventStore(newLifecycleTestDB(t))
	if err != nil {
		t.Fatalf("new event store: %v", err)
	}
	now := time.Now().UTC()

	events := []domainlifecycle.Event{
		{
			EventID:       "evt-chaos-started-1",
			SchemaVersion: 1,
			ProjectID:     "project-chaos",
			RunID:         "run-chaos",
			TaskID:        "task-chaos",
			JobID:         "job-chaos",
			SessionID:     "session-chaos",
			WorkerID:      "worker-chaos",
			SourceRuntime: "worker",
			PipelineType:  "agent.workflow.run",
			EventType:     domainlifecycle.EventStarted,
			OccurredAt:    now,
			Payload:       map[string]any{"runtime_alive": true},
		},
		{
			EventID:       "evt-chaos-failed-1",
			SchemaVersion: 1,
			ProjectID:     "project-chaos",
			RunID:         "run-chaos",
			TaskID:        "task-chaos",
			JobID:         "job-chaos",
			SessionID:     "session-chaos",
			WorkerID:      "worker-chaos",
			SourceRuntime: "worker",
			PipelineType:  "agent.workflow.run",
			EventType:     domainlifecycle.EventFailed,
			OccurredAt:    now.Add(2 * time.Second),
			Payload:       map[string]any{"runtime_alive": false, "error": "panic"},
		},
		{
			EventID:       "evt-chaos-started-2",
			SchemaVersion: 1,
			ProjectID:     "project-chaos",
			RunID:         "run-chaos",
			TaskID:        "task-chaos",
			JobID:         "job-chaos",
			SessionID:     "session-chaos",
			WorkerID:      "worker-chaos-2",
			SourceRuntime: "worker",
			PipelineType:  "agent.workflow.run",
			EventType:     domainlifecycle.EventStarted,
			OccurredAt:    now.Add(5 * time.Second),
			Payload:       map[string]any{"runtime_alive": true},
		},
	}

	for index, event := range events {
		appended, appendErr := store.Append(context.Background(), event)
		if appendErr != nil {
			t.Fatalf("append event %d: %v", index+1, appendErr)
		}
		if appended.EventSeq != int64(index+1) {
			t.Fatalf("expected monotonic event_seq %d, got %d", index+1, appended.EventSeq)
		}
	}

	history, err := loadHistoryBySessionID(store, "session-chaos")
	if err != nil {
		t.Fatalf("load history: %v", err)
	}
	if len(history) != 3 {
		t.Fatalf("expected 3 persisted events, got %d", len(history))
	}
	if history[2].EventType != string(domainlifecycle.EventStarted) {
		t.Fatalf("expected replay tail to be restart started event, got %s", history[2].EventType)
	}

	snapshot, err := loadSnapshotBySessionID(store, "session-chaos")
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	if snapshot == nil {
		t.Fatalf("expected snapshot to exist")
	}
	if snapshot.CurrentState != "healthy_active" {
		t.Fatalf("expected recovered snapshot state healthy_active, got %q", snapshot.CurrentState)
	}
}

func loadSnapshotBySessionID(store *EventStore, sessionID string) (*projectSessionRecord, error) {
	if store == nil || store.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	var snapshot projectSessionRecord
	err := store.db.WithContext(context.Background()).Where("session_id = ?", sessionID).Take(&snapshot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func loadHistoryBySessionID(store *EventStore, sessionID string) ([]eventHistoryRecord, error) {
	if store == nil || store.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	var records []eventHistoryRecord
	err := store.db.WithContext(context.Background()).Where("session_id = ?", sessionID).Order("event_seq ASC").Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func loadDeliveriesByEventID(store *EventStore, eventID string) ([]feedbackDeliveryRecord, error) {
	if store == nil || store.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	var records []feedbackDeliveryRecord
	err := store.db.WithContext(context.Background()).Where("event_id = ?", eventID).Order("listener_id ASC").Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

type fakeTableChangeWatcher struct {
	published []domainrealtime.TableChangeEvent
}

func (watcher *fakeTableChangeWatcher) Publish(_ context.Context, event domainrealtime.TableChangeEvent) error {
	watcher.published = append(watcher.published, event)
	return nil
}

func (watcher *fakeTableChangeWatcher) Watch(_ context.Context, _ string, _ func(domainrealtime.TableChangeEvent) error) error {
	return nil
}

type fakeFailingTableChangeWatcher struct{}

func (watcher *fakeFailingTableChangeWatcher) Publish(_ context.Context, _ domainrealtime.TableChangeEvent) error {
	return errors.New("publish failed")
}

func (watcher *fakeFailingTableChangeWatcher) Watch(_ context.Context, _ string, _ func(domainrealtime.TableChangeEvent) error) error {
	return nil
}
