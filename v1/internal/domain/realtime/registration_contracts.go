package realtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"agentic-orchestrator/internal/domain/failures"
)

type RegistrationStatus string

const (
	RegistrationStatusPending  RegistrationStatus = "pending"
	RegistrationStatusAccepted RegistrationStatus = "accepted"
	RegistrationStatusRejected RegistrationStatus = "rejected"
	RegistrationStatusRevoked  RegistrationStatus = "revoked"
)

type RegistrationSubmission struct {
	SubmissionID  string
	WorkerID      string
	RequestedAt   time.Time
	ExpiresAt     time.Time
	Status        RegistrationStatus
	Capabilities  []Capability
	RejectReasons []string
	ResolvedAt    time.Time
}

func (status RegistrationStatus) Validate() error {
	switch status {
	case RegistrationStatusPending, RegistrationStatusAccepted, RegistrationStatusRejected, RegistrationStatusRevoked:
		return nil
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported registration status %q", status))
	}
}

func (submission RegistrationSubmission) Validate() error {
	if strings.TrimSpace(submission.SubmissionID) == "" {
		return failures.WrapTerminal(errors.New("submission_id is required"))
	}
	if strings.TrimSpace(submission.WorkerID) == "" {
		return failures.WrapTerminal(errors.New("worker_id is required"))
	}
	if submission.RequestedAt.IsZero() {
		return failures.WrapTerminal(errors.New("requested_at is required"))
	}
	if submission.ExpiresAt.IsZero() {
		return failures.WrapTerminal(errors.New("expires_at is required"))
	}
	if !submission.ExpiresAt.After(submission.RequestedAt) {
		return failures.WrapTerminal(errors.New("expires_at must be after requested_at"))
	}
	if err := submission.Status.Validate(); err != nil {
		return err
	}
	if len(submission.Capabilities) == 0 {
		return failures.WrapTerminal(errors.New("at least one capability is required"))
	}
	for _, capability := range submission.Capabilities {
		if err := capability.Validate(); err != nil {
			return err
		}
	}
	if submission.Status != RegistrationStatusPending && submission.ResolvedAt.IsZero() {
		return failures.WrapTerminal(errors.New("resolved_at is required when status is resolved"))
	}
	return nil
}

type RegistrationSubmissionEvent struct {
	SubmissionID string       `json:"submission_id"`
	WorkerID     string       `json:"worker_id"`
	RequestedAt  time.Time    `json:"requested_at"`
	ExpiresAt    time.Time    `json:"expires_at"`
	Capabilities []Capability `json:"capabilities"`
}

func (event RegistrationSubmissionEvent) Validate() error {
	if strings.TrimSpace(event.SubmissionID) == "" {
		return failures.WrapTerminal(errors.New("submission_id is required"))
	}
	if strings.TrimSpace(event.WorkerID) == "" {
		return failures.WrapTerminal(errors.New("worker_id is required"))
	}
	if event.RequestedAt.IsZero() {
		return failures.WrapTerminal(errors.New("requested_at is required"))
	}
	if event.ExpiresAt.IsZero() {
		return failures.WrapTerminal(errors.New("expires_at is required"))
	}
	if !event.ExpiresAt.After(event.RequestedAt) {
		return failures.WrapTerminal(errors.New("expires_at must be after requested_at"))
	}
	if len(event.Capabilities) == 0 {
		return failures.WrapTerminal(errors.New("at least one capability is required"))
	}
	for _, capability := range event.Capabilities {
		if err := capability.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type RegistrationDecision string

const (
	RegistrationDecisionAccept RegistrationDecision = "ok"
	RegistrationDecisionReject RegistrationDecision = "reject"
)

type RegistrationDecisionEvent struct {
	SubmissionID   string               `json:"submission_id"`
	WorkerID       string               `json:"worker_id"`
	Decision       RegistrationDecision `json:"decision"`
	Reasons        []string             `json:"reasons,omitempty"`
	RespondedAt    time.Time            `json:"responded_at"`
	RegistrationID int64                `json:"registration_epoch,omitempty"`
}

func (event RegistrationDecisionEvent) Validate() error {
	if strings.TrimSpace(event.SubmissionID) == "" {
		return failures.WrapTerminal(errors.New("submission_id is required"))
	}
	if strings.TrimSpace(event.WorkerID) == "" {
		return failures.WrapTerminal(errors.New("worker_id is required"))
	}
	switch event.Decision {
	case RegistrationDecisionAccept, RegistrationDecisionReject:
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported registration decision %q", event.Decision))
	}
	if event.RespondedAt.IsZero() {
		return failures.WrapTerminal(errors.New("responded_at is required"))
	}
	return nil
}

type InvalidationIntent struct {
	WorkerID string    `json:"worker_id"`
	Epoch    int64     `json:"epoch"`
	Reason   string    `json:"reason"`
	IssuedAt time.Time `json:"issued_at"`
}

func (intent InvalidationIntent) Validate() error {
	if strings.TrimSpace(intent.WorkerID) == "" {
		return failures.WrapTerminal(errors.New("worker_id is required"))
	}
	if intent.Epoch <= 0 {
		return failures.WrapTerminal(errors.New("epoch must be greater than zero"))
	}
	if strings.TrimSpace(intent.Reason) == "" {
		return failures.WrapTerminal(errors.New("reason is required"))
	}
	if intent.IssuedAt.IsZero() {
		return failures.WrapTerminal(errors.New("issued_at is required"))
	}
	return nil
}

type WorkerLifecycleTransport interface {
	HeartbeatTransport
	RuntimeActivityTransport
	PublishRegistrationSubmission(ctx context.Context, event RegistrationSubmissionEvent) error
	PublishRegistrationDecision(ctx context.Context, event RegistrationDecisionEvent) error
	ListenRegistrationSubmissions(ctx context.Context, handler func(RegistrationSubmissionEvent) error) error
	ListenRegistrationDecisions(ctx context.Context, handler func(RegistrationDecisionEvent) error) error
	PublishInvalidationIntent(ctx context.Context, intent InvalidationIntent) error
	ListenInvalidationIntents(ctx context.Context, handler func(InvalidationIntent) error) error
}
