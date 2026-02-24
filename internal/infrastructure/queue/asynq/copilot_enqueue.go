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
