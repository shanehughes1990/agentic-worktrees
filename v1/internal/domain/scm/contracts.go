package scm

import (
	"agentic-orchestrator/internal/domain/failures"
	"context"
	"errors"
	"fmt"
	"strings"
)

type Repository struct {
	Provider string
	Owner    string
	Name     string
}

const ProviderGitHub = "github"

var SupportedProviders = []string{ProviderGitHub}

func (repository Repository) Validate() error {
	provider := strings.ToLower(strings.TrimSpace(repository.Provider))
	if provider == "" {
		return failures.WrapTerminal(errors.New("provider is required"))
	}
	if !isSupportedProvider(provider) {
		return failures.WrapTerminal(fmt.Errorf("unsupported provider %q", repository.Provider))
	}
	if strings.TrimSpace(repository.Owner) == "" {
		return failures.WrapTerminal(errors.New("owner is required"))
	}
	if strings.TrimSpace(repository.Name) == "" {
		return failures.WrapTerminal(errors.New("name is required"))
	}
	return nil
}

func isSupportedProvider(provider string) bool {
	for _, supportedProvider := range SupportedProviders {
		if provider == supportedProvider {
			return true
		}
	}
	return false
}

type RepoCacheKey string

func RepoCacheKeyFromRepository(repository Repository) RepoCacheKey {
	return RepoCacheKey(strings.ToLower(strings.TrimSpace(repository.Provider) + "/" + strings.TrimSpace(repository.Owner) + "/" + strings.TrimSpace(repository.Name)))
}

func (key RepoCacheKey) Validate() error {
	if strings.TrimSpace(string(key)) == "" {
		return failures.WrapTerminal(errors.New("repo_cache_key is required"))
	}
	return nil
}

type RepoLease struct {
	CacheKey RepoCacheKey
	OwnerID  string
	Token    string
}

func (lease RepoLease) Validate() error {
	if err := lease.CacheKey.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(lease.OwnerID) == "" {
		return failures.WrapTerminal(errors.New("owner_id is required"))
	}
	if strings.TrimSpace(lease.Token) == "" {
		return failures.WrapTerminal(errors.New("token is required"))
	}
	return nil
}

type SourceState struct {
	DefaultBranch string
	HeadSHA       string
}

func (state SourceState) Validate() error {
	if strings.TrimSpace(state.DefaultBranch) == "" {
		return failures.WrapTerminal(errors.New("default_branch is required"))
	}
	if strings.TrimSpace(state.HeadSHA) == "" {
		return failures.WrapTerminal(errors.New("head_sha is required"))
	}
	return nil
}

type SyncStrategy string

const (
	SyncStrategyMerge  SyncStrategy = "merge"
	SyncStrategyRebase SyncStrategy = "rebase"
)

func (strategy SyncStrategy) Canonical() SyncStrategy {
	normalized := strings.ToLower(strings.TrimSpace(string(strategy)))
	if normalized == "" {
		return SyncStrategyMerge
	}
	return SyncStrategy(normalized)
}

type WorktreeSpec struct {
	BaseBranch   string
	TargetBranch string
	Path         string
	SyncStrategy SyncStrategy
}

func (spec WorktreeSpec) EffectiveSyncStrategy() SyncStrategy {
	return spec.SyncStrategy.Canonical()
}

func (spec WorktreeSpec) Validate() error {
	if strings.TrimSpace(spec.BaseBranch) == "" {
		return failures.WrapTerminal(errors.New("base_branch is required"))
	}
	if strings.TrimSpace(spec.TargetBranch) == "" {
		return failures.WrapTerminal(errors.New("target_branch is required"))
	}
	if strings.TrimSpace(spec.Path) == "" {
		return failures.WrapTerminal(errors.New("path is required"))
	}
	switch spec.EffectiveSyncStrategy() {
	case SyncStrategyMerge, SyncStrategyRebase:
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported sync strategy %q", spec.SyncStrategy))
	}
	return nil
}

type WorktreeState struct {
	Path      string
	Branch    string
	Base      string
	HeadSHA   string
	IsInSync  bool
	IsCleaned bool
}

func (state WorktreeState) Validate() error {
	if strings.TrimSpace(state.Path) == "" {
		return failures.WrapTerminal(errors.New("path is required"))
	}
	if strings.TrimSpace(state.Branch) == "" {
		return failures.WrapTerminal(errors.New("branch is required"))
	}
	if strings.TrimSpace(state.Base) == "" {
		return failures.WrapTerminal(errors.New("base is required"))
	}
	if strings.TrimSpace(state.HeadSHA) == "" {
		return failures.WrapTerminal(errors.New("head_sha is required"))
	}
	return nil
}

type BranchSpec struct {
	BaseBranch   string
	TargetBranch string
}

func (spec BranchSpec) Validate() error {
	if strings.TrimSpace(spec.BaseBranch) == "" {
		return failures.WrapTerminal(errors.New("base_branch is required"))
	}
	if strings.TrimSpace(spec.TargetBranch) == "" {
		return failures.WrapTerminal(errors.New("target_branch is required"))
	}
	return nil
}

type BranchState struct {
	Name    string
	Base    string
	HeadSHA string
	InSync  bool
}

func (state BranchState) Validate() error {
	if strings.TrimSpace(state.Name) == "" {
		return failures.WrapTerminal(errors.New("name is required"))
	}
	if strings.TrimSpace(state.Base) == "" {
		return failures.WrapTerminal(errors.New("base is required"))
	}
	if strings.TrimSpace(state.HeadSHA) == "" {
		return failures.WrapTerminal(errors.New("head_sha is required"))
	}
	return nil
}

type PullRequestSpec struct {
	Repository   Repository
	SourceBranch string
	TargetBranch string
	Title        string
	Body         string
}

func (spec PullRequestSpec) Validate() error {
	if err := spec.Repository.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(spec.SourceBranch) == "" {
		return failures.WrapTerminal(errors.New("source_branch is required"))
	}
	if strings.TrimSpace(spec.TargetBranch) == "" {
		return failures.WrapTerminal(errors.New("target_branch is required"))
	}
	if strings.TrimSpace(spec.Title) == "" {
		return failures.WrapTerminal(errors.New("title is required"))
	}
	return nil
}

type PullRequestState struct {
	Number  int
	URL     string
	State   string
	Merged  bool
	HeadSHA string
}

func (state PullRequestState) Validate() error {
	if state.Number <= 0 {
		return failures.WrapTerminal(fmt.Errorf("pull request number must be positive: %d", state.Number))
	}
	if strings.TrimSpace(state.URL) == "" {
		return failures.WrapTerminal(errors.New("pull request url is required"))
	}
	if strings.TrimSpace(state.State) == "" {
		return failures.WrapTerminal(errors.New("pull request state is required"))
	}
	if strings.TrimSpace(state.HeadSHA) == "" {
		return failures.WrapTerminal(errors.New("pull request head_sha is required"))
	}
	return nil
}

type ReviewDecision string

const (
	ReviewDecisionApprove        ReviewDecision = "approve"
	ReviewDecisionRequestChanges ReviewDecision = "request_changes"
	ReviewDecisionComment        ReviewDecision = "comment"
)

type ReviewSpec struct {
	Repository        Repository
	PullRequestNumber int
	Decision          ReviewDecision
	Body              string
}

func (spec ReviewSpec) Validate() error {
	if err := spec.Repository.Validate(); err != nil {
		return err
	}
	if spec.PullRequestNumber <= 0 {
		return failures.WrapTerminal(errors.New("pull_request_number is required"))
	}
	switch spec.Decision {
	case ReviewDecisionApprove, ReviewDecisionRequestChanges, ReviewDecisionComment:
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported review decision %q", spec.Decision))
	}
	if strings.TrimSpace(spec.Body) == "" {
		return failures.WrapTerminal(errors.New("body is required"))
	}
	return nil
}

type MergeReadiness struct {
	CanMerge bool
	Reason   string
}

func (readiness MergeReadiness) Validate() error {
	if readiness.CanMerge {
		return nil
	}
	if strings.TrimSpace(readiness.Reason) == "" {
		return failures.WrapTerminal(errors.New("reason is required when can_merge is false"))
	}
	return nil
}

type MergeMethod string

const (
	MergeMethodMerge  MergeMethod = "merge"
	MergeMethodSquash MergeMethod = "squash"
	MergeMethodRebase MergeMethod = "rebase"
)

func (method MergeMethod) Canonical() MergeMethod {
	normalized := strings.ToLower(strings.TrimSpace(string(method)))
	if normalized == "" {
		return MergeMethodSquash
	}
	return MergeMethod(normalized)
}

type MergePullRequestSpec struct {
	Repository        Repository
	PullRequestNumber int
	Method            MergeMethod
	CommitTitle       string
	CommitMessage     string
}

func (spec MergePullRequestSpec) Validate() error {
	if err := spec.Repository.Validate(); err != nil {
		return err
	}
	if spec.PullRequestNumber <= 0 {
		return failures.WrapTerminal(errors.New("pull_request_number is required"))
	}
	switch spec.Method.Canonical() {
	case MergeMethodMerge, MergeMethodSquash, MergeMethodRebase:
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported merge method %q", spec.Method))
	}
	return nil
}

type Orchestrator interface {
	SourceState(ctx context.Context, repository Repository) (SourceState, error)
	EnsureWorktree(ctx context.Context, repository Repository, spec WorktreeSpec) (WorktreeState, error)
	SyncWorktree(ctx context.Context, repository Repository, path string) (WorktreeState, error)
	CleanupWorktree(ctx context.Context, repository Repository, path string) error
	EnsureBranch(ctx context.Context, repository Repository, spec BranchSpec) (BranchState, error)
	SyncBranch(ctx context.Context, repository Repository, branchName string) (BranchState, error)
	CreateOrUpdatePullRequest(ctx context.Context, spec PullRequestSpec) (PullRequestState, error)
	GetPullRequest(ctx context.Context, repository Repository, pullRequestNumber int) (PullRequestState, error)
	SubmitReview(ctx context.Context, spec ReviewSpec) (ReviewDecision, error)
	CheckMergeReadiness(ctx context.Context, repository Repository, pullRequestNumber int) (MergeReadiness, error)
	MergePullRequest(ctx context.Context, spec MergePullRequestSpec) (PullRequestState, error)
}
