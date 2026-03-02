package postgres

import (
	applicationstream "agentic-orchestrator/internal/application/stream"
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
	db *gorm.DB
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
	payloadJSON, err := json.Marshal(event.Payload)
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
	return event, nil
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
