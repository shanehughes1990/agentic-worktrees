package taskengine

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type JobKind string

const (
	JobKindIngestionAgent        JobKind = "ingestion.agent.run"
	JobKindAgentWorkflow         JobKind = "agent.workflow.run"
	JobKindSCMWorkflow           JobKind = "scm.workflow.run"
	JobKindProjectDocumentUploadPrepare JobKind = "project.document.upload.prepare"
	JobKindProjectDocumentDelete        JobKind = "project.document.delete"
)

type CorrelationIDs struct {
	RunID     string
	TaskID    string
	JobID     string
	ProjectID string
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

type Lease struct {
	OwnerID    string
	Token      string
	AcquiredAt time.Time
	ExpiresAt  time.Time
}

func (lease Lease) Validate() error {
	if strings.TrimSpace(lease.OwnerID) == "" {
		return fmt.Errorf("%w: owner_id is required", ErrInvalidLeaseContract)
	}
	if strings.TrimSpace(lease.Token) == "" {
		return fmt.Errorf("%w: token is required", ErrInvalidLeaseContract)
	}
	if lease.AcquiredAt.IsZero() {
		return fmt.Errorf("%w: acquired_at is required", ErrInvalidLeaseContract)
	}
	if lease.ExpiresAt.IsZero() {
		return fmt.Errorf("%w: expires_at is required", ErrInvalidLeaseContract)
	}
	if !lease.ExpiresAt.After(lease.AcquiredAt) {
		return fmt.Errorf("%w: expires_at must be after acquired_at", ErrInvalidLeaseContract)
	}
	return nil
}

type LeaseRenewalRequest struct {
	OwnerID   string
	RenewedAt time.Time
	ExpiresAt time.Time
}

func (lease Lease) Renew(request LeaseRenewalRequest) (Lease, error) {
	if err := lease.Validate(); err != nil {
		return Lease{}, err
	}
	if strings.TrimSpace(request.OwnerID) == "" {
		return Lease{}, fmt.Errorf("%w: owner_id is required", ErrInvalidLeaseRenewal)
	}
	if request.OwnerID != lease.OwnerID {
		return Lease{}, fmt.Errorf("%w: owner_id does not own lease", ErrInvalidLeaseRenewal)
	}
	if request.RenewedAt.IsZero() {
		return Lease{}, fmt.Errorf("%w: renewed_at is required", ErrInvalidLeaseRenewal)
	}
	if request.ExpiresAt.IsZero() {
		return Lease{}, fmt.Errorf("%w: expires_at is required", ErrInvalidLeaseRenewal)
	}
	if request.RenewedAt.After(lease.ExpiresAt) {
		return Lease{}, fmt.Errorf("%w: lease is already expired", ErrInvalidLeaseRenewal)
	}
	if !request.ExpiresAt.After(request.RenewedAt) {
		return Lease{}, fmt.Errorf("%w: expires_at must be after renewed_at", ErrInvalidLeaseRenewal)
	}
	if !request.ExpiresAt.After(lease.ExpiresAt) {
		return Lease{}, fmt.Errorf("%w: renewal must extend lease expiration", ErrInvalidLeaseRenewal)
	}
	renewedLease := lease
	renewedLease.ExpiresAt = request.ExpiresAt
	return renewedLease, nil
}
