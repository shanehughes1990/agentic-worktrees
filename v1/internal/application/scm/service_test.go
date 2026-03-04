package scm

import (
	"agentic-orchestrator/internal/application/taskengine"
	"agentic-orchestrator/internal/domain/failures"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"errors"
	"testing"
)

type fakeOrchestrator struct {
	sourceStateResult      domainscm.SourceState
	repositoryStateResult    domainscm.RepositoryState
	branchStateResult      domainscm.BranchState
	pullRequestState       domainscm.PullRequestState
	reviewDecision         domainscm.ReviewDecision
	mergeReadiness         domainscm.MergeReadiness
	sourceStateErr         error
	ensureRepositoryErr      error
	cleanupRepositoryErr     error
	createPullRequestErr   error
	checkMergeReadinessErr error
	capturedRepositorySpec   domainscm.RepositorySpec
}

func (fake *fakeOrchestrator) SourceState(_ context.Context, _ domainscm.Repository) (domainscm.SourceState, error) {
	return fake.sourceStateResult, fake.sourceStateErr
}

func (fake *fakeOrchestrator) EnsureRepository(_ context.Context, _ domainscm.Repository, spec domainscm.RepositorySpec) (domainscm.RepositoryState, error) {
	fake.capturedRepositorySpec = spec
	return fake.repositoryStateResult, fake.ensureRepositoryErr
}

func (fake *fakeOrchestrator) SyncRepository(_ context.Context, _ domainscm.Repository, _ string) (domainscm.RepositoryState, error) {
	return fake.repositoryStateResult, nil
}

func (fake *fakeOrchestrator) CleanupRepository(_ context.Context, _ domainscm.Repository, _ string) error {
	return fake.cleanupRepositoryErr
}

func (fake *fakeOrchestrator) EnsureBranch(_ context.Context, _ domainscm.Repository, _ domainscm.BranchSpec) (domainscm.BranchState, error) {
	return fake.branchStateResult, nil
}

func (fake *fakeOrchestrator) SyncBranch(_ context.Context, _ domainscm.Repository, _ string) (domainscm.BranchState, error) {
	return fake.branchStateResult, nil
}

func (fake *fakeOrchestrator) CreateOrUpdatePullRequest(_ context.Context, _ domainscm.PullRequestSpec) (domainscm.PullRequestState, error) {
	return fake.pullRequestState, fake.createPullRequestErr
}

func (fake *fakeOrchestrator) GetPullRequest(_ context.Context, _ domainscm.Repository, _ int) (domainscm.PullRequestState, error) {
	return fake.pullRequestState, nil
}

func (fake *fakeOrchestrator) SubmitReview(_ context.Context, _ domainscm.ReviewSpec) (domainscm.ReviewDecision, error) {
	return fake.reviewDecision, nil
}

func (fake *fakeOrchestrator) CheckMergeReadiness(_ context.Context, _ domainscm.Repository, _ int) (domainscm.MergeReadiness, error) {
	return fake.mergeReadiness, fake.checkMergeReadinessErr
}

type fakeLeaseManagerForService struct {
	acquireErr error
}

func (fake *fakeLeaseManagerForService) Acquire(_ context.Context, request RepoLeaseAcquireRequest) (domainscm.RepoLease, error) {
	if fake.acquireErr != nil {
		return domainscm.RepoLease{}, fake.acquireErr
	}
	return domainscm.RepoLease{CacheKey: request.CacheKey, OwnerID: request.OwnerID, Token: request.Token}, nil
}

func (fake *fakeLeaseManagerForService) Release(_ context.Context, _ domainscm.RepoLease) error {
	return nil
}

func validMetadata() Metadata {
	return Metadata{
		CorrelationIDs: taskengine.CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1"},
		IdempotencyKey: "id-1",
	}
}

func TestSourceStateRejectsInvalidMetadata(t *testing.T) {
	service, err := NewService(&fakeOrchestrator{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, sourceStateErr := service.SourceState(context.Background(), SourceStateRequest{
		Repository: domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		Metadata:   Metadata{},
	})
	if !failures.IsClass(sourceStateErr, failures.ClassTerminal) {
		t.Fatalf("expected terminal error, got %q (%v)", failures.ClassOf(sourceStateErr), sourceStateErr)
	}
}

func TestEnsureRepositoryClassifiesUnknownErrorsAsTransient(t *testing.T) {
	orchestrator := &fakeOrchestrator{
		ensureRepositoryErr: errors.New("temporary transport failure"),
		repositoryStateResult: domainscm.RepositoryState{
			Path:    "/tmp/repositories/run-1-task-1",
			Branch:  "feature/one",
			Base:    "main",
			HeadSHA: "abc123",
		},
	}
	service, err := NewService(orchestrator)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, ensureErr := service.EnsureRepository(context.Background(), EnsureRepositoryRequest{
		Repository: domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		Spec:       domainscm.RepositorySpec{BaseBranch: "main", TargetBranch: "feature/one", Path: "/tmp/repositories/run-1-task-1"},
		Metadata:   validMetadata(),
	})
	if !failures.IsClass(ensureErr, failures.ClassTransient) {
		t.Fatalf("expected transient error, got %q (%v)", failures.ClassOf(ensureErr), ensureErr)
	}
}

func TestEnsureRepositoryClassifiesLeaseAcquireUnknownErrorAsTransient(t *testing.T) {
	orchestrator := &fakeOrchestrator{repositoryStateResult: domainscm.RepositoryState{Path: "/tmp/repository", Branch: "feature/one", Base: "main", HeadSHA: "abc"}}
	service, err := NewServiceWithLeaseManager(orchestrator, &fakeLeaseManagerForService{acquireErr: errors.New("lease backend unavailable")})
	if err != nil {
		t.Fatalf("new service with lease manager: %v", err)
	}

	_, ensureErr := service.EnsureRepository(context.Background(), EnsureRepositoryRequest{
		Repository: domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		Spec:       domainscm.RepositorySpec{BaseBranch: "main", TargetBranch: "feature/one", Path: "/tmp/repository"},
		Metadata:   validMetadata(),
	})
	if !failures.IsClass(ensureErr, failures.ClassTransient) {
		t.Fatalf("expected transient lease acquire error, got %q (%v)", failures.ClassOf(ensureErr), ensureErr)
	}
}

func TestEnsureRepositoryDefaultsSyncStrategyToMerge(t *testing.T) {
	orchestrator := &fakeOrchestrator{repositoryStateResult: domainscm.RepositoryState{
		Path:    "/tmp/repositories/run-1-task-1",
		Branch:  "feature/one",
		Base:    "main",
		HeadSHA: "abc123",
	}}
	service, err := NewService(orchestrator)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, ensureErr := service.EnsureRepository(context.Background(), EnsureRepositoryRequest{
		Repository: domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		Spec:       domainscm.RepositorySpec{BaseBranch: "main", TargetBranch: "feature/one", Path: "/tmp/repositories/run-1-task-1"},
		Metadata:   validMetadata(),
	})
	if ensureErr != nil {
		t.Fatalf("ensure repository: %v", ensureErr)
	}
	if orchestrator.capturedRepositorySpec.SyncStrategy != domainscm.SyncStrategyMerge {
		t.Fatalf("expected sync strategy merge, got %q", orchestrator.capturedRepositorySpec.SyncStrategy)
	}
}

func TestEnsureRepositoryPanicsWhenRebaseStrategyIsSelected(t *testing.T) {
	orchestrator := &fakeOrchestrator{repositoryStateResult: domainscm.RepositoryState{
		Path:    "/tmp/repositories/run-1-task-1",
		Branch:  "feature/one",
		Base:    "main",
		HeadSHA: "abc123",
	}}
	service, err := NewService(orchestrator)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic for rebase sync strategy")
		}
	}()
	_, _ = service.EnsureRepository(context.Background(), EnsureRepositoryRequest{
		Repository: domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		Spec: domainscm.RepositorySpec{
			BaseBranch:   "main",
			TargetBranch: "feature/one",
			Path:         "/tmp/repositories/run-1-task-1",
			SyncStrategy: domainscm.SyncStrategyRebase,
		},
		Metadata: validMetadata(),
	})
}

func TestCreateOrUpdatePullRequestReturnsState(t *testing.T) {
	orchestrator := &fakeOrchestrator{
		pullRequestState: domainscm.PullRequestState{Number: 12, URL: "https://github.com/acme/repo/pull/12", State: "open", HeadSHA: "abc123"},
	}
	service, err := NewService(orchestrator)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	result, createErr := service.CreateOrUpdatePullRequest(context.Background(), CreateOrUpdatePullRequestRequest{
		Spec: domainscm.PullRequestSpec{
			Repository:   domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
			SourceBranch: "feature/one",
			TargetBranch: "main",
			Title:        "Feature one",
			Body:         "Implements feature one",
		},
		Metadata: validMetadata(),
	})
	if createErr != nil {
		t.Fatalf("create or update pull request: %v", createErr)
	}
	if result.Number != 12 {
		t.Fatalf("expected pull request number 12, got %d", result.Number)
	}
}

func TestCheckMergeReadinessReturnsMergeableResult(t *testing.T) {
	orchestrator := &fakeOrchestrator{mergeReadiness: domainscm.MergeReadiness{CanMerge: true}}
	service, err := NewService(orchestrator)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	result, checkErr := service.CheckMergeReadiness(context.Background(), CheckMergeReadinessRequest{
		Repository:        domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		PullRequestNumber: 8,
		Metadata:          validMetadata(),
	})
	if checkErr != nil {
		t.Fatalf("check merge readiness: %v", checkErr)
	}
	if !result.CanMerge {
		t.Fatalf("expected mergeable result")
	}
}

func (fake *fakeOrchestrator) MergePullRequest(_ context.Context, _ domainscm.MergePullRequestSpec) (domainscm.PullRequestState, error) {
	return fake.pullRequestState, fake.createPullRequestErr
}
