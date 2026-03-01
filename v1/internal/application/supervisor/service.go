package supervisor

import (
	"agentic-orchestrator/internal/application/taskengine"
	domainsupervisor "agentic-orchestrator/internal/domain/supervisor"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrEventStoreRequired = errors.New("supervisor: event store is required")

type EventStore interface {
	Append(ctx context.Context, decision domainsupervisor.Decision) error
	ListByCorrelation(ctx context.Context, correlation domainsupervisor.CorrelationIDs) ([]domainsupervisor.Decision, error)
}

type Service struct {
	eventStore EventStore
	rules      []domainsupervisor.Rule
}

func NewService(eventStore EventStore, rules []domainsupervisor.Rule) (*Service, error) {
	if eventStore == nil {
		return nil, ErrEventStoreRequired
	}
	if len(rules) == 0 {
		rules = DefaultRules()
	}
	return &Service{eventStore: eventStore, rules: rules}, nil
}

func (service *Service) EvaluateSignal(ctx context.Context, signal domainsupervisor.Signal) (domainsupervisor.Decision, error) {
	if service == nil {
		return domainsupervisor.Decision{}, ErrEventStoreRequired
	}
	history, err := service.eventStore.ListByCorrelation(ctx, signal.CorrelationIDs)
	if err != nil {
		return domainsupervisor.Decision{}, fmt.Errorf("load supervisor history: %w", err)
	}
	currentState := domainsupervisor.StateIdle
	if len(history) > 0 {
		currentState = history[len(history)-1].ToState
	}
	decision, err := domainsupervisor.Evaluate(ctx, currentState, signal, service.rules)
	if err != nil {
		return domainsupervisor.Decision{}, err
	}
	if err := service.eventStore.Append(ctx, decision); err != nil {
		return domainsupervisor.Decision{}, fmt.Errorf("append supervisor decision: %w", err)
	}
	return decision, nil
}

func (service *Service) History(ctx context.Context, correlation domainsupervisor.CorrelationIDs) ([]domainsupervisor.Decision, error) {
	if service == nil {
		return nil, ErrEventStoreRequired
	}
	if err := correlation.Validate(); err != nil {
		return nil, err
	}
	return service.eventStore.ListByCorrelation(ctx, correlation)
}

func (service *Service) OnAdmitted(ctx context.Context, record taskengine.AdmissionRecord) error {
	_, err := service.OnAdmission(ctx, record)
	return err
}

func (service *Service) OnAdmission(ctx context.Context, record taskengine.AdmissionRecord) (domainsupervisor.Decision, error) {
	signal := domainsupervisor.Signal{
		Type:           domainsupervisor.SignalJobAdmitted,
		CorrelationIDs: domainsupervisor.CorrelationIDs{RunID: record.RunID, TaskID: record.TaskID, JobID: record.JobID},
		JobKind:        string(record.JobKind),
		IdempotencyKey: record.IdempotencyKey,
		OccurredAt:     record.EnqueuedAt,
		Metadata: map[string]string{
			"queue":         record.Queue,
			"queue_task_id": record.QueueTaskID,
		},
	}
	return service.EvaluateSignal(ctx, signal)
}

func (service *Service) OnExecution(ctx context.Context, record taskengine.ExecutionRecord, attempt int, maxRetry int) (domainsupervisor.Decision, error) {
	signalType := domainsupervisor.SignalExecutionProgressed
	switch record.Status {
	case taskengine.ExecutionStatusFailed:
		signalType = domainsupervisor.SignalExecutionFailed
	case taskengine.ExecutionStatusSucceeded:
		signalType = domainsupervisor.SignalExecutionSucceeded
	case taskengine.ExecutionStatusRunning, taskengine.ExecutionStatusSkipped:
		signalType = domainsupervisor.SignalExecutionProgressed
	}
	signal := domainsupervisor.Signal{
		Type:           signalType,
		CorrelationIDs: domainsupervisor.CorrelationIDs{RunID: record.RunID, TaskID: record.TaskID, JobID: record.JobID},
		JobKind:        string(record.JobKind),
		IdempotencyKey: record.IdempotencyKey,
		Attempt:        attempt,
		MaxRetry:       maxRetry,
		OccurredAt:     record.UpdatedAt,
		Metadata: map[string]string{
			"step":          record.Step,
			"status":        string(record.Status),
			"error_message": record.ErrorMessage,
		},
	}
	if signalType == domainsupervisor.SignalExecutionFailed {
		signal.FailureClass = classifyExecutionFailure(record)
	}
	return service.EvaluateSignal(ctx, signal)
}

func (service *Service) OnCheckpointSaved(ctx context.Context, correlation taskengine.CorrelationIDs, jobKind taskengine.JobKind, idempotencyKey string, step string) (domainsupervisor.Decision, error) {
	signal := domainsupervisor.Signal{
		Type:           domainsupervisor.SignalCheckpointSaved,
		CorrelationIDs: domainsupervisor.CorrelationIDs{RunID: correlation.RunID, TaskID: correlation.TaskID, JobID: correlation.JobID},
		JobKind:        string(jobKind),
		IdempotencyKey: idempotencyKey,
		OccurredAt:     time.Now().UTC(),
		Metadata: map[string]string{
			"step": step,
		},
	}
	return service.EvaluateSignal(ctx, signal)
}

func (service *Service) OnTrackerAttention(ctx context.Context, correlation taskengine.CorrelationIDs, reason string) (domainsupervisor.Decision, error) {
	signal := domainsupervisor.Signal{
		Type:           domainsupervisor.SignalTrackerAttentionNeeded,
		CorrelationIDs: domainsupervisor.CorrelationIDs{RunID: correlation.RunID, TaskID: correlation.TaskID, JobID: correlation.JobID},
		AttentionZone:  domainsupervisor.AttentionZoneTracker,
		OccurredAt:     time.Now().UTC(),
		Metadata:       map[string]string{"reason": reason},
	}
	return service.EvaluateSignal(ctx, signal)
}

func (service *Service) OnIssueOpened(ctx context.Context, correlation taskengine.CorrelationIDs, source string, issueReference string) (domainsupervisor.Decision, error) {
	signal := domainsupervisor.Signal{
		Type:           domainsupervisor.SignalIssueOpened,
		CorrelationIDs: domainsupervisor.CorrelationIDs{RunID: correlation.RunID, TaskID: correlation.TaskID, JobID: correlation.JobID},
		AttentionZone:  domainsupervisor.AttentionZoneTracker,
		OccurredAt:     time.Now().UTC(),
		Metadata: map[string]string{
			"source":          strings.TrimSpace(source),
			"issue_reference": strings.TrimSpace(issueReference),
		},
	}
	return service.EvaluateSignal(ctx, signal)
}

var _ taskengine.AdmissionSignalSink = (*Service)(nil)
