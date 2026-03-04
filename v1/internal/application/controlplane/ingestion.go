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
	BoardID             string
	SelectedDocumentIDs []string
	RepositorySourceBranches []RepositorySourceBranch
	UserPrompt          string
	Model               string
	SourceBranch        string
}

type RepositorySourceBranch struct {
	RepositoryID string
	Branch       string
}

func (input RunIngestionAgentInput) Validate() error {
	if strings.TrimSpace(input.ProjectID) == "" {
		return fmt.Errorf("project_id is required")
	}
	hasPrompt := strings.TrimSpace(input.UserPrompt) != ""
	hasSelectedDocuments := hasNonEmptyValues(input.SelectedDocumentIDs)
	if !hasPrompt && !hasSelectedDocuments {
		return fmt.Errorf("user_prompt or selected_document_ids is required")
	}
	for index, repositorySourceBranch := range input.RepositorySourceBranches {
		if strings.TrimSpace(repositorySourceBranch.RepositoryID) == "" {
			return fmt.Errorf("repository_source_branches[%d].repository_id is required", index)
		}
		if strings.TrimSpace(repositorySourceBranch.Branch) == "" {
			return fmt.Errorf("repository_source_branches[%d].branch is required", index)
		}
	}
	return nil
}

const defaultIngestionModel = "gpt-5.3-codex"
const defaultIngestionSourceBranch = "main"

type IngestionAgentPayload struct {
	RunID                     string   `json:"run_id"`
	TaskID                    string   `json:"task_id"`
	JobID                     string   `json:"job_id"`
	BoardID                   string   `json:"board_id"`
	StreamID                  string   `json:"stream_id"`
	ProjectID                 string   `json:"project_id"`
	SelectedDocumentLocations []string `json:"selected_document_locations"`
	PreferSelectedDocuments   bool     `json:"prefer_selected_documents"`
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
	SourceBranch  string `json:"source_branch,omitempty"`
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
	selectedRepositoryBranches := mapRepositorySourceBranches(input.RepositorySourceBranches)
	boardID := strings.TrimSpace(input.BoardID)
	if service.projectRepository != nil {
		setup, err := service.projectRepository.GetProjectSetup(ctx, strings.TrimSpace(input.ProjectID))
		if err != nil {
			return nil, fmt.Errorf("load project setup: %w", err)
		}
		if setup == nil {
			return nil, fmt.Errorf("project setup not found")
		}
		knownRepositoryIDs := make(map[string]struct{}, len(setup.Repositories))
		for _, repository := range setup.Repositories {
			repositoryID := strings.TrimSpace(repository.RepositoryID)
			repositoryURL := strings.TrimSpace(repository.RepositoryURL)
			if repositoryID == "" || repositoryURL == "" {
				continue
			}
			knownRepositoryIDs[repositoryID] = struct{}{}
			resolvedSourceBranch := strings.TrimSpace(selectedRepositoryBranches[repositoryID])
			if resolvedSourceBranch == "" {
				resolvedSourceBranch = strings.TrimSpace(input.SourceBranch)
			}
			if resolvedSourceBranch == "" {
				resolvedSourceBranch = defaultIngestionSourceBranch
			}
			sourceRepositories = append(sourceRepositories, IngestionSourceRepository{RepositoryID: repositoryID, RepositoryURL: repositoryURL, SourceBranch: resolvedSourceBranch})
		}
		if boardID == "" && len(setup.Boards) > 0 {
			boardID = strings.TrimSpace(setup.Boards[0].BoardID)
		}
		for selectedRepositoryID := range selectedRepositoryBranches {
			if _, exists := knownRepositoryIDs[selectedRepositoryID]; !exists {
				return nil, fmt.Errorf("repository_source_branches contains unknown repository_id %q", selectedRepositoryID)
			}
		}
	}
	if boardID == "" {
		boardID = boardIDFromName("default")
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
	preferSelectedDocuments := len(normalizedDocumentLocations) > 0
	runID := fmt.Sprintf("ingest-%d", time.Now().UTC().UnixNano())
	taskID := "ingestion"
	jobID := fmt.Sprintf("ingestion-agent-%d", time.Now().UTC().UnixNano())
	systemPrompt := strings.TrimSpace(prompts.WorkerToolingBaseline)
	userPrompt := strings.TrimSpace(input.UserPrompt)
	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = defaultIngestionModel
	}
	sourceBranch := strings.TrimSpace(input.SourceBranch)
	if sourceBranch == "" {
		sourceBranch = defaultIngestionSourceBranch
	}
	idempotencyKey := ingestionIdempotencyKey(strings.TrimSpace(input.ProjectID), boardID, normalizedDocumentLocations, sourceRepositoryFingerprints(sourceRepositories), sourceBranch, model, strings.TrimSpace(systemPrompt), userPrompt)
	payload := IngestionAgentPayload{
		RunID:                     runID,
		TaskID:                    taskID,
		JobID:                     jobID,
		BoardID:                   boardID,
		StreamID:                  jobID,
		ProjectID:                 strings.TrimSpace(input.ProjectID),
		SelectedDocumentLocations: normalizedDocumentLocations,
		PreferSelectedDocuments:   preferSelectedDocuments,
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

func ingestionIdempotencyKey(projectID string, boardID string, documentIDs []string, sourceRepositories []string, sourceBranch string, model string, systemPrompt string, userPrompt string) string {
	hasher := sha256.New()
	_, _ = hasher.Write([]byte(strings.TrimSpace(projectID)))
	_, _ = hasher.Write([]byte("|" + strings.TrimSpace(boardID)))
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
		sourceBranch := strings.TrimSpace(repository.SourceBranch)
		if repositoryID == "" && repositoryURL == "" {
			continue
		}
		fingerprints = append(fingerprints, repositoryID+":"+repositoryURL+":"+sourceBranch)
	}
	return fingerprints
}

func mapRepositorySourceBranches(values []RepositorySourceBranch) map[string]string {
	result := make(map[string]string, len(values))
	for _, value := range values {
		repositoryID := strings.TrimSpace(value.RepositoryID)
		branch := strings.TrimSpace(value.Branch)
		if repositoryID == "" || branch == "" {
			continue
		}
		result[repositoryID] = branch
	}
	return result
}

func hasNonEmptyValues(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}
