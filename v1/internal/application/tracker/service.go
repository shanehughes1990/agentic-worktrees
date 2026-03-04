package tracker

import (
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"errors"
	"strings"
	"time"
)

type BoardStore interface {
	UpsertBoard(ctx context.Context, board domaintracker.Board) error
	ListBoards(ctx context.Context, projectID string) ([]domaintracker.Board, error)
	LoadBoard(ctx context.Context, projectID string, boardID string) (domaintracker.Board, error)
	DeleteBoard(ctx context.Context, projectID string, boardID string) error
	ClaimNextTask(ctx context.Context, projectID string, boardID string, agentID string, leaseTTL time.Duration) (domaintracker.Board, domaintracker.Task, string, error)
	ApplyTaskResult(ctx context.Context, projectID string, boardID string, claimToken string, taskID string, nextState domaintracker.TaskState, outcome domaintracker.TaskOutcome) (domaintracker.Board, error)
}

type ClaimNextTaskRequest struct {
	ProjectID string
	BoardID   string
	AgentID   string
	LeaseTTL  time.Duration
}

func (request ClaimNextTaskRequest) Validate() error {
	if strings.TrimSpace(request.ProjectID) == "" {
		return failures.WrapTerminal(errors.New("project_id is required"))
	}
	if strings.TrimSpace(request.BoardID) == "" {
		return failures.WrapTerminal(errors.New("board_id is required"))
	}
	if strings.TrimSpace(request.AgentID) == "" {
		return failures.WrapTerminal(errors.New("agent_id is required"))
	}
	if request.LeaseTTL <= 0 {
		return failures.WrapTerminal(errors.New("lease_ttl must be greater than zero"))
	}
	return nil
}

type ClaimedTask struct {
	Board      domaintracker.Board
	Task       domaintracker.Task
	ClaimToken string
}

type ApplyTaskResultRequest struct {
	ProjectID       string
	BoardID         string
	ClaimToken      string
	TaskID          string
	NextState       domaintracker.TaskState
	OutcomeStatus   domaintracker.OutcomeStatus
	OutcomeSummary  string
	OutcomeErrorCode string
	OutcomeErrorMessage string
}

func (request ApplyTaskResultRequest) Validate() error {
	if strings.TrimSpace(request.ProjectID) == "" {
		return failures.WrapTerminal(errors.New("project_id is required"))
	}
	if strings.TrimSpace(request.BoardID) == "" {
		return failures.WrapTerminal(errors.New("board_id is required"))
	}
	if strings.TrimSpace(request.ClaimToken) == "" {
		return failures.WrapTerminal(errors.New("claim_token is required"))
	}
	if strings.TrimSpace(request.TaskID) == "" {
		return failures.WrapTerminal(errors.New("task_id is required"))
	}
	if err := request.NextState.Validate(); err != nil {
		return err
	}
	outcome := domaintracker.TaskOutcome{
		Status:       request.OutcomeStatus,
		Summary:      strings.TrimSpace(request.OutcomeSummary),
		ErrorCode:    strings.TrimSpace(request.OutcomeErrorCode),
		ErrorMessage: strings.TrimSpace(request.OutcomeErrorMessage),
	}
	if err := outcome.Validate(); err != nil {
		return err
	}
	return nil
}

type AppliedTaskResult struct {
	Board domaintracker.Board
}

type Service struct {
	boardStore BoardStore
}

func NewTaskMutationService(boardStore BoardStore) (*Service, error) {
	if boardStore == nil {
		return nil, failures.WrapTerminal(errors.New("tracker board store is required"))
	}
	return &Service{boardStore: boardStore}, nil
}

func (service *Service) UpsertBoard(ctx context.Context, board domaintracker.Board) error {
	if service == nil || service.boardStore == nil {
		return failures.WrapTerminal(errors.New("tracker service is not initialized"))
	}
	if err := board.Validate(); err != nil {
		return err
	}
	if err := service.boardStore.UpsertBoard(ctx, board); err != nil {
		return ensureClassified(err)
	}
	return nil
}

func (service *Service) ListBoards(ctx context.Context, projectID string) ([]domaintracker.Board, error) {
	if service == nil || service.boardStore == nil {
		return nil, failures.WrapTerminal(errors.New("tracker service is not initialized"))
	}
	cleanProjectID := strings.TrimSpace(projectID)
	if cleanProjectID == "" {
		return nil, failures.WrapTerminal(errors.New("project_id is required"))
	}
	boards, err := service.boardStore.ListBoards(ctx, cleanProjectID)
	if err != nil {
		return nil, ensureClassified(err)
	}
	return boards, nil
}

func (service *Service) LoadBoard(ctx context.Context, projectID string, boardID string) (domaintracker.Board, error) {
	if service == nil || service.boardStore == nil {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("tracker service is not initialized"))
	}
	cleanProjectID := strings.TrimSpace(projectID)
	if cleanProjectID == "" {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("project_id is required"))
	}
	cleanBoardID := strings.TrimSpace(boardID)
	if cleanBoardID == "" {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("board_id is required"))
	}
	board, err := service.boardStore.LoadBoard(ctx, cleanProjectID, cleanBoardID)
	if err != nil {
		return domaintracker.Board{}, ensureClassified(err)
	}
	return board, nil
}

func (service *Service) DeleteBoard(ctx context.Context, projectID string, boardID string) error {
	if service == nil || service.boardStore == nil {
		return failures.WrapTerminal(errors.New("tracker service is not initialized"))
	}
	cleanProjectID := strings.TrimSpace(projectID)
	if cleanProjectID == "" {
		return failures.WrapTerminal(errors.New("project_id is required"))
	}
	cleanBoardID := strings.TrimSpace(boardID)
	if cleanBoardID == "" {
		return failures.WrapTerminal(errors.New("board_id is required"))
	}
	if err := service.boardStore.DeleteBoard(ctx, cleanProjectID, cleanBoardID); err != nil {
		return ensureClassified(err)
	}
	return nil
}

func (service *Service) ClaimNextTask(ctx context.Context, request ClaimNextTaskRequest) (ClaimedTask, error) {
	if service == nil || service.boardStore == nil {
		return ClaimedTask{}, failures.WrapTerminal(errors.New("tracker service is not initialized"))
	}
	if err := request.Validate(); err != nil {
		return ClaimedTask{}, err
	}
	board, task, claimToken, err := service.boardStore.ClaimNextTask(
		ctx,
		strings.TrimSpace(request.ProjectID),
		strings.TrimSpace(request.BoardID),
		strings.TrimSpace(request.AgentID),
		request.LeaseTTL,
	)
	if err != nil {
		return ClaimedTask{}, ensureClassified(err)
	}
	return ClaimedTask{Board: board, Task: task, ClaimToken: claimToken}, nil
}

func (service *Service) ApplyTaskResult(ctx context.Context, request ApplyTaskResultRequest) (AppliedTaskResult, error) {
	if service == nil || service.boardStore == nil {
		return AppliedTaskResult{}, failures.WrapTerminal(errors.New("tracker service is not initialized"))
	}
	if err := request.Validate(); err != nil {
		return AppliedTaskResult{}, err
	}
	outcome := domaintracker.TaskOutcome{
		Status:       request.OutcomeStatus,
		Summary:      strings.TrimSpace(request.OutcomeSummary),
		ErrorCode:    strings.TrimSpace(request.OutcomeErrorCode),
		ErrorMessage: strings.TrimSpace(request.OutcomeErrorMessage),
	}
	board, err := service.boardStore.ApplyTaskResult(
		ctx,
		strings.TrimSpace(request.ProjectID),
		strings.TrimSpace(request.BoardID),
		strings.TrimSpace(request.ClaimToken),
		strings.TrimSpace(request.TaskID),
		request.NextState,
		outcome,
	)
	if err != nil {
		return AppliedTaskResult{}, ensureClassified(err)
	}
	return AppliedTaskResult{Board: board}, nil
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
