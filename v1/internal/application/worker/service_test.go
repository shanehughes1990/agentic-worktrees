package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"agentic-orchestrator/internal/application/taskengine"
	domainworker "agentic-orchestrator/internal/domain/worker"
)

type fakeRepository struct {
	worker   *domainworker.Worker
	settings domainworker.Settings
	updated  *domainworker.Worker
	getSettingsErr error
	removedWorkerID string
	removedEpoch int64
}

func (repository *fakeRepository) Register(ctx context.Context, workerID string, capabilities []taskengine.JobKind, heartbeatAt time.Time, leaseExpiresAt time.Time) (*domainworker.Worker, error) {
	repository.worker = &domainworker.Worker{WorkerID: workerID, Epoch: 1, State: domainworker.StateHealthy, DesiredState: domainworker.StateHealthy, Capabilities: capabilities, LastHeartbeat: heartbeatAt, LeaseExpiresAt: leaseExpiresAt, UpdatedAt: heartbeatAt}
	return repository.worker, nil
}

func (repository *fakeRepository) RenewHeartbeat(ctx context.Context, workerID string, epoch int64, heartbeatAt time.Time, leaseExpiresAt time.Time) (*domainworker.Worker, error) {
	if repository.worker == nil {
		repository.worker = &domainworker.Worker{WorkerID: workerID, Epoch: epoch, State: domainworker.StateHealthy, DesiredState: domainworker.StateHealthy, Capabilities: []taskengine.JobKind{taskengine.JobKindAgentWorkflow}, UpdatedAt: heartbeatAt}
	}
	repository.worker.LastHeartbeat = heartbeatAt
	repository.worker.LeaseExpiresAt = leaseExpiresAt
	return repository.worker, nil
}

func (repository *fakeRepository) UpdateState(ctx context.Context, workerID string, epoch int64, state domainworker.State, desiredState domainworker.State, reason string, changedAt time.Time) (*domainworker.Worker, error) {
	repository.updated = &domainworker.Worker{WorkerID: workerID, Epoch: epoch, State: state, DesiredState: desiredState, Capabilities: []taskengine.JobKind{taskengine.JobKindAgentWorkflow}, LastHeartbeat: changedAt.Add(-time.Second), LeaseExpiresAt: changedAt.Add(time.Second), RogueReason: reason, UpdatedAt: changedAt}
	return repository.updated, nil
}

func (repository *fakeRepository) RemoveRegistration(ctx context.Context, workerID string, epoch int64) error {
	repository.removedWorkerID = workerID
	repository.removedEpoch = epoch
	return nil
}

func (repository *fakeRepository) ListWorkers(ctx context.Context, limit int) ([]domainworker.Worker, error) {
	if repository.worker == nil {
		return []domainworker.Worker{}, nil
	}
	return []domainworker.Worker{*repository.worker}, nil
}

func (repository *fakeRepository) ListStaleWorkers(ctx context.Context, staleBefore time.Time, limit int) ([]domainworker.Worker, error) {
	if repository.worker == nil {
		return []domainworker.Worker{}, nil
	}
	return []domainworker.Worker{*repository.worker}, nil
}

func (repository *fakeRepository) GetSettings(ctx context.Context) (domainworker.Settings, error) {
	if repository.getSettingsErr != nil {
		return domainworker.Settings{}, repository.getSettingsErr
	}
	if repository.settings.HeartbeatInterval == 0 {
		repository.settings = domainworker.Settings{HeartbeatInterval: 15 * time.Second, ResponseDeadline: 5 * time.Second, StaleAfter: 45 * time.Second, DrainTimeout: 20 * time.Second, TerminateTimeout: 10 * time.Second, RogueThreshold: 3, UpdatedAt: time.Now().UTC()}
	}
	return repository.settings, nil
}

func (repository *fakeRepository) UpsertSettings(ctx context.Context, settings domainworker.Settings) (domainworker.Settings, error) {
	repository.settings = settings
	return settings, nil
}

type fakeEngine struct {
	requests []taskengine.EnqueueRequest
}

func (engine *fakeEngine) Enqueue(ctx context.Context, request taskengine.EnqueueRequest) (taskengine.EnqueueResult, error) {
	engine.requests = append(engine.requests, request)
	return taskengine.EnqueueResult{QueueTaskID: "id", Duplicate: false}, nil
}

func TestServiceHeartbeatReturnsStoppingForShutdownState(t *testing.T) {
	repository := &fakeRepository{worker: &domainworker.Worker{WorkerID: "worker-1", Epoch: 1, State: domainworker.StateDraining, DesiredState: domainworker.StateShutdownRequested, Capabilities: []taskengine.JobKind{taskengine.JobKindAgentWorkflow}, LastHeartbeat: time.Now().UTC(), LeaseExpiresAt: time.Now().UTC().Add(time.Minute), UpdatedAt: time.Now().UTC()}}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	_, err = service.Heartbeat(context.Background(), "worker-1", 1, 15*time.Second)
	if err == nil {
		t.Fatalf("expected heartbeat error")
	}
	if err != ErrApplicationStopping {
		t.Fatalf("expected ErrApplicationStopping, got %v", err)
	}
}

func TestCoordinatorEnqueuesShutdownTasksForStaleWorker(t *testing.T) {
	repository := &fakeRepository{worker: &domainworker.Worker{WorkerID: "worker-1", Epoch: 1, State: domainworker.StateStale, DesiredState: domainworker.StateHealthy, Capabilities: []taskengine.JobKind{taskengine.JobKindAgentWorkflow}, LastHeartbeat: time.Now().UTC().Add(-time.Minute), LeaseExpiresAt: time.Now().UTC().Add(-time.Second), UpdatedAt: time.Now().UTC()}}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	engine := &fakeEngine{}
	coordinator, err := NewCoordinator(service, engine)
	if err != nil {
		t.Fatalf("new coordinator: %v", err)
	}
	if err := coordinator.ProbeAndEscalate(context.Background()); err != nil {
		t.Fatalf("probe and escalate: %v", err)
	}
	if len(engine.requests) != 3 {
		t.Fatalf("expected 3 shutdown enqueue requests, got %d", len(engine.requests))
	}
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

func TestDeregisterRemovesRegistration(t *testing.T) {
	repository := &fakeRepository{}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	worker, err := service.Deregister(context.Background(), "worker-1", 7, "shutdown")
	if err != nil {
		t.Fatalf("deregister: %v", err)
	}
	if worker == nil {
		t.Fatalf("expected worker on deregister")
	}
	if repository.removedWorkerID != "worker-1" || repository.removedEpoch != 7 {
		t.Fatalf("expected registration removal for worker-1 epoch 7, got %q epoch %d", repository.removedWorkerID, repository.removedEpoch)
	}
}

func TestForceDeregisterRemovesRegistration(t *testing.T) {
	repository := &fakeRepository{}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	worker, err := service.ForceDeregister(context.Background(), "worker-2", 3, "forced")
	if err != nil {
		t.Fatalf("force deregister: %v", err)
	}
	if worker == nil {
		t.Fatalf("expected worker on force deregister")
	}
	if repository.removedWorkerID != "worker-2" || repository.removedEpoch != 3 {
		t.Fatalf("expected registration removal for worker-2 epoch 3, got %q epoch %d", repository.removedWorkerID, repository.removedEpoch)
	}
}
