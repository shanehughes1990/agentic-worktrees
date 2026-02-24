package gitflow

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
)

type TaskExecutionRequest struct {
	BoardID        string
	RunID          string
	TaskID         string
	TaskTitle      string
	TaskDetail     string
	SourceBranch   string
	RepositoryRoot string
}

type TaskExecutor struct {
	git        GitPort
	decomposer appcopilot.Decomposer
}

func NewTaskExecutor(git GitPort, decomposer appcopilot.Decomposer) *TaskExecutor {
	return &TaskExecutor{git: git, decomposer: decomposer}
}

func (executor *TaskExecutor) ExecuteTask(ctx context.Context, request TaskExecutionRequest) (executeErr error) {
	cleanBoardID := strings.TrimSpace(request.BoardID)
	cleanRunID := strings.TrimSpace(request.RunID)
	cleanTaskID := strings.TrimSpace(request.TaskID)
	cleanSourceBranch := strings.TrimSpace(request.SourceBranch)
	cleanRepositoryRoot := strings.TrimSpace(request.RepositoryRoot)

	if cleanBoardID == "" {
		return WrapTerminal(fmt.Errorf("board_id is required"))
	}
	if cleanRunID == "" {
		cleanRunID = cleanBoardID
	}
	if cleanTaskID == "" {
		return WrapTerminal(fmt.Errorf("task_id is required"))
	}
	if cleanSourceBranch == "" {
		return WrapTerminal(fmt.Errorf("source_branch is required"))
	}
	if cleanRepositoryRoot == "" {
		return WrapTerminal(fmt.Errorf("repository_root is required"))
	}

	taskBranch := fmt.Sprintf("task/%s/%s", sanitizeBranchSegment(cleanRunID), sanitizeBranchSegment(cleanTaskID))
	worktreePath := fmt.Sprintf(".worktree/%s-%s", sanitizeWorktreeSegment(cleanRunID), sanitizeWorktreeSegment(cleanTaskID))
	worktreeCreated := false

	defer func() {
		if !worktreeCreated {
			return
		}
		cleanupErr := executor.git.CleanupTaskWorktree(ctx, cleanRepositoryRoot, worktreePath, taskBranch)
		if cleanupErr == nil {
			return
		}
		wrappedCleanupErr := EnsureClassified(fmt.Errorf("cleanup task worktree: %w", cleanupErr), FailureClassTerminal)
		if executeErr == nil {
			executeErr = wrappedCleanupErr
		}
	}()

	if err := executor.git.CreateTaskWorktree(ctx, cleanRepositoryRoot, cleanSourceBranch, taskBranch, worktreePath); err != nil {
		return EnsureClassified(fmt.Errorf("create task worktree: %w", err), FailureClassTerminal)
	}
	worktreeCreated = true

	absoluteWorktreePath := filepath.Join(cleanRepositoryRoot, filepath.FromSlash(worktreePath))
	if executor.decomposer != nil {
		_, err := executor.decomposer.Decompose(ctx, appcopilot.DecomposeRequest{
			RunID:            cleanRunID,
			WorkingDirectory: absoluteWorktreePath,
			Prompt:           buildTaskImplementationPrompt(request),
		})
		if err != nil {
			return EnsureClassified(fmt.Errorf("implement task with agent: %w", err), FailureClassTerminal)
		}
	}

	mergeAttempt, err := executor.git.MergeTaskBranch(ctx, cleanRepositoryRoot, cleanSourceBranch, taskBranch)
	if err != nil {
		return EnsureClassified(fmt.Errorf("merge task branch: %w", err), FailureClassTerminal)
	}

	if len(mergeAttempt.ConflictFiles) > 0 {
		advice := ""
		if executor.decomposer != nil {
			result, adviceErr := executor.decomposer.Decompose(ctx, appcopilot.DecomposeRequest{
				RunID:            cleanRunID,
				WorkingDirectory: cleanRepositoryRoot,
				Prompt:           buildConflictResolutionPrompt(cleanTaskID, mergeAttempt.ConflictFiles),
			})
			if adviceErr == nil {
				advice = result.Response
			}
		}
		if resolveErr := executor.git.ResolveConflicts(ctx, cleanRepositoryRoot, mergeAttempt.ConflictFiles, advice); resolveErr != nil {
			return EnsureClassified(fmt.Errorf("resolve merge conflicts: %w", resolveErr), FailureClassTerminal)
		}
		if commitErr := executor.git.Commit(ctx, cleanRepositoryRoot, fmt.Sprintf("Resolve merge conflicts for task %s", cleanTaskID)); commitErr != nil {
			return EnsureClassified(fmt.Errorf("commit conflict resolution: %w", commitErr), FailureClassTerminal)
		}
	}

	return nil
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
	return fmt.Sprintf("Implement task %s for board %s. Task title: %s. Task detail: %s. Apply minimal correct code changes in this worktree and ensure code remains buildable.",
		strings.TrimSpace(request.TaskID),
		strings.TrimSpace(request.BoardID),
		strings.TrimSpace(request.TaskTitle),
		strings.TrimSpace(request.TaskDetail),
	)
}

func buildConflictResolutionPrompt(taskID string, files []string) string {
	return fmt.Sprintf("Resolve merge conflicts for task %s in files: %s. Keep only task-scoped changes and preserve correctness.", strings.TrimSpace(taskID), strings.Join(files, ", "))
}
