package lifecycle

import (
	"testing"
	"time"
)

func TestClassifyPrimaryModes(t *testing.T) {
	now := time.Now().UTC()
	stale := now.Add(-10 * time.Minute)
	fresh := now.Add(-30 * time.Second)
	policy := ClassificationPolicy{WaitingInputThreshold: 2 * time.Minute, StuckThreshold: 5 * time.Minute, DebounceWindow: 0}

	tests := []struct {
		name     string
		input    ClassificationInput
		expected State
	}{
		{name: "healthy active", input: ClassificationInput{Now: now, EventType: EventStarted, RuntimeAlive: true, LastActivityAt: &fresh, LastCheckpointAt: &fresh}, expected: StateHealthyActive},
		{name: "healthy waiting input", input: ClassificationInput{Now: now, EventType: EventStarted, RuntimeAlive: true, WaitingInput: true, LastActivityAt: &fresh, LastCheckpointAt: &fresh}, expected: StateHealthyWaitingInput},
		{name: "stale waiting input", input: ClassificationInput{Now: now, EventType: EventStarted, RuntimeAlive: true, WaitingInput: true, LastActivityAt: &stale, LastCheckpointAt: &fresh}, expected: StateStaleNeedsNudge},
		{name: "stuck due checkpoint", input: ClassificationInput{Now: now, EventType: EventStarted, RuntimeAlive: true, LastActivityAt: &fresh, LastCheckpointAt: &stale}, expected: StateStuckNeedsIntervention},
		{name: "exited unexpected", input: ClassificationInput{Now: now, EventType: EventFailed, RuntimeAlive: false, LastActivityAt: &fresh, LastCheckpointAt: &fresh}, expected: StateExitedUnexpected},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Classify(testCase.input, policy)
			if result.State != testCase.expected {
				t.Fatalf("expected state %s, got %s", testCase.expected, result.State)
			}
		})
	}
}

func TestClassifyAppliesDebounce(t *testing.T) {
	now := time.Now().UTC()
	previousChanged := now.Add(-5 * time.Second)
	stale := now.Add(-10 * time.Minute)
	policy := ClassificationPolicy{WaitingInputThreshold: 2 * time.Minute, StuckThreshold: 5 * time.Minute, DebounceWindow: 30 * time.Second}

	result := Classify(ClassificationInput{
		Now:               now,
		EventType:         EventStarted,
		RuntimeAlive:      true,
		WaitingInput:      true,
		LastActivityAt:    &stale,
		PreviousState:     StateHealthyActive,
		PreviousChangedAt: &previousChanged,
	}, policy)
	if result.State != StateHealthyActive {
		t.Fatalf("expected debounce to preserve previous state, got %s", result.State)
	}
	if result.ReasonCode != "debounce_suppressed" {
		t.Fatalf("expected debounce reason code, got %s", result.ReasonCode)
	}
}

func TestClassifyAllowsRecoveryAfterFailureWithinDebounceWindow(t *testing.T) {
	now := time.Now().UTC()
	previousChanged := now.Add(-5 * time.Second)
	fresh := now.Add(-10 * time.Second)
	policy := ClassificationPolicy{WaitingInputThreshold: 2 * time.Minute, StuckThreshold: 5 * time.Minute, DebounceWindow: 30 * time.Second}

	result := Classify(ClassificationInput{
		Now:               now,
		EventType:         EventStarted,
		RuntimeAlive:      true,
		LastActivityAt:    &fresh,
		LastCheckpointAt:  &fresh,
		PreviousState:     StateExitedUnexpected,
		PreviousChangedAt: &previousChanged,
	}, policy)

	if result.State != StateHealthyActive {
		t.Fatalf("expected started event to recover state, got %s", result.State)
	}
	if result.ReasonCode != "healthy_active" {
		t.Fatalf("expected healthy_active reason code, got %s", result.ReasonCode)
	}
}
