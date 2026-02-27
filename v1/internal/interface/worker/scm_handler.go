package worker

import (
	applicationscm "agentic-orchestrator/internal/application/scm"
	"agentic-orchestrator/internal/application/taskengine"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type SCMWorkflowPayload struct {
	Operation       string `json:"operation"`
	Provider        string `json:"provider"`
	Owner           string `json:"owner"`
	Repository      string `json:"repository"`
	RunID           string `json:"run_id"`
	TaskID          string `json:"task_id"`
	JobID           string `json:"job_id"`
	IdempotencyKey  string `json:"idempotency_key"`
	WorktreePath    string `json:"worktree_path,omitempty"`
	BaseBranch      string `json:"base_branch,omitempty"`
	TargetBranch    string `json:"target_branch,omitempty"`
	PullRequestID   int    `json:"pull_request_number,omitempty"`
	PullRequestTitle string `json:"pull_request_title,omitempty"`
	PullRequestBody  string `json:"pull_request_body,omitempty"`
	ReviewDecision  string `json:"review_decision,omitempty"`
	ReviewBody      string `json:"review_body,omitempty"`
}

type scmService interface {
	SourceState(ctx context.Context, request applicationscm.SourceStateRequest) (domainscm.SourceState, error)
	EnsureWorktree(ctx context.Context, request applicationscm.EnsureWorktreeRequest) (domainscm.WorktreeState, error)
	SyncWorktree(ctx context.Context, request applicationscm.SyncWorktreeRequest) (domainscm.WorktreeState, error)
	CleanupWorktree(ctx context.Context, request applicationscm.CleanupWorktreeRequest) error
	EnsureBranch(ctx context.Context, request applicationscm.EnsureBranchRequest) (domainscm.BranchState, error)
	SyncBranch(ctx context.Context, request applicationscm.SyncBranchRequest) (domainscm.BranchState, error)
	CreateOrUpdatePullRequest(ctx context.Context, request applicationscm.CreateOrUpdatePullRequestRequest) (domainscm.PullRequestState, error)
	GetPullRequest(ctx context.Context, request applicationscm.GetPullRequestRequest) (domainscm.PullRequestState, error)
	SubmitReview(ctx context.Context, request applicationscm.SubmitReviewRequest) (domainscm.ReviewDecision, error)
	CheckMergeReadiness(ctx context.Context, request applicationscm.CheckMergeReadinessRequest) (domainscm.MergeReadiness, error)
}

type SCMWorkflowHandler struct {
	service scmService
}

func NewSCMWorkflowHandler(service scmService) (*SCMWorkflowHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("scm service is required")
	}
	return &SCMWorkflowHandler{service: service}, nil
}

func (handler *SCMWorkflowHandler) Handle(ctx context.Context, job taskengine.Job) error {
	var payload SCMWorkflowPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("decode scm workflow payload: %w", err)
	}
	repository := domainscm.Repository{Provider: payload.Provider, Owner: payload.Owner, Name: payload.Repository}
	metadata := applicationscm.Metadata{CorrelationIDs: taskengine.CorrelationIDs{RunID: payload.RunID, TaskID: payload.TaskID, JobID: payload.JobID}, IdempotencyKey: payload.IdempotencyKey}

	switch strings.TrimSpace(payload.Operation) {
	case "source_state":
		_, err := handler.service.SourceState(ctx, applicationscm.SourceStateRequest{Repository: repository, Metadata: metadata})
		return err
	case "ensure_worktree":
		_, err := handler.service.EnsureWorktree(ctx, applicationscm.EnsureWorktreeRequest{Repository: repository, Spec: domainscm.WorktreeSpec{BaseBranch: payload.BaseBranch, TargetBranch: payload.TargetBranch, Path: payload.WorktreePath}, Metadata: metadata})
		return err
	case "sync_worktree":
		_, err := handler.service.SyncWorktree(ctx, applicationscm.SyncWorktreeRequest{Repository: repository, Path: payload.WorktreePath, Metadata: metadata})
		return err
	case "cleanup_worktree":
		return handler.service.CleanupWorktree(ctx, applicationscm.CleanupWorktreeRequest{Repository: repository, Path: payload.WorktreePath, Metadata: metadata})
	case "ensure_branch":
		_, err := handler.service.EnsureBranch(ctx, applicationscm.EnsureBranchRequest{Repository: repository, Spec: domainscm.BranchSpec{BaseBranch: payload.BaseBranch, TargetBranch: payload.TargetBranch}, Metadata: metadata})
		return err
	case "sync_branch":
		_, err := handler.service.SyncBranch(ctx, applicationscm.SyncBranchRequest{Repository: repository, BranchName: payload.TargetBranch, Metadata: metadata})
		return err
	case "upsert_pull_request":
		_, err := handler.service.CreateOrUpdatePullRequest(ctx, applicationscm.CreateOrUpdatePullRequestRequest{Spec: domainscm.PullRequestSpec{Repository: repository, SourceBranch: payload.TargetBranch, TargetBranch: payload.BaseBranch, Title: payload.PullRequestTitle, Body: payload.PullRequestBody}, Metadata: metadata})
		return err
	case "get_pull_request":
		_, err := handler.service.GetPullRequest(ctx, applicationscm.GetPullRequestRequest{Repository: repository, PullRequestNumber: payload.PullRequestID, Metadata: metadata})
		return err
	case "submit_review":
		_, err := handler.service.SubmitReview(ctx, applicationscm.SubmitReviewRequest{Spec: domainscm.ReviewSpec{Repository: repository, PullRequestNumber: payload.PullRequestID, Decision: domainscm.ReviewDecision(payload.ReviewDecision), Body: payload.ReviewBody}, Metadata: metadata})
		return err
	case "check_merge_readiness":
		_, err := handler.service.CheckMergeReadiness(ctx, applicationscm.CheckMergeReadinessRequest{Repository: repository, PullRequestNumber: payload.PullRequestID, Metadata: metadata})
		return err
	default:
		return fmt.Errorf("unsupported scm operation %q", payload.Operation)
	}
}
