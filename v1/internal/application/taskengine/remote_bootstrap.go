package taskengine

import (
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"fmt"
	"strings"
)

type SCMBootstrapPort interface {
	SourceState(ctx context.Context, repository domainscm.Repository) (domainscm.SourceState, error)
	EnsureRepository(ctx context.Context, repository domainscm.Repository, spec domainscm.RepositorySpec) (domainscm.RepositoryState, error)
}

type RemoteBootstrapRequest struct {
	Job            Job
	CorrelationIDs CorrelationIDs
	IdempotencyKey string
	Repository     domainscm.Repository
	BaseBranch     string
	TargetBranch   string
	RepositoryPath   string
}

func (request RemoteBootstrapRequest) Validate() error {
	if strings.TrimSpace(string(request.Job.Kind)) == "" {
		return fmt.Errorf("%w: job kind is required", ErrInvalidRemoteExecutionRequest)
	}
	if strings.TrimSpace(request.Job.QueueTaskID) == "" {
		return fmt.Errorf("%w: queue_task_id is required", ErrInvalidRemoteExecutionRequest)
	}
	if len(request.Job.Payload) == 0 {
		return fmt.Errorf("%w: payload is required", ErrInvalidRemoteExecutionRequest)
	}
	if strings.TrimSpace(request.CorrelationIDs.RunID) == "" {
		return fmt.Errorf("%w: run_id is required", ErrInvalidRemoteExecutionRequest)
	}
	if strings.TrimSpace(request.CorrelationIDs.TaskID) == "" {
		return fmt.Errorf("%w: task_id is required", ErrInvalidRemoteExecutionRequest)
	}
	if strings.TrimSpace(request.CorrelationIDs.JobID) == "" {
		return fmt.Errorf("%w: job_id is required", ErrInvalidRemoteExecutionRequest)
	}
	if strings.TrimSpace(request.IdempotencyKey) == "" {
		return fmt.Errorf("%w: idempotency_key is required", ErrInvalidRemoteExecutionRequest)
	}
	if err := request.Repository.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(request.TargetBranch) == "" {
		return fmt.Errorf("%w: target_branch is required", ErrInvalidRemoteExecutionRequest)
	}
	if strings.TrimSpace(request.RepositoryPath) == "" {
		return fmt.Errorf("%w: repository_path is required", ErrInvalidRemoteExecutionRequest)
	}
	return nil
}

type RemoteBootstrapService struct {
	scm    SCMBootstrapPort
	remote RemoteWorkerAdapter
}

func NewRemoteBootstrapService(scm SCMBootstrapPort, remote RemoteWorkerAdapter) (*RemoteBootstrapService, error) {
	if scm == nil {
		return nil, fmt.Errorf("%w: scm bootstrap port is required", ErrInvalidRemoteExecutionRequest)
	}
	if remote == nil {
		return nil, fmt.Errorf("%w: remote worker adapter is required", ErrInvalidRemoteExecutionRequest)
	}
	return &RemoteBootstrapService{scm: scm, remote: remote}, nil
}

func (service *RemoteBootstrapService) BuildRemoteExecutionRequest(ctx context.Context, request RemoteBootstrapRequest) (RemoteExecutionRequest, error) {
	if service == nil || service.scm == nil {
		return RemoteExecutionRequest{}, fmt.Errorf("%w: remote bootstrap service is not initialized", ErrInvalidRemoteExecutionRequest)
	}
	if err := request.Validate(); err != nil {
		return RemoteExecutionRequest{}, err
	}
	baseBranch := strings.TrimSpace(request.BaseBranch)
	if baseBranch == "" {
		sourceState, err := service.scm.SourceState(ctx, request.Repository)
		if err != nil {
			return RemoteExecutionRequest{}, err
		}
		baseBranch = strings.TrimSpace(sourceState.DefaultBranch)
		if baseBranch == "" {
			return RemoteExecutionRequest{}, fmt.Errorf("%w: source_state default_branch is required", ErrInvalidRemoteExecutionRequest)
		}
	}
	_, err := service.scm.EnsureRepository(ctx, request.Repository, domainscm.RepositorySpec{
		BaseBranch:   baseBranch,
		TargetBranch: request.TargetBranch,
		Path:         request.RepositoryPath,
	})
	if err != nil {
		return RemoteExecutionRequest{}, err
	}
	remoteRequest := RemoteExecutionRequest{
		Job:            request.Job,
		CorrelationIDs: request.CorrelationIDs,
		IdempotencyKey: request.IdempotencyKey,
		ResumeCheckpoint: &RemoteCheckpoint{
			Step:  "ensure_repository",
			Token: request.IdempotencyKey,
		},
	}
	if err := remoteRequest.Validate(); err != nil {
		return RemoteExecutionRequest{}, err
	}
	return remoteRequest, nil
}

func (service *RemoteBootstrapService) Execute(ctx context.Context, request RemoteBootstrapRequest) (RemoteExecutionResult, error) {
	if service == nil || service.remote == nil {
		return RemoteExecutionResult{}, fmt.Errorf("%w: remote bootstrap service is not initialized", ErrInvalidRemoteExecutionRequest)
	}
	remoteRequest, err := service.BuildRemoteExecutionRequest(ctx, request)
	if err != nil {
		return RemoteExecutionResult{}, err
	}
	result, err := service.remote.Execute(ctx, remoteRequest)
	if err != nil {
		return RemoteExecutionResult{}, err
	}
	if result.CompletedCheckpoint == nil {
		result.CompletedCheckpoint = &RemoteCheckpoint{Step: "ensure_repository", Token: request.IdempotencyKey}
	}
	if err := result.Validate(); err != nil {
		return RemoteExecutionResult{}, err
	}
	return result, nil
}
