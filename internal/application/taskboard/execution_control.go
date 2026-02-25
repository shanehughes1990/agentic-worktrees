package taskboard

import (
	"context"
	"fmt"
	"strings"
)

type ExecutionCleaner interface {
	CleanupBoardRun(ctx context.Context, boardID string, repositoryRoot string) error
}

type ExecutionControlService struct {
	registry *ExecutionRegistry
	cleaner  ExecutionCleaner
}

func NewExecutionControlService(registry *ExecutionRegistry, cleaner ExecutionCleaner) *ExecutionControlService {
	return &ExecutionControlService{registry: registry, cleaner: cleaner}
}

func (service *ExecutionControlService) CancelAndCleanup(ctx context.Context, boardID string, repositoryRoot string) (bool, error) {
	cleanBoardID := strings.TrimSpace(boardID)
	cleanRepositoryRoot := strings.TrimSpace(repositoryRoot)
	if cleanBoardID == "" {
		return false, fmt.Errorf("board_id is required")
	}
	if cleanRepositoryRoot == "" {
		return false, fmt.Errorf("repository_root is required")
	}

	canceled := false
	if service.registry != nil {
		canceled = service.registry.Cancel(cleanBoardID)
	}

	if service.cleaner != nil {
		if err := service.cleaner.CleanupBoardRun(ctx, cleanBoardID, cleanRepositoryRoot); err != nil {
			return canceled, fmt.Errorf("cleanup board runner artifacts: %w", err)
		}
	}

	return canceled, nil
}
