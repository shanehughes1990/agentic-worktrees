package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
	infracopilot "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/copilot"
	infragit "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/git"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/logging/logruslogger"
	queueasynq "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
	filesystemsource "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/taskboard/filesystemsource"
	jsontaskboard "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/taskboard/jsonrepo"
	"github.com/shanehughes1990/agentic-worktrees/internal/interface/dashboard"
	workeriface "github.com/shanehughes1990/agentic-worktrees/internal/interface/worker"
	"github.com/sirupsen/logrus"
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
	logger                 *logrus.Logger
}

type taskPipelineExecutorAdapter struct {
	dispatcher       *appgitflow.Service
	taskboardService *apptaskboard.Service
	worktreeRoot     string
	pollInterval     time.Duration
	logger           *logrus.Logger
}

type taskResumeCheckpointAdapter struct {
	taskboardService *apptaskboard.Service
}

func (adapter *taskResumeCheckpointAdapter) CheckpointResumeSession(ctx context.Context, boardID string, taskID string, resumeSessionID string) error {
	if adapter == nil || adapter.taskboardService == nil {
		return nil
	}
	return adapter.taskboardService.CheckpointTaskResumeSession(ctx, boardID, taskID, resumeSessionID)
}

func (adapter *taskPipelineExecutorAdapter) ExecuteTask(ctx context.Context, request apptaskboard.TaskExecutionRequest) (apptaskboard.TaskExecutionOutcome, error) {
	entry := adapter.entry().WithFields(logrus.Fields{
		"event":           "core.task_pipeline_adapter.execute_task",
		"run_id":          strings.TrimSpace(request.RunID),
		"board_id":        strings.TrimSpace(request.BoardID),
		"task_id":         strings.TrimSpace(request.TaskID),
		"source_branch":   strings.TrimSpace(request.SourceBranch),
		"repository_root": strings.TrimSpace(request.RepositoryRoot),
	})
	if adapter == nil || adapter.dispatcher == nil {
		entry.Error("task executor adapter is not configured")
		return apptaskboard.TaskExecutionOutcome{}, fmt.Errorf("task executor adapter is not configured")
	}
	if adapter.taskboardService == nil {
		entry.Error("taskboard service is required")
		return apptaskboard.TaskExecutionOutcome{}, fmt.Errorf("taskboard service is required")
	}
	entry.Info("dispatching task execution")

	startResult, err := adapter.dispatcher.Start(ctx, appgitflow.StartRequest{
		RunID:           request.RunID,
		BoardID:         request.BoardID,
		TaskID:          request.TaskID,
		TaskTitle:       request.TaskTitle,
		TaskDetail:      request.TaskDetail,
		ResumeSessionID: request.ResumeSessionID,
		RepositoryRoot:  request.RepositoryRoot,
		SourceBranch:    request.SourceBranch,
		WorktreeRoot:    strings.TrimSpace(adapter.worktreeRoot),
	})
	if err != nil {
		entry.WithError(err).Error("failed to dispatch git worktree flow")
		return apptaskboard.TaskExecutionOutcome{}, err
	}
	entry.WithFields(logrus.Fields{"queue_task_id": startResult.QueueTaskID, "task_branch": startResult.TaskBranch, "worktree": startResult.Worktree}).Info("dispatched git worktree flow")

	pollInterval := adapter.pollInterval
	if pollInterval <= 0 {
		pollInterval = 250 * time.Millisecond
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			entry.WithError(ctx.Err()).Warn("task execution polling canceled")
			return apptaskboard.TaskExecutionOutcome{TaskBranch: startResult.TaskBranch, Worktree: startResult.Worktree}, ctx.Err()
		case <-ticker.C:
			task, getErr := adapter.taskboardService.GetTaskByID(ctx, strings.TrimSpace(request.BoardID), strings.TrimSpace(request.TaskID))
			if getErr != nil {
				entry.WithError(getErr).Error("failed to load task state while polling")
				return apptaskboard.TaskExecutionOutcome{TaskBranch: startResult.TaskBranch, Worktree: startResult.Worktree}, fmt.Errorf("load task state: %w", getErr)
			}
			if task == nil {
				continue
			}

			outcome := apptaskboard.TaskExecutionOutcome{
				TaskBranch: startResult.TaskBranch,
				Worktree:   startResult.Worktree,
			}
			if task.Outcome != nil {
				outcome.Status = strings.TrimSpace(task.Outcome.Status)
				outcome.Reason = strings.TrimSpace(task.Outcome.Reason)
				if strings.TrimSpace(task.Outcome.TaskBranch) != "" {
					outcome.TaskBranch = strings.TrimSpace(task.Outcome.TaskBranch)
				}
				if strings.TrimSpace(task.Outcome.Worktree) != "" {
					outcome.Worktree = strings.TrimSpace(task.Outcome.Worktree)
				}
				outcome.ResumeSessionID = strings.TrimSpace(task.Outcome.ResumeSessionID)
			}

			switch task.Status {
			case domaintaskboard.StatusCompleted:
				entry.WithFields(logrus.Fields{"final_status": outcome.Status, "reason": outcome.Reason, "resume_session_id": outcome.ResumeSessionID}).Info("task reached completed status")
				if outcome.Status == "" {
					outcome.Status = "merged"
				}
				if outcome.Reason == "" {
					outcome.Reason = "task execution completed"
				}
				return outcome, nil
			case domaintaskboard.StatusBlocked:
				entry.WithFields(logrus.Fields{"final_status": outcome.Status, "reason": outcome.Reason, "resume_session_id": outcome.ResumeSessionID}).Error("task reached blocked status")
				if outcome.Status == "" {
					outcome.Status = "failed"
				}
				if outcome.Reason == "" {
					outcome.Reason = "task execution failed"
				}
				return outcome, fmt.Errorf("%s", outcome.Reason)
			case domaintaskboard.StatusNotStarted:
				if strings.EqualFold(outcome.Status, "canceled") {
					entry.WithFields(logrus.Fields{"final_status": outcome.Status, "reason": outcome.Reason, "resume_session_id": outcome.ResumeSessionID}).Warn("task returned to not-started with canceled outcome")
					if outcome.Reason == "" {
						outcome.Reason = "task execution canceled"
					}
					return outcome, context.Canceled
				}
			}
		}
	}
}

func (adapter *taskPipelineExecutorAdapter) entry() *logrus.Entry {
	if adapter == nil || adapter.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(adapter.logger)
}

func Init() (*Runtime, error) {
	cfg, err := LoadAppConfigFromEnv()
	if err != nil {
		return nil, err
	}

	logger, err := logruslogger.New(cfg.Logging.Format, cfg.Logging.Level, defaultRuntimeLogFilePath(cfg))
	if err != nil {
		return nil, err
	}

	queueCfg, err := queueasynq.NewConfig(cfg.Redis.URI)
	if err != nil {
		return nil, err
	}
	queueCfg = queueCfg.WithLogger(logruslogger.NewAsynqAdapter(logger))

	taskboardRepository, err := jsontaskboard.NewRepositoryWithWorkflowDirectory(runtimeTaskboardsDirectory(cfg), runtimeWorkflowsDirectory(cfg))
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
	runtimeWorkflowRepo := queueasynq.NewRuntimeWorkflowRepository(queueCfg, logger)
	authenticator := infracopilot.NewAuthenticator(copilotConfig, logger)
	authService := appcopilot.NewAuthService(authenticator, logger)
	decomposer := infracopilot.NewDecomposer(copilotConfig, logger)
	gitAdapter := infragit.NewAdapter(logger)
	sourceBranchService := appgitflow.NewSourceBranchService(gitAdapter, logger)
	gitWorktreeDispatcher := queueasynq.NewGitWorktreeDispatcher(queueClient, logger)
	gitflowService := appgitflow.NewService(gitWorktreeDispatcher, logger)
	gitflowRunner := appgitflow.NewRunner(gitAdapter, gitWorktreeDispatcher, taskboardRepository, logger)
	taskboardService := apptaskboard.NewService(taskboardRepository, logger)
	taskExecutor := appgitflow.NewTaskExecutorWithLogger(gitAdapter, decomposer, logger, &taskResumeCheckpointAdapter{taskboardService: taskboardService})
	executionRegistry := apptaskboard.NewExecutionRegistry()
	effectiveMaxAgents := queueCfg.Concurrency
	if effectiveMaxAgents < 1 {
		effectiveMaxAgents = 1
	}
	executionPipeline := apptaskboard.NewExecutionPipelineService(taskboardService, &taskPipelineExecutorAdapter{dispatcher: gitflowService, taskboardService: taskboardService, worktreeRoot: runtimeRootDirectory(cfg), logger: logger}, taskboardRepository, effectiveMaxAgents, logger)
	taskboardExecutionHandler := workeriface.NewTaskboardExecuteHandler(executionPipeline, executionRegistry, logger)
	taskboardExecutionDispatcher := queueasynq.NewTaskboardExecutionDispatcher(queueClient, logger)
	executionCommand := apptaskboard.NewExecutionCommandService(taskboardExecutionDispatcher, logger)
	executionControl := apptaskboard.NewExecutionControlService(executionRegistry, taskExecutor, logger)
	copilotHandler := workeriface.NewCopilotDecomposeHandler(decomposer, taskboardRepository, taskboardRepository, logger)
	gitWorktreeFlowHandler := workeriface.NewGitWorktreeFlowHandler(taskExecutor, taskboardService, logger)
	gitConflictResolveHandler := workeriface.NewGitConflictResolveHandler(gitflowRunner, decomposer, logger)
	workerRegistrations := []queueasynq.HandlerRegistration{
		{TaskType: tasks.TaskTypeCopilotDecompose, Handler: copilotHandler},
		{TaskType: tasks.TaskTypeGitWorktreeFlow, Handler: gitWorktreeFlowHandler},
		{TaskType: tasks.TaskTypeGitConflictResolve, Handler: gitConflictResolveHandler},
		{TaskType: tasks.TaskTypeTaskboardExecute, Handler: taskboardExecutionHandler},
	}

	ingestionDispatcher := queueasynq.NewTaskboardIngestionDispatcher(queueClient, copilotConfig, logger)
	sourceAdapter := filesystemsource.NewAdapter()
	ingestionCommand := apptaskboard.NewIngestionService(ingestionDispatcher, taskboardRepository, taskboardRepository, sourceAdapter, sourceAdapter, cfg.Copilot.Model, logger)
	runtimeWorkflowService := apptaskboard.NewRuntimeWorkflowService(runtimeWorkflowRepo, logger)

	repositoryRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("determine repository root from working directory: %w", err)
	}

	defaultSourceBranch := ""
	resolveBranchCtx, resolveBranchCancel := context.WithTimeout(context.Background(), 5*time.Second)
	resolvedSourceBranch, resolveBranchErr := sourceBranchService.Resolve(resolveBranchCtx, repositoryRoot)
	resolveBranchCancel()
	if resolveBranchErr != nil {
		logger.WithError(resolveBranchErr).WithField("repository_root", repositoryRoot).Warn("failed to resolve current source branch for dashboard prefill")
	} else {
		defaultSourceBranch = resolvedSourceBranch
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
		logger:                 logger,
	}
	logger.WithFields(logrus.Fields{"event": "core.init.complete", "repository_root": repositoryRoot, "max_agents": effectiveMaxAgents}).Info("runtime initialized")
	runtime.ui = dashboard.New(
		func(ctx context.Context, request apptaskboard.IngestRequest, redisURI string) (apptaskboard.IngestionResult, error) {
			cleanRedisURI := strings.TrimSpace(redisURI)
			if cleanRedisURI == "" || cleanRedisURI == cfg.Redis.URI {
				return runtime.ingestionCommand.Ingest(ctx, request)
			}
			overrideCfg, overrideErr := queueasynq.NewConfig(cleanRedisURI)
			if overrideErr != nil {
				return apptaskboard.IngestionResult{}, overrideErr
			}
			overrideCfg = overrideCfg.WithLogger(logruslogger.NewAsynqAdapter(logger))
			overrideClient := queueasynq.NewClient(overrideCfg)
			defer overrideClient.Close()
			overrideDispatcher := queueasynq.NewTaskboardIngestionDispatcher(overrideClient, copilotConfig, logger)
			overrideIngestion := apptaskboard.NewIngestionService(overrideDispatcher, taskboardRepository, taskboardRepository, sourceAdapter, sourceAdapter, cfg.Copilot.Model, logger)
			return overrideIngestion.Ingest(ctx, request)
		},
		func(ctx context.Context, boardID string, sourceBranch string, maxTasks int, redisURI string) (string, error) {
			cleanRedisURI := strings.TrimSpace(redisURI)
			if cleanRedisURI == "" || cleanRedisURI == cfg.Redis.URI {
				return runtime.executionCommand.Start(ctx, apptaskboard.StartExecutionRequest{
					BoardID:        boardID,
					RepositoryRoot: runtime.repositoryRoot,
					SourceBranch:   sourceBranch,
					MaxTasks:       maxTasks,
				})
			}
			overrideCfg, overrideErr := queueasynq.NewConfig(cleanRedisURI)
			if overrideErr != nil {
				return "", overrideErr
			}
			overrideCfg = overrideCfg.WithLogger(logruslogger.NewAsynqAdapter(logger))
			overrideClient := queueasynq.NewClient(overrideCfg)
			defer overrideClient.Close()
			overrideDispatcher := queueasynq.NewTaskboardExecutionDispatcher(overrideClient, logger)
			overrideExecutionCommand := apptaskboard.NewExecutionCommandService(overrideDispatcher, logger)
			return overrideExecutionCommand.Start(ctx, apptaskboard.StartExecutionRequest{
				BoardID:        boardID,
				RepositoryRoot: runtime.repositoryRoot,
				SourceBranch:   sourceBranch,
				MaxTasks:       maxTasks,
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
		func(ctx context.Context, redisURI string) ([]apptaskboard.IngestionWorkflow, error) {
			cleanRedisURI := strings.TrimSpace(redisURI)
			if cleanRedisURI == "" || cleanRedisURI == cfg.Redis.URI {
				return runtime.runtimeWorkflowService.ListWorkflows(ctx)
			}
			overrideCfg, overrideErr := queueasynq.NewConfig(cleanRedisURI)
			if overrideErr != nil {
				return nil, overrideErr
			}
			overrideRepo := queueasynq.NewRuntimeWorkflowRepository(overrideCfg, logger)
			defer overrideRepo.Close()
			overrideService := apptaskboard.NewRuntimeWorkflowService(overrideRepo, logger)
			return overrideService.ListWorkflows(ctx)
		},
		func(ctx context.Context, runID string, redisURI string) (*apptaskboard.IngestionWorkflow, error) {
			cleanRedisURI := strings.TrimSpace(redisURI)
			if cleanRedisURI == "" || cleanRedisURI == cfg.Redis.URI {
				return runtime.runtimeWorkflowService.GetWorkflowStatus(ctx, runID)
			}
			overrideCfg, overrideErr := queueasynq.NewConfig(cleanRedisURI)
			if overrideErr != nil {
				return nil, overrideErr
			}
			overrideRepo := queueasynq.NewRuntimeWorkflowRepository(overrideCfg, logger)
			defer overrideRepo.Close()
			overrideService := apptaskboard.NewRuntimeWorkflowService(overrideRepo, logger)
			return overrideService.GetWorkflowStatus(ctx, runID)
		},
		func(ctx context.Context, runID string, redisURI string) (string, error) {
			cleanRedisURI := strings.TrimSpace(redisURI)
			service := runtime.runtimeWorkflowService
			if cleanRedisURI != "" && cleanRedisURI != cfg.Redis.URI {
				overrideCfg, overrideErr := queueasynq.NewConfig(cleanRedisURI)
				if overrideErr != nil {
					return "", overrideErr
				}
				overrideRepo := queueasynq.NewRuntimeWorkflowRepository(overrideCfg, logger)
				defer overrideRepo.Close()
				service = apptaskboard.NewRuntimeWorkflowService(overrideRepo, logger)
			}
			result, cancelErr := service.CancelWorkflow(ctx, runID)
			if cancelErr != nil {
				return "", cancelErr
			}
			if result.MatchedTasks == 0 {
				return fmt.Sprintf("No runtime tasks found for run %s", strings.TrimSpace(runID)), nil
			}
			return fmt.Sprintf("Canceled workflow %s: canceled=%d signaled_active=%d uncancelable=%d matched=%d", result.RunID, result.CanceledTasks, result.SignaledActive, result.UncancelableTasks, result.MatchedTasks), nil
		},
		runtime.authService.Status,
		runtime.authService.Authenticate,
		runtime.authService.KillOrphanedProcesses,
		runtime.repositoryRoot,
		defaultSourceBranch,
		cfg.Redis.URI,
		effectiveMaxAgents,
	)

	return runtime, nil
}

func (runtime *Runtime) Run() error {
	entry := runtime.entry().WithField("event", "core.runtime.run")
	entry.Info("starting runtime")
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
			entry.WithError(err).Error("worker exited with error")
			runtime.ui.Stop()
			_ = waitUI()
			return fmt.Errorf("asynq worker stopped: %w", err)
		}
		entry.Info("worker exited; stopping ui")
		runtime.ui.Stop()
		return waitUI()

	case err := <-uiErr:
		entry.WithError(err).Warn("ui exited; stopping worker")
		cancelWorker()
		workerStopErr := waitWorker()
		if err != nil {
			return err
		}
		return workerStopErr

	case <-sigCtx.Done():
		entry.WithField("signal", sigCtx.Err()).Info("shutdown signal received")
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

func (runtime *Runtime) entry() *logrus.Entry {
	if runtime == nil || runtime.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(runtime.logger)
}
