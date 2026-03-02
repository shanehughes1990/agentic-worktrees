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
	StateDegraded          State = "degraded"
	StateStale             State = "stale"
	StateDraining          State = "draining"
	StateShutdownRequested State = "shutdown_requested"
	StateDeregistered      State = "deregistered"
	StateTerminated        State = "terminated"
)

func (state State) Validate() error {
	switch state {
	case StateRegistered, StateHealthy, StateDegraded, StateStale, StateDraining, StateShutdownRequested, StateDeregistered, StateTerminated:
		return nil
	default:
		return fmt.Errorf("unsupported worker state %q", state)
	}
}

type Worker struct {
	WorkerID       string
	Epoch          int64
	State          State
	DesiredState   State
	Capabilities   []taskengine.JobKind
	LastHeartbeat  time.Time
	LeaseExpiresAt time.Time
	RogueReason    string
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
	if err := worker.DesiredState.Validate(); err != nil {
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
	StaleAfter        time.Duration
	DrainTimeout      time.Duration
	TerminateTimeout  time.Duration
	RogueThreshold    int
	UpdatedAt         time.Time
}

func (settings Settings) Validate() error {
	if settings.HeartbeatInterval <= 0 {
		return fmt.Errorf("heartbeat_interval must be greater than zero")
	}
	if settings.ResponseDeadline <= 0 {
		return fmt.Errorf("response_deadline must be greater than zero")
	}
	if settings.StaleAfter <= settings.HeartbeatInterval {
		return fmt.Errorf("stale_after must be greater than heartbeat_interval")
	}
	if settings.DrainTimeout <= 0 {
		return fmt.Errorf("drain_timeout must be greater than zero")
	}
	if settings.TerminateTimeout <= 0 {
		return fmt.Errorf("terminate_timeout must be greater than zero")
	}
	if settings.RogueThreshold <= 0 {
		return fmt.Errorf("rogue_threshold must be greater than zero")
	}
	if settings.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}
	return nil
}
