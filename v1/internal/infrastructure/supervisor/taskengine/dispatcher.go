package taskengine

import (
	applicationsupervisor "agentic-orchestrator/internal/application/supervisor"
	"agentic-orchestrator/internal/application/taskengine"
	domainsupervisor "agentic-orchestrator/internal/domain/supervisor"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Dispatcher struct {
	scheduler *taskengine.Scheduler
}

func NewDispatcher(scheduler *taskengine.Scheduler) (*Dispatcher, error) {
	if scheduler == nil {
		return nil, fmt.Errorf("supervisor task dispatcher: scheduler is required")
	}
	return &Dispatcher{scheduler: scheduler}, nil
}

type agentWorkflowPayload struct {
	SessionID      string `json:"session_id"`
	Prompt         string `json:"prompt"`
	Provider       string `json:"provider"`
	Owner          string `json:"owner"`
	Repository     string `json:"repository"`
	RunID          string `json:"run_id"`
	TaskID         string `json:"task_id"`
	JobID          string `json:"job_id"`
	ProjectID      string `json:"project_id"`
	IdempotencyKey string `json:"idempotency_key"`
}

type scmWorkflowPayload struct {
	Operation        string `json:"operation"`
	Provider         string `json:"provider"`
	Owner            string `json:"owner"`
	Repository       string `json:"repository"`
	RunID            string `json:"run_id"`
	TaskID           string `json:"task_id"`
	JobID            string `json:"job_id"`
	ProjectID        string `json:"project_id"`
	IdempotencyKey   string `json:"idempotency_key"`
	PullRequestID    int    `json:"pull_request_number,omitempty"`
	MergeMethod      string `json:"merge_method,omitempty"`
	PullRequestTitle string `json:"pull_request_title,omitempty"`
	PullRequestBody  string `json:"pull_request_body,omitempty"`
	ReviewDecision   string `json:"review_decision,omitempty"`
	ReviewBody       string `json:"review_body,omitempty"`
}

func (dispatcher *Dispatcher) Dispatch(ctx context.Context, decision domainsupervisor.Decision) error {
	if dispatcher == nil || dispatcher.scheduler == nil {
		return fmt.Errorf("supervisor task dispatcher: scheduler is not configured")
	}
	switch decision.Action {
	case domainsupervisor.ActionStartTask:
		return dispatcher.enqueueAgentWorkflow(ctx, decision, "start")
	case domainsupervisor.ActionRequestRework:
		return dispatcher.enqueueAgentWorkflow(ctx, decision, "rework")
	case domainsupervisor.ActionMerge:
		return dispatcher.enqueueMergeWorkflow(ctx, decision)
	case domainsupervisor.ActionContinue, domainsupervisor.ActionRetry, domainsupervisor.ActionBlock, domainsupervisor.ActionEscalate, domainsupervisor.ActionRefuse:
		return nil
	default:
		return nil
	}
}

func (dispatcher *Dispatcher) enqueueAgentWorkflow(ctx context.Context, decision domainsupervisor.Decision, mode string) error {
	provider, owner, repository := providerOwnerRepositoryFromDecision(decision)
	if provider == "" || owner == "" || repository == "" {
		return nil
	}
	issueReference := strings.TrimSpace(decision.Metadata["issue_reference"])
	prompt := fmt.Sprintf("Supervisor requested %s action due to %s", mode, decision.Reason)
	if issueReference != "" {
		prompt = fmt.Sprintf("%s on issue %s", prompt, issueReference)
	}
	jobID := nextSupervisorJobID(decision, mode)
	idempotencyKey := nextSupervisorIdempotency(decision, mode)
	payload := agentWorkflowPayload{
		SessionID:      fmt.Sprintf("%s/%s", decision.CorrelationIDs.RunID, decision.CorrelationIDs.TaskID),
		Prompt:         prompt,
		Provider:       provider,
		Owner:          owner,
		Repository:     repository,
		RunID:          decision.CorrelationIDs.RunID,
		TaskID:         decision.CorrelationIDs.TaskID,
		JobID:          jobID,
		ProjectID:      decision.CorrelationIDs.ProjectID,
		IdempotencyKey: idempotencyKey,
	}
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode supervisor agent workflow payload: %w", err)
	}
	_, err = dispatcher.scheduler.Enqueue(ctx, taskengine.EnqueueRequest{
		Kind:           taskengine.JobKindAgentWorkflow,
		Payload:        encodedPayload,
		IdempotencyKey: idempotencyKey,
		CorrelationIDs: taskengine.CorrelationIDs{RunID: decision.CorrelationIDs.RunID, TaskID: decision.CorrelationIDs.TaskID, JobID: jobID, ProjectID: decision.CorrelationIDs.ProjectID},
	})
	if err != nil {
		return fmt.Errorf("enqueue supervisor agent workflow: %w", err)
	}
	return nil
}

func (dispatcher *Dispatcher) enqueueMergeWorkflow(ctx context.Context, decision domainsupervisor.Decision) error {
	provider, owner, repository := providerOwnerRepositoryFromDecision(decision)
	if provider == "" || owner == "" || repository == "" {
		return nil
	}
	pullRequestNumber, err := strconv.Atoi(strings.TrimSpace(decision.Metadata["pull_request_number"]))
	if err != nil || pullRequestNumber <= 0 {
		return nil
	}
	mergeMethod := strings.TrimSpace(decision.Metadata["merge_method"])
	if mergeMethod == "" {
		mergeMethod = "squash"
	}
	jobID := nextSupervisorJobID(decision, "merge")
	idempotencyKey := nextSupervisorIdempotency(decision, "merge")
	payload := scmWorkflowPayload{
		Operation:      "merge_pull_request",
		Provider:       provider,
		Owner:          owner,
		Repository:     repository,
		RunID:          decision.CorrelationIDs.RunID,
		TaskID:         decision.CorrelationIDs.TaskID,
		JobID:          jobID,
		ProjectID:      decision.CorrelationIDs.ProjectID,
		IdempotencyKey: idempotencyKey,
		PullRequestID:  pullRequestNumber,
		MergeMethod:    mergeMethod,
	}
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode supervisor merge payload: %w", err)
	}
	_, err = dispatcher.scheduler.Enqueue(ctx, taskengine.EnqueueRequest{
		Kind:           taskengine.JobKindSCMWorkflow,
		Payload:        encodedPayload,
		IdempotencyKey: idempotencyKey,
		CorrelationIDs: taskengine.CorrelationIDs{RunID: decision.CorrelationIDs.RunID, TaskID: decision.CorrelationIDs.TaskID, JobID: jobID, ProjectID: decision.CorrelationIDs.ProjectID},
	})
	if err != nil {
		return fmt.Errorf("enqueue supervisor merge workflow: %w", err)
	}
	return nil
}

func providerOwnerRepositoryFromDecision(decision domainsupervisor.Decision) (string, string, string) {
	provider := strings.TrimSpace(decision.Metadata["provider"])
	owner := strings.TrimSpace(decision.Metadata["owner"])
	repository := strings.TrimSpace(decision.Metadata["repository"])
	if provider != "" && owner != "" && repository != "" {
		return provider, owner, repository
	}
	source := strings.TrimSpace(decision.Metadata["source"])
	if source == "" {
		source = strings.TrimSpace(decision.Metadata["issue_reference"])
	}
	parts := strings.Split(strings.Trim(strings.TrimSpace(source), "/"), "/")
	if len(parts) < 2 {
		return provider, owner, repository
	}
	repo := parts[1]
	if hash := strings.Index(repo, "#"); hash >= 0 {
		repo = repo[:hash]
	}
	if provider == "" {
		provider = "github"
	}
	if owner == "" {
		owner = strings.TrimSpace(parts[0])
	}
	if repository == "" {
		repository = strings.TrimSpace(repo)
	}
	return provider, owner, repository
}

func nextSupervisorJobID(decision domainsupervisor.Decision, suffix string) string {
	return fmt.Sprintf("%s-supervisor-%s-%d", strings.TrimSpace(decision.CorrelationIDs.JobID), strings.TrimSpace(suffix), decision.OccurredAt.Unix())
}

func nextSupervisorIdempotency(decision domainsupervisor.Decision, suffix string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", strings.TrimSpace(decision.CorrelationIDs.RunID), strings.TrimSpace(decision.CorrelationIDs.TaskID), strings.TrimSpace(decision.CorrelationIDs.JobID), strings.TrimSpace(suffix), decision.OccurredAt.UTC().Format(time.RFC3339Nano))
}

var _ applicationsupervisor.DecisionDispatcher = (*Dispatcher)(nil)
