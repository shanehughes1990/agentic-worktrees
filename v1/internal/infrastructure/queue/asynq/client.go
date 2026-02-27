package asynq

import (
	"context"

	"github.com/hibiken/asynq"
)

type Client struct {
	inner *asynq.Client
}

func NewClient(cfg Config) *Client {
	return &Client{inner: asynq.NewClient(cfg.redisConnOpt)}
}

func (client *Client) Enqueue(ctx context.Context, taskType string, payload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	task := asynq.NewTask(taskType, payload)
	return client.inner.EnqueueContext(ctx, task, opts...)
}

func (client *Client) Close() error {
	return client.inner.Close()
}
