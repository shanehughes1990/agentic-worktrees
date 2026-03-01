package agent

import (
	domainstream "agentic-orchestrator/internal/domain/stream"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SessionStateReader struct {
	basePath string
}

func NewSessionStateReader(basePath string) (*SessionStateReader, error) {
	normalizedBasePath := strings.TrimSpace(basePath)
	if normalizedBasePath == "" {
		homeDirectory, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("session state reader: user home directory: %w", err)
		}
		normalizedBasePath = filepath.Join(homeDirectory, ".copilot", "session-state")
	}
	return &SessionStateReader{basePath: normalizedBasePath}, nil
}

func (reader *SessionStateReader) ReadSessionEvents(ctx context.Context, sessionID string, correlationIDs domainstream.CorrelationIDs, limit int) ([]domainstream.Event, error) {
	normalizedSessionID := strings.TrimSpace(sessionID)
	if normalizedSessionID == "" {
		return nil, fmt.Errorf("session state reader: session_id is required")
	}
	eventsPath := filepath.Join(reader.basePath, normalizedSessionID, "events.jsonl")
	file, err := os.Open(eventsPath)
	if err != nil {
		return nil, fmt.Errorf("session state reader: open events file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()
	events := make([]domainstream.Event, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if limit > 0 && len(events) >= limit {
			break
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		raw := map[string]any{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		event, ok := normalizeSessionStateEvent(raw, normalizedSessionID, correlationIDs)
		if !ok {
			continue
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("session state reader: scan events file: %w", err)
	}
	return events, nil
}

func normalizeSessionStateEvent(raw map[string]any, sessionID string, correlationIDs domainstream.CorrelationIDs) (domainstream.Event, bool) {
	now := time.Now().UTC()
	event := domainstream.Event{
		EventID:       fmt.Sprintf("session-file-%d", now.UnixNano()),
		OccurredAt:    now,
		Source:        domainstream.SourceSessionFile,
		EventType:     domainstream.EventSessionRecovered,
		CorrelationIDs: correlationIDs,
		Payload: map[string]any{
			"session_id": sessionID,
			"raw":        raw,
		},
	}
	event.CorrelationIDs.SessionID = sessionID
	if strings.TrimSpace(event.CorrelationIDs.CorrelationID) == "" {
		event.CorrelationIDs.CorrelationID = sessionID
	}
	if err := event.Validate(); err != nil {
		return domainstream.Event{}, false
	}
	return event, true
}
