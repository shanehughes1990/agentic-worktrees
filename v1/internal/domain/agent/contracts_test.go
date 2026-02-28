package agent

import (
	"agentic-orchestrator/internal/domain/failures"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"testing"
)

func validMetadata() Metadata {
	return Metadata{
		CorrelationIDs: CorrelationIDs{
			RunID:  "run-1",
			TaskID: "task-1",
			JobID:  "job-1",
		},
		IdempotencyKey: "id-1",
	}
}

func validRepository() domainscm.Repository {
	return domainscm.Repository{
		Provider: "github",
		Owner:    "acme",
		Name:     "repo",
	}
}

func TestExecutionRequestValidateRequiresPrompt(t *testing.T) {
	err := (ExecutionRequest{
		Session: SessionRef{
			SessionID:  "session-1",
			Repository: validRepository(),
		},
		Metadata: validMetadata(),
	}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestSessionIntrospectionRequestValidateRequiresMetadata(t *testing.T) {
	err := (SessionIntrospectionRequest{
		SessionID: "session-1",
		Metadata:  Metadata{},
	}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestSessionStateValidateAcceptsOptionalSCMArtifacts(t *testing.T) {
	state := SessionState{
		SessionID:   "session-1",
		Repository:  validRepository(),
		SourceState: domainscm.SourceState{DefaultBranch: "main", HeadSHA: "abc123"},
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("validate session state: %v", err)
	}
}

func TestCheckpointValidateRequiresFields(t *testing.T) {
	err := (Checkpoint{}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}
