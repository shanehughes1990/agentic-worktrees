package postgres

import (
	domainstream "agentic-orchestrator/internal/domain/stream"
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newStreamTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func TestEventStoreAppendAndListFromOffset(t *testing.T) {
	store, err := NewEventStore(newStreamTestDB(t))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	event, err := store.Append(context.Background(), domainstream.Event{
		EventID:    "evt-1",
		OccurredAt: time.Now().UTC(),
		Source:     domainstream.SourceACP,
		EventType:  domainstream.EventType("stream.agent.chunk"),
		CorrelationIDs: domainstream.CorrelationIDs{
			RunID:         "run-1",
			TaskID:        "task-1",
			JobID:         "job-1",
			SessionID:     "session-1",
			CorrelationID: "corr-1",
		},
		Payload: map[string]any{"text": "hello"},
	})
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if event.StreamOffset == 0 {
		t.Fatalf("expected stream offset to be assigned")
	}

	events, err := store.ListFromOffset(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("list from offset: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventID != "evt-1" {
		t.Fatalf("unexpected event id: %s", events[0].EventID)
	}
}

func TestEventStoreListFromOffsetHonorsOffsetAndLimit(t *testing.T) {
	store, err := NewEventStore(newStreamTestDB(t))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	for index := range 3 {
		_, err := store.Append(context.Background(), domainstream.Event{
			EventID:    "evt-" + string(rune('a'+index)),
			OccurredAt: time.Now().UTC(),
			Source:     domainstream.SourceWorker,
			EventType:  domainstream.EventType("stream.session.updated"),
			CorrelationIDs: domainstream.CorrelationIDs{
				RunID:         "run-1",
				TaskID:        "task-1",
				JobID:         "job-1",
				SessionID:     "session-1",
				CorrelationID: "corr-1",
			},
			Payload: map[string]any{"index": index},
		})
		if err != nil {
			t.Fatalf("append %d: %v", index, err)
		}
	}
	events, err := store.ListFromOffset(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("list from offset: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].StreamOffset != 2 {
		t.Fatalf("unexpected offset: %d", events[0].StreamOffset)
	}
}
