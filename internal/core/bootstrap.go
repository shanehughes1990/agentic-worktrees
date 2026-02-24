package core

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	infracopilot "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/copilot"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/logging/logruslogger"
	queueasynq "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
	jsontaskboard "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/taskboard/jsonrepo"
	"github.com/shanehughes1990/agentic-worktrees/internal/interface/dashboard"
	workeriface "github.com/shanehughes1990/agentic-worktrees/internal/interface/worker"
)

type Runtime struct {
	worker                 *queueasynq.Server
	workerRegistrations    []queueasynq.HandlerRegistration
	queueClient            *queueasynq.Client
	runtimeWorkflowRepo    *queueasynq.RuntimeWorkflowRepository
	ui                     *dashboard.UI
	taskboardService       *apptaskboard.Service
	ingestionCommand       *apptaskboard.IngestionService
	authService            *appcopilot.AuthService
	runtimeWorkflowService *apptaskboard.RuntimeWorkflowService
}

func Init() (*Runtime, error) {
	cfg, err := LoadAppConfigFromEnv()
	if err != nil {
		return nil, err
	}

	logger, err := logruslogger.New(cfg.Logging.Format, cfg.Logging.Level, cfg.Logging.FilePath)
	if err != nil {
		return nil, err
	}

	queueCfg, err := queueasynq.NewConfig(cfg.Redis.URI)
	if err != nil {
		return nil, err
	}
	queueCfg = queueCfg.WithLogger(logruslogger.NewAsynqAdapter(logger))

	taskboardRepository, err := jsontaskboard.NewRepository(cfg.Taskboard.JSONDirectory)
	if err != nil {
		return nil, err
	}

	copilotConfig := infracopilot.ClientConfig{
		GitHubToken:       cfg.Copilot.GitHubToken,
		CLIPath:           cfg.Copilot.CLIPath,
		CLIURL:            cfg.Copilot.CLIURL,
		AuthStatusCommand: cfg.Copilot.AuthStatusCommand,
		AuthLoginCommand:  cfg.Copilot.AuthLoginCommand,
		LogLevel:          "error",
		DefaultModel:      cfg.Copilot.Model,
		SkillDirectories:  cfg.Copilot.SkillDirectories,
	}.Normalized()

	queueClient := queueasynq.NewClient(queueCfg)
	runtimeWorkflowRepo := queueasynq.NewRuntimeWorkflowRepository(queueCfg)
	authenticator := infracopilot.NewAuthenticator(copilotConfig, logger)
	authService := appcopilot.NewAuthService(authenticator)
	decomposer := infracopilot.NewDecomposer(copilotConfig, logger)
	copilotHandler := workeriface.NewCopilotDecomposeHandler(decomposer, taskboardRepository, taskboardRepository, logger)
	workerRegistrations := []queueasynq.HandlerRegistration{
		{TaskType: tasks.TaskTypeCopilotDecompose, Handler: copilotHandler},
	}

	ingestionDispatcher := queueasynq.NewTaskboardIngestionDispatcher(queueClient, copilotConfig, logger)
	ingestionCommand := apptaskboard.NewIngestionService(ingestionDispatcher, taskboardRepository, taskboardRepository, cfg.Copilot.Model)
	runtimeWorkflowService := apptaskboard.NewRuntimeWorkflowService(runtimeWorkflowRepo)

	runtime := &Runtime{
		worker:                 queueasynq.NewServer(queueCfg),
		workerRegistrations:    workerRegistrations,
		queueClient:            queueClient,
		runtimeWorkflowRepo:    runtimeWorkflowRepo,
		taskboardService:       apptaskboard.NewService(taskboardRepository),
		ingestionCommand:       ingestionCommand,
		authService:            authService,
		runtimeWorkflowService: runtimeWorkflowService,
	}
	runtime.ui = dashboard.New(
		runtime.ingestionCommand.IngestDirectory,
		runtime.runtimeWorkflowService.ListWorkflows,
		runtime.runtimeWorkflowService.GetWorkflowStatus,
		runtime.authService.Status,
		runtime.authService.Authenticate,
	)

	return runtime, nil
}

func (runtime *Runtime) Run() error {
	defer runtime.queueClient.Close()
	defer func() {
		if runtime.runtimeWorkflowRepo != nil {
			_ = runtime.runtimeWorkflowRepo.Close()
		}
	}()

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	workerErr := make(chan error, 1)
	go func() {
		workerErr <- runtime.worker.Run(workerCtx, runtime.workerRegistrations)
	}()

	uiErr := make(chan error, 1)
	go func() {
		uiErr <- runtime.ui.Run()
	}()

	sigCtx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	shutdownDeadline := 5 * time.Second

	waitWorker := func() error {
		timer := time.NewTimer(shutdownDeadline)
		defer timer.Stop()
		select {
		case err := <-workerErr:
			if err != nil && !errors.Is(err, asynq.ErrServerClosed) {
				return fmt.Errorf("asynq worker stopped: %w", err)
			}
			return nil
		case <-timer.C:
			return fmt.Errorf("shutdown timeout waiting for worker")
		}
	}

	waitUI := func() error {
		timer := time.NewTimer(shutdownDeadline)
		defer timer.Stop()
		select {
		case err := <-uiErr:
			return err
		case <-timer.C:
			return fmt.Errorf("shutdown timeout waiting for dashboard")
		}
	}

	select {
	case err := <-workerErr:
		if err != nil && !errors.Is(err, asynq.ErrServerClosed) {
			runtime.ui.Stop()
			_ = waitUI()
			return fmt.Errorf("asynq worker stopped: %w", err)
		}
		runtime.ui.Stop()
		return waitUI()

	case err := <-uiErr:
		cancelWorker()
		workerStopErr := waitWorker()
		if err != nil {
			return err
		}
		return workerStopErr

	case <-sigCtx.Done():
		cancelWorker()
		if err := waitWorker(); err != nil {
			runtime.ui.Stop()
			_ = waitUI()
			return err
		}
		runtime.ui.Stop()
		return waitUI()
	}
}
