package taskboard

import (
	"context"
	"fmt"
	"strings"
)

type StartExecutionRequest struct {
	BoardID        string
	SourceBranch   string
	RepositoryRoot string
}

type ExecutionDispatcher interface {
	EnqueueTaskboardExecution(ctx context.Context, request StartExecutionRequest) (string, error)
}

type ExecutionCommandService struct {
	dispatcher ExecutionDispatcher
}

func NewExecutionCommandService(dispatcher ExecutionDispatcher) *ExecutionCommandService {
	return &ExecutionCommandService{dispatcher: dispatcher}
}

func (service *ExecutionCommandService) Start(ctx context.Context, request StartExecutionRequest) (string, error) {
	if service.dispatcher == nil {
		return "", fmt.Errorf("dispatcher is required")
	}
	request.BoardID = strings.TrimSpace(request.BoardID)
	request.SourceBranch = strings.TrimSpace(request.SourceBranch)
	request.RepositoryRoot = strings.TrimSpace(request.RepositoryRoot)

	if request.BoardID == "" {
		return "", fmt.Errorf("board_id is required")
	}
	if request.SourceBranch == "" {
		return "", fmt.Errorf("source_branch is required")
	}
	if request.RepositoryRoot == "" {
		return "", fmt.Errorf("repository_root is required")
	}

	taskID, err := service.dispatcher.EnqueueTaskboardExecution(ctx, request)
	if err != nil {
		return "", fmt.Errorf("enqueue taskboard execution: %w", err)
	}
	return taskID, nil
}
