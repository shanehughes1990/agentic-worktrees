package bootstrap

import (
	applicationagent "agentic-orchestrator/internal/application/agent"
	applicationscm "agentic-orchestrator/internal/application/scm"
	"agentic-orchestrator/internal/application/taskengine"
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"agentic-orchestrator/internal/infrastructure/healthcheck"
	"agentic-orchestrator/internal/infrastructure/observability"
	infrascm "agentic-orchestrator/internal/infrastructure/scm"
	infratracker "agentic-orchestrator/internal/infrastructure/tracker"
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
	taskEnginePlatform    taskengine.Consumer
	agentService          *applicationagent.Service
	scmService            *applicationscm.Service
	trackerService        *applicationtracker.Service
}

type workerJobRegistration struct {
	kind    taskengine.JobKind
	handler taskengine.Handler
	label   string
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
	repoLeaseManager := infrascm.NewInMemoryRepoLeaseManager()
	scmService, err := applicationscm.NewServiceWithLeaseManager(githubAdapter, repoLeaseManager)
	if err != nil {
		return nil, fmt.Errorf("init scm service: %w", err)
	}
	agentService, err := applicationagent.NewService(githubAdapter)
	if err != nil {
		return nil, fmt.Errorf("init agent service: %w", err)
	}
	localTrackerProvider, err := infratracker.NewLocalJSONProvider(config.TrackerLocalJSONBasePath)
	if err != nil {
		return nil, fmt.Errorf("init local tracker provider: %w", err)
	}
	trackerProviderRegistry, err := infratracker.NewProviderRegistry(map[domaintracker.SourceKind]applicationtracker.Provider{
		domaintracker.SourceKindLocalJSON: localTrackerProvider,
		domaintracker.SourceKindJira:      infratracker.NewJiraProvider(),
		domaintracker.SourceKindLinear:    infratracker.NewLinearProvider(),
	})
	if err != nil {
		return nil, fmt.Errorf("init tracker provider registry: %w", err)
	}
	trackerService, err := applicationtracker.NewService(trackerProviderRegistry)
	if err != nil {
		return nil, fmt.Errorf("init tracker service: %w", err)
	}

	return &WorkerApp{
		config:                config,
		observabilityPlatform: observabilityPlatform,
		healthPlatform:        healthPlatform,
		taskScheduler:         taskScheduler,
		taskEnginePlatform:    taskEnginePlatform,
		agentService:          agentService,
		scmService:            scmService,
		trackerService:        trackerService,
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

	ingestionHandler, err := workerinterface.NewIngestionAgentHandler(app.trackerService)
	if err != nil {
		return fmt.Errorf("create ingestion agent handler: %w", err)
	}
	agentHandler, err := workerinterface.NewAgentWorkflowHandler(app.agentService)
	if err != nil {
		return fmt.Errorf("create agent workflow handler: %w", err)
	}
	scmHandler, err := workerinterface.NewSCMWorkflowHandler(app.scmService)
	if err != nil {
		return fmt.Errorf("create scm workflow handler: %w", err)
	}
	if err := registerWorkerJobs(
		context.Background(),
		app.taskEnginePlatform,
		fmt.Sprintf("%s-worker-%d", strings.TrimSpace(app.config.ServiceName), app.config.WorkerPort),
		[]workerJobRegistration{
			{kind: taskengine.JobKindIngestionAgent, handler: ingestionHandler, label: "ingestion agent"},
			{kind: taskengine.JobKindAgentWorkflow, handler: agentHandler, label: "agent workflow"},
			{kind: taskengine.JobKindSCMWorkflow, handler: scmHandler, label: "scm workflow"},
		},
	); err != nil {
		return err
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

func registerWorkerJobs(ctx context.Context, consumer taskengine.Consumer, workerID string, registrations []workerJobRegistration) error {
	if consumer == nil {
		return fmt.Errorf("task engine platform is not initialized")
	}

	capabilities := make([]taskengine.WorkerCapability, 0, len(registrations))
	for _, registration := range registrations {
		if err := consumer.Register(registration.kind, registration.handler); err != nil {
			return fmt.Errorf("register %s handler: %w", registration.label, err)
		}
		capabilities = append(capabilities, taskengine.WorkerCapability{Kind: registration.kind})
	}

	advertiser, ok := consumer.(taskengine.WorkerCapabilityAdvertiser)
	if !ok {
		return nil
	}

	advertisement := taskengine.WorkerCapabilityAdvertisement{
		WorkerID:     workerID,
		Capabilities: capabilities,
	}
	if err := advertisement.Validate(); err != nil {
		return fmt.Errorf("advertise worker capabilities: %w", err)
	}
	if err := advertiser.Advertise(ctx, advertisement); err != nil {
		return fmt.Errorf("advertise worker capabilities: %w", err)
	}
	return nil
}
