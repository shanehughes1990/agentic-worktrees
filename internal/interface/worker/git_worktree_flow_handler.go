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

type taskExecutor interface {
	ExecuteTask(ctx context.Context, request appgitflow.TaskExecutionRequest) (appgitflow.TaskExecutionResult, error)
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
	entry.Info("processing git worktree flow task")

	boardID := strings.TrimSpace(payload.BoardID)
	if boardID == "" {
		boardID = strings.TrimSpace(payload.RunID)
	}

	if task, getErr := handler.taskboardService.GetTaskByID(ctx, boardID, strings.TrimSpace(payload.TaskID)); getErr == nil && task != nil && task.Outcome != nil {
		persistedResumeSessionID := strings.TrimSpace(task.Outcome.ResumeSessionID)
		if persistedResumeSessionID != "" {
			if strings.TrimSpace(payload.ResumeSessionID) != persistedResumeSessionID {
				entry.WithField("resume_session_id", persistedResumeSessionID).Info("using latest persisted resume session for recovered task execution")
			}
			payload.ResumeSessionID = persistedResumeSessionID
		}
	}

	result, err := handler.executor.ExecuteTask(ctx, appgitflow.TaskExecutionRequest{
		BoardID:         boardID,
		RunID:           payload.RunID,
		TaskID:          payload.TaskID,
		TaskTitle:       payload.TaskTitle,
		TaskDetail:      payload.TaskDetail,
		ResumeSessionID: payload.ResumeSessionID,
		SourceBranch:    payload.SourceBranch,
		RepositoryRoot:  payload.RepositoryRoot,
		WorktreePath:    payload.WorktreePath,
	})
	if err == nil {
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
		entry.WithError(err).Warn("git worktree flow task interrupted; retrying for automatic resume")
		return fmt.Errorf("git worktree flow interrupted: %w", err)
	}

	if markErr := handler.taskboardService.MarkTaskBlockedWithOutcome(ctx, boardID, strings.TrimSpace(payload.TaskID), outcome); markErr != nil {
		entry.WithError(markErr).Error("failed to mark task blocked after git worktree flow")
		return fmt.Errorf("mark task blocked: %w", markErr)
	}
	entry.WithError(err).Error("git worktree flow task failed")
	if appgitflow.IsTerminalFailure(err) {
		return fmt.Errorf("%w: git worktree flow: %v", asynq.SkipRetry, err)
	}
	return fmt.Errorf("git worktree flow: %w", err)
}

func (handler *GitWorktreeFlowHandler) entry() *logrus.Entry {
	if handler.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(handler.logger)
}
