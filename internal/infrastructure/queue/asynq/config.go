package asynq

import (
	"errors"
	"runtime"
	"strings"

	"github.com/hibiken/asynq"
)

var ErrRedisURIRequired = errors.New("redis URI is required")

type Config struct {
	redisConnOpt asynq.RedisConnOpt
	Concurrency  int
	Queues       map[string]int
	Logger       asynq.Logger
}

func NewConfig(redisURI string) (Config, error) {
	trimmedRedisURI := strings.TrimSpace(redisURI)
	if trimmedRedisURI == "" {
		return Config{}, ErrRedisURIRequired
	}

	redisConnOpt, err := asynq.ParseRedisURI(trimmedRedisURI)
	if err != nil {
		return Config{}, err
	}

	concurrency := runtime.GOMAXPROCS(0)
	if concurrency < 1 {
		concurrency = 1
	}

	return Config{
		redisConnOpt: redisConnOpt,
		Concurrency:  concurrency,
		Queues: map[string]int{
			"ingestion": 6,
			"agent":     3,
			"default":   1,
		},
	}, nil
}

func (cfg Config) WithLogger(logger asynq.Logger) Config {
	cfg.Logger = logger
	return cfg
}

func (cfg Config) serverConfig() asynq.Config {
	serverCfg := asynq.Config{
		Concurrency: cfg.Concurrency,
		Logger:      cfg.Logger,
	}
	if len(cfg.Queues) > 0 {
		serverCfg.Queues = cfg.Queues
	}
	return serverCfg
}
