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
}
