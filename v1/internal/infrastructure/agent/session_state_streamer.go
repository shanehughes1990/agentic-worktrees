package agent

import (
	"agentic-orchestrator/internal/domain/stream"
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type SessionStateStreamer struct {
	basePath          string
	pollInterval      time.Duration
	heartbeatInterval time.Duration
}

type sessionStateCursor struct {
	byteOffset      int64
	lineNumber      int64
	sequence        int64
	context         stream.CorrelationIDs
	heartbeatWindow int64
}

func NewSessionStateStreamer(basePath string) (*SessionStateStreamer, error) {
	normalizedBasePath := strings.TrimSpace(basePath)
	if normalizedBasePath == "" {
		var err error
		normalizedBasePath, err = defaultCopilotSessionStatePath()
		if err != nil {
			return nil, err
		}
	}
	return &SessionStateStreamer{basePath: normalizedBasePath, pollInterval: 750 * time.Millisecond, heartbeatInterval: 15 * time.Second}, nil
}

func defaultCopilotSessionStatePath() (string, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("session state streamer: user home directory: %w", err)
	}
	// TODO(shane): make this runtime-configurable through an internal worker config source; do not require user-supplied boot configuration.
	return filepath.Join(homeDirectory, ".copilot", "session-state"), nil
}

func (streamer *SessionStateStreamer) Run(ctx context.Context, publish func(context.Context, stream.Event) error) error {
	if streamer == nil {
		return fmt.Errorf("session state streamer: streamer is nil")
	}
	if publish == nil {
		return fmt.Errorf("session state streamer: publish callback is required")
	}
	cursors := map[string]sessionStateCursor{}
	ticker := time.NewTicker(streamer.pollInterval)
	defer ticker.Stop()

	for {
		if err := streamer.pollSessionDirectories(ctx, publish, cursors); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (streamer *SessionStateStreamer) pollSessionDirectories(ctx context.Context, publish func(context.Context, stream.Event) error, cursors map[string]sessionStateCursor) error {
	entries, err := os.ReadDir(streamer.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("session state streamer: read base path: %w", err)
	}

	sessionIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sessionID := strings.TrimSpace(entry.Name())
		if sessionID == "" {
			continue
		}
		sessionIDs = append(sessionIDs, sessionID)
	}
	sort.Strings(sessionIDs)

	for _, sessionID := range sessionIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		cursor := cursors[sessionID]
		nextCursor, err := streamer.processSessionEventFile(ctx, sessionID, cursor, publish)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		nextCursor, err = streamer.publishDerivedSessionHeartbeats(ctx, sessionID, nextCursor, publish)
		if err != nil {
			return err
		}
		cursors[sessionID] = nextCursor
	}
	return nil
}

func (streamer *SessionStateStreamer) publishDerivedSessionHeartbeats(ctx context.Context, sessionID string, cursor sessionStateCursor, publish func(context.Context, stream.Event) error) (sessionStateCursor, error) {
	if streamer == nil || publish == nil {
		return cursor, nil
	}
	if streamer.heartbeatInterval <= 0 {
		return cursor, nil
	}
	now := time.Now().UTC()
	window := now.UnixNano() / int64(streamer.heartbeatInterval)
	if window == cursor.heartbeatWindow {
		return cursor, nil
	}
	cursor.heartbeatWindow = window

	sessionDirectory := filepath.Join(streamer.basePath, sessionID)
	eventsPath := filepath.Join(sessionDirectory, "events.jsonl")
	eventsInfo, err := os.Stat(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cursor, nil
		}
		return cursor, err
	}

	contextIDs := cursor.context
	contextIDs.SessionID = strings.TrimSpace(sessionID)
	if strings.TrimSpace(contextIDs.CorrelationID) == "" {
		contextIDs.CorrelationID = "session:" + strings.TrimSpace(sessionID)
	}

	activityFreshness := now.Sub(eventsInfo.ModTime().UTC())
	activityEvent := stream.Event{
		EventID:        buildSessionHeartbeatEventID(sessionID, "activity", window),
		OccurredAt:     now,
		Source:         stream.SourceSessionFile,
		EventType:      stream.EventSessionHealth,
		CorrelationIDs: contextIDs,
		Payload: map[string]any{
			"session_id":                 sessionID,
			"session_event_seq":          cursor.sequence,
			"session_file_path":          "events.jsonl",
			"heartbeat_layer":            "activity",
			"heartbeat_source":           "session_file_mtime",
			"heartbeat_freshness_ms":     activityFreshness.Milliseconds(),
			"heartbeat_status":           freshnessStatus(activityFreshness, 20*time.Second),
			"last_activity_heartbeat_at": now.Format(time.RFC3339),
			"last_activity_at":           eventsInfo.ModTime().UTC().Format(time.RFC3339),
			"heartbeat_quorum_state":     quorumStateFromFreshness(activityFreshness, 20*time.Second),
			"session_refs":               buildSessionReferences(sessionDirectory, sessionID),
		},
	}
	if err := activityEvent.Validate(); err == nil {
		if err := publish(ctx, activityEvent); err != nil {
			return cursor, err
		}
	}

	refs := buildSessionReferences(sessionDirectory, sessionID)
	lastContextAt := latestReferenceUpdate(refs)
	contextFreshness := now.Sub(lastContextAt)
	contextEvent := stream.Event{
		EventID:        buildSessionHeartbeatEventID(sessionID, "context", window),
		OccurredAt:     now,
		Source:         stream.SourceSessionFile,
		EventType:      stream.EventSessionHealth,
		CorrelationIDs: contextIDs,
		Payload: map[string]any{
			"session_id":                sessionID,
			"session_event_seq":         cursor.sequence,
			"heartbeat_layer":           "context",
			"heartbeat_source":          "session_refs",
			"heartbeat_freshness_ms":    contextFreshness.Milliseconds(),
			"heartbeat_status":          freshnessStatus(contextFreshness, 60*time.Second),
			"last_context_heartbeat_at": now.Format(time.RFC3339),
			"last_context_update_at":    lastContextAt.UTC().Format(time.RFC3339),
			"heartbeat_quorum_state":    "running_observed",
			"session_refs":              refs,
		},
	}
	if err := contextEvent.Validate(); err == nil {
		if err := publish(ctx, contextEvent); err != nil {
			return cursor, err
		}
	}
	return cursor, nil
}

func (streamer *SessionStateStreamer) processSessionEventFile(ctx context.Context, sessionID string, cursor sessionStateCursor, publish func(context.Context, stream.Event) error) (sessionStateCursor, error) {
	eventsPath := filepath.Join(streamer.basePath, sessionID, "events.jsonl")
	file, err := os.Open(eventsPath)
	if err != nil {
		return cursor, err
	}
	defer func() {
		_ = file.Close()
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return cursor, fmt.Errorf("session state streamer: stat events file: %w", err)
	}
	if cursor.byteOffset > fileInfo.Size() {
		cursor.byteOffset = 0
		cursor.lineNumber = 0
		cursor.sequence = 0
	}

	if _, err := file.Seek(cursor.byteOffset, io.SeekStart); err != nil {
		return cursor, fmt.Errorf("session state streamer: seek events file: %w", err)
	}

	reader := bufio.NewReader(file)
	for {
		if err := ctx.Err(); err != nil {
			return cursor, err
		}

		lineStartOffset := cursor.byteOffset
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			if readErr == io.EOF {
				// Keep partial line for the next poll when the writer flushes the newline.
				return cursor, nil
			}
			return cursor, fmt.Errorf("session state streamer: read events file line: %w", readErr)
		}

		cursor.byteOffset += int64(len(line))
		cursor.lineNumber += 1
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		raw := map[string]any{}
		if err := json.Unmarshal([]byte(trimmedLine), &raw); err != nil {
			continue
		}

		nextContext := mergeCorrelationContext(cursor.context, extractCorrelationContext(raw))
		nextContext.SessionID = sessionID
		if strings.TrimSpace(nextContext.CorrelationID) == "" {
			nextContext.CorrelationID = "session:" + sessionID
		}
		cursor.context = nextContext

		cursor.sequence += 1
		event := stream.Event{
			EventID:        buildSessionFileEventID(sessionID, lineStartOffset, trimmedLine),
			OccurredAt:     parseOccurredAt(raw),
			Source:         stream.SourceSessionFile,
			EventType:      mapSessionEventType(raw),
			CorrelationIDs: nextContext,
			Payload: map[string]any{
				"session_id":          sessionID,
				"session_event_seq":   cursor.sequence,
				"session_file_path":   "events.jsonl",
				"session_file_offset": lineStartOffset,
				"session_file_line":   cursor.lineNumber,
				"session_event_type":  strings.TrimSpace(stringValue(raw, "type")),
				"raw_event":           raw,
				"session_refs":        buildSessionReferences(filepath.Join(streamer.basePath, sessionID), sessionID),
			},
		}
		if err := event.Validate(); err != nil {
			continue
		}
		if err := publish(ctx, event); err != nil {
			return cursor, err
		}
	}
}

func parseOccurredAt(raw map[string]any) time.Time {
	timestamp := strings.TrimSpace(stringValue(raw, "timestamp"))
	if timestamp == "" {
		return time.Now().UTC()
	}
	parsed, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return time.Now().UTC()
	}
	return parsed.UTC()
}

func mapSessionEventType(raw map[string]any) stream.EventType {
	rawType := strings.ToLower(strings.TrimSpace(stringValue(raw, "type")))
	switch rawType {
	case "session.start":
		return stream.EventSessionStarted
	case "session.end", "session.ended", "session.terminated":
		return stream.EventSessionEnded
	case "assistant.turn_end":
		return stream.EventAgentTurnCompleted
	case "tool.execution_start":
		return stream.EventToolStarted
	case "tool.execution_complete":
		return stream.EventToolCompleted
	case "permission.requested":
		return stream.EventPermissionRequested
	case "permission.decided":
		return stream.EventPermissionDecided
	}
	if strings.Contains(rawType, "checkpoint") {
		return stream.EventSessionCheckpointed
	}
	if rawType == "assistant.message" {
		return stream.EventAgentChunk
	}
	return stream.EventSessionUpdated
}

func mergeCorrelationContext(current stream.CorrelationIDs, discovered stream.CorrelationIDs) stream.CorrelationIDs {
	merged := current
	if strings.TrimSpace(discovered.RunID) != "" {
		merged.RunID = strings.TrimSpace(discovered.RunID)
	}
	if strings.TrimSpace(discovered.TaskID) != "" {
		merged.TaskID = strings.TrimSpace(discovered.TaskID)
	}
	if strings.TrimSpace(discovered.JobID) != "" {
		merged.JobID = strings.TrimSpace(discovered.JobID)
	}
	if strings.TrimSpace(discovered.ProjectID) != "" {
		merged.ProjectID = strings.TrimSpace(discovered.ProjectID)
	}
	if strings.TrimSpace(discovered.CorrelationID) != "" {
		merged.CorrelationID = strings.TrimSpace(discovered.CorrelationID)
	}
	return merged
}

func extractCorrelationContext(raw map[string]any) stream.CorrelationIDs {
	return stream.CorrelationIDs{
		RunID:         findFirstStringValue(raw, "run_id", "runId", "runID"),
		TaskID:        findFirstStringValue(raw, "task_id", "taskId", "taskID"),
		JobID:         findFirstStringValue(raw, "job_id", "jobId", "jobID"),
		ProjectID:     findFirstStringValue(raw, "project_id", "projectId", "projectID"),
		CorrelationID: findFirstStringValue(raw, "correlation_id", "correlationId", "correlationID"),
	}
}

func findFirstStringValue(root any, keys ...string) string {
	seenKeys := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		normalized := strings.ToLower(strings.TrimSpace(key))
		if normalized == "" {
			continue
		}
		seenKeys[normalized] = struct{}{}
	}
	return findFirstStringValueRecursive(root, seenKeys)
}

func findFirstStringValueRecursive(node any, keys map[string]struct{}) string {
	switch typed := node.(type) {
	case map[string]any:
		for key, value := range typed {
			if _, exists := keys[strings.ToLower(strings.TrimSpace(key))]; exists {
				if valueString, ok := value.(string); ok {
					trimmed := strings.TrimSpace(valueString)
					if trimmed != "" {
						return trimmed
					}
				}
			}
		}
		for _, value := range typed {
			if nested := findFirstStringValueRecursive(value, keys); nested != "" {
				return nested
			}
		}
	case []any:
		for _, value := range typed {
			if nested := findFirstStringValueRecursive(value, keys); nested != "" {
				return nested
			}
		}
	}
	return ""
}

func buildSessionReferences(sessionDirectory string, sessionID string) map[string]any {
	return map[string]any{
		"workspace_yaml":    describeReferenceFile(filepath.Join(sessionDirectory, "workspace.yaml"), filepath.ToSlash(filepath.Join("session-state", sessionID, "workspace.yaml"))),
		"plan_md":           describeReferenceFile(filepath.Join(sessionDirectory, "plan.md"), filepath.ToSlash(filepath.Join("session-state", sessionID, "plan.md"))),
		"checkpoints_index": describeReferenceFile(filepath.Join(sessionDirectory, "checkpoints", "index.md"), filepath.ToSlash(filepath.Join("session-state", sessionID, "checkpoints", "index.md"))),
	}
}

func describeReferenceFile(path string, internalLocation string) map[string]any {
	reference := map[string]any{
		"internal_path": strings.TrimSpace(internalLocation),
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		reference["exists"] = false
		return reference
	}
	reference["exists"] = true
	reference["size_bytes"] = fileInfo.Size()
	reference["updated_at"] = fileInfo.ModTime().UTC().Format(time.RFC3339Nano)
	return reference
}

func buildSessionFileEventID(sessionID string, offset int64, line string) string {
	hash := sha1.Sum([]byte(line)) // #nosec G401 -- deterministic non-crypto hash for event identity
	return fmt.Sprintf("session-file-%s-%d-%s", strings.TrimSpace(sessionID), offset, hex.EncodeToString(hash[:8]))
}

func buildSessionHeartbeatEventID(sessionID string, layer string, window int64) string {
	return fmt.Sprintf("session-heartbeat-%s-%s-%d", strings.TrimSpace(sessionID), strings.TrimSpace(layer), window)
}

func freshnessStatus(value time.Duration, freshThreshold time.Duration) string {
	if freshThreshold <= 0 {
		freshThreshold = 1 * time.Second
	}
	if value <= freshThreshold {
		return "fresh"
	}
	return "stale"
}

func quorumStateFromFreshness(value time.Duration, freshThreshold time.Duration) string {
	if freshnessStatus(value, freshThreshold) == "fresh" {
		return "running_confident"
	}
	return "running_degraded"
}

func latestReferenceUpdate(references map[string]any) time.Time {
	latest := time.Time{}
	for _, key := range []string{"workspace_yaml", "plan_md", "checkpoints_index"} {
		raw, exists := references[key]
		if !exists || raw == nil {
			continue
		}
		reference, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		updatedAtRaw, _ := reference["updated_at"].(string)
		updatedAt := strings.TrimSpace(updatedAtRaw)
		if updatedAt == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			continue
		}
		if latest.IsZero() || parsed.After(latest) {
			latest = parsed.UTC()
		}
	}
	if latest.IsZero() {
		return time.Now().UTC()
	}
	return latest
}

func stringValue(source map[string]any, key string) string {
	raw, exists := source[key]
	if !exists || raw == nil {
		return ""
	}
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return value
}
