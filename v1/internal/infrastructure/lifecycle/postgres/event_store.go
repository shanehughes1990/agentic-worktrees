package postgres

import (
	applicationlifecycle "agentic-orchestrator/internal/application/lifecycle"
	"agentic-orchestrator/internal/domain/failures"
	domainlifecycle "agentic-orchestrator/internal/domain/lifecycle"
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
	domainsecurity "agentic-orchestrator/internal/domain/security"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const listenerCircuitOpenThreshold = 3

type eventHistoryRecord struct {
	gorm.Model
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	EventID         string    `gorm:"column:event_id;size:255;not null;uniqueIndex"`
	SchemaVersion   int       `gorm:"column:schema_version;not null"`
	ProjectID       string    `gorm:"column:project_id;size:255;not null;index:idx_history_project_seq,priority:1;index:idx_history_project_occurred,priority:1"`
	RunID           string    `gorm:"column:run_id;size:255;index"`
	TaskID          string    `gorm:"column:task_id;size:255;index"`
	JobID           string    `gorm:"column:job_id;size:255;index"`
	SessionID       string    `gorm:"column:session_id;size:255;not null;index:idx_history_session_seq,priority:1"`
	WorkerID        string    `gorm:"column:worker_id;size:255"`
	SourceRuntime   string    `gorm:"column:source_runtime;size:128;not null"`
	PipelineType    string    `gorm:"column:pipeline_type;size:128;not null;index"`
	ProjectEventSeq int64     `gorm:"column:project_event_seq;not null;index:idx_history_project_seq,priority:2"`
	EventSeq        int64     `gorm:"column:event_seq;not null;index:idx_history_session_seq,priority:2"`
	OccurredAt      time.Time `gorm:"column:occurred_at;not null;index:idx_history_project_occurred,priority:2"`
	IngestedAt      time.Time `gorm:"column:ingested_at;not null"`
	EventType       string    `gorm:"column:event_type;size:128;not null;index"`
	PayloadJSON     string    `gorm:"column:payload_json;type:text;not null"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;autoCreateTime"`
}

func (eventHistoryRecord) TableName() string {
	return "project_session_history"
}

type projectSessionRecord struct {
	gorm.Model
	ID                  uint64     `gorm:"primaryKey;autoIncrement"`
	ProjectID           string     `gorm:"column:project_id;size:255;not null;index:idx_project_sessions_project_updated,priority:1;index:idx_project_sessions_project_state,priority:1"`
	RunID               string     `gorm:"column:run_id;size:255;index"`
	PipelineType        string     `gorm:"column:pipeline_type;size:128;not null;index:idx_project_sessions_project_pipeline,priority:2"`
	TaskID              string     `gorm:"column:task_id;size:255;index"`
	JobID               string     `gorm:"column:job_id;size:255;index"`
	SessionID           string     `gorm:"column:session_id;size:255;not null;uniqueIndex"`
	WorkerID            string     `gorm:"column:worker_id;size:255"`
	SourceRuntime       string     `gorm:"column:source_runtime;size:128;not null"`
	CurrentState        string     `gorm:"column:current_state;size:64;not null;index:idx_project_sessions_project_state,priority:2"`
	CurrentSeverity     string     `gorm:"column:current_severity;size:32;not null"`
	LastReasonCode      string     `gorm:"column:last_reason_code;size:128"`
	LastReasonSummary   string     `gorm:"column:last_reason_summary;size:1024"`
	LastLivenessAt      *time.Time `gorm:"column:last_liveness_at"`
	LastActivityAt      *time.Time `gorm:"column:last_activity_at"`
	LastCheckpointAt    *time.Time `gorm:"column:last_checkpoint_at"`
	LastEventSeq        int64      `gorm:"column:last_event_seq;not null"`
	LastProjectEventSeq int64      `gorm:"column:last_project_event_seq;not null"`
	StartedAt           time.Time  `gorm:"column:started_at;not null"`
	EndedAt             *time.Time `gorm:"column:ended_at"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;not null;autoUpdateTime;index:idx_project_sessions_project_updated,priority:2"`
}

func (projectSessionRecord) TableName() string {
	return "project_sessions"
}

type feedbackDeliveryRecord struct {
	gorm.Model
	ID             uint64     `gorm:"primaryKey;autoIncrement"`
	EventID        string     `gorm:"column:event_id;size:255;not null;uniqueIndex:ux_feedback_delivery_event_listener,priority:1;index"`
	ListenerID     string     `gorm:"column:listener_id;size:255;not null;uniqueIndex:ux_feedback_delivery_event_listener,priority:2;index:idx_feedback_delivery_listener_status_next,priority:1"`
	ListenerType   string     `gorm:"column:listener_type;size:64;not null;index:idx_feedback_delivery_listener_type_updated,priority:1"`
	DeliveryStatus string     `gorm:"column:delivery_status;size:64;not null;index:idx_feedback_delivery_listener_status_next,priority:2"`
	AttemptCount   int        `gorm:"column:attempt_count;not null"`
	LastError      string     `gorm:"column:last_error;size:2048"`
	NextAttemptAt  *time.Time `gorm:"column:next_attempt_at;index:idx_feedback_delivery_listener_status_next,priority:3"`
	LastAttemptAt  *time.Time `gorm:"column:last_attempt_at"`
	DeliveredAt    *time.Time `gorm:"column:delivered_at"`
	CursorEventSeq int64      `gorm:"column:cursor_event_seq"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null;autoUpdateTime;index:idx_feedback_delivery_listener_type_updated,priority:2"`
}

func (feedbackDeliveryRecord) TableName() string {
	return "project_session_feedback_deliveries"
}

type EventStore struct {
	db        *gorm.DB
	policy    domainlifecycle.ClassificationPolicy
	watcher   domainrealtime.TableChangeWatcher
	listeners []domainlifecycle.ListenerTarget
}

func NewEventStore(db *gorm.DB) (*EventStore, error) {
	if db == nil {
		return nil, fmt.Errorf("lifecycle event store: db is required")
	}
	if err := db.AutoMigrate(&eventHistoryRecord{}, &projectSessionRecord{}, &feedbackDeliveryRecord{}); err != nil {
		return nil, fmt.Errorf("lifecycle event store: migrate: %w", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ux_project_session_history_session_seq ON project_session_history (session_id, event_seq)").Error; err != nil {
		return nil, fmt.Errorf("lifecycle event store: create session sequence index: %w", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ux_project_session_history_project_seq ON project_session_history (project_id, project_event_seq)").Error; err != nil {
		return nil, fmt.Errorf("lifecycle event store: create project sequence index: %w", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_project_sessions_project_pipeline ON project_sessions (project_id, pipeline_type)").Error; err != nil {
		return nil, fmt.Errorf("lifecycle event store: create project sessions pipeline index: %w", err)
	}
	return &EventStore{
		db:     db,
		policy: domainlifecycle.DefaultClassificationPolicy(),
		listeners: []domainlifecycle.ListenerTarget{
			{ListenerID: "graphql_default", ListenerType: domainlifecycle.ListenerTypeGraphQL},
			{ListenerID: "internal_default", ListenerType: domainlifecycle.ListenerTypeInternal},
		},
	}, nil
}

func (store *EventStore) RegisterListeners(listeners []domainlifecycle.ListenerTarget) {
	if store == nil {
		return
	}
	if len(listeners) == 0 {
		return
	}
	existing := map[string]struct{}{}
	for _, listener := range store.listeners {
		existing[strings.TrimSpace(listener.ListenerID)] = struct{}{}
	}
	for _, listener := range listeners {
		if err := listener.Validate(); err != nil {
			continue
		}
		listenerID := strings.TrimSpace(listener.ListenerID)
		if listenerID == "" {
			continue
		}
		if _, found := existing[listenerID]; found {
			continue
		}
		store.listeners = append(store.listeners, listener)
		existing[listenerID] = struct{}{}
	}
}

func (store *EventStore) Append(ctx context.Context, event domainlifecycle.Event) (domainlifecycle.Event, error) {
	if store == nil || store.db == nil {
		return domainlifecycle.Event{}, fmt.Errorf("lifecycle event store: db is not initialized")
	}
	if err := event.ValidateForAppend(); err != nil {
		return domainlifecycle.Event{}, err
	}
	existing, err := store.loadByEventID(ctx, strings.TrimSpace(event.EventID))
	if err != nil {
		return domainlifecycle.Event{}, err
	}
	if existing != nil {
		return mapRecordToEvent(*existing), nil
	}

	for attempt := 0; attempt < 3; attempt++ {
		persisted, appendErr := store.appendOnce(ctx, event)
		if appendErr == nil {
			store.publishTableChangeSignals(ctx, persisted)
			return persisted, nil
		}
		if isDuplicateEventIDError(appendErr) {
			record, getErr := store.loadByEventID(ctx, strings.TrimSpace(event.EventID))
			if getErr != nil {
				return domainlifecycle.Event{}, getErr
			}
			if record != nil {
				return mapRecordToEvent(*record), nil
			}
		}
		if !isSequenceConflictError(appendErr) {
			return domainlifecycle.Event{}, appendErr
		}
	}
	return domainlifecycle.Event{}, fmt.Errorf("lifecycle event store: append retries exhausted")
}

func (store *EventStore) SetTableChangeWatcher(watcher domainrealtime.TableChangeWatcher) {
	if store == nil {
		return
	}
	store.watcher = watcher
}

func (store *EventStore) MarkDeliveryAttempt(ctx context.Context, eventID string, listenerID string, deliveryErr error) error {
	if store == nil || store.db == nil {
		return fmt.Errorf("lifecycle event store: db is not initialized")
	}
	normalizedEventID := strings.TrimSpace(eventID)
	normalizedListenerID := strings.TrimSpace(listenerID)
	if normalizedEventID == "" || normalizedListenerID == "" {
		return fmt.Errorf("lifecycle event store: event_id and listener_id are required")
	}
	return store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row feedbackDeliveryRecord
		if err := tx.Where("event_id = ? AND listener_id = ?", normalizedEventID, normalizedListenerID).Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return fmt.Errorf("lifecycle event store: load delivery row: %w", err)
		}
		now := time.Now().UTC()
		if deliveryErr == nil {
			row.DeliveryStatus = "delivered"
			row.LastError = ""
			row.DeliveredAt = &now
			row.LastAttemptAt = &now
			row.NextAttemptAt = nil
		} else {
			row.AttemptCount++
			row.LastAttemptAt = &now
			row.LastError = strings.TrimSpace(deliveryErr.Error())
			if failures.ClassOf(deliveryErr) == failures.ClassTransient {
				if row.AttemptCount >= listenerCircuitOpenThreshold {
					row.DeliveryStatus = "circuit_open"
					row.NextAttemptAt = nil
					row.LastError = fmt.Sprintf("circuit breaker open: %s", row.LastError)
				} else {
					row.DeliveryStatus = "retrying"
					next := now.Add(deliveryBackoffDuration(row.AttemptCount))
					row.NextAttemptAt = &next
				}
			} else {
				row.DeliveryStatus = "failed_terminal"
				row.NextAttemptAt = nil
			}
		}
		if err := tx.Save(&row).Error; err != nil {
			return fmt.Errorf("lifecycle event store: update delivery row: %w", err)
		}
		return nil
	})
}

func (store *EventStore) publishTableChangeSignals(ctx context.Context, event domainlifecycle.Event) {
	if store == nil || store.watcher == nil {
		return
	}
	livePayload := map[string]any{
		"event_id":          strings.TrimSpace(event.EventID),
		"event_type":        strings.TrimSpace(string(event.EventType)),
		"source_runtime":    strings.TrimSpace(event.SourceRuntime),
		"pipeline_type":     strings.TrimSpace(event.PipelineType),
		"project_id":        strings.TrimSpace(event.ProjectID),
		"run_id":            strings.TrimSpace(event.RunID),
		"task_id":           strings.TrimSpace(event.TaskID),
		"job_id":            strings.TrimSpace(event.JobID),
		"session_id":        strings.TrimSpace(event.SessionID),
		"correlation_id":    "session:" + strings.TrimSpace(event.SessionID),
		"session_event_seq": event.EventSeq,
		"project_event_seq": event.ProjectEventSeq,
		"occurred_at":       event.OccurredAt.UTC().Format(time.RFC3339Nano),
		"payload":           clonePayload(event.Payload),
	}
	base := domainrealtime.TableChangeEvent{
		ProjectID:       strings.TrimSpace(event.ProjectID),
		RunID:           strings.TrimSpace(event.RunID),
		TaskID:          strings.TrimSpace(event.TaskID),
		JobID:           strings.TrimSpace(event.JobID),
		SessionID:       strings.TrimSpace(event.SessionID),
		ProjectEventSeq: event.ProjectEventSeq,
		SessionEventSeq: event.EventSeq,
		OccurredAt:      event.OccurredAt.UTC(),
		Payload:         livePayload,
	}
	_ = store.watcher.Publish(ctx, mergeTableEvent(base, "lifecycle_project_session_history", "project_session_history", "insert"))
	base.Payload = nil
	_ = store.watcher.Publish(ctx, mergeTableEvent(base, "lifecycle_project_sessions", "project_sessions", "upsert"))
}

func mergeTableEvent(base domainrealtime.TableChangeEvent, topic string, table string, operation string) domainrealtime.TableChangeEvent {
	base.Topic = topic
	base.Table = table
	base.Operation = operation
	return base
}

func (store *EventStore) appendOnce(ctx context.Context, event domainlifecycle.Event) (domainlifecycle.Event, error) {
	payload := clonePayload(event.Payload)
	payload, _ = domainsecurity.RedactPayload(payload)
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return domainlifecycle.Event{}, fmt.Errorf("lifecycle event store: marshal payload: %w", err)
	}
	var persisted domainlifecycle.Event
	txErr := store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var maxSessionSeq int64
		if err := tx.Model(&eventHistoryRecord{}).
			Where("session_id = ?", strings.TrimSpace(event.SessionID)).
			Select("COALESCE(MAX(event_seq), 0)").
			Scan(&maxSessionSeq).Error; err != nil {
			return fmt.Errorf("lifecycle event store: select session max sequence: %w", err)
		}
		var maxProjectSeq int64
		if err := tx.Model(&eventHistoryRecord{}).
			Where("project_id = ?", strings.TrimSpace(event.ProjectID)).
			Select("COALESCE(MAX(project_event_seq), 0)").
			Scan(&maxProjectSeq).Error; err != nil {
			return fmt.Errorf("lifecycle event store: select project max sequence: %w", err)
		}

		now := time.Now().UTC()
		nextSessionSeq := maxSessionSeq + 1
		nextProjectSeq := maxProjectSeq + 1
		record := eventHistoryRecord{
			EventID:         strings.TrimSpace(event.EventID),
			SchemaVersion:   event.SchemaVersion,
			ProjectID:       strings.TrimSpace(event.ProjectID),
			RunID:           strings.TrimSpace(event.RunID),
			TaskID:          strings.TrimSpace(event.TaskID),
			JobID:           strings.TrimSpace(event.JobID),
			SessionID:       strings.TrimSpace(event.SessionID),
			WorkerID:        strings.TrimSpace(event.WorkerID),
			SourceRuntime:   strings.TrimSpace(event.SourceRuntime),
			PipelineType:    strings.TrimSpace(event.PipelineType),
			ProjectEventSeq: nextProjectSeq,
			EventSeq:        nextSessionSeq,
			OccurredAt:      event.OccurredAt.UTC(),
			IngestedAt:      now,
			EventType:       strings.TrimSpace(string(event.EventType)),
			PayloadJSON:     string(payloadJSON),
			CreatedAt:       now,
		}

		expectedSessionSeq := expectedSessionSequence(event.EventSeq, payload)
		gapDetected := expectedSessionSeq > 0 && expectedSessionSeq > record.EventSeq
		gapReconciled := expectedSessionSeq > 0 && expectedSessionSeq < record.EventSeq
		if gapDetected || gapReconciled {
			payload["expected_event_seq"] = expectedSessionSeq
			payload["observed_event_seq"] = record.EventSeq
			payload["gap_detected"] = gapDetected
			payload["gap_reconciled"] = gapReconciled
			payloadJSON, err = json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("lifecycle event store: marshal payload with gap metadata: %w", err)
			}
			record.PayloadJSON = string(payloadJSON)
		}

		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("lifecycle event store: append: %w", err)
		}
		store.persistDeliveryRowsBestEffort(tx, record)

		if gapDetected {
			nextSessionSeq++
			nextProjectSeq++
			gapEvent, buildErr := buildSyntheticGapHistoryRecord(record, domainlifecycle.EventGapDetected, expectedSessionSeq, nextSessionSeq, nextProjectSeq, now)
			if buildErr != nil {
				return buildErr
			}
			if err := tx.Create(&gapEvent).Error; err != nil {
				return fmt.Errorf("lifecycle event store: append synthetic gap_detected: %w", err)
			}
			store.persistDeliveryRowsBestEffort(tx, gapEvent)
			if err := upsertSessionSnapshot(tx, gapEvent, syntheticGapPayload(gapEvent), store.policy); err != nil {
				return err
			}
		}
		if gapReconciled {
			nextSessionSeq++
			nextProjectSeq++
			gapEvent, buildErr := buildSyntheticGapHistoryRecord(record, domainlifecycle.EventGapReconciled, expectedSessionSeq, nextSessionSeq, nextProjectSeq, now)
			if buildErr != nil {
				return buildErr
			}
			if err := tx.Create(&gapEvent).Error; err != nil {
				return fmt.Errorf("lifecycle event store: append synthetic gap_reconciled: %w", err)
			}
			store.persistDeliveryRowsBestEffort(tx, gapEvent)
			if err := upsertSessionSnapshot(tx, gapEvent, syntheticGapPayload(gapEvent), store.policy); err != nil {
				return err
			}
		}

		if err := upsertSessionSnapshot(tx, record, payload, store.policy); err != nil {
			return err
		}
		persisted = mapRecordToEvent(record)
		persisted.Payload = payload
		return nil
	})
	if txErr != nil {
		return domainlifecycle.Event{}, txErr
	}
	return persisted, nil
}

func (store *EventStore) persistDeliveryRowsBestEffort(tx *gorm.DB, history eventHistoryRecord) {
	if store == nil || tx == nil || len(store.listeners) == 0 {
		return
	}
	now := time.Now().UTC()
	rows := make([]feedbackDeliveryRecord, 0, len(store.listeners))
	for _, listener := range store.listeners {
		if err := listener.Validate(); err != nil {
			continue
		}
		rows = append(rows, feedbackDeliveryRecord{
			EventID:        strings.TrimSpace(history.EventID),
			ListenerID:     strings.TrimSpace(listener.ListenerID),
			ListenerType:   strings.TrimSpace(string(listener.ListenerType)),
			DeliveryStatus: "pending",
			AttemptCount:   0,
			NextAttemptAt:  &now,
			CursorEventSeq: history.EventSeq,
		})
	}
	if len(rows) == 0 {
		return
	}
	_ = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

func (store *EventStore) loadByEventID(ctx context.Context, eventID string) (*eventHistoryRecord, error) {
	if strings.TrimSpace(eventID) == "" {
		return nil, nil
	}
	var record eventHistoryRecord
	err := store.db.WithContext(ctx).Where("event_id = ?", strings.TrimSpace(eventID)).Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("lifecycle event store: load by event id: %w", err)
	}
	return &record, nil
}

func mapRecordToEvent(record eventHistoryRecord) domainlifecycle.Event {
	payload := map[string]any{}
	if strings.TrimSpace(record.PayloadJSON) != "" {
		_ = json.Unmarshal([]byte(record.PayloadJSON), &payload)
	}
	return domainlifecycle.Event{
		EventID:         record.EventID,
		SchemaVersion:   record.SchemaVersion,
		ProjectID:       record.ProjectID,
		RunID:           record.RunID,
		TaskID:          record.TaskID,
		JobID:           record.JobID,
		SessionID:       record.SessionID,
		WorkerID:        record.WorkerID,
		SourceRuntime:   record.SourceRuntime,
		PipelineType:    record.PipelineType,
		EventType:       domainlifecycle.EventType(record.EventType),
		OccurredAt:      record.OccurredAt,
		Payload:         payload,
		EventSeq:        record.EventSeq,
		ProjectEventSeq: record.ProjectEventSeq,
	}
}

func isDuplicateEventIDError(err error) bool {
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "event_id") && (strings.Contains(message, "duplicate") || strings.Contains(message, "unique constraint"))
}

func isSequenceConflictError(err error) bool {
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	if strings.Contains(message, "session_id") && strings.Contains(message, "event_seq") && (strings.Contains(message, "duplicate") || strings.Contains(message, "unique constraint")) {
		return true
	}
	if strings.Contains(message, "project_id") && strings.Contains(message, "project_event_seq") && (strings.Contains(message, "duplicate") || strings.Contains(message, "unique constraint")) {
		return true
	}
	return false
}

func upsertSessionSnapshot(tx *gorm.DB, history eventHistoryRecord, payload map[string]any, policy domainlifecycle.ClassificationPolicy) error {
	var snapshot projectSessionRecord
	err := tx.Where("session_id = ?", strings.TrimSpace(history.SessionID)).Take(&snapshot).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("lifecycle event store: load project session snapshot: %w", err)
	}

	now := history.OccurredAt.UTC()
	previousChangedAt := snapshot.UpdatedAt
	state := strings.TrimSpace(snapshot.CurrentState)
	severity := strings.TrimSpace(snapshot.CurrentSeverity)
	reasonCode := ""
	if !isGapEventType(history.EventType) || state == "" {
		classification := domainlifecycle.Classify(domainlifecycle.ClassificationInput{
			Now:               now,
			EventType:         domainlifecycle.EventType(strings.TrimSpace(history.EventType)),
			RuntimeAlive:      runtimeAliveForEvent(history.EventType, payload),
			WaitingInput:      payloadBool(payload, "waiting_input"),
			LastActivityAt:    maxTimePtr(snapshot.LastActivityAt, &history.OccurredAt),
			LastCheckpointAt:  maxTimePtr(snapshot.LastCheckpointAt, payloadTime(payload, "last_checkpoint_at")),
			PreviousState:     domainlifecycle.State(strings.TrimSpace(snapshot.CurrentState)),
			PreviousChangedAt: timePtr(previousChangedAt),
		}, policy)
		state = string(classification.State)
		severity = classification.Severity
		reasonCode = classification.ReasonCode
	}
	if state == "" {
		state = string(domainlifecycle.StateHealthyActive)
	}
	if severity == "" {
		severity = "info"
	}
	lastReasonSummary := strings.TrimSpace(readString(payload, "error"))
	if lastReasonSummary == "" {
		lastReasonSummary = strings.TrimSpace(readString(payload, "reason_summary"))
	}
	lastReasonCode := strings.TrimSpace(readString(payload, "reason_code"))
	if lastReasonCode == "" && history.EventType == string(domainlifecycle.EventFailed) {
		lastReasonCode = "handler_error_" + strings.TrimSpace(readString(payload, "failure_class"))
	}
	if lastReasonCode == "" {
		lastReasonCode = reasonCode
	}
	if lastReasonCode == "" && isGapEventType(history.EventType) {
		lastReasonCode = strings.TrimSpace(history.EventType)
	}

	activityAt := history.OccurredAt
	livenessAt := payloadTime(payload, "last_liveness_at")
	if livenessAt == nil {
		livenessAt = &activityAt
	}
	checkpointAt := payloadTime(payload, "last_checkpoint_at")
	if checkpointAt == nil && history.EventType == string(domainlifecycle.EventCompleted) {
		checkpointAt = &activityAt
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newSnapshot := projectSessionRecord{
			ProjectID:           strings.TrimSpace(history.ProjectID),
			RunID:               strings.TrimSpace(history.RunID),
			PipelineType:        strings.TrimSpace(history.PipelineType),
			TaskID:              strings.TrimSpace(history.TaskID),
			JobID:               strings.TrimSpace(history.JobID),
			SessionID:           strings.TrimSpace(history.SessionID),
			WorkerID:            strings.TrimSpace(history.WorkerID),
			SourceRuntime:       strings.TrimSpace(history.SourceRuntime),
			CurrentState:        state,
			CurrentSeverity:     severity,
			LastReasonCode:      lastReasonCode,
			LastReasonSummary:   lastReasonSummary,
			LastLivenessAt:      livenessAt,
			LastActivityAt:      &activityAt,
			LastCheckpointAt:    checkpointAt,
			LastEventSeq:        history.EventSeq,
			LastProjectEventSeq: history.ProjectEventSeq,
			StartedAt:           history.OccurredAt,
		}
		if isTerminalEvent(history.EventType) {
			endedAt := history.OccurredAt
			newSnapshot.EndedAt = &endedAt
		}
		if err := tx.Create(&newSnapshot).Error; err != nil {
			return fmt.Errorf("lifecycle event store: create project session snapshot: %w", err)
		}
		return nil
	}

	if history.EventSeq > snapshot.LastEventSeq {
		snapshot.LastEventSeq = history.EventSeq
		snapshot.CurrentState = state
		snapshot.CurrentSeverity = severity
		snapshot.LastReasonCode = lastReasonCode
		snapshot.LastReasonSummary = lastReasonSummary
		snapshot.RunID = strings.TrimSpace(history.RunID)
		snapshot.TaskID = strings.TrimSpace(history.TaskID)
		snapshot.JobID = strings.TrimSpace(history.JobID)
		snapshot.WorkerID = strings.TrimSpace(history.WorkerID)
		snapshot.PipelineType = strings.TrimSpace(history.PipelineType)
		snapshot.SourceRuntime = strings.TrimSpace(history.SourceRuntime)
	}
	if history.ProjectEventSeq > snapshot.LastProjectEventSeq {
		snapshot.LastProjectEventSeq = history.ProjectEventSeq
	}

	snapshot.LastActivityAt = maxTimePtr(snapshot.LastActivityAt, &activityAt)
	snapshot.LastLivenessAt = maxTimePtr(snapshot.LastLivenessAt, livenessAt)
	snapshot.LastCheckpointAt = maxTimePtr(snapshot.LastCheckpointAt, checkpointAt)
	snapshot.StartedAt = minTime(snapshot.StartedAt, history.OccurredAt)
	if isTerminalEvent(history.EventType) {
		endedAt := history.OccurredAt
		snapshot.EndedAt = maxTimePtr(snapshot.EndedAt, &endedAt)
	}

	if err := tx.Save(&snapshot).Error; err != nil {
		return fmt.Errorf("lifecycle event store: update project session snapshot: %w", err)
	}
	return nil
}

func isTerminalEvent(eventType string) bool {
	trimmed := strings.TrimSpace(eventType)
	return trimmed == string(domainlifecycle.EventCompleted) || trimmed == string(domainlifecycle.EventFailed)
}

func isGapEventType(eventType string) bool {
	trimmed := strings.TrimSpace(eventType)
	return trimmed == string(domainlifecycle.EventGapDetected) || trimmed == string(domainlifecycle.EventGapReconciled)
}

func expectedSessionSequence(eventSeqHint int64, payload map[string]any) int64 {
	if eventSeqHint > 0 {
		return eventSeqHint
	}
	return payloadInt64(payload, "expected_event_seq")
}

func payloadInt64(payload map[string]any, key string) int64 {
	if payload == nil {
		return 0
	}
	rawValue, exists := payload[key]
	if !exists || rawValue == nil {
		return 0
	}
	switch value := rawValue.(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	}
	return 0
}

func buildSyntheticGapHistoryRecord(base eventHistoryRecord, eventType domainlifecycle.EventType, expectedSeq int64, eventSeq int64, projectSeq int64, ingestedAt time.Time) (eventHistoryRecord, error) {
	payload, err := json.Marshal(map[string]any{
		"related_event_id":   strings.TrimSpace(base.EventID),
		"expected_event_seq": expectedSeq,
		"observed_event_seq": base.EventSeq,
	})
	if err != nil {
		return eventHistoryRecord{}, fmt.Errorf("lifecycle event store: marshal synthetic gap payload: %w", err)
	}
	return eventHistoryRecord{
		EventID:         fmt.Sprintf("lifecycle:gap:%s:%s:%d:%d", strings.TrimSpace(base.SessionID), strings.TrimSpace(string(eventType)), expectedSeq, base.EventSeq),
		SchemaVersion:   base.SchemaVersion,
		ProjectID:       strings.TrimSpace(base.ProjectID),
		RunID:           strings.TrimSpace(base.RunID),
		TaskID:          strings.TrimSpace(base.TaskID),
		JobID:           strings.TrimSpace(base.JobID),
		SessionID:       strings.TrimSpace(base.SessionID),
		WorkerID:        strings.TrimSpace(base.WorkerID),
		SourceRuntime:   strings.TrimSpace(base.SourceRuntime),
		PipelineType:    strings.TrimSpace(base.PipelineType),
		ProjectEventSeq: projectSeq,
		EventSeq:        eventSeq,
		OccurredAt:      base.OccurredAt,
		IngestedAt:      ingestedAt.UTC(),
		EventType:       string(eventType),
		PayloadJSON:     string(payload),
		CreatedAt:       ingestedAt.UTC(),
	}, nil
}

func syntheticGapPayload(record eventHistoryRecord) map[string]any {
	result := map[string]any{}
	if strings.TrimSpace(record.PayloadJSON) == "" {
		return result
	}
	_ = json.Unmarshal([]byte(record.PayloadJSON), &result)
	return result
}

func clonePayload(payload map[string]any) map[string]any {
	if payload == nil {
		return map[string]any{}
	}
	copy := make(map[string]any, len(payload))
	for key, value := range payload {
		copy[key] = value
	}
	return copy
}

func payloadBool(payload map[string]any, key string) bool {
	if payload == nil {
		return false
	}
	rawValue, exists := payload[key]
	if !exists || rawValue == nil {
		return false
	}
	value, ok := rawValue.(bool)
	if ok {
		return value
	}
	stringValue, ok := rawValue.(string)
	if !ok {
		return false
	}
	normalized := strings.TrimSpace(strings.ToLower(stringValue))
	return normalized == "true" || normalized == "1" || normalized == "yes"
}

func runtimeAliveForEvent(eventType string, payload map[string]any) bool {
	if payload != nil {
		if _, exists := payload["runtime_alive"]; exists {
			return payloadBool(payload, "runtime_alive")
		}
	}
	if strings.TrimSpace(eventType) == string(domainlifecycle.EventFailed) {
		return false
	}
	return true
}

func timePtr(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	copy := value.UTC()
	return &copy
}

func readString(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	rawValue, exists := payload[key]
	if !exists || rawValue == nil {
		return ""
	}
	value, ok := rawValue.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func payloadTime(payload map[string]any, key string) *time.Time {
	value := strings.TrimSpace(readString(payload, key))
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	parsed = parsed.UTC()
	return &parsed
}

func maxTimePtr(existing *time.Time, candidate *time.Time) *time.Time {
	if candidate == nil {
		return existing
	}
	if existing == nil || candidate.After(*existing) {
		clone := *candidate
		return &clone
	}
	return existing
}

func minTime(existing time.Time, candidate time.Time) time.Time {
	if existing.IsZero() {
		return candidate
	}
	if candidate.IsZero() {
		return existing
	}
	if candidate.Before(existing) {
		return candidate
	}
	return existing
}

func deliveryBackoffDuration(attemptCount int) time.Duration {
	if attemptCount <= 0 {
		return 30 * time.Second
	}
	if attemptCount > 5 {
		attemptCount = 5
	}
	return time.Duration(1<<(attemptCount-1)) * 30 * time.Second
}

var _ applicationlifecycle.EventStore = (*EventStore)(nil)
