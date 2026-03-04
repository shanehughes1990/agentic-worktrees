package worker

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	applicationingestion "agentic-orchestrator/internal/application/ingestion"
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type IngestionAgentHandler struct {
	service *applicationingestion.Service
}

func NewIngestionAgentHandler(service *applicationingestion.Service) (*IngestionAgentHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("ingestion service is required")
	}
	return &IngestionAgentHandler{service: service}, nil
}

func (handler *IngestionAgentHandler) Handle(ctx context.Context, job taskengine.Job) error {
	if handler == nil || handler.service == nil {
		return fmt.Errorf("ingestion agent handler is not initialized")
	}
	var payload applicationcontrolplane.IngestionAgentPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("decode ingestion agent payload: %w", err)
	}
	request := applicationingestion.Request{
		RunID:                     strings.TrimSpace(payload.RunID),
		JobID:                     strings.TrimSpace(payload.JobID),
		ProjectID:                 strings.TrimSpace(payload.ProjectID),
		BoardID:                   strings.TrimSpace(payload.BoardID),
		StreamID:                  strings.TrimSpace(payload.StreamID),
		SelectedDocumentLocations: payload.SelectedDocumentLocations,
		PreferSelectedDocuments:   payload.PreferSelectedDocuments,
		SourceRepositories:        mapSourceRepositories(payload.SourceRepositories),
		SourceBranch:              strings.TrimSpace(payload.SourceBranch),
		Model:                     strings.TrimSpace(payload.Model),
		SystemPrompt:              strings.TrimSpace(payload.SystemPrompt),
		UserPrompt:                strings.TrimSpace(payload.UserPrompt),
	}
	if _, err := handler.service.Execute(ctx, request); err != nil {
		return err
	}
	return nil
}

func mapSourceRepositories(payloadRepositories []applicationcontrolplane.IngestionSourceRepository) []applicationingestion.SourceRepository {
	mapped := make([]applicationingestion.SourceRepository, 0, len(payloadRepositories))
	for _, repository := range payloadRepositories {
		repositoryID := strings.TrimSpace(repository.RepositoryID)
		repositoryURL := strings.TrimSpace(repository.RepositoryURL)
		if repositoryID == "" || repositoryURL == "" {
			continue
		}
		mapped = append(mapped, applicationingestion.SourceRepository{RepositoryID: repositoryID, RepositoryURL: repositoryURL})
	}
	return mapped
}
