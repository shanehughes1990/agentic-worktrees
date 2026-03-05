package lifecycle

import (
	domainlifecycle "agentic-orchestrator/internal/domain/lifecycle"
	"context"
	"errors"
)

var ErrStoreRequired = errors.New("lifecycle: event store is required")

type EventStore interface {
	Append(ctx context.Context, event domainlifecycle.Event) (domainlifecycle.Event, error)
}

type Service struct {
	store EventStore
}

func NewService(store EventStore) (*Service, error) {
	if store == nil {
		return nil, ErrStoreRequired
	}
	return &Service{store: store}, nil
}

func (service *Service) AppendEvent(ctx context.Context, event domainlifecycle.Event) (domainlifecycle.Event, error) {
	if service == nil || service.store == nil {
		return domainlifecycle.Event{}, ErrStoreRequired
	}
	if err := event.ValidateForAppend(); err != nil {
		return domainlifecycle.Event{}, err
	}
	return service.store.Append(ctx, event)
}
