package resolvers

import (
	applicationsupervisor "agentic-orchestrator/internal/application/supervisor"
	domainfailures "agentic-orchestrator/internal/domain/failures"
	domainsupervisor "agentic-orchestrator/internal/domain/supervisor"
	"agentic-orchestrator/internal/interface/graphql/models"
	"context"
	"fmt"
	"sort"
)

func toSupervisorSignalType(value domainsupervisor.SignalType) models.SupervisorSignalType {
	switch value {
	case domainsupervisor.SignalJobAdmitted:
		return models.SupervisorSignalTypeJobAdmitted
	case domainsupervisor.SignalExecutionProgressed:
		return models.SupervisorSignalTypeExecutionProgressed
	case domainsupervisor.SignalExecutionFailed:
		return models.SupervisorSignalTypeExecutionFailed
	case domainsupervisor.SignalExecutionSucceeded:
		return models.SupervisorSignalTypeExecutionSucceeded
	case domainsupervisor.SignalCheckpointSaved:
		return models.SupervisorSignalTypeCheckpointSaved
	case domainsupervisor.SignalTrackerAttentionNeeded:
		return models.SupervisorSignalTypeTrackerAttentionNeeded
	case domainsupervisor.SignalTrackerAttentionCleared:
		return models.SupervisorSignalTypeTrackerAttentionCleared
	case domainsupervisor.SignalSCMAttentionNeeded:
		return models.SupervisorSignalTypeScmAttentionNeeded
	case domainsupervisor.SignalSCMAttentionCleared:
		return models.SupervisorSignalTypeScmAttentionCleared
	case domainsupervisor.SignalPRConflictDetected:
		return models.SupervisorSignalTypePrConflictDetected
	case domainsupervisor.SignalPRReviewChangesRequested:
		return models.SupervisorSignalTypePrReviewChangesRequested
	case domainsupervisor.SignalPRChecksFailed:
		return models.SupervisorSignalTypePrChecksFailed
	case domainsupervisor.SignalPRChecksPassed:
		return models.SupervisorSignalTypePrChecksPassed
	case domainsupervisor.SignalPRMergeRequested:
		return models.SupervisorSignalTypePrMergeRequested
	case domainsupervisor.SignalIssueOpened:
		return models.SupervisorSignalTypeIssueOpened
	case domainsupervisor.SignalIssueApproved:
		return models.SupervisorSignalTypeIssueApproved
	default:
		return models.SupervisorSignalTypeManualOverride
	}
}

func toSupervisorState(value domainsupervisor.State) models.SupervisorState {
	switch value {
	case domainsupervisor.StateIdle:
		return models.SupervisorStateIdle
	case domainsupervisor.StateExecuting:
		return models.SupervisorStateExecuting
	case domainsupervisor.StateReviewing:
		return models.SupervisorStateReviewing
	case domainsupervisor.StateRework:
		return models.SupervisorStateRework
	case domainsupervisor.StateMergeReady:
		return models.SupervisorStateMergeReady
	case domainsupervisor.StateBlocked:
		return models.SupervisorStateBlocked
	case domainsupervisor.StateEscalated:
		return models.SupervisorStateEscalated
	case domainsupervisor.StateMerged:
		return models.SupervisorStateMerged
	case domainsupervisor.StateRefused:
		return models.SupervisorStateRefused
	default:
		return models.SupervisorStateCompleted
	}
}

func toSupervisorAction(value domainsupervisor.ActionCode) models.SupervisorActionCode {
	switch value {
	case domainsupervisor.ActionContinue:
		return models.SupervisorActionCodeContinue
	case domainsupervisor.ActionRetry:
		return models.SupervisorActionCodeRetry
	case domainsupervisor.ActionBlock:
		return models.SupervisorActionCodeBlock
	case domainsupervisor.ActionEscalate:
		return models.SupervisorActionCodeEscalate
	case domainsupervisor.ActionRequestRework:
		return models.SupervisorActionCodeRequestRework
	case domainsupervisor.ActionMerge:
		return models.SupervisorActionCodeMerge
	case domainsupervisor.ActionRefuse:
		return models.SupervisorActionCodeRefuse
	default:
		return models.SupervisorActionCodeStartTask
	}
}

func toSupervisorReason(value domainsupervisor.ReasonCode) models.SupervisorReasonCode {
	switch value {
	case domainsupervisor.ReasonJobAdmitted:
		return models.SupervisorReasonCodeJobAdmitted
	case domainsupervisor.ReasonExecutionProgressed:
		return models.SupervisorReasonCodeExecutionProgressed
	case domainsupervisor.ReasonExecutionSucceeded:
		return models.SupervisorReasonCodeExecutionSucceeded
	case domainsupervisor.ReasonExecutionFailedRetry:
		return models.SupervisorReasonCodeExecutionFailedRetry
	case domainsupervisor.ReasonExecutionFailedMaxed:
		return models.SupervisorReasonCodeExecutionFailedMaxRetries
	case domainsupervisor.ReasonExecutionFailedFatal:
		return models.SupervisorReasonCodeExecutionFailedTerminal
	case domainsupervisor.ReasonTrackerAttention:
		return models.SupervisorReasonCodeTrackerAttentionRequired
	case domainsupervisor.ReasonTrackerAttentionClear:
		return models.SupervisorReasonCodeTrackerAttentionCleared
	case domainsupervisor.ReasonSCMAttention:
		return models.SupervisorReasonCodeScmAttentionRequired
	case domainsupervisor.ReasonSCMAttentionClear:
		return models.SupervisorReasonCodeScmAttentionCleared
	case domainsupervisor.ReasonPRConflictDetected:
		return models.SupervisorReasonCodePrConflictDetected
	case domainsupervisor.ReasonPRReviewChangesRequested:
		return models.SupervisorReasonCodePrReviewChangesRequested
	case domainsupervisor.ReasonPRChecksFailed:
		return models.SupervisorReasonCodePrChecksFailed
	case domainsupervisor.ReasonPRChecksPassed:
		return models.SupervisorReasonCodePrChecksPassed
	case domainsupervisor.ReasonPRMergeApproved:
		return models.SupervisorReasonCodePrMergeApproved
	case domainsupervisor.ReasonPRMergeRefused:
		return models.SupervisorReasonCodePrMergeRefused
	case domainsupervisor.ReasonIssueAwaitingApproval:
		return models.SupervisorReasonCodeIssueAwaitingApproval
	case domainsupervisor.ReasonIssueTaskKickoff:
		return models.SupervisorReasonCodeIssueTaskKickoff
	case domainsupervisor.ReasonManualOverride:
		return models.SupervisorReasonCodeManualOverride
	default:
		return models.SupervisorReasonCodePolicyDefaultContinue
	}
}

func toSupervisorAttentionZone(value domainsupervisor.AttentionZone) models.SupervisorAttentionZone {
	switch value {
	case domainsupervisor.AttentionZoneTracker:
		return models.SupervisorAttentionZoneTracker
	case domainsupervisor.AttentionZoneSCM:
		return models.SupervisorAttentionZoneScm
	case domainsupervisor.AttentionZoneExecution:
		return models.SupervisorAttentionZoneExecution
	default:
		return models.SupervisorAttentionZoneNone
	}
}

func toFailureClass(value domainfailures.Class) models.FailureClass {
	switch value {
	case domainfailures.ClassTerminal:
		return models.FailureClassTerminal
	case domainfailures.ClassTransient:
		return models.FailureClassTransient
	default:
		return models.FailureClassUnknown
	}
}

func mapSupervisorDecision(decision domainsupervisor.Decision) *models.SupervisorDecision {
	entries := make([]*models.SupervisorDecisionMetadataEntry, 0, len(decision.Metadata))
	for key, value := range decision.Metadata {
		entries = append(entries, &models.SupervisorDecisionMetadataEntry{Key: key, Value: value})
	}
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].Key < entries[right].Key
	})
	return &models.SupervisorDecision{
		RunID:         decision.CorrelationIDs.RunID,
		TaskID:        decision.CorrelationIDs.TaskID,
		JobID:         decision.CorrelationIDs.JobID,
		SignalType:    toSupervisorSignalType(decision.SignalType),
		FromState:     toSupervisorState(decision.FromState),
		ToState:       toSupervisorState(decision.ToState),
		Action:        toSupervisorAction(decision.Action),
		Reason:        toSupervisorReason(decision.Reason),
		RuleName:      decision.RuleName,
		RulePriority:  int32(decision.RulePriority),
		OccurredAt:    decision.OccurredAt.UTC(),
		AttentionZone: toSupervisorAttentionZone(decision.AttentionZone),
		Attempt:       int32(decision.Attempt),
		MaxRetry:      int32(decision.MaxRetry),
		FailureClass:  toFailureClass(decision.FailureClass),
		Metadata:      entries,
	}
}

func loadSupervisorHistoryResult(ctx context.Context, supervisorService *applicationsupervisor.Service, correlation models.SupervisorCorrelationInput) models.SupervisorDecisionHistoryResult {
	if supervisorService == nil {
		return models.GraphError{Code: models.GraphErrorCodeUnavailable, Message: "supervisor service is not configured"}
	}
	history, err := supervisorService.History(ctx, domainsupervisor.CorrelationIDs{RunID: correlation.RunID, TaskID: correlation.TaskID, JobID: correlation.JobID})
	if err != nil {
		return graphErrorFromError(fmt.Errorf("load supervisor history: %w", err))
	}
	mapped := make([]*models.SupervisorDecision, 0, len(history))
	for _, decision := range history {
		mapped = append(mapped, mapSupervisorDecision(decision))
	}
	return models.SupervisorDecisionHistorySuccess{Decisions: mapped}
}
