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
	called                 string
	err                    error
	lastMetadata           applicationscm.Metadata
	lastEnsureWorktreeSpec domainscm.WorktreeSpec
}

func (fake *fakeSCMService) SourceState(ctx context.Context, request applicationscm.SourceStateRequest) (domainscm.SourceState, error) {
	fake.called = "source_state"
	fake.lastMetadata = request.Metadata
	return domainscm.SourceState{DefaultBranch: "main", HeadSHA: "abc"}, fake.err
}
func (fake *fakeSCMService) EnsureWorktree(ctx context.Context, request applicationscm.EnsureWorktreeRequest) (domainscm.WorktreeState, error) {
	fake.called = "ensure_worktree"
	fake.lastEnsureWorktreeSpec = request.Spec
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

	payload, _ := json.Marshal(SCMWorkflowPayload{Operation: "ensure_worktree", Provider: "github", Owner: "acme", Repository: "repo", RunID: "run-1", TaskID: "task-1", JobID: "job-1", IdempotencyKey: "id-1", BaseBranch: "main", TargetBranch: "feature", WorktreePath: "/tmp/worktree", SyncStrategy: "merge"})
	err = handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, Payload: payload})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if service.called != "ensure_worktree" {
		t.Fatalf("expected ensure_worktree call, got %q", service.called)
	}
	if service.lastEnsureWorktreeSpec.SyncStrategy != domainscm.SyncStrategyMerge {
		t.Fatalf("expected merge sync strategy, got %q", service.lastEnsureWorktreeSpec.SyncStrategy)
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

func TestSCMWorkflowHandlerPropagatesCorrelationIDs(t *testing.T) {
	service := &fakeSCMService{}
	handler, err := NewSCMWorkflowHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(SCMWorkflowPayload{Operation: "source_state", Provider: "github", Owner: "acme", Repository: "repo", RunID: "run-x", TaskID: "task-x", JobID: "job-x", IdempotencyKey: "id-x"})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if service.lastMetadata.CorrelationIDs.RunID != "run-x" {
		t.Fatalf("expected run_id run-x, got %q", service.lastMetadata.CorrelationIDs.RunID)
	}
	if service.lastMetadata.CorrelationIDs.TaskID != "task-x" {
		t.Fatalf("expected task_id task-x, got %q", service.lastMetadata.CorrelationIDs.TaskID)
	}
	if service.lastMetadata.CorrelationIDs.JobID != "job-x" {
		t.Fatalf("expected job_id job-x, got %q", service.lastMetadata.CorrelationIDs.JobID)
	}
}

func TestSCMWorkflowHandlerSkipsOperationWhenCheckpointMatches(t *testing.T) {
	service := &fakeSCMService{}
	handler, err := NewSCMWorkflowHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(SCMWorkflowPayload{
		Operation: "source_state", Provider: "github", Owner: "acme", Repository: "repo",
		RunID: "run-1", TaskID: "task-1", JobID: "job-1", IdempotencyKey: "id-1",
		CompletedCheckpoint: &taskengine.Checkpoint{Step: "source_state", Token: "id-1"},
	})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected no service call when checkpoint matches, got %q", service.called)
	}
}

func TestSCMWorkflowHandlerExecutesWhenCheckpointStepDiffers(t *testing.T) {
	service := &fakeSCMService{}
	handler, err := NewSCMWorkflowHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(SCMWorkflowPayload{
		Operation: "source_state", Provider: "github", Owner: "acme", Repository: "repo",
		RunID: "run-1", TaskID: "task-1", JobID: "job-1", IdempotencyKey: "id-1",
		CompletedCheckpoint: &taskengine.Checkpoint{Step: "ensure_worktree", Token: "id-1"},
	})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if service.called != "source_state" {
		t.Fatalf("expected source_state call when checkpoint step differs, got %q", service.called)
	}
}

func TestSCMWorkflowHandlerSkipsOperationWhenResumeCheckpointMatches(t *testing.T) {
	service := &fakeSCMService{}
	handler, err := NewSCMWorkflowHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(SCMWorkflowPayload{
		Operation: "source_state", Provider: "github", Owner: "acme", Repository: "repo",
		RunID: "run-1", TaskID: "task-1", JobID: "job-1", IdempotencyKey: "id-1",
		ResumeCheckpoint: &taskengine.Checkpoint{Step: "source_state", Token: "id-1"},
	})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected no service call when resume checkpoint matches, got %q", service.called)
	}
}

func TestSCMWorkflowHandlerSkipsOperationWhenPersistedCheckpointMatches(t *testing.T) {
	service := &fakeSCMService{}
	store := &fakeCheckpointStore{loadedCheckpoint: &taskengine.Checkpoint{Step: "source_state", Token: "id-1"}}
	handler, err := NewSCMWorkflowHandlerWithCheckpointStore(service, store)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(SCMWorkflowPayload{Operation: "source_state", Provider: "github", Owner: "acme", Repository: "repo", RunID: "run-1", TaskID: "task-1", JobID: "job-1", IdempotencyKey: "id-1"})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected no service call when persisted checkpoint matches, got %q", service.called)
	}
}

func TestSCMWorkflowHandlerPersistsCheckpointAfterSuccess(t *testing.T) {
	service := &fakeSCMService{}
	store := &fakeCheckpointStore{}
	handler, err := NewSCMWorkflowHandlerWithCheckpointStore(service, store)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(SCMWorkflowPayload{Operation: "source_state", Provider: "github", Owner: "acme", Repository: "repo", RunID: "run-1", TaskID: "task-1", JobID: "job-1", IdempotencyKey: "id-1"})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if store.savedCheckpoint == nil {
		t.Fatalf("expected checkpoint persisted")
	}
	if store.savedCheckpoint.Step != "source_state" || store.savedCheckpoint.Token != "id-1" {
		t.Fatalf("expected source_state checkpoint, got %+v", store.savedCheckpoint)
	}
}

func TestSCMWorkflowHandlerReturnsSaveError(t *testing.T) {
	service := &fakeSCMService{}
	store := &fakeCheckpointStore{saveErr: errors.New("save failed")}
	handler, err := NewSCMWorkflowHandlerWithCheckpointStore(service, store)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(SCMWorkflowPayload{Operation: "source_state", Provider: "github", Owner: "acme", Repository: "repo", RunID: "run-1", TaskID: "task-1", JobID: "job-1", IdempotencyKey: "id-1"})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, Payload: payload}); err == nil {
		t.Fatalf("expected checkpoint save error")
	}
}


func TestSCMWorkflowHandlerRecordsExecutionJournalSkipped(t *testing.T) {
	service := &fakeSCMService{}
	store := &fakeCheckpointStore{loadedCheckpoint: &taskengine.Checkpoint{Step: "source_state", Token: "id-1"}}
	journal := &fakeExecutionJournal{}
	handler, err := NewSCMWorkflowHandlerWithReliability(service, store, journal)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(SCMWorkflowPayload{Operation: "source_state", Provider: "github", Owner: "acme", Repository: "repo", RunID: "run-1", TaskID: "task-1", JobID: "job-1", IdempotencyKey: "id-1"})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(journal.records) < 2 {
		t.Fatalf("expected at least 2 journal records, got %d", len(journal.records))
	}
	if journal.records[len(journal.records)-1].Status != taskengine.ExecutionStatusSkipped {
		t.Fatalf("expected last status skipped, got %q", journal.records[len(journal.records)-1].Status)
	}
}

func TestSCMWorkflowHandlerRecordsExecutionJournalFailure(t *testing.T) {
	service := &fakeSCMService{err: errors.New("boom")}
	journal := &fakeExecutionJournal{}
	handler, err := NewSCMWorkflowHandlerWithReliability(service, nil, journal)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(SCMWorkflowPayload{Operation: "source_state", Provider: "github", Owner: "acme", Repository: "repo", RunID: "run-1", TaskID: "task-1", JobID: "job-1", IdempotencyKey: "id-1"})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, Payload: payload}); err == nil {
		t.Fatalf("expected error")
	}
	if len(journal.records) < 2 {
		t.Fatalf("expected at least 2 journal records, got %d", len(journal.records))
	}
	if journal.records[len(journal.records)-1].Status != taskengine.ExecutionStatusFailed {
		t.Fatalf("expected last status failed, got %q", journal.records[len(journal.records)-1].Status)
	}
}

func TestSCMWorkflowHandlerIgnoresExecutionJournalWriteErrors(t *testing.T) {
	service := &fakeSCMService{}
	journal := &fakeExecutionJournal{err: errors.New("journal unavailable")}
	handler, err := NewSCMWorkflowHandlerWithReliability(service, nil, journal)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(SCMWorkflowPayload{Operation: "source_state", Provider: "github", Owner: "acme", Repository: "repo", RunID: "run-1", TaskID: "task-1", JobID: "job-1", IdempotencyKey: "id-1"})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindSCMWorkflow, Payload: payload}); err != nil {
		t.Fatalf("expected journal errors to be ignored, got: %v", err)
	}
	if service.called != "source_state" {
		t.Fatalf("expected source_state to execute despite journal errors, got %q", service.called)
	}
}
