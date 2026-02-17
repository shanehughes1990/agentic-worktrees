package worker

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"

	"github.com/shanehughes1990/agentic-worktrees/internal/application"
	infraagent "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/agent"
	infralogger "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/logger"
	infraqueue "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue"
	interfaceworkercommands "github.com/shanehughes1990/agentic-worktrees/internal/interface/worker/commands"
)

type Runtime struct {
	logger        *logrus.Logger
	agent         *infraagent.CopilotClient
	resultClient  *infraqueue.AsynqGenerateTaskBoardClient
	queueClient   *infraqueue.AsynqClient
	server        *asynq.Server
	handler       *asynq.ServeMux
}

func New() (*Runtime, error) {
	logger, err := infralogger.New(envOrDefault("LOG_LEVEL", "info"), envOrDefault("LOG_FORMAT", "json"))
	if err != nil {
		return nil, err
	}

	agentClient, err := infraagent.NewCopilotClient(logger, os.Getenv("COPILOT_CLI_PATH"), os.Getenv("COPILOT_CLI_URL"), os.Getenv("GITHUB_TOKEN"))
	if err != nil {
		return nil, err
	}

	executeCommand, err := application.NewExecuteGenerateTaskBoardCommand(agentClient)
	if err != nil {
		return nil, err
	}

	queueName := envOrDefault("ASYNQ_QUEUE", "default")
	redisAddress := envOrDefault("REDIS_ADDR", "127.0.0.1:6379")

	server, err := infraqueue.NewAsynqServer(redisAddress, queueName, intEnvOrDefault("ASYNQ_CONCURRENCY", 10))
	if err != nil {
		return nil, err
	}

	queueClient, err := infraqueue.NewAsynqClient(redisAddress)
	if err != nil {
		return nil, err
	}

	resultQueue := envOrDefault("ASYNQ_RESULT_QUEUE", queueName+"-result")
	resultClient, err := infraqueue.NewAsynqGenerateTaskBoardClient(queueClient, resultQueue)
	if err != nil {
		_ = queueClient.Close()
		return nil, err
	}

	taskHandler, err := interfaceworkercommands.NewGenerateTaskBoardHandler(executeCommand, resultClient)
	if err != nil {
		_ = queueClient.Close()
		return nil, err
	}

	handler := asynq.NewServeMux()
	handler.HandleFunc(application.AsynqTaskTypeGenerateTaskBoard, taskHandler.Handle)

	return &Runtime{logger: logger, agent: agentClient, resultClient: resultClient, queueClient: queueClient, server: server, handler: handler}, nil
}

func (r *Runtime) Run(ctx context.Context) error {
	if r == nil {
		return fmt.Errorf("runtime cannot be nil")
	}
	if r.server == nil || r.handler == nil {
		return fmt.Errorf("runtime server is not initialized")
	}
	if r.resultClient == nil {
		return fmt.Errorf("runtime result client is not initialized")
	}
	if r.queueClient == nil {
		return fmt.Errorf("runtime queue client is not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if err := r.agent.Start(ctx); err != nil {
		return err
	}
	defer func() {
		if err := r.agent.Stop(); err != nil {
			r.logger.WithError(err).Error("failed to stop copilot client")
		}
	}()
	defer func() {
		if err := r.queueClient.Close(); err != nil {
			r.logger.WithError(err).Error("failed to close result queue client")
		}
	}()

	return r.server.Run(r.handler)
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func intEnvOrDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
