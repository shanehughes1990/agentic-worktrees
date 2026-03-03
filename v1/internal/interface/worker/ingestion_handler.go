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
		ProjectID:                 strings.TrimSpace(payload.ProjectID),
		SelectedDocumentLocations: payload.SelectedDocumentLocations,
		SystemPrompt:              strings.TrimSpace(payload.SystemPrompt),
		UserPrompt:                strings.TrimSpace(payload.UserPrompt),
	}
	if _, err := handler.service.Execute(ctx, request); err != nil {
		return err
	}
	return nil
}
