package postgres

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type deadLetterEventRecord struct {
	gorm.Model
	Queue          string `gorm:"column:queue;size:128;not null;index:idx_dead_letter_queue_task,priority:1"`
	TaskID         string `gorm:"column:task_id;size:255;not null;index:idx_dead_letter_queue_task,priority:2"`
	JobKind        string `gorm:"column:job_kind;size:128"`
	Action         string `gorm:"column:action;size:64;not null"`
	LastError      string `gorm:"column:last_error"`
	Reason         string `gorm:"column:reason"`
	Actor          string `gorm:"column:actor"`
	OccurredAtUnix int64  `gorm:"column:occurred_at_unix;not null"`
}

func (deadLetterEventRecord) TableName() string {
	return "dead_letter_events"
}

type DeadLetterAudit struct {
	db *gorm.DB
}

func NewDeadLetterAudit(db *gorm.DB) (*DeadLetterAudit, error) {
	if db == nil {
		return nil, fmt.Errorf("dead-letter audit db is required")
	}
	if err := db.AutoMigrate(&deadLetterEventRecord{}); err != nil {
		return nil, fmt.Errorf("dead-letter audit migrate: %w", err)
	}
	return &DeadLetterAudit{db: db}, nil
}

func (audit *DeadLetterAudit) Record(ctx context.Context, event taskengine.DeadLetterEvent) error {
	if audit == nil || audit.db == nil {
		return fmt.Errorf("dead-letter audit is not initialized")
	}
	if err := event.Validate(); err != nil {
		return err
	}
	model := deadLetterEventRecord{
		Queue:          strings.TrimSpace(event.Queue),
		TaskID:         strings.TrimSpace(event.TaskID),
		JobKind:        strings.TrimSpace(string(event.JobKind)),
		Action:         strings.TrimSpace(string(event.Action)),
		LastError:      strings.TrimSpace(event.LastError),
		Reason:         strings.TrimSpace(event.Reason),
		Actor:          strings.TrimSpace(event.Actor),
		OccurredAtUnix: event.OccurredAt.Unix(),
	}
	if err := audit.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("dead-letter audit record: %w", err)
	}
	return nil
}

var _ taskengine.DeadLetterAudit = (*DeadLetterAudit)(nil)
