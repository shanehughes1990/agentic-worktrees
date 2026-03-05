package worker

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	applicationscm "agentic-orchestrator/internal/application/scm"
	"agentic-orchestrator/internal/application/taskengine"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
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
	RepositoryPath          string                 `json:"repository_path,omitempty"`
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

type SCMRuntimeService interface {
	SourceState(ctx context.Context, request applicationscm.SourceStateRequest) (domainscm.SourceState, error)
	EnsureRepository(ctx context.Context, request applicationscm.EnsureRepositoryRequest) (domainscm.RepositoryState, error)
	SyncRepository(ctx context.Context, request applicationscm.SyncRepositoryRequest) (domainscm.RepositoryState, error)
	CleanupRepository(ctx context.Context, request applicationscm.CleanupRepositoryRequest) error
	EnsureBranch(ctx context.Context, request applicationscm.EnsureBranchRequest) (domainscm.BranchState, error)
	SyncBranch(ctx context.Context, request applicationscm.SyncBranchRequest) (domainscm.BranchState, error)
	CreateOrUpdatePullRequest(ctx context.Context, request applicationscm.CreateOrUpdatePullRequestRequest) (domainscm.PullRequestState, error)
	GetPullRequest(ctx context.Context, request applicationscm.GetPullRequestRequest) (domainscm.PullRequestState, error)
	SubmitReview(ctx context.Context, request applicationscm.SubmitReviewRequest) (domainscm.ReviewDecision, error)
	CheckMergeReadiness(ctx context.Context, request applicationscm.CheckMergeReadinessRequest) (domainscm.MergeReadiness, error)
	MergePullRequest(ctx context.Context, request applicationscm.MergePullRequestRequest) (domainscm.PullRequestState, error)
}

type projectSetupLookup interface {
	GetProjectSetup(ctx context.Context, projectID string) (*applicationcontrolplane.ProjectSetup, error)
}

type scmServiceFactoryFunc func(ctx context.Context, projectID string, scm applicationcontrolplane.ProjectSCM, repository applicationcontrolplane.ProjectRepository) (SCMRuntimeService, error)

type scmRuntimeResolver interface {
	Resolve(ctx context.Context, payload SCMWorkflowPayload) (SCMRuntimeService, domainscm.Repository, error)
}

type staticSCMRuntimeResolver struct {
	service SCMRuntimeService
}

func (resolver *staticSCMRuntimeResolver) Resolve(ctx context.Context, payload SCMWorkflowPayload) (SCMRuntimeService, domainscm.Repository, error) {
	_ = ctx
	if resolver == nil || resolver.service == nil {
		return nil, domainscm.Repository{}, fmt.Errorf("scm service is required")
	}
	return resolver.service, domainscm.Repository{Provider: payload.Provider, Owner: payload.Owner, Name: payload.Repository}, nil
}

type projectSetupSCMRuntimeResolver struct {
	projectRepository projectSetupLookup
	serviceFactory    scmServiceFactoryFunc
}

func (resolver *projectSetupSCMRuntimeResolver) Resolve(ctx context.Context, payload SCMWorkflowPayload) (SCMRuntimeService, domainscm.Repository, error) {
	if resolver == nil || resolver.projectRepository == nil {
		return nil, domainscm.Repository{}, fmt.Errorf("project setup repository is required")
	}
	if resolver.serviceFactory == nil {
		return nil, domainscm.Repository{}, fmt.Errorf("scm service factory is required")
	}
	projectID := strings.TrimSpace(payload.ProjectID)
	if projectID == "" {
		return nil, domainscm.Repository{}, fmt.Errorf("project_id is required for scm workflow execution")
	}
	setup, err := resolver.projectRepository.GetProjectSetup(ctx, projectID)
	if err != nil {
		return nil, domainscm.Repository{}, fmt.Errorf("load project setup: %w", err)
	}
	if setup == nil {
		return nil, domainscm.Repository{}, fmt.Errorf("project setup not found for project_id %q", projectID)
	}
	repositoryConfig, err := primaryProjectRepository(setup.Repositories)
	if err != nil {
		return nil, domainscm.Repository{}, err
	}
	scmConfig, err := projectSCMByID(setup.SCMs, repositoryConfig.SCMID)
	if err != nil {
		return nil, domainscm.Repository{}, err
	}
	owner, repositoryName, err := ownerRepositoryFromURL(repositoryConfig.RepositoryURL)
	if err != nil {
		return nil, domainscm.Repository{}, err
	}
	service, err := resolver.serviceFactory(ctx, projectID, scmConfig, repositoryConfig)
	if err != nil {
		return nil, domainscm.Repository{}, err
	}
	return service, domainscm.Repository{Provider: scmConfig.SCMProvider, Owner: owner, Name: repositoryName}, nil
}

func primaryProjectRepository(repositories []applicationcontrolplane.ProjectRepository) (applicationcontrolplane.ProjectRepository, error) {
	if len(repositories) == 0 {
		return applicationcontrolplane.ProjectRepository{}, fmt.Errorf("project setup requires at least one repository")
	}
	for _, repository := range repositories {
		if repository.IsPrimary {
			return repository, nil
		}
	}
	return repositories[0], nil
}

func projectSCMByID(scms []applicationcontrolplane.ProjectSCM, scmID string) (applicationcontrolplane.ProjectSCM, error) {
	trimmedSCMID := strings.TrimSpace(scmID)
	if trimmedSCMID == "" {
		return applicationcontrolplane.ProjectSCM{}, fmt.Errorf("repository scm_id is required")
	}
	for _, scm := range scms {
		if strings.TrimSpace(scm.SCMID) == trimmedSCMID {
			return scm, nil
		}
	}
	return applicationcontrolplane.ProjectSCM{}, fmt.Errorf("project scm not found for scm_id %q", trimmedSCMID)
}

func ownerRepositoryFromURL(repositoryURL string) (string, string, error) {
	trimmedURL := strings.TrimSpace(repositoryURL)
	if trimmedURL == "" {
		return "", "", fmt.Errorf("project repository_url is required")
	}
	if strings.HasPrefix(trimmedURL, "git@") {
		parts := strings.SplitN(trimmedURL, ":", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("project repository_url %q is invalid", trimmedURL)
		}
		return ownerRepositoryFromPath(parts[1], trimmedURL)
	}
	parsedURL, err := url.Parse(trimmedURL)
	if err != nil || parsedURL.Host == "" {
		return "", "", fmt.Errorf("project repository_url %q is invalid", trimmedURL)
	}
	return ownerRepositoryFromPath(parsedURL.Path, trimmedURL)
}

func ownerRepositoryFromPath(pathValue string, rawURL string) (string, string, error) {
	trimmedPath := strings.Trim(strings.TrimSpace(pathValue), "/")
	parts := strings.Split(trimmedPath, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("project repository_url %q must include owner and repository", rawURL)
	}
	owner := strings.TrimSpace(parts[0])
	repository := strings.TrimSuffix(strings.TrimSpace(parts[1]), ".git")
	if owner == "" || repository == "" {
		return "", "", fmt.Errorf("project repository_url %q must include owner and repository", rawURL)
	}
	return owner, repository, nil
}

type SCMWorkflowHandler struct {
	runtimeResolver   scmRuntimeResolver
	checkpointStore   taskengine.CheckpointStore
	executionJournal  taskengine.ExecutionJournal
}

func NewSCMWorkflowHandler(service SCMRuntimeService) (*SCMWorkflowHandler, error) {
	return newSCMWorkflowHandler(&staticSCMRuntimeResolver{service: service}, nil, nil)
}

func NewSCMWorkflowHandlerWithCheckpointStore(service SCMRuntimeService, checkpointStore taskengine.CheckpointStore) (*SCMWorkflowHandler, error) {
	return newSCMWorkflowHandler(&staticSCMRuntimeResolver{service: service}, checkpointStore, nil)
}

func NewSCMWorkflowHandlerWithReliability(service SCMRuntimeService, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal) (*SCMWorkflowHandler, error) {
	return newSCMWorkflowHandler(&staticSCMRuntimeResolver{service: service}, checkpointStore, executionJournal)
}
func NewSCMWorkflowHandlerWithProjectSetup(projectRepository projectSetupLookup, serviceFactory scmServiceFactoryFunc, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal) (*SCMWorkflowHandler, error) {
	return newSCMWorkflowHandler(&projectSetupSCMRuntimeResolver{projectRepository: projectRepository, serviceFactory: serviceFactory}, checkpointStore, executionJournal)
}

func newSCMWorkflowHandler(runtimeResolver scmRuntimeResolver, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal) (*SCMWorkflowHandler, error) {
	if runtimeResolver == nil {
		return nil, fmt.Errorf("scm runtime resolver is required")
	}
	return &SCMWorkflowHandler{runtimeResolver: runtimeResolver, checkpointStore: checkpointStore, executionJournal: executionJournal}, nil
}

func (handler *SCMWorkflowHandler) Handle(ctx context.Context, job taskengine.Job) error {
	var payload SCMWorkflowPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("decode scm workflow payload: %w", err)
	}
	operation := strings.TrimSpace(payload.Operation)
	idempotencyKey := strings.TrimSpace(payload.IdempotencyKey)
	repositoryPath := scopedRepositoryPath(payload.ProjectID, payload.RepositoryPath)
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

	service, repository, err := handler.runtimeResolver.Resolve(ctx, payload)
	if err != nil {
		handler.safeRecordExecution(ctx, metadata.CorrelationIDs, job.Kind, idempotencyKey, operation, taskengine.ExecutionStatusFailed, err.Error())
		return err
	}

	var executionErr error
	switch operation {
	case "source_state":
		_, executionErr = service.SourceState(ctx, applicationscm.SourceStateRequest{Repository: repository, Metadata: metadata})
	case "ensure_repository":
		_, executionErr = service.EnsureRepository(ctx, applicationscm.EnsureRepositoryRequest{Repository: repository, Spec: domainscm.RepositorySpec{BaseBranch: payload.BaseBranch, TargetBranch: payload.TargetBranch, Path: repositoryPath, SyncStrategy: domainscm.SyncStrategy(payload.SyncStrategy)}, Metadata: metadata})
	case "sync_repository":
		_, executionErr = service.SyncRepository(ctx, applicationscm.SyncRepositoryRequest{Repository: repository, Path: repositoryPath, Metadata: metadata})
	case "cleanup_repository":
		executionErr = service.CleanupRepository(ctx, applicationscm.CleanupRepositoryRequest{Repository: repository, Path: repositoryPath, Metadata: metadata})
	case "ensure_branch":
		_, executionErr = service.EnsureBranch(ctx, applicationscm.EnsureBranchRequest{Repository: repository, Spec: domainscm.BranchSpec{BaseBranch: payload.BaseBranch, TargetBranch: payload.TargetBranch}, Metadata: metadata})
	case "sync_branch":
		_, executionErr = service.SyncBranch(ctx, applicationscm.SyncBranchRequest{Repository: repository, BranchName: payload.TargetBranch, Metadata: metadata})
	case "upsert_pull_request":
		_, executionErr = service.CreateOrUpdatePullRequest(ctx, applicationscm.CreateOrUpdatePullRequestRequest{Spec: domainscm.PullRequestSpec{Repository: repository, SourceBranch: payload.TargetBranch, TargetBranch: payload.BaseBranch, Title: payload.PullRequestTitle, Body: payload.PullRequestBody}, Metadata: metadata})
	case "get_pull_request":
		_, executionErr = service.GetPullRequest(ctx, applicationscm.GetPullRequestRequest{Repository: repository, PullRequestNumber: payload.PullRequestID, Metadata: metadata})
	case "submit_review":
		_, executionErr = service.SubmitReview(ctx, applicationscm.SubmitReviewRequest{Spec: domainscm.ReviewSpec{Repository: repository, PullRequestNumber: payload.PullRequestID, Decision: domainscm.ReviewDecision(payload.ReviewDecision), Body: payload.ReviewBody}, Metadata: metadata})
	case "check_merge_readiness":
		readiness, readinessErr := service.CheckMergeReadiness(ctx, applicationscm.CheckMergeReadinessRequest{Repository: repository, PullRequestNumber: payload.PullRequestID, Metadata: metadata})
		executionErr = readinessErr
		if readinessErr == nil {
			_ = readiness
		}
	case "merge_pull_request":
		_, executionErr = service.MergePullRequest(ctx, applicationscm.MergePullRequestRequest{Spec: domainscm.MergePullRequestSpec{Repository: repository, PullRequestNumber: payload.PullRequestID, Method: domainscm.MergeMethod(payload.MergeMethod)}, Metadata: metadata})
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
	}
	handler.safeRecordExecution(ctx, metadata.CorrelationIDs, job.Kind, idempotencyKey, operation, taskengine.ExecutionStatusSucceeded, "")
	return nil
}

func scopedRepositoryPath(projectID string, repositoryPath string) string {
	trimmedPath := strings.TrimSpace(repositoryPath)
	if trimmedPath == "" || filepath.IsAbs(trimmedPath) {
		return trimmedPath
	}
	cleanPath := filepath.Clean(trimmedPath)
	projectRoot := filepath.Join(strings.TrimSpace(projectID), "repositories")
	if strings.TrimSpace(projectID) == "" {
		return cleanPath
	}
	if cleanPath == projectRoot || strings.HasPrefix(cleanPath, projectRoot+string(filepath.Separator)) {
		return cleanPath
	}
	return filepath.Join(projectRoot, cleanPath)
}

func (handler *SCMWorkflowHandler) safeRecordExecution(ctx context.Context, correlationIDs taskengine.CorrelationIDs, kind taskengine.JobKind, idempotencyKey string, step string, status taskengine.ExecutionStatus, errorMessage string) {
	_, _ = handler.recordExecution(ctx, correlationIDs, kind, idempotencyKey, step, status, errorMessage)
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
