package taskboard

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type IngestionJob struct {
	RunID            string
	Prompt           string
	Model            string
	WorkingDirectory string
}

type IngestionResult struct {
	RunID   string
	BoardID string
}

type IngestionDispatcher interface {
	EnqueueIngestion(ctx context.Context, job IngestionJob) (string, error)
}

type IngestionService struct {
	dispatcher   IngestionDispatcher
	repository   Repository
	workflowRepo WorkflowRepository
	model        string
}

func NewIngestionService(dispatcher IngestionDispatcher, repository Repository, workflowRepo WorkflowRepository, model string) *IngestionService {
	return &IngestionService{
		dispatcher:   dispatcher,
		repository:   repository,
		workflowRepo: workflowRepo,
		model:        strings.TrimSpace(model),
	}
}

func (service *IngestionService) IngestDirectory(ctx context.Context, directory string) (IngestionResult, error) {
	cleanDirectory := strings.TrimSpace(directory)
	if cleanDirectory == "" {
		return IngestionResult{}, fmt.Errorf("directory is required")
	}

	runID := uuid.NewString()
	job := IngestionJob{
		RunID:            runID,
		Prompt:           BuildTaskboardPrompt(cleanDirectory),
		Model:            service.model,
		WorkingDirectory: cleanDirectory,
	}
	taskID, err := service.dispatcher.EnqueueIngestion(ctx, job)
	if err != nil {
		return IngestionResult{}, fmt.Errorf("enqueue ingestion: %w", err)
	}

	workflow := &IngestionWorkflow{TaskID: taskID, Status: WorkflowStatusQueued, Message: "ingestion queued"}
	workflow.Normalize(runID)
	if err := service.workflowRepo.SaveWorkflow(ctx, workflow); err != nil {
		return IngestionResult{}, fmt.Errorf("save workflow queued status: %w", err)
	}

	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return IngestionResult{}, ctx.Err()
		case <-ticker.C:
			currentWorkflow, err := service.workflowRepo.GetWorkflow(ctx, runID)
			if err != nil {
				return IngestionResult{}, fmt.Errorf("check workflow status: %w", err)
			}
			if currentWorkflow != nil && currentWorkflow.Status == WorkflowStatusFailed {
				message := strings.TrimSpace(currentWorkflow.Message)
				if message == "" {
					message = "ingestion workflow failed"
				}
				return IngestionResult{}, errors.New(message)
			}

			board, err := service.repository.GetByBoardID(ctx, runID)
			if err != nil {
				return IngestionResult{}, fmt.Errorf("check taskboard result: %w", err)
			}
			if board != nil {
				return IngestionResult{RunID: runID, BoardID: board.BoardID}, nil
			}
		}
	}
}

func (service *IngestionService) GetWorkflowStatus(ctx context.Context, runID string) (*IngestionWorkflow, error) {
	cleanRunID := strings.TrimSpace(runID)
	if cleanRunID == "" {
		return nil, fmt.Errorf("run_id is required")
	}

	workflow, err := service.workflowRepo.GetWorkflow(ctx, cleanRunID)
	if err != nil {
		return nil, fmt.Errorf("load workflow status: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", cleanRunID)
	}
	return workflow, nil
}

func (service *IngestionService) ListWorkflows(ctx context.Context) ([]IngestionWorkflow, error) {
	workflows, err := service.workflowRepo.ListWorkflows(ctx)
	if err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}
	return workflows, nil
}
