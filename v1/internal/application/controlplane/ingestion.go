package controlplane

import (
	"agentic-orchestrator/internal/application/prompts"
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type RunIngestionAgentInput struct {
	ProjectID           string
	SelectedDocumentIDs []string
	UserPrompt          string
	Model               string
	SourceBranch        string
}

func (input RunIngestionAgentInput) Validate() error {
	if strings.TrimSpace(input.ProjectID) == "" {
		return fmt.Errorf("project_id is required")
	}
	if len(input.SelectedDocumentIDs) == 0 {
		return fmt.Errorf("selected_document_ids are required")
	}
	return nil
}

const taskboardIngestionSystemPrompt = "You are a taskboard synthesis agent. Build a canonical execution taskboard from the provided project documents. Extract epics and actionable tasks, preserve dependencies, avoid inventing unsupported facts, and output a deterministic plan suitable for execution orchestration."

const defaultIngestionModel = "gpt-5.3-codex"
const defaultIngestionSourceBranch = "main"

type IngestionAgentPayload struct {
	RunID                     string   `json:"run_id"`
	TaskID                    string   `json:"task_id"`
	JobID                     string   `json:"job_id"`
	ProjectID                 string   `json:"project_id"`
	SelectedDocumentLocations []string `json:"selected_document_locations"`
	SourceRepositories        []IngestionSourceRepository `json:"source_repositories,omitempty"`
	SourceBranch              string   `json:"source_branch"`
	Model                     string   `json:"model"`
	SystemPrompt              string   `json:"system_prompt"`
	UserPrompt                string   `json:"user_prompt,omitempty"`
	IdempotencyKey            string   `json:"idempotency_key"`
}

type IngestionSourceRepository struct {
	RepositoryID  string `json:"repository_id"`
	RepositoryURL string `json:"repository_url"`
}

type RunIngestionAgentResult struct {
	RunID       string
	TaskID      string
	JobID       string
	QueueTaskID string
	Duplicate   bool
}

func (service *Service) RunIngestionAgent(ctx context.Context, input RunIngestionAgentInput) (*RunIngestionAgentResult, error) {
	if service == nil || service.scheduler == nil {
		return nil, fmt.Errorf("task scheduler is not configured")
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	sourceRepositories := make([]IngestionSourceRepository, 0)
	if service.projectRepository != nil {
		setup, err := service.projectRepository.GetProjectSetup(ctx, strings.TrimSpace(input.ProjectID))
		if err != nil {
			return nil, fmt.Errorf("load project setup: %w", err)
		}
		if setup == nil {
			return nil, fmt.Errorf("project setup not found")
		}
		for _, repository := range setup.Repositories {
			repositoryID := strings.TrimSpace(repository.RepositoryID)
			repositoryURL := strings.TrimSpace(repository.RepositoryURL)
			if repositoryID == "" || repositoryURL == "" {
				continue
			}
			sourceRepositories = append(sourceRepositories, IngestionSourceRepository{RepositoryID: repositoryID, RepositoryURL: repositoryURL})
		}
	}
	normalizedDocumentLocations := make([]string, 0, len(input.SelectedDocumentIDs))
	for _, rawDocumentID := range input.SelectedDocumentIDs {
		documentID := strings.TrimSpace(rawDocumentID)
		if documentID == "" {
			continue
		}
		if service.projectDocumentRepository != nil {
			document, err := service.projectDocumentRepository.GetProjectDocument(ctx, strings.TrimSpace(input.ProjectID), documentID)
			if err != nil {
				return nil, fmt.Errorf("load project document %q: %w", documentID, err)
			}
			if document == nil {
				return nil, fmt.Errorf("selected project document %q not found", documentID)
			}
			objectPath := strings.TrimSpace(document.ObjectPath)
			if objectPath == "" {
				return nil, fmt.Errorf("selected project document %q has no object path", documentID)
			}
			normalizedDocumentLocations = append(normalizedDocumentLocations, objectPath)
			continue
		}
		normalizedDocumentLocations = append(normalizedDocumentLocations, documentID)
	}
	if len(normalizedDocumentLocations) == 0 {
		return nil, fmt.Errorf("selected_document_ids must contain at least one non-empty value")
	}
	runID := fmt.Sprintf("ingest-%d", time.Now().UTC().UnixNano())
	taskID := "ingestion"
	jobID := fmt.Sprintf("ingestion-agent-%d", time.Now().UTC().UnixNano())
	systemPrompt := strings.TrimSpace(prompts.WorkerToolingBaseline + "\n\n" + taskboardIngestionSystemPrompt)
	userPrompt := strings.TrimSpace(input.UserPrompt)
	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = defaultIngestionModel
	}
	sourceBranch := strings.TrimSpace(input.SourceBranch)
	if sourceBranch == "" {
		sourceBranch = defaultIngestionSourceBranch
	}
	idempotencyKey := ingestionIdempotencyKey(strings.TrimSpace(input.ProjectID), normalizedDocumentLocations, sourceRepositoryFingerprints(sourceRepositories), sourceBranch, model, strings.TrimSpace(systemPrompt), userPrompt)
	payload := IngestionAgentPayload{
		RunID:                     runID,
		TaskID:                    taskID,
		JobID:                     jobID,
		ProjectID:                 strings.TrimSpace(input.ProjectID),
		SelectedDocumentLocations: normalizedDocumentLocations,
		SourceRepositories:        sourceRepositories,
		SourceBranch:              sourceBranch,
		Model:                     model,
		SystemPrompt:              strings.TrimSpace(systemPrompt),
		UserPrompt:                userPrompt,
		IdempotencyKey:            idempotencyKey,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal ingestion payload: %w", err)
	}
	enqueueResult, err := service.scheduler.Enqueue(ctx, taskengine.EnqueueRequest{
		Kind:           taskengine.JobKindIngestionAgent,
		Payload:        payloadBytes,
		IdempotencyKey: idempotencyKey,
		CorrelationIDs: taskengine.CorrelationIDs{
			RunID:     runID,
			TaskID:    taskID,
			JobID:     jobID,
			ProjectID: strings.TrimSpace(input.ProjectID),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("enqueue ingestion agent job: %w", err)
	}
	return &RunIngestionAgentResult{RunID: runID, TaskID: taskID, JobID: jobID, QueueTaskID: enqueueResult.QueueTaskID, Duplicate: enqueueResult.Duplicate}, nil
}

func ingestionIdempotencyKey(projectID string, documentIDs []string, sourceRepositories []string, sourceBranch string, model string, systemPrompt string, userPrompt string) string {
	hasher := sha256.New()
	_, _ = hasher.Write([]byte(strings.TrimSpace(projectID)))
	for _, documentID := range documentIDs {
		_, _ = hasher.Write([]byte("|" + strings.TrimSpace(documentID)))
	}
	for _, repository := range sourceRepositories {
		_, _ = hasher.Write([]byte("|" + strings.TrimSpace(repository)))
	}
	_, _ = hasher.Write([]byte("|" + strings.TrimSpace(sourceBranch)))
	_, _ = hasher.Write([]byte("|" + strings.TrimSpace(model)))
	_, _ = hasher.Write([]byte("|" + strings.TrimSpace(systemPrompt)))
	_, _ = hasher.Write([]byte("|" + strings.TrimSpace(userPrompt)))
	return "ingestion-agent:" + hex.EncodeToString(hasher.Sum(nil))
}

func sourceRepositoryFingerprints(repositories []IngestionSourceRepository) []string {
	fingerprints := make([]string, 0, len(repositories))
	for _, repository := range repositories {
		repositoryID := strings.TrimSpace(repository.RepositoryID)
		repositoryURL := strings.TrimSpace(repository.RepositoryURL)
		if repositoryID == "" && repositoryURL == "" {
			continue
		}
		fingerprints = append(fingerprints, repositoryID+":"+repositoryURL)
	}
	return fingerprints
}
