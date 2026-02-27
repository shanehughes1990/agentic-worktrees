package asynq

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
)

func (client *Client) EnqueueCopilotDecompose(ctx context.Context, payload tasks.CopilotDecomposePayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	task, taskOpts, err := tasks.NewCopilotDecomposeTask(payload, opts...)
	if err != nil {
		return nil, err
	}
	return client.inner.EnqueueContext(ctx, task, taskOpts...)
}

func (client *Client) EnqueueGitWorktreeFlow(ctx context.Context, payload tasks.GitWorktreeFlowPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	task, taskOpts, err := tasks.NewGitWorktreeFlowTask(payload, opts...)
	if err != nil {
		return nil, err
	}
	return client.inner.EnqueueContext(ctx, task, taskOpts...)
}

func (client *Client) EnqueueGitConflictResolve(ctx context.Context, payload tasks.GitConflictResolvePayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	task, taskOpts, err := tasks.NewGitConflictResolveTask(payload, opts...)
	if err != nil {
		return nil, err
	}
	return client.inner.EnqueueContext(ctx, task, taskOpts...)
}

func (client *Client) EnqueueTaskboardExecute(ctx context.Context, payload tasks.TaskboardExecutePayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	task, taskOpts, err := tasks.NewTaskboardExecuteTask(payload, opts...)
	if err != nil {
		return nil, err
	}
	return client.inner.EnqueueContext(ctx, task, taskOpts...)
}
