package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/logging/logruslogger"
	queueasynq "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq"
)

func main() {
	logger := logruslogger.NewFromEnv()

	redisURI := os.Getenv("ASYNQ_REDIS_URI")
	cfg, err := queueasynq.NewConfig(redisURI)
	if err != nil {
		logger.WithError(err).Fatal("invalid asynq redis configuration")
	}
	cfg = cfg.WithLogger(logruslogger.NewAsynqAdapter(logger))

	server := queueasynq.NewServer(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := server.Run(ctx, nil); err != nil {
		logger.WithError(err).Fatal("worker stopped with error")
	}
}
