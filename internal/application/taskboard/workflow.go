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
)

type IngestionWorkflow struct {
	RunID     string         `json:"run_id"`
	TaskID    string         `json:"task_id,omitempty"`
	Status    WorkflowStatus `json:"status"`
	Message   string         `json:"message,omitempty"`
	Stream    string         `json:"stream,omitempty"`
	BoardID   string         `json:"board_id,omitempty"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedAt time.Time      `json:"created_at"`
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
