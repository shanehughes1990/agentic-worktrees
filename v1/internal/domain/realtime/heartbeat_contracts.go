package realtime

import (
	"context"
	"errors"
	"strings"
	"time"

	"agentic-orchestrator/internal/domain/failures"
)

type HeartbeatRequest struct {
	RequestID   string    `json:"request_id"`
	WorkerID    string    `json:"worker_id"`
	Epoch       int64     `json:"epoch"`
	RequestedAt time.Time `json:"requested_at"`
	DeadlineAt  time.Time `json:"deadline_at"`
}

func (request HeartbeatRequest) Validate() error {
	if strings.TrimSpace(request.RequestID) == "" {
		return failures.WrapTerminal(errors.New("request_id is required"))
	}
	if strings.TrimSpace(request.WorkerID) == "" {
		return failures.WrapTerminal(errors.New("worker_id is required"))
	}
	if request.Epoch <= 0 {
		return failures.WrapTerminal(errors.New("epoch must be greater than zero"))
	}
	if request.RequestedAt.IsZero() {
		return failures.WrapTerminal(errors.New("requested_at is required"))
	}
	if request.DeadlineAt.IsZero() {
		return failures.WrapTerminal(errors.New("deadline_at is required"))
	}
	if !request.DeadlineAt.After(request.RequestedAt) {
		return failures.WrapTerminal(errors.New("deadline_at must be after requested_at"))
	}
	return nil
}

type HeartbeatResponse struct {
	RequestID   string    `json:"request_id"`
	WorkerID    string    `json:"worker_id"`
	Epoch       int64     `json:"epoch"`
	ReceivedAt  time.Time `json:"received_at"`
	RespondedAt time.Time `json:"responded_at"`
	Healthy     bool      `json:"healthy"`
	Reason      string    `json:"reason,omitempty"`
}

func (response HeartbeatResponse) Validate() error {
	if strings.TrimSpace(response.RequestID) == "" {
		return failures.WrapTerminal(errors.New("request_id is required"))
	}
	if strings.TrimSpace(response.WorkerID) == "" {
		return failures.WrapTerminal(errors.New("worker_id is required"))
	}
	if response.Epoch <= 0 {
		return failures.WrapTerminal(errors.New("epoch must be greater than zero"))
	}
	if response.ReceivedAt.IsZero() {
		return failures.WrapTerminal(errors.New("received_at is required"))
	}
	if response.RespondedAt.IsZero() {
		return failures.WrapTerminal(errors.New("responded_at is required"))
	}
	if response.RespondedAt.Before(response.ReceivedAt) {
		return failures.WrapTerminal(errors.New("responded_at must be after or equal to received_at"))
	}
	return nil
}

type HeartbeatTransport interface {
	PublishRequest(ctx context.Context, request HeartbeatRequest) error
	PublishResponse(ctx context.Context, response HeartbeatResponse) error
	ListenRequests(ctx context.Context, handler func(HeartbeatRequest) error) error
	ListenResponses(ctx context.Context, handler func(HeartbeatResponse) error) error
}
