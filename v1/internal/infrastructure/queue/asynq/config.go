package asynq

import "strings"

type Config struct {
	RedisAddress  string
	RedisPassword string
	RedisDatabase int
	Concurrency   int
}

func (config Config) normalized() Config {
	normalizedConfig := config
	if strings.TrimSpace(normalizedConfig.RedisAddress) == "" {
		normalizedConfig.RedisAddress = "127.0.0.1:6379"
	}
	if normalizedConfig.Concurrency <= 0 {
		normalizedConfig.Concurrency = 10
	}
	if normalizedConfig.RedisDatabase < 0 {
		normalizedConfig.RedisDatabase = 0
	}
	return normalizedConfig
}
