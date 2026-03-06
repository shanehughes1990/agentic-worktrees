package postgres

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ControlPlaneQueryRepository struct {
	db *gorm.DB
}

type lifecycleTreeSnapshotRow struct {
	gorm.Model
	ProjectID       string    `gorm:"column:project_id"`
	RunID           string    `gorm:"column:run_id"`
	TaskID          string    `gorm:"column:task_id"`
	JobID           string    `gorm:"column:job_id"`
	SessionID       string    `gorm:"column:session_id"`
	PipelineType    string    `gorm:"column:pipeline_type"`
	SourceRuntime   string    `gorm:"column:source_runtime"`
	CurrentState    string    `gorm:"column:current_state"`
	CurrentSeverity string    `gorm:"column:current_severity"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
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
		gorm.Model
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
		gorm.Model
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

func (repository *ControlPlaneQueryRepository) ListLifecycleSessionSnapshots(ctx context.Context, projectID string, pipelineType string, limit int) ([]applicationcontrolplane.LifecycleSessionSnapshot, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("control-plane query repository is not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	pipelineType = strings.TrimSpace(pipelineType)
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if limit <= 0 {
		limit = 100
	}

	type row struct {
		gorm.Model
		ProjectID           string     `gorm:"column:project_id"`
		RunID               string     `gorm:"column:run_id"`
		TaskID              string     `gorm:"column:task_id"`
		JobID               string     `gorm:"column:job_id"`
		SessionID           string     `gorm:"column:session_id"`
		PipelineType        string     `gorm:"column:pipeline_type"`
		SourceRuntime       string     `gorm:"column:source_runtime"`
		CurrentState        string     `gorm:"column:current_state"`
		CurrentSeverity     string     `gorm:"column:current_severity"`
		LastReasonCode      string     `gorm:"column:last_reason_code"`
		LastReasonSummary   string     `gorm:"column:last_reason_summary"`
		LastEventSeq        int64      `gorm:"column:last_event_seq"`
		LastProjectEventSeq int64      `gorm:"column:last_project_event_seq"`
		LastLivenessAt      *time.Time `gorm:"column:last_liveness_at"`
		LastActivityAt      *time.Time `gorm:"column:last_activity_at"`
		LastCheckpointAt    *time.Time `gorm:"column:last_checkpoint_at"`
		StartedAt           time.Time  `gorm:"column:started_at"`
		EndedAt             *time.Time `gorm:"column:ended_at"`
		UpdatedAt           time.Time  `gorm:"column:updated_at"`
	}

	rows := make([]row, 0)
	query := repository.db.WithContext(ctx).
		Table("project_sessions").
		Select("project_id, run_id, task_id, job_id, session_id, pipeline_type, source_runtime, current_state, current_severity, last_reason_code, last_reason_summary, last_event_seq, last_project_event_seq, last_liveness_at, last_activity_at, last_checkpoint_at, started_at, ended_at, updated_at").
		Where("project_id = ?", projectID)
	if pipelineType != "" {
		query = query.Where("pipeline_type = ?", pipelineType)
	}
	if err := query.Order("updated_at DESC").Limit(limit).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list lifecycle session snapshots: %w", err)
	}

	result := make([]applicationcontrolplane.LifecycleSessionSnapshot, 0, len(rows))
	for _, value := range rows {
		result = append(result, applicationcontrolplane.LifecycleSessionSnapshot{
			ProjectID:           value.ProjectID,
			RunID:               value.RunID,
			TaskID:              value.TaskID,
			JobID:               value.JobID,
			SessionID:           value.SessionID,
			PipelineType:        value.PipelineType,
			SourceRuntime:       value.SourceRuntime,
			CurrentState:        value.CurrentState,
			CurrentSeverity:     value.CurrentSeverity,
			LastReasonCode:      value.LastReasonCode,
			LastReasonSummary:   value.LastReasonSummary,
			LastEventSeq:        value.LastEventSeq,
			LastProjectEventSeq: value.LastProjectEventSeq,
			LastLivenessAt:      utcTimePtr(value.LastLivenessAt),
			LastActivityAt:      utcTimePtr(value.LastActivityAt),
			LastCheckpointAt:    utcTimePtr(value.LastCheckpointAt),
			StartedAt:           value.StartedAt.UTC(),
			EndedAt:             utcTimePtr(value.EndedAt),
			UpdatedAt:           value.UpdatedAt.UTC(),
		})
	}
	return result, nil
}

func (repository *ControlPlaneQueryRepository) ListLifecycleSessionHistory(ctx context.Context, projectID string, sessionID string, fromEventSeq int64, limit int) ([]applicationcontrolplane.LifecycleHistoryEvent, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("control-plane query repository is not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	if fromEventSeq < 0 {
		fromEventSeq = 0
	}
	if limit <= 0 {
		limit = 100
	}

	type row struct {
		gorm.Model
		EventID         string    `gorm:"column:event_id"`
		ProjectID       string    `gorm:"column:project_id"`
		RunID           string    `gorm:"column:run_id"`
		TaskID          string    `gorm:"column:task_id"`
		JobID           string    `gorm:"column:job_id"`
		SessionID       string    `gorm:"column:session_id"`
		PipelineType    string    `gorm:"column:pipeline_type"`
		SourceRuntime   string    `gorm:"column:source_runtime"`
		EventType       string    `gorm:"column:event_type"`
		EventSeq        int64     `gorm:"column:event_seq"`
		ProjectEventSeq int64     `gorm:"column:project_event_seq"`
		OccurredAt      time.Time `gorm:"column:occurred_at"`
		PayloadJSON     string    `gorm:"column:payload_json"`
	}

	rows := make([]row, 0)
	err := repository.db.WithContext(ctx).
		Table("project_session_history").
		Select("event_id, project_id, run_id, task_id, job_id, session_id, pipeline_type, source_runtime, event_type, event_seq, project_event_seq, occurred_at, payload_json").
		Where("project_id = ? AND session_id = ? AND event_seq > ?", projectID, sessionID, fromEventSeq).
		Order("event_seq ASC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list lifecycle session history: %w", err)
	}

	result := make([]applicationcontrolplane.LifecycleHistoryEvent, 0, len(rows))
	for _, value := range rows {
		result = append(result, applicationcontrolplane.LifecycleHistoryEvent{
			EventID:         value.EventID,
			ProjectID:       value.ProjectID,
			RunID:           value.RunID,
			TaskID:          value.TaskID,
			JobID:           value.JobID,
			SessionID:       value.SessionID,
			PipelineType:    value.PipelineType,
			SourceRuntime:   value.SourceRuntime,
			EventType:       value.EventType,
			EventSeq:        value.EventSeq,
			ProjectEventSeq: value.ProjectEventSeq,
			OccurredAt:      value.OccurredAt.UTC(),
			PayloadJSON:     value.PayloadJSON,
		})
	}
	return result, nil
}

func (repository *ControlPlaneQueryRepository) ListLifecycleTreeNodes(ctx context.Context, filter applicationcontrolplane.LifecycleTreeFilter, limit int) ([]applicationcontrolplane.LifecycleTreeNode, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("control-plane query repository is not initialized")
	}
	if strings.TrimSpace(filter.ProjectID) == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if limit <= 0 {
		limit = 200
	}

	rows := make([]lifecycleTreeSnapshotRow, 0)
	query := repository.db.WithContext(ctx).
		Table("project_sessions").
		Select("project_id, run_id, task_id, job_id, session_id, pipeline_type, source_runtime, current_state, current_severity, updated_at").
		Where("project_id = ?", strings.TrimSpace(filter.ProjectID))
	if strings.TrimSpace(filter.PipelineType) != "" {
		query = query.Where("pipeline_type = ?", strings.TrimSpace(filter.PipelineType))
	}
	if strings.TrimSpace(filter.RunID) != "" {
		query = query.Where("run_id = ?", strings.TrimSpace(filter.RunID))
	}
	if strings.TrimSpace(filter.TaskID) != "" {
		query = query.Where("task_id = ?", strings.TrimSpace(filter.TaskID))
	}
	if strings.TrimSpace(filter.JobID) != "" {
		query = query.Where("job_id = ?", strings.TrimSpace(filter.JobID))
	}
	if err := query.Order("updated_at DESC").Limit(limit).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list lifecycle tree nodes: %w", err)
	}

	runNodes := map[string]*applicationcontrolplane.LifecycleTreeNode{}
	taskNodes := map[string]*applicationcontrolplane.LifecycleTreeNode{}
	jobNodes := map[string]*applicationcontrolplane.LifecycleTreeNode{}
	sessionNodes := map[string]*applicationcontrolplane.LifecycleTreeNode{}

	for _, value := range rows {
		runID := strings.TrimSpace(value.RunID)
		taskID := strings.TrimSpace(value.TaskID)
		jobID := strings.TrimSpace(value.JobID)
		sessionID := strings.TrimSpace(value.SessionID)
		if runID == "" || taskID == "" || jobID == "" || sessionID == "" {
			continue
		}

		runNodeID := "run:" + runID
		taskNodeID := runNodeID + "/task:" + taskID
		jobNodeID := taskNodeID + "/job:" + jobID
		sessionNodeID := jobNodeID + "/session:" + sessionID

		runNode := ensureTreeNode(runNodes, runNodeID, "", applicationcontrolplane.LifecycleTreeNodeTypeRun, value)
		updateTreeNodeAggregate(runNode, value)
		taskNode := ensureTreeNode(taskNodes, taskNodeID, runNodeID, applicationcontrolplane.LifecycleTreeNodeTypeTask, value)
		updateTreeNodeAggregate(taskNode, value)
		jobNode := ensureTreeNode(jobNodes, jobNodeID, taskNodeID, applicationcontrolplane.LifecycleTreeNodeTypeJob, value)
		updateTreeNodeAggregate(jobNode, value)

		sessionNode := ensureTreeNode(sessionNodes, sessionNodeID, jobNodeID, applicationcontrolplane.LifecycleTreeNodeTypeSession, value)
		sessionNode.NodeID = sessionNodeID
		sessionNode.ParentNodeID = jobNodeID
		sessionNode.RunID = runID
		sessionNode.TaskID = taskID
		sessionNode.JobID = jobID
		sessionNode.SessionID = sessionID
		sessionNode.PipelineType = strings.TrimSpace(value.PipelineType)
		sessionNode.SourceRuntime = strings.TrimSpace(value.SourceRuntime)
		sessionNode.CurrentState = strings.TrimSpace(value.CurrentState)
		sessionNode.CurrentSeverity = strings.TrimSpace(value.CurrentSeverity)
		sessionNode.SessionCount = 1
		sessionNode.UpdatedAt = value.UpdatedAt.UTC()
	}

	result := make([]applicationcontrolplane.LifecycleTreeNode, 0, len(runNodes)+len(taskNodes)+len(jobNodes)+len(sessionNodes))
	for _, nodes := range []map[string]*applicationcontrolplane.LifecycleTreeNode{runNodes, taskNodes, jobNodes, sessionNodes} {
		for _, node := range nodes {
			result = append(result, *node)
		}
	}
	sort.SliceStable(result, func(left int, right int) bool {
		if result[left].UpdatedAt.Equal(result[right].UpdatedAt) {
			return result[left].NodeID < result[right].NodeID
		}
		return result[left].UpdatedAt.After(result[right].UpdatedAt)
	})
	return result, nil
}

func ensureTreeNode(store map[string]*applicationcontrolplane.LifecycleTreeNode, nodeID string, parentNodeID string, nodeType applicationcontrolplane.LifecycleTreeNodeType, value lifecycleTreeSnapshotRow) *applicationcontrolplane.LifecycleTreeNode {
	node, exists := store[nodeID]
	if exists {
		return node
	}
	node = &applicationcontrolplane.LifecycleTreeNode{
		NodeID:          nodeID,
		ParentNodeID:    parentNodeID,
		NodeType:        nodeType,
		ProjectID:       strings.TrimSpace(value.ProjectID),
		RunID:           strings.TrimSpace(value.RunID),
		TaskID:          strings.TrimSpace(value.TaskID),
		JobID:           strings.TrimSpace(value.JobID),
		SessionID:       strings.TrimSpace(value.SessionID),
		PipelineType:    strings.TrimSpace(value.PipelineType),
		SourceRuntime:   strings.TrimSpace(value.SourceRuntime),
		CurrentState:    strings.TrimSpace(value.CurrentState),
		CurrentSeverity: strings.TrimSpace(value.CurrentSeverity),
		SessionCount:    0,
		UpdatedAt:       value.UpdatedAt.UTC(),
	}
	store[nodeID] = node
	return node
}

func updateTreeNodeAggregate(node *applicationcontrolplane.LifecycleTreeNode, value lifecycleTreeSnapshotRow) {
	if node == nil {
		return
	}
	node.SessionCount++
	if value.UpdatedAt.After(node.UpdatedAt) {
		node.CurrentState = strings.TrimSpace(value.CurrentState)
		node.CurrentSeverity = strings.TrimSpace(value.CurrentSeverity)
		node.PipelineType = strings.TrimSpace(value.PipelineType)
		node.SourceRuntime = strings.TrimSpace(value.SourceRuntime)
		node.UpdatedAt = value.UpdatedAt.UTC()
	}
}

func utcTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	converted := value.UTC()
	return &converted
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
