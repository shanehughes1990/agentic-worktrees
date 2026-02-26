package gitflow

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
)

type TaskExecutionRequest struct {
	BoardID         string
	RunID           string
	TaskID          string
	TaskTitle       string
	TaskDetail      string
	ResumeSessionID string
	ExecutionAttempt int
	SourceBranch    string
	RepositoryRoot  string
	WorktreePath    string
	WorktreeRoot    string
}

type TaskExecutionResult struct {
	Status          string
	Reason          string
	TaskBranch      string
	Worktree        string
	ResumeSessionID string
}

type ResumeSessionCheckpoint interface {
	CheckpointResumeSession(ctx context.Context, boardID string, taskID string, resumeSessionID string) error
}

type TaskExecutor struct {
	git               GitPort
	decomposer        appcopilot.Decomposer
	resumeCheckpoint  ResumeSessionCheckpoint
}

func NewTaskExecutor(git GitPort, decomposer appcopilot.Decomposer, resumeCheckpoints ...ResumeSessionCheckpoint) *TaskExecutor {
	var resumeCheckpoint ResumeSessionCheckpoint
	if len(resumeCheckpoints) > 0 {
		resumeCheckpoint = resumeCheckpoints[0]
	}
	return &TaskExecutor{git: git, decomposer: decomposer, resumeCheckpoint: resumeCheckpoint}
}

func (executor *TaskExecutor) ExecuteTask(ctx context.Context, request TaskExecutionRequest) (result TaskExecutionResult, executeErr error) {
	cleanBoardID := strings.TrimSpace(request.BoardID)
	cleanRunID := strings.TrimSpace(request.RunID)
	cleanTaskID := strings.TrimSpace(request.TaskID)
	cleanSourceBranch := strings.TrimSpace(request.SourceBranch)
	cleanRepositoryRoot := strings.TrimSpace(request.RepositoryRoot)

	if cleanBoardID == "" {
		return TaskExecutionResult{}, WrapTerminal(fmt.Errorf("board_id is required"))
	}
	if cleanRunID == "" {
		cleanRunID = cleanBoardID
	}
	if cleanTaskID == "" {
		return TaskExecutionResult{}, WrapTerminal(fmt.Errorf("task_id is required"))
	}
	if cleanSourceBranch == "" {
		return TaskExecutionResult{}, WrapTerminal(fmt.Errorf("source_branch is required"))
	}
	if cleanRepositoryRoot == "" {
		return TaskExecutionResult{}, WrapTerminal(fmt.Errorf("repository_root is required"))
	}

	taskBranch := fmt.Sprintf("task/%s/%s", sanitizeBranchSegment(cleanRunID), sanitizeBranchSegment(cleanTaskID))
	worktreePath := strings.TrimSpace(request.WorktreePath)
	if worktreePath == "" {
		worktreePath = buildWorktreePath(request.WorktreeRoot, cleanRunID, cleanTaskID)
	}

	failedResult := func(status string, reason string, resumeSessionID string) TaskExecutionResult {
		cleanStatus := strings.TrimSpace(status)
		if cleanStatus == "" {
			cleanStatus = "failed"
		}
		cleanReason := strings.TrimSpace(reason)
		if cleanReason == "" {
			cleanReason = "task execution failed"
		}
		return TaskExecutionResult{
			Status:          cleanStatus,
			Reason:          cleanReason,
			TaskBranch:      taskBranch,
			Worktree:        worktreePath,
			ResumeSessionID: strings.TrimSpace(resumeSessionID),
		}
	}

	if err := executor.git.CreateTaskWorktree(ctx, cleanRepositoryRoot, cleanSourceBranch, taskBranch, worktreePath); err != nil {
		return failedResult("failed", err.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("create task worktree: %w", err), FailureClassTerminal)
	}

	absoluteWorktreePath := filepath.Join(cleanRepositoryRoot, filepath.FromSlash(worktreePath))
	if executor.decomposer != nil {
		decomposeResult, err := executor.decomposer.Decompose(ctx, appcopilot.DecomposeRequest{
			RunID:            cleanRunID,
			ResumeSessionID:  strings.TrimSpace(request.ResumeSessionID),
			WorkingDirectory: absoluteWorktreePath,
			Prompt:           buildTaskImplementationPrompt(request),
		})
		if err != nil {
			defaultClass := FailureClassTerminal
			if IsTransientInfrastructureFailure(err) {
				defaultClass = FailureClassTransient
			}
			classifiedErr := EnsureClassified(fmt.Errorf("implement task with agent: %w", err), defaultClass)
			result := TaskExecutionResult{
				Status:          "failed",
				Reason:          err.Error(),
				TaskBranch:      taskBranch,
				Worktree:        worktreePath,
				ResumeSessionID: strings.TrimSpace(decomposeResult.SessionID),
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				result.Status = "canceled"
				result.Reason = "task execution canceled"
				classifiedErr = EnsureClassified(fmt.Errorf("implement task with agent: %w", err), FailureClassTransient)
			}
			return result, classifiedErr
		}
		if strings.TrimSpace(decomposeResult.SessionID) != "" {
			request.ResumeSessionID = strings.TrimSpace(decomposeResult.SessionID)
			if checkpointErr := executor.checkpointResumeSession(ctx, cleanBoardID, cleanTaskID, request.ResumeSessionID); checkpointErr != nil {
				return failedResult("failed", checkpointErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("checkpoint resume session: %w", checkpointErr), FailureClassTransient)
			}
		}
		if stageErr := executor.git.StageAll(ctx, absoluteWorktreePath); stageErr != nil {
			return failedResult("failed", stageErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("stage task worktree changes: %w", stageErr), FailureClassTerminal)
		}
		if commitErr := executor.git.Commit(ctx, absoluteWorktreePath, fmt.Sprintf("Implement task %s", cleanTaskID)); commitErr != nil {
			return failedResult("failed", commitErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("commit task worktree changes: %w", commitErr), FailureClassTerminal)
		}
	}

	syncAttempt, err := executor.git.SyncTaskBranchWithSource(ctx, cleanRepositoryRoot, cleanSourceBranch, taskBranch, worktreePath)
	if err != nil {
		return failedResult("failed", err.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("sync task branch with source: %w", err), FailureClassTerminal)
	}

	if len(syncAttempt.ConflictFiles) > 0 {
		if executor.decomposer == nil {
			return failedResult("failed", "resolve pre-merge sync conflicts: copilot decomposer is required", request.ResumeSessionID), EnsureClassified(fmt.Errorf("resolve pre-merge sync conflicts: copilot decomposer is required"), FailureClassTerminal)
		}
		resolutionResult, resolutionErr := executor.decomposer.Decompose(ctx, appcopilot.DecomposeRequest{
			RunID:            cleanRunID,
			ResumeSessionID:  strings.TrimSpace(request.ResumeSessionID),
			WorkingDirectory: absoluteWorktreePath,
			Prompt:           buildSourceSyncConflictResolutionPrompt(cleanTaskID, cleanSourceBranch, syncAttempt.ConflictFiles),
		})
		if resolutionErr != nil {
			defaultClass := FailureClassTerminal
			if IsTransientInfrastructureFailure(resolutionErr) {
				defaultClass = FailureClassTransient
			}
			resumeSessionID := strings.TrimSpace(request.ResumeSessionID)
			if strings.TrimSpace(resolutionResult.SessionID) != "" {
				resumeSessionID = strings.TrimSpace(resolutionResult.SessionID)
				_ = executor.checkpointResumeSession(ctx, cleanBoardID, cleanTaskID, resumeSessionID)
			}
			return failedResult("failed", resolutionErr.Error(), resumeSessionID), EnsureClassified(fmt.Errorf("resolve pre-merge sync conflicts with copilot: %w", resolutionErr), defaultClass)
		}
		if strings.TrimSpace(resolutionResult.SessionID) != "" {
			request.ResumeSessionID = strings.TrimSpace(resolutionResult.SessionID)
			if checkpointErr := executor.checkpointResumeSession(ctx, cleanBoardID, cleanTaskID, request.ResumeSessionID); checkpointErr != nil {
				return failedResult("failed", checkpointErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("checkpoint resume session: %w", checkpointErr), FailureClassTransient)
			}
		}
		if resolveErr := executor.git.ResolveConflicts(ctx, absoluteWorktreePath, syncAttempt.ConflictFiles, resolutionResult.Response); resolveErr != nil {
			return failedResult("failed", resolveErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("resolve pre-merge sync conflicts: %w", resolveErr), FailureClassTerminal)
		}
		if stageErr := executor.git.StageAll(ctx, absoluteWorktreePath); stageErr != nil {
			return failedResult("failed", stageErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("stage pre-merge sync conflict resolution changes: %w", stageErr), FailureClassTerminal)
		}
		if commitErr := executor.git.Commit(ctx, absoluteWorktreePath, fmt.Sprintf("Resolve pre-merge sync conflicts for task %s", cleanTaskID)); commitErr != nil {
			return failedResult("failed", commitErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("commit pre-merge sync conflict resolution: %w", commitErr), FailureClassTerminal)
		}
		syncAttempt.NoChanges = false
	}

	if !syncAttempt.NoChanges {
		if executor.decomposer != nil {
			recheckResult, recheckErr := executor.decomposer.Decompose(ctx, appcopilot.DecomposeRequest{
				RunID:            cleanRunID,
				ResumeSessionID:  strings.TrimSpace(request.ResumeSessionID),
				WorkingDirectory: absoluteWorktreePath,
				Prompt:           buildPostSourceSyncRecheckPrompt(request, cleanSourceBranch),
			})
			if recheckErr != nil {
				defaultClass := FailureClassTerminal
				if IsTransientInfrastructureFailure(recheckErr) {
					defaultClass = FailureClassTransient
				}
				resumeSessionID := strings.TrimSpace(request.ResumeSessionID)
				if strings.TrimSpace(recheckResult.SessionID) != "" {
					resumeSessionID = strings.TrimSpace(recheckResult.SessionID)
					_ = executor.checkpointResumeSession(ctx, cleanBoardID, cleanTaskID, resumeSessionID)
				}
				return failedResult("failed", recheckErr.Error(), resumeSessionID), EnsureClassified(fmt.Errorf("post-sync recheck with agent: %w", recheckErr), defaultClass)
			}
			if strings.TrimSpace(recheckResult.SessionID) != "" {
				request.ResumeSessionID = strings.TrimSpace(recheckResult.SessionID)
				if checkpointErr := executor.checkpointResumeSession(ctx, cleanBoardID, cleanTaskID, request.ResumeSessionID); checkpointErr != nil {
					return failedResult("failed", checkpointErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("checkpoint resume session: %w", checkpointErr), FailureClassTransient)
				}
			}
			if stageErr := executor.git.StageAll(ctx, absoluteWorktreePath); stageErr != nil {
				return failedResult("failed", stageErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("stage post-sync recheck changes: %w", stageErr), FailureClassTerminal)
			}
			if commitErr := executor.git.Commit(ctx, absoluteWorktreePath, fmt.Sprintf("Recheck task %s after source sync", cleanTaskID)); commitErr != nil {
				return failedResult("failed", commitErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("commit post-sync recheck changes: %w", commitErr), FailureClassTerminal)
			}
		}

		if validateErr := executor.git.ValidateWorktree(ctx, absoluteWorktreePath); validateErr != nil {
			return failedResult("failed", validateErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("validate task worktree after source sync: %w", validateErr), FailureClassTerminal)
		}
	}

	mergeAttempt, err := executor.git.MergeTaskBranch(ctx, cleanRepositoryRoot, cleanSourceBranch, taskBranch)
	if err != nil {
		return failedResult("failed", err.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("merge task branch: %w", err), FailureClassTerminal)
	}

	if len(mergeAttempt.ConflictFiles) > 0 {
		if executor.decomposer == nil {
			return failedResult("failed", "resolve merge conflicts: copilot decomposer is required", request.ResumeSessionID), EnsureClassified(fmt.Errorf("resolve merge conflicts: copilot decomposer is required"), FailureClassTerminal)
		}
		resolutionResult, resolutionErr := executor.decomposer.Decompose(ctx, appcopilot.DecomposeRequest{
			RunID:            cleanRunID,
			ResumeSessionID:  strings.TrimSpace(request.ResumeSessionID),
			WorkingDirectory: cleanRepositoryRoot,
			Prompt:           buildConflictResolutionPrompt(cleanTaskID, mergeAttempt.ConflictFiles),
		})
		if resolutionErr != nil {
			defaultClass := FailureClassTerminal
			if IsTransientInfrastructureFailure(resolutionErr) {
				defaultClass = FailureClassTransient
			}
			resumeSessionID := strings.TrimSpace(request.ResumeSessionID)
			if strings.TrimSpace(resolutionResult.SessionID) != "" {
				resumeSessionID = strings.TrimSpace(resolutionResult.SessionID)
				_ = executor.checkpointResumeSession(ctx, cleanBoardID, cleanTaskID, resumeSessionID)
			}
			return failedResult("failed", resolutionErr.Error(), resumeSessionID), EnsureClassified(fmt.Errorf("resolve merge conflicts with copilot: %w", resolutionErr), defaultClass)
		}
		if strings.TrimSpace(resolutionResult.SessionID) != "" {
			request.ResumeSessionID = strings.TrimSpace(resolutionResult.SessionID)
			if checkpointErr := executor.checkpointResumeSession(ctx, cleanBoardID, cleanTaskID, request.ResumeSessionID); checkpointErr != nil {
				return failedResult("failed", checkpointErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("checkpoint resume session: %w", checkpointErr), FailureClassTransient)
			}
		}
		if resolveErr := executor.git.ResolveConflicts(ctx, cleanRepositoryRoot, mergeAttempt.ConflictFiles, resolutionResult.Response); resolveErr != nil {
			return failedResult("failed", resolveErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("resolve merge conflicts: %w", resolveErr), FailureClassTerminal)
		}
		if stageErr := executor.git.StageAll(ctx, cleanRepositoryRoot); stageErr != nil {
			return failedResult("failed", stageErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("stage conflict resolution changes: %w", stageErr), FailureClassTerminal)
		}
		if commitErr := executor.git.Commit(ctx, cleanRepositoryRoot, fmt.Sprintf("Resolve merge conflicts for task %s", cleanTaskID)); commitErr != nil {
			return failedResult("failed", commitErr.Error(), request.ResumeSessionID), EnsureClassified(fmt.Errorf("commit conflict resolution: %w", commitErr), FailureClassTerminal)
		}
	}

	result = TaskExecutionResult{
		TaskBranch:      taskBranch,
		Worktree:        worktreePath,
		Status:          "merged",
		Reason:          "task merged into source branch and cleaned up",
		ResumeSessionID: strings.TrimSpace(request.ResumeSessionID),
	}
	if mergeAttempt.NoChanges {
		result.Status = "no_changes"
		result.Reason = "task execution produced no mergeable changes"
	}

	if cleanupErr := executor.git.CleanupTaskWorktree(ctx, cleanRepositoryRoot, worktreePath, taskBranch); cleanupErr != nil {
		return result, EnsureClassified(fmt.Errorf("cleanup task worktree: %w", cleanupErr), FailureClassTerminal)
	}
	return result, nil
}

func (executor *TaskExecutor) checkpointResumeSession(ctx context.Context, boardID string, taskID string, resumeSessionID string) error {
	if executor == nil || executor.resumeCheckpoint == nil {
		return nil
	}
	cleanSessionID := strings.TrimSpace(resumeSessionID)
	if cleanSessionID == "" {
		return nil
	}
	return executor.resumeCheckpoint.CheckpointResumeSession(ctx, strings.TrimSpace(boardID), strings.TrimSpace(taskID), cleanSessionID)
}

func (executor *TaskExecutor) CleanupBoardRun(ctx context.Context, boardID string, repositoryRoot string) error {
	cleanBoardID := strings.TrimSpace(boardID)
	cleanRepositoryRoot := strings.TrimSpace(repositoryRoot)
	if cleanBoardID == "" {
		return WrapTerminal(fmt.Errorf("board_id is required"))
	}
	if cleanRepositoryRoot == "" {
		return WrapTerminal(fmt.Errorf("repository_root is required"))
	}
	if executor == nil || executor.git == nil {
		return WrapTerminal(fmt.Errorf("git port is required"))
	}

	runPrefix := sanitizeBranchSegment(cleanBoardID)
	if err := executor.git.CleanupRunArtifacts(ctx, cleanRepositoryRoot, runPrefix); err != nil {
		return EnsureClassified(fmt.Errorf("cleanup run artifacts: %w", err), FailureClassTerminal)
	}
	return nil
}

func buildTaskImplementationPrompt(request TaskExecutionRequest) string {
	if request.ExecutionAttempt > 1 {
		return fmt.Sprintf("Previous attempt made no forward progress for task %s on board %s. Continue only the remaining work, make minimal correct code changes in this worktree, and if blocked return an explicit BLOCKED reason. Task title: %s. Task detail: %s.",
			strings.TrimSpace(request.TaskID),
			strings.TrimSpace(request.BoardID),
			strings.TrimSpace(request.TaskTitle),
			strings.TrimSpace(request.TaskDetail),
		)
	}
	return fmt.Sprintf("Implement task %s for board %s. Task title: %s. Task detail: %s. Apply minimal correct code changes in this worktree and ensure code remains buildable.",
		strings.TrimSpace(request.TaskID),
		strings.TrimSpace(request.BoardID),
		strings.TrimSpace(request.TaskTitle),
		strings.TrimSpace(request.TaskDetail),
	)
}

func buildConflictResolutionPrompt(taskID string, files []string) string {
	return fmt.Sprintf("Resolve merge conflicts for task %s in files: %s. You must fully resolve all conflict markers, preserve intended task changes, ensure the repository builds/tests still pass for impacted scope, and leave the repository in a merge-ready clean state for commit.", strings.TrimSpace(taskID), strings.Join(files, ", "))
}

func buildSourceSyncConflictResolutionPrompt(taskID string, sourceBranch string, files []string) string {
	return fmt.Sprintf("You are synchronizing task %s with the latest %s branch before final merge. Resolve these conflicts in the task worktree files: %s. Preserve task intent while integrating latest upstream changes, ensure conflict markers are fully removed, and keep repository state commit-ready.", strings.TrimSpace(taskID), strings.TrimSpace(sourceBranch), strings.Join(files, ", "))
}

func buildPostSourceSyncRecheckPrompt(request TaskExecutionRequest, sourceBranch string) string {
	return fmt.Sprintf("Recheck task %s for board %s after syncing with latest %s. Confirm the task is still complete and correct, adjust implementation if required, and ensure tests/build expectations remain satisfied. Task title: %s. Task detail: %s.",
		strings.TrimSpace(request.TaskID),
		strings.TrimSpace(request.BoardID),
		strings.TrimSpace(sourceBranch),
		strings.TrimSpace(request.TaskTitle),
		strings.TrimSpace(request.TaskDetail),
	)
}
