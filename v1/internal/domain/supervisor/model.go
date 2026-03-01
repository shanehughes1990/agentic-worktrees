package supervisor

import (
	"agentic-orchestrator/internal/domain/failures"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

var (
	ErrInvalidCorrelationIDs = errors.New("supervisor: invalid correlation identifiers")
	ErrInvalidSignal         = errors.New("supervisor: invalid signal")
	ErrInvalidDecision       = errors.New("supervisor: invalid decision")
	ErrNoMatchingRule        = errors.New("supervisor: no matching rule")
)

type State string

const (
	StateIdle       State = "idle"
	StateExecuting  State = "executing"
	StateReviewing  State = "reviewing"
	StateRework     State = "rework"
	StateMergeReady State = "merge_ready"
	StateBlocked    State = "blocked"
	StateEscalated  State = "escalated"
	StateMerged     State = "merged"
	StateRefused    State = "refused"
	StateCompleted  State = "completed"
)

type AttentionZone string

const (
	AttentionZoneNone      AttentionZone = "none"
	AttentionZoneTracker   AttentionZone = "tracker"
	AttentionZoneSCM       AttentionZone = "scm"
	AttentionZoneExecution AttentionZone = "execution"
)

type ActionCode string

const (
	ActionContinue      ActionCode = "continue"
	ActionRetry         ActionCode = "retry"
	ActionBlock         ActionCode = "block"
	ActionEscalate      ActionCode = "escalate"
	ActionRequestRework ActionCode = "request_rework"
	ActionMerge         ActionCode = "merge"
	ActionRefuse        ActionCode = "refuse"
	ActionStartTask     ActionCode = "start_task"
)

type ReasonCode string

const (
	ReasonJobAdmitted                ReasonCode = "job_admitted"
	ReasonExecutionProgressed        ReasonCode = "execution_progressed"
	ReasonExecutionSucceeded         ReasonCode = "execution_succeeded"
	ReasonExecutionFailedRetry       ReasonCode = "execution_failed_retry"
	ReasonExecutionFailedMaxed       ReasonCode = "execution_failed_max_retries"
	ReasonExecutionFailedFatal       ReasonCode = "execution_failed_terminal"
	ReasonTrackerAttention           ReasonCode = "tracker_attention_required"
	ReasonTrackerAttentionClear      ReasonCode = "tracker_attention_cleared"
	ReasonSCMAttention               ReasonCode = "scm_attention_required"
	ReasonSCMAttentionClear          ReasonCode = "scm_attention_cleared"
	ReasonPRConflictDetected         ReasonCode = "pr_conflict_detected"
	ReasonPRReviewChangesRequested   ReasonCode = "pr_review_changes_requested"
	ReasonPRChecksFailed             ReasonCode = "pr_checks_failed"
	ReasonPRChecksPassed             ReasonCode = "pr_checks_passed"
	ReasonPRMergeApproved            ReasonCode = "pr_merge_approved"
	ReasonPRMergeRefused             ReasonCode = "pr_merge_refused"
	ReasonIssueTaskKickoff           ReasonCode = "issue_task_kickoff"
	ReasonManualOverride             ReasonCode = "manual_override"
	ReasonPolicyDefault              ReasonCode = "policy_default_continue"
)

type SignalType string

const (
	SignalJobAdmitted               SignalType = "job_admitted"
	SignalExecutionProgressed       SignalType = "execution_progressed"
	SignalExecutionFailed           SignalType = "execution_failed"
	SignalExecutionSucceeded        SignalType = "execution_succeeded"
	SignalCheckpointSaved           SignalType = "checkpoint_saved"
	SignalTrackerAttentionNeeded    SignalType = "tracker_attention_needed"
	SignalTrackerAttentionCleared   SignalType = "tracker_attention_cleared"
	SignalSCMAttentionNeeded        SignalType = "scm_attention_needed"
	SignalSCMAttentionCleared       SignalType = "scm_attention_cleared"
	SignalPRConflictDetected        SignalType = "pr_conflict_detected"
	SignalPRReviewChangesRequested  SignalType = "pr_review_changes_requested"
	SignalPRChecksFailed            SignalType = "pr_checks_failed"
	SignalPRChecksPassed            SignalType = "pr_checks_passed"
	SignalPRMergeRequested          SignalType = "pr_merge_requested"
	SignalIssueOpened               SignalType = "issue_opened"
	SignalManualOverride            SignalType = "manual_override"
)

type CorrelationIDs struct {
	RunID  string
	TaskID string
	JobID  string
}

func (ids CorrelationIDs) Validate() error {
	if strings.TrimSpace(ids.RunID) == "" {
		return fmt.Errorf("%w: run_id is required", ErrInvalidCorrelationIDs)
	}
	if strings.TrimSpace(ids.TaskID) == "" {
		return fmt.Errorf("%w: task_id is required", ErrInvalidCorrelationIDs)
	}
	if strings.TrimSpace(ids.JobID) == "" {
		return fmt.Errorf("%w: job_id is required", ErrInvalidCorrelationIDs)
	}
	return nil
}

type Signal struct {
	Type           SignalType
	CorrelationIDs CorrelationIDs
	JobKind        string
	IdempotencyKey string
	Attempt        int
	MaxRetry       int
	FailureClass   failures.Class
	AttentionZone  AttentionZone
	OccurredAt     time.Time
	Metadata       map[string]string
}

func (signal Signal) Validate() error {
	if err := signal.CorrelationIDs.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(string(signal.Type)) == "" {
		return fmt.Errorf("%w: type is required", ErrInvalidSignal)
	}
	if signal.Attempt < 0 {
		return fmt.Errorf("%w: attempt cannot be negative", ErrInvalidSignal)
	}
	if signal.MaxRetry < 0 {
		return fmt.Errorf("%w: max_retry cannot be negative", ErrInvalidSignal)
	}
	if signal.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at is required", ErrInvalidSignal)
	}
	supportedTypes := []SignalType{
		SignalJobAdmitted,
		SignalExecutionProgressed,
		SignalExecutionFailed,
		SignalExecutionSucceeded,
		SignalCheckpointSaved,
		SignalTrackerAttentionNeeded,
		SignalTrackerAttentionCleared,
		SignalSCMAttentionNeeded,
		SignalSCMAttentionCleared,
		SignalPRConflictDetected,
		SignalPRReviewChangesRequested,
		SignalPRChecksFailed,
		SignalPRChecksPassed,
		SignalPRMergeRequested,
		SignalIssueOpened,
		SignalManualOverride,
	}
	if !slices.Contains(supportedTypes, signal.Type) {
		return fmt.Errorf("%w: unsupported signal type %q", ErrInvalidSignal, signal.Type)
	}
	if signal.Type != SignalExecutionFailed && signal.FailureClass != "" {
		return fmt.Errorf("%w: failure class is only valid for execution_failed signals", ErrInvalidSignal)
	}
	if signal.Type == SignalExecutionFailed {
		if signal.FailureClass == "" {
			return fmt.Errorf("%w: failure class is required for execution_failed signals", ErrInvalidSignal)
		}
		if signal.FailureClass != failures.ClassTerminal && signal.FailureClass != failures.ClassTransient && signal.FailureClass != failures.ClassUnknown {
			return fmt.Errorf("%w: unsupported failure class %q", ErrInvalidSignal, signal.FailureClass)
		}
	}
	return nil
}

type Decision struct {
	CorrelationIDs CorrelationIDs
	SignalType      SignalType
	FromState       State
	ToState         State
	Action          ActionCode
	Reason          ReasonCode
	RuleName        string
	RulePriority    int
	OccurredAt      time.Time
	Attempt         int
	MaxRetry        int
	FailureClass    failures.Class
	AttentionZone   AttentionZone
	Metadata        map[string]string
}

func (decision Decision) Validate() error {
	if err := decision.CorrelationIDs.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(string(decision.SignalType)) == "" {
		return fmt.Errorf("%w: signal_type is required", ErrInvalidDecision)
	}
	if strings.TrimSpace(string(decision.FromState)) == "" {
		return fmt.Errorf("%w: from_state is required", ErrInvalidDecision)
	}
	if strings.TrimSpace(string(decision.ToState)) == "" {
		return fmt.Errorf("%w: to_state is required", ErrInvalidDecision)
	}
	if strings.TrimSpace(string(decision.Action)) == "" {
		return fmt.Errorf("%w: action is required", ErrInvalidDecision)
	}
	if strings.TrimSpace(string(decision.Reason)) == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidDecision)
	}
	if strings.TrimSpace(decision.RuleName) == "" {
		return fmt.Errorf("%w: rule_name is required", ErrInvalidDecision)
	}
	if decision.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at is required", ErrInvalidDecision)
	}
	if !IsTransitionAllowed(decision.FromState, decision.ToState) {
		return fmt.Errorf("%w: transition %q -> %q is not allowed", ErrInvalidDecision, decision.FromState, decision.ToState)
	}
	return nil
}

func IsTransitionAllowed(from State, to State) bool {
	allowedTransitions := map[State][]State{
		StateIdle:       {StateIdle, StateExecuting, StateReviewing, StateRework, StateBlocked},
		StateExecuting:  {StateExecuting, StateReviewing, StateRework, StateBlocked, StateEscalated, StateCompleted},
		StateReviewing:  {StateReviewing, StateRework, StateMergeReady, StateBlocked, StateRefused},
		StateRework:     {StateRework, StateExecuting, StateReviewing, StateBlocked, StateEscalated, StateRefused},
		StateMergeReady: {StateMergeReady, StateMerged, StateRefused, StateRework},
		StateBlocked:    {StateBlocked, StateExecuting, StateReviewing, StateRework, StateEscalated, StateRefused},
		StateEscalated:  {StateEscalated, StateRework, StateRefused},
		StateMerged:     {StateMerged},
		StateRefused:    {StateRefused, StateRework},
		StateCompleted:  {StateCompleted, StateReviewing, StateMergeReady, StateMerged},
	}
	nextStates, exists := allowedTransitions[from]
	if !exists {
		return false
	}
	return slices.Contains(nextStates, to)
}
