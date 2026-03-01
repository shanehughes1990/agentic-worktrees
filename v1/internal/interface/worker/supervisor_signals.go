package worker

import (
	"agentic-orchestrator/internal/application/taskengine"
	domainsupervisor "agentic-orchestrator/internal/domain/supervisor"
	"context"
)

type supervisorSignalService interface {
	OnExecution(ctx context.Context, record taskengine.ExecutionRecord, attempt int, maxRetry int) (domainsupervisor.Decision, error)
	OnCheckpointSaved(ctx context.Context, correlation taskengine.CorrelationIDs, jobKind taskengine.JobKind, idempotencyKey string, step string) (domainsupervisor.Decision, error)
	OnTrackerAttention(ctx context.Context, correlation taskengine.CorrelationIDs, reason string) (domainsupervisor.Decision, error)
	OnIssueOpened(ctx context.Context, correlation taskengine.CorrelationIDs, source string, issueReference string) (domainsupervisor.Decision, error)
	OnIssueApproved(ctx context.Context, correlation taskengine.CorrelationIDs, source string, issueReference string, approvedBy string) (domainsupervisor.Decision, error)
	OnPRChecksEvaluated(ctx context.Context, correlation taskengine.CorrelationIDs, provider string, owner string, repository string, pullRequestNumber int, canMerge bool, reason string) (domainsupervisor.Decision, error)
	OnPRMergeRequested(ctx context.Context, correlation taskengine.CorrelationIDs, provider string, owner string, repository string, pullRequestNumber int, mergeMethod string) (domainsupervisor.Decision, error)
}
