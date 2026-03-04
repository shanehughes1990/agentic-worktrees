package taskengine

import (
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"errors"
	"testing"
)

type fakeSCMBootstrapPort struct {
	sourceStateRequest   domainscm.Repository
	ensureRepositoryRepo   domainscm.Repository
	ensureRepositorySpec   domainscm.RepositorySpec
	sourceStateResponse  domainscm.SourceState
	ensureRepositoryResult domainscm.RepositoryState
	sourceStateErr       error
	ensureRepositoryErr    error
	sourceStateCalls     int
	ensureRepositoryCalls  int
}

func (fake *fakeSCMBootstrapPort) SourceState(ctx context.Context, repository domainscm.Repository) (domainscm.SourceState, error) {
	_ = ctx
	fake.sourceStateCalls++
	fake.sourceStateRequest = repository
	if fake.sourceStateErr != nil {
		return domainscm.SourceState{}, fake.sourceStateErr
	}
	if fake.sourceStateResponse.DefaultBranch == "" {
		return domainscm.SourceState{DefaultBranch: "main", HeadSHA: "sha-main"}, nil
	}
	return fake.sourceStateResponse, nil
}

func (fake *fakeSCMBootstrapPort) EnsureRepository(ctx context.Context, repository domainscm.Repository, spec domainscm.RepositorySpec) (domainscm.RepositoryState, error) {
	_ = ctx
	fake.ensureRepositoryCalls++
	fake.ensureRepositoryRepo = repository
	fake.ensureRepositorySpec = spec
	if fake.ensureRepositoryErr != nil {
		return domainscm.RepositoryState{}, fake.ensureRepositoryErr
	}
	if fake.ensureRepositoryResult.Path == "" {
		return domainscm.RepositoryState{Path: spec.Path, Branch: spec.TargetBranch, Base: spec.BaseBranch, HeadSHA: "sha-repository", IsInSync: true}, nil
	}
	return fake.ensureRepositoryResult, nil
}

type fakeRemoteWorkerAdapter struct {
	request RemoteExecutionRequest
	result  RemoteExecutionResult
	err     error
	calls   int
}

func (fake *fakeRemoteWorkerAdapter) Execute(ctx context.Context, request RemoteExecutionRequest) (RemoteExecutionResult, error) {
	_ = ctx
	fake.calls++
	fake.request = request
	if fake.err != nil {
		return RemoteExecutionResult{}, fake.err
	}
	if fake.result.WorkerID == "" {
		return RemoteExecutionResult{WorkerID: "remote-worker-1", CompletedCheckpoint: request.ResumeCheckpoint}, nil
	}
	return fake.result, nil
}

func validRemoteBootstrapRequest() RemoteBootstrapRequest {
	return RemoteBootstrapRequest{
		Job: Job{
			Kind:        JobKindSCMWorkflow,
			QueueTaskID: "queue-task-1",
			Payload:     []byte(`{"operation":"source_state"}`),
		},
		CorrelationIDs: CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1"},
		IdempotencyKey: "id-1",
		Repository:     domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		TargetBranch:   "feature/remote",
		RepositoryPath:   "/tmp/repositories/feature-remote",
	}
}

func TestRemoteBootstrapBuildUsesSourceStateDefaultBranch(t *testing.T) {
	scm := &fakeSCMBootstrapPort{sourceStateResponse: domainscm.SourceState{DefaultBranch: "main", HeadSHA: "sha-main"}}
	service, err := NewRemoteBootstrapService(scm, &fakeRemoteWorkerAdapter{})
	if err != nil {
		t.Fatalf("new remote bootstrap service: %v", err)
	}

	request := validRemoteBootstrapRequest()
	remoteRequest, buildErr := service.BuildRemoteExecutionRequest(context.Background(), request)
	if buildErr != nil {
		t.Fatalf("build remote request: %v", buildErr)
	}
	if scm.sourceStateCalls != 1 {
		t.Fatalf("expected one source_state call, got %d", scm.sourceStateCalls)
	}
	if scm.ensureRepositoryCalls != 1 {
		t.Fatalf("expected one ensure_repository call, got %d", scm.ensureRepositoryCalls)
	}
	if scm.ensureRepositorySpec.BaseBranch != "main" {
		t.Fatalf("expected base branch main, got %q", scm.ensureRepositorySpec.BaseBranch)
	}
	if remoteRequest.ResumeCheckpoint == nil || remoteRequest.ResumeCheckpoint.Step != "ensure_repository" {
		t.Fatalf("expected ensure_repository checkpoint, got %+v", remoteRequest.ResumeCheckpoint)
	}
}

func TestRemoteBootstrapBuildUsesExplicitBaseBranch(t *testing.T) {
	scm := &fakeSCMBootstrapPort{}
	service, err := NewRemoteBootstrapService(scm, &fakeRemoteWorkerAdapter{})
	if err != nil {
		t.Fatalf("new remote bootstrap service: %v", err)
	}
	request := validRemoteBootstrapRequest()
	request.BaseBranch = "release/1.0"

	_, buildErr := service.BuildRemoteExecutionRequest(context.Background(), request)
	if buildErr != nil {
		t.Fatalf("build remote request: %v", buildErr)
	}
	if scm.sourceStateCalls != 0 {
		t.Fatalf("expected source_state to be skipped when base branch provided, got %d calls", scm.sourceStateCalls)
	}
	if scm.ensureRepositorySpec.BaseBranch != "release/1.0" {
		t.Fatalf("expected explicit base branch, got %q", scm.ensureRepositorySpec.BaseBranch)
	}
}

func TestRemoteBootstrapExecuteDispatchesToRemoteAdapter(t *testing.T) {
	scm := &fakeSCMBootstrapPort{}
	remote := &fakeRemoteWorkerAdapter{}
	service, err := NewRemoteBootstrapService(scm, remote)
	if err != nil {
		t.Fatalf("new remote bootstrap service: %v", err)
	}

	request := validRemoteBootstrapRequest()
	result, executeErr := service.Execute(context.Background(), request)
	if executeErr != nil {
		t.Fatalf("execute: %v", executeErr)
	}
	if remote.calls != 1 {
		t.Fatalf("expected one remote adapter call, got %d", remote.calls)
	}
	if remote.request.Job.QueueTaskID != request.Job.QueueTaskID {
		t.Fatalf("expected queue_task_id %q, got %q", request.Job.QueueTaskID, remote.request.Job.QueueTaskID)
	}
	if result.WorkerID != "remote-worker-1" {
		t.Fatalf("expected remote-worker-1, got %q", result.WorkerID)
	}
}

func TestRemoteBootstrapExecuteReturnsBootstrapError(t *testing.T) {
	scm := &fakeSCMBootstrapPort{ensureRepositoryErr: errors.New("cannot create repository")}
	service, err := NewRemoteBootstrapService(scm, &fakeRemoteWorkerAdapter{})
	if err != nil {
		t.Fatalf("new remote bootstrap service: %v", err)
	}

	_, executeErr := service.Execute(context.Background(), validRemoteBootstrapRequest())
	if executeErr == nil {
		t.Fatalf("expected bootstrap error")
	}
}

func TestRemoteBootstrapRequestValidateRequiresRepositoryPath(t *testing.T) {
	request := validRemoteBootstrapRequest()
	request.RepositoryPath = ""
	if err := request.Validate(); !errors.Is(err, ErrInvalidRemoteExecutionRequest) {
		t.Fatalf("expected ErrInvalidRemoteExecutionRequest, got %v", err)
	}
}
