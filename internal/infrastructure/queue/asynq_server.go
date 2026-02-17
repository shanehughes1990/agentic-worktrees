package queue

import (
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
)

func NewAsynqServer(redisAddress string, queueName string, concurrency int) (*asynq.Server, error) {
	redusClientOpt, err := asynq.ParseRedisURI(redisAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid redis address: %w", err)
	}
	if strings.TrimSpace(queueName) == "" {
		return nil, fmt.Errorf("queue name is required")
	}
	if concurrency <= 0 {
		return nil, fmt.Errorf("concurrency must be greater than zero")
	}

	return asynq.NewServer(
		redusClientOpt,
		asynq.Config{
			Concurrency: concurrency,
			Queues:      map[string]int{queueName: 1},
		},
	), nil
}
