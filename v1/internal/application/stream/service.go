package stream

import (
	domainstream "agentic-orchestrator/internal/domain/stream"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrStoreRequired          = errors.New("stream: event store is required")
	ErrPromptInjectorRequired = errors.New("stream: prompt injector is not configured")
)

type PromptInjector interface {
	InjectPrompt(ctx context.Context, sessionID string, prompt string) error
}

type SessionHealthEvaluator interface {
	Evaluate(ctx context.Context, sessionID string) (map[string]any, error)
}

type Service struct {
	store             EventStore
	injector          PromptInjector
	healthEvaluator   SessionHealthEvaluator
	mu                sync.Mutex
	subscribers       map[uint64]chan domainstream.Event
	changeSubscribers map[uint64]chan struct{}
	nextID            uint64
}

func NewService(store EventStore) (*Service, error) {
	if store == nil {
		return nil, ErrStoreRequired
	}
	return &Service{
		store:             store,
		subscribers:       map[uint64]chan domainstream.Event{},
		changeSubscribers: map[uint64]chan struct{}{},
	}, nil
}

func (service *Service) SetPromptInjector(injector PromptInjector) {
	if service == nil {
		return
	}
	service.injector = injector
}

func (service *Service) SetHealthEvaluator(evaluator SessionHealthEvaluator) {
	if service == nil {
		return
	}
	service.healthEvaluator = evaluator
}

func (service *Service) AppendAndPublish(ctx context.Context, event domainstream.Event) (domainstream.Event, error) {
	if service == nil || service.store == nil {
		return domainstream.Event{}, ErrStoreRequired
	}
	persistedEvent, err := service.store.Append(ctx, event)
	if err != nil {
		return domainstream.Event{}, err
	}
	service.broadcast(persistedEvent)
	service.broadcastChange()
	return persistedEvent, nil
}

func (service *Service) ReplayFromOffset(ctx context.Context, offset uint64, limit int) ([]domainstream.Event, error) {
	if service == nil || service.store == nil {
		return nil, ErrStoreRequired
	}
	return service.store.ListFromOffset(ctx, offset, limit)
}

func (service *Service) Subscribe(bufferSize int) (uint64, <-chan domainstream.Event, func()) {
	if bufferSize <= 0 {
		bufferSize = 64
	}
	channel := make(chan domainstream.Event, bufferSize)
	id := atomic.AddUint64(&service.nextID, 1)
	service.mu.Lock()
	service.subscribers[id] = channel
	service.mu.Unlock()
	cancel := func() {
		service.mu.Lock()
		current, exists := service.subscribers[id]
		if exists {
			delete(service.subscribers, id)
			close(current)
		}
		service.mu.Unlock()
	}
	return id, channel, cancel
}

func (service *Service) SubscribeChanges(bufferSize int) (uint64, <-chan struct{}, func()) {
	if bufferSize <= 0 {
		bufferSize = 32
	}
	channel := make(chan struct{}, bufferSize)
	id := atomic.AddUint64(&service.nextID, 1)
	service.mu.Lock()
	service.changeSubscribers[id] = channel
	service.mu.Unlock()
	cancel := func() {
		service.mu.Lock()
		current, exists := service.changeSubscribers[id]
		if exists {
			delete(service.changeSubscribers, id)
			close(current)
		}
		service.mu.Unlock()
	}
	return id, channel, cancel
}

func (service *Service) NotifyExternalChange() {
	if service == nil {
		return
	}
	service.broadcastChange()
}

func (service *Service) PublishLive(event domainstream.Event) {
	if service == nil {
		return
	}
	service.broadcast(event)
}

func (service *Service) InjectPrompt(ctx context.Context, sessionID string, prompt string, correlationIDs domainstream.CorrelationIDs) (domainstream.Event, error) {
	if service == nil || service.store == nil {
		return domainstream.Event{}, ErrStoreRequired
	}
	if service.injector == nil {
		return domainstream.Event{}, ErrPromptInjectorRequired
	}
	normalizedSessionID := strings.TrimSpace(sessionID)
	if normalizedSessionID == "" {
		return domainstream.Event{}, fmt.Errorf("stream: session_id is required")
	}
	normalizedPrompt := strings.TrimSpace(prompt)
	if normalizedPrompt == "" {
		return domainstream.Event{}, fmt.Errorf("stream: prompt is required")
	}
	if err := service.injector.InjectPrompt(ctx, normalizedSessionID, normalizedPrompt); err != nil {
		return domainstream.Event{}, err
	}
	event := domainstream.Event{
		EventID:        fmt.Sprintf("inject-%d", time.Now().UTC().UnixNano()),
		OccurredAt:     time.Now().UTC(),
		Source:         domainstream.SourceWorker,
		EventType:      domainstream.EventSessionInjectedPrompt,
		CorrelationIDs: correlationIDs,
		Payload: map[string]any{
			"session_id": normalizedSessionID,
			"prompt":     normalizedPrompt,
		},
	}
	return service.AppendAndPublish(ctx, event)
}

func (service *Service) PublishHealth(ctx context.Context, sessionID string, correlationIDs domainstream.CorrelationIDs) (domainstream.Event, error) {
	if service == nil || service.store == nil {
		return domainstream.Event{}, ErrStoreRequired
	}
	if service.healthEvaluator == nil {
		return domainstream.Event{}, fmt.Errorf("stream: health evaluator is not configured")
	}
	status, err := service.healthEvaluator.Evaluate(ctx, sessionID)
	if err != nil {
		return domainstream.Event{}, err
	}
	event := domainstream.Event{
		EventID:        fmt.Sprintf("health-%d", time.Now().UTC().UnixNano()),
		OccurredAt:     time.Now().UTC(),
		Source:         domainstream.SourceWorker,
		EventType:      domainstream.EventSessionHealth,
		CorrelationIDs: correlationIDs,
		Payload:        status,
	}
	return service.AppendAndPublish(ctx, event)
}

func (service *Service) broadcast(event domainstream.Event) {
	service.mu.Lock()
	for _, channel := range service.subscribers {
		select {
		case channel <- event:
		default:
			select {
			case <-channel:
			default:
			}
			select {
			case channel <- event:
			default:
			}
		}
	}
	service.mu.Unlock()
}

func (service *Service) broadcastChange() {
	service.mu.Lock()
	for _, channel := range service.changeSubscribers {
		select {
		case channel <- struct{}{}:
		default:
			select {
			case <-channel:
			default:
			}
			select {
			case channel <- struct{}{}:
			default:
			}
		}
	}
	service.mu.Unlock()
}
