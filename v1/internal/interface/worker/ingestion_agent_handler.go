package worker

import (
	"agentic-orchestrator/internal/application/taskengine"
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type IngestionBoardSourcePayload struct {
	Kind     string         `json:"kind"`
	Location string         `json:"location,omitempty"`
	BoardID  string         `json:"board_id,omitempty"`
	Config   map[string]any `json:"config,omitempty"`
}

type IngestionAgentPayload struct {
	RunID          string                      `json:"run_id"`
	TaskID         string                      `json:"task_id"`
	JobID          string                      `json:"job_id"`
	IdempotencyKey string                      `json:"idempotency_key"`
	Prompt         string                      `json:"prompt"`
	ProjectID      string                      `json:"project_id"`
	WorkflowID     string                      `json:"workflow_id"`
	BoardSource    IngestionBoardSourcePayload `json:"board_source"`
}

type trackerService interface {
	SyncBoard(ctx context.Context, request applicationtracker.SyncBoardRequest) (domaintracker.Board, error)
}

type IngestionAgentHandler struct {
	service trackerService
}

func NewIngestionAgentHandler(service trackerService) (*IngestionAgentHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("tracker service is required")
	}
	return &IngestionAgentHandler{service: service}, nil
}

func (handler *IngestionAgentHandler) Handle(ctx context.Context, job taskengine.Job) error {
	var payload IngestionAgentPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("decode ingestion agent payload: %w", err)
	}
	request := applicationtracker.SyncBoardRequest{
		RunID:      strings.TrimSpace(payload.RunID),
		Prompt:     strings.TrimSpace(payload.Prompt),
		ProjectID:  strings.TrimSpace(payload.ProjectID),
		WorkflowID: strings.TrimSpace(payload.WorkflowID),
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKind(strings.TrimSpace(payload.BoardSource.Kind)),
			Location: strings.TrimSpace(payload.BoardSource.Location),
			BoardID:  strings.TrimSpace(payload.BoardSource.BoardID),
			Config:   payload.BoardSource.Config,
		},
	}
	_, err := handler.service.SyncBoard(ctx, request)
	return err
}
