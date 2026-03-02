package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"agentic-orchestrator/internal/application/taskengine"
	domainworker "agentic-orchestrator/internal/domain/worker"
)

var (
	ErrRepositoryRequired = errors.New("worker repository is required")
	ErrEpochMismatch      = errors.New("worker epoch mismatch")
	ErrApplicationStopping = errors.New("application stopping")
	ErrWorkerNotRegistered = errors.New("worker not registered")
	ErrSettingsNotFound   = errors.New("worker settings not found")
)

type Repository interface {
	Register(ctx context.Context, workerID string, capabilities []taskengine.JobKind, heartbeatAt time.Time, leaseExpiresAt time.Time) (*domainworker.Worker, error)
	RenewHeartbeat(ctx context.Context, workerID string, epoch int64, heartbeatAt time.Time, leaseExpiresAt time.Time) (*domainworker.Worker, error)
	UpdateState(ctx context.Context, workerID string, epoch int64, state domainworker.State, desiredState domainworker.State, reason string, changedAt time.Time) (*domainworker.Worker, error)
	RemoveRegistration(ctx context.Context, workerID string, epoch int64) error
	ListWorkers(ctx context.Context, limit int) ([]domainworker.Worker, error)
	ListStaleWorkers(ctx context.Context, staleBefore time.Time, limit int) ([]domainworker.Worker, error)
	GetSettings(ctx context.Context) (domainworker.Settings, error)
	UpsertSettings(ctx context.Context, settings domainworker.Settings) (domainworker.Settings, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) (*Service, error) {
	if repository == nil {
		return nil, ErrRepositoryRequired
	}
	return &Service{repository: repository}, nil
}

func DefaultSettings(now time.Time) domainworker.Settings {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return domainworker.Settings{
		HeartbeatInterval: 15 * time.Second,
		ResponseDeadline:  5 * time.Second,
		StaleAfter:        45 * time.Second,
		DrainTimeout:      20 * time.Second,
		TerminateTimeout:  10 * time.Second,
		RogueThreshold:    3,
		UpdatedAt:         now.UTC(),
	}
}

func (service *Service) Register(ctx context.Context, workerID string, capabilities []taskengine.JobKind, heartbeatInterval time.Duration) (*domainworker.Worker, error) {
	if service == nil || service.repository == nil {
		return nil, ErrRepositoryRequired
	}
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return nil, fmt.Errorf("worker_id is required")
	}
	if len(capabilities) == 0 {
		return nil, fmt.Errorf("at least one capability is required")
	}
	if heartbeatInterval <= 0 {
		return nil, fmt.Errorf("heartbeat_interval must be greater than zero")
	}
	now := time.Now().UTC()
	return service.repository.Register(ctx, workerID, capabilities, now, now.Add(heartbeatInterval*3))
}

func (service *Service) Heartbeat(ctx context.Context, workerID string, epoch int64, heartbeatInterval time.Duration) (*domainworker.Worker, error) {
	if service == nil || service.repository == nil {
		return nil, ErrRepositoryRequired
	}
	now := time.Now().UTC()
	worker, err := service.repository.RenewHeartbeat(ctx, strings.TrimSpace(workerID), epoch, now, now.Add(heartbeatInterval*3))
	if err != nil {
		if errors.Is(err, ErrEpochMismatch) || errors.Is(err, ErrWorkerNotRegistered) {
			return nil, ErrApplicationStopping
		}
		return nil, err
	}
	if worker.DesiredState == domainworker.StateShutdownRequested || worker.DesiredState == domainworker.StateDraining || worker.DesiredState == domainworker.StateTerminated || worker.DesiredState == domainworker.StateDeregistered {
		return worker, ErrApplicationStopping
	}
	if worker.State == domainworker.StateDeregistered {
		return worker, ErrApplicationStopping
	}
	return worker, nil
}

func (service *Service) RequestShutdown(ctx context.Context, workerID string, epoch int64, reason string) (*domainworker.Worker, error) {
	if service == nil || service.repository == nil {
		return nil, ErrRepositoryRequired
	}
	return service.repository.UpdateState(ctx, strings.TrimSpace(workerID), epoch, domainworker.StateDraining, domainworker.StateShutdownRequested, strings.TrimSpace(reason), time.Now().UTC())
}

func (service *Service) ForceDeregister(ctx context.Context, workerID string, epoch int64, reason string) (*domainworker.Worker, error) {
	if service == nil || service.repository == nil {
		return nil, ErrRepositoryRequired
	}
	worker, err := service.repository.UpdateState(ctx, strings.TrimSpace(workerID), epoch, domainworker.StateTerminated, domainworker.StateDeregistered, strings.TrimSpace(reason), time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if err := service.repository.RemoveRegistration(ctx, strings.TrimSpace(workerID), epoch); err != nil {
		return nil, err
	}
	return worker, nil
}

func (service *Service) Deregister(ctx context.Context, workerID string, epoch int64, reason string) (*domainworker.Worker, error) {
	if service == nil || service.repository == nil {
		return nil, ErrRepositoryRequired
	}
	worker, err := service.repository.UpdateState(ctx, strings.TrimSpace(workerID), epoch, domainworker.StateDeregistered, domainworker.StateDeregistered, strings.TrimSpace(reason), time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if err := service.repository.RemoveRegistration(ctx, strings.TrimSpace(workerID), epoch); err != nil {
		return nil, err
	}
	return worker, nil
}

func (service *Service) ListWorkers(ctx context.Context, limit int) ([]domainworker.Worker, error) {
	if service == nil || service.repository == nil {
		return nil, ErrRepositoryRequired
	}
	if limit <= 0 {
		limit = 50
	}
	return service.repository.ListWorkers(ctx, limit)
}

func (service *Service) GetSettings(ctx context.Context) (domainworker.Settings, error) {
	if service == nil || service.repository == nil {
		return domainworker.Settings{}, ErrRepositoryRequired
	}
	return service.repository.GetSettings(ctx)
}

func (service *Service) UpdateSettings(ctx context.Context, settings domainworker.Settings) (domainworker.Settings, error) {
	if service == nil || service.repository == nil {
		return domainworker.Settings{}, ErrRepositoryRequired
	}
	if settings.UpdatedAt.IsZero() {
		settings.UpdatedAt = time.Now().UTC()
	}
	if err := settings.Validate(); err != nil {
		return domainworker.Settings{}, err
	}
	return service.repository.UpsertSettings(ctx, settings)
}

func (service *Service) EnsureBaseSettings(ctx context.Context, defaults domainworker.Settings) (domainworker.Settings, error) {
	if service == nil || service.repository == nil {
		return domainworker.Settings{}, ErrRepositoryRequired
	}
	settings, err := service.repository.GetSettings(ctx)
	if err == nil {
		return settings, nil
	}
	if !errors.Is(err, ErrSettingsNotFound) {
		return domainworker.Settings{}, err
	}
	if defaults.UpdatedAt.IsZero() {
		defaults.UpdatedAt = time.Now().UTC()
	}
	if validateErr := defaults.Validate(); validateErr != nil {
		return domainworker.Settings{}, validateErr
	}
	return service.repository.UpsertSettings(ctx, defaults)
}
