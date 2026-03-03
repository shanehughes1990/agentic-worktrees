package realtime

import (
	"fmt"
	"strings"
	"time"

	"agentic-orchestrator/internal/application/taskengine"
)

type State string

const (
	StateRegistered        State = "registered"
	StateHealthy           State = "healthy"
	StateDeregistered      State = "deregistered"
)

func (state State) Validate() error {
	switch state {
	case StateRegistered, StateHealthy, StateDeregistered:
		return nil
	default:
		return fmt.Errorf("unsupported worker state %q", state)
	}
}

type Worker struct {
	WorkerID       string
	Epoch          int64
	State          State
	Capabilities   []taskengine.JobKind
	LastHeartbeat  time.Time
	LeaseExpiresAt time.Time
	UpdatedAt      time.Time
}

func (worker Worker) Validate() error {
	if strings.TrimSpace(worker.WorkerID) == "" {
		return fmt.Errorf("worker_id is required")
	}
	if worker.Epoch <= 0 {
		return fmt.Errorf("epoch must be greater than zero")
	}
	if err := worker.State.Validate(); err != nil {
		return err
	}
	if len(worker.Capabilities) == 0 {
		return fmt.Errorf("at least one capability is required")
	}
	if worker.LastHeartbeat.IsZero() {
		return fmt.Errorf("last_heartbeat is required")
	}
	if worker.LeaseExpiresAt.IsZero() {
		return fmt.Errorf("lease_expires_at is required")
	}
	if !worker.LeaseExpiresAt.After(worker.LastHeartbeat) {
		return fmt.Errorf("lease_expires_at must be after last_heartbeat")
	}
	if worker.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}
	return nil
}

type Settings struct {
	HeartbeatInterval time.Duration
	ResponseDeadline  time.Duration
	UpdatedAt         time.Time
}

func (settings Settings) Validate() error {
	if settings.HeartbeatInterval <= 0 {
		return fmt.Errorf("heartbeat_interval must be greater than zero")
	}
	if settings.ResponseDeadline <= 0 {
		return fmt.Errorf("response_deadline must be greater than zero")
	}
	if settings.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}
	return nil
}
