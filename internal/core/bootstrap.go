package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	infracopilot "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/copilot"
	infragit "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/git"
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
	executionCommand       *apptaskboard.ExecutionCommandService
	executionControl       *apptaskboard.ExecutionControlService
	authService            *appcopilot.AuthService
	runtimeWorkflowService *apptaskboard.RuntimeWorkflowService
	repositoryRoot         string
}

type taskPipelineExecutorAdapter struct {
	inner *appgitflow.TaskExecutor
}

func (adapter *taskPipelineExecutorAdapter) ExecuteTask(ctx context.Context, request apptaskboard.TaskExecutionRequest) (apptaskboard.TaskExecutionOutcome, error) {
	if adapter == nil || adapter.inner == nil {
		return apptaskboard.TaskExecutionOutcome{}, fmt.Errorf("task executor adapter is not configured")
	}
	result, err := adapter.inner.ExecuteTask(ctx, appgitflow.TaskExecutionRequest{
		BoardID:        request.BoardID,
		RunID:          request.RunID,
		TaskID:         request.TaskID,
		TaskTitle:      request.TaskTitle,
		TaskDetail:     request.TaskDetail,
		SourceBranch:   request.SourceBranch,
		RepositoryRoot: request.RepositoryRoot,
	})
	if err != nil {
		return apptaskboard.TaskExecutionOutcome{}, err
	}
	return apptaskboard.TaskExecutionOutcome{
		Status:     result.Status,
		Reason:     result.Reason,
		TaskBranch: result.TaskBranch,
		Worktree:   result.Worktree,
	}, nil
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
	gitAdapter := infragit.NewAdapter(logger)
	gitWorktreeDispatcher := queueasynq.NewGitWorktreeDispatcher(queueClient, logger)
	gitflowRunner := appgitflow.NewRunner(gitAdapter, gitWorktreeDispatcher, taskboardRepository)
	taskExecutor := appgitflow.NewTaskExecutor(gitAdapter, decomposer)
	executionRegistry := apptaskboard.NewExecutionRegistry()
	taskboardService := apptaskboard.NewService(taskboardRepository)
	executionPipeline := apptaskboard.NewExecutionPipelineService(taskboardService, &taskPipelineExecutorAdapter{inner: taskExecutor}, taskboardRepository, cfg.Taskboard.MaxConcurrentAgents)
	taskboardExecutionHandler := workeriface.NewTaskboardExecuteHandler(executionPipeline, executionRegistry, logger)
	taskboardExecutionDispatcher := queueasynq.NewTaskboardExecutionDispatcher(queueClient, logger)
	executionCommand := apptaskboard.NewExecutionCommandService(taskboardExecutionDispatcher)
	executionControl := apptaskboard.NewExecutionControlService(executionRegistry, taskExecutor)
	copilotHandler := workeriface.NewCopilotDecomposeHandler(decomposer, taskboardRepository, taskboardRepository, logger)
	gitWorktreeFlowHandler := workeriface.NewGitWorktreeFlowHandler(gitflowRunner, logger)
	gitConflictResolveHandler := workeriface.NewGitConflictResolveHandler(gitflowRunner, decomposer, logger)
	workerRegistrations := []queueasynq.HandlerRegistration{
		{TaskType: tasks.TaskTypeCopilotDecompose, Handler: copilotHandler},
		{TaskType: tasks.TaskTypeGitWorktreeFlow, Handler: gitWorktreeFlowHandler},
		{TaskType: tasks.TaskTypeGitConflictResolve, Handler: gitConflictResolveHandler},
		{TaskType: tasks.TaskTypeTaskboardExecute, Handler: taskboardExecutionHandler},
	}

	ingestionDispatcher := queueasynq.NewTaskboardIngestionDispatcher(queueClient, copilotConfig, logger)
	ingestionCommand := apptaskboard.NewIngestionService(ingestionDispatcher, taskboardRepository, taskboardRepository, cfg.Copilot.Model)
	runtimeWorkflowService := apptaskboard.NewRuntimeWorkflowService(runtimeWorkflowRepo)

	repositoryRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("determine repository root from working directory: %w", err)
	}

	runtime := &Runtime{
		worker:                 queueasynq.NewServer(queueCfg),
		workerRegistrations:    workerRegistrations,
		queueClient:            queueClient,
		runtimeWorkflowRepo:    runtimeWorkflowRepo,
		taskboardService:       taskboardService,
		ingestionCommand:       ingestionCommand,
		executionCommand:       executionCommand,
		executionControl:       executionControl,
		authService:            authService,
		runtimeWorkflowService: runtimeWorkflowService,
		repositoryRoot:         repositoryRoot,
	}
	runtime.ui = dashboard.New(
		runtime.ingestionCommand.IngestDirectory,
		func(ctx context.Context, boardID string, sourceBranch string) (string, error) {
			return runtime.executionCommand.Start(ctx, apptaskboard.StartExecutionRequest{
				BoardID:        boardID,
				RepositoryRoot: runtime.repositoryRoot,
				SourceBranch:   sourceBranch,
			})
		},
		func(ctx context.Context, boardID string) (string, error) {
			canceled, err := runtime.executionControl.CancelAndCleanup(ctx, boardID, runtime.repositoryRoot)
			if err != nil {
				return "", err
			}
			if canceled {
				return fmt.Sprintf("Canceled runner for board %s and cleaned artifacts", boardID), nil
			}
			return fmt.Sprintf("No active runner for board %s; cleanup completed", boardID), nil
		},
		func(ctx context.Context) ([]string, error) {
			return runtime.taskboardService.ListBoardIDs(ctx)
		},
		func(ctx context.Context, boardID string) ([]string, error) {
			readyTasks, err := runtime.taskboardService.GetReadyTasks(ctx, boardID)
			if err != nil {
				return nil, err
			}
			readyTaskIDs := make([]string, 0, len(readyTasks))
			for _, readyTask := range readyTasks {
				if readyTask == nil {
					continue
				}
				readyTaskIDs = append(readyTaskIDs, readyTask.ID)
			}
			return readyTaskIDs, nil
		},
		runtime.runtimeWorkflowService.ListWorkflows,
		runtime.runtimeWorkflowService.GetWorkflowStatus,
		runtime.authService.Status,
		runtime.authService.Authenticate,
		runtime.repositoryRoot,
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
