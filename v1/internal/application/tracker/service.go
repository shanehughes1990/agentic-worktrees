package tracker

import (
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"errors"
	"fmt"
	"strings"
)

type ProviderSyncRequest struct {
	RunID      string
	ProjectID  string
	WorkflowID string
	Source     domaintracker.SourceRef
}

func (request ProviderSyncRequest) Validate() error {
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

type Provider interface {
	SyncBoard(ctx context.Context, request ProviderSyncRequest) (domaintracker.Board, error)
}

type ProviderResolver interface {
	Resolve(ctx context.Context, request ProviderSyncRequest) (Provider, error)
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
	return request.ProviderSyncRequest().Validate()
}

func (request SyncBoardRequest) ProviderSyncRequest() ProviderSyncRequest {
	return ProviderSyncRequest{
		RunID:      request.RunID,
		ProjectID:  request.ProjectID,
		WorkflowID: request.WorkflowID,
		Source:     request.Source,
	}
}

type Service struct {
	providerResolver ProviderResolver
}

func NewService(providerResolver ProviderResolver) (*Service, error) {
	if providerResolver == nil {
		return nil, failures.WrapTerminal(errors.New("tracker provider resolver is required"))
	}
	return &Service{providerResolver: providerResolver}, nil
}

func (service *Service) SyncBoard(ctx context.Context, request SyncBoardRequest) (domaintracker.Board, error) {
	if err := request.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	providerRequest := request.ProviderSyncRequest()
	provider, err := service.providerResolver.Resolve(ctx, providerRequest)
	if err != nil {
		return domaintracker.Board{}, ensureClassified(err)
	}
	if provider == nil {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("tracker provider resolver returned nil provider"))
	}
	board, err := provider.SyncBoard(ctx, providerRequest)
	if err != nil {
		return domaintracker.Board{}, ensureClassified(err)
	}
	if err := board.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	if board.RunID != request.RunID {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("board run_id %q does not match request run_id %q", board.RunID, request.RunID))
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
