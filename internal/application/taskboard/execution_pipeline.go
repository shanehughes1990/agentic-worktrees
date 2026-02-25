package taskboard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

type TaskPipelineExecutor interface {
	ExecuteTask(ctx context.Context, request TaskExecutionRequest) (TaskExecutionOutcome, error)
}

type TaskExecutionRequest struct {
	BoardID         string
	RunID           string
	TaskID          string
	TaskTitle       string
	TaskDetail      string
	ResumeSessionID string
	SourceBranch    string
	RepositoryRoot  string
}

type TaskExecutionOutcome struct {
	Status          string
	Reason          string
	TaskBranch      string
	Worktree        string
	ResumeSessionID string
}

type ExecutionPipelineService struct {
	taskboardService *Service
	executor         TaskPipelineExecutor
	workflowRepo     WorkflowRepository
	maxAgents        int
}

func NewExecutionPipelineService(taskboardService *Service, executor TaskPipelineExecutor, workflowRepo WorkflowRepository, maxAgents int) *ExecutionPipelineService {
	if maxAgents < 1 {
		maxAgents = 1
	}
	return &ExecutionPipelineService{taskboardService: taskboardService, executor: executor, workflowRepo: workflowRepo, maxAgents: maxAgents}
}

func (service *ExecutionPipelineService) ExecuteBoard(ctx context.Context, boardID string, sourceBranch string, repositoryRoot string, maxTasks int) error {
	cleanBoardID := strings.TrimSpace(boardID)
	cleanSourceBranch := strings.TrimSpace(sourceBranch)
	cleanRepositoryRoot := strings.TrimSpace(repositoryRoot)
	cleanMaxTasks := maxTasks

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
	if cleanMaxTasks < 0 {
		return fmt.Errorf("max_tasks cannot be negative")
	}

	_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusRunning, "taskboard execution starting", map[string]any{
		"event":           "pipeline_start",
		"max_agents":      service.maxAgents,
		"max_tasks":       cleanMaxTasks,
		"source_branch":   cleanSourceBranch,
		"repository_root": cleanRepositoryRoot,
		"resumed_at":      time.Now().UTC().Format(time.RFC3339),
	})

	requeuedCount, requeueErr := service.taskboardService.RequeueInProgressTasks(ctx, cleanBoardID, "runner interrupted previously; re-queued on resume")
	if requeueErr != nil {
		return fmt.Errorf("requeue interrupted in-progress tasks: %w", requeueErr)
	}
	if requeuedCount > 0 {
		_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusRunning, "requeued interrupted tasks", map[string]any{
			"event":          "resume_requeue",
			"requeued_tasks": requeuedCount,
		})
	}

	if err := service.taskboardService.AnnotateCompletedTasksWithoutOutcome(ctx, cleanBoardID, "task was already completed before this pipeline run"); err != nil {
		return fmt.Errorf("annotate pre-completed tasks: %w", err)
	}

	round := 0
	tasksExecuted := 0
	for {
		select {
		case <-ctx.Done():
			_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusFailed, "taskboard execution canceled", map[string]any{
				"event": "pipeline_canceled",
			})
			return ctx.Err()
		default:
		}
		round++

		readyTasks, err := service.taskboardService.GetReadyTasks(ctx, cleanBoardID)
		if err != nil {
			return fmt.Errorf("load ready tasks: %w", err)
		}
		readyTaskIDs := make([]string, 0, len(readyTasks))
		for _, readyTask := range readyTasks {
			if readyTask == nil {
				continue
			}
			readyTaskIDs = append(readyTaskIDs, readyTask.ID)
		}
		_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusRunning, "ready task scan complete", map[string]any{
			"event":       "round_ready",
			"round":       round,
			"ready_tasks": readyTaskIDs,
		})

		if len(readyTasks) == 0 {
			completed, completedErr := service.taskboardService.IsBoardCompleted(ctx, cleanBoardID)
			if completedErr != nil {
				return fmt.Errorf("check board completion: %w", completedErr)
			}
			if completed {
				_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusCompleted, "taskboard execution completed", map[string]any{
					"event": "pipeline_completed",
				})
				return nil
			}
			_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusFailed, "pipeline blocked by unresolved dependencies", map[string]any{
				"event": "pipeline_blocked",
				"round": round,
			})
			return fmt.Errorf("no ready tasks remain for board %s, but board is not completed", cleanBoardID)
		}

		batchSize := service.maxAgents
		if batchSize > len(readyTasks) {
			batchSize = len(readyTasks)
		}
		if cleanMaxTasks > 0 {
			remaining := cleanMaxTasks - tasksExecuted
			if remaining <= 0 {
				_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusCompleted, "taskboard execution stopped at task limit", map[string]any{
					"event":          "pipeline_limit_reached",
					"max_tasks":      cleanMaxTasks,
					"executed_tasks": tasksExecuted,
				})
				return nil
			}
			if batchSize > remaining {
				batchSize = remaining
			}
		}
		batch := readyTasks[:batchSize]
		batchTaskIDs := make([]string, 0, len(batch))
		for _, task := range batch {
			if task == nil {
				continue
			}
			batchTaskIDs = append(batchTaskIDs, task.ID)
		}
		_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusRunning, "starting task batch", map[string]any{
			"event":       "batch_start",
			"round":       round,
			"batch_tasks": batchTaskIDs,
		})

		type taskResult struct {
			taskID  string
			outcome TaskExecutionOutcome
			err     error
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
				BoardID:    cleanBoardID,
				RunID:      cleanBoardID,
				TaskID:     task.ID,
				TaskTitle:  task.Title,
				TaskDetail: task.Description,
				ResumeSessionID: func() string {
					if task.Outcome == nil {
						return ""
					}
					return strings.TrimSpace(task.Outcome.ResumeSessionID)
				}(),
				SourceBranch:   cleanSourceBranch,
				RepositoryRoot: cleanRepositoryRoot,
			}

			group.Add(1)
			go func(taskID string, taskRequest TaskExecutionRequest) {
				defer group.Done()
				outcome, err := service.executor.ExecuteTask(ctx, taskRequest)
				results <- taskResult{taskID: taskID, outcome: outcome, err: err}
			}(task.ID, request)
		}

		group.Wait()
		close(results)

		if len(startedTaskIDs) == 0 {
			_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusFailed, "no tasks started in batch", map[string]any{
				"event": "batch_empty",
				"round": round,
			})
			return fmt.Errorf("no tasks were started from ready batch")
		}

		failedTaskErrors := map[string]error{}
		outcomesByTaskID := map[string]TaskExecutionOutcome{}
		canceledTaskIDs := make([]string, 0)
		for result := range results {
			outcomesByTaskID[result.taskID] = result.outcome
			if result.err == nil {
				continue
			}
			if errors.Is(result.err, context.Canceled) || errors.Is(result.err, context.DeadlineExceeded) {
				canceledTaskIDs = append(canceledTaskIDs, result.taskID)
				continue
			}
			failedTaskErrors[result.taskID] = result.err
		}

		sort.Strings(startedTaskIDs)
		for _, taskID := range startedTaskIDs {
			executionErr := failedTaskErrors[taskID]
			outcome := outcomesByTaskID[taskID]
			isCanceled := false
			for _, canceledTaskID := range canceledTaskIDs {
				if canceledTaskID == taskID {
					isCanceled = true
					break
				}
			}
			if isCanceled {
				if strings.TrimSpace(outcome.Status) == "" {
					outcome.Status = "canceled"
				}
				if strings.TrimSpace(outcome.Reason) == "" {
					outcome.Reason = "task execution canceled"
				}
				if markErr := service.taskboardService.MarkTaskCanceledWithOutcome(ctx, cleanBoardID, taskID, domaintaskboard.TaskOutcome{
					Status:          outcome.Status,
					Reason:          outcome.Reason,
					TaskBranch:      outcome.TaskBranch,
					Worktree:        outcome.Worktree,
					ResumeSessionID: outcome.ResumeSessionID,
				}); markErr != nil {
					return fmt.Errorf("mark task canceled for resume: %w", markErr)
				}
				_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusRunning, "task canceled", map[string]any{
					"event":             "task_canceled",
					"task_id":           taskID,
					"reason":            outcome.Reason,
					"resume_session_id": outcome.ResumeSessionID,
				})
				continue
			}
			if executionErr != nil {
				if strings.TrimSpace(outcome.Status) == "" {
					outcome.Status = "failed"
				}
				if strings.TrimSpace(outcome.Reason) == "" {
					outcome.Reason = executionErr.Error()
				}
				if markErr := service.taskboardService.MarkTaskBlockedWithOutcome(ctx, cleanBoardID, taskID, domaintaskboard.TaskOutcome{
					Status:          outcome.Status,
					Reason:          outcome.Reason,
					TaskBranch:      outcome.TaskBranch,
					Worktree:        outcome.Worktree,
					ResumeSessionID: outcome.ResumeSessionID,
				}); markErr != nil {
					return fmt.Errorf("mark task blocked after failure: %w", markErr)
				}
				_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusRunning, "task blocked", map[string]any{
					"event":       "task_blocked",
					"task_id":     taskID,
					"reason":      outcome.Reason,
					"task_branch": outcome.TaskBranch,
					"worktree":    outcome.Worktree,
				})
				continue
			}
			if strings.TrimSpace(outcome.Status) == "" {
				outcome.Status = "merged"
			}
			if strings.TrimSpace(outcome.Reason) == "" {
				outcome.Reason = "task execution completed"
			}
			if err := service.taskboardService.MarkTaskCompletedWithOutcome(ctx, cleanBoardID, taskID, domaintaskboard.TaskOutcome{
				Status:          outcome.Status,
				Reason:          outcome.Reason,
				TaskBranch:      outcome.TaskBranch,
				Worktree:        outcome.Worktree,
				ResumeSessionID: outcome.ResumeSessionID,
			}); err != nil {
				return fmt.Errorf("mark task completed: %w", err)
			}
			_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusRunning, "task completed", map[string]any{
				"event":       "task_completed",
				"task_id":     taskID,
				"status":      outcome.Status,
				"reason":      outcome.Reason,
				"task_branch": outcome.TaskBranch,
				"worktree":    outcome.Worktree,
			})
		}

		if len(canceledTaskIDs) > 0 {
			sort.Strings(canceledTaskIDs)
			_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusFailed, "taskboard execution canceled", map[string]any{
				"event":             "pipeline_canceled",
				"round":             round,
				"canceled_task_ids": canceledTaskIDs,
			})
			return context.Canceled
		}

		if len(failedTaskErrors) > 0 {
			failedTaskIDs := make([]string, 0, len(failedTaskErrors))
			for taskID := range failedTaskErrors {
				failedTaskIDs = append(failedTaskIDs, taskID)
			}
			sort.Strings(failedTaskIDs)
			_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusFailed, "batch failed", map[string]any{
				"event":           "batch_failed",
				"round":           round,
				"failed_task_ids": failedTaskIDs,
			})
			return fmt.Errorf("task execution failed for %d tasks: %s", len(failedTaskIDs), strings.Join(failedTaskIDs, ", "))
		}

		_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusRunning, "batch completed", map[string]any{
			"event":         "batch_completed",
			"round":         round,
			"completed_ids": startedTaskIDs,
		})

		tasksExecuted += len(startedTaskIDs)
		if cleanMaxTasks > 0 && tasksExecuted >= cleanMaxTasks {
			_ = service.appendWorkflowEvent(ctx, cleanBoardID, cleanBoardID, WorkflowStatusCompleted, "taskboard execution stopped at task limit", map[string]any{
				"event":          "pipeline_limit_reached",
				"max_tasks":      cleanMaxTasks,
				"executed_tasks": tasksExecuted,
			})
			return nil
		}
	}
}

func (service *ExecutionPipelineService) appendWorkflowEvent(ctx context.Context, runID string, boardID string, status WorkflowStatus, message string, details map[string]any) error {
	if service.workflowRepo == nil {
		return nil
	}

	cleanRunID := strings.TrimSpace(runID)
	if cleanRunID == "" {
		return nil
	}

	workflow, err := service.workflowRepo.GetWorkflow(ctx, cleanRunID)
	if err != nil {
		return fmt.Errorf("load workflow stream: %w", err)
	}
	if workflow == nil {
		workflow = &IngestionWorkflow{RunID: cleanRunID}
	}

	eventPayload := map[string]any{
		"time":     time.Now().UTC().Format(time.RFC3339Nano),
		"run_id":   cleanRunID,
		"board_id": strings.TrimSpace(boardID),
		"message":  strings.TrimSpace(message),
	}
	for key, value := range details {
		eventPayload[key] = value
	}
	eventJSON, marshalErr := json.Marshal(eventPayload)
	if marshalErr != nil {
		return fmt.Errorf("marshal workflow event: %w", marshalErr)
	}

	existingStream := strings.TrimSpace(workflow.Stream)
	if existingStream == "" {
		workflow.Stream = string(eventJSON)
	} else {
		workflow.Stream = existingStream + "\n" + string(eventJSON)
	}

	workflow.Status = status
	workflow.Message = strings.TrimSpace(message)
	workflow.BoardID = strings.TrimSpace(boardID)
	workflow.Normalize(cleanRunID)

	if saveErr := service.workflowRepo.SaveWorkflow(ctx, workflow); saveErr != nil {
		return fmt.Errorf("save workflow stream: %w", saveErr)
	}
	return nil
}
