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
	worktreeStateResult    domainscm.WorktreeState
	branchStateResult      domainscm.BranchState
	pullRequestState       domainscm.PullRequestState
	reviewDecision         domainscm.ReviewDecision
	mergeReadiness         domainscm.MergeReadiness
	sourceStateErr         error
	ensureWorktreeErr      error
	cleanupWorktreeErr     error
	createPullRequestErr   error
	checkMergeReadinessErr error
	capturedWorktreeSpec   domainscm.WorktreeSpec
}

func (fake *fakeOrchestrator) SourceState(_ context.Context, _ domainscm.Repository) (domainscm.SourceState, error) {
	return fake.sourceStateResult, fake.sourceStateErr
}

func (fake *fakeOrchestrator) EnsureWorktree(_ context.Context, _ domainscm.Repository, spec domainscm.WorktreeSpec) (domainscm.WorktreeState, error) {
	fake.capturedWorktreeSpec = spec
	return fake.worktreeStateResult, fake.ensureWorktreeErr
}

func (fake *fakeOrchestrator) SyncWorktree(_ context.Context, _ domainscm.Repository, _ string) (domainscm.WorktreeState, error) {
	return fake.worktreeStateResult, nil
}

func (fake *fakeOrchestrator) CleanupWorktree(_ context.Context, _ domainscm.Repository, _ string) error {
	return fake.cleanupWorktreeErr
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

func TestEnsureWorktreeClassifiesUnknownErrorsAsTransient(t *testing.T) {
	orchestrator := &fakeOrchestrator{
		ensureWorktreeErr: errors.New("temporary transport failure"),
		worktreeStateResult: domainscm.WorktreeState{
			Path:    "/tmp/worktrees/run-1-task-1",
			Branch:  "feature/one",
			Base:    "main",
			HeadSHA: "abc123",
		},
	}
	service, err := NewService(orchestrator)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, ensureErr := service.EnsureWorktree(context.Background(), EnsureWorktreeRequest{
		Repository: domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		Spec:       domainscm.WorktreeSpec{BaseBranch: "main", TargetBranch: "feature/one", Path: "/tmp/worktrees/run-1-task-1"},
		Metadata:   validMetadata(),
	})
	if !failures.IsClass(ensureErr, failures.ClassTransient) {
		t.Fatalf("expected transient error, got %q (%v)", failures.ClassOf(ensureErr), ensureErr)
	}
}

func TestEnsureWorktreeDefaultsSyncStrategyToMerge(t *testing.T) {
	orchestrator := &fakeOrchestrator{worktreeStateResult: domainscm.WorktreeState{
		Path:    "/tmp/worktrees/run-1-task-1",
		Branch:  "feature/one",
		Base:    "main",
		HeadSHA: "abc123",
	}}
	service, err := NewService(orchestrator)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, ensureErr := service.EnsureWorktree(context.Background(), EnsureWorktreeRequest{
		Repository: domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		Spec:       domainscm.WorktreeSpec{BaseBranch: "main", TargetBranch: "feature/one", Path: "/tmp/worktrees/run-1-task-1"},
		Metadata:   validMetadata(),
	})
	if ensureErr != nil {
		t.Fatalf("ensure worktree: %v", ensureErr)
	}
	if orchestrator.capturedWorktreeSpec.SyncStrategy != domainscm.SyncStrategyMerge {
		t.Fatalf("expected sync strategy merge, got %q", orchestrator.capturedWorktreeSpec.SyncStrategy)
	}
}

func TestEnsureWorktreePanicsWhenRebaseStrategyIsSelected(t *testing.T) {
	orchestrator := &fakeOrchestrator{worktreeStateResult: domainscm.WorktreeState{
		Path:    "/tmp/worktrees/run-1-task-1",
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
	_, _ = service.EnsureWorktree(context.Background(), EnsureWorktreeRequest{
		Repository: domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		Spec: domainscm.WorktreeSpec{
			BaseBranch:   "main",
			TargetBranch: "feature/one",
			Path:         "/tmp/worktrees/run-1-task-1",
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
