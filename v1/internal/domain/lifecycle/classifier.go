package lifecycle

import "time"

type State string

const (
	StateHealthyActive          State = "healthy_active"
	StateHealthyWaitingInput    State = "healthy_waiting_input"
	StateStaleNeedsNudge        State = "stale_needs_nudge"
	StateStuckNeedsIntervention State = "stuck_needs_intervention"
	StateExitedUnexpected       State = "exited_unexpected"
	StateCompleted              State = "completed"
)

type ClassificationPolicy struct {
	WaitingInputThreshold time.Duration
	StuckThreshold        time.Duration
	DebounceWindow        time.Duration
}

func DefaultClassificationPolicy() ClassificationPolicy {
	return ClassificationPolicy{
		WaitingInputThreshold: 2 * time.Minute,
		StuckThreshold:        5 * time.Minute,
		DebounceWindow:        30 * time.Second,
	}
}

type ClassificationInput struct {
	Now               time.Time
	EventType         EventType
	RuntimeAlive      bool
	WaitingInput      bool
	LastActivityAt    *time.Time
	LastCheckpointAt  *time.Time
	PreviousState     State
	PreviousChangedAt *time.Time
}

type ClassificationResult struct {
	State      State
	Severity   string
	ReasonCode string
}

func Classify(input ClassificationInput, policy ClassificationPolicy) ClassificationResult {
	now := input.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	defaults := DefaultClassificationPolicy()
	if policy.WaitingInputThreshold <= 0 {
		policy.WaitingInputThreshold = defaults.WaitingInputThreshold
	}
	if policy.StuckThreshold <= 0 {
		policy.StuckThreshold = defaults.StuckThreshold
	}
	if policy.DebounceWindow < 0 {
		policy.DebounceWindow = 0
	}

	candidate := classifyWithoutDebounce(input, now, policy)
	if candidate.State == StateCompleted || candidate.State == StateExitedUnexpected {
		return candidate
	}
	if input.EventType == EventStarted && input.RuntimeAlive && input.PreviousState == StateExitedUnexpected {
		return candidate
	}
	if input.PreviousState != "" && input.PreviousState != candidate.State && input.PreviousChangedAt != nil {
		if now.Sub(*input.PreviousChangedAt) < policy.DebounceWindow {
			return ClassificationResult{State: input.PreviousState, Severity: severityForState(input.PreviousState), ReasonCode: "debounce_suppressed"}
		}
	}
	return candidate
}

func classifyWithoutDebounce(input ClassificationInput, now time.Time, policy ClassificationPolicy) ClassificationResult {
	if input.EventType == EventCompleted {
		return ClassificationResult{State: StateCompleted, Severity: severityForState(StateCompleted), ReasonCode: "completed"}
	}
	if !input.RuntimeAlive || input.EventType == EventFailed {
		return ClassificationResult{State: StateExitedUnexpected, Severity: severityForState(StateExitedUnexpected), ReasonCode: "runtime_not_alive"}
	}
	if input.WaitingInput {
		if isStale(now, input.LastActivityAt, policy.WaitingInputThreshold) {
			return ClassificationResult{State: StateStaleNeedsNudge, Severity: severityForState(StateStaleNeedsNudge), ReasonCode: "waiting_input_stale"}
		}
		return ClassificationResult{State: StateHealthyWaitingInput, Severity: severityForState(StateHealthyWaitingInput), ReasonCode: "waiting_input"}
	}
	if isStale(now, input.LastActivityAt, policy.StuckThreshold) {
		return ClassificationResult{State: StateStuckNeedsIntervention, Severity: severityForState(StateStuckNeedsIntervention), ReasonCode: "activity_stale"}
	}
	if isStale(now, input.LastCheckpointAt, policy.StuckThreshold) {
		return ClassificationResult{State: StateStuckNeedsIntervention, Severity: severityForState(StateStuckNeedsIntervention), ReasonCode: "checkpoint_stale"}
	}
	return ClassificationResult{State: StateHealthyActive, Severity: severityForState(StateHealthyActive), ReasonCode: "healthy_active"}
}

func severityForState(state State) string {
	switch state {
	case StateExitedUnexpected, StateStuckNeedsIntervention:
		return "error"
	case StateStaleNeedsNudge:
		return "warning"
	default:
		return "info"
	}
}

func isStale(now time.Time, value *time.Time, threshold time.Duration) bool {
	if value == nil || threshold <= 0 {
		return false
	}
	return now.Sub(value.UTC()) >= threshold
}
