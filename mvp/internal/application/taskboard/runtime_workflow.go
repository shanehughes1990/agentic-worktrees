package taskboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type RuntimeWorkflowRepository interface {
	ListRuntimeWorkflows(ctx context.Context) ([]IngestionWorkflow, error)
	GetRuntimeWorkflow(ctx context.Context, runID string) (*IngestionWorkflow, error)
	CancelRuntimeWorkflow(ctx context.Context, runID string) (WorkflowCancelResult, error)
}

type RuntimeWorkflowService struct {
	repository RuntimeWorkflowRepository
	logger     *logrus.Logger
}

func NewRuntimeWorkflowService(repository RuntimeWorkflowRepository, loggers ...*logrus.Logger) *RuntimeWorkflowService {
	var logger *logrus.Logger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &RuntimeWorkflowService{repository: repository, logger: logger}
}

func (service *RuntimeWorkflowService) ListWorkflows(ctx context.Context) ([]IngestionWorkflow, error) {
	entry := service.entry().WithField("event", "taskboard.runtime_workflow.list")
	workflows, err := service.repository.ListRuntimeWorkflows(ctx)
	if err != nil {
		entry.WithError(err).Error("failed to list runtime workflows")
		return nil, fmt.Errorf("list runtime workflows: %w", err)
	}
	entry.WithField("workflow_count", len(workflows)).Info("listed runtime workflows")
	return workflows, nil
}

func (service *RuntimeWorkflowService) GetWorkflowStatus(ctx context.Context, runID string) (*IngestionWorkflow, error) {
	cleanRunID := strings.TrimSpace(runID)
	entry := service.entry().WithFields(logrus.Fields{"event": "taskboard.runtime_workflow.get", "run_id": cleanRunID})
	if cleanRunID == "" {
		entry.Error("run_id is required")
		return nil, fmt.Errorf("run_id is required")
	}
	workflow, err := service.repository.GetRuntimeWorkflow(ctx, cleanRunID)
	if err != nil {
		entry.WithError(err).Error("failed to load runtime workflow status")
		return nil, fmt.Errorf("load runtime workflow status: %w", err)
	}
	if workflow == nil {
		entry.Warn("workflow not found")
		return nil, fmt.Errorf("workflow not found: %s", cleanRunID)
	}
	entry.WithField("status", workflow.Status).Info("loaded runtime workflow status")
	return workflow, nil
}

func (service *RuntimeWorkflowService) CancelWorkflow(ctx context.Context, runID string) (WorkflowCancelResult, error) {
	cleanRunID := strings.TrimSpace(runID)
	entry := service.entry().WithFields(logrus.Fields{"event": "taskboard.runtime_workflow.cancel", "run_id": cleanRunID})
	if cleanRunID == "" {
		entry.Error("run_id is required")
		return WorkflowCancelResult{}, fmt.Errorf("run_id is required")
	}
	result, err := service.repository.CancelRuntimeWorkflow(ctx, cleanRunID)
	if err != nil {
		entry.WithError(err).Error("failed to cancel runtime workflow")
		return WorkflowCancelResult{}, fmt.Errorf("cancel runtime workflow: %w", err)
	}
	entry.WithFields(logrus.Fields{"matched_tasks": result.MatchedTasks, "canceled_tasks": result.CanceledTasks, "signaled_active": result.SignaledActive, "uncancelable_tasks": result.UncancelableTasks}).Info("canceled runtime workflow")
	return result, nil
}

func (service *RuntimeWorkflowService) entry() *logrus.Entry {
	if service == nil || service.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(service.logger)
}
