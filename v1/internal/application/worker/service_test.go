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
	worker          *domainrealtime.Worker
	settings        domainrealtime.Settings
	updated         *domainrealtime.Worker
	getSettingsErr  error
	removedWorkerID string
	removedEpoch    int64
	submission      domainrealtime.RegistrationSubmission
}

func (repository *fakeRepository) Register(ctx context.Context, workerID string, capabilities []taskengine.JobKind, heartbeatAt time.Time, leaseExpiresAt time.Time) (*domainrealtime.Worker, error) {
	repository.worker = &domainrealtime.Worker{WorkerID: workerID, Epoch: 1, State: domainrealtime.StateHealthy, Capabilities: capabilities, LastHeartbeat: heartbeatAt, LeaseExpiresAt: leaseExpiresAt, UpdatedAt: heartbeatAt}
	return repository.worker, nil
}

func (repository *fakeRepository) UpdateState(ctx context.Context, workerID string, epoch int64, state domainrealtime.State, changedAt time.Time) (*domainrealtime.Worker, error) {
	repository.updated = &domainrealtime.Worker{WorkerID: workerID, Epoch: epoch, State: state, Capabilities: []taskengine.JobKind{taskengine.JobKindAgentWorkflow}, LastHeartbeat: changedAt.Add(-time.Second), LeaseExpiresAt: changedAt.Add(time.Second), UpdatedAt: changedAt}
	return repository.updated, nil
}

func (repository *fakeRepository) TouchHeartbeat(ctx context.Context, workerID string, epoch int64, heartbeatAt time.Time, leaseExpiresAt time.Time) (*domainrealtime.Worker, error) {
	repository.updated = &domainrealtime.Worker{WorkerID: workerID, Epoch: epoch, State: domainrealtime.StateHealthy, Capabilities: []taskengine.JobKind{taskengine.JobKindAgentWorkflow}, LastHeartbeat: heartbeatAt, LeaseExpiresAt: leaseExpiresAt, UpdatedAt: heartbeatAt}
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

func (repository *fakeRepository) CreateRegistrationSubmission(ctx context.Context, submission domainrealtime.RegistrationSubmission) (domainrealtime.RegistrationSubmission, error) {
	repository.submission = submission
	return submission, nil
}

func (repository *fakeRepository) GetRegistrationSubmission(ctx context.Context, submissionID string) (domainrealtime.RegistrationSubmission, error) {
	if repository.submission.SubmissionID == "" || repository.submission.SubmissionID != submissionID {
		return domainrealtime.RegistrationSubmission{}, errors.New("submission not found")
	}
	return repository.submission, nil
}

func (repository *fakeRepository) ListPendingRegistrationSubmissions(ctx context.Context, limit int) ([]domainrealtime.RegistrationSubmission, error) {
	if repository.submission.SubmissionID == "" {
		return []domainrealtime.RegistrationSubmission{}, nil
	}
	return []domainrealtime.RegistrationSubmission{repository.submission}, nil
}

func (repository *fakeRepository) ResolveRegistrationSubmission(ctx context.Context, submissionID string, status domainrealtime.RegistrationStatus, reasons []string, resolvedAt time.Time) (domainrealtime.RegistrationSubmission, error) {
	repository.submission.Status = status
	repository.submission.RejectReasons = reasons
	repository.submission.ResolvedAt = resolvedAt
	return repository.submission, nil
}

func (repository *fakeRepository) RevokeRegistrationSubmission(ctx context.Context, submissionID string, reason string, revokedAt time.Time) (domainrealtime.RegistrationSubmission, error) {
	repository.submission.Status = domainrealtime.RegistrationStatusRevoked
	repository.submission.RejectReasons = []string{reason}
	repository.submission.ResolvedAt = revokedAt
	return repository.submission, nil
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
