package worker

import (
	applicationscm "agentic-orchestrator/internal/application/scm"
	"agentic-orchestrator/internal/application/taskengine"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type fakeSCMService struct {
	called string
	err    error
}

func (fake *fakeSCMService) SourceState(ctx context.Context, request applicationscm.SourceStateRequest) (domainscm.SourceState, error) {
	fake.called = "source_state"
	return domainscm.SourceState{DefaultBranch: "main", HeadSHA: "abc"}, fake.err
}
func (fake *fakeSCMService) EnsureWorktree(ctx context.Context, request applicationscm.EnsureWorktreeRequest) (domainscm.WorktreeState, error) {
	fake.called = "ensure_worktree"
	return domainscm.WorktreeState{Path: request.Spec.Path, Branch: request.Spec.TargetBranch, Base: request.Spec.BaseBranch, HeadSHA: "abc"}, fake.err
}
func (fake *fakeSCMService) SyncWorktree(ctx context.Context, request applicationscm.SyncWorktreeRequest) (domainscm.WorktreeState, error) {
	fake.called = "sync_worktree"
	return domainscm.WorktreeState{Path: request.Path, Branch: "feature", Base: "main", HeadSHA: "abc"}, fake.err
}
func (fake *fakeSCMService) CleanupWorktree(ctx context.Context, request applicationscm.CleanupWorktreeRequest) error {
	fake.called = "cleanup_worktree"
	return fake.err
}
func (fake *fakeSCMService) EnsureBranch(ctx context.Context, request applicationscm.EnsureBranchRequest) (domainscm.BranchState, error) {
	fake.called = "ensure_branch"
	return domainscm.BranchState{Name: request.Spec.TargetBranch, Base: request.Spec.BaseBranch, HeadSHA: "abc"}, fake.err
}
func (fake *fakeSCMService) SyncBranch(ctx context.Context, request applicationscm.SyncBranchRequest) (domainscm.BranchState, error) {
	fake.called = "sync_branch"
	return domainscm.BranchState{Name: request.BranchName, Base: request.BranchName, HeadSHA: "abc"}, fake.err
}
func (fake *fakeSCMService) CreateOrUpdatePullRequest(ctx context.Context, request applicationscm.CreateOrUpdatePullRequestRequest) (domainscm.PullRequestState, error) {
	fake.called = "upsert_pull_request"
	return domainscm.PullRequestState{Number: 1, URL: "https://example/pull/1", State: "open", HeadSHA: "abc"}, fake.err
}
func (fake *fakeSCMService) GetPullRequest(ctx context.Context, request applicationscm.GetPullRequestRequest) (domainscm.PullRequestState, error) {
	fake.called = "get_pull_request"
	return domainscm.PullRequestState{Number: request.PullRequestNumber, URL: "https://example/pull/1", State: "open", HeadSHA: "abc"}, fake.err
}
func (fake *fakeSCMService) SubmitReview(ctx context.Context, request applicationscm.SubmitReviewRequest) (domainscm.ReviewDecision, error) {
	fake.called = "submit_review"
	return request.Spec.Decision, fake.err
}
func (fake *fakeSCMService) CheckMergeReadiness(ctx context.Context, request applicationscm.CheckMergeReadinessRequest) (domainscm.MergeReadiness, error) {
	fake.called = "check_merge_readiness"
	return domainscm.MergeReadiness{CanMerge: true}, fake.err
}

func TestSCMWorkflowHandlerDispatchesEnsureWorktree(t *testing.T) {
	service := &fakeSCMService{}
	handler, err := NewSCMWorkflowHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	payload, _ := json.Marshal(SCMWorkflowPayload{Operation: "ensure_worktree", Provider: "github", Owner: "acme", Repository: "repo", RunID: "run-1", TaskID: "task-1", JobID: "job-1", IdempotencyKey: "id-1", BaseBranch: "main", TargetBranch: "feature", WorktreePath: "/tmp/worktree"})
	err = handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, Payload: payload})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if service.called != "ensure_worktree" {
		t.Fatalf("expected ensure_worktree call, got %q", service.called)
	}
}

func TestSCMWorkflowHandlerReturnsServiceError(t *testing.T) {
	service := &fakeSCMService{err: errors.New("boom")}
	handler, err := NewSCMWorkflowHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(SCMWorkflowPayload{Operation: "source_state", Provider: "github", Owner: "acme", Repository: "repo", RunID: "run-1", TaskID: "task-1", JobID: "job-1", IdempotencyKey: "id-1"})
	err = handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, Payload: payload})
	if err == nil {
		t.Fatalf("expected service error")
	}
}
