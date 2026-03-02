package postgres

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type executionRecord struct {
	gorm.Model
	RunID          string `gorm:"column:run_id;size:255;not null;uniqueIndex:idx_job_execution_identity,priority:1"`
	TaskID         string `gorm:"column:task_id;size:255;not null;uniqueIndex:idx_job_execution_identity,priority:2"`
	JobID          string `gorm:"column:job_id;size:255;not null;uniqueIndex:idx_job_execution_identity,priority:3"`
	ProjectID      string `gorm:"column:project_id;size:255;index"`
	JobKind        string `gorm:"column:job_kind;not null"`
	IdempotencyKey string `gorm:"column:idempotency_key;not null"`
	Step           string `gorm:"column:step;not null"`
	Status         string `gorm:"column:status;not null"`
	ErrorMessage   string `gorm:"column:error_message"`
}

func (executionRecord) TableName() string {
	return "job_execution_events"
}

type PostgresExecutionJournal struct {
	db *gorm.DB
}

func NewPostgresExecutionJournal(db *gorm.DB) (*PostgresExecutionJournal, error) {
	if db == nil {
		return nil, fmt.Errorf("postgres execution journal: db is required")
	}
	if err := db.AutoMigrate(&executionRecord{}); err != nil {
		return nil, fmt.Errorf("postgres execution journal: migrate: %w", err)
	}
	return &PostgresExecutionJournal{db: db}, nil
}

func (journal *PostgresExecutionJournal) Upsert(ctx context.Context, record taskengine.ExecutionRecord) error {
	if journal == nil || journal.db == nil {
		return fmt.Errorf("postgres execution journal: db is not initialized")
	}
	if err := record.Validate(); err != nil {
		return err
	}
	model := executionRecord{
		RunID:          strings.TrimSpace(record.RunID),
		TaskID:         strings.TrimSpace(record.TaskID),
		JobID:          strings.TrimSpace(record.JobID),
		ProjectID:      strings.TrimSpace(record.ProjectID),
		JobKind:        strings.TrimSpace(string(record.JobKind)),
		IdempotencyKey: strings.TrimSpace(record.IdempotencyKey),
		Step:           strings.TrimSpace(record.Step),
		Status:         strings.TrimSpace(string(record.Status)),
		ErrorMessage:   strings.TrimSpace(record.ErrorMessage),
	}
	if err := journal.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "run_id"}, {Name: "task_id"}, {Name: "job_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"project_id", "job_kind", "idempotency_key", "step", "status", "error_message", "updated_at"}),
	}).Create(&model).Error; err != nil {
		return fmt.Errorf("postgres execution journal: upsert: %w", err)
	}
	return nil
}

func (journal *PostgresExecutionJournal) Load(ctx context.Context, runID string, taskID string, jobID string) (*taskengine.ExecutionRecord, error) {
	if journal == nil || journal.db == nil {
		return nil, fmt.Errorf("postgres execution journal: db is not initialized")
	}
	runID = strings.TrimSpace(runID)
	taskID = strings.TrimSpace(taskID)
	jobID = strings.TrimSpace(jobID)
	if runID == "" || taskID == "" || jobID == "" {
		return nil, fmt.Errorf("postgres execution journal: run_id, task_id, job_id are required")
	}
	var model executionRecord
	err := journal.db.WithContext(ctx).First(&model, "run_id = ? AND task_id = ? AND job_id = ?", runID, taskID, jobID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("postgres execution journal: load: %w", err)
	}
	record := taskengine.ExecutionRecord{
		RunID:          model.RunID,
		TaskID:         model.TaskID,
		JobID:          model.JobID,
		ProjectID:      model.ProjectID,
		JobKind:        taskengine.JobKind(model.JobKind),
		IdempotencyKey: model.IdempotencyKey,
		Step:           model.Step,
		Status:         taskengine.ExecutionStatus(model.Status),
		ErrorMessage:   model.ErrorMessage,
		UpdatedAt:      model.UpdatedAt,
	}
	if err := record.Validate(); err != nil {
		return nil, err
	}
	return &record, nil
}

var _ taskengine.ExecutionJournal = (*PostgresExecutionJournal)(nil)
