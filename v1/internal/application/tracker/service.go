package tracker

import (
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"errors"
	"fmt"
	"strings"
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

func ensureClassified(err error) error {
	if err == nil {
		return nil
	}
	if failures.ClassOf(err) != failures.ClassUnknown {
		return err
	}
	return failures.WrapTransient(err)
}
