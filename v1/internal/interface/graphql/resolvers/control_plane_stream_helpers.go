package resolvers

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	domainstream "agentic-orchestrator/internal/domain/stream"
	"agentic-orchestrator/internal/interface/graphql/models"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func isActiveRuntimeWorkerProjectEvent(event domainstream.Event) bool {
	if event.Source != domainstream.SourceWorker {
		return false
	}
	if isWorkerSessionLifecycleEvent(event.EventType) {
		return true
	}
	runtimeActivity, _ := event.Payload["runtime_activity"].(bool)
	runtimeEvent := strings.ToLower(strings.TrimSpace(payloadString(event.Payload, "runtime_event")))
	if runtimeEvent == "" {
		runtimeEvent = strings.ToLower(strings.TrimSpace(payloadString(event.Payload, "lifecycle_event_type")))
	}
	if !runtimeActivity && runtimeEvent == "" {
		return false
	}
	switch runtimeEvent {
	case "started", "heartbeat", "updated", "healthy", "active", "enqueued", "completed", "failed", "terminated", "cancelled", "canceled", "exited", "paused":
		return true
	default:
		return false
	}
}

func bootstrapProjectActiveEvents(ctx context.Context, controlPlaneService *applicationcontrolplane.Service, projectID string) []models.StreamEventResult {
	if controlPlaneService == nil || strings.TrimSpace(projectID) == "" {
		return nil
	}
	snapshots, err := controlPlaneService.LifecycleSessionSnapshots(ctx, projectID, "", 300)
	if err != nil {
		return nil
	}
	results := make([]models.StreamEventResult, 0, len(snapshots))
	for _, snapshot := range snapshots {
		if !isActiveLifecycleSnapshot(snapshot) {
			continue
		}
		results = append(results, models.StreamEventSuccess{Event: mapStreamEvent(streamEventFromLifecycleSnapshot(snapshot))})
	}
	return results
}

func bootstrapSessionActivityEvents(ctx context.Context, controlPlaneService *applicationcontrolplane.Service, correlation models.CorrelationInput) []models.StreamEventResult {
	if controlPlaneService == nil {
		return nil
	}
	projectID := strings.TrimSpace(derefString(correlation.ProjectID))
	if projectID == "" {
		return nil
	}
	snapshots, err := controlPlaneService.LifecycleSessionSnapshots(ctx, projectID, "", 300)
	if err != nil {
		return nil
	}
	results := make([]models.StreamEventResult, 0, 1)
	for _, snapshot := range snapshots {
		if !matchesLifecycleSnapshotCorrelation(snapshot, correlation) {
			continue
		}
		seedEvent := streamEventFromLifecycleSnapshot(snapshot)
		if strings.TrimSpace(snapshot.ProjectID) != "" && strings.TrimSpace(snapshot.SessionID) != "" {
			history, err := controlPlaneService.LifecycleSessionHistory(ctx, snapshot.ProjectID, snapshot.SessionID, maxInt64(snapshot.LastEventSeq-1, 0), 1)
			if err == nil && len(history) > 0 {
				latest := history[0]
				for _, candidate := range history {
					if candidate.EventSeq > latest.EventSeq {
						latest = candidate
					}
				}
				seedEvent = streamEventFromLifecycleHistory(latest)
			}
		}
		results = append(results, models.StreamEventSuccess{Event: mapStreamEvent(seedEvent)})
		break
	}
	return results
}

func streamEventFromLifecycleHistory(record applicationcontrolplane.LifecycleHistoryEvent) domainstream.Event {
	payload := map[string]any{}
	if strings.TrimSpace(record.PayloadJSON) != "" {
		_ = json.Unmarshal([]byte(record.PayloadJSON), &payload)
	}
	payload["lifecycle_event_type"] = strings.TrimSpace(record.EventType)
	payload["session_event_seq"] = record.EventSeq
	payload["project_event_seq"] = record.ProjectEventSeq

	return domainstream.Event{
		EventID:      strings.TrimSpace(record.EventID),
		StreamOffset: uint64(record.ProjectEventSeq),
		OccurredAt:   record.OccurredAt.UTC(),
		Source:       streamSourceFromLifecycleRuntime(record.SourceRuntime),
		EventType:    lifecycleHistoryEventTypeToStreamEventType(record.EventType),
		CorrelationIDs: domainstream.CorrelationIDs{
			RunID:         strings.TrimSpace(record.RunID),
			TaskID:        strings.TrimSpace(record.TaskID),
			JobID:         strings.TrimSpace(record.JobID),
			ProjectID:     strings.TrimSpace(record.ProjectID),
			SessionID:     strings.TrimSpace(record.SessionID),
			CorrelationID: "session:" + strings.TrimSpace(record.SessionID),
		},
		Payload: payload,
	}
}

func lifecycleHistoryEventTypeToStreamEventType(eventType string) domainstream.EventType {
	normalized := strings.TrimSpace(strings.ToLower(eventType))
	switch normalized {
	case "started":
		return domainstream.EventSessionStarted
	case "completed", "failed":
		return domainstream.EventSessionEnded
	case "heartbeat", "runtime_heartbeat", "process_heartbeat", "activity_heartbeat", "tool_heartbeat", "log_heartbeat", "heartbeat_quorum_degraded", "heartbeat_quorum_recovered", "gap_detected", "gap_reconciled":
		return domainstream.EventSessionHealth
	default:
		return domainstream.EventSessionUpdated
	}
}

func streamSourceFromLifecycleRuntime(runtime string) domainstream.Source {
	if strings.Contains(strings.ToLower(strings.TrimSpace(runtime)), "worker") {
		return domainstream.SourceWorker
	}
	return domainstream.SourceACP
}

func maxInt64(value int64, floor int64) int64 {
	if value < floor {
		return floor
	}
	return value
}

func matchesLifecycleSnapshotCorrelation(snapshot applicationcontrolplane.LifecycleSessionSnapshot, correlation models.CorrelationInput) bool {
	if value := strings.TrimSpace(derefString(correlation.RunID)); value != "" && strings.TrimSpace(snapshot.RunID) != value {
		return false
	}
	if value := strings.TrimSpace(derefString(correlation.TaskID)); value != "" && strings.TrimSpace(snapshot.TaskID) != value {
		return false
	}
	if value := strings.TrimSpace(derefString(correlation.JobID)); value != "" && strings.TrimSpace(snapshot.JobID) != value {
		return false
	}
	if value := strings.TrimSpace(derefString(correlation.ProjectID)); value != "" && strings.TrimSpace(snapshot.ProjectID) != value {
		return false
	}
	return true
}

func isActiveLifecycleSnapshot(snapshot applicationcontrolplane.LifecycleSessionSnapshot) bool {
	if snapshot.EndedAt != nil {
		return false
	}
	state := strings.ToLower(strings.TrimSpace(snapshot.CurrentState))
	if state == "" {
		return true
	}
	if strings.Contains(state, "completed") || strings.Contains(state, "exited") {
		return false
	}
	return true
}

func streamEventFromLifecycleSnapshot(snapshot applicationcontrolplane.LifecycleSessionSnapshot) domainstream.Event {
	timestamp := snapshot.UpdatedAt.UTC()
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}
	return domainstream.Event{
		EventID:      fmt.Sprintf("seed:%s:%d:%d", strings.TrimSpace(snapshot.SessionID), snapshot.LastProjectEventSeq, timestamp.UnixNano()),
		StreamOffset: uint64(snapshot.LastProjectEventSeq),
		OccurredAt:   timestamp,
		Source:       domainstream.SourceWorker,
		EventType:    domainstream.EventSessionHealth,
		CorrelationIDs: domainstream.CorrelationIDs{
			RunID:         strings.TrimSpace(snapshot.RunID),
			TaskID:        strings.TrimSpace(snapshot.TaskID),
			JobID:         strings.TrimSpace(snapshot.JobID),
			ProjectID:     strings.TrimSpace(snapshot.ProjectID),
			SessionID:     strings.TrimSpace(snapshot.SessionID),
			CorrelationID: "session:" + strings.TrimSpace(snapshot.SessionID),
		},
		Payload: map[string]any{
			"runtime_activity":     true,
			"runtime_event":        "health",
			"seeded_from_snapshot": true,
			"current_state":        strings.TrimSpace(snapshot.CurrentState),
			"current_severity":     strings.TrimSpace(snapshot.CurrentSeverity),
			"pipeline_type":        strings.TrimSpace(snapshot.PipelineType),
		},
	}
}

func isWorkerSessionLifecycleEvent(eventType domainstream.EventType) bool {
	switch eventType {
	case domainstream.EventSessionStarted, domainstream.EventSessionHealth, domainstream.EventSessionEnded:
		return true
	default:
		return false
	}
}

func payloadString(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	raw, ok := payload[key]
	if !ok || raw == nil {
		return ""
	}
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return value
}
