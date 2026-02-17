package producer

import (
	"fmt"
	"strings"

	"github.com/hibiken/asynq"

	queuedomain "github.com/shanehughes1990/agentic-worktrees/internal/features/queue/domain"
)

type Client struct {
	client    *asynq.Client
	queueName string
}

func NewClient(redisAddr string, queueName string) (*Client, error) {
	if strings.TrimSpace(redisAddr) == "" {
		return nil, fmt.Errorf("redis address cannot be empty")
	}
	return &Client{
		client:    asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
		queueName: strings.TrimSpace(queueName),
	}, nil
}

func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

func (c *Client) EnqueuePlanBoard(payload queuedomain.PlanBoardPayload) (*asynq.TaskInfo, error) {
	task, err := queuedomain.NewPlanBoardTask(payload, c.queueName)
	if err != nil {
		return nil, err
	}
	return c.client.Enqueue(task)
}
