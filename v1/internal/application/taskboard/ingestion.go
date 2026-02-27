package taskboard

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
	"github.com/sirupsen/logrus"
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
	dispatcher                     IngestionDispatcher
	repository                     Repository
	workflowRepo                   WorkflowRepository
	sourceLister                   domaintaskboard.SourceLister
	sourceReader                   domaintaskboard.SourceReader
	sourceWorkingDirectoryResolver domaintaskboard.SourceWorkingDirectoryResolver
	model                          string
	normalizers                    []DocumentNormalizer
	logger                         *logrus.Logger
}

func NewIngestionService(dispatcher IngestionDispatcher, repository Repository, workflowRepo WorkflowRepository, sourceLister domaintaskboard.SourceLister, sourceReader domaintaskboard.SourceReader, model string, loggers ...*logrus.Logger) *IngestionService {
	return NewIngestionServiceWithNormalizers(dispatcher, repository, workflowRepo, sourceLister, sourceReader, model, nil, loggers...)
}

func NewIngestionServiceWithNormalizers(dispatcher IngestionDispatcher, repository Repository, workflowRepo WorkflowRepository, sourceLister domaintaskboard.SourceLister, sourceReader domaintaskboard.SourceReader, model string, normalizers []DocumentNormalizer, loggers ...*logrus.Logger) *IngestionService {
	if len(normalizers) == 0 {
		normalizers = DefaultDocumentNormalizers()
	}
	var logger *logrus.Logger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	var sourceWorkingDirectoryResolver domaintaskboard.SourceWorkingDirectoryResolver
	if resolver, ok := sourceLister.(domaintaskboard.SourceWorkingDirectoryResolver); ok {
		sourceWorkingDirectoryResolver = resolver
	} else if resolver, ok := sourceReader.(domaintaskboard.SourceWorkingDirectoryResolver); ok {
		sourceWorkingDirectoryResolver = resolver
	}
	return &IngestionService{
		dispatcher:                     dispatcher,
		repository:                     repository,
		workflowRepo:                   workflowRepo,
		sourceLister:                   sourceLister,
		sourceReader:                   sourceReader,
		sourceWorkingDirectoryResolver: sourceWorkingDirectoryResolver,
		model:                          strings.TrimSpace(model),
		normalizers:                    normalizers,
		logger:                         logger,
	}
}

func (service *IngestionService) Ingest(ctx context.Context, request IngestRequest) (IngestionResult, error) {
	entry := service.entry().WithFields(logrus.Fields{
		"event":       "taskboard.ingestion.ingest",
		"source_path": strings.TrimSpace(request.SourcePath),
		"source_type": strings.TrimSpace(string(request.SourceType)),
	})
	if service.sourceLister == nil {
		entry.Error("source lister is required")
		return IngestionResult{}, fmt.Errorf("source lister is required")
	}
	if service.sourceReader == nil {
		entry.Error("source reader is required")
		return IngestionResult{}, fmt.Errorf("source reader is required")
	}

	cleanSourcePath := strings.TrimSpace(request.SourcePath)
	if cleanSourcePath == "" {
		entry.Error("source path is required")
		return IngestionResult{}, fmt.Errorf("source path is required")
	}

	sourceType := strings.TrimSpace(string(request.SourceType))
	if sourceType == "" {
		inferredSourceType, inferErr := service.inferSourceType(ctx, cleanSourcePath)
		if inferErr != nil {
			entry.WithError(inferErr).Error("failed to infer source type")
			return IngestionResult{}, inferErr
		}
		sourceType = string(inferredSourceType)
	}

	cleanSourceType := IngestionSourceType(sourceType)
	if cleanSourceType != IngestionSourceTypeFile && cleanSourceType != IngestionSourceTypeFolder {
		entry.WithField("source_type", cleanSourceType).Error("source type must be file or folder")
		return IngestionResult{}, fmt.Errorf("source type must be file or folder")
	}

	workingDirectory, err := service.resolveWorkingDirectory(ctx, cleanSourceType, cleanSourcePath)
	if err != nil {
		entry.WithError(err).Error("failed to resolve ingestion working directory")
		return IngestionResult{}, fmt.Errorf("resolve working directory: %w", err)
	}

	sourceKind := domaintaskboard.SourceKindFolder
	if cleanSourceType == IngestionSourceTypeFile {
		sourceKind = domaintaskboard.SourceKindFile
	}
	documents, err := NormalizeSourceDocuments(ctx, domaintaskboard.SourceMetadata{
		Identity: domaintaskboard.SourceIdentity{
			Kind:    sourceKind,
			Locator: cleanSourcePath,
		},
	}, request.Folder, service.sourceLister, service.sourceReader, service.normalizers)
	if err != nil {
		entry.WithError(err).Error("failed to normalize source documents")
		return IngestionResult{}, fmt.Errorf("normalize documents: %w", err)
	}
	entry.WithField("document_count", len(documents)).Info("normalized source documents")

	runID := uuid.NewString()
	job := IngestionJob{
		RunID:            runID,
		Prompt:           BuildTaskboardPrompt(cleanSourcePath, documents...),
		Model:            service.model,
		WorkingDirectory: workingDirectory,
	}
	taskID, err := service.dispatcher.EnqueueIngestion(ctx, job)
	if err != nil {
		entry.WithError(err).WithField("run_id", runID).Error("failed to enqueue ingestion")
		return IngestionResult{}, fmt.Errorf("enqueue ingestion: %w", err)
	}
	entry.WithFields(logrus.Fields{"run_id": runID, "queue_task_id": taskID, "working_directory": workingDirectory, "model": service.model}).Info("ingestion queued")

	workflow := &IngestionWorkflow{TaskID: taskID, Status: WorkflowStatusQueued, Message: "ingestion queued"}
	workflow.TaskType = WorkflowTaskTypeCopilotDecompose
	workflow.BoardID = runID
	workflow.Details = map[string]any{
		"run_id":            runID,
		"queue_task_id":     strings.TrimSpace(taskID),
		"source_path":       cleanSourcePath,
		"source_type":       string(cleanSourceType),
		"working_directory": workingDirectory,
		"model":             service.model,
	}
	workflow.Normalize(runID)
	if err := service.workflowRepo.SaveWorkflow(ctx, workflow); err != nil {
		entry.WithError(err).WithField("run_id", runID).Error("failed to save queued workflow")
		return IngestionResult{}, fmt.Errorf("save workflow queued status: %w", err)
	}

	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			entry.WithError(ctx.Err()).WithField("run_id", runID).Warn("ingestion polling canceled")
			return IngestionResult{}, ctx.Err()
		case <-ticker.C:
			currentWorkflow, err := service.workflowRepo.GetWorkflow(ctx, runID)
			if err != nil {
				entry.WithError(err).WithField("run_id", runID).Error("failed to load workflow while polling")
				return IngestionResult{}, fmt.Errorf("check workflow status: %w", err)
			}
			if currentWorkflow != nil && currentWorkflow.Status == WorkflowStatusFailed {
				message := strings.TrimSpace(currentWorkflow.Message)
				if message == "" {
					message = "ingestion workflow failed"
				}
				entry.WithField("run_id", runID).WithField("reason", message).Error("ingestion workflow failed")
				return IngestionResult{}, errors.New(message)
			}

			board, err := service.repository.GetByBoardID(ctx, runID)
			if err != nil {
				entry.WithError(err).WithField("run_id", runID).Error("failed to load taskboard while polling")
				return IngestionResult{}, fmt.Errorf("check taskboard result: %w", err)
			}
			if board != nil {
				entry.WithFields(logrus.Fields{"run_id": runID, "board_id": board.BoardID}).Info("ingestion completed with board result")
				return IngestionResult{RunID: runID, BoardID: board.BoardID}, nil
			}
		}
	}
}

func (service *IngestionService) resolveWorkingDirectory(ctx context.Context, sourceType IngestionSourceType, sourcePath string) (string, error) {
	if service.sourceWorkingDirectoryResolver != nil {
		sourceKind := domaintaskboard.SourceKindFolder
		if sourceType == IngestionSourceTypeFile {
			sourceKind = domaintaskboard.SourceKindFile
		}
		workingDirectory, err := service.sourceWorkingDirectoryResolver.ResolveWorkingDirectory(ctx, domaintaskboard.SourceIdentity{
			Kind:    sourceKind,
			Locator: sourcePath,
		})
		if err != nil {
			return "", err
		}
		cleanWorkingDirectory := strings.TrimSpace(workingDirectory)
		if cleanWorkingDirectory != "" {
			return cleanWorkingDirectory, nil
		}
	}
	if sourceType != IngestionSourceTypeFile {
		return sourcePath, nil
	}
	lastSeparator := strings.LastIndexAny(sourcePath, `/\`)
	if lastSeparator < 0 {
		return ".", nil
	}
	if lastSeparator == 0 {
		return sourcePath[:1], nil
	}
	return sourcePath[:lastSeparator], nil
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
	entry := service.entry().WithFields(logrus.Fields{"event": "taskboard.ingestion.get_workflow_status", "run_id": cleanRunID})
	if cleanRunID == "" {
		entry.Error("run_id is required")
		return nil, fmt.Errorf("run_id is required")
	}

	workflow, err := service.workflowRepo.GetWorkflow(ctx, cleanRunID)
	if err != nil {
		entry.WithError(err).Error("failed to load workflow status")
		return nil, fmt.Errorf("load workflow status: %w", err)
	}
	if workflow == nil {
		entry.Warn("workflow not found")
		return nil, fmt.Errorf("workflow not found: %s", cleanRunID)
	}
	entry.WithField("status", workflow.Status).Info("loaded workflow status")
	return workflow, nil
}

func (service *IngestionService) ListWorkflows(ctx context.Context) ([]IngestionWorkflow, error) {
	entry := service.entry().WithField("event", "taskboard.ingestion.list_workflows")
	workflows, err := service.workflowRepo.ListWorkflows(ctx)
	if err != nil {
		entry.WithError(err).Error("failed to list workflows")
		return nil, fmt.Errorf("list workflows: %w", err)
	}
	entry.WithField("workflow_count", len(workflows)).Info("listed workflows")
	return workflows, nil
}

func (service *IngestionService) entry() *logrus.Entry {
	if service == nil || service.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(service.logger)
}
