package postgres

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ControlPlaneQueryRepository struct {
	db *gorm.DB
}

func NewControlPlaneQueryRepository(db *gorm.DB) (*ControlPlaneQueryRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("control-plane query repository db is required")
	}
	return &ControlPlaneQueryRepository{db: db}, nil
}

func (repository *ControlPlaneQueryRepository) ListSessions(ctx context.Context, limit int) ([]applicationcontrolplane.SessionSummary, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("control-plane query repository is not initialized")
	}
	if limit <= 0 {
		limit = 50
	}
	type row struct {
		RunID        string `gorm:"column:run_id"`
		TaskCount    int    `gorm:"column:task_count"`
		JobCount     int    `gorm:"column:job_count"`
		UpdatedAtRaw string `gorm:"column:updated_at"`
	}
	rows := make([]row, 0)
	err := repository.db.WithContext(ctx).
		Model(&admissionLedgerRecord{}).
		Select("run_id, COUNT(DISTINCT task_id) as task_count, COUNT(*) as job_count, MAX(updated_at) as updated_at").
		Group("run_id").
		Order("MAX(updated_at) DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	result := make([]applicationcontrolplane.SessionSummary, 0, len(rows))
	for _, item := range rows {
		updatedAt, parseErr := parseAggregateTime(item.UpdatedAtRaw)
		if parseErr != nil {
			return nil, fmt.Errorf("list sessions: parse updated_at: %w", parseErr)
		}
		result = append(result, applicationcontrolplane.SessionSummary{RunID: item.RunID, TaskCount: item.TaskCount, JobCount: item.JobCount, UpdatedAt: updatedAt.UTC()})
	}
	return result, nil
}

func (repository *ControlPlaneQueryRepository) GetSession(ctx context.Context, runID string) (*applicationcontrolplane.SessionSummary, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("control-plane query repository is not initialized")
	}
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil, fmt.Errorf("run_id is required")
	}
	type row struct {
		RunID        string `gorm:"column:run_id"`
		TaskCount    int    `gorm:"column:task_count"`
		JobCount     int    `gorm:"column:job_count"`
		UpdatedAtRaw string `gorm:"column:updated_at"`
	}
	rows := make([]row, 0, 1)
	err := repository.db.WithContext(ctx).
		Model(&admissionLedgerRecord{}).
		Select("run_id, COUNT(DISTINCT task_id) as task_count, COUNT(*) as job_count, MAX(updated_at) as updated_at").
		Where("run_id = ?", runID).
		Group("run_id").
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	updatedAt, parseErr := parseAggregateTime(rows[0].UpdatedAtRaw)
	if parseErr != nil {
		return nil, fmt.Errorf("get session: parse updated_at: %w", parseErr)
	}
	summary := applicationcontrolplane.SessionSummary{RunID: rows[0].RunID, TaskCount: rows[0].TaskCount, JobCount: rows[0].JobCount, UpdatedAt: updatedAt.UTC()}
	return &summary, nil
}

func (repository *ControlPlaneQueryRepository) ListWorkflowJobs(ctx context.Context, runID string, taskID string, limit int) ([]applicationcontrolplane.WorkflowJob, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("control-plane query repository is not initialized")
	}
	runID = strings.TrimSpace(runID)
	taskID = strings.TrimSpace(taskID)
	if runID == "" {
		return nil, fmt.Errorf("run_id is required")
	}
	if limit <= 0 {
		limit = 100
	}
	query := repository.db.WithContext(ctx).Model(&admissionLedgerRecord{}).Where("run_id = ?", runID)
	if taskID != "" {
		query = query.Where("task_id = ?", taskID)
	}
	records := make([]admissionLedgerRecord, 0)
	if err := query.Order("updated_at DESC").Limit(limit).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list workflow jobs: %w", err)
	}
	result := make([]applicationcontrolplane.WorkflowJob, 0, len(records))
	for _, record := range records {
		result = append(result, applicationcontrolplane.WorkflowJob{
			RunID:          record.RunID,
			TaskID:         record.TaskID,
			JobID:          record.JobID,
			ProjectID:      record.ProjectID,
			JobKind:        taskengine.JobKind(record.JobKind),
			IdempotencyKey: record.IdempotencyKey,
			QueueTaskID:    record.QueueTaskID,
			Queue:          record.Queue,
			Status:         record.Status,
			Duplicate:      record.Duplicate,
			EnqueuedAt:     time.Unix(record.EnqueuedAtUnix, 0).UTC(),
			UpdatedAt:      record.UpdatedAt.UTC(),
		})
	}
	return result, nil
}

func (repository *ControlPlaneQueryRepository) ListExecutionHistory(ctx context.Context, filter applicationcontrolplane.CorrelationFilter, limit int) ([]applicationcontrolplane.ExecutionHistoryRecord, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("control-plane query repository is not initialized")
	}
	if limit <= 0 {
		limit = 100
	}
	query := repository.db.WithContext(ctx).Model(&executionRecord{}).Where("run_id = ? AND task_id = ?", filter.RunID, filter.TaskID)
	if strings.TrimSpace(filter.JobID) != "" {
		query = query.Where("job_id = ?", strings.TrimSpace(filter.JobID))
	}
	if strings.TrimSpace(filter.ProjectID) != "" {
		query = query.Where("project_id = ?", strings.TrimSpace(filter.ProjectID))
	}
	records := make([]executionRecord, 0)
	if err := query.Order("updated_at DESC").Limit(limit).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list execution history: %w", err)
	}
	result := make([]applicationcontrolplane.ExecutionHistoryRecord, 0, len(records))
	for _, record := range records {
		result = append(result, applicationcontrolplane.ExecutionHistoryRecord{
			RunID:          record.RunID,
			TaskID:         record.TaskID,
			JobID:          record.JobID,
			ProjectID:      record.ProjectID,
			JobKind:        taskengine.JobKind(record.JobKind),
			IdempotencyKey: record.IdempotencyKey,
			Step:           record.Step,
			Status:         taskengine.ExecutionStatus(record.Status),
			ErrorMessage:   record.ErrorMessage,
			UpdatedAt:      record.UpdatedAt.UTC(),
		})
	}
	return result, nil
}

func (repository *ControlPlaneQueryRepository) ListDeadLetterHistory(ctx context.Context, queue string, limit int) ([]applicationcontrolplane.DeadLetterHistoryRecord, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("control-plane query repository is not initialized")
	}
	if limit <= 0 {
		limit = 100
	}
	query := repository.db.WithContext(ctx).Model(&deadLetterEventRecord{})
	if strings.TrimSpace(queue) != "" {
		query = query.Where("queue = ?", strings.TrimSpace(queue))
	}
	records := make([]deadLetterEventRecord, 0)
	if err := query.Order("occurred_at_unix DESC").Limit(limit).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list dead letter history: %w", err)
	}
	result := make([]applicationcontrolplane.DeadLetterHistoryRecord, 0, len(records))
	for _, record := range records {
		result = append(result, applicationcontrolplane.DeadLetterHistoryRecord{
			Queue:      record.Queue,
			TaskID:     record.TaskID,
			JobKind:    taskengine.JobKind(record.JobKind),
			Action:     taskengine.DeadLetterAction(record.Action),
			LastError:  record.LastError,
			Reason:     record.Reason,
			Actor:      record.Actor,
			OccurredAt: time.Unix(record.OccurredAtUnix, 0).UTC(),
		})
	}
	return result, nil
}

func parseAggregateTime(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("empty aggregate time")
	}
	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format %q", trimmed)
}

var _ applicationcontrolplane.QueryRepository = (*ControlPlaneQueryRepository)(nil)
