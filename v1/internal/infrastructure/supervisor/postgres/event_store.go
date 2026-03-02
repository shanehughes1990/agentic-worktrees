package postgres

import (
	applicationsupervisor "agentic-orchestrator/internal/application/supervisor"
	"agentic-orchestrator/internal/domain/failures"
	domainsupervisor "agentic-orchestrator/internal/domain/supervisor"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type supervisorEventRecord struct {
	ID            uint      `gorm:"primaryKey"`
	RunID         string    `gorm:"column:run_id;size:255;not null;index:idx_supervisor_corr,priority:1"`
	TaskID        string    `gorm:"column:task_id;size:255;not null;index:idx_supervisor_corr,priority:2"`
	JobID         string    `gorm:"column:job_id;size:255;not null;index:idx_supervisor_corr,priority:3"`
	ProjectID     string    `gorm:"column:project_id;size:255;index:idx_supervisor_corr,priority:4"`
	SignalType    string    `gorm:"column:signal_type;size:128;not null"`
	FromState     string    `gorm:"column:from_state;size:64;not null"`
	ToState       string    `gorm:"column:to_state;size:64;not null"`
	ActionCode    string    `gorm:"column:action_code;size:64;not null"`
	ReasonCode    string    `gorm:"column:reason_code;size:128;not null"`
	RuleName      string    `gorm:"column:rule_name;size:255;not null"`
	RulePriority  int       `gorm:"column:rule_priority;not null"`
	Attempt       int       `gorm:"column:attempt;not null;default:0"`
	MaxRetry      int       `gorm:"column:max_retry;not null;default:0"`
	FailureClass  string    `gorm:"column:failure_class;size:64"`
	AttentionZone string    `gorm:"column:attention_zone;size:64"`
	MetadataJSON  string    `gorm:"column:metadata_json;type:text"`
	OccurredAt    time.Time `gorm:"column:occurred_at;not null;index"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;autoCreateTime"`
}

func (supervisorEventRecord) TableName() string {
	return "supervisor_events"
}

type EventStore struct {
	db *gorm.DB
}

func NewEventStore(db *gorm.DB) (*EventStore, error) {
	if db == nil {
		return nil, fmt.Errorf("supervisor event store: db is required")
	}
	if err := db.AutoMigrate(&supervisorEventRecord{}); err != nil {
		return nil, fmt.Errorf("supervisor event store: migrate: %w", err)
	}
	return &EventStore{db: db}, nil
}

func (store *EventStore) Append(ctx context.Context, decision domainsupervisor.Decision) error {
	if store == nil || store.db == nil {
		return fmt.Errorf("supervisor event store: db is not initialized")
	}
	if err := decision.Validate(); err != nil {
		return err
	}
	metadataJSON, err := json.Marshal(decision.Metadata)
	if err != nil {
		return fmt.Errorf("supervisor event store: marshal metadata: %w", err)
	}
	record := supervisorEventRecord{
		RunID:         strings.TrimSpace(decision.CorrelationIDs.RunID),
		TaskID:        strings.TrimSpace(decision.CorrelationIDs.TaskID),
		JobID:         strings.TrimSpace(decision.CorrelationIDs.JobID),
		ProjectID:     strings.TrimSpace(decision.CorrelationIDs.ProjectID),
		SignalType:    strings.TrimSpace(string(decision.SignalType)),
		FromState:     strings.TrimSpace(string(decision.FromState)),
		ToState:       strings.TrimSpace(string(decision.ToState)),
		ActionCode:    strings.TrimSpace(string(decision.Action)),
		ReasonCode:    strings.TrimSpace(string(decision.Reason)),
		RuleName:      strings.TrimSpace(decision.RuleName),
		RulePriority:  decision.RulePriority,
		Attempt:       decision.Attempt,
		MaxRetry:      decision.MaxRetry,
		FailureClass:  strings.TrimSpace(string(decision.FailureClass)),
		AttentionZone: strings.TrimSpace(string(decision.AttentionZone)),
		MetadataJSON:  string(metadataJSON),
		OccurredAt:    decision.OccurredAt,
	}
	if err := store.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("supervisor event store: append: %w", err)
	}
	return nil
}

func (store *EventStore) ListByCorrelation(ctx context.Context, correlation domainsupervisor.CorrelationIDs) ([]domainsupervisor.Decision, error) {
	if store == nil || store.db == nil {
		return nil, fmt.Errorf("supervisor event store: db is not initialized")
	}
	if err := correlation.Validate(); err != nil {
		return nil, err
	}
	var records []supervisorEventRecord
	query := store.db.WithContext(ctx).
		Order("id ASC").
		Where("run_id = ? AND task_id = ? AND job_id = ?", strings.TrimSpace(correlation.RunID), strings.TrimSpace(correlation.TaskID), strings.TrimSpace(correlation.JobID))
	if strings.TrimSpace(correlation.ProjectID) != "" {
		query = query.Where("project_id = ?", strings.TrimSpace(correlation.ProjectID))
	}
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("supervisor event store: list by correlation: %w", err)
	}
	decisions := make([]domainsupervisor.Decision, 0, len(records))
	for _, record := range records {
		metadata := map[string]string{}
		if strings.TrimSpace(record.MetadataJSON) != "" {
			if err := json.Unmarshal([]byte(record.MetadataJSON), &metadata); err != nil {
				return nil, fmt.Errorf("supervisor event store: unmarshal metadata: %w", err)
			}
		}
		decision := domainsupervisor.Decision{
			CorrelationIDs: domainsupervisor.CorrelationIDs{RunID: record.RunID, TaskID: record.TaskID, JobID: record.JobID, ProjectID: record.ProjectID},
			SignalType:      domainsupervisor.SignalType(record.SignalType),
			FromState:       domainsupervisor.State(record.FromState),
			ToState:         domainsupervisor.State(record.ToState),
			Action:          domainsupervisor.ActionCode(record.ActionCode),
			Reason:          domainsupervisor.ReasonCode(record.ReasonCode),
			RuleName:        record.RuleName,
			RulePriority:    record.RulePriority,
			OccurredAt:      record.OccurredAt,
			Attempt:         record.Attempt,
			MaxRetry:        record.MaxRetry,
			FailureClass:    failures.Class(record.FailureClass),
			AttentionZone:   domainsupervisor.AttentionZone(record.AttentionZone),
			Metadata:        metadata,
		}
		if err := decision.Validate(); err != nil {
			return nil, err
		}
		decisions = append(decisions, decision)
	}
	return decisions, nil
}

var _ applicationsupervisor.EventStore = (*EventStore)(nil)
