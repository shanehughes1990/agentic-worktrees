package taskengine

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type AdmissionSignalSink interface {
	OnAdmitted(ctx context.Context, record AdmissionRecord) error
}

type Scheduler struct {
	engine              Engine
	policies            map[JobKind]JobPolicy
	admissionLedger     AdmissionLedger
	admissionSignalSink AdmissionSignalSink
}

func NewScheduler(engine Engine, policies map[JobKind]JobPolicy) (*Scheduler, error) {
	if engine == nil {
		return nil, fmt.Errorf("%w: engine is required", ErrInvalidEnqueueRequest)
	}
	if len(policies) == 0 {
		return nil, fmt.Errorf("%w: at least one job policy is required", ErrInvalidEnqueueRequest)
	}
	return &Scheduler{engine: engine, policies: policies}, nil
}

func (scheduler *Scheduler) SetAdmissionLedger(ledger AdmissionLedger) {
	if scheduler == nil {
		return
	}
	scheduler.admissionLedger = ledger
}

func (scheduler *Scheduler) SetAdmissionSignalSink(sink AdmissionSignalSink) {
	if scheduler == nil {
		return
	}
	scheduler.admissionSignalSink = sink
}

func (scheduler *Scheduler) Enqueue(ctx context.Context, request EnqueueRequest) (EnqueueResult, error) {
	if scheduler == nil || scheduler.engine == nil {
		return EnqueueResult{}, fmt.Errorf("%w: scheduler is not initialized", ErrInvalidEnqueueRequest)
	}
	policy, exists := scheduler.policies[request.Kind]
	if !exists {
		return EnqueueResult{}, fmt.Errorf("%w: unknown job kind %q", ErrInvalidEnqueueRequest, request.Kind)
	}
	if len(request.Payload) == 0 {
		return EnqueueResult{}, fmt.Errorf("%w: payload is required", ErrInvalidEnqueueRequest)
	}

	normalizedRequest := request
	if strings.TrimSpace(normalizedRequest.Queue) == "" {
		normalizedRequest.Queue = policy.DefaultQueue
	}
	if strings.TrimSpace(normalizedRequest.IdempotencyKey) == "" && policy.RequireIdempotencyKey {
		return EnqueueResult{}, fmt.Errorf("%w: idempotency_key is required for %q", ErrInvalidEnqueueRequest, request.Kind)
	}
	if normalizedRequest.UniqueFor <= 0 {
		normalizedRequest.UniqueFor = policy.DefaultUniqueFor
	}
	if policy.RequireUniqueFor && normalizedRequest.UniqueFor <= 0 {
		return EnqueueResult{}, fmt.Errorf("%w: unique_for is required for %q", ErrInvalidEnqueueRequest, request.Kind)
	}
	if normalizedRequest.Timeout <= 0 {
		normalizedRequest.Timeout = policy.DefaultTimeout
	}
	if normalizedRequest.Timeout <= 0 {
		return EnqueueResult{}, fmt.Errorf("%w: timeout must be greater than zero", ErrInvalidEnqueueRequest)
	}
	if normalizedRequest.MaxRetry < 0 {
		return EnqueueResult{}, fmt.Errorf("%w: max_retry cannot be negative", ErrInvalidEnqueueRequest)
	}
	if normalizedRequest.MaxRetry == 0 {
		normalizedRequest.MaxRetry = policy.DefaultMaxRetry
	}
	if strings.TrimSpace(normalizedRequest.Queue) == "" {
		return EnqueueResult{}, fmt.Errorf("%w: queue is required", ErrInvalidEnqueueRequest)
	}

	result, err := scheduler.engine.Enqueue(ctx, normalizedRequest)
	if err != nil {
		return EnqueueResult{}, err
	}
	record := AdmissionRecord{
		RunID:          strings.TrimSpace(normalizedRequest.CorrelationIDs.RunID),
		TaskID:         strings.TrimSpace(normalizedRequest.CorrelationIDs.TaskID),
		JobID:          strings.TrimSpace(normalizedRequest.CorrelationIDs.JobID),
		ProjectID:      strings.TrimSpace(normalizedRequest.CorrelationIDs.ProjectID),
		JobKind:        normalizedRequest.Kind,
		IdempotencyKey: strings.TrimSpace(normalizedRequest.IdempotencyKey),
		QueueTaskID:    strings.TrimSpace(result.QueueTaskID),
		Queue:          strings.TrimSpace(normalizedRequest.Queue),
		Status:         AdmissionStatusQueued,
		Duplicate:      result.Duplicate,
		EnqueuedAt:     time.Now().UTC(),
	}
	if scheduler.admissionLedger != nil {
		if err := scheduler.admissionLedger.Upsert(ctx, record); err != nil {
			return EnqueueResult{}, fmt.Errorf("persist admission ledger: %w", err)
		}
	}
	if scheduler.admissionSignalSink != nil {
		if err := scheduler.admissionSignalSink.OnAdmitted(ctx, record); err != nil {
			return EnqueueResult{}, fmt.Errorf("emit admission supervisor signal: %w", err)
		}
	}

	return result, nil
}
