package tracker

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"errors"
	"fmt"
)

type ProviderRegistry struct {
	providers map[domaintracker.SourceKind]applicationtracker.Provider
}

func NewProviderRegistry(providers map[domaintracker.SourceKind]applicationtracker.Provider) (*ProviderRegistry, error) {
	if len(providers) == 0 {
		return nil, failures.WrapTerminal(errors.New("tracker providers are required"))
	}
	registered := make(map[domaintracker.SourceKind]applicationtracker.Provider, len(providers))
	for kind, provider := range providers {
		if err := kind.Validate(); err != nil {
			return nil, err
		}
		if provider == nil {
			return nil, failures.WrapTerminal(fmt.Errorf("tracker provider is required for kind %q", kind))
		}
		registered[kind] = provider
	}
	return &ProviderRegistry{providers: registered}, nil
}

func (registry *ProviderRegistry) Resolve(ctx context.Context, request applicationtracker.ProviderSyncRequest) (applicationtracker.Provider, error) {
	_ = ctx
	if err := request.Validate(); err != nil {
		return nil, err
	}
	if registry == nil {
		return nil, failures.WrapTerminal(errors.New("tracker provider registry is not initialized"))
	}
	provider, exists := registry.providers[request.Source.Kind]
	if !exists {
		return nil, failures.WrapTerminal(fmt.Errorf("tracker provider for kind %q is not registered", request.Source.Kind))
	}
	return provider, nil
}
