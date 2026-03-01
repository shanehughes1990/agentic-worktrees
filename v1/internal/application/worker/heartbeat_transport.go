package worker

import (
	"context"
	"time"
)

type HeartbeatRequest struct {
	RequestID   string    `json:"request_id"`
	WorkerID    string    `json:"worker_id"`
	Epoch       int64     `json:"epoch"`
	RequestedAt time.Time `json:"requested_at"`
	DeadlineAt  time.Time `json:"deadline_at"`
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

type HeartbeatTransport interface {
	PublishRequest(ctx context.Context, request HeartbeatRequest) error
	PublishResponse(ctx context.Context, response HeartbeatResponse) error
	ListenRequests(ctx context.Context, handler func(HeartbeatRequest) error) error
	ListenResponses(ctx context.Context, handler func(HeartbeatResponse) error) error
}
