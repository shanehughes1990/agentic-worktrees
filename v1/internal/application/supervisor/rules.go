package supervisor

import (
	"agentic-orchestrator/internal/domain/failures"
	domainsupervisor "agentic-orchestrator/internal/domain/supervisor"
	"context"
)

type rule struct {
	name     string
	priority int
	evaluate func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool)
}

func (rule rule) Name() string {
	return rule.name
}

func (rule rule) Priority() int {
	return rule.priority
}

func (rule rule) Evaluate(ctx context.Context, state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool, error) {
	_ = ctx
	decision, matched := rule.evaluate(state, signal)
	return decision, matched, nil
}

func DefaultRules() []domainsupervisor.Rule {
	return []domainsupervisor.Rule{
		rule{name: "pr-merge-request-ready", priority: 120, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalPRMergeRequested || state != domainsupervisor.StateMergeReady {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateMerged, Action: domainsupervisor.ActionMerge, Reason: domainsupervisor.ReasonPRMergeApproved, AttentionZone: domainsupervisor.AttentionZoneSCM}, true
		}},
		rule{name: "pr-merge-request-not-ready", priority: 119, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalPRMergeRequested || state == domainsupervisor.StateMergeReady {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateRefused, Action: domainsupervisor.ActionRefuse, Reason: domainsupervisor.ReasonPRMergeRefused, AttentionZone: domainsupervisor.AttentionZoneSCM}, true
		}},
		rule{name: "pr-conflict-detected", priority: 115, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalPRConflictDetected {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateRework, Action: domainsupervisor.ActionRequestRework, Reason: domainsupervisor.ReasonPRConflictDetected, AttentionZone: domainsupervisor.AttentionZoneSCM}, true
		}},
		rule{name: "pr-review-changes-requested", priority: 114, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalPRReviewChangesRequested {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateRework, Action: domainsupervisor.ActionRequestRework, Reason: domainsupervisor.ReasonPRReviewChangesRequested, AttentionZone: domainsupervisor.AttentionZoneSCM}, true
		}},
		rule{name: "pr-checks-failed", priority: 113, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalPRChecksFailed {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateBlocked, Action: domainsupervisor.ActionRequestRework, Reason: domainsupervisor.ReasonPRChecksFailed, AttentionZone: domainsupervisor.AttentionZoneSCM}, true
		}},
		rule{name: "pr-checks-passed", priority: 112, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalPRChecksPassed {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateMergeReady, Action: domainsupervisor.ActionContinue, Reason: domainsupervisor.ReasonPRChecksPassed, AttentionZone: domainsupervisor.AttentionZoneSCM}, true
		}},
		rule{name: "issue-opened-kickoff", priority: 110, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalIssueOpened {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateExecuting, Action: domainsupervisor.ActionStartTask, Reason: domainsupervisor.ReasonIssueTaskKickoff, AttentionZone: domainsupervisor.AttentionZoneTracker}, true
		}},
		rule{name: "tracker-attention-needed", priority: 100, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalTrackerAttentionNeeded {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateBlocked, Action: domainsupervisor.ActionBlock, Reason: domainsupervisor.ReasonTrackerAttention, AttentionZone: domainsupervisor.AttentionZoneTracker}, true
		}},
		rule{name: "tracker-attention-cleared", priority: 95, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalTrackerAttentionCleared {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateExecuting, Action: domainsupervisor.ActionContinue, Reason: domainsupervisor.ReasonTrackerAttentionClear, AttentionZone: domainsupervisor.AttentionZoneTracker}, true
		}},
		rule{name: "scm-attention-needed", priority: 94, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalSCMAttentionNeeded {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateBlocked, Action: domainsupervisor.ActionBlock, Reason: domainsupervisor.ReasonSCMAttention, AttentionZone: domainsupervisor.AttentionZoneSCM}, true
		}},
		rule{name: "scm-attention-cleared", priority: 93, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalSCMAttentionCleared {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateExecuting, Action: domainsupervisor.ActionContinue, Reason: domainsupervisor.ReasonSCMAttentionClear, AttentionZone: domainsupervisor.AttentionZoneSCM}, true
		}},
		rule{name: "execution-failed-terminal", priority: 90, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalExecutionFailed || signal.FailureClass != failures.ClassTerminal {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateEscalated, Action: domainsupervisor.ActionEscalate, Reason: domainsupervisor.ReasonExecutionFailedFatal, AttentionZone: domainsupervisor.AttentionZoneExecution}, true
		}},
		rule{name: "execution-failed-max-retries", priority: 89, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalExecutionFailed {
				return domainsupervisor.Decision{}, false
			}
			if signal.MaxRetry > 0 && signal.Attempt >= signal.MaxRetry {
				return domainsupervisor.Decision{ToState: domainsupervisor.StateEscalated, Action: domainsupervisor.ActionEscalate, Reason: domainsupervisor.ReasonExecutionFailedMaxed, AttentionZone: domainsupervisor.AttentionZoneExecution}, true
			}
			return domainsupervisor.Decision{}, false
		}},
		rule{name: "execution-failed-retry", priority: 88, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalExecutionFailed {
				return domainsupervisor.Decision{}, false
			}
			if signal.FailureClass != failures.ClassTransient {
				return domainsupervisor.Decision{}, false
			}
			if signal.MaxRetry > 0 && signal.Attempt < signal.MaxRetry {
				return domainsupervisor.Decision{ToState: domainsupervisor.StateExecuting, Action: domainsupervisor.ActionRetry, Reason: domainsupervisor.ReasonExecutionFailedRetry, AttentionZone: domainsupervisor.AttentionZoneExecution}, true
			}
			if signal.MaxRetry <= 0 {
				return domainsupervisor.Decision{ToState: domainsupervisor.StateRework, Action: domainsupervisor.ActionRequestRework, Reason: domainsupervisor.ReasonExecutionFailedRetry, AttentionZone: domainsupervisor.AttentionZoneExecution}, true
			}
			return domainsupervisor.Decision{}, false
		}},
		rule{name: "execution-succeeded", priority: 70, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalExecutionSucceeded {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateCompleted, Action: domainsupervisor.ActionContinue, Reason: domainsupervisor.ReasonExecutionSucceeded, AttentionZone: domainsupervisor.AttentionZoneNone}, true
		}},
		rule{name: "execution-progress", priority: 60, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalExecutionProgressed && signal.Type != domainsupervisor.SignalCheckpointSaved {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateExecuting, Action: domainsupervisor.ActionContinue, Reason: domainsupervisor.ReasonExecutionProgressed, AttentionZone: domainsupervisor.AttentionZoneNone}, true
		}},
		rule{name: "job-admitted", priority: 50, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalJobAdmitted {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateExecuting, Action: domainsupervisor.ActionContinue, Reason: domainsupervisor.ReasonJobAdmitted, AttentionZone: domainsupervisor.AttentionZoneNone}, true
		}},
		rule{name: "manual-override", priority: 40, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			if signal.Type != domainsupervisor.SignalManualOverride {
				return domainsupervisor.Decision{}, false
			}
			return domainsupervisor.Decision{ToState: domainsupervisor.StateExecuting, Action: domainsupervisor.ActionContinue, Reason: domainsupervisor.ReasonManualOverride, AttentionZone: domainsupervisor.AttentionZoneNone}, true
		}},
		rule{name: "default-continue", priority: 0, evaluate: func(state domainsupervisor.State, signal domainsupervisor.Signal) (domainsupervisor.Decision, bool) {
			return domainsupervisor.Decision{ToState: state, Action: domainsupervisor.ActionContinue, Reason: domainsupervisor.ReasonPolicyDefault, AttentionZone: domainsupervisor.AttentionZoneNone}, true
		}},
	}
}
