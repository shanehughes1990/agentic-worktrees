package taskboard

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
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

type IngestionSourceType string

const (
	IngestionSourceTypeFile   IngestionSourceType = "file"
	IngestionSourceTypeFolder IngestionSourceType = "folder"
)

type FolderTraversalOptions struct {
	WalkDepth        int
	IgnorePaths      []string
	IgnoreExtensions []string
}

type IngestRequest struct {
	SourcePath string
	SourceType IngestionSourceType
	Folder     FolderTraversalOptions
}

type IngestionDispatcher interface {
	EnqueueIngestion(ctx context.Context, job IngestionJob) (string, error)
}

type IngestionService struct {
	dispatcher   IngestionDispatcher
	repository   Repository
	workflowRepo WorkflowRepository
	sourceLister domaintaskboard.SourceLister
	sourceReader domaintaskboard.SourceReader
	model        string
	normalizers  []DocumentNormalizer
}

func NewIngestionService(dispatcher IngestionDispatcher, repository Repository, workflowRepo WorkflowRepository, sourceLister domaintaskboard.SourceLister, sourceReader domaintaskboard.SourceReader, model string) *IngestionService {
	return NewIngestionServiceWithNormalizers(dispatcher, repository, workflowRepo, sourceLister, sourceReader, model, nil)
}

func NewIngestionServiceWithNormalizers(dispatcher IngestionDispatcher, repository Repository, workflowRepo WorkflowRepository, sourceLister domaintaskboard.SourceLister, sourceReader domaintaskboard.SourceReader, model string, normalizers []DocumentNormalizer) *IngestionService {
	if len(normalizers) == 0 {
		normalizers = DefaultDocumentNormalizers()
	}
	return &IngestionService{
		dispatcher:   dispatcher,
		repository:   repository,
		workflowRepo: workflowRepo,
		sourceLister: sourceLister,
		sourceReader: sourceReader,
		model:        strings.TrimSpace(model),
		normalizers:  normalizers,
	}
}

func (service *IngestionService) Ingest(ctx context.Context, request IngestRequest) (IngestionResult, error) {
	if service.sourceLister == nil {
		return IngestionResult{}, fmt.Errorf("source lister is required")
	}
	if service.sourceReader == nil {
		return IngestionResult{}, fmt.Errorf("source reader is required")
	}

	cleanSourcePath := strings.TrimSpace(request.SourcePath)
	if cleanSourcePath == "" {
		return IngestionResult{}, fmt.Errorf("source path is required")
	}

	sourceType := strings.TrimSpace(string(request.SourceType))
	if sourceType == "" {
		inferredSourceType, inferErr := service.inferSourceType(ctx, cleanSourcePath)
		if inferErr != nil {
			return IngestionResult{}, inferErr
		}
		sourceType = string(inferredSourceType)
	}

	cleanSourceType := IngestionSourceType(sourceType)
	if cleanSourceType != IngestionSourceTypeFile && cleanSourceType != IngestionSourceTypeFolder {
		return IngestionResult{}, fmt.Errorf("source type must be file or folder")
	}

	workingDirectory := cleanSourcePath
	if cleanSourceType == IngestionSourceTypeFile {
		workingDirectory = filepath.Dir(cleanSourcePath)
	}

	documents, err := NormalizeSourceDocumentsWithSourcePort(ctx, cleanSourcePath, cleanSourceType, request.Folder, service.sourceLister, service.sourceReader, service.normalizers)
	if err != nil {
		return IngestionResult{}, fmt.Errorf("normalize documents: %w", err)
	}

	runID := uuid.NewString()
	job := IngestionJob{
		RunID:            runID,
		Prompt:           BuildTaskboardPrompt(cleanSourcePath, documents...),
		Model:            service.model,
		WorkingDirectory: workingDirectory,
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

func (service *IngestionService) inferSourceType(ctx context.Context, sourcePath string) (IngestionSourceType, error) {
	_, folderErr := service.sourceLister.List(ctx, domaintaskboard.SourceMetadata{
		Identity: domaintaskboard.SourceIdentity{
			Kind:    domaintaskboard.SourceKindFolder,
			Locator: sourcePath,
		},
	}, domaintaskboard.SourceListOptions{WalkDepth: 0})
	if folderErr == nil {
		return IngestionSourceTypeFolder, nil
	}

	_, fileErr := service.sourceReader.Read(ctx, domaintaskboard.SourceIdentity{
		Kind:    domaintaskboard.SourceKindFile,
		Locator: sourcePath,
	})
	if fileErr == nil {
		return IngestionSourceTypeFile, nil
	}

	return "", fmt.Errorf("determine source type: %w", errors.Join(folderErr, fileErr))
}

func (service *IngestionService) IngestDirectory(ctx context.Context, directory string) (IngestionResult, error) {
	return service.Ingest(ctx, IngestRequest{
		SourcePath: strings.TrimSpace(directory),
		SourceType: IngestionSourceTypeFolder,
		Folder: FolderTraversalOptions{
			WalkDepth: -1,
		},
	})
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
