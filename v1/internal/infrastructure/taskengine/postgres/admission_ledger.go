package postgres

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type admissionLedgerRecord struct {
	gorm.Model
	RunID          string `gorm:"column:run_id;size:255;not null;uniqueIndex:idx_workflow_job_identity,priority:1"`
	TaskID         string `gorm:"column:task_id;size:255;not null;uniqueIndex:idx_workflow_job_identity,priority:2"`
	JobID          string `gorm:"column:job_id;size:255;not null;uniqueIndex:idx_workflow_job_identity,priority:3"`
	IdempotencyKey string `gorm:"column:idempotency_key;size:255;not null;uniqueIndex:idx_workflow_job_identity,priority:4"`
	JobKind        string `gorm:"column:job_kind;size:128;not null"`
	QueueTaskID    string `gorm:"column:queue_task_id;size:255;not null"`
	Queue          string `gorm:"column:queue;size:128;not null"`
	Status         string `gorm:"column:status;size:64;not null"`
	Duplicate      bool   `gorm:"column:duplicate;not null;default:false"`
	EnqueuedAtUnix int64  `gorm:"column:enqueued_at_unix;not null"`
}

func (admissionLedgerRecord) TableName() string {
	return "workflow_jobs"
}

type AdmissionLedger struct {
	db *gorm.DB
}

func NewAdmissionLedger(db *gorm.DB) (*AdmissionLedger, error) {
	if db == nil {
		return nil, fmt.Errorf("admission ledger db is required")
	}
	if err := db.AutoMigrate(&admissionLedgerRecord{}); err != nil {
		return nil, fmt.Errorf("admission ledger migrate: %w", err)
	}
	return &AdmissionLedger{db: db}, nil
}

func (ledger *AdmissionLedger) Upsert(ctx context.Context, record taskengine.AdmissionRecord) error {
	if ledger == nil || ledger.db == nil {
		return fmt.Errorf("admission ledger is not initialized")
	}
	if err := record.Validate(); err != nil {
		return err
	}
	model := admissionLedgerRecord{
		RunID:          strings.TrimSpace(record.RunID),
		TaskID:         strings.TrimSpace(record.TaskID),
		JobID:          strings.TrimSpace(record.JobID),
		IdempotencyKey: strings.TrimSpace(record.IdempotencyKey),
		JobKind:        strings.TrimSpace(string(record.JobKind)),
		QueueTaskID:    strings.TrimSpace(record.QueueTaskID),
		Queue:          strings.TrimSpace(record.Queue),
		Status:         strings.TrimSpace(string(record.Status)),
		Duplicate:      record.Duplicate,
		EnqueuedAtUnix: record.EnqueuedAt.Unix(),
	}
	if err := ledger.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "run_id"}, {Name: "task_id"}, {Name: "job_id"}, {Name: "idempotency_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"job_kind", "queue_task_id", "queue", "status", "duplicate", "enqueued_at_unix", "updated_at"}),
	}).Create(&model).Error; err != nil {
		return fmt.Errorf("admission ledger upsert: %w", err)
	}
	return nil
}

var _ taskengine.AdmissionLedger = (*AdmissionLedger)(nil)
