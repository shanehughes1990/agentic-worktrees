package gitflow

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type StartRequest struct {
	RunID          string
	BoardID        string
	TaskID         string
	TaskTitle      string
	TaskDetail     string
	ResumeSessionID string
	RepositoryRoot string
	SourceBranch   string
	WorktreeRoot   string
}

type StartResult struct {
	QueueTaskID string
	TaskBranch  string
	Worktree    string
}

type WorktreeFlowJob struct {
	RunID          string
	BoardID        string
	TaskID         string
	TaskTitle      string
	TaskDetail     string
	ResumeSessionID string
	RepositoryRoot string
	SourceBranch   string
	TaskBranch     string
	WorktreePath   string
}

type Dispatcher interface {
	EnqueueWorktreeFlow(ctx context.Context, job WorktreeFlowJob) (string, error)
}

type Service struct {
	dispatcher Dispatcher
	logger     *logrus.Logger
}

func NewService(dispatcher Dispatcher, loggers ...*logrus.Logger) *Service {
	var logger *logrus.Logger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &Service{dispatcher: dispatcher, logger: logger}
}

func (service *Service) Start(ctx context.Context, request StartRequest) (StartResult, error) {
	entry := service.entry().WithFields(logrus.Fields{
		"event":           "gitflow.start",
		"run_id":          strings.TrimSpace(request.RunID),
		"board_id":        strings.TrimSpace(request.BoardID),
		"task_id":         strings.TrimSpace(request.TaskID),
		"source_branch":   strings.TrimSpace(request.SourceBranch),
		"repository_root": strings.TrimSpace(request.RepositoryRoot),
	})
	if service.dispatcher == nil {
		entry.Error("dispatcher is required")
		return StartResult{}, fmt.Errorf("dispatcher is required")
	}

	runID := strings.TrimSpace(request.RunID)
	taskID := strings.TrimSpace(request.TaskID)
	repositoryRoot := strings.TrimSpace(request.RepositoryRoot)
	sourceBranch := strings.TrimSpace(request.SourceBranch)

	if runID == "" {
		runID = uuid.NewString()
		entry = entry.WithField("run_id", runID)
	}
	if taskID == "" {
		entry.Error("task_id is required")
		return StartResult{}, fmt.Errorf("task_id is required")
	}
	if repositoryRoot == "" {
		entry.Error("repository_root is required")
		return StartResult{}, fmt.Errorf("repository_root is required")
	}
	if sourceBranch == "" {
		entry.Error("source_branch is required")
		return StartResult{}, fmt.Errorf("source_branch is required")
	}

	taskBranch := fmt.Sprintf("task/%s/%s", sanitizeBranchSegment(runID), sanitizeBranchSegment(taskID))
	worktreePath := buildWorktreePath(request.WorktreeRoot, runID, taskID)

	queueTaskID, err := service.dispatcher.EnqueueWorktreeFlow(ctx, WorktreeFlowJob{
		RunID:          runID,
		BoardID:        strings.TrimSpace(request.BoardID),
		TaskID:         taskID,
		TaskTitle:      strings.TrimSpace(request.TaskTitle),
		TaskDetail:     strings.TrimSpace(request.TaskDetail),
		ResumeSessionID: strings.TrimSpace(request.ResumeSessionID),
		RepositoryRoot: repositoryRoot,
		SourceBranch:   sourceBranch,
		TaskBranch:     taskBranch,
		WorktreePath:   worktreePath,
	})
	if err != nil {
		entry.WithError(err).WithFields(logrus.Fields{"task_branch": taskBranch, "worktree_path": worktreePath}).Error("failed to enqueue git worktree flow")
		return StartResult{}, fmt.Errorf("enqueue worktree flow: %w", err)
	}

	entry.WithFields(logrus.Fields{"queue_task_id": queueTaskID, "task_branch": taskBranch, "worktree_path": worktreePath}).Info("enqueued git worktree flow")

	return StartResult{QueueTaskID: queueTaskID, TaskBranch: taskBranch, Worktree: worktreePath}, nil
}

func (service *Service) entry() *logrus.Entry {
	if service == nil || service.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(service.logger)
}

func buildWorktreePath(worktreeRoot, runID, taskID string) string {
	cleanRoot := filepath.ToSlash(filepath.Clean(strings.TrimSpace(worktreeRoot)))
	if cleanRoot == "" || cleanRoot == "." {
		cleanRoot = ".worktree"
	}
	return filepath.ToSlash(filepath.Join(cleanRoot, "worktrees", fmt.Sprintf("%s-%s", sanitizeWorktreeSegment(runID), sanitizeWorktreeSegment(taskID))))
}

func sanitizeBranchSegment(value string) string {
	cleanValue := strings.TrimSpace(value)
	cleanValue = strings.ReplaceAll(cleanValue, " ", "-")
	cleanValue = strings.ReplaceAll(cleanValue, "..", "-")
	cleanValue = strings.ReplaceAll(cleanValue, "~", "-")
	cleanValue = strings.ReplaceAll(cleanValue, "^", "-")
	cleanValue = strings.ReplaceAll(cleanValue, ":", "-")
	cleanValue = strings.ReplaceAll(cleanValue, "?", "-")
	cleanValue = strings.ReplaceAll(cleanValue, "*", "-")
	cleanValue = strings.ReplaceAll(cleanValue, "[", "-")
	cleanValue = strings.ReplaceAll(cleanValue, "\\", "-")
	cleanValue = strings.ReplaceAll(cleanValue, "@{", "-")
	cleanValue = strings.ReplaceAll(cleanValue, "//", "/")
	cleanValue = strings.Trim(cleanValue, "/.-")
	if cleanValue == "" {
		return "value"
	}
	return cleanValue
}

func sanitizeWorktreeSegment(value string) string {
	cleanValue := strings.TrimSpace(value)
	cleanValue = strings.ReplaceAll(cleanValue, " ", "-")
	cleanValue = strings.ReplaceAll(cleanValue, "/", "-")
	cleanValue = strings.ReplaceAll(cleanValue, "\\", "-")
	cleanValue = strings.ReplaceAll(cleanValue, "..", "-")
	cleanValue = strings.Trim(cleanValue, "-.")
	if cleanValue == "" {
		return "value"
	}
	return cleanValue
}
