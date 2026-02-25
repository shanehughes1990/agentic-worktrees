package taskboard

import (
	"context"
	"fmt"
	"strings"
)

type RuntimeWorkflowRepository interface {
	ListRuntimeWorkflows(ctx context.Context) ([]IngestionWorkflow, error)
	GetRuntimeWorkflow(ctx context.Context, runID string) (*IngestionWorkflow, error)
	CancelRuntimeWorkflow(ctx context.Context, runID string) (WorkflowCancelResult, error)
}

type RuntimeWorkflowService struct {
	repository RuntimeWorkflowRepository
}

func NewRuntimeWorkflowService(repository RuntimeWorkflowRepository) *RuntimeWorkflowService {
	return &RuntimeWorkflowService{repository: repository}
}

func (service *RuntimeWorkflowService) ListWorkflows(ctx context.Context) ([]IngestionWorkflow, error) {
	workflows, err := service.repository.ListRuntimeWorkflows(ctx)
	if err != nil {
		return nil, fmt.Errorf("list runtime workflows: %w", err)
	}
	return workflows, nil
}

func (service *RuntimeWorkflowService) GetWorkflowStatus(ctx context.Context, runID string) (*IngestionWorkflow, error) {
	cleanRunID := strings.TrimSpace(runID)
	if cleanRunID == "" {
		return nil, fmt.Errorf("run_id is required")
	}
	workflow, err := service.repository.GetRuntimeWorkflow(ctx, cleanRunID)
	if err != nil {
		return nil, fmt.Errorf("load runtime workflow status: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", cleanRunID)
	}
	return workflow, nil
}

func (service *RuntimeWorkflowService) CancelWorkflow(ctx context.Context, runID string) (WorkflowCancelResult, error) {
	cleanRunID := strings.TrimSpace(runID)
	if cleanRunID == "" {
		return WorkflowCancelResult{}, fmt.Errorf("run_id is required")
	}
	result, err := service.repository.CancelRuntimeWorkflow(ctx, cleanRunID)
	if err != nil {
		return WorkflowCancelResult{}, fmt.Errorf("cancel runtime workflow: %w", err)
	}
	return result, nil
}
