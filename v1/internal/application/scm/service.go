package scm

import (
	"agentic-orchestrator/internal/application/taskengine"
	"agentic-orchestrator/internal/domain/failures"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"errors"
	"strings"
)

type Metadata struct {
	CorrelationIDs taskengine.CorrelationIDs
	IdempotencyKey string
}

func (metadata Metadata) Validate() error {
	if strings.TrimSpace(metadata.CorrelationIDs.RunID) == "" {
		return failures.WrapTerminal(errors.New("correlation run_id is required"))
	}
	if strings.TrimSpace(metadata.CorrelationIDs.TaskID) == "" {
		return failures.WrapTerminal(errors.New("correlation task_id is required"))
	}
	if strings.TrimSpace(metadata.CorrelationIDs.JobID) == "" {
		return failures.WrapTerminal(errors.New("correlation job_id is required"))
	}
	if strings.TrimSpace(metadata.IdempotencyKey) == "" {
		return failures.WrapTerminal(errors.New("idempotency_key is required"))
	}
	return nil
}

type SourceStateRequest struct {
	Repository domainscm.Repository
	Metadata   Metadata
}

func (request SourceStateRequest) Validate() error {
	if err := request.Repository.Validate(); err != nil {
		return err
	}
	return request.Metadata.Validate()
}

type EnsureWorktreeRequest struct {
	Repository domainscm.Repository
	Spec       domainscm.WorktreeSpec
	Metadata   Metadata
}

func (request EnsureWorktreeRequest) Validate() error {
	if err := request.Repository.Validate(); err != nil {
		return err
	}
	if err := request.Spec.Validate(); err != nil {
		return err
	}
	return request.Metadata.Validate()
}

type SyncWorktreeRequest struct {
	Repository domainscm.Repository
	Path       string
	Metadata   Metadata
}

func (request SyncWorktreeRequest) Validate() error {
	if err := request.Repository.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(request.Path) == "" {
		return failures.WrapTerminal(errors.New("path is required"))
	}
	return request.Metadata.Validate()
}

type CleanupWorktreeRequest struct {
	Repository domainscm.Repository
	Path       string
	Metadata   Metadata
}

func (request CleanupWorktreeRequest) Validate() error {
	if err := request.Repository.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(request.Path) == "" {
		return failures.WrapTerminal(errors.New("path is required"))
	}
	return request.Metadata.Validate()
}

type EnsureBranchRequest struct {
	Repository domainscm.Repository
	Spec       domainscm.BranchSpec
	Metadata   Metadata
}

func (request EnsureBranchRequest) Validate() error {
	if err := request.Repository.Validate(); err != nil {
		return err
	}
	if err := request.Spec.Validate(); err != nil {
		return err
	}
	return request.Metadata.Validate()
}

type SyncBranchRequest struct {
	Repository domainscm.Repository
	BranchName string
	Metadata   Metadata
}

func (request SyncBranchRequest) Validate() error {
	if err := request.Repository.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(request.BranchName) == "" {
		return failures.WrapTerminal(errors.New("branch_name is required"))
	}
	return request.Metadata.Validate()
}

type CreateOrUpdatePullRequestRequest struct {
	Spec     domainscm.PullRequestSpec
	Metadata Metadata
}

func (request CreateOrUpdatePullRequestRequest) Validate() error {
	if err := request.Spec.Validate(); err != nil {
		return err
	}
	return request.Metadata.Validate()
}

type GetPullRequestRequest struct {
	Repository        domainscm.Repository
	PullRequestNumber int
	Metadata          Metadata
}

func (request GetPullRequestRequest) Validate() error {
	if err := request.Repository.Validate(); err != nil {
		return err
	}
	if request.PullRequestNumber <= 0 {
		return failures.WrapTerminal(errors.New("pull_request_number is required"))
	}
	return request.Metadata.Validate()
}

type SubmitReviewRequest struct {
	Spec     domainscm.ReviewSpec
	Metadata Metadata
}

func (request SubmitReviewRequest) Validate() error {
	if err := request.Spec.Validate(); err != nil {
		return err
	}
	return request.Metadata.Validate()
}

type CheckMergeReadinessRequest struct {
	Repository        domainscm.Repository
	PullRequestNumber int
	Metadata          Metadata
}

func (request CheckMergeReadinessRequest) Validate() error {
	if err := request.Repository.Validate(); err != nil {
		return err
	}
	if request.PullRequestNumber <= 0 {
		return failures.WrapTerminal(errors.New("pull_request_number is required"))
	}
	return request.Metadata.Validate()
}

type Service struct {
	orchestrator domainscm.Orchestrator
}

func NewService(orchestrator domainscm.Orchestrator) (*Service, error) {
	if orchestrator == nil {
		return nil, failures.WrapTerminal(errors.New("scm orchestrator is required"))
	}
	return &Service{orchestrator: orchestrator}, nil
}

func (service *Service) SourceState(ctx context.Context, request SourceStateRequest) (domainscm.SourceState, error) {
	if err := request.Validate(); err != nil {
		return domainscm.SourceState{}, err
	}
	state, err := service.orchestrator.SourceState(ctx, request.Repository)
	if err != nil {
		return domainscm.SourceState{}, ensureClassified(err)
	}
	if err := state.Validate(); err != nil {
		return domainscm.SourceState{}, err
	}
	return state, nil
}

func (service *Service) EnsureWorktree(ctx context.Context, request EnsureWorktreeRequest) (domainscm.WorktreeState, error) {
	if err := request.Validate(); err != nil {
		return domainscm.WorktreeState{}, err
	}
	state, err := service.orchestrator.EnsureWorktree(ctx, request.Repository, request.Spec)
	if err != nil {
		return domainscm.WorktreeState{}, ensureClassified(err)
	}
	if err := state.Validate(); err != nil {
		return domainscm.WorktreeState{}, err
	}
	return state, nil
}

func (service *Service) SyncWorktree(ctx context.Context, request SyncWorktreeRequest) (domainscm.WorktreeState, error) {
	if err := request.Validate(); err != nil {
		return domainscm.WorktreeState{}, err
	}
	state, err := service.orchestrator.SyncWorktree(ctx, request.Repository, request.Path)
	if err != nil {
		return domainscm.WorktreeState{}, ensureClassified(err)
	}
	if err := state.Validate(); err != nil {
		return domainscm.WorktreeState{}, err
	}
	return state, nil
}

func (service *Service) CleanupWorktree(ctx context.Context, request CleanupWorktreeRequest) error {
	if err := request.Validate(); err != nil {
		return err
	}
	if err := service.orchestrator.CleanupWorktree(ctx, request.Repository, request.Path); err != nil {
		return ensureClassified(err)
	}
	return nil
}

func (service *Service) EnsureBranch(ctx context.Context, request EnsureBranchRequest) (domainscm.BranchState, error) {
	if err := request.Validate(); err != nil {
		return domainscm.BranchState{}, err
	}
	state, err := service.orchestrator.EnsureBranch(ctx, request.Repository, request.Spec)
	if err != nil {
		return domainscm.BranchState{}, ensureClassified(err)
	}
	if err := state.Validate(); err != nil {
		return domainscm.BranchState{}, err
	}
	return state, nil
}

func (service *Service) SyncBranch(ctx context.Context, request SyncBranchRequest) (domainscm.BranchState, error) {
	if err := request.Validate(); err != nil {
		return domainscm.BranchState{}, err
	}
	state, err := service.orchestrator.SyncBranch(ctx, request.Repository, request.BranchName)
	if err != nil {
		return domainscm.BranchState{}, ensureClassified(err)
	}
	if err := state.Validate(); err != nil {
		return domainscm.BranchState{}, err
	}
	return state, nil
}

func (service *Service) CreateOrUpdatePullRequest(ctx context.Context, request CreateOrUpdatePullRequestRequest) (domainscm.PullRequestState, error) {
	if err := request.Validate(); err != nil {
		return domainscm.PullRequestState{}, err
	}
	state, err := service.orchestrator.CreateOrUpdatePullRequest(ctx, request.Spec)
	if err != nil {
		return domainscm.PullRequestState{}, ensureClassified(err)
	}
	if err := state.Validate(); err != nil {
		return domainscm.PullRequestState{}, err
	}
	return state, nil
}

func (service *Service) GetPullRequest(ctx context.Context, request GetPullRequestRequest) (domainscm.PullRequestState, error) {
	if err := request.Validate(); err != nil {
		return domainscm.PullRequestState{}, err
	}
	state, err := service.orchestrator.GetPullRequest(ctx, request.Repository, request.PullRequestNumber)
	if err != nil {
		return domainscm.PullRequestState{}, ensureClassified(err)
	}
	if err := state.Validate(); err != nil {
		return domainscm.PullRequestState{}, err
	}
	return state, nil
}

func (service *Service) SubmitReview(ctx context.Context, request SubmitReviewRequest) (domainscm.ReviewDecision, error) {
	if err := request.Validate(); err != nil {
		return "", err
	}
	decision, err := service.orchestrator.SubmitReview(ctx, request.Spec)
	if err != nil {
		return "", ensureClassified(err)
	}
	return decision, nil
}

func (service *Service) CheckMergeReadiness(ctx context.Context, request CheckMergeReadinessRequest) (domainscm.MergeReadiness, error) {
	if err := request.Validate(); err != nil {
		return domainscm.MergeReadiness{}, err
	}
	result, err := service.orchestrator.CheckMergeReadiness(ctx, request.Repository, request.PullRequestNumber)
	if err != nil {
		return domainscm.MergeReadiness{}, ensureClassified(err)
	}
	if err := result.Validate(); err != nil {
		return domainscm.MergeReadiness{}, err
	}
	return result, nil
}

func ensureClassified(err error) error {
	if err == nil {
		return nil
	}
	if failures.ClassOf(err) != failures.ClassUnknown {
		return err
	}
	return failures.WrapTransient(err)
}
