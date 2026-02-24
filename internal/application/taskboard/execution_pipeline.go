package taskboard

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type TaskPipelineExecutor interface {
	ExecuteTask(ctx context.Context, request TaskExecutionRequest) error
}

type TaskExecutionRequest struct {
	BoardID        string
	RunID          string
	TaskID         string
	TaskTitle      string
	TaskDetail     string
	SourceBranch   string
	RepositoryRoot string
}

type ExecutionPipelineService struct {
	taskboardService *Service
	executor         TaskPipelineExecutor
	maxAgents        int
}

func NewExecutionPipelineService(taskboardService *Service, executor TaskPipelineExecutor, maxAgents int) *ExecutionPipelineService {
	if maxAgents < 1 {
		maxAgents = 1
	}
	return &ExecutionPipelineService{taskboardService: taskboardService, executor: executor, maxAgents: maxAgents}
}

func (service *ExecutionPipelineService) ExecuteBoard(ctx context.Context, boardID string, sourceBranch string, repositoryRoot string) error {
	cleanBoardID := strings.TrimSpace(boardID)
	cleanSourceBranch := strings.TrimSpace(sourceBranch)
	cleanRepositoryRoot := strings.TrimSpace(repositoryRoot)

	if service.taskboardService == nil {
		return fmt.Errorf("taskboard service is required")
	}
	if service.executor == nil {
		return fmt.Errorf("task executor is required")
	}
	if cleanBoardID == "" {
		return fmt.Errorf("board_id is required")
	}
	if cleanSourceBranch == "" {
		return fmt.Errorf("source_branch is required")
	}
	if cleanRepositoryRoot == "" {
		return fmt.Errorf("repository_root is required")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		readyTasks, err := service.taskboardService.GetReadyTasks(ctx, cleanBoardID)
		if err != nil {
			return fmt.Errorf("load ready tasks: %w", err)
		}
		if len(readyTasks) == 0 {
			completed, completedErr := service.taskboardService.IsBoardCompleted(ctx, cleanBoardID)
			if completedErr != nil {
				return fmt.Errorf("check board completion: %w", completedErr)
			}
			if completed {
				return nil
			}
			return fmt.Errorf("no ready tasks remain for board %s, but board is not completed", cleanBoardID)
		}

		batchSize := service.maxAgents
		if batchSize > len(readyTasks) {
			batchSize = len(readyTasks)
		}
		batch := readyTasks[:batchSize]

		type taskResult struct {
			taskID string
			err    error
		}

		results := make(chan taskResult, len(batch))
		var group sync.WaitGroup
		startedTaskIDs := make([]string, 0, len(batch))

		for _, task := range batch {
			if task == nil {
				continue
			}
			if err := service.taskboardService.MarkTaskInProgress(ctx, cleanBoardID, task.ID); err != nil {
				return fmt.Errorf("mark task in-progress: %w", err)
			}
			startedTaskIDs = append(startedTaskIDs, task.ID)

			request := TaskExecutionRequest{
				BoardID:        cleanBoardID,
				RunID:          cleanBoardID,
				TaskID:         task.ID,
				TaskTitle:      task.Title,
				TaskDetail:     task.Description,
				SourceBranch:   cleanSourceBranch,
				RepositoryRoot: cleanRepositoryRoot,
			}

			group.Add(1)
			go func(taskID string, taskRequest TaskExecutionRequest) {
				defer group.Done()
				results <- taskResult{taskID: taskID, err: service.executor.ExecuteTask(ctx, taskRequest)}
			}(task.ID, request)
		}

		group.Wait()
		close(results)

		if len(startedTaskIDs) == 0 {
			return fmt.Errorf("no tasks were started from ready batch")
		}

		failedTaskErrors := map[string]error{}
		for result := range results {
			if result.err == nil {
				continue
			}
			failedTaskErrors[result.taskID] = result.err
		}

		sort.Strings(startedTaskIDs)
		for _, taskID := range startedTaskIDs {
			executionErr := failedTaskErrors[taskID]
			if executionErr != nil {
				if markErr := service.taskboardService.MarkTaskBlocked(ctx, cleanBoardID, taskID); markErr != nil {
					return fmt.Errorf("mark task blocked after failure: %w", markErr)
				}
				continue
			}
			if err := service.taskboardService.MarkTaskCompleted(ctx, cleanBoardID, taskID); err != nil {
				return fmt.Errorf("mark task completed: %w", err)
			}
		}

		if len(failedTaskErrors) > 0 {
			failedTaskIDs := make([]string, 0, len(failedTaskErrors))
			for taskID := range failedTaskErrors {
				failedTaskIDs = append(failedTaskIDs, taskID)
			}
			sort.Strings(failedTaskIDs)
			return fmt.Errorf("task execution failed for %d tasks: %s", len(failedTaskIDs), strings.Join(failedTaskIDs, ", "))
		}
	}
}
