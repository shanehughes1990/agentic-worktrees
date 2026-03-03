package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"agentic-orchestrator/internal/application/taskengine"
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
)

var (
	ErrRepositoryRequired = errors.New("worker repository is required")
	ErrEpochMismatch      = errors.New("worker epoch mismatch")
	ErrSettingsNotFound   = errors.New("worker settings not found")
)

type Repository interface {
	Register(ctx context.Context, workerID string, capabilities []taskengine.JobKind, heartbeatAt time.Time, leaseExpiresAt time.Time) (*domainrealtime.Worker, error)
	TouchHeartbeat(ctx context.Context, workerID string, epoch int64, heartbeatAt time.Time, leaseExpiresAt time.Time) (*domainrealtime.Worker, error)
	UpdateState(ctx context.Context, workerID string, epoch int64, state domainrealtime.State, changedAt time.Time) (*domainrealtime.Worker, error)
	RemoveRegistration(ctx context.Context, workerID string, epoch int64) error
	ListWorkers(ctx context.Context, limit int) ([]domainrealtime.Worker, error)
	GetSettings(ctx context.Context) (domainrealtime.Settings, error)
	UpsertSettings(ctx context.Context, settings domainrealtime.Settings) (domainrealtime.Settings, error)
	CreateRegistrationSubmission(ctx context.Context, submission domainrealtime.RegistrationSubmission) (domainrealtime.RegistrationSubmission, error)
	ListPendingRegistrationSubmissions(ctx context.Context, limit int) ([]domainrealtime.RegistrationSubmission, error)
	ResolveRegistrationSubmission(ctx context.Context, submissionID string, status domainrealtime.RegistrationStatus, reasons []string, resolvedAt time.Time) (domainrealtime.RegistrationSubmission, error)
	RevokeRegistrationSubmission(ctx context.Context, submissionID string, reason string, revokedAt time.Time) (domainrealtime.RegistrationSubmission, error)
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

func DefaultSettings(now time.Time) domainrealtime.Settings {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return domainrealtime.Settings{
		HeartbeatInterval: 15 * time.Second,
		ResponseDeadline:  5 * time.Second,
		UpdatedAt:         now.UTC(),
	}
}

func (service *Service) Register(ctx context.Context, workerID string, capabilities []taskengine.JobKind, heartbeatInterval time.Duration) (*domainrealtime.Worker, error) {
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

func (service *Service) RecordHeartbeat(ctx context.Context, workerID string, epoch int64, heartbeatInterval time.Duration) (*domainrealtime.Worker, error) {
	if service == nil || service.repository == nil {
		return nil, ErrRepositoryRequired
	}
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return nil, fmt.Errorf("worker_id is required")
	}
	if epoch <= 0 {
		return nil, fmt.Errorf("epoch must be greater than zero")
	}
	if heartbeatInterval <= 0 {
		return nil, fmt.Errorf("heartbeat_interval must be greater than zero")
	}
	now := time.Now().UTC()
	return service.repository.TouchHeartbeat(ctx, workerID, epoch, now, now.Add(heartbeatInterval*3))
}

func (service *Service) Invalidate(ctx context.Context, workerID string, epoch int64) (*domainrealtime.Worker, error) {
	if service == nil || service.repository == nil {
		return nil, ErrRepositoryRequired
	}
	return service.repository.UpdateState(ctx, strings.TrimSpace(workerID), epoch, domainrealtime.StateInvalidated, time.Now().UTC())
}

func (service *Service) CreateRegistrationSubmission(ctx context.Context, submission domainrealtime.RegistrationSubmission) (domainrealtime.RegistrationSubmission, error) {
	if service == nil || service.repository == nil {
		return domainrealtime.RegistrationSubmission{}, ErrRepositoryRequired
	}
	if submission.Status == "" {
		submission.Status = domainrealtime.RegistrationStatusPending
	}
	if err := submission.Validate(); err != nil {
		return domainrealtime.RegistrationSubmission{}, err
	}
	return service.repository.CreateRegistrationSubmission(ctx, submission)
}

func (service *Service) ListPendingRegistrationSubmissions(ctx context.Context, limit int) ([]domainrealtime.RegistrationSubmission, error) {
	if service == nil || service.repository == nil {
		return nil, ErrRepositoryRequired
	}
	if limit <= 0 {
		limit = 200
	}
	return service.repository.ListPendingRegistrationSubmissions(ctx, limit)
}

func (service *Service) ResolveRegistrationSubmission(ctx context.Context, submissionID string, accepted bool, reasons []string) (domainrealtime.RegistrationSubmission, error) {
	if service == nil || service.repository == nil {
		return domainrealtime.RegistrationSubmission{}, ErrRepositoryRequired
	}
	status := domainrealtime.RegistrationStatusRejected
	if accepted {
		status = domainrealtime.RegistrationStatusAccepted
	}
	return service.repository.ResolveRegistrationSubmission(ctx, strings.TrimSpace(submissionID), status, reasons, time.Now().UTC())
}

func (service *Service) RevokeRegistrationSubmission(ctx context.Context, submissionID string, reason string) (domainrealtime.RegistrationSubmission, error) {
	if service == nil || service.repository == nil {
		return domainrealtime.RegistrationSubmission{}, ErrRepositoryRequired
	}
	return service.repository.RevokeRegistrationSubmission(ctx, strings.TrimSpace(submissionID), strings.TrimSpace(reason), time.Now().UTC())
}

func (service *Service) ForceDeregister(ctx context.Context, workerID string, epoch int64) (*domainrealtime.Worker, error) {
	if service == nil || service.repository == nil {
		return nil, ErrRepositoryRequired
	}
	worker, err := service.repository.UpdateState(ctx, strings.TrimSpace(workerID), epoch, domainrealtime.StateDeregistered, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if err := service.repository.RemoveRegistration(ctx, strings.TrimSpace(workerID), epoch); err != nil {
		return nil, err
	}
	return worker, nil
}

func (service *Service) Deregister(ctx context.Context, workerID string, epoch int64) (*domainrealtime.Worker, error) {
	if service == nil || service.repository == nil {
		return nil, ErrRepositoryRequired
	}
	worker, err := service.repository.UpdateState(ctx, strings.TrimSpace(workerID), epoch, domainrealtime.StateDeregistered, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if err := service.repository.RemoveRegistration(ctx, strings.TrimSpace(workerID), epoch); err != nil {
		return nil, err
	}
	return worker, nil
}

func (service *Service) ListWorkers(ctx context.Context, limit int) ([]domainrealtime.Worker, error) {
	if service == nil || service.repository == nil {
		return nil, ErrRepositoryRequired
	}
	if limit <= 0 {
		limit = 50
	}
	return service.repository.ListWorkers(ctx, limit)
}

func (service *Service) GetSettings(ctx context.Context) (domainrealtime.Settings, error) {
	if service == nil || service.repository == nil {
		return domainrealtime.Settings{}, ErrRepositoryRequired
	}
	return service.repository.GetSettings(ctx)
}

func (service *Service) UpdateSettings(ctx context.Context, settings domainrealtime.Settings) (domainrealtime.Settings, error) {
	if service == nil || service.repository == nil {
		return domainrealtime.Settings{}, ErrRepositoryRequired
	}
	if settings.UpdatedAt.IsZero() {
		settings.UpdatedAt = time.Now().UTC()
	}
	if err := settings.Validate(); err != nil {
		return domainrealtime.Settings{}, err
	}
	return service.repository.UpsertSettings(ctx, settings)
}

func (service *Service) EnsureBaseSettings(ctx context.Context, defaults domainrealtime.Settings) (domainrealtime.Settings, error) {
	if service == nil || service.repository == nil {
		return domainrealtime.Settings{}, ErrRepositoryRequired
	}
	settings, err := service.repository.GetSettings(ctx)
	if err == nil {
		return settings, nil
	}
	if !errors.Is(err, ErrSettingsNotFound) {
		return domainrealtime.Settings{}, err
	}
	if defaults.UpdatedAt.IsZero() {
		defaults.UpdatedAt = time.Now().UTC()
	}
	if validateErr := defaults.Validate(); validateErr != nil {
		return domainrealtime.Settings{}, validateErr
	}
	return service.repository.UpsertSettings(ctx, defaults)
}
