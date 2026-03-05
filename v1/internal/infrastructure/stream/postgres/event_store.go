package postgres

import (
	applicationstream "agentic-orchestrator/internal/application/stream"
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
	domainsecurity "agentic-orchestrator/internal/domain/security"
	domainstream "agentic-orchestrator/internal/domain/stream"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type streamEventRecord struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement"`
	EventID       string    `gorm:"column:event_id;size:255;not null;uniqueIndex"`
	RunID         string    `gorm:"column:run_id;size:255;index:idx_stream_corr,priority:1"`
	TaskID        string    `gorm:"column:task_id;size:255;index:idx_stream_corr,priority:2"`
	JobID         string    `gorm:"column:job_id;size:255;index:idx_stream_corr,priority:3"`
	ProjectID     string    `gorm:"column:project_id;size:255;index:idx_stream_corr,priority:4"`
	SessionID     string    `gorm:"column:session_id;size:255;index:idx_stream_session"`
	CorrelationID string    `gorm:"column:correlation_id;size:255;not null;index:idx_stream_corr_id"`
	Source        string    `gorm:"column:source;size:64;not null"`
	EventType     string    `gorm:"column:event_type;size:128;not null;index"`
	PayloadJSON   string    `gorm:"column:payload_json;type:text;not null"`
	OccurredAt    time.Time `gorm:"column:occurred_at;not null;index"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;autoCreateTime"`
}

func (streamEventRecord) TableName() string {
	return "stream_events"
}

type EventStore struct {
	db      *gorm.DB
	watcher domainrealtime.TableChangeWatcher
}

func NewEventStore(db *gorm.DB) (*EventStore, error) {
	if db == nil {
		return nil, fmt.Errorf("stream event store: db is required")
	}
	if err := db.AutoMigrate(&streamEventRecord{}); err != nil {
		return nil, fmt.Errorf("stream event store: migrate: %w", err)
	}
	return &EventStore{db: db}, nil
}

func (store *EventStore) Append(ctx context.Context, event domainstream.Event) (domainstream.Event, error) {
	if store == nil || store.db == nil {
		return domainstream.Event{}, fmt.Errorf("stream event store: db is not initialized")
	}
	if err := event.Validate(); err != nil {
		return domainstream.Event{}, err
	}
	redactedPayload, _ := domainsecurity.RedactPayload(event.Payload)
	event.Payload = redactedPayload
	payloadJSON, err := json.Marshal(redactedPayload)
	if err != nil {
		return domainstream.Event{}, fmt.Errorf("stream event store: marshal payload: %w", err)
	}
	record := streamEventRecord{
		EventID:       strings.TrimSpace(event.EventID),
		RunID:         strings.TrimSpace(event.CorrelationIDs.RunID),
		TaskID:        strings.TrimSpace(event.CorrelationIDs.TaskID),
		JobID:         strings.TrimSpace(event.CorrelationIDs.JobID),
		ProjectID:     strings.TrimSpace(event.CorrelationIDs.ProjectID),
		SessionID:     strings.TrimSpace(event.CorrelationIDs.SessionID),
		CorrelationID: strings.TrimSpace(event.CorrelationIDs.CorrelationID),
		Source:        strings.TrimSpace(string(event.Source)),
		EventType:     strings.TrimSpace(string(event.EventType)),
		PayloadJSON:   string(payloadJSON),
		OccurredAt:    event.OccurredAt.UTC(),
	}
	if err := store.db.WithContext(ctx).Create(&record).Error; err != nil {
		return domainstream.Event{}, fmt.Errorf("stream event store: append: %w", err)
	}
	event.StreamOffset = record.ID
	store.publishTableChangeSignal(ctx, event, record)
	return event, nil
}

func (store *EventStore) SetTableChangeWatcher(watcher domainrealtime.TableChangeWatcher) {
	if store == nil {
		return
	}
	store.watcher = watcher
}

func (store *EventStore) publishTableChangeSignal(ctx context.Context, event domainstream.Event, record streamEventRecord) {
	if store == nil || store.watcher == nil {
		return
	}
	_ = store.watcher.Publish(ctx, domainrealtime.TableChangeEvent{
		Topic:           "stream_events_live",
		Table:           "stream_events",
		Operation:       "insert",
		ProjectID:       strings.TrimSpace(record.ProjectID),
		RunID:           strings.TrimSpace(record.RunID),
		TaskID:          strings.TrimSpace(record.TaskID),
		JobID:           strings.TrimSpace(record.JobID),
		SessionID:       strings.TrimSpace(record.SessionID),
		ProjectEventSeq: int64(record.ID),
		SessionEventSeq: int64(record.ID),
		OccurredAt:      record.OccurredAt.UTC(),
		Payload: map[string]any{
			"event_id":       strings.TrimSpace(event.EventID),
			"stream_offset":  int64(event.StreamOffset),
			"occurred_at":    event.OccurredAt.UTC().Format(time.RFC3339Nano),
			"source":         strings.TrimSpace(string(event.Source)),
			"event_type":     strings.TrimSpace(string(event.EventType)),
			"run_id":         strings.TrimSpace(event.CorrelationIDs.RunID),
			"task_id":        strings.TrimSpace(event.CorrelationIDs.TaskID),
			"job_id":         strings.TrimSpace(event.CorrelationIDs.JobID),
			"project_id":     strings.TrimSpace(event.CorrelationIDs.ProjectID),
			"session_id":     strings.TrimSpace(event.CorrelationIDs.SessionID),
			"correlation_id": strings.TrimSpace(event.CorrelationIDs.CorrelationID),
			"payload":        event.Payload,
		},
	})
	_ = store.watcher.Publish(ctx, domainrealtime.TableChangeEvent{
		Topic:           "stream_events_changed",
		Table:           "stream_events",
		Operation:       "insert",
		ProjectID:       strings.TrimSpace(record.ProjectID),
		RunID:           strings.TrimSpace(record.RunID),
		TaskID:          strings.TrimSpace(record.TaskID),
		JobID:           strings.TrimSpace(record.JobID),
		SessionID:       strings.TrimSpace(record.SessionID),
		ProjectEventSeq: int64(record.ID),
		SessionEventSeq: int64(record.ID),
		OccurredAt:      record.OccurredAt.UTC(),
	})
}

func (store *EventStore) ListFromOffset(ctx context.Context, offset uint64, limit int) ([]domainstream.Event, error) {
	if store == nil || store.db == nil {
		return nil, fmt.Errorf("stream event store: db is not initialized")
	}
	if limit <= 0 {
		limit = applicationstream.DefaultReplayLimit
	}
	var records []streamEventRecord
	if err := store.db.WithContext(ctx).
		Where("id > ?", offset).
		Order("id ASC").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("stream event store: list from offset: %w", err)
	}
	events := make([]domainstream.Event, 0, len(records))
	for _, record := range records {
		payload := map[string]any{}
		if err := json.Unmarshal([]byte(record.PayloadJSON), &payload); err != nil {
			return nil, fmt.Errorf("stream event store: unmarshal payload: %w", err)
		}
		event := domainstream.Event{
			EventID:      record.EventID,
			StreamOffset: record.ID,
			OccurredAt:   record.OccurredAt,
			Source:       domainstream.Source(record.Source),
			EventType:    domainstream.EventType(record.EventType),
			CorrelationIDs: domainstream.CorrelationIDs{
				RunID:         record.RunID,
				TaskID:        record.TaskID,
				JobID:         record.JobID,
				ProjectID:     record.ProjectID,
				SessionID:     record.SessionID,
				CorrelationID: record.CorrelationID,
			},
			Payload: payload,
		}
		if err := event.Validate(); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

var _ applicationstream.EventStore = (*EventStore)(nil)
