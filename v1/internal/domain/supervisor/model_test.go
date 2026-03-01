package supervisor

import (
	"agentic-orchestrator/internal/domain/failures"
	"testing"
	"time"
)

func TestSignalValidate(t *testing.T) {
	signal := Signal{
		Type: SignalExecutionFailed,
		CorrelationIDs: CorrelationIDs{
			RunID:  "run-1",
			TaskID: "task-1",
			JobID:  "job-1",
		},
		Attempt:      1,
		MaxRetry:     3,
		FailureClass: failures.ClassTransient,
		OccurredAt:   time.Now().UTC(),
	}
	if err := signal.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
}

func TestIsTransitionAllowed(t *testing.T) {
	if !IsTransitionAllowed(StateExecuting, StateEscalated) {
		t.Fatalf("expected executing -> escalated to be allowed")
	}
	if IsTransitionAllowed(StateCompleted, StateExecuting) {
		t.Fatalf("expected completed -> executing to be disallowed")
	}
}
