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

func (fake *fakeSCMPort) EnsureRepository(_ context.Context, _ domainscm.Repository, _ domainscm.RepositorySpec) (domainscm.RepositoryState, error) {
	return domainscm.RepositoryState{}, nil
}

func (fake *fakeSCMPort) SyncRepository(_ context.Context, _ domainscm.Repository, _ string) (domainscm.RepositoryState, error) {
	return domainscm.RepositoryState{}, nil
}

func (fake *fakeSCMPort) CleanupRepository(_ context.Context, _ domainscm.Repository, _ string) error {
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

// TestExecuteOnlyUsesPortInterface verifies that the agent service is constructed
// with a domainagent.SCMPort interface and never holds a concrete adapter reference,
// enforcing the no-provider-bypass invariant through the type system.
func TestExecuteOnlyUsesPortInterface(t *testing.T) {
	scm := &fakeSCMPort{sourceState: domainscm.SourceState{DefaultBranch: "main", HeadSHA: "sha1"}}
	service, err := NewService(scm)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.Execute(context.Background(), validExecutionRequest()); err != nil {
		t.Fatalf("execute via port double: %v", err)
	}
	if scm.sourceStateCalls != 1 {
		t.Fatalf("expected exactly one source state call through port, got %d", scm.sourceStateCalls)
	}
}

func TestExecuteRequiresNonEmptyIdempotencyKey(t *testing.T) {
	scm := &fakeSCMPort{}
	service, err := NewService(scm)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	req := validExecutionRequest()
	req.Metadata.IdempotencyKey = ""
	if executeErr := service.Execute(context.Background(), req); !failures.IsClass(executeErr, failures.ClassTerminal) {
		t.Fatalf("expected terminal error for empty idempotency key, got %v", executeErr)
	}
}

func TestExecuteRequiresCorrelationRunID(t *testing.T) {
	scm := &fakeSCMPort{}
	service, err := NewService(scm)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	req := validExecutionRequest()
	req.Metadata.CorrelationIDs.RunID = ""
	if executeErr := service.Execute(context.Background(), req); !failures.IsClass(executeErr, failures.ClassTerminal) {
		t.Fatalf("expected terminal error for empty run_id, got %v", executeErr)
	}
}

func TestExecuteWithCheckpointSkipsSourceState(t *testing.T) {
	scm := &fakeSCMPort{sourceState: domainscm.SourceState{DefaultBranch: "main", HeadSHA: "sha1"}}
	service, err := NewService(scm)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	req := validExecutionRequest()
	req.ResumeCheckpoint = &domainagent.Checkpoint{Step: "source_state", Token: "id-1"}
	if executeErr := service.Execute(context.Background(), req); executeErr != nil {
		t.Fatalf("execute with checkpoint: %v", executeErr)
	}
	if scm.sourceStateCalls != 0 {
		t.Fatalf("expected source_state to be skipped when checkpoint matches, got %d calls", scm.sourceStateCalls)
	}
}

func TestExecuteWithCheckpointTokenMismatchQueriesSourceState(t *testing.T) {
	scm := &fakeSCMPort{sourceState: domainscm.SourceState{DefaultBranch: "main", HeadSHA: "sha1"}}
	service, err := NewService(scm)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	req := validExecutionRequest()
	req.ResumeCheckpoint = &domainagent.Checkpoint{Step: "source_state", Token: "id-other"}
	if executeErr := service.Execute(context.Background(), req); executeErr != nil {
		t.Fatalf("execute with mismatched checkpoint token: %v", executeErr)
	}
	if scm.sourceStateCalls != 1 {
		t.Fatalf("expected source_state call when checkpoint token mismatches, got %d calls", scm.sourceStateCalls)
	}
}

func TestExecutePreservesTerminalClassification(t *testing.T) {
	terminalErr := failures.WrapTerminal(errors.New("auth token revoked"))
	scm := &fakeSCMPort{sourceStateErr: terminalErr}
	service, err := NewService(scm)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if executeErr := service.Execute(context.Background(), validExecutionRequest()); !failures.IsClass(executeErr, failures.ClassTerminal) {
		t.Fatalf("expected terminal classification preserved, got %q (%v)", failures.ClassOf(executeErr), executeErr)
	}
}
