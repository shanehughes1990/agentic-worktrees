package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	urcli "github.com/urfave/cli/v3"
	"github.com/xo/dburl"

	"github.com/shanehughes1990/agentic-worktrees/internal/application"
	infraagent "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/agent"
	infradatabase "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/database"
	interfacefilesystem "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/filesystem"
	infralogger "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/logger"
	infraqueue "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue"
	infrarepositories "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/repositories"
	interfaceclicommands "github.com/shanehughes1990/agentic-worktrees/internal/interface/cli/commands"
	interfaceworkercommands "github.com/shanehughes1990/agentic-worktrees/internal/interface/worker/commands"
)

type runtimeWorker interface {
	Run(ctx context.Context) error
}

type Runtime struct {
	command *urcli.Command

	logLevel            string
	logFormat           string
	databaseDSN         string
	redisAddress        string
	queueName           string
	resultQueueName     string
	asynqConcurrency    int
	waitTimeout         time.Duration
	copilotCLIPath      string
	copilotCLIURL       string
	githubToken         string
	copilotStartTimeout time.Duration

	logger         *logrus.Logger
	redisConnOpt   asynq.RedisConnOpt
	queueClient    *infraqueue.AsynqClient
	prepareCommand *application.PrepareGenerateTaskBoardCommand
	generateClient *infraqueue.AsynqGenerateTaskBoardClient

	workerRunner  runtimeWorker
	workerErrorCh chan error
	workerCancel  context.CancelFunc
}

type combinedWorkerRunner struct {
	agent        *infraagent.CopilotClient
	server       *asynq.Server
	handlerMux   *asynq.ServeMux
	logger       *logrus.Logger
	startTimeout time.Duration
}

func (r combinedWorkerRunner) Run(ctx context.Context) error {
	if r.agent == nil {
		return fmt.Errorf("agent client cannot be nil")
	}
	if r.server == nil {
		return fmt.Errorf("worker server cannot be nil")
	}
	if r.handlerMux == nil {
		return fmt.Errorf("worker handler mux cannot be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	startTimeout := r.startTimeout
	if startTimeout <= 0 {
		startTimeout = 10 * time.Second
	}
	startCtx, cancelStart := context.WithTimeout(ctx, startTimeout)
	defer cancelStart()
	if err := r.agent.Start(startCtx); err != nil {
		return fmt.Errorf("start copilot client: %w", err)
	}
	defer func() {
		if err := r.agent.Stop(); err != nil && r.logger != nil {
			r.logger.WithError(err).Error("failed to stop copilot client")
		}
	}()

	return r.server.Run(r.handlerMux)
}

func New() (*Runtime, error) {
	runtime := &Runtime{
		logLevel:            "info",
		logFormat:           "json",
		databaseDSN:         defaultDatabaseDSN(),
		redisAddress:        "127.0.0.1:6379",
		queueName:           "default",
		resultQueueName:     "default-result",
		asynqConcurrency:    10,
		waitTimeout:         10 * time.Minute,
		copilotStartTimeout: 10 * time.Second,
	}
	runtime.command = runtime.buildCommand()
	return runtime, nil
}

func (r *Runtime) buildCommand() *urcli.Command {
	rootDirectory := "docs"
	traversalDepth := 2
	boardPrompt := interfaceclicommands.DefaultBoardPrompt
	boardModel := "gpt-5.3-codex"
	runner := interfaceclicommands.NewGenerateTaskBoardRunner(func() (*interfaceclicommands.GenerateTaskBoardExecutionDependencies, error) {
		if r.prepareCommand == nil {
			return nil, fmt.Errorf("prepare command cannot be nil")
		}
		if r.generateClient == nil {
			return nil, fmt.Errorf("generate client cannot be nil")
		}
		if r.redisConnOpt == nil {
			return nil, fmt.Errorf("redis connection options cannot be nil")
		}
		return &interfaceclicommands.GenerateTaskBoardExecutionDependencies{
			PrepareCommand:  r.prepareCommand,
			GenerateClient:  r.generateClient,
			RedisConnOpt:    r.redisConnOpt,
			ResultQueueName: r.resultQueueName,
			WaitTimeout:     r.waitTimeout,
			WorkerErrorCh:   r.workerErrorCh,
		}, nil
	})

	return &urcli.Command{
		Name:  "cli",
		Usage: "bootstrap cli",
		Flags: []urcli.Flag{
			&urcli.StringFlag{Name: "LOG_LEVEL", Value: r.logLevel, Required: true, Destination: &r.logLevel, Sources: urcli.EnvVars("LOG_LEVEL")},
			&urcli.StringFlag{Name: "LOG_FORMAT", Value: r.logFormat, Required: true, Destination: &r.logFormat, Sources: urcli.EnvVars("LOG_FORMAT")},
			&urcli.StringFlag{Name: "DATABASE_DSN", Value: r.databaseDSN, Required: true, Destination: &r.databaseDSN, Sources: urcli.EnvVars("DATABASE_DSN")},
			&urcli.StringFlag{Name: "REDIS_ADDR", Value: r.redisAddress, Required: true, Destination: &r.redisAddress, Sources: urcli.EnvVars("REDIS_ADDR")},
			&urcli.StringFlag{Name: "ASYNQ_QUEUE", Value: r.queueName, Required: true, Destination: &r.queueName, Sources: urcli.EnvVars("ASYNQ_QUEUE")},
			&urcli.StringFlag{Name: "ASYNQ_RESULT_QUEUE", Value: r.resultQueueName, Required: true, Destination: &r.resultQueueName, Sources: urcli.EnvVars("ASYNQ_RESULT_QUEUE")},
			&urcli.IntFlag{Name: "ASYNQ_CONCURRENCY", Value: r.asynqConcurrency, Required: true, Destination: &r.asynqConcurrency, Sources: urcli.EnvVars("ASYNQ_CONCURRENCY")},
			&urcli.DurationFlag{Name: "ASYNQ_WAIT_TIMEOUT", Value: r.waitTimeout, Required: true, Destination: &r.waitTimeout, Sources: urcli.EnvVars("ASYNQ_WAIT_TIMEOUT")},
			&urcli.DurationFlag{Name: "COPILOT_START_TIMEOUT", Value: r.copilotStartTimeout, Required: true, Destination: &r.copilotStartTimeout, Sources: urcli.EnvVars("COPILOT_START_TIMEOUT")},
			&urcli.StringFlag{Name: "COPILOT_CLI_PATH", Value: "", Destination: &r.copilotCLIPath, Sources: urcli.EnvVars("COPILOT_CLI_PATH")},
			&urcli.StringFlag{Name: "COPILOT_CLI_URL", Value: "", Destination: &r.copilotCLIURL, Sources: urcli.EnvVars("COPILOT_CLI_URL")},
			&urcli.StringFlag{Name: "GITHUB_TOKEN", Value: "", Destination: &r.githubToken, Sources: urcli.EnvVars("GITHUB_TOKEN")},
		},
		Before: func(hookCtx context.Context, _ *urcli.Command) (context.Context, error) {
			if err := r.validateConfig(); err != nil {
				return hookCtx, err
			}
			if err := r.bootstrap(hookCtx); err != nil {
				return hookCtx, err
			}
			return hookCtx, nil
		},
		Commands: []*urcli.Command{
			interfaceclicommands.NewGenerateTaskBoardCommand(runner, interfaceclicommands.GenerateTaskBoardCommandOptions{
				RootDirectory:  &rootDirectory,
				TraversalDepth: &traversalDepth,
				BoardPrompt:    &boardPrompt,
				BoardModel:     &boardModel,
				LoggerProvider: func() *logrus.Logger { return r.logger },
			}),
		},
	}
}

func (r *Runtime) validateConfig() error {
	trimmedLogLevel := strings.TrimSpace(r.logLevel)
	if trimmedLogLevel == "" {
		return fmt.Errorf("LOG_LEVEL is required")
	}
	if _, err := logrus.ParseLevel(trimmedLogLevel); err != nil {
		return fmt.Errorf("invalid LOG_LEVEL %q: %w", r.logLevel, err)
	}

	trimmedLogFormat := strings.ToLower(strings.TrimSpace(r.logFormat))
	if trimmedLogFormat == "" {
		return fmt.Errorf("LOG_FORMAT is required")
	}
	if trimmedLogFormat != "json" && trimmedLogFormat != "text" {
		return fmt.Errorf("invalid LOG_FORMAT %q: expected json or text", r.logFormat)
	}

	if strings.TrimSpace(r.databaseDSN) == "" {
		return fmt.Errorf("DATABASE_DSN is required")
	}
	if _, err := dburl.Parse(strings.TrimSpace(r.databaseDSN)); err != nil {
		return fmt.Errorf("invalid DATABASE_DSN %q: %w", r.databaseDSN, err)
	}

	if strings.TrimSpace(r.redisAddress) == "" {
		return fmt.Errorf("REDIS_ADDR is required")
	}
	redisConnOpt, err := infraqueue.ResolveRedisClientOpt(r.redisAddress)
	if err != nil {
		return fmt.Errorf("invalid REDIS_ADDR %q: %w", r.redisAddress, err)
	}
	r.redisConnOpt = redisConnOpt

	if strings.TrimSpace(r.queueName) == "" {
		return fmt.Errorf("ASYNQ_QUEUE is required")
	}
	if strings.TrimSpace(r.resultQueueName) == "" {
		return fmt.Errorf("ASYNQ_RESULT_QUEUE is required")
	}
	if r.asynqConcurrency <= 0 {
		return fmt.Errorf("ASYNQ_CONCURRENCY must be greater than zero")
	}
	if r.waitTimeout <= 0 {
		return fmt.Errorf("ASYNQ_WAIT_TIMEOUT must be greater than zero")
	}
	if r.copilotStartTimeout <= 0 {
		return fmt.Errorf("COPILOT_START_TIMEOUT must be greater than zero")
	}

	return nil
}

func (r *Runtime) bootstrap(ctx context.Context) error {
	if err := r.validateConfig(); err != nil {
		return err
	}

	redisConnOpt := r.redisConnOpt
	if redisConnOpt == nil {
		resolvedRedisConnOpt, err := infraqueue.ResolveRedisClientOpt(r.redisAddress)
		if err != nil {
			return err
		}
		redisConnOpt = resolvedRedisConnOpt
		r.redisConnOpt = resolvedRedisConnOpt
	}

	appLogger, err := infralogger.New(r.logLevel, r.logFormat)
	if err != nil {
		return err
	}
	dbClient, err := infradatabase.NewGormClient(appLogger, r.databaseDSN)
	if err != nil {
		return err
	}
	boardRepository, err := infrarepositories.NewSQLiteBoardRepository(dbClient.DB())
	if err != nil {
		return err
	}
	persistCommand, err := application.NewPersistGenerateTaskBoardResultCommand(boardRepository)
	if err != nil {
		return err
	}
	persistHandler, err := interfaceclicommands.NewPersistGeneratedBoardResultHandler(persistCommand)
	if err != nil {
		return err
	}

	queueClient, err := infraqueue.NewAsynqClientWithRedisOpt(redisConnOpt)
	if err != nil {
		return err
	}
	generateClient, err := infraqueue.NewAsynqGenerateTaskBoardClient(queueClient, r.queueName)
	if err != nil {
		_ = queueClient.Close()
		return err
	}
	resultPublisher, err := infraqueue.NewAsynqGenerateTaskBoardClient(queueClient, r.resultQueueName)
	if err != nil {
		_ = queueClient.Close()
		return err
	}
	prepareCommand, err := application.NewPrepareGenerateTaskBoardCommand(interfacefilesystem.NewDocumentationLoader())
	if err != nil {
		_ = queueClient.Close()
		return err
	}

	agentClient, err := infraagent.NewCopilotClient(appLogger, r.copilotCLIPath, r.copilotCLIURL, r.githubToken)
	if err != nil {
		_ = queueClient.Close()
		return err
	}
	executeCommand, err := application.NewExecuteGenerateTaskBoardCommand(agentClient)
	if err != nil {
		_ = queueClient.Close()
		return err
	}
	generateHandler, err := interfaceworkercommands.NewGenerateTaskBoardHandler(executeCommand, resultPublisher)
	if err != nil {
		_ = queueClient.Close()
		return err
	}

	workerServer, err := infraqueue.NewAsynqServerWithQueuesAndRedisOpt(redisConnOpt, map[string]int{r.queueName: 1, r.resultQueueName: 1}, r.asynqConcurrency)
	if err != nil {
		_ = queueClient.Close()
		return err
	}
	workerMux := asynq.NewServeMux()
	workerMux.HandleFunc(application.AsynqTaskTypeGenerateTaskBoard, generateHandler.Handle)
	workerMux.HandleFunc(application.AsynqTaskTypeGenerateTaskBoardResult, persistHandler.Handle)

	r.workerRunner = &combinedWorkerRunner{
		agent:        agentClient,
		server:       workerServer,
		handlerMux:   workerMux,
		logger:       appLogger,
		startTimeout: r.copilotStartTimeout,
	}
	r.logger = appLogger
	r.queueClient = queueClient
	r.prepareCommand = prepareCommand
	r.generateClient = generateClient
	r.workerErrorCh = make(chan error, 1)

	workerCtx, workerCancel := context.WithCancel(ctx)
	r.workerCancel = workerCancel
	go func() {
		workerErr := r.workerRunner.Run(workerCtx)
		if workerErr == nil {
			return
		}
		select {
		case r.workerErrorCh <- workerErr:
		default:
		}
	}()

	select {
	case workerErr := <-r.workerErrorCh:
		workerCancel()
		_ = queueClient.Close()
		return fmt.Errorf("worker bootstrap failed: %w", workerErr)
	case <-time.After(2 * time.Second):
	}

	return nil
}

func (r *Runtime) Run(ctx context.Context) error {
	if r == nil {
		return fmt.Errorf("runtime cannot be nil")
	}
	if r.command == nil {
		return fmt.Errorf("command cannot be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	defer func() {
		if r.workerCancel != nil {
			r.workerCancel()
		}
		if r.queueClient != nil {
			_ = r.queueClient.Close()
		}
	}()
	return r.command.Run(ctx, os.Args)
}

func defaultDatabaseDSN() string {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "sqlite:agentic-worktrees.db"
	}
	databasePath := filepath.Join(workingDirectory, "agentic-worktrees.db")
	return "sqlite:///" + filepath.ToSlash(databasePath)
}
