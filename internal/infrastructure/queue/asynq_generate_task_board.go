package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/shanehughes1990/agentic-worktrees/internal/application"
)

type AsynqClient struct {
	client *asynq.Client
}

func NewAsynqClient(redisAddress string) (*AsynqClient, error) {
	if strings.TrimSpace(redisAddress) == "" {
		return nil, fmt.Errorf("redis address is required")
	}
	return &AsynqClient{client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddress})}, nil
}

func (c *AsynqClient) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

type AsynqGenerateTaskBoardClient struct {
	asynqClient *AsynqClient
	queueName   string
}

func NewAsynqGenerateTaskBoardClient(asynqClient *AsynqClient, queueName string) (*AsynqGenerateTaskBoardClient, error) {
	if asynqClient == nil || asynqClient.client == nil {
		return nil, fmt.Errorf("asynq client is required")
	}
	if strings.TrimSpace(queueName) == "" {
		return nil, fmt.Errorf("queue name is required")
	}
	return &AsynqGenerateTaskBoardClient{asynqClient: asynqClient, queueName: queueName}, nil
}

func (c *AsynqGenerateTaskBoardClient) EnqueueGenerateTaskBoard(ctx context.Context, payload application.GenerateTaskBoardPayload) (string, error) {
	if c == nil || c.asynqClient == nil || c.asynqClient.client == nil {
		return "", fmt.Errorf("asynq client cannot be nil")
	}
	if err := payload.Validate(); err != nil {
		return "", err
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	task := asynq.NewTask(application.AsynqTaskTypeGenerateTaskBoard, payloadBytes)
	info, err := c.asynqClient.client.EnqueueContext(ctx, task, asynq.Queue(c.queueName), asynq.TaskID(payload.Metadata.JobID))
	if err != nil {
		return "", fmt.Errorf("enqueue task: %w", err)
	}
	return info.ID, nil
}

func (c *AsynqGenerateTaskBoardClient) EnqueueGenerateTaskBoardResult(ctx context.Context, result application.GenerateTaskBoardResultMessage) (string, error) {
	if c == nil || c.asynqClient == nil || c.asynqClient.client == nil {
		return "", fmt.Errorf("asynq client cannot be nil")
	}
	if err := result.Validate(); err != nil {
		return "", err
	}

	payloadBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal result: %w", err)
	}

	task := asynq.NewTask(application.AsynqTaskTypeGenerateTaskBoardResult, payloadBytes)
	info, err := c.asynqClient.client.EnqueueContext(ctx, task, asynq.Queue(c.queueName))
	if err != nil {
		return "", fmt.Errorf("enqueue result task: %w", err)
	}
	return info.ID, nil
}
