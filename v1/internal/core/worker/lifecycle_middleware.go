package worker

import (
	applicationlifecycle "agentic-orchestrator/internal/application/lifecycle"
	"agentic-orchestrator/internal/application/taskengine"
	"agentic-orchestrator/internal/domain/failures"
	domainlifecycle "agentic-orchestrator/internal/domain/lifecycle"
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const lifecycleSchemaVersion = 1

type jobLifecycleMiddleware struct {
	workerID string
	service  *applicationlifecycle.Service
	transport domainrealtime.WorkerLifecycleTransport
	handler  taskengine.Handler
}

func newJobLifecycleMiddleware(workerID string, service *applicationlifecycle.Service, transport domainrealtime.WorkerLifecycleTransport, handler taskengine.Handler) taskengine.Handler {
	if service == nil || handler == nil {
		return handler
	}
	return &jobLifecycleMiddleware{workerID: strings.TrimSpace(workerID), service: service, transport: transport, handler: handler}
}

func (middleware *jobLifecycleMiddleware) Handle(ctx context.Context, job taskengine.Job) error {
	if middleware == nil || middleware.service == nil || middleware.handler == nil {
		return fmt.Errorf("job lifecycle middleware is not initialized")
	}
	meta := extractLifecycleMetadata(job)
	startedID := lifecycleEventID(job, meta.IdempotencyKey, meta.JobID, string(domainlifecycle.EventStarted), meta.RetryCount)
	startedPayload := map[string]any{
		"queue_task_id": strings.TrimSpace(job.QueueTaskID),
		"job_kind":      string(job.Kind),
		"retry_count":   meta.RetryCount,
		"max_retry":     meta.MaxRetry,
	}
	if _, err := middleware.service.AppendEvent(ctx, domainlifecycle.Event{
		EventID:       startedID,
		SchemaVersion: lifecycleSchemaVersion,
		ProjectID:     meta.ProjectID,
		RunID:         meta.RunID,
		TaskID:        meta.TaskID,
		JobID:         meta.JobID,
		SessionID:     meta.SessionID,
		WorkerID:      middleware.workerID,
		SourceRuntime: "worker",
		PipelineType:  string(job.Kind),
		EventType:     domainlifecycle.EventStarted,
		OccurredAt:    time.Now().UTC(),
		Payload:       startedPayload,
	}); err != nil {
		return fmt.Errorf("append lifecycle started event: %w", err)
	}
	middleware.publishRuntimeSignal(ctx, meta, "started", startedPayload, startedID)
	heartbeatCtx, stopHeartbeats := context.WithCancel(ctx)
	defer stopHeartbeats()
	go middleware.runRuntimeHeartbeatSignals(heartbeatCtx, meta, job)

	handlerErr := middleware.handler.Handle(ctx, job)
	stopHeartbeats()
	completedType := domainlifecycle.EventCompleted
	if handlerErr != nil {
		completedType = domainlifecycle.EventFailed
	}
	payload := map[string]any{
		"queue_task_id": strings.TrimSpace(job.QueueTaskID),
		"job_kind":      string(job.Kind),
		"runtime_alive": handlerErr == nil,
		"retry_count":   meta.RetryCount,
		"max_retry":     meta.MaxRetry,
	}
	if handlerErr != nil {
		payload["error"] = strings.TrimSpace(handlerErr.Error())
		payload["failure_class"] = string(failures.ClassOf(handlerErr))
	}
	runtimeTerminalEventType := "completed"
	if completedType == domainlifecycle.EventFailed {
		runtimeTerminalEventType = "failed"
	}
	completedID := lifecycleEventID(job, meta.IdempotencyKey, meta.JobID, string(completedType), meta.RetryCount)
	if _, err := middleware.service.AppendEvent(ctx, domainlifecycle.Event{
		EventID:       completedID,
		SchemaVersion: lifecycleSchemaVersion,
		ProjectID:     meta.ProjectID,
		RunID:         meta.RunID,
		TaskID:        meta.TaskID,
		JobID:         meta.JobID,
		SessionID:     meta.SessionID,
		WorkerID:      middleware.workerID,
		SourceRuntime: "worker",
		PipelineType:  string(job.Kind),
		EventType:     completedType,
		OccurredAt:    time.Now().UTC(),
		Payload:       payload,
	}); err != nil {
		wrapped := fmt.Errorf("append lifecycle %s event: %w", strings.TrimSpace(string(completedType)), err)
		if handlerErr != nil {
			return errors.Join(handlerErr, wrapped)
		}
		return wrapped
	}
	middleware.publishRuntimeSignal(ctx, meta, runtimeTerminalEventType, payload, completedID)
	return handlerErr
}

type lifecycleMetadata struct {
	ProjectID      string
	RunID          string
	TaskID         string
	JobID          string
	SessionID      string
	PipelineType   string
	IdempotencyKey string
	RetryCount     int
	MaxRetry       int
}

func extractLifecycleMetadata(job taskengine.Job) lifecycleMetadata {
	metadata := lifecycleMetadata{}
	payload := map[string]any{}
	if err := json.Unmarshal(job.Payload, &payload); err == nil {
		metadata.ProjectID = readString(payload, "project_id")
		metadata.RunID = readString(payload, "run_id")
		metadata.TaskID = readString(payload, "task_id")
		metadata.JobID = readString(payload, "job_id")
		metadata.SessionID = readString(payload, "session_id")
		metadata.IdempotencyKey = readString(payload, "idempotency_key")
	}
	if strings.TrimSpace(metadata.JobID) == "" {
		metadata.JobID = strings.TrimSpace(job.QueueTaskID)
	}
	if strings.TrimSpace(metadata.ProjectID) == "" {
		metadata.ProjectID = "project:unknown"
	}
	if strings.TrimSpace(metadata.SessionID) == "" {
		metadata.SessionID = buildSyntheticSessionID(job.Kind, metadata.ProjectID, metadata.RunID, metadata.TaskID, metadata.JobID, metadata.IdempotencyKey)
	}
	metadata.PipelineType = strings.TrimSpace(string(job.Kind))
	if job.RetryCount < 0 {
		metadata.RetryCount = 0
	} else {
		metadata.RetryCount = job.RetryCount
	}
	if job.MaxRetry < 0 {
		metadata.MaxRetry = 0
	} else {
		metadata.MaxRetry = job.MaxRetry
	}
	return metadata
}

func buildSyntheticSessionID(kind taskengine.JobKind, projectID string, runID string, taskID string, jobID string, idempotencyKey string) string {
	parts := []string{strings.TrimSpace(string(kind)), strings.TrimSpace(projectID), strings.TrimSpace(runID), strings.TrimSpace(taskID), strings.TrimSpace(jobID), strings.TrimSpace(idempotencyKey)}
	joined := strings.Join(parts, ":")
	hash := sha1.Sum([]byte(joined))
	return fmt.Sprintf("synthetic:%s:%s", strings.TrimSpace(string(kind)), hex.EncodeToString(hash[:8]))
}

func lifecycleEventID(job taskengine.Job, idempotencyKey string, jobID string, phase string, retryCount int) string {
	base := strings.TrimSpace(idempotencyKey)
	if base == "" {
		base = strings.TrimSpace(jobID)
	}
	if base == "" {
		base = strings.TrimSpace(job.QueueTaskID)
	}
	if base == "" {
		hash := sha1.Sum(job.Payload)
		base = hex.EncodeToString(hash[:8])
	}
	if retryCount < 0 {
		retryCount = 0
	}
	return fmt.Sprintf("lifecycle:%s:%s:%s:attempt-%d", strings.TrimSpace(string(job.Kind)), base, strings.TrimSpace(phase), retryCount)
}

func readString(payload map[string]any, key string) string {
	value, ok := payload[key]
	if !ok || value == nil {
		return ""
	}
	stringValue, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(stringValue)
}

func (middleware *jobLifecycleMiddleware) runRuntimeHeartbeatSignals(ctx context.Context, meta lifecycleMetadata, job taskengine.Job) {
	if middleware == nil || middleware.transport == nil {
		return
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			middleware.publishRuntimeSignal(ctx, meta, "heartbeat", map[string]any{
				"queue_task_id": strings.TrimSpace(job.QueueTaskID),
				"job_kind":      string(job.Kind),
				"retry_count":   meta.RetryCount,
				"max_retry":     meta.MaxRetry,
			}, "")
		}
	}
}

func (middleware *jobLifecycleMiddleware) publishRuntimeSignal(ctx context.Context, meta lifecycleMetadata, eventType string, payload map[string]any, signalID string) {
	if middleware == nil {
		return
	}
	now := time.Now().UTC()
	resolvedSignalID := strings.TrimSpace(signalID)
	if resolvedSignalID == "" {
		resolvedSignalID = domainrealtime.RuntimeActivitySignalID(meta.PipelineType, meta.SessionID, eventType, now)
	}

	if middleware.transport != nil {
		_ = middleware.transport.PublishRuntimeActivity(ctx, domainrealtime.RuntimeActivitySignal{
			SignalID:     resolvedSignalID,
			ProjectID:    strings.TrimSpace(meta.ProjectID),
			RunID:        strings.TrimSpace(meta.RunID),
			TaskID:       strings.TrimSpace(meta.TaskID),
			JobID:        strings.TrimSpace(meta.JobID),
			SessionID:    strings.TrimSpace(meta.SessionID),
			PipelineType: strings.TrimSpace(meta.PipelineType),
			WorkerID:     strings.TrimSpace(middleware.workerID),
			EventType:    strings.TrimSpace(eventType),
			OccurredAt:   now,
			Payload:      payload,
		})
	}

	if strings.TrimSpace(strings.ToLower(eventType)) != "heartbeat" || middleware.service == nil {
		return
	}
	if _, err := middleware.service.AppendEvent(ctx, domainlifecycle.Event{
		EventID:       resolvedSignalID,
		SchemaVersion: lifecycleSchemaVersion,
		ProjectID:     strings.TrimSpace(meta.ProjectID),
		RunID:         strings.TrimSpace(meta.RunID),
		TaskID:        strings.TrimSpace(meta.TaskID),
		JobID:         strings.TrimSpace(meta.JobID),
		SessionID:     strings.TrimSpace(meta.SessionID),
		WorkerID:      strings.TrimSpace(middleware.workerID),
		SourceRuntime: "worker",
		PipelineType:  strings.TrimSpace(meta.PipelineType),
		EventType:     domainlifecycle.EventType("heartbeat"),
		OccurredAt:    now,
		Payload:       payload,
	}); err != nil {
		return
	}
}
