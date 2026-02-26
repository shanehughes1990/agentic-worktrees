package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
	"github.com/sirupsen/logrus"
)

type GitWorktreeFlowHandler struct {
	executor         taskExecutor
	taskboardService *apptaskboard.Service
	logger           *logrus.Logger
}

const noProgressOutcomeStatus = "no_changes"

var readRetryContext = func(ctx context.Context) (int, int) {
	retryCount, ok := asynq.GetRetryCount(ctx)
	if !ok || retryCount < 0 {
		retryCount = 0
	}
	maxRetry, ok := asynq.GetMaxRetry(ctx)
	if !ok || maxRetry < 0 {
		maxRetry = 0
	}
	return retryCount, maxRetry
}

type taskExecutor interface {
	ExecuteTask(ctx context.Context, request appgitflow.TaskExecutionRequest) (appgitflow.TaskExecutionResult, error)
	ReconcileCompletedTaskWorktree(ctx context.Context, request appgitflow.TaskExecutionRequest) error
}

func NewGitWorktreeFlowHandler(executor taskExecutor, taskboardService *apptaskboard.Service, logger *logrus.Logger) *GitWorktreeFlowHandler {
	return &GitWorktreeFlowHandler{executor: executor, taskboardService: taskboardService, logger: logger}
}

func (handler *GitWorktreeFlowHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if handler.executor == nil {
		return fmt.Errorf("task executor is required")
	}
	if handler.taskboardService == nil {
		return fmt.Errorf("taskboard service is required")
	}

	var payload tasks.GitWorktreeFlowPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode git worktree flow payload: %w", err)
	}

	entry := handler.entry().WithFields(logrus.Fields{
		"event":           "worker.git_worktree_flow",
		"run_id":          strings.TrimSpace(payload.RunID),
		"task_id":         strings.TrimSpace(payload.TaskID),
		"task_type":       task.Type(),
		"source_branch":   strings.TrimSpace(payload.SourceBranch),
		"task_branch":     strings.TrimSpace(payload.TaskBranch),
		"worktree_path":   strings.TrimSpace(payload.WorktreePath),
		"repository_root": strings.TrimSpace(payload.RepositoryRoot),
	})
	queueTaskID, _ := asynq.GetTaskID(ctx)
	cleanQueueTaskID := strings.TrimSpace(queueTaskID)
	correlationID := buildCorrelationID(strings.TrimSpace(payload.RunID), cleanQueueTaskID)
	tracePathID := correlationID
	if tracePathID == "" {
		tracePathID = strings.TrimSpace(payload.RunID)
	}
	if tracePathID == "" {
		tracePathID = strings.TrimSpace(payload.TaskID)
	}
	entry = entry.WithFields(logrus.Fields{"queue_task_id": cleanQueueTaskID, "correlation_id": correlationID, "trace_path_id": tracePathID})
	retryCount, maxRetry := readRetryContext(ctx)
	attempt := retryCount + 1
	entry = entry.WithFields(logrus.Fields{
		"retry_count": retryCount,
		"max_retry":   maxRetry,
		"attempt":     attempt,
	})
	entry.WithFields(logrus.Fields{"phase": "received", "resume_session_id": strings.TrimSpace(payload.ResumeSessionID)}).Info("git worktree flow state snapshot")
	entry.Info("processing git worktree flow task")

	boardID := strings.TrimSpace(payload.BoardID)
	if boardID == "" {
		boardID = strings.TrimSpace(payload.RunID)
	}

	if taskState, getErr := handler.taskboardService.GetTaskByID(ctx, boardID, strings.TrimSpace(payload.TaskID)); getErr == nil && taskState != nil {
		if taskState.Status == domaintaskboard.StatusCompleted && taskState.Outcome != nil && strings.TrimSpace(taskState.Outcome.Worktree) != "" {
			reconcileRequest := appgitflow.TaskExecutionRequest{
				BoardID:       boardID,
				RunID:         payload.RunID,
				TaskID:        payload.TaskID,
				TaskBranch:    strings.TrimSpace(taskState.Outcome.TaskBranch),
				QueueTaskID:   cleanQueueTaskID,
				CorrelationID: correlationID,
				SourceBranch:  payload.SourceBranch,
				RepositoryRoot: payload.RepositoryRoot,
				WorktreePath:  strings.TrimSpace(taskState.Outcome.Worktree),
			}
			if reconcileErr := handler.executor.ReconcileCompletedTaskWorktree(ctx, reconcileRequest); reconcileErr != nil {
				entry.WithError(reconcileErr).WithFields(logrus.Fields{"phase": "completed_reconcile_failed", "task_status": taskState.Status}).Error("completed task stale worktree reconcile failed")
				return fmt.Errorf("reconcile completed task worktree: %w", reconcileErr)
			}
			entry.WithFields(logrus.Fields{"phase": "completed_reconcile_skipped_execution", "task_status": taskState.Status}).Info("completed task had stale worktree reconciled; skipping execution")
			return nil
		}

		if taskState.Outcome != nil {
			persistedResumeSessionID := strings.TrimSpace(taskState.Outcome.ResumeSessionID)
			if persistedResumeSessionID != "" {
				if strings.TrimSpace(payload.ResumeSessionID) != persistedResumeSessionID {
					entry.WithField("resume_session_id", persistedResumeSessionID).Info("using latest persisted resume session for recovered task execution")
				}
				payload.ResumeSessionID = persistedResumeSessionID
			}
		}
	}

	result, err := handler.executor.ExecuteTask(ctx, appgitflow.TaskExecutionRequest{
		BoardID:         boardID,
		RunID:           payload.RunID,
		TaskID:          payload.TaskID,
		QueueTaskID:     cleanQueueTaskID,
		CorrelationID:   correlationID,
		TaskTitle:       payload.TaskTitle,
		TaskDetail:      payload.TaskDetail,
		ResumeSessionID: payload.ResumeSessionID,
		ExecutionAttempt: attempt,
		SourceBranch:    payload.SourceBranch,
		RepositoryRoot:  payload.RepositoryRoot,
		WorktreePath:    payload.WorktreePath,
	})
	if err == nil {
		if strings.EqualFold(strings.TrimSpace(result.Status), noProgressOutcomeStatus) {
			outcome := domaintaskboard.TaskOutcome{
				Status:          "interrupted",
				Reason:          fmt.Sprintf("no forward progress detected; auto-retry attempt %d/%d", attempt, maxRetry+1),
				TaskBranch:      strings.TrimSpace(result.TaskBranch),
				Worktree:        strings.TrimSpace(result.Worktree),
				ResumeSessionID: strings.TrimSpace(result.ResumeSessionID),
			}
			if retryCount < maxRetry {
				if markErr := handler.taskboardService.MarkTaskCanceledWithOutcome(ctx, boardID, strings.TrimSpace(payload.TaskID), outcome); markErr != nil {
					entry.WithError(markErr).Error("failed to requeue task after no-progress result")
					return fmt.Errorf("requeue task after no-progress result: %w", markErr)
				}
				entry.WithFields(logrus.Fields{"phase": "no_progress_retry", "status": outcome.Status, "reason": outcome.Reason, "resume_session_id": outcome.ResumeSessionID, "retry_next_attempt": attempt + 1}).Warn("git worktree flow state snapshot")
				entry.WithField("retry_next_attempt", attempt+1).Warn("git worktree flow returned no progress; triggering automatic retry")
				return fmt.Errorf("git worktree flow no progress on attempt %d/%d", attempt, maxRetry+1)
			}

			outcome.Status = "blocked"
			outcome.Reason = fmt.Sprintf("no forward progress after %d attempts", maxRetry+1)
			if markErr := handler.taskboardService.MarkTaskBlockedWithOutcome(ctx, boardID, strings.TrimSpace(payload.TaskID), outcome); markErr != nil {
				entry.WithError(markErr).Error("failed to mark task blocked after no-progress retry exhaustion")
				return fmt.Errorf("mark task blocked after no-progress retry exhaustion: %w", markErr)
			}
			entry.WithFields(logrus.Fields{"phase": "no_progress_exhausted", "status": outcome.Status, "reason": outcome.Reason, "resume_session_id": outcome.ResumeSessionID}).Error("git worktree flow state snapshot")
			entry.Error("git worktree flow no-progress retries exhausted")
			return fmt.Errorf("%w: git worktree flow no progress after retries", asynq.SkipRetry)
		}

		markErr := handler.taskboardService.MarkTaskCompletedWithOutcome(ctx, boardID, strings.TrimSpace(payload.TaskID), domaintaskboard.TaskOutcome{
			Status:          result.Status,
			Reason:          result.Reason,
			TaskBranch:      result.TaskBranch,
			Worktree:        result.Worktree,
			ResumeSessionID: result.ResumeSessionID,
		})
		if markErr != nil {
			entry.WithError(markErr).Error("failed to mark task completed after git worktree flow")
			return fmt.Errorf("mark task completed: %w", markErr)
		}
		entry.WithFields(logrus.Fields{"phase": "completed", "status": result.Status, "reason": result.Reason, "resume_session_id": result.ResumeSessionID}).Info("git worktree flow state snapshot")
		entry.Info("git worktree flow task completed")
		return nil
	}

	outcome := domaintaskboard.TaskOutcome{
		Status:          strings.TrimSpace(result.Status),
		Reason:          strings.TrimSpace(result.Reason),
		TaskBranch:      strings.TrimSpace(result.TaskBranch),
		Worktree:        strings.TrimSpace(result.Worktree),
		ResumeSessionID: strings.TrimSpace(result.ResumeSessionID),
	}
	if outcome.Status == "" {
		outcome.Status = "failed"
	}
	if outcome.Reason == "" {
		outcome.Reason = err.Error()
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		if outcome.Status == "failed" {
			outcome.Status = "canceled"
		}
		if markErr := handler.taskboardService.MarkTaskCanceledWithOutcome(ctx, boardID, strings.TrimSpace(payload.TaskID), outcome); markErr != nil {
			entry.WithError(markErr).Error("failed to mark task canceled after git worktree flow")
			return fmt.Errorf("mark task canceled: %w", markErr)
		}
		entry.WithFields(logrus.Fields{"phase": "interrupted", "status": outcome.Status, "reason": outcome.Reason, "resume_session_id": outcome.ResumeSessionID}).WithError(err).Warn("git worktree flow state snapshot")
		entry.WithError(err).Warn("git worktree flow task interrupted; retrying for automatic resume")
		return fmt.Errorf("git worktree flow interrupted: %w", err)
	}

	if !appgitflow.IsTerminalFailure(err) {
		if strings.TrimSpace(outcome.Status) == "" || strings.EqualFold(strings.TrimSpace(outcome.Status), "failed") {
			outcome.Status = "interrupted"
		}
		if strings.TrimSpace(outcome.Reason) == "" {
			outcome.Reason = "transient task execution failure"
		}
		if markErr := handler.taskboardService.MarkTaskCanceledWithOutcome(ctx, boardID, strings.TrimSpace(payload.TaskID), outcome); markErr != nil {
			entry.WithError(markErr).Error("failed to requeue task after transient git worktree flow failure")
			return fmt.Errorf("requeue task after transient failure: %w", markErr)
		}
		entry.WithFields(logrus.Fields{"phase": "transient_failure", "status": outcome.Status, "reason": outcome.Reason, "resume_session_id": outcome.ResumeSessionID}).WithError(err).Warn("git worktree flow state snapshot")
		entry.WithError(err).Warn("git worktree flow transient failure; task re-queued for automatic restart")
		return fmt.Errorf("git worktree flow transient failure: %w", err)
	}

	if markErr := handler.taskboardService.MarkTaskBlockedWithOutcome(ctx, boardID, strings.TrimSpace(payload.TaskID), outcome); markErr != nil {
		entry.WithError(markErr).Error("failed to mark task blocked after git worktree flow")
		return fmt.Errorf("mark task blocked: %w", markErr)
	}
	entry.WithFields(logrus.Fields{"phase": "terminal_failure", "status": outcome.Status, "reason": outcome.Reason, "resume_session_id": outcome.ResumeSessionID}).WithError(err).Error("git worktree flow state snapshot")
	entry.WithError(err).Error("git worktree flow task failed")
	return fmt.Errorf("%w: git worktree flow: %v", asynq.SkipRetry, err)
}

func (handler *GitWorktreeFlowHandler) entry() *logrus.Entry {
	if handler.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(handler.logger)
}
