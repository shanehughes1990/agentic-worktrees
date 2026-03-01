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

func enqueueSCMWorkflow(ctx context.Context, resolver *Resolver, input models.EnqueueSCMWorkflowInput) (*models.EnqueueSCMWorkflowResult, error) {
	if resolver == nil || resolver.TaskScheduler == nil {
		return nil, fmt.Errorf("task scheduler is not configured")
	}
	payload := scmWorkflowPayload{
		Operation:      input.Operation,
		Provider:       input.Provider,
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
		payload.MergeMethod = *input.MergeMethod
	}
	if input.PullRequestTitle != nil {
		payload.PullRequestTitle = *input.PullRequestTitle
	}
	if input.PullRequestBody != nil {
		payload.PullRequestBody = *input.PullRequestBody
	}
	if input.ReviewDecision != nil {
		payload.ReviewDecision = *input.ReviewDecision
	}
	if input.ReviewBody != nil {
		payload.ReviewBody = *input.ReviewBody
	}

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode scm workflow payload: %w", err)
	}

	result, err := resolver.TaskScheduler.Enqueue(ctx, taskengine.EnqueueRequest{
		Kind:           taskengine.JobKindSCMWorkflow,
		Payload:        encodedPayload,
		IdempotencyKey: input.IdempotencyKey,
		CorrelationIDs: taskengine.CorrelationIDs{RunID: input.RunID, TaskID: input.TaskID, JobID: input.JobID},
	})
	if err != nil {
		return nil, fmt.Errorf("enqueue scm workflow: %w", err)
	}
	return &models.EnqueueSCMWorkflowResult{QueueTaskID: result.QueueTaskID, Duplicate: result.Duplicate}, nil
}
