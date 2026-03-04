package tracker

import (
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type IngestionSourceRequest struct {
	RunID      string
	ProjectID  string
	WorkflowID string
	Source     domaintracker.SourceRef
}

func (request IngestionSourceRequest) Validate() error {
	if strings.TrimSpace(request.RunID) == "" {
		return failures.WrapTerminal(errors.New("run_id is required"))
	}
	if strings.TrimSpace(request.ProjectID) == "" {
		return failures.WrapTerminal(errors.New("project_id is required"))
	}
	if strings.TrimSpace(request.WorkflowID) == "" {
		return failures.WrapTerminal(errors.New("workflow_id is required"))
	}
	return request.Source.Validate()
}

type IngestionSource interface {
	SyncBoard(ctx context.Context, request IngestionSourceRequest) (domaintracker.Board, error)
}

type IngestionSourceResolver interface {
	Resolve(ctx context.Context, request IngestionSourceRequest) (IngestionSource, error)
}

type BoardStore interface {
	UpsertBoard(ctx context.Context, board domaintracker.Board) error
	ClaimNextTask(ctx context.Context, projectID string, boardID string, workerID string) (domaintracker.Board, domaintracker.Task, string, int64, error)
	ApplyTaskResult(ctx context.Context, projectID string, boardID string, claimID string, taskID string, nextStatus domaintracker.Status, outcome domaintracker.TaskOutcome) (domaintracker.Board, int64, error)
}

type ClaimNextTaskRequest struct {
	ProjectID string
	BoardID   string
	WorkerID  string
}

func (request ClaimNextTaskRequest) Validate() error {
	if strings.TrimSpace(request.ProjectID) == "" {
		return failures.WrapTerminal(errors.New("project_id is required"))
	}
	if strings.TrimSpace(request.BoardID) == "" {
		return failures.WrapTerminal(errors.New("board_id is required"))
	}
	if strings.TrimSpace(request.WorkerID) == "" {
		return failures.WrapTerminal(errors.New("worker_id is required"))
	}
	return nil
}

type ClaimedTask struct {
	Board    domaintracker.Board
	Task     domaintracker.Task
	ClaimID  string
	Revision int64
}

type ApplyTaskResultRequest struct {
	ProjectID       string
	BoardID         string
	ClaimID         string
	TaskID          string
	NextStatus      domaintracker.Status
	OutcomeStatus   string
	OutcomeReason   string
	TaskBranch      string
	Repository        string
	ResumeSessionID string
}

func (request ApplyTaskResultRequest) Validate() error {
	if strings.TrimSpace(request.ProjectID) == "" {
		return failures.WrapTerminal(errors.New("project_id is required"))
	}
	if strings.TrimSpace(request.BoardID) == "" {
		return failures.WrapTerminal(errors.New("board_id is required"))
	}
	if strings.TrimSpace(request.ClaimID) == "" {
		return failures.WrapTerminal(errors.New("claim_id is required"))
	}
	if strings.TrimSpace(request.TaskID) == "" {
		return failures.WrapTerminal(errors.New("task_id is required"))
	}
	if err := request.NextStatus.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(request.OutcomeStatus) == "" {
		return failures.WrapTerminal(errors.New("outcome_status is required"))
	}
	return nil
}

type AppliedTaskResult struct {
	Board    domaintracker.Board
	Revision int64
}

type SyncBoardRequest struct {
	RunID      string
	Prompt     string
	ProjectID  string
	WorkflowID string
	Source     domaintracker.SourceRef
}

func (request SyncBoardRequest) Validate() error {
	if strings.TrimSpace(request.Prompt) == "" {
		return failures.WrapTerminal(errors.New("prompt is required"))
	}
	return request.IngestionSourceRequest().Validate()
}

func (request SyncBoardRequest) IngestionSourceRequest() IngestionSourceRequest {
	return IngestionSourceRequest{
		RunID:      request.RunID,
		ProjectID:  request.ProjectID,
		WorkflowID: request.WorkflowID,
		Source:     request.Source,
	}
}

type Service struct {
	sourceResolver IngestionSourceResolver
	boardStore       BoardStore
}

func NewService(sourceResolver IngestionSourceResolver, boardStore BoardStore) (*Service, error) {
	if sourceResolver == nil {
		return nil, failures.WrapTerminal(errors.New("tracker ingestion source resolver is required"))
	}
	if boardStore == nil {
		return nil, failures.WrapTerminal(errors.New("tracker board store is required"))
	}
	return &Service{sourceResolver: sourceResolver, boardStore: boardStore}, nil
}

func NewTaskMutationService(boardStore BoardStore) (*Service, error) {
	if boardStore == nil {
		return nil, failures.WrapTerminal(errors.New("tracker board store is required"))
	}
	return &Service{boardStore: boardStore}, nil
}

func (service *Service) SyncBoard(ctx context.Context, request SyncBoardRequest) (domaintracker.Board, error) {
	if err := request.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	sourceRequest := request.IngestionSourceRequest()
	source, err := service.sourceResolver.Resolve(ctx, sourceRequest)
	if err != nil {
		return domaintracker.Board{}, ensureClassified(err)
	}
	if source == nil {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("tracker ingestion source resolver returned nil source"))
	}
	board, err := source.SyncBoard(ctx, sourceRequest)
	if err != nil {
		return domaintracker.Board{}, ensureClassified(err)
	}
	if err := board.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	if board.RunID != request.RunID {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("board run_id %q does not match request run_id %q", board.RunID, request.RunID))
	}
	if err := service.boardStore.UpsertBoard(ctx, board); err != nil {
		return domaintracker.Board{}, ensureClassified(err)
	}
	return board, nil
}

func (service *Service) ClaimNextTask(ctx context.Context, request ClaimNextTaskRequest) (ClaimedTask, error) {
	if err := request.Validate(); err != nil {
		return ClaimedTask{}, err
	}
	board, task, claimID, revision, err := service.boardStore.ClaimNextTask(
		ctx,
		strings.TrimSpace(request.ProjectID),
		strings.TrimSpace(request.BoardID),
		strings.TrimSpace(request.WorkerID),
	)
	if err != nil {
		return ClaimedTask{}, ensureClassified(err)
	}
	return ClaimedTask{
		Board:    board,
		Task:     task,
		ClaimID:  claimID,
		Revision: revision,
	}, nil
}

func (service *Service) ApplyTaskResult(ctx context.Context, request ApplyTaskResultRequest) (AppliedTaskResult, error) {
	if err := request.Validate(); err != nil {
		return AppliedTaskResult{}, err
	}
	outcome := domaintracker.TaskOutcome{
		Status:          strings.TrimSpace(request.OutcomeStatus),
		Reason:          strings.TrimSpace(request.OutcomeReason),
		TaskBranch:      strings.TrimSpace(request.TaskBranch),
		Repository:        strings.TrimSpace(request.Repository),
		ResumeSessionID: strings.TrimSpace(request.ResumeSessionID),
		UpdatedAt:       time.Now().UTC(),
	}
	if err := outcome.Validate(); err != nil {
		return AppliedTaskResult{}, err
	}
	board, revision, err := service.boardStore.ApplyTaskResult(
		ctx,
		strings.TrimSpace(request.ProjectID),
		strings.TrimSpace(request.BoardID),
		strings.TrimSpace(request.ClaimID),
		strings.TrimSpace(request.TaskID),
		request.NextStatus,
		outcome,
	)
	if err != nil {
		return AppliedTaskResult{}, ensureClassified(err)
	}
	return AppliedTaskResult{Board: board, Revision: revision}, nil
}

func ensureClassified(err error) error {
	if err == nil {
		return nil
	}
	if failures.ClassOf(err) != failures.ClassUnknown {
		return err
	}
	return failures.WrapTransient(err)
}
