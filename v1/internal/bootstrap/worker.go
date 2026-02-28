package bootstrap

import (
	applicationagent "agentic-orchestrator/internal/application/agent"
	applicationscm "agentic-orchestrator/internal/application/scm"
	"agentic-orchestrator/internal/application/taskengine"
	"agentic-orchestrator/internal/infrastructure/healthcheck"
	"agentic-orchestrator/internal/infrastructure/observability"
	asynqengine "agentic-orchestrator/internal/infrastructure/queue/asynq"
	infrascm "agentic-orchestrator/internal/infrastructure/scm"
	workerinterface "agentic-orchestrator/internal/interface/worker"
	"context"
	"errors"
	"fmt"
	"os/signal"
	"strings"
	"syscall"
)

type WorkerApp struct {
	config                WorkerConfig
	observabilityPlatform *observability.Platform
	healthPlatform        *healthcheck.Platform
	taskScheduler         *taskengine.Scheduler
	taskEnginePlatform    *asynqengine.Platform
	agentService          *applicationagent.Service
	scmService            *applicationscm.Service
}

func InitWorker() (*WorkerApp, error) {
	config, err := LoadWorkerConfigFromEnv()
	if err != nil {
		return nil, err
	}

	observabilityPlatform, healthPlatform, err := bootstrapPlatforms(context.Background(), config.BaseConfig)
	if err != nil {
		return nil, err
	}

	taskScheduler, taskEnginePlatform, err := bootstrapTaskEngine(config.BaseConfig, observabilityPlatform)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(config.SCMGitHubToken) == "" {
		return nil, fmt.Errorf("worker requires SCM_GITHUB_TOKEN for github scm execution")
	}

	githubAdapter, err := infrascm.NewGitHubAdapter(infrascm.GitHubAdapterConfig{APIBaseURL: config.SCMGitHubAPIBaseURL, RepoPath: config.SCMLocalRepositoryPath}, nil, infrascm.NewStaticTokenProvider(config.SCMGitHubToken), infrascm.NewExecGitRunner())
	if err != nil {
		return nil, fmt.Errorf("init github scm adapter: %w", err)
	}
	scmService, err := applicationscm.NewService(githubAdapter)
	if err != nil {
		return nil, fmt.Errorf("init scm service: %w", err)
	}
	agentService, err := applicationagent.NewService(githubAdapter)
	if err != nil {
		return nil, fmt.Errorf("init agent service: %w", err)
	}

	return &WorkerApp{
		config:                config,
		observabilityPlatform: observabilityPlatform,
		healthPlatform:        healthPlatform,
		taskScheduler:         taskScheduler,
		taskEnginePlatform:    taskEnginePlatform,
		agentService:          agentService,
		scmService:            scmService,
	}, nil
}

func (app *WorkerApp) Run() error {
	if app == nil {
		return errors.New("worker app is not initialized")
	}

	entry := app.observabilityPlatform.ServiceEntry()
	if entry != nil {
		entry.WithFields(map[string]any{
			"runtime":             "worker",
			"env":                 app.config.Environment,
			"task_engine_backend": app.config.TaskEngineBackend,
		}).Info("worker runtime starting")
	}

	ingestionHandler := workerinterface.NewIngestionAgentHandler()
	if err := app.taskEnginePlatform.Register(taskengine.JobKindIngestionAgent, ingestionHandler); err != nil {
		return fmt.Errorf("register ingestion agent handler: %w", err)
	}
	agentHandler, err := workerinterface.NewAgentWorkflowHandler(app.agentService)
	if err != nil {
		return fmt.Errorf("create agent workflow handler: %w", err)
	}
	if err := app.taskEnginePlatform.Register(taskengine.JobKindAgentWorkflow, agentHandler); err != nil {
		return fmt.Errorf("register agent workflow handler: %w", err)
	}
	scmHandler, err := workerinterface.NewSCMWorkflowHandler(app.scmService)
	if err != nil {
		return fmt.Errorf("create scm workflow handler: %w", err)
	}
	if err := app.taskEnginePlatform.Register(taskengine.JobKindSCMWorkflow, scmHandler); err != nil {
		return fmt.Errorf("register scm workflow handler: %w", err)
	}
	if err := app.taskEnginePlatform.Start(); err != nil {
		if entry != nil {
			entry.WithError(err).WithField("runtime", "worker").Error("failed to start task engine worker")
		}
		return fmt.Errorf("start task engine worker: %w", err)
	}
	if entry != nil {
		entry.WithField("runtime", "worker").Info("task engine worker started")
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-signalCtx.Done()

	if entry != nil {
		entry.WithField("runtime", "worker").Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.config.ShutdownTimeout)
	defer cancel()

	var shutdownErr error
	if app.taskEnginePlatform != nil {
		if err := app.taskEnginePlatform.Shutdown(shutdownCtx); err != nil {
			if entry != nil {
				entry.WithError(err).WithField("runtime", "worker").Error("shutdown task engine platform failed")
			}
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown task engine platform: %w", err))
		}
	}
	if err := app.healthPlatform.Shutdown(shutdownCtx); err != nil {
		if entry != nil {
			entry.WithError(err).WithField("runtime", "worker").Error("shutdown health platform failed")
		}
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown health platform: %w", err))
	}
	if err := app.observabilityPlatform.Shutdown(shutdownCtx); err != nil {
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown observability platform: %w", err))
	}
	return shutdownErr
}
