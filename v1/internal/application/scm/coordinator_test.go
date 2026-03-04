package scm

import (
	"agentic-orchestrator/internal/application/taskengine"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"errors"
	"testing"
)

type fakeLeaseManager struct {
	acquireRequest RepoLeaseAcquireRequest
	releaseLease   domainscm.RepoLease
	acquireErr     error
	releaseErr     error
	acquireCalls   int
	releaseCalls   int
}

func (fake *fakeLeaseManager) Acquire(ctx context.Context, request RepoLeaseAcquireRequest) (domainscm.RepoLease, error) {
	_ = ctx
	fake.acquireCalls++
	fake.acquireRequest = request
	if fake.acquireErr != nil {
		return domainscm.RepoLease{}, fake.acquireErr
	}
	return domainscm.RepoLease{CacheKey: request.CacheKey, OwnerID: request.OwnerID, Token: request.Token}, nil
}

func (fake *fakeLeaseManager) Release(ctx context.Context, lease domainscm.RepoLease) error {
	_ = ctx
	fake.releaseCalls++
	fake.releaseLease = lease
	return fake.releaseErr
}

func TestEnsureRepositoryCoordinatorAcquiresAndReleasesLease(t *testing.T) {
	orchestrator := &fakeOrchestrator{repositoryStateResult: domainscm.RepositoryState{Path: "/tmp/repository", Branch: "feature/one", Base: "main", HeadSHA: "abc"}}
	leaseManager := &fakeLeaseManager{}
	coordinator, err := NewEnsureRepositoryCoordinator(orchestrator, leaseManager)
	if err != nil {
		t.Fatalf("new coordinator: %v", err)
	}

	state, ensureErr := coordinator.Ensure(context.Background(), EnsureRepositoryRequest{
		Repository: domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		Spec:       domainscm.RepositorySpec{BaseBranch: "main", TargetBranch: "feature/one", Path: "/tmp/repository"},
		Metadata: Metadata{CorrelationIDs: taskengine.CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1"}, IdempotencyKey: "id-1"},
	})
	if ensureErr != nil {
		t.Fatalf("ensure: %v", ensureErr)
	}
	if state.Path == "" {
		t.Fatalf("expected repository state")
	}
	if leaseManager.acquireCalls != 1 {
		t.Fatalf("expected one acquire call, got %d", leaseManager.acquireCalls)
	}
	if leaseManager.releaseCalls != 1 {
		t.Fatalf("expected one release call, got %d", leaseManager.releaseCalls)
	}
	if leaseManager.acquireRequest.CacheKey != domainscm.RepoCacheKey("github/acme/repo") {
		t.Fatalf("expected cache key github/acme/repo, got %q", leaseManager.acquireRequest.CacheKey)
	}
}

func TestEnsureRepositoryCoordinatorReleasesLeaseWhenEnsureFails(t *testing.T) {
	orchestrator := &fakeOrchestrator{ensureRepositoryErr: errors.New("git fetch failed")}
	leaseManager := &fakeLeaseManager{}
	coordinator, err := NewEnsureRepositoryCoordinator(orchestrator, leaseManager)
	if err != nil {
		t.Fatalf("new coordinator: %v", err)
	}

	_, ensureErr := coordinator.Ensure(context.Background(), EnsureRepositoryRequest{
		Repository: domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		Spec:       domainscm.RepositorySpec{BaseBranch: "main", TargetBranch: "feature/one", Path: "/tmp/repository"},
		Metadata: Metadata{CorrelationIDs: taskengine.CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1"}, IdempotencyKey: "id-1"},
	})
	if ensureErr == nil {
		t.Fatalf("expected ensure error")
	}
	if leaseManager.releaseCalls != 1 {
		t.Fatalf("expected release call on ensure failure, got %d", leaseManager.releaseCalls)
	}
}

func TestEnsureRepositoryCoordinatorPropagatesAcquireError(t *testing.T) {
	orchestrator := &fakeOrchestrator{repositoryStateResult: domainscm.RepositoryState{Path: "/tmp/repository", Branch: "feature/one", Base: "main", HeadSHA: "abc"}}
	leaseManager := &fakeLeaseManager{acquireErr: errors.New("lease busy")}
	coordinator, err := NewEnsureRepositoryCoordinator(orchestrator, leaseManager)
	if err != nil {
		t.Fatalf("new coordinator: %v", err)
	}

	_, ensureErr := coordinator.Ensure(context.Background(), EnsureRepositoryRequest{
		Repository: domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		Spec:       domainscm.RepositorySpec{BaseBranch: "main", TargetBranch: "feature/one", Path: "/tmp/repository"},
		Metadata: Metadata{CorrelationIDs: taskengine.CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1"}, IdempotencyKey: "id-1"},
	})
	if ensureErr == nil {
		t.Fatalf("expected acquire error")
	}
}

func TestEnsureRepositoryCoordinatorPanicsOnRebaseStrategy(t *testing.T) {
	orchestrator := &fakeOrchestrator{repositoryStateResult: domainscm.RepositoryState{Path: "/tmp/repository", Branch: "feature/one", Base: "main", HeadSHA: "abc"}}
	coordinator, err := NewEnsureRepositoryCoordinator(orchestrator, nil)
	if err != nil {
		t.Fatalf("new coordinator: %v", err)
	}
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	_, _ = coordinator.Ensure(context.Background(), EnsureRepositoryRequest{
		Repository: domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		Spec: domainscm.RepositorySpec{
			BaseBranch:   "main",
			TargetBranch: "feature/one",
			Path:         "/tmp/repository",
			SyncStrategy: domainscm.SyncStrategyRebase,
		},
		Metadata: Metadata{CorrelationIDs: taskengine.CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1"}, IdempotencyKey: "id-1"},
	})
}
