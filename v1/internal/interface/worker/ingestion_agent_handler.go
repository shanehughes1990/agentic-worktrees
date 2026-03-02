package worker

import (
	"agentic-orchestrator/internal/application/taskengine"
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type IngestionBoardSourcePayload struct {
	BoardID                  string         `json:"board_id"`
	Kind                     string         `json:"kind"`
	Location                 string         `json:"location,omitempty"`
	ExternalBoardID          string         `json:"external_board_id,omitempty"`
	AppliesToAllRepositories bool           `json:"applies_to_all_repositories"`
	RepositoryIDs            []string       `json:"repository_ids,omitempty"`
	Config                   map[string]any `json:"config,omitempty"`
}

type IngestionAgentPayload struct {
	RunID          string                        `json:"run_id"`
	TaskID         string                        `json:"task_id"`
	JobID          string                        `json:"job_id"`
	IdempotencyKey string                        `json:"idempotency_key"`
	Prompt         string                        `json:"prompt"`
	ProjectID      string                        `json:"project_id"`
	WorkflowID     string                        `json:"workflow_id"`
	BoardSources   []IngestionBoardSourcePayload `json:"board_sources"`
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
	if len(payload.BoardSources) == 0 {
		return fmt.Errorf("ingestion board_sources is required")
	}
	correlation := taskengine.CorrelationIDs{RunID: payload.RunID, TaskID: payload.TaskID, JobID: payload.JobID, ProjectID: payload.ProjectID}
	for _, boardSource := range payload.BoardSources {
		sourceKind := domaintracker.SourceKind(strings.TrimSpace(boardSource.Kind))
		sourceLocation := strings.TrimSpace(boardSource.Location)
		if sourceKind == domaintracker.SourceKindLocalJSON {
			sourceLocation = scopedLocalTrackerLocation(payload.ProjectID, sourceLocation)
		}
		request := applicationtracker.SyncBoardRequest{
			RunID:      strings.TrimSpace(payload.RunID),
			Prompt:     strings.TrimSpace(payload.Prompt),
			ProjectID:  strings.TrimSpace(payload.ProjectID),
			WorkflowID: strings.TrimSpace(payload.WorkflowID),
			Source: domaintracker.SourceRef{
				Kind:     sourceKind,
				Location: sourceLocation,
				BoardID:  strings.TrimSpace(boardSource.ExternalBoardID),
				Config:   boardSource.Config,
			},
		}
		board, err := handler.service.SyncBoard(ctx, request)
		if err != nil {
			handler.safeSupervisorAttention(ctx, correlation, err.Error())
			return err
		}
		if sourceKind == domaintracker.SourceKindGitHubIssues {
			references := issueReferencesFromBoard(board)
			if len(references) == 0 {
				handler.safeSupervisorIssueOpened(ctx, correlation, sourceLocation, sourceLocation)
				continue
			}
			for _, reference := range references {
				handler.safeSupervisorIssueOpened(ctx, correlation, sourceLocation, reference)
			}
		}
	}
	return nil
}

func scopedLocalTrackerLocation(projectID string, location string) string {
	trimmedLocation := strings.TrimSpace(location)
	if trimmedLocation == "" || filepath.IsAbs(trimmedLocation) {
		return trimmedLocation
	}
	cleanLocation := filepath.Clean(trimmedLocation)
	trimmedProjectID := strings.TrimSpace(projectID)
	if trimmedProjectID == "" {
		return cleanLocation
	}
	projectRoot := filepath.Join(trimmedProjectID, "tracker")
	if cleanLocation == projectRoot || strings.HasPrefix(cleanLocation, projectRoot+string(filepath.Separator)) {
		return cleanLocation
	}
	return filepath.Join(projectRoot, cleanLocation)
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
