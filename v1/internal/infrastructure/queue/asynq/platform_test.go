package asynq

import (
	"agentic-orchestrator/internal/domain/failures"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
)

func TestMapHandlerErrorToQueuePolicyMarksTerminalAsSkipRetry(t *testing.T) {
	handleErr := failures.WrapTerminal(errors.New("terminal failure"))
	mapped := mapHandlerErrorToQueuePolicy(handleErr)
	if !errors.Is(mapped, asynq.SkipRetry) {
		t.Fatalf("expected mapped terminal error to include asynq.SkipRetry")
	}
	if !errors.Is(mapped, handleErr) {
		t.Fatalf("expected mapped error to retain original terminal error")
	}
}

func TestMapHandlerErrorToQueuePolicyLeavesTransientRetryable(t *testing.T) {
	handleErr := failures.WrapTransient(errors.New("transient failure"))
	mapped := mapHandlerErrorToQueuePolicy(handleErr)
	if errors.Is(mapped, asynq.SkipRetry) {
		t.Fatalf("did not expect transient errors to include asynq.SkipRetry")
	}
	if !errors.Is(mapped, handleErr) {
		t.Fatalf("expected mapped transient error to retain original error")
	}
}
