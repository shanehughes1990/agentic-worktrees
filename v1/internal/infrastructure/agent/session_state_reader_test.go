package agent

import (
	domainstream "agentic-orchestrator/internal/domain/stream"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSessionStateReaderReadSessionEvents(t *testing.T) {
	tempDirectory := t.TempDir()
	sessionID := "session-1"
	sessionDirectory := filepath.Join(tempDirectory, sessionID)
	if err := os.MkdirAll(sessionDirectory, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	eventsPath := filepath.Join(sessionDirectory, "events.jsonl")
	if err := os.WriteFile(eventsPath, []byte("{\"k\":\"v\"}\n"), 0o644); err != nil {
		t.Fatalf("write events: %v", err)
	}
	reader, err := NewSessionStateReader(tempDirectory)
	if err != nil {
		t.Fatalf("new reader: %v", err)
	}
	events, err := reader.ReadSessionEvents(context.Background(), sessionID, domainstream.CorrelationIDs{CorrelationID: "corr-1"}, 10)
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != domainstream.EventSessionRecovered {
		t.Fatalf("unexpected event type: %s", events[0].EventType)
	}
}
