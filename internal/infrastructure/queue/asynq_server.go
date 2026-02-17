package queue

import (
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
)

func NewAsynqServer(redisAddress string, queueName string, concurrency int) (*asynq.Server, error) {
	if strings.TrimSpace(queueName) == "" {
		return nil, fmt.Errorf("queue name is required")
	}
	return NewAsynqServerWithQueues(redisAddress, map[string]int{queueName: 1}, concurrency)
}

func NewAsynqServerWithQueues(redisAddress string, queues map[string]int, concurrency int) (*asynq.Server, error) {
	redisClientOpt, err := ResolveRedisClientOpt(redisAddress)
	if err != nil {
		return nil, err
	}
	return NewAsynqServerWithQueuesAndRedisOpt(redisClientOpt, queues, concurrency)
}

func NewAsynqServerWithQueuesAndRedisOpt(redisClientOpt asynq.RedisConnOpt, queues map[string]int, concurrency int) (*asynq.Server, error) {
	if redisClientOpt == nil {
		return nil, fmt.Errorf("redis connection options are required")
	}
	if len(queues) == 0 {
		return nil, fmt.Errorf("at least one queue is required")
	}
	for name, weight := range queues {
		if strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("queue name is required")
		}
		if weight <= 0 {
			return nil, fmt.Errorf("queue weight must be greater than zero")
		}
	}
	if concurrency <= 0 {
		return nil, fmt.Errorf("concurrency must be greater than zero")
	}

	return asynq.NewServer(
		redisClientOpt,
		asynq.Config{
			Concurrency: concurrency,
			Queues:      queues,
		},
	), nil
}

func ResolveRedisClientOpt(redisAddress string) (asynq.RedisConnOpt, error) {
	trimmedAddress := strings.TrimSpace(redisAddress)
	if trimmedAddress == "" {
		return nil, fmt.Errorf("redis address is required")
	}
	if strings.Contains(trimmedAddress, "://") {
		parsed, err := asynq.ParseRedisURI(trimmedAddress)
		if err != nil {
			return nil, fmt.Errorf("invalid redis address: %w", err)
		}
		return parsed, nil
	}
	return asynq.RedisClientOpt{Addr: trimmedAddress}, nil
}
