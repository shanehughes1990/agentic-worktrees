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
	service           trackerService
	supervisorService supervisorSignalService
}

func NewIngestionAgentHandler(service trackerService) (*IngestionAgentHandler, error) {
	return NewIngestionAgentHandlerWithSupervisor(service, nil)
}

func NewIngestionAgentHandlerWithSupervisor(service trackerService, supervisorService supervisorSignalService) (*IngestionAgentHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("tracker service is required")
	}
	return &IngestionAgentHandler{service: service, supervisorService: supervisorService}, nil
}

func (handler *IngestionAgentHandler) Handle(ctx context.Context, job taskengine.Job) error {
	var payload IngestionAgentPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("decode ingestion agent payload: %w", err)
	}
	sourceKind := domaintracker.SourceKind(strings.TrimSpace(payload.BoardSource.Kind))
	sourceLocation := strings.TrimSpace(payload.BoardSource.Location)
	request := applicationtracker.SyncBoardRequest{
		RunID:      strings.TrimSpace(payload.RunID),
		Prompt:     strings.TrimSpace(payload.Prompt),
		ProjectID:  strings.TrimSpace(payload.ProjectID),
		WorkflowID: strings.TrimSpace(payload.WorkflowID),
		Source: domaintracker.SourceRef{
			Kind:     sourceKind,
			Location: sourceLocation,
			BoardID:  strings.TrimSpace(payload.BoardSource.BoardID),
			Config:   payload.BoardSource.Config,
		},
	}
	board, err := handler.service.SyncBoard(ctx, request)
	correlation := taskengine.CorrelationIDs{RunID: payload.RunID, TaskID: payload.TaskID, JobID: payload.JobID}
	if err != nil {
		handler.safeSupervisorAttention(ctx, correlation, err.Error())
		return err
	}
	if sourceKind == domaintracker.SourceKindGitHubIssues {
		references := issueReferencesFromBoard(board)
		if len(references) == 0 {
			handler.safeSupervisorIssueOpened(ctx, correlation, sourceLocation, sourceLocation)
			return nil
		}
		for _, reference := range references {
			handler.safeSupervisorIssueOpened(ctx, correlation, sourceLocation, reference)
		}
	}
	return nil
}

func issueReferencesFromBoard(board domaintracker.Board) []string {
	references := make([]string, 0)
	for _, epic := range board.Epics {
		for _, task := range epic.Tasks {
			if task.Metadata == nil {
				continue
			}
			reference, ok := task.Metadata["issue_reference"].(string)
			if !ok {
				continue
			}
			cleanReference := strings.TrimSpace(reference)
			if cleanReference == "" {
				continue
			}
			references = append(references, cleanReference)
		}
	}
	return references
}

func (handler *IngestionAgentHandler) safeSupervisorAttention(ctx context.Context, correlation taskengine.CorrelationIDs, reason string) {
	if handler == nil || handler.supervisorService == nil {
		return
	}
	_, _ = handler.supervisorService.OnTrackerAttention(ctx, correlation, reason)
}

func (handler *IngestionAgentHandler) safeSupervisorIssueOpened(ctx context.Context, correlation taskengine.CorrelationIDs, source string, issueReference string) {
	if handler == nil || handler.supervisorService == nil {
		return
	}
	_, _ = handler.supervisorService.OnIssueOpened(ctx, correlation, strings.TrimSpace(source), strings.TrimSpace(issueReference))
}
