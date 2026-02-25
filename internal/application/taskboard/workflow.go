package taskboard

import (
	"context"
	"strings"
	"time"
)

type WorkflowStatus string

const (
	WorkflowStatusQueued    WorkflowStatus = "queued"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCanceled  WorkflowStatus = "canceled"
)

const (
	WorkflowTaskTypeCopilotDecompose = "copilot.decompose"
	WorkflowTaskTypeGitWorktreeFlow  = "git.worktree.flow"
	WorkflowTaskTypeGitConflict      = "git.conflict.resolve"
	WorkflowTaskTypeTaskboardExecute = "taskboard.execute"
)

type IngestionWorkflow struct {
	RunID      string         `json:"run_id"`
	TaskID     string         `json:"task_id,omitempty"`
	TaskType   string         `json:"task_type,omitempty"`
	Status     WorkflowStatus `json:"status"`
	Message    string         `json:"message,omitempty"`
	Stream     string         `json:"stream,omitempty"`
	BoardID    string         `json:"board_id,omitempty"`
	UpdatedAt  time.Time      `json:"updated_at"`
	CreatedAt  time.Time      `json:"created_at"`
	Cancelable bool           `json:"cancelable,omitempty"`
}

type WorkflowCancelResult struct {
	RunID             string `json:"run_id"`
	MatchedTasks      int    `json:"matched_tasks"`
	CanceledTasks     int    `json:"canceled_tasks"`
	SignaledActive    int    `json:"signaled_active_tasks"`
	UncancelableTasks int    `json:"uncancelable_tasks"`
}

func IsIngestionWorkflowTaskType(taskType string) bool {
	switch strings.TrimSpace(taskType) {
	case WorkflowTaskTypeCopilotDecompose:
		return true
	default:
		return false
	}
}

func IsWorktreeWorkflowTaskType(taskType string) bool {
	switch strings.TrimSpace(taskType) {
	case WorkflowTaskTypeTaskboardExecute, WorkflowTaskTypeGitWorktreeFlow, WorkflowTaskTypeGitConflict:
		return true
	default:
		return false
	}
}

func (workflow *IngestionWorkflow) Normalize(runID string) {
	now := time.Now().UTC()
	if workflow.CreatedAt.IsZero() {
		workflow.CreatedAt = now
	}
	workflow.UpdatedAt = now
	workflow.RunID = strings.TrimSpace(runID)
}

type WorkflowRepository interface {
	GetWorkflow(ctx context.Context, runID string) (*IngestionWorkflow, error)
	ListWorkflows(ctx context.Context) ([]IngestionWorkflow, error)
	SaveWorkflow(ctx context.Context, workflow *IngestionWorkflow) error
}
