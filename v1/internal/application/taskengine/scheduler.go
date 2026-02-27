package taskengine

import (
	"context"
	"fmt"
	"strings"
)

type Scheduler struct {
	engine   Engine
	policies map[JobKind]JobPolicy
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

	return scheduler.engine.Enqueue(ctx, normalizedRequest)
}
