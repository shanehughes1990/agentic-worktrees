package tracker

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"fmt"
)

type ExternalProvider struct {
	sourceKind domaintracker.SourceKind
}

func NewJiraProvider() *ExternalProvider {
	return &ExternalProvider{sourceKind: domaintracker.SourceKindJira}
}

func NewLinearProvider() *ExternalProvider {
	return &ExternalProvider{sourceKind: domaintracker.SourceKindLinear}
}

func (provider *ExternalProvider) SyncBoard(ctx context.Context, request applicationtracker.ProviderSyncRequest) (domaintracker.Board, error) {
	_ = ctx
	if err := request.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	if request.Source.Kind != provider.sourceKind {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("external provider kind %q does not support source kind %q", provider.sourceKind, request.Source.Kind))
	}
	return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("%s tracker provider sync boundary is defined but not implemented", provider.sourceKind))
}
