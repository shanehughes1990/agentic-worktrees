package asynq

import (
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
)

type Config struct {
	RedisURL    string
	Concurrency int
}

type APIConfig struct {
	RedisURL string
}

type WorkerConfig struct {
	RedisURL    string
	Concurrency int
}

func (config Config) normalized() Config {
	normalizedConfig := config
	if strings.TrimSpace(normalizedConfig.RedisURL) == "" {
		normalizedConfig.RedisURL = "redis://127.0.0.1:6379/0"
	}
	if normalizedConfig.Concurrency <= 0 {
		normalizedConfig.Concurrency = 10
	}
	return normalizedConfig
}

func (config APIConfig) normalized() APIConfig {
	normalizedConfig := config
	if strings.TrimSpace(normalizedConfig.RedisURL) == "" {
		normalizedConfig.RedisURL = "redis://127.0.0.1:6379/0"
	}
	return normalizedConfig
}

func (config WorkerConfig) normalized() WorkerConfig {
	normalizedConfig := config
	if strings.TrimSpace(normalizedConfig.RedisURL) == "" {
		normalizedConfig.RedisURL = "redis://127.0.0.1:6379/0"
	}
	if normalizedConfig.Concurrency <= 0 {
		normalizedConfig.Concurrency = 10
	}
	return normalizedConfig
}

func (config Config) redisClientOpt() (asynq.RedisConnOpt, error) {
	normalized := config.normalized()
	redisConnOpt, err := asynq.ParseRedisURI(normalized.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return redisConnOpt, nil
}

func (config APIConfig) redisClientOpt() (asynq.RedisConnOpt, error) {
	normalized := config.normalized()
	redisConnOpt, err := asynq.ParseRedisURI(normalized.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return redisConnOpt, nil
}

func (config WorkerConfig) redisClientOpt() (asynq.RedisConnOpt, error) {
	normalized := config.normalized()
	redisConnOpt, err := asynq.ParseRedisURI(normalized.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return redisConnOpt, nil
}
