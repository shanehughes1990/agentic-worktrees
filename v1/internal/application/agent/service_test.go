package agent

import (
	domainagent "agentic-orchestrator/internal/domain/agent"
	"agentic-orchestrator/internal/domain/failures"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"errors"
	"testing"
)

type fakeSCMPort struct {
	sourceState      domainscm.SourceState
	sourceStateErr   error
	sourceStateCalls int
	lastRepository   domainscm.Repository
}

func (fake *fakeSCMPort) SourceState(_ context.Context, repository domainscm.Repository) (domainscm.SourceState, error) {
	fake.sourceStateCalls++
	fake.lastRepository = repository
	return fake.sourceState, fake.sourceStateErr
}

func (fake *fakeSCMPort) EnsureWorktree(_ context.Context, _ domainscm.Repository, _ domainscm.WorktreeSpec) (domainscm.WorktreeState, error) {
	return domainscm.WorktreeState{}, nil
}

func (fake *fakeSCMPort) SyncWorktree(_ context.Context, _ domainscm.Repository, _ string) (domainscm.WorktreeState, error) {
	return domainscm.WorktreeState{}, nil
}

func (fake *fakeSCMPort) CleanupWorktree(_ context.Context, _ domainscm.Repository, _ string) error {
	return nil
}

func (fake *fakeSCMPort) EnsureBranch(_ context.Context, _ domainscm.Repository, _ domainscm.BranchSpec) (domainscm.BranchState, error) {
	return domainscm.BranchState{}, nil
}

func (fake *fakeSCMPort) SyncBranch(_ context.Context, _ domainscm.Repository, _ string) (domainscm.BranchState, error) {
	return domainscm.BranchState{}, nil
}

func (fake *fakeSCMPort) CreateOrUpdatePullRequest(_ context.Context, _ domainscm.PullRequestSpec) (domainscm.PullRequestState, error) {
	return domainscm.PullRequestState{}, nil
}

func (fake *fakeSCMPort) GetPullRequest(_ context.Context, _ domainscm.Repository, _ int) (domainscm.PullRequestState, error) {
	return domainscm.PullRequestState{}, nil
}

func (fake *fakeSCMPort) SubmitReview(_ context.Context, _ domainscm.ReviewSpec) (domainscm.ReviewDecision, error) {
	return "", nil
}

func (fake *fakeSCMPort) CheckMergeReadiness(_ context.Context, _ domainscm.Repository, _ int) (domainscm.MergeReadiness, error) {
	return domainscm.MergeReadiness{}, nil
}

func validExecutionRequest() domainagent.ExecutionRequest {
	return domainagent.ExecutionRequest{
		Session: domainagent.SessionRef{
			SessionID: "session-1",
			Repository: domainscm.Repository{
				Provider: "github",
				Owner:    "acme",
				Name:     "repo",
			},
		},
		Prompt: "do the thing",
		Metadata: domainagent.Metadata{
			CorrelationIDs: domainagent.CorrelationIDs{
				RunID:  "run-1",
				TaskID: "task-1",
				JobID:  "job-1",
			},
			IdempotencyKey: "id-1",
		},
	}
}

func TestNewServiceRequiresSCMPort(t *testing.T) {
	_, err := NewService(nil)
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestExecuteQueriesSourceState(t *testing.T) {
	scm := &fakeSCMPort{sourceState: domainscm.SourceState{DefaultBranch: "main", HeadSHA: "abc123"}}
	service, err := NewService(scm)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	request := validExecutionRequest()
	if executeErr := service.Execute(context.Background(), request); executeErr != nil {
		t.Fatalf("execute: %v", executeErr)
	}
	if scm.sourceStateCalls != 1 {
		t.Fatalf("expected one source state call, got %d", scm.sourceStateCalls)
	}
	if scm.lastRepository != request.Session.Repository {
		t.Fatalf("expected source state to use request repository")
	}
}

func TestExecuteClassifiesUnknownErrorsAsTransient(t *testing.T) {
	scm := &fakeSCMPort{sourceStateErr: errors.New("temporary github outage")}
	service, err := NewService(scm)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	executeErr := service.Execute(context.Background(), validExecutionRequest())
	if !failures.IsClass(executeErr, failures.ClassTransient) {
		t.Fatalf("expected transient error, got %q (%v)", failures.ClassOf(executeErr), executeErr)
	}
}

func TestIntrospectSessionBuildsValidatedState(t *testing.T) {
	scm := &fakeSCMPort{sourceState: domainscm.SourceState{DefaultBranch: "main", HeadSHA: "abc123"}}
	service, err := NewService(scm)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	request := validExecutionRequest()
	state, introspectErr := service.IntrospectSession(context.Background(), domainagent.SessionIntrospectionRequest{
		Session:  request.Session,
		Metadata: request.Metadata,
	})
	if introspectErr != nil {
		t.Fatalf("introspect session: %v", introspectErr)
	}
	if state.SessionID != request.Session.SessionID {
		t.Fatalf("expected session id %q, got %q", request.Session.SessionID, state.SessionID)
	}
	if state.SourceState.HeadSHA != "abc123" {
		t.Fatalf("expected source state head sha abc123, got %q", state.SourceState.HeadSHA)
	}
}
