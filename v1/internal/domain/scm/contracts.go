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

func (repository Repository) Validate() error {
	if strings.TrimSpace(repository.Provider) == "" {
		return failures.WrapTerminal(errors.New("provider is required"))
	}
	if strings.TrimSpace(repository.Owner) == "" {
		return failures.WrapTerminal(errors.New("owner is required"))
	}
	if strings.TrimSpace(repository.Name) == "" {
		return failures.WrapTerminal(errors.New("name is required"))
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

type WorktreeSpec struct {
	BaseBranch   string
	TargetBranch string
	Path         string
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
}
