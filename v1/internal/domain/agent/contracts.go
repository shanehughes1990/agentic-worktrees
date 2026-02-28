package agent

import (
	"agentic-orchestrator/internal/domain/failures"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"errors"
	"strings"
)

type CorrelationIDs struct {
	RunID  string
	TaskID string
	JobID  string
}

func (ids CorrelationIDs) Validate() error {
	if strings.TrimSpace(ids.RunID) == "" {
		return failures.WrapTerminal(errors.New("run_id is required"))
	}
	if strings.TrimSpace(ids.TaskID) == "" {
		return failures.WrapTerminal(errors.New("task_id is required"))
	}
	if strings.TrimSpace(ids.JobID) == "" {
		return failures.WrapTerminal(errors.New("job_id is required"))
	}
	return nil
}

type Metadata struct {
	CorrelationIDs CorrelationIDs
	IdempotencyKey string
}

func (metadata Metadata) Validate() error {
	if err := metadata.CorrelationIDs.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(metadata.IdempotencyKey) == "" {
		return failures.WrapTerminal(errors.New("idempotency_key is required"))
	}
	return nil
}

type SessionRef struct {
	SessionID  string
	Repository domainscm.Repository
}

func (session SessionRef) Validate() error {
	if strings.TrimSpace(session.SessionID) == "" {
		return failures.WrapTerminal(errors.New("session_id is required"))
	}
	return session.Repository.Validate()
}

type ExecutionRequest struct {
	Session  SessionRef
	Prompt   string
	Metadata Metadata
}

func (request ExecutionRequest) Validate() error {
	if err := request.Session.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(request.Prompt) == "" {
		return failures.WrapTerminal(errors.New("prompt is required"))
	}
	return request.Metadata.Validate()
}

type SessionIntrospectionRequest struct {
	Session  SessionRef
	Metadata Metadata
}

func (request SessionIntrospectionRequest) Validate() error {
	if err := request.Session.Validate(); err != nil {
		return err
	}
	return request.Metadata.Validate()
}

type Checkpoint struct {
	Step  string
	Token string
}

func (checkpoint Checkpoint) Validate() error {
	if strings.TrimSpace(checkpoint.Step) == "" {
		return failures.WrapTerminal(errors.New("checkpoint step is required"))
	}
	if strings.TrimSpace(checkpoint.Token) == "" {
		return failures.WrapTerminal(errors.New("checkpoint token is required"))
	}
	return nil
}

type SessionState struct {
	SessionID      string
	Repository     domainscm.Repository
	SourceState    domainscm.SourceState
	WorktreeState  *domainscm.WorktreeState
	BranchState    *domainscm.BranchState
	PullRequest    *domainscm.PullRequestState
	LastCheckpoint *Checkpoint
}

func (state SessionState) Validate() error {
	if strings.TrimSpace(state.SessionID) == "" {
		return failures.WrapTerminal(errors.New("session_id is required"))
	}
	if err := state.Repository.Validate(); err != nil {
		return err
	}
	if err := state.SourceState.Validate(); err != nil {
		return err
	}
	if state.WorktreeState != nil {
		if err := state.WorktreeState.Validate(); err != nil {
			return err
		}
	}
	if state.BranchState != nil {
		if err := state.BranchState.Validate(); err != nil {
			return err
		}
	}
	if state.PullRequest != nil {
		if err := state.PullRequest.Validate(); err != nil {
			return err
		}
	}
	if state.LastCheckpoint != nil {
		if err := state.LastCheckpoint.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type SCMPort interface {
	SourceState(ctx context.Context, repository domainscm.Repository) (domainscm.SourceState, error)
	EnsureWorktree(ctx context.Context, repository domainscm.Repository, spec domainscm.WorktreeSpec) (domainscm.WorktreeState, error)
	SyncWorktree(ctx context.Context, repository domainscm.Repository, path string) (domainscm.WorktreeState, error)
	CleanupWorktree(ctx context.Context, repository domainscm.Repository, path string) error
	EnsureBranch(ctx context.Context, repository domainscm.Repository, spec domainscm.BranchSpec) (domainscm.BranchState, error)
	SyncBranch(ctx context.Context, repository domainscm.Repository, branchName string) (domainscm.BranchState, error)
	CreateOrUpdatePullRequest(ctx context.Context, spec domainscm.PullRequestSpec) (domainscm.PullRequestState, error)
	GetPullRequest(ctx context.Context, repository domainscm.Repository, pullRequestNumber int) (domainscm.PullRequestState, error)
	SubmitReview(ctx context.Context, spec domainscm.ReviewSpec) (domainscm.ReviewDecision, error)
	CheckMergeReadiness(ctx context.Context, repository domainscm.Repository, pullRequestNumber int) (domainscm.MergeReadiness, error)
}

type ExecutionPort interface {
	Execute(ctx context.Context, request ExecutionRequest) error
}

type SessionIntrospectionPort interface {
	IntrospectSession(ctx context.Context, request SessionIntrospectionRequest) (SessionState, error)
}
