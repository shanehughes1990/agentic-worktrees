package gitflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
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
}

func NewService(dispatcher Dispatcher) *Service {
	return &Service{dispatcher: dispatcher}
}

func (service *Service) Start(ctx context.Context, request StartRequest) (StartResult, error) {
	if service.dispatcher == nil {
		return StartResult{}, fmt.Errorf("dispatcher is required")
	}

	runID := strings.TrimSpace(request.RunID)
	taskID := strings.TrimSpace(request.TaskID)
	repositoryRoot := strings.TrimSpace(request.RepositoryRoot)
	sourceBranch := strings.TrimSpace(request.SourceBranch)

	if runID == "" {
		runID = uuid.NewString()
	}
	if taskID == "" {
		return StartResult{}, fmt.Errorf("task_id is required")
	}
	if repositoryRoot == "" {
		return StartResult{}, fmt.Errorf("repository_root is required")
	}
	if sourceBranch == "" {
		return StartResult{}, fmt.Errorf("source_branch is required")
	}

	taskBranch := fmt.Sprintf("task/%s/%s", sanitizeBranchSegment(runID), sanitizeBranchSegment(taskID))
	worktreePath := fmt.Sprintf(".worktree/%s-%s", sanitizeWorktreeSegment(runID), sanitizeWorktreeSegment(taskID))

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
		return StartResult{}, fmt.Errorf("enqueue worktree flow: %w", err)
	}

	return StartResult{QueueTaskID: queueTaskID, TaskBranch: taskBranch, Worktree: worktreePath}, nil
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
