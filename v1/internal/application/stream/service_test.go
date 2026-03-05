package stream

import (
	domainstream "agentic-orchestrator/internal/domain/stream"
	"context"
	"testing"
	"time"
)

type fakeStore struct {
	events []domainstream.Event
}

func (store *fakeStore) Append(_ context.Context, event domainstream.Event) (domainstream.Event, error) {
	event.StreamOffset = uint64(len(store.events) + 1)
	store.events = append(store.events, event)
	return event, nil
}

func (store *fakeStore) ListFromOffset(_ context.Context, offset uint64, limit int) ([]domainstream.Event, error) {
	start := 0
	for index := range store.events {
		if store.events[index].StreamOffset > offset {
			start = index
			break
		}
	}
	end := len(store.events)
	if limit > 0 && start+limit < end {
		end = start + limit
	}
	return append([]domainstream.Event(nil), store.events[start:end]...), nil
}

type fakeInjector struct {
	sessionID string
	prompt    string
}

func (injector *fakeInjector) InjectPrompt(_ context.Context, sessionID string, prompt string) error {
	injector.sessionID = sessionID
	injector.prompt = prompt
	return nil
}

func TestServiceAppendAndReplay(t *testing.T) {
	store := &fakeStore{}
	service, err := NewService(store)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	persistedEvent, err := service.AppendAndPublish(context.Background(), domainstream.Event{
		EventID:    "evt-1",
		OccurredAt: time.Now().UTC(),
		Source:     domainstream.SourceWorker,
		EventType:  domainstream.EventSessionUpdated,
		CorrelationIDs: domainstream.CorrelationIDs{
			CorrelationID: "corr-1",
		},
		Payload: map[string]any{"key": "value"},
	})
	if err != nil {
		t.Fatalf("append and publish: %v", err)
	}
	if persistedEvent.StreamOffset != 1 {
		t.Fatalf("expected offset 1, got %d", persistedEvent.StreamOffset)
	}
	events, err := service.ReplayFromOffset(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestServiceInjectPromptPublishesEvent(t *testing.T) {
	store := &fakeStore{}
	injector := &fakeInjector{}
	service, err := NewService(store)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.SetPromptInjector(injector)
	injectedEvent, err := service.InjectPrompt(context.Background(), "session-1", "do it", domainstream.CorrelationIDs{CorrelationID: "corr-1"})
	if err != nil {
		t.Fatalf("inject prompt: %v", err)
	}
	if injector.sessionID != "session-1" || injector.prompt != "do it" {
		t.Fatalf("unexpected injector call: session=%s prompt=%s", injector.sessionID, injector.prompt)
	}
	if injectedEvent.EventType != domainstream.EventSessionInjectedPrompt {
		t.Fatalf("unexpected event type: %s", injectedEvent.EventType)
	}
}

func TestServiceDropsOldestEventOnBackpressure(t *testing.T) {
	store := &fakeStore{}
	service, err := NewService(store)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	_, channel, _ := service.Subscribe(1)
	_, err = service.AppendAndPublish(context.Background(), domainstream.Event{
		EventID:    "evt-1",
		OccurredAt: time.Now().UTC(),
		Source:     domainstream.SourceWorker,
		EventType:  domainstream.EventSessionUpdated,
		CorrelationIDs: domainstream.CorrelationIDs{
			CorrelationID: "corr-1",
		},
		Payload: map[string]any{"sequence": 1},
	})
	if err != nil {
		t.Fatalf("append first event: %v", err)
	}
	_, err = service.AppendAndPublish(context.Background(), domainstream.Event{
		EventID:    "evt-2",
		OccurredAt: time.Now().UTC(),
		Source:     domainstream.SourceWorker,
		EventType:  domainstream.EventSessionUpdated,
		CorrelationIDs: domainstream.CorrelationIDs{
			CorrelationID: "corr-1",
		},
		Payload: map[string]any{"sequence": 2},
	})
	if err != nil {
		t.Fatalf("append second event: %v", err)
	}
	latest, open := <-channel
	if !open {
		t.Fatalf("expected subscriber channel to remain open")
	}
	sequence, ok := latest.Payload["sequence"].(int)
	if !ok {
		t.Fatalf("expected integer sequence payload, got %#v", latest.Payload["sequence"])
	}
	if sequence != 2 {
		t.Fatalf("expected latest event sequence 2 after backpressure drop, got %d", sequence)
	}

	select {
	case extra := <-channel:
		t.Fatalf("expected no additional buffered events, got %#v", extra)
	default:
	}
}

func TestServiceNotifyExternalChangeSignalsSubscribers(t *testing.T) {
	store := &fakeStore{}
	service, err := NewService(store)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	_, changes, cancel := service.SubscribeChanges(1)
	defer cancel()

	service.NotifyExternalChange()

	select {
	case <-changes:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected external change signal")
	}
}

func TestServicePublishLiveSignalsSubscribers(t *testing.T) {
	store := &fakeStore{}
	service, err := NewService(store)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	_, channel, cancel := service.Subscribe(1)
	defer cancel()

	service.PublishLive(domainstream.Event{
		EventID:    "live-1",
		OccurredAt: time.Now().UTC(),
		Source:     domainstream.SourceWorker,
		EventType:  domainstream.EventSessionUpdated,
		CorrelationIDs: domainstream.CorrelationIDs{
			CorrelationID: "corr-live-1",
		},
		Payload: map[string]any{"live": true},
	})

	select {
	case received := <-channel:
		if received.EventID != "live-1" {
			t.Fatalf("expected live event id live-1, got %s", received.EventID)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected live event on subscription")
	}
}
