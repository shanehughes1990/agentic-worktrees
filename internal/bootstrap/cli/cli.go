package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	urcli "github.com/urfave/cli/v3"

	"github.com/shanehughes1990/agentic-worktrees/internal/application"
	bootstrapworker "github.com/shanehughes1990/agentic-worktrees/internal/bootstrap/worker"
	infradatabase "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/database"
	interfacefilesystem "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/filesystem"
	infralogger "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/logger"
	infraqueue "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue"
	interfaceclicommands "github.com/shanehughes1990/agentic-worktrees/internal/interface/cli/commands"
)

type Runtime struct {
	command       *urcli.Command
	queueClient   *infraqueue.AsynqClient
	workerRuntime *bootstrapworker.Runtime
}

type bootstrapper struct{}

func (bootstrapper) Bootstrap(logLevel string, logFormat string, databaseDSN string) (*logrus.Logger, error) {
	configuredLogger, err := infralogger.New(logLevel, logFormat)
	if err != nil {
		return nil, err
	}
	if _, err := infradatabase.NewGormClient(configuredLogger, databaseDSN); err != nil {
		return nil, fmt.Errorf("bootstrap database: %w", err)
	}
	return configuredLogger, nil
}

type runner struct {
	prepareCommand *application.PrepareGenerateTaskBoardCommand
	enqueuer       application.GenerateTaskBoardEnqueuer
}

func (r runner) Run(ctx context.Context, input application.GenerateTaskBoardInput) (string, error) {
	if r.prepareCommand == nil {
		return "", fmt.Errorf("prepare command cannot be nil")
	}
	if r.enqueuer == nil {
		return "", fmt.Errorf("enqueuer cannot be nil")
	}
	payload, err := r.prepareCommand.Execute(ctx, input)
	if err != nil {
		return "", err
	}
	return r.enqueuer.EnqueueGenerateTaskBoard(ctx, payload)
}

func New() (*Runtime, error) {
	redisAddress := envOrDefault("REDIS_ADDR", "127.0.0.1:6379")
	queueName := envOrDefault("ASYNQ_QUEUE", "default")

	queueClient, err := infraqueue.NewAsynqClient(redisAddress)
	if err != nil {
		return nil, err
	}
	generateTaskBoardClient, err := infraqueue.NewAsynqGenerateTaskBoardClient(queueClient, queueName)
	if err != nil {
		_ = queueClient.Close()
		return nil, err
	}

	prepareCommand, err := application.NewPrepareGenerateTaskBoardCommand(interfacefilesystem.NewDocumentationLoader())
	if err != nil {
		_ = queueClient.Close()
		return nil, err
	}

	workerRuntime, err := bootstrapworker.New()
	if err != nil {
		_ = queueClient.Close()
		return nil, err
	}

	command := interfaceclicommands.NewRootCommand(bootstrapper{}, runner{prepareCommand: prepareCommand, enqueuer: generateTaskBoardClient}, workerRuntime)
	return &Runtime{command: command, queueClient: queueClient, workerRuntime: workerRuntime}, nil
}

func (r *Runtime) Run(ctx context.Context) error {
	if r == nil {
		return fmt.Errorf("runtime cannot be nil")
	}
	if r.command == nil {
		return fmt.Errorf("command cannot be nil")
	}
	if r.queueClient == nil {
		return fmt.Errorf("queue client cannot be nil")
	}
	if r.workerRuntime == nil {
		return fmt.Errorf("worker runtime cannot be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	defer func() {
		_ = r.queueClient.Close()
	}()
	return r.command.Run(ctx, os.Args)
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
