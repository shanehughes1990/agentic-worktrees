package worker

import (
	applicationscm "agentic-orchestrator/internal/application/scm"
	"agentic-orchestrator/internal/application/taskengine"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type SCMWorkflowPayload struct {
	Operation             string                 `json:"operation"`
	Provider              string                 `json:"provider"`
	Owner                 string                 `json:"owner"`
	Repository            string                 `json:"repository"`
	RunID                 string                 `json:"run_id"`
	TaskID                string                 `json:"task_id"`
	JobID                 string                 `json:"job_id"`
	ProjectID             string                 `json:"project_id"`
	IdempotencyKey        string                 `json:"idempotency_key"`
	WorktreePath          string                 `json:"worktree_path,omitempty"`
	BaseBranch            string                 `json:"base_branch,omitempty"`
	TargetBranch          string                 `json:"target_branch,omitempty"`
	SyncStrategy          string                 `json:"sync_strategy,omitempty"`
	PullRequestID         int                    `json:"pull_request_number,omitempty"`
	MergeMethod           string                 `json:"merge_method,omitempty"`
	PullRequestTitle      string                 `json:"pull_request_title,omitempty"`
	PullRequestBody       string                 `json:"pull_request_body,omitempty"`
	ReviewDecision        string                 `json:"review_decision,omitempty"`
	ReviewBody            string                 `json:"review_body,omitempty"`
	ResumeCheckpoint      *taskengine.Checkpoint `json:"resume_checkpoint,omitempty"`
	ResumeCheckpointStep  string                 `json:"resume_checkpoint_step,omitempty"`
	ResumeCheckpointToken string                 `json:"resume_checkpoint_token,omitempty"`
	CompletedCheckpoint   *taskengine.Checkpoint `json:"completed_checkpoint,omitempty"`
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
	MergePullRequest(ctx context.Context, request applicationscm.MergePullRequestRequest) (domainscm.PullRequestState, error)
}

type SCMWorkflowHandler struct {
	service           scmService
	checkpointStore   taskengine.CheckpointStore
	executionJournal  taskengine.ExecutionJournal
	supervisorService supervisorSignalService
}

func NewSCMWorkflowHandler(service scmService) (*SCMWorkflowHandler, error) {
	return NewSCMWorkflowHandlerWithCheckpointStore(service, nil)
}

func NewSCMWorkflowHandlerWithCheckpointStore(service scmService, checkpointStore taskengine.CheckpointStore) (*SCMWorkflowHandler, error) {
	return NewSCMWorkflowHandlerWithReliability(service, checkpointStore, nil)
}

func NewSCMWorkflowHandlerWithReliability(service scmService, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal) (*SCMWorkflowHandler, error) {
	return newSCMWorkflowHandler(service, checkpointStore, executionJournal, nil)
}

func NewSCMWorkflowHandlerWithSupervisor(service scmService, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal, supervisorService supervisorSignalService) (*SCMWorkflowHandler, error) {
	return newSCMWorkflowHandler(service, checkpointStore, executionJournal, supervisorService)
}

func newSCMWorkflowHandler(service scmService, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal, supervisorService supervisorSignalService) (*SCMWorkflowHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("scm service is required")
	}
	return &SCMWorkflowHandler{service: service, checkpointStore: checkpointStore, executionJournal: executionJournal, supervisorService: supervisorService}, nil
}

func (handler *SCMWorkflowHandler) Handle(ctx context.Context, job taskengine.Job) error {
	var payload SCMWorkflowPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("decode scm workflow payload: %w", err)
	}
	operation := strings.TrimSpace(payload.Operation)
	idempotencyKey := strings.TrimSpace(payload.IdempotencyKey)
	repository := domainscm.Repository{Provider: payload.Provider, Owner: payload.Owner, Name: payload.Repository}
	metadata := applicationscm.Metadata{CorrelationIDs: taskengine.CorrelationIDs{RunID: payload.RunID, TaskID: payload.TaskID, JobID: payload.JobID, ProjectID: payload.ProjectID}, IdempotencyKey: idempotencyKey}

	handler.safeRecordExecution(ctx, metadata.CorrelationIDs, job.Kind, idempotencyKey, operation, taskengine.ExecutionStatusRunning, "")

	retryCheckpoint := taskengine.RetryCheckpointContract{
		ResumeCheckpoint:      payload.ResumeCheckpoint,
		CompletedCheckpoint:   payload.CompletedCheckpoint,
		ResumeCheckpointStep:  payload.ResumeCheckpointStep,
		ResumeCheckpointToken: payload.ResumeCheckpointToken,
	}
	effectiveCheckpoint := retryCheckpoint.Checkpoint()
	if handler.checkpointStore != nil && idempotencyKey != "" {
		persistedCheckpoint, err := handler.checkpointStore.Load(ctx, idempotencyKey)
		if err != nil {
			handler.safeRecordExecution(ctx, metadata.CorrelationIDs, job.Kind, idempotencyKey, operation, taskengine.ExecutionStatusFailed, err.Error())
			return fmt.Errorf("load persisted checkpoint: %w", err)
		}
		if persistedCheckpoint != nil {
			effectiveCheckpoint = persistedCheckpoint
		}
	}
	if taskengine.CheckpointMatches(effectiveCheckpoint, operation, idempotencyKey) {
		handler.safeRecordExecution(ctx, metadata.CorrelationIDs, job.Kind, idempotencyKey, operation, taskengine.ExecutionStatusSkipped, "")
		return nil
	}

	var executionErr error
	switch operation {
	case "source_state":
		_, executionErr = handler.service.SourceState(ctx, applicationscm.SourceStateRequest{Repository: repository, Metadata: metadata})
	case "ensure_worktree":
		_, executionErr = handler.service.EnsureWorktree(ctx, applicationscm.EnsureWorktreeRequest{Repository: repository, Spec: domainscm.WorktreeSpec{BaseBranch: payload.BaseBranch, TargetBranch: payload.TargetBranch, Path: payload.WorktreePath, SyncStrategy: domainscm.SyncStrategy(payload.SyncStrategy)}, Metadata: metadata})
	case "sync_worktree":
		_, executionErr = handler.service.SyncWorktree(ctx, applicationscm.SyncWorktreeRequest{Repository: repository, Path: payload.WorktreePath, Metadata: metadata})
	case "cleanup_worktree":
		executionErr = handler.service.CleanupWorktree(ctx, applicationscm.CleanupWorktreeRequest{Repository: repository, Path: payload.WorktreePath, Metadata: metadata})
	case "ensure_branch":
		_, executionErr = handler.service.EnsureBranch(ctx, applicationscm.EnsureBranchRequest{Repository: repository, Spec: domainscm.BranchSpec{BaseBranch: payload.BaseBranch, TargetBranch: payload.TargetBranch}, Metadata: metadata})
	case "sync_branch":
		_, executionErr = handler.service.SyncBranch(ctx, applicationscm.SyncBranchRequest{Repository: repository, BranchName: payload.TargetBranch, Metadata: metadata})
	case "upsert_pull_request":
		_, executionErr = handler.service.CreateOrUpdatePullRequest(ctx, applicationscm.CreateOrUpdatePullRequestRequest{Spec: domainscm.PullRequestSpec{Repository: repository, SourceBranch: payload.TargetBranch, TargetBranch: payload.BaseBranch, Title: payload.PullRequestTitle, Body: payload.PullRequestBody}, Metadata: metadata})
	case "get_pull_request":
		_, executionErr = handler.service.GetPullRequest(ctx, applicationscm.GetPullRequestRequest{Repository: repository, PullRequestNumber: payload.PullRequestID, Metadata: metadata})
	case "submit_review":
		_, executionErr = handler.service.SubmitReview(ctx, applicationscm.SubmitReviewRequest{Spec: domainscm.ReviewSpec{Repository: repository, PullRequestNumber: payload.PullRequestID, Decision: domainscm.ReviewDecision(payload.ReviewDecision), Body: payload.ReviewBody}, Metadata: metadata})
	case "check_merge_readiness":
		readiness, readinessErr := handler.service.CheckMergeReadiness(ctx, applicationscm.CheckMergeReadinessRequest{Repository: repository, PullRequestNumber: payload.PullRequestID, Metadata: metadata})
		executionErr = readinessErr
		if readinessErr == nil {
			handler.safeSupervisorPRChecksEvaluated(ctx, metadata.CorrelationIDs, repository, payload.PullRequestID, readiness)
			if readiness.CanMerge {
				handler.safeSupervisorPRMergeRequested(ctx, metadata.CorrelationIDs, repository, payload.PullRequestID, payload.MergeMethod)
			}
		}
	case "merge_pull_request":
		_, executionErr = handler.service.MergePullRequest(ctx, applicationscm.MergePullRequestRequest{Spec: domainscm.MergePullRequestSpec{Repository: repository, PullRequestNumber: payload.PullRequestID, Method: domainscm.MergeMethod(payload.MergeMethod)}, Metadata: metadata})
	default:
		return fmt.Errorf("unsupported scm operation %q", payload.Operation)
	}
	if executionErr != nil {
		handler.safeRecordExecution(ctx, metadata.CorrelationIDs, job.Kind, idempotencyKey, operation, taskengine.ExecutionStatusFailed, executionErr.Error())
		return executionErr
	}
	if handler.checkpointStore != nil && idempotencyKey != "" {
		if err := handler.checkpointStore.Save(ctx, idempotencyKey, taskengine.Checkpoint{Step: operation, Token: idempotencyKey}); err != nil {
			handler.safeRecordExecution(ctx, metadata.CorrelationIDs, job.Kind, idempotencyKey, operation, taskengine.ExecutionStatusFailed, err.Error())
			return fmt.Errorf("persist completed checkpoint: %w", err)
		}
		handler.safeSupervisorCheckpoint(ctx, metadata.CorrelationIDs, job.Kind, idempotencyKey, operation)
	}
	handler.safeRecordExecution(ctx, metadata.CorrelationIDs, job.Kind, idempotencyKey, operation, taskengine.ExecutionStatusSucceeded, "")
	return nil
}

func (handler *SCMWorkflowHandler) safeRecordExecution(ctx context.Context, correlationIDs taskengine.CorrelationIDs, kind taskengine.JobKind, idempotencyKey string, step string, status taskengine.ExecutionStatus, errorMessage string) {
	record, err := handler.recordExecution(ctx, correlationIDs, kind, idempotencyKey, step, status, errorMessage)
	if err != nil {
		return
	}
	handler.safeSupervisorExecution(ctx, record)
}

func (handler *SCMWorkflowHandler) recordExecution(ctx context.Context, correlationIDs taskengine.CorrelationIDs, kind taskengine.JobKind, idempotencyKey string, step string, status taskengine.ExecutionStatus, errorMessage string) (taskengine.ExecutionRecord, error) {
	record := taskengine.ExecutionRecord{
		RunID:          correlationIDs.RunID,
		TaskID:         correlationIDs.TaskID,
		JobID:          correlationIDs.JobID,
		ProjectID:      correlationIDs.ProjectID,
		JobKind:        kind,
		IdempotencyKey: idempotencyKey,
		Step:           step,
		Status:         status,
		ErrorMessage:   strings.TrimSpace(errorMessage),
		UpdatedAt:      time.Now().UTC(),
	}
	if handler == nil || handler.executionJournal == nil {
		return record, nil
	}
	if err := handler.executionJournal.Upsert(ctx, record); err != nil {
		return taskengine.ExecutionRecord{}, fmt.Errorf("record execution journal: %w", err)
	}
	return record, nil
}

func (handler *SCMWorkflowHandler) safeSupervisorExecution(ctx context.Context, record taskengine.ExecutionRecord) {
	if handler == nil || handler.supervisorService == nil {
		return
	}
	_, _ = handler.supervisorService.OnExecution(ctx, record, 0, 0)
}

func (handler *SCMWorkflowHandler) safeSupervisorCheckpoint(ctx context.Context, correlation taskengine.CorrelationIDs, kind taskengine.JobKind, idempotencyKey string, step string) {
	if handler == nil || handler.supervisorService == nil {
		return
	}
	_, _ = handler.supervisorService.OnCheckpointSaved(ctx, correlation, kind, idempotencyKey, step)
}

func (handler *SCMWorkflowHandler) safeSupervisorPRChecksEvaluated(ctx context.Context, correlation taskengine.CorrelationIDs, repository domainscm.Repository, pullRequestNumber int, readiness domainscm.MergeReadiness) {
	if handler == nil || handler.supervisorService == nil {
		return
	}
	_, _ = handler.supervisorService.OnPRChecksEvaluated(ctx, correlation, repository.Provider, repository.Owner, repository.Name, pullRequestNumber, readiness.CanMerge, readiness.Reason)
}

func (handler *SCMWorkflowHandler) safeSupervisorPRMergeRequested(ctx context.Context, correlation taskengine.CorrelationIDs, repository domainscm.Repository, pullRequestNumber int, mergeMethod string) {
	if handler == nil || handler.supervisorService == nil {
		return
	}
	_, _ = handler.supervisorService.OnPRMergeRequested(ctx, correlation, repository.Provider, repository.Owner, repository.Name, pullRequestNumber, mergeMethod)
}
