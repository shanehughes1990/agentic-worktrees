package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"agentic-orchestrator/internal/application/taskengine"
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
)

type fakeRepository struct {
	worker           *domainrealtime.Worker
	settings         domainrealtime.Settings
	updated          *domainrealtime.Worker
	getSettingsErr   error
	removedWorkerID  string
	removedEpoch     int64
}

func (repository *fakeRepository) Register(ctx context.Context, workerID string, capabilities []taskengine.JobKind, heartbeatAt time.Time, leaseExpiresAt time.Time) (*domainrealtime.Worker, error) {
	repository.worker = &domainrealtime.Worker{WorkerID: workerID, Epoch: 1, State: domainrealtime.StateHealthy, Capabilities: capabilities, LastHeartbeat: heartbeatAt, LeaseExpiresAt: leaseExpiresAt, UpdatedAt: heartbeatAt}
	return repository.worker, nil
}

func (repository *fakeRepository) UpdateState(ctx context.Context, workerID string, epoch int64, state domainrealtime.State, changedAt time.Time) (*domainrealtime.Worker, error) {
	repository.updated = &domainrealtime.Worker{WorkerID: workerID, Epoch: epoch, State: state, Capabilities: []taskengine.JobKind{taskengine.JobKindAgentWorkflow}, LastHeartbeat: changedAt.Add(-time.Second), LeaseExpiresAt: changedAt.Add(time.Second), UpdatedAt: changedAt}
	return repository.updated, nil
}

func (repository *fakeRepository) RemoveRegistration(ctx context.Context, workerID string, epoch int64) error {
	repository.removedWorkerID = workerID
	repository.removedEpoch = epoch
	return nil
}

func (repository *fakeRepository) ListWorkers(ctx context.Context, limit int) ([]domainrealtime.Worker, error) {
	if repository.worker == nil {
		return []domainrealtime.Worker{}, nil
	}
	return []domainrealtime.Worker{*repository.worker}, nil
}

func (repository *fakeRepository) GetSettings(ctx context.Context) (domainrealtime.Settings, error) {
	if repository.getSettingsErr != nil {
		return domainrealtime.Settings{}, repository.getSettingsErr
	}
	if repository.settings.HeartbeatInterval == 0 {
		repository.settings = domainrealtime.Settings{HeartbeatInterval: 15 * time.Second, ResponseDeadline: 5 * time.Second, UpdatedAt: time.Now().UTC()}
	}
	return repository.settings, nil
}

func (repository *fakeRepository) UpsertSettings(ctx context.Context, settings domainrealtime.Settings) (domainrealtime.Settings, error) {
	repository.settings = settings
	return settings, nil
}

func TestEnsureBaseSettingsCreatesWhenMissing(t *testing.T) {
	repository := &fakeRepository{getSettingsErr: ErrSettingsNotFound}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	settings, err := service.EnsureBaseSettings(context.Background(), DefaultSettings(time.Now().UTC()))
	if err != nil {
		t.Fatalf("ensure base settings: %v", err)
	}
	if settings.HeartbeatInterval != 15*time.Second {
		t.Fatalf("expected heartbeat interval 15s, got %s", settings.HeartbeatInterval)
	}
	if repository.settings.HeartbeatInterval != 15*time.Second {
		t.Fatalf("expected repository settings to be persisted")
	}
}

func TestEnsureBaseSettingsReturnsErrorForNonNotFound(t *testing.T) {
	repository := &fakeRepository{getSettingsErr: errors.New("db unavailable")}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	_, err = service.EnsureBaseSettings(context.Background(), DefaultSettings(time.Now().UTC()))
	if err == nil {
		t.Fatalf("expected ensure base settings error")
	}
}

