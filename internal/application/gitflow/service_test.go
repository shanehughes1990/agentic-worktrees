package gitflow

import (
	"context"
	"errors"
	"testing"
)

type fakeDispatcher struct {
	enqueue func(context.Context, WorktreeFlowJob) (string, error)
}

func (dispatcher *fakeDispatcher) EnqueueWorktreeFlow(ctx context.Context, job WorktreeFlowJob) (string, error) {
	if dispatcher.enqueue == nil {
		return "job-1", nil
	}
	return dispatcher.enqueue(ctx, job)
}

func TestServiceStartBuildsTaskBranchAndWorktreePath(t *testing.T) {
	service := NewService(&fakeDispatcher{})

	result, err := service.Start(context.Background(), StartRequest{
		RunID:          "run-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
	})
	if err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}
	if result.TaskBranch != "task/run-1/task-1" {
		t.Fatalf("unexpected task branch: %s", result.TaskBranch)
	}
	if result.Worktree != ".worktree/run-1-task-1" {
		t.Fatalf("unexpected worktree path: %s", result.Worktree)
	}
}

func TestServiceStartValidatesInput(t *testing.T) {
	service := NewService(&fakeDispatcher{})

	_, err := service.Start(context.Background(), StartRequest{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestServiceStartPropagatesDispatcherError(t *testing.T) {
	expectedErr := errors.New("enqueue failed")
	service := NewService(&fakeDispatcher{enqueue: func(_ context.Context, _ WorktreeFlowJob) (string, error) {
		return "", expectedErr
	}})

	_, err := service.Start(context.Background(), StartRequest{
		RunID:          "run-1",
		TaskID:         "task-1",
		RepositoryRoot: ".",
		SourceBranch:   "revamp",
	})
	if err == nil {
		t.Fatalf("expected dispatcher error")
	}
}
