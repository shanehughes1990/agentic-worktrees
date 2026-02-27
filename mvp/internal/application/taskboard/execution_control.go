package taskboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type ExecutionCleaner interface {
	CleanupBoardRun(ctx context.Context, boardID string, repositoryRoot string) error
}

type ExecutionControlService struct {
	registry *ExecutionRegistry
	cleaner  ExecutionCleaner
	logger   *logrus.Logger
}

func NewExecutionControlService(registry *ExecutionRegistry, cleaner ExecutionCleaner, loggers ...*logrus.Logger) *ExecutionControlService {
	var logger *logrus.Logger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &ExecutionControlService{registry: registry, cleaner: cleaner, logger: logger}
}

func (service *ExecutionControlService) CancelAndCleanup(ctx context.Context, boardID string, repositoryRoot string) (bool, error) {
	cleanBoardID := strings.TrimSpace(boardID)
	cleanRepositoryRoot := strings.TrimSpace(repositoryRoot)
	entry := service.entry().WithFields(logrus.Fields{"event": "taskboard.execution_control.cancel_cleanup", "board_id": cleanBoardID, "repository_root": cleanRepositoryRoot})
	if cleanBoardID == "" {
		entry.Error("board_id is required")
		return false, fmt.Errorf("board_id is required")
	}
	if cleanRepositoryRoot == "" {
		entry.Error("repository_root is required")
		return false, fmt.Errorf("repository_root is required")
	}

	canceled := false
	if service.registry != nil {
		canceled = service.registry.Cancel(cleanBoardID)
	}
	entry.WithField("runner_canceled", canceled).Info("execution cancel signal processed")

	if service.cleaner != nil {
		if err := service.cleaner.CleanupBoardRun(ctx, cleanBoardID, cleanRepositoryRoot); err != nil {
			entry.WithError(err).Error("cleanup board runner artifacts failed")
			return canceled, fmt.Errorf("cleanup board runner artifacts: %w", err)
		}
	}

	entry.Info("execution cancel and cleanup completed")

	return canceled, nil
}

func (service *ExecutionControlService) entry() *logrus.Entry {
	if service == nil || service.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(service.logger)
}
