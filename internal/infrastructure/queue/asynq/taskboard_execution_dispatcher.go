package asynq

import (
	"context"

	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
	"github.com/sirupsen/logrus"
)

type TaskboardExecutionDispatcher struct {
	client *Client
	logger *logrus.Logger
}

func NewTaskboardExecutionDispatcher(client *Client, logger *logrus.Logger) *TaskboardExecutionDispatcher {
	return &TaskboardExecutionDispatcher{client: client, logger: logger}
}

func (dispatcher *TaskboardExecutionDispatcher) EnqueueTaskboardExecution(ctx context.Context, request apptaskboard.StartExecutionRequest) (string, error) {
	entry := dispatcher.entry().WithFields(logrus.Fields{
		"event":           "taskboard.execute.enqueue",
		"board_id":        request.BoardID,
		"source_branch":   request.SourceBranch,
		"repository_root": request.RepositoryRoot,
	})
	entry.Info("enqueueing taskboard execution pipeline")

	taskInfo, err := dispatcher.client.EnqueueTaskboardExecute(ctx, tasks.TaskboardExecutePayload{
		BoardID:        request.BoardID,
		SourceBranch:   request.SourceBranch,
		RepositoryRoot: request.RepositoryRoot,
		IdempotencyKey: request.BoardID + ":" + request.SourceBranch,
	})
	if err != nil {
		entry.WithError(err).Error("failed to enqueue taskboard execution pipeline")
		return "", err
	}
	if taskInfo == nil {
		entry.Warn("enqueue returned nil task info")
		return "", nil
	}

	entry.WithField("task_queue_id", taskInfo.ID).Info("taskboard execution pipeline enqueued")
	return taskInfo.ID, nil
}

func (dispatcher *TaskboardExecutionDispatcher) entry() *logrus.Entry {
	if dispatcher.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(dispatcher.logger)
}
