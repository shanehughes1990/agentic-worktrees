package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	checkpointstore "github.com/shanehughes1990/agentic-worktrees/internal/features/checkpoints/store"
	"github.com/shanehughes1990/agentic-worktrees/internal/features/ingestion/adk"
	ingestionservice "github.com/shanehughes1990/agentic-worktrees/internal/features/ingestion/service"
	"github.com/shanehughes1990/agentic-worktrees/internal/features/queue/consumer"
)

func (runtime *Runtime) runWorker(ctx context.Context) error {
	if strings.TrimSpace(runtime.config.ADKBoardURL) == "" {
		return fmt.Errorf("COPILOT_ADK_BOARD_URL is required for worker")
	}

	adkClient, err := adk.NewClient(runtime.config.ADKBoardURL, runtime.config.ADKAuthToken, 60*time.Second)
	if err != nil {
		return err
	}

	ingestionSvc := ingestionservice.New(adkClient)
	checkpointSvc := checkpointstore.NewJSONStore(runtime.config.CheckpointPath)
	handler := &consumer.Handler{
		IngestionService: ingestionSvc,
		CheckpointStore:  checkpointSvc,
		Logger:           runtime.logger,
	}

	worker, err := consumer.NewWorker(runtime.config.RedisAddr, runtime.config.QueueName, runtime.config.Concurrency, handler)
	if err != nil {
		return err
	}

	runtime.logger.WithFields(logrus.Fields{
		"component":   "worker",
		"redis_addr":  runtime.config.RedisAddr,
		"queue":       runtime.config.QueueName,
		"concurrency": runtime.config.Concurrency,
	}).Info("worker started")

	return worker.Run(ctx)
}
