package resolvers

import (
	"agentic-orchestrator/internal/application/taskengine"
	"agentic-orchestrator/internal/interface/graphql/models"
	"context"
	"encoding/json"
	"fmt"
)

type scmWorkflowPayload struct {
	Operation        string `json:"operation"`
	Provider         string `json:"provider"`
	Owner            string `json:"owner"`
	Repository       string `json:"repository"`
	RunID            string `json:"run_id"`
	TaskID           string `json:"task_id"`
	JobID            string `json:"job_id"`
	IdempotencyKey   string `json:"idempotency_key"`
	WorktreePath     string `json:"worktree_path,omitempty"`
	BaseBranch       string `json:"base_branch,omitempty"`
	TargetBranch     string `json:"target_branch,omitempty"`
	PullRequestID    int    `json:"pull_request_number,omitempty"`
	MergeMethod      string `json:"merge_method,omitempty"`
	PullRequestTitle string `json:"pull_request_title,omitempty"`
	PullRequestBody  string `json:"pull_request_body,omitempty"`
	ReviewDecision   string `json:"review_decision,omitempty"`
	ReviewBody       string `json:"review_body,omitempty"`
}

func allSupportedSCMOperations() []models.SCMOperation {
	return []models.SCMOperation{
		models.SCMOperationSourceState,
		models.SCMOperationEnsureWorktree,
		models.SCMOperationSyncWorktree,
		models.SCMOperationCleanupWorktree,
		models.SCMOperationEnsureBranch,
		models.SCMOperationSyncBranch,
		models.SCMOperationUpsertPullRequest,
		models.SCMOperationGetPullRequest,
		models.SCMOperationSubmitReview,
		models.SCMOperationCheckMergeReadiness,
		models.SCMOperationMergePullRequest,
	}
}

func toSCMOperation(value models.SCMOperation) string {
	switch value {
	case models.SCMOperationSourceState:
		return "source_state"
	case models.SCMOperationEnsureWorktree:
		return "ensure_worktree"
	case models.SCMOperationSyncWorktree:
		return "sync_worktree"
	case models.SCMOperationCleanupWorktree:
		return "cleanup_worktree"
	case models.SCMOperationEnsureBranch:
		return "ensure_branch"
	case models.SCMOperationSyncBranch:
		return "sync_branch"
	case models.SCMOperationUpsertPullRequest:
		return "upsert_pull_request"
	case models.SCMOperationGetPullRequest:
		return "get_pull_request"
	case models.SCMOperationSubmitReview:
		return "submit_review"
	case models.SCMOperationCheckMergeReadiness:
		return "check_merge_readiness"
	case models.SCMOperationMergePullRequest:
		return "merge_pull_request"
	default:
		return ""
	}
}

func toSCMProvider(value models.SCMProvider) string {
	switch value {
	case models.SCMProviderGithub:
		return "github"
	default:
		return ""
	}
}

func toSCMMergeMethod(value models.SCMMergeMethod) string {
	switch value {
	case models.SCMMergeMethodMerge:
		return "merge"
	case models.SCMMergeMethodSquash:
		return "squash"
	case models.SCMMergeMethodRebase:
		return "rebase"
	default:
		return ""
	}
}

func toSCMReviewDecision(value models.SCMReviewDecision) string {
	switch value {
	case models.SCMReviewDecisionApprove:
		return "approve"
	case models.SCMReviewDecisionRequestChanges:
		return "request_changes"
	case models.SCMReviewDecisionComment:
		return "comment"
	default:
		return ""
	}
}

func enqueueSCMWorkflow(ctx context.Context, resolver *Resolver, input models.EnqueueSCMWorkflowInput) (models.EnqueueSCMWorkflowResult, error) {
	if resolver == nil || resolver.TaskScheduler == nil {
		return models.GraphError{Code: models.GraphErrorCodeUnavailable, Message: "task scheduler is not configured"}, nil
	}
	operation := toSCMOperation(input.Operation)
	if operation == "" {
		return models.GraphError{Code: models.GraphErrorCodeValidation, Message: "operation is invalid", Field: strPtr("operation")}, nil
	}
	provider := toSCMProvider(input.Provider)
	if provider == "" {
		return models.GraphError{Code: models.GraphErrorCodeValidation, Message: "provider is invalid", Field: strPtr("provider")}, nil
	}

	payload := scmWorkflowPayload{
		Operation:      operation,
		Provider:       provider,
		Owner:          input.Owner,
		Repository:     input.Repository,
		RunID:          input.RunID,
		TaskID:         input.TaskID,
		JobID:          input.JobID,
		IdempotencyKey: input.IdempotencyKey,
	}
	if input.WorktreePath != nil {
		payload.WorktreePath = *input.WorktreePath
	}
	if input.BaseBranch != nil {
		payload.BaseBranch = *input.BaseBranch
	}
	if input.TargetBranch != nil {
		payload.TargetBranch = *input.TargetBranch
	}
	if input.PullRequestNumber != nil {
		payload.PullRequestID = int(*input.PullRequestNumber)
	}
	if input.MergeMethod != nil {
		mergeMethod := toSCMMergeMethod(*input.MergeMethod)
		if mergeMethod == "" {
			return models.GraphError{Code: models.GraphErrorCodeValidation, Message: "mergeMethod is invalid", Field: strPtr("mergeMethod")}, nil
		}
		payload.MergeMethod = mergeMethod
	}
	if input.PullRequestTitle != nil {
		payload.PullRequestTitle = *input.PullRequestTitle
	}
	if input.PullRequestBody != nil {
		payload.PullRequestBody = *input.PullRequestBody
	}
	if input.ReviewDecision != nil {
		reviewDecision := toSCMReviewDecision(*input.ReviewDecision)
		if reviewDecision == "" {
			return models.GraphError{Code: models.GraphErrorCodeValidation, Message: "reviewDecision is invalid", Field: strPtr("reviewDecision")}, nil
		}
		payload.ReviewDecision = reviewDecision
	}
	if input.ReviewBody != nil {
		payload.ReviewBody = *input.ReviewBody
	}

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return graphErrorFromError(fmt.Errorf("encode scm workflow payload: %w", err)), nil
	}

	result, err := resolver.TaskScheduler.Enqueue(ctx, taskengine.EnqueueRequest{
		Kind:           taskengine.JobKindSCMWorkflow,
		Payload:        encodedPayload,
		IdempotencyKey: input.IdempotencyKey,
		CorrelationIDs: taskengine.CorrelationIDs{RunID: input.RunID, TaskID: input.TaskID, JobID: input.JobID},
	})
	if err != nil {
		return graphErrorFromError(fmt.Errorf("enqueue scm workflow: %w", err)), nil
	}
	return models.EnqueueSCMWorkflowSuccess{QueueTaskID: result.QueueTaskID, Duplicate: result.Duplicate}, nil
}

func strPtr(value string) *string {
	return &value
}
