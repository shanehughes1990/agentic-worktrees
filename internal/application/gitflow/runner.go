package gitflow

import (
	"context"
	"fmt"
	"strings"

	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	domaingitflow "github.com/shanehughes1990/agentic-worktrees/internal/domain/gitflow"
	"github.com/sirupsen/logrus"
)

type MergeAttempt struct {
	ConflictFiles []string
	NoChanges     bool
}

type GitPort interface {
	CreateTaskWorktree(ctx context.Context, repositoryRoot string, sourceBranch string, taskBranch string, worktreePath string) error
	SyncTaskBranchWithSource(ctx context.Context, repositoryRoot string, sourceBranch string, taskBranch string, worktreePath string) (MergeAttempt, error)
	MergeTaskBranch(ctx context.Context, repositoryRoot string, sourceBranch string, taskBranch string) (MergeAttempt, error)
	ResolveConflicts(ctx context.Context, repositoryRoot string, conflictFiles []string, copilotAdvice string) error
	ValidateWorktree(ctx context.Context, repositoryRoot string) error
	StageAll(ctx context.Context, repositoryRoot string) error
	Commit(ctx context.Context, repositoryRoot string, message string) error
	CleanupTaskWorktree(ctx context.Context, repositoryRoot string, worktreePath string, taskBranch string) error
	CleanupRunArtifacts(ctx context.Context, repositoryRoot string, runPrefix string) error
}

type ConflictResolveJob struct {
	RunID          string
	TaskID         string
	RepositoryRoot string
	SourceBranch   string
	TaskBranch     string
	WorktreePath   string
	ConflictFiles  []string
	IdempotencyKey string
}

type ConflictDispatcher interface {
	EnqueueConflictResolution(ctx context.Context, job ConflictResolveJob) (string, error)
}

type Runner struct {
	git                GitPort
	conflictDispatcher ConflictDispatcher
	workflowRepository apptaskboard.WorkflowRepository
	logger             *logrus.Logger
}

func NewRunner(git GitPort, conflictDispatcher ConflictDispatcher, workflowRepository apptaskboard.WorkflowRepository, loggers ...*logrus.Logger) *Runner {
	var logger *logrus.Logger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &Runner{git: git, conflictDispatcher: conflictDispatcher, workflowRepository: workflowRepository, logger: logger}
}

func (runner *Runner) RunWorktreeFlow(ctx context.Context, job WorktreeFlowJob) error {
	entry := runner.entry().WithFields(logrus.Fields{"event": "gitflow.runner.worktree_flow", "run_id": strings.TrimSpace(job.RunID), "task_id": strings.TrimSpace(job.TaskID), "source_branch": strings.TrimSpace(job.SourceBranch), "task_branch": strings.TrimSpace(job.TaskBranch), "worktree_path": strings.TrimSpace(job.WorktreePath), "repository_root": strings.TrimSpace(job.RepositoryRoot)})
	session := &domaingitflow.TaskExecutionSession{
		RunID:        strings.TrimSpace(job.RunID),
		TaskID:       strings.TrimSpace(job.TaskID),
		SourceBranch: strings.TrimSpace(job.SourceBranch),
		TaskBranch:   strings.TrimSpace(job.TaskBranch),
		WorktreePath: strings.TrimSpace(job.WorktreePath),
	}
	session.Normalize()
	if err := session.ValidateBasics(); err != nil {
		entry.WithError(err).Error("invalid worktree flow session")
		return WrapTerminal(err)
	}
	if strings.TrimSpace(job.RepositoryRoot) == "" {
		entry.Error("repository_root is required")
		return WrapTerminal(fmt.Errorf("repository_root is required"))
	}
	entry.Info("starting git worktree flow")

	runner.saveWorkflow(ctx, session.RunID, apptaskboard.WorkflowStatusRunning, "git worktree flow running")

	if err := runner.git.CreateTaskWorktree(ctx, job.RepositoryRoot, session.SourceBranch, session.TaskBranch, session.WorktreePath); err != nil {
		entry.WithError(err).Error("create task worktree failed")
		runner.saveWorkflow(ctx, session.RunID, apptaskboard.WorkflowStatusFailed, fmt.Sprintf("create task worktree failed: %v", err))
		return EnsureClassified(fmt.Errorf("create task worktree: %w", err), FailureClassTerminal)
	}

	mergeAttempt, err := runner.git.MergeTaskBranch(ctx, job.RepositoryRoot, session.SourceBranch, session.TaskBranch)
	if err != nil {
		entry.WithError(err).Error("merge task branch failed")
		runner.saveWorkflow(ctx, session.RunID, apptaskboard.WorkflowStatusFailed, fmt.Sprintf("merge task branch failed: %v", err))
		return EnsureClassified(fmt.Errorf("merge task branch: %w", err), FailureClassTerminal)
	}

	if len(mergeAttempt.ConflictFiles) > 0 {
		entry.WithField("conflict_count", len(mergeAttempt.ConflictFiles)).Warn("merge conflicts detected; queuing resolver")
		runner.saveWorkflow(ctx, session.RunID, apptaskboard.WorkflowStatusRunning, "merge conflict detected; conflict resolver task queued")
		if runner.conflictDispatcher == nil {
			entry.Error("conflict dispatcher is required")
			return WrapTerminal(fmt.Errorf("conflict dispatcher is required"))
		}
		_, enqueueErr := runner.conflictDispatcher.EnqueueConflictResolution(ctx, ConflictResolveJob{
			RunID:          session.RunID,
			TaskID:         session.TaskID,
			RepositoryRoot: job.RepositoryRoot,
			SourceBranch:   session.SourceBranch,
			TaskBranch:     session.TaskBranch,
			WorktreePath:   session.WorktreePath,
			ConflictFiles:  mergeAttempt.ConflictFiles,
			IdempotencyKey: fmt.Sprintf("%s:%s:conflict", session.RunID, session.TaskID),
		})
		if enqueueErr != nil {
			entry.WithError(enqueueErr).Error("enqueue conflict resolver failed")
			runner.saveWorkflow(ctx, session.RunID, apptaskboard.WorkflowStatusFailed, fmt.Sprintf("enqueue conflict resolver failed: %v", enqueueErr))
			return EnsureClassified(fmt.Errorf("enqueue conflict resolver: %w", enqueueErr), FailureClassTransient)
		}
		entry.Info("queued conflict resolver task")
		return nil
	}

	if err := runner.git.CleanupTaskWorktree(ctx, job.RepositoryRoot, session.WorktreePath, session.TaskBranch); err != nil {
		entry.WithError(err).Error("cleanup worktree failed")
		runner.saveWorkflow(ctx, session.RunID, apptaskboard.WorkflowStatusFailed, fmt.Sprintf("cleanup worktree failed: %v", err))
		return EnsureClassified(fmt.Errorf("cleanup task worktree: %w", err), FailureClassTerminal)
	}

	runner.saveWorkflow(ctx, session.RunID, apptaskboard.WorkflowStatusCompleted, "git worktree flow completed")
	entry.Info("git worktree flow completed")
	return nil
}

func (runner *Runner) RunConflictResolution(ctx context.Context, job ConflictResolveJob, copilotAdvice string) error {
	entry := runner.entry().WithFields(logrus.Fields{"event": "gitflow.runner.conflict_resolution", "run_id": strings.TrimSpace(job.RunID), "task_id": strings.TrimSpace(job.TaskID), "source_branch": strings.TrimSpace(job.SourceBranch), "task_branch": strings.TrimSpace(job.TaskBranch), "worktree_path": strings.TrimSpace(job.WorktreePath), "repository_root": strings.TrimSpace(job.RepositoryRoot), "conflict_count": len(job.ConflictFiles)})
	session := &domaingitflow.TaskExecutionSession{
		RunID:        strings.TrimSpace(job.RunID),
		TaskID:       strings.TrimSpace(job.TaskID),
		SourceBranch: strings.TrimSpace(job.SourceBranch),
		TaskBranch:   strings.TrimSpace(job.TaskBranch),
		WorktreePath: strings.TrimSpace(job.WorktreePath),
	}
	session.Normalize()
	if err := session.ValidateBasics(); err != nil {
		entry.WithError(err).Error("invalid conflict resolution session")
		return WrapTerminal(err)
	}
	if strings.TrimSpace(job.RepositoryRoot) == "" {
		entry.Error("repository_root is required")
		return WrapTerminal(fmt.Errorf("repository_root is required"))
	}
	if len(job.ConflictFiles) == 0 {
		entry.Error("conflict_files is required")
		return WrapTerminal(fmt.Errorf("conflict_files is required"))
	}
	entry.Info("starting conflict resolution")

	runner.saveWorkflow(ctx, session.RunID, apptaskboard.WorkflowStatusRunning, "conflict resolution running")

	if err := runner.git.ResolveConflicts(ctx, job.RepositoryRoot, job.ConflictFiles, copilotAdvice); err != nil {
		entry.WithError(err).Error("resolve conflicts failed")
		runner.saveWorkflow(ctx, session.RunID, apptaskboard.WorkflowStatusFailed, fmt.Sprintf("resolve conflicts failed: %v", err))
		return EnsureClassified(fmt.Errorf("resolve conflicts: %w", err), FailureClassTerminal)
	}

	if err := runner.git.Commit(ctx, job.RepositoryRoot, fmt.Sprintf("Resolve merge conflicts for task %s", session.TaskID)); err != nil {
		entry.WithError(err).Error("commit conflict resolution failed")
		runner.saveWorkflow(ctx, session.RunID, apptaskboard.WorkflowStatusFailed, fmt.Sprintf("commit conflict resolution failed: %v", err))
		return EnsureClassified(fmt.Errorf("commit conflict resolution: %w", err), FailureClassTerminal)
	}

	if err := runner.git.CleanupTaskWorktree(ctx, job.RepositoryRoot, session.WorktreePath, session.TaskBranch); err != nil {
		entry.WithError(err).Error("cleanup worktree failed")
		runner.saveWorkflow(ctx, session.RunID, apptaskboard.WorkflowStatusFailed, fmt.Sprintf("cleanup worktree failed: %v", err))
		return EnsureClassified(fmt.Errorf("cleanup task worktree: %w", err), FailureClassTerminal)
	}

	runner.saveWorkflow(ctx, session.RunID, apptaskboard.WorkflowStatusCompleted, "conflicts resolved and merge completed")
	entry.Info("conflict resolution completed")
	return nil
}

func (runner *Runner) saveWorkflow(ctx context.Context, runID string, status apptaskboard.WorkflowStatus, message string) {
	if runner.workflowRepository == nil {
		return
	}
	workflow := &apptaskboard.IngestionWorkflow{RunID: strings.TrimSpace(runID), Status: status, Message: strings.TrimSpace(message)}
	workflow.Normalize(runID)
	_ = runner.workflowRepository.SaveWorkflow(ctx, workflow)
}

func (runner *Runner) entry() *logrus.Entry {
	if runner == nil || runner.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(runner.logger)
}
