package taskboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type StartExecutionRequest struct {
	BoardID        string
	SourceBranch   string
	RepositoryRoot string
	MaxTasks       int
}

type ExecutionDispatcher interface {
	EnqueueTaskboardExecution(ctx context.Context, request StartExecutionRequest) (string, error)
}

type ExecutionCommandService struct {
	dispatcher ExecutionDispatcher
	logger     *logrus.Logger
}

func NewExecutionCommandService(dispatcher ExecutionDispatcher, loggers ...*logrus.Logger) *ExecutionCommandService {
	var logger *logrus.Logger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &ExecutionCommandService{dispatcher: dispatcher, logger: logger}
}

func (service *ExecutionCommandService) Start(ctx context.Context, request StartExecutionRequest) (string, error) {
	entry := service.entry().WithFields(logrus.Fields{
		"event":           "taskboard.execution_command.start",
		"board_id":        strings.TrimSpace(request.BoardID),
		"source_branch":   strings.TrimSpace(request.SourceBranch),
		"repository_root": strings.TrimSpace(request.RepositoryRoot),
		"max_tasks":       request.MaxTasks,
	})
	if service.dispatcher == nil {
		entry.Error("dispatcher is required")
		return "", fmt.Errorf("dispatcher is required")
	}
	request.BoardID = strings.TrimSpace(request.BoardID)
	request.SourceBranch = strings.TrimSpace(request.SourceBranch)
	request.RepositoryRoot = strings.TrimSpace(request.RepositoryRoot)

	if request.BoardID == "" {
		entry.Error("board_id is required")
		return "", fmt.Errorf("board_id is required")
	}
	if request.SourceBranch == "" {
		entry.Error("source_branch is required")
		return "", fmt.Errorf("source_branch is required")
	}
	if request.RepositoryRoot == "" {
		entry.Error("repository_root is required")
		return "", fmt.Errorf("repository_root is required")
	}
	if request.MaxTasks < 0 {
		entry.Error("max_tasks cannot be negative")
		return "", fmt.Errorf("max_tasks cannot be negative")
	}

	taskID, err := service.dispatcher.EnqueueTaskboardExecution(ctx, request)
	if err != nil {
		entry.WithError(err).Error("failed to enqueue taskboard execution")
		return "", fmt.Errorf("enqueue taskboard execution: %w", err)
	}
	entry.WithField("queue_task_id", taskID).Info("enqueued taskboard execution")
	return taskID, nil
}

func (service *ExecutionCommandService) entry() *logrus.Entry {
	if service == nil || service.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(service.logger)
}
