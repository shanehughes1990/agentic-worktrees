package controlplane

import (
	domainlifecycle "agentic-orchestrator/internal/domain/lifecycle"
	"context"
	"fmt"
	"strings"
	"time"
)

type LifecycleEventService interface {
	AppendEvent(ctx context.Context, event domainlifecycle.Event) (domainlifecycle.Event, error)
}

type ManualInterventionAction string

const (
	ManualInterventionActionNudge     ManualInterventionAction = "nudge"
	ManualInterventionActionRetry     ManualInterventionAction = "retry"
	ManualInterventionActionPause     ManualInterventionAction = "pause"
	ManualInterventionActionTerminate ManualInterventionAction = "terminate"
	ManualInterventionActionRestore   ManualInterventionAction = "restore"
)

type ManualInterventionRequest struct {
	ProjectID string
	SessionID string
	Action    ManualInterventionAction
	Reason    string
	ActorID   string
	Force     bool
}

type ManualInterventionResult struct {
	EventID         string
	EventSeq        int64
	ProjectEventSeq int64
	Action          string
	ResultingState  string
}

type InterventionMetrics struct {
	ProjectID              string
	InterventionCount      int
	SuccessfulOutcomeCount int
	FailedOutcomeCount     int
	AverageRecoverySeconds int64
}

func (service *Service) ApplyManualIntervention(ctx context.Context, request ManualInterventionRequest) (ManualInterventionResult, error) {
	if service == nil || service.lifecycleService == nil {
		return ManualInterventionResult{}, fmt.Errorf("lifecycle service is not configured")
	}
	request.ProjectID = strings.TrimSpace(request.ProjectID)
	request.SessionID = strings.TrimSpace(request.SessionID)
	request.ActorID = strings.TrimSpace(request.ActorID)
	request.Reason = strings.TrimSpace(request.Reason)
	request.Action = ManualInterventionAction(strings.ToLower(strings.TrimSpace(string(request.Action))))
	if request.ProjectID == "" || request.SessionID == "" {
		return ManualInterventionResult{}, fmt.Errorf("project_id and session_id are required")
	}
	if request.ActorID == "" {
		return ManualInterventionResult{}, fmt.Errorf("actor_id is required")
	}
	if request.Reason == "" {
		return ManualInterventionResult{}, fmt.Errorf("reason is required")
	}
	if !isSupportedManualInterventionAction(request.Action) {
		return ManualInterventionResult{}, fmt.Errorf("invalid manual intervention action %q", request.Action)
	}
	if isDestructiveManualInterventionAction(request.Action) {
		if !strings.HasPrefix(request.ActorID, "admin:") {
			return ManualInterventionResult{}, fmt.Errorf("forbidden: actor is not authorized for %s", request.Action)
		}
		if !request.Force {
			return ManualInterventionResult{}, fmt.Errorf("force confirmation is required for %s", request.Action)
		}
		if len(request.Reason) < 12 {
			return ManualInterventionResult{}, fmt.Errorf("reason must be at least 12 characters for %s", request.Action)
		}
	}

	snapshot, err := service.findLifecycleSessionSnapshot(ctx, request.ProjectID, request.SessionID)
	if err != nil {
		return ManualInterventionResult{}, err
	}
	if snapshot == nil {
		return ManualInterventionResult{}, fmt.Errorf("session not found")
	}

	if request.Action == ManualInterventionActionRetry {
		if err := service.requeueSessionWorkflowTask(ctx, *snapshot); err != nil {
			return ManualInterventionResult{}, err
		}
	}

	now := time.Now().UTC()
	resultingState := interventionResultState(request.Action)
	appended, err := service.lifecycleService.AppendEvent(ctx, domainlifecycle.Event{
		EventID:       fmt.Sprintf("manual:%s:%s:%d", request.Action, request.SessionID, now.UnixNano()),
		SchemaVersion: 1,
		ProjectID:     request.ProjectID,
		RunID:         snapshot.RunID,
		TaskID:        snapshot.TaskID,
		JobID:         snapshot.JobID,
		SessionID:     request.SessionID,
		SourceRuntime: "api_controlplane",
		PipelineType:  snapshot.PipelineType,
		EventType:     domainlifecycle.EventType("manual_" + string(request.Action)),
		OccurredAt:    now,
		Payload: map[string]any{
			"manual_action":     string(request.Action),
			"actor_id":          request.ActorID,
			"reason":            request.Reason,
			"force":             request.Force,
			"previous_state":    snapshot.CurrentState,
			"resulting_state":   resultingState,
			"source_runtime":    "api_controlplane",
			"intervention_time": now.Format(time.RFC3339),
		},
	})
	if err != nil {
		return ManualInterventionResult{}, fmt.Errorf("persist manual intervention: %w", err)
	}

	return ManualInterventionResult{
		EventID:         appended.EventID,
		EventSeq:        appended.EventSeq,
		ProjectEventSeq: appended.ProjectEventSeq,
		Action:          string(request.Action),
		ResultingState:  resultingState,
	}, nil
}

func (service *Service) requeueSessionWorkflowTask(ctx context.Context, snapshot LifecycleSessionSnapshot) error {
	if service == nil || service.queryRepository == nil {
		return fmt.Errorf("control-plane query repository is not configured")
	}
	if service.deadLetterManager == nil {
		return fmt.Errorf("dead-letter manager is not configured")
	}
	runID := strings.TrimSpace(snapshot.RunID)
	taskID := strings.TrimSpace(snapshot.TaskID)
	jobID := strings.TrimSpace(snapshot.JobID)
	if runID == "" || taskID == "" || jobID == "" {
		return fmt.Errorf("session correlation ids are incomplete for retry")
	}
	jobs, err := service.queryRepository.ListWorkflowJobs(ctx, runID, taskID, 500)
	if err != nil {
		return fmt.Errorf("load workflow jobs for retry: %w", err)
	}
	for _, job := range jobs {
		if strings.TrimSpace(job.JobID) != jobID {
			continue
		}
		queue := strings.TrimSpace(job.Queue)
		queueTaskID := strings.TrimSpace(job.QueueTaskID)
		if queue == "" || queueTaskID == "" {
			return fmt.Errorf("workflow job %q is missing queue identifiers for retry", jobID)
		}
		if err := service.deadLetterManager.RequeueDeadLetter(ctx, queue, queueTaskID); err != nil {
			return fmt.Errorf("requeue workflow job %q: %w", jobID, err)
		}
		return nil
	}
	return fmt.Errorf("workflow job %q not found for retry", jobID)
}

func (service *Service) findLifecycleSessionSnapshot(ctx context.Context, projectID string, sessionID string) (*LifecycleSessionSnapshot, error) {
	snapshots, err := service.LifecycleSessionSnapshots(ctx, projectID, "", 500)
	if err != nil {
		return nil, err
	}
	for _, snapshot := range snapshots {
		if strings.TrimSpace(snapshot.SessionID) == sessionID {
			copy := snapshot
			return &copy, nil
		}
	}
	return nil, nil
}

func isSupportedManualInterventionAction(action ManualInterventionAction) bool {
	switch action {
	case ManualInterventionActionNudge, ManualInterventionActionRetry, ManualInterventionActionPause, ManualInterventionActionTerminate, ManualInterventionActionRestore:
		return true
	default:
		return false
	}
}

func isDestructiveManualInterventionAction(action ManualInterventionAction) bool {
	return action == ManualInterventionActionPause || action == ManualInterventionActionTerminate
}

func interventionResultState(action ManualInterventionAction) string {
	switch action {
	case ManualInterventionActionNudge:
		return "stale_needs_nudge"
	case ManualInterventionActionRetry, ManualInterventionActionRestore:
		return "healthy_active"
	case ManualInterventionActionPause:
		return "healthy_waiting_input"
	case ManualInterventionActionTerminate:
		return "exited_unexpected"
	default:
		return "healthy_active"
	}
}

func (service *Service) InterventionMetrics(ctx context.Context, projectID string, limit int) (InterventionMetrics, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return InterventionMetrics{}, fmt.Errorf("project_id is required")
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	snapshots, err := service.LifecycleSessionSnapshots(ctx, projectID, "", limit)
	if err != nil {
		return InterventionMetrics{}, err
	}

	metrics := InterventionMetrics{ProjectID: projectID}
	var totalRecoverySeconds int64
	for _, snapshot := range snapshots {
		history, historyErr := service.LifecycleSessionHistory(ctx, projectID, snapshot.SessionID, 0, 500)
		if historyErr != nil {
			return InterventionMetrics{}, historyErr
		}
		for index, record := range history {
			eventType := strings.TrimSpace(record.EventType)
			if !strings.HasPrefix(eventType, "manual_") {
				continue
			}
			metrics.InterventionCount++
			next := nextOutcomeEvent(history[index+1:])
			if next == nil {
				continue
			}
			if isSuccessfulOutcome(next.EventType) {
				metrics.SuccessfulOutcomeCount++
				if next.OccurredAt.After(record.OccurredAt) {
					totalRecoverySeconds += int64(next.OccurredAt.Sub(record.OccurredAt).Seconds())
				}
			} else {
				metrics.FailedOutcomeCount++
			}
		}
	}
	if metrics.SuccessfulOutcomeCount > 0 {
		metrics.AverageRecoverySeconds = totalRecoverySeconds / int64(metrics.SuccessfulOutcomeCount)
	}
	return metrics, nil
}

func nextOutcomeEvent(events []LifecycleHistoryEvent) *LifecycleHistoryEvent {
	for _, event := range events {
		typeValue := strings.TrimSpace(event.EventType)
		if strings.HasPrefix(typeValue, "manual_") {
			continue
		}
		copy := event
		return &copy
	}
	return nil
}

func isSuccessfulOutcome(eventType string) bool {
	normalized := strings.ToLower(strings.TrimSpace(eventType))
	if strings.Contains(normalized, "failed") || strings.Contains(normalized, "exited") {
		return false
	}
	if strings.Contains(normalized, "completed") || strings.Contains(normalized, "recovered") || strings.Contains(normalized, "healthy") || strings.Contains(normalized, "started") {
		return true
	}
	return false
}
