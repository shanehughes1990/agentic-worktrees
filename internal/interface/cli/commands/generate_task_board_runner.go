package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"

	"github.com/shanehughes1990/agentic-worktrees/internal/application"
	infraqueue "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue"
)

type GenerateTaskBoardExecutionDependencies struct {
	PrepareCommand  *application.PrepareGenerateTaskBoardCommand
	GenerateClient  *infraqueue.AsynqGenerateTaskBoardClient
	RedisConnOpt    asynq.RedisConnOpt
	ResultQueueName string
	WaitTimeout     time.Duration
	WorkerErrorCh   <-chan error
}

type GenerateTaskBoardDependenciesProvider func() (*GenerateTaskBoardExecutionDependencies, error)

func NewGenerateTaskBoardRunner(provider GenerateTaskBoardDependenciesProvider) GenerateTaskBoardRunFunc {
	return func(ctx context.Context, input application.GenerateTaskBoardInput) (string, error) {
		if provider == nil {
			return "", fmt.Errorf("generate task board dependencies provider cannot be nil")
		}
		dependencies, err := provider()
		if err != nil {
			return "", err
		}
		if dependencies == nil {
			return "", fmt.Errorf("generate task board dependencies cannot be nil")
		}
		if dependencies.PrepareCommand == nil {
			return "", fmt.Errorf("prepare command cannot be nil")
		}
		if dependencies.GenerateClient == nil {
			return "", fmt.Errorf("generate client cannot be nil")
		}
		if dependencies.RedisConnOpt == nil {
			return "", fmt.Errorf("redis connection options cannot be nil")
		}

		payload, err := dependencies.PrepareCommand.Execute(ctx, input)
		if err != nil {
			return "", err
		}
		taskID, err := dependencies.GenerateClient.EnqueueGenerateTaskBoard(ctx, payload)
		if err != nil {
			return "", err
		}

		resultTaskID := payload.Metadata.JobID + "-result"
		if err := waitForTaskCompletion(ctx, dependencies.RedisConnOpt, dependencies.ResultQueueName, resultTaskID, dependencies.WaitTimeout, dependencies.WorkerErrorCh); err != nil {
			return "", err
		}
		return taskID, nil
	}
}

func waitForTaskCompletion(ctx context.Context, redisConnOpt asynq.RedisConnOpt, queueName string, taskID string, timeout time.Duration, workerErrorCh <-chan error) error {
	if redisConnOpt == nil {
		return fmt.Errorf("redis connection options are required")
	}
	inspector := asynq.NewInspector(redisConnOpt)
	defer inspector.Close()

	deadline := time.Now().Add(timeout)
	for {
		if workerErrorCh != nil {
			select {
			case workerErr := <-workerErrorCh:
				if workerErr != nil {
					return fmt.Errorf("worker runtime failed: %w", workerErr)
				}
			default:
			}
		}
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
		}
		if timeout > 0 && time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for task %q in queue %q", taskID, queueName)
		}

		info, err := inspector.GetTaskInfo(queueName, taskID)
		if err == nil {
			switch info.State {
			case asynq.TaskStateCompleted:
				return nil
			case asynq.TaskStateArchived:
				return fmt.Errorf("task %q archived: %s", taskID, strings.TrimSpace(info.LastErr))
			}
		} else if !errors.Is(err, asynq.ErrTaskNotFound) && !errors.Is(err, asynq.ErrQueueNotFound) {
			return fmt.Errorf("get task info for %q: %w", taskID, err)
		}

		time.Sleep(500 * time.Millisecond)
	}
}
