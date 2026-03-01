package bootstrap

import (
	applicationagent "agentic-orchestrator/internal/application/agent"
	applicationscm "agentic-orchestrator/internal/application/scm"
	"agentic-orchestrator/internal/application/taskengine"
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	postgresdb "agentic-orchestrator/internal/infrastructure/database/postgres"
	"agentic-orchestrator/internal/infrastructure/healthcheck"
	"agentic-orchestrator/internal/infrastructure/observability"
	infrascm "agentic-orchestrator/internal/infrastructure/scm"
	infrataskenginepostgres "agentic-orchestrator/internal/infrastructure/taskengine/postgres"
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
	databaseClient        *postgresdb.Client
	workerRegistry        *infrataskenginepostgres.WorkerRegistry
	checkpointStore       taskengine.CheckpointStore
	executionJournal      taskengine.ExecutionJournal
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
	if err := ensureRuntimeFilesystem(config.BaseConfig); err != nil {
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

	databaseClient, err := postgresdb.Open(context.Background(), postgresdb.Config{DSN: config.DatabaseDSN}, observabilityPlatform.ServiceEntry())
	if err != nil {
		return nil, fmt.Errorf("init postgres client: %w", err)
	}
	admissionLedger, err := infrataskenginepostgres.NewAdmissionLedger(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init postgres admission ledger: %w", err)
	}
	taskScheduler.SetAdmissionLedger(admissionLedger)
	deadLetterAudit, err := infrataskenginepostgres.NewDeadLetterAudit(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init postgres dead-letter audit: %w", err)
	}
	taskEnginePlatform.SetDeadLetterAudit(deadLetterAudit)
	workerRegistry, err := infrataskenginepostgres.NewWorkerRegistry(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init postgres worker registry: %w", err)
	}
	checkpointStore, err := infrataskenginepostgres.NewPostgresCheckpointStore(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init postgres checkpoint store: %w", err)
	}
	executionJournal, err := infrataskenginepostgres.NewPostgresExecutionJournal(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init postgres execution journal: %w", err)
	}
	repoLeaseManager, err := infrascm.NewPostgresRepoLeaseManager(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init postgres repo lease manager: %w", err)
	}

	if strings.TrimSpace(config.SCMGitHubToken) == "" {
		return nil, fmt.Errorf("worker requires SCM_GITHUB_TOKEN for github scm execution")
	}

	githubAdapter, err := infrascm.NewGitHubAdapter(infrascm.GitHubAdapterConfig{
		APIBaseURL:       config.SCMGitHubAPIBaseURL,
		RepoPath:         config.RepositorySourcePath(),
		WorktreeRootPath: config.WorktreesPath(),
	}, nil, infrascm.NewStaticTokenProvider(config.SCMGitHubToken), infrascm.NewExecGitRunner())
	if err != nil {
		return nil, fmt.Errorf("init github scm adapter: %w", err)
	}
	scmService, err := applicationscm.NewServiceWithLeaseManager(githubAdapter, repoLeaseManager)
	if err != nil {
		return nil, fmt.Errorf("init scm service: %w", err)
	}
	agentService, err := applicationagent.NewService(githubAdapter)
	if err != nil {
		return nil, fmt.Errorf("init agent service: %w", err)
	}
	localTrackerProvider, err := infratracker.NewLocalJSONProvider(config.TrackerPath())
	if err != nil {
		return nil, fmt.Errorf("init local tracker provider: %w", err)
	}
	trackerSnapshotProvider, err := infratracker.NewPostgresBoardSnapshotProvider(databaseClient.DB(), localTrackerProvider)
	if err != nil {
		return nil, fmt.Errorf("init postgres tracker board snapshot provider: %w", err)
	}
	trackerProviderRegistry, err := infratracker.NewProviderRegistry(map[domaintracker.SourceKind]applicationtracker.Provider{
		domaintracker.SourceKindLocalJSON: trackerSnapshotProvider,
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
		databaseClient:        databaseClient,
		workerRegistry:        workerRegistry,
		checkpointStore:       checkpointStore,
		executionJournal:      executionJournal,
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
	agentHandler, err := workerinterface.NewAgentWorkflowHandlerWithReliability(app.agentService, app.checkpointStore, app.executionJournal)
	if err != nil {
		return fmt.Errorf("create agent workflow handler: %w", err)
	}
	scmHandler, err := workerinterface.NewSCMWorkflowHandlerWithReliability(app.scmService, app.checkpointStore, app.executionJournal)
	if err != nil {
		return fmt.Errorf("create scm workflow handler: %w", err)
	}
	workerID := fmt.Sprintf("%s-worker-%d", strings.TrimSpace(app.config.ServiceName), app.config.WorkerPort)
	registrations := []workerJobRegistration{
		{kind: taskengine.JobKindIngestionAgent, handler: ingestionHandler, label: "ingestion agent"},
		{kind: taskengine.JobKindAgentWorkflow, handler: agentHandler, label: "agent workflow"},
		{kind: taskengine.JobKindSCMWorkflow, handler: scmHandler, label: "scm workflow"},
	}
	if err := registerWorkerJobs(context.Background(), app.taskEnginePlatform, workerID, registrations); err != nil {
		return err
	}
	if app.workerRegistry != nil {
		capabilities := make([]taskengine.WorkerCapability, 0, len(registrations))
		for _, registration := range registrations {
			capabilities = append(capabilities, taskengine.WorkerCapability{Kind: registration.kind})
		}
		if err := app.workerRegistry.Upsert(context.Background(), taskengine.WorkerCapabilityAdvertisement{WorkerID: workerID, Capabilities: capabilities}); err != nil {
			return fmt.Errorf("persist worker capability advertisement: %w", err)
		}
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
	if app.databaseClient != nil {
		if err := app.databaseClient.Close(); err != nil {
			if entry != nil {
				entry.WithError(err).WithField("runtime", "worker").Error("shutdown postgres client failed")
			}
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown postgres client: %w", err))
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
