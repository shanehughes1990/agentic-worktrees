package agent

import (
	"agentic-orchestrator/internal/domain/stream"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSessionStateStreamerProcessSessionEventFile(t *testing.T) {
	t.Parallel()

	tempDirectory := t.TempDir()
	sessionID := "session-1"
	sessionDirectory := filepath.Join(tempDirectory, sessionID)
	if err := os.MkdirAll(filepath.Join(sessionDirectory, "checkpoints"), 0o755); err != nil {
		t.Fatalf("mkdir checkpoints: %v", err)
	}

	eventsPath := filepath.Join(sessionDirectory, "events.jsonl")
	eventsJSONL := "{" +
		"\"type\":\"session.start\",\"timestamp\":\"2026-03-05T03:03:39.694Z\",\"data\":{\"run_id\":\"run-1\",\"project_id\":\"project-1\"}}\n" +
		"{\"type\":\"tool.execution_start\",\"timestamp\":\"2026-03-05T03:03:45.677Z\",\"data\":{\"task_id\":\"task-1\",\"job_id\":\"job-1\"}}\n"
	if err := os.WriteFile(eventsPath, []byte(eventsJSONL), 0o644); err != nil {
		t.Fatalf("write events.jsonl: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionDirectory, "workspace.yaml"), []byte("workspace: test\n"), 0o644); err != nil {
		t.Fatalf("write workspace.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionDirectory, "plan.md"), []byte("# Plan\n"), 0o644); err != nil {
		t.Fatalf("write plan.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionDirectory, "checkpoints", "index.md"), []byte("# Index\n"), 0o644); err != nil {
		t.Fatalf("write checkpoints index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionDirectory, "session.db"), []byte("internal"), 0o644); err != nil {
		t.Fatalf("write session.db: %v", err)
	}

	streamer := &SessionStateStreamer{basePath: tempDirectory, pollInterval: 1}
	cursor := sessionStateCursor{}
	published := make([]stream.Event, 0, 2)

	nextCursor, err := streamer.processSessionEventFile(context.Background(), sessionID, cursor, func(_ context.Context, event stream.Event) error {
		published = append(published, event)
		return nil
	})
	if err != nil {
		t.Fatalf("process session event file: %v", err)
	}
	if nextCursor.sequence != 2 {
		t.Fatalf("expected sequence=2, got %d", nextCursor.sequence)
	}
	if len(published) != 2 {
		t.Fatalf("expected 2 events, got %d", len(published))
	}

	if published[0].EventType != stream.EventSessionStarted {
		t.Fatalf("expected first event type %q, got %q", stream.EventSessionStarted, published[0].EventType)
	}
	if published[1].EventType != stream.EventToolStarted {
		t.Fatalf("expected second event type %q, got %q", stream.EventToolStarted, published[1].EventType)
	}

	if published[1].CorrelationIDs.RunID != "run-1" {
		t.Fatalf("expected run_id propagation from prior line, got %q", published[1].CorrelationIDs.RunID)
	}
	if published[1].CorrelationIDs.ProjectID != "project-1" {
		t.Fatalf("expected project_id propagation from prior line, got %q", published[1].CorrelationIDs.ProjectID)
	}
	if published[1].CorrelationIDs.TaskID != "task-1" {
		t.Fatalf("expected task_id extraction, got %q", published[1].CorrelationIDs.TaskID)
	}
	if published[1].CorrelationIDs.JobID != "job-1" {
		t.Fatalf("expected job_id extraction, got %q", published[1].CorrelationIDs.JobID)
	}

	references, ok := published[0].Payload["session_refs"].(map[string]any)
	if !ok {
		t.Fatalf("expected session_refs map in payload")
	}
	if _, exists := references["workspace_yaml"]; !exists {
		t.Fatalf("expected workspace_yaml reference")
	}
	if _, exists := references["plan_md"]; !exists {
		t.Fatalf("expected plan_md reference")
	}
	if _, exists := references["checkpoints_index"]; !exists {
		t.Fatalf("expected checkpoints_index reference")
	}
	if _, exists := references["session_db"]; exists {
		t.Fatalf("session_db should not be included in session_refs")
	}
	if published[0].Payload["session_file_path"] != "events.jsonl" {
		t.Fatalf("expected redacted session_file_path, got %v", published[0].Payload["session_file_path"])
	}
	workspaceReference, ok := references["workspace_yaml"].(map[string]any)
	if !ok {
		t.Fatalf("expected workspace_yaml reference map")
	}
	internalPath, _ := workspaceReference["internal_path"].(string)
	if strings.Contains(internalPath, tempDirectory) {
		t.Fatalf("workspace_yaml internal_path must not expose absolute base directory: %q", internalPath)
	}
}

func TestSessionStateStreamerPollSessionDirectoriesPublishesDerivedHeartbeats(t *testing.T) {
	t.Parallel()

	tempDirectory := t.TempDir()
	sessionID := "session-heartbeat-1"
	sessionDirectory := filepath.Join(tempDirectory, sessionID)
	if err := os.MkdirAll(filepath.Join(sessionDirectory, "checkpoints"), 0o755); err != nil {
		t.Fatalf("mkdir checkpoints: %v", err)
	}

	eventsPath := filepath.Join(sessionDirectory, "events.jsonl")
	eventsJSONL := "{" +
		"\"type\":\"session.start\",\"timestamp\":\"2026-03-05T03:03:39.694Z\",\"data\":{\"run_id\":\"run-1\"}}\n"
	if err := os.WriteFile(eventsPath, []byte(eventsJSONL), 0o644); err != nil {
		t.Fatalf("write events.jsonl: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionDirectory, "workspace.yaml"), []byte("workspace: test\n"), 0o644); err != nil {
		t.Fatalf("write workspace.yaml: %v", err)
	}

	streamer := &SessionStateStreamer{basePath: tempDirectory, pollInterval: 1, heartbeatInterval: time.Second}
	published := make([]stream.Event, 0, 4)
	cursors := map[string]sessionStateCursor{}

	if err := streamer.pollSessionDirectories(context.Background(), func(_ context.Context, event stream.Event) error {
		published = append(published, event)
		return nil
	}, cursors); err != nil {
		t.Fatalf("poll session directories: %v", err)
	}

	hasActivityHeartbeat := false
	hasContextHeartbeat := false
	for _, event := range published {
		if event.EventType != stream.EventSessionHealth {
			continue
		}
		layer, _ := event.Payload["heartbeat_layer"].(string)
		if layer == "activity" {
			hasActivityHeartbeat = true
		}
		if layer == "context" {
			hasContextHeartbeat = true
		}
	}
	if !hasActivityHeartbeat || !hasContextHeartbeat {
		t.Fatalf("expected activity and context derived heartbeats, got activity=%t context=%t", hasActivityHeartbeat, hasContextHeartbeat)
	}
}
