package taskengine

import (
	"context"
	"time"
)

type JobKind string

const (
	JobKindIngestionAgent JobKind = "ingestion.agent.run"
	JobKindSCMWorkflow    JobKind = "scm.workflow.run"
)

type CorrelationIDs struct {
	RunID  string
	TaskID string
	JobID  string
}

type EnqueueRequest struct {
	Kind           JobKind
	Payload        []byte
	Queue          string
	IdempotencyKey string
	UniqueFor      time.Duration
	Timeout        time.Duration
	MaxRetry       int
	CorrelationIDs CorrelationIDs
}

type EnqueueResult struct {
	QueueTaskID string
	Duplicate   bool
}

type JobPolicy struct {
	DefaultQueue          string
	RequireIdempotencyKey bool
	RequireUniqueFor      bool
	DefaultUniqueFor      time.Duration
	DefaultTimeout        time.Duration
	DefaultMaxRetry       int
}

type Job struct {
	Kind        JobKind
	QueueTaskID string
	Payload     []byte
}

type Handler interface {
	Handle(ctx context.Context, job Job) error
}

type HandlerFunc func(ctx context.Context, job Job) error

func (function HandlerFunc) Handle(ctx context.Context, job Job) error {
	return function(ctx, job)
}

type Engine interface {
	Enqueue(ctx context.Context, request EnqueueRequest) (EnqueueResult, error)
}

type Consumer interface {
	Register(kind JobKind, handler Handler) error
	Start() error
	Shutdown(ctx context.Context) error
}
