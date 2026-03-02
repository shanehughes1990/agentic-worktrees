package worker

import (
	applicationagent "agentic-orchestrator/internal/application/agent"
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	applicationscm "agentic-orchestrator/internal/application/scm"
	applicationsupervisor "agentic-orchestrator/internal/application/supervisor"
	"agentic-orchestrator/internal/application/taskengine"
	applicationworker "agentic-orchestrator/internal/application/worker"
	"agentic-orchestrator/internal/core/shared/healthcheck"
	"agentic-orchestrator/internal/core/shared/observability"
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
	postgresdb "agentic-orchestrator/internal/infrastructure/database/postgres"
	asynqengine "agentic-orchestrator/internal/infrastructure/queue/asynq"
	infrascm "agentic-orchestrator/internal/infrastructure/scm"
	infrasupervisorpostgres "agentic-orchestrator/internal/infrastructure/supervisor/postgres"
	infrasupervisortaskengine "agentic-orchestrator/internal/infrastructure/supervisor/taskengine"
	infrataskenginepostgres "agentic-orchestrator/internal/infrastructure/taskengine/postgres"
	workerinterface "agentic-orchestrator/internal/interface/worker"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type WorkerApp struct {
	config                 Config
	httpServer             *http.Server
	observabilityPlatform  *observability.Platform
	healthPlatform         *healthcheck.Platform
	taskScheduler          *taskengine.Scheduler
	taskEnginePlatform     taskengine.Consumer
	databaseClient         *postgresdb.Client
	checkpointStore        taskengine.CheckpointStore
	executionJournal       taskengine.ExecutionJournal
	projectSetupRepository applicationcontrolplane.ProjectSetupRepository
	supervisorService      *applicationsupervisor.Service
	workerService          *applicationworker.Service
}

type workerJobRegistration struct {
	kind    taskengine.JobKind
	handler taskengine.Handler
	label   string
}

func New() (*WorkerApp, error) {
	config, err := LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}
	if err := ensureRuntimeFilesystem(config); err != nil {
		return nil, err
	}

	observabilityPlatform, healthPlatform, err := bootstrapPlatforms(context.Background(), config)
	if err != nil {
		return nil, err
	}

	taskScheduler, taskEnginePlatform, err := bootstrapTaskEngine(config, observabilityPlatform)
	if err != nil {
		return nil, err
	}

	databaseClient, err := postgresdb.Open(context.Background(), postgresdb.Config{DSN: config.App.DatabaseDSN}, observabilityPlatform.ServiceEntry())
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
	checkpointStore, err := infrataskenginepostgres.NewPostgresCheckpointStore(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init postgres checkpoint store: %w", err)
	}
	executionJournal, err := infrataskenginepostgres.NewPostgresExecutionJournal(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init postgres execution journal: %w", err)
	}
	supervisorEventStore, err := infrasupervisorpostgres.NewEventStore(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init postgres supervisor event store: %w", err)
	}
	supervisorService, err := applicationsupervisor.NewService(supervisorEventStore, nil)
	if err != nil {
		return nil, fmt.Errorf("init supervisor service: %w", err)
	}
	supervisorDispatcher, err := infrasupervisortaskengine.NewDispatcher(taskScheduler)
	if err != nil {
		return nil, fmt.Errorf("init supervisor task dispatcher: %w", err)
	}
	supervisorService.SetDispatcher(supervisorDispatcher)
	taskScheduler.SetAdmissionSignalSink(supervisorService)

	workerRegistry, err := infrataskenginepostgres.NewWorkerRegistry(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init worker registry: %w", err)
	}
	workerService, err := applicationworker.NewService(workerRegistry)
	if err != nil {
		return nil, fmt.Errorf("init worker service: %w", err)
	}
	projectSetupRepository, err := infrataskenginepostgres.NewProjectSetupRepository(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init project setup repository: %w", err)
	}

	mux := http.NewServeMux()
	healthPlatform.Mount(mux)

	return &WorkerApp{
		config:                 config,
		httpServer:             &http.Server{Addr: fmt.Sprintf(":%d", config.App.Port), Handler: mux},
		observabilityPlatform:  observabilityPlatform,
		healthPlatform:         healthPlatform,
		taskScheduler:          taskScheduler,
		taskEnginePlatform:     taskEnginePlatform,
		databaseClient:         databaseClient,
		checkpointStore:        checkpointStore,
		executionJournal:       executionJournal,
		projectSetupRepository: projectSetupRepository,
		supervisorService:      supervisorService,
		workerService:          workerService,
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
			"addr":                app.httpServer.Addr,
			"env":                 app.config.OTEL.ServiceEnvironment,
			"task_engine_backend": app.config.TaskEngineBackend,
		}).Info("worker runtime starting")
	}

	if app.projectSetupRepository == nil {
		return fmt.Errorf("project setup repository is not initialized")
	}
	repoLeaseManager, err := infrascm.NewPostgresRepoLeaseManager(app.databaseClient.DB())
	if err != nil {
		return fmt.Errorf("init postgres repo lease manager: %w", err)
	}

	buildGitHubAdapter := func(projectID string, repository applicationcontrolplane.ProjectRepository) (*infrascm.GitHubAdapter, error) {
		if strings.TrimSpace(repository.SCMProvider) != "github" {
			return nil, fmt.Errorf("unsupported scm provider %q", repository.SCMProvider)
		}
		if strings.TrimSpace(repository.SCMToken) == "" {
			return nil, fmt.Errorf("project repository scm_token is required")
		}
		githubAdapter, adapterErr := infrascm.NewGitHubAdapter(infrascm.GitHubAdapterConfig{
			RepoPath:         filepath.Join(app.config.ProjectPath(projectID), "repositories", "source"),
			WorktreeRootPath: app.config.ProjectsPath(),
		}, nil, infrascm.NewStaticTokenProvider(repository.SCMToken), infrascm.NewExecGitRunner())
		if adapterErr != nil {
			return nil, fmt.Errorf("init github scm adapter: %w", adapterErr)
		}
		return githubAdapter, nil
	}

	agentHandler, err := workerinterface.NewAgentWorkflowHandlerWithProjectSetup(
		app.projectSetupRepository,
		func(ctx context.Context, projectID string, repository applicationcontrolplane.ProjectRepository) (workerinterface.AgentRuntimeService, error) {
			_ = ctx
			githubAdapter, adapterErr := buildGitHubAdapter(projectID, repository)
			if adapterErr != nil {
				return nil, adapterErr
			}
			service, serviceErr := applicationagent.NewService(githubAdapter)
			if serviceErr != nil {
				return nil, fmt.Errorf("init agent service: %w", serviceErr)
			}
			return service, nil
		},
		app.checkpointStore,
		app.executionJournal,
		app.supervisorService,
	)
	if err != nil {
		return fmt.Errorf("create agent workflow handler: %w", err)
	}
	scmHandler, err := workerinterface.NewSCMWorkflowHandlerWithProjectSetup(
		app.projectSetupRepository,
		func(ctx context.Context, projectID string, repository applicationcontrolplane.ProjectRepository) (workerinterface.SCMRuntimeService, error) {
			_ = ctx
			githubAdapter, adapterErr := buildGitHubAdapter(projectID, repository)
			if adapterErr != nil {
				return nil, adapterErr
			}
			service, serviceErr := applicationscm.NewServiceWithLeaseManager(githubAdapter, repoLeaseManager)
			if serviceErr != nil {
				return nil, fmt.Errorf("init scm service: %w", serviceErr)
			}
			return service, nil
		},
		app.checkpointStore,
		app.executionJournal,
		app.supervisorService,
	)
	if err != nil {
		return fmt.Errorf("create scm workflow handler: %w", err)
	}
	hostname, hostnameErr := os.Hostname()
	if hostnameErr != nil {
		hostname = "unknown-host"
	}
	workerID := buildWorkerID(strings.TrimSpace(app.config.OTEL.ServiceName), hostname, app.config.App.Port)
	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	settings, err := waitForWorkerSettings(signalCtx, app.workerService, 2*time.Second)
	if err != nil {
		return fmt.Errorf("wait for worker settings: %w", err)
	}

	registrations := []workerJobRegistration{
		{kind: taskengine.JobKindAgentWorkflow, handler: agentHandler, label: "agent workflow"},
		{kind: taskengine.JobKindSCMWorkflow, handler: scmHandler, label: "scm workflow"},
	}
	if err := registerWorkerJobs(context.Background(), app.taskEnginePlatform, workerID, registrations); err != nil {
		return err
	}
	capabilities := workerCapabilities(registrations)
	registeredWorker, err := app.workerService.Register(context.Background(), workerID, capabilities, settings.HeartbeatInterval)
	if err != nil {
		return fmt.Errorf("register worker lifecycle: %w", err)
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

	serverErrCh := make(chan error, 1)
	go func() {
		err := app.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}
		serverErrCh <- nil
	}()

	heartbeatCtx, heartbeatCancel := context.WithCancel(signalCtx)
	defer heartbeatCancel()
	heartbeatErrCh := make(chan error, 1)
	go func() {
		err := runWorkerHeartbeat(heartbeatCtx, app.workerService, workerID, registeredWorker.Epoch, settings.HeartbeatInterval)
		if err == nil || errors.Is(err, context.Canceled) {
			return
		}
		heartbeatErrCh <- err
	}()

	shutdownReason := "shutdown signal"
	var runErr error

	select {
	case err := <-serverErrCh:
		if err != nil {
			shutdownReason = "health server error"
			runErr = fmt.Errorf("worker health server error: %w", err)
			if entry != nil {
				entry.WithError(err).WithField("runtime", "worker").Error("worker health server error")
			}
		} else {
			return nil
		}
	case err := <-heartbeatErrCh:
		shutdownReason = "application stopping"
		runErr = err
		if entry != nil {
			entry.WithError(err).WithField("runtime", "worker").Warn("worker heartbeat loop stopped")
		}
	case <-signalCtx.Done():
		if entry != nil {
			entry.WithField("runtime", "worker").Info("shutdown signal received")
		}
	}
	heartbeatCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.config.App.ShutdownTimeout)
	defer cancel()

	var shutdownErr error
	if app.workerService != nil {
		if _, err := app.workerService.RequestShutdown(shutdownCtx, workerID, registeredWorker.Epoch, shutdownReason); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("request shutdown state: %w", err))
		}
	}
	if err := app.httpServer.Shutdown(shutdownCtx); err != nil {
		if entry != nil {
			entry.WithError(err).WithField("runtime", "worker").Error("shutdown worker health server failed")
		}
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown worker health server: %w", err))
	}
	if app.taskEnginePlatform != nil {
		if err := app.taskEnginePlatform.Shutdown(shutdownCtx); err != nil {
			if entry != nil {
				entry.WithError(err).WithField("runtime", "worker").Error("shutdown task engine platform failed")
			}
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown task engine platform: %w", err))
		}
	}
	if app.workerService != nil {
		if _, err := app.workerService.Deregister(shutdownCtx, workerID, registeredWorker.Epoch, shutdownReason); err != nil {
			if _, forceErr := app.workerService.ForceDeregister(shutdownCtx, workerID, registeredWorker.Epoch, shutdownReason); forceErr != nil {
				shutdownErr = errors.Join(shutdownErr, fmt.Errorf("force deregister worker: %w", forceErr))
			}
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("deregister worker: %w", err))
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
	return errors.Join(runErr, shutdownErr)
}

func ensureRuntimeFilesystem(config Config) error {
	directories := []string{
		config.ApplicationRootPath(),
		config.ProjectsPath(),
		config.RepositorySourcePath(),
		config.LogsPath(),
	}
	for _, directory := range directories {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return fmt.Errorf("create runtime directory %q: %w", directory, err)
		}
	}
	return nil
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

func runWorkerHeartbeat(ctx context.Context, service *applicationworker.Service, workerID string, epoch int64, interval time.Duration) error {
	if service == nil {
		return fmt.Errorf("worker service is not initialized")
	}
	if interval <= 0 {
		return fmt.Errorf("heartbeat interval must be greater than zero")
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			worker, err := service.Heartbeat(ctx, workerID, epoch, interval)
			if err != nil {
				if errors.Is(err, applicationworker.ErrApplicationStopping) {
					return fmt.Errorf("%w: worker requested shutdown", applicationworker.ErrApplicationStopping)
				}
				return err
			}
			if worker != nil && (worker.DesiredState == domainrealtime.StateShutdownRequested || worker.DesiredState == domainrealtime.StateDraining || worker.DesiredState == domainrealtime.StateTerminated || worker.DesiredState == domainrealtime.StateDeregistered) {
				return fmt.Errorf("%w: desired state is %s", applicationworker.ErrApplicationStopping, worker.DesiredState)
			}
		}
	}
}

func workerCapabilities(registrations []workerJobRegistration) []taskengine.JobKind {
	capabilities := make([]taskengine.JobKind, 0, len(registrations))
	for _, registration := range registrations {
		capabilities = append(capabilities, registration.kind)
	}
	return capabilities
}

func waitForWorkerSettings(ctx context.Context, service *applicationworker.Service, retryInterval time.Duration) (domainrealtime.Settings, error) {
	if service == nil {
		return domainrealtime.Settings{}, fmt.Errorf("worker service is not initialized")
	}
	if retryInterval <= 0 {
		retryInterval = 2 * time.Second
	}
	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()
	for {
		settings, err := service.GetSettings(ctx)
		if err == nil {
			return settings, nil
		}
		if ctx.Err() != nil {
			return domainrealtime.Settings{}, ctx.Err()
		}
		<-ticker.C
	}
}

func buildWorkerID(serviceName string, hostname string, workerPort int) string {
	trimmedServiceName := strings.TrimSpace(serviceName)
	if trimmedServiceName == "" {
		trimmedServiceName = "worker"
	}
	trimmedHostname := strings.TrimSpace(hostname)
	if trimmedHostname == "" {
		trimmedHostname = "unknown-host"
	}
	return fmt.Sprintf("%s-%s-worker-%d", trimmedServiceName, trimmedHostname, workerPort)
}

func bootstrapPlatforms(ctx context.Context, config Config) (*observability.Platform, *healthcheck.Platform, error) {
	serviceVersion := strings.TrimSpace(os.Getenv("SERVICE_VERSION"))
	if serviceVersion == "" {
		serviceVersion = "development"
	}

	observabilityPlatform, err := observability.Bootstrap(ctx, observability.Config{
		ServiceName:  config.OTEL.ServiceName,
		Environment:  config.OTEL.ServiceEnvironment,
		Version:      serviceVersion,
		LogFormat:    observability.LogFormatText,
		LogLevel:     observability.LogLevelInfo,
		OTLPEndpoint: config.OTEL.OTLPEndpoint,
		OTLPHeaders:  parseOTLPHeaders(config.OTEL.OTLPHeaders),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("bootstrap observability: %w", err)
	}

	healthPlatform := healthcheck.Bootstrap(healthcheck.Config{
		LivenessPath:  config.Health.LivePath,
		ReadinessPath: config.Health.ReadyPath,
		Metadata: map[string]string{
			"service":     config.OTEL.ServiceName,
			"environment": config.OTEL.ServiceEnvironment,
			"version":     serviceVersion,
		},
	})

	return observabilityPlatform, healthPlatform, nil
}

func bootstrapTaskEngine(config Config, observabilityPlatform *observability.Platform) (*taskengine.Scheduler, *asynqengine.Platform, error) {
	if config.TaskEngineBackend != "asynq" {
		return nil, nil, fmt.Errorf("unsupported task engine backend: %s", config.TaskEngineBackend)
	}

	entry := observabilityPlatform.ServiceEntry()
	if entry != nil {
		entry = entry.WithFields(map[string]any{"component": "taskengine", "backend": config.TaskEngineBackend})
	}

	platform := asynqengine.NewPlatform(asynqengine.Config{
		RedisAddress:  config.TaskEngineRedisAddress,
		RedisPassword: config.TaskEngineRedisPassword,
		RedisDatabase: config.TaskEngineRedisDatabase,
		Concurrency:   config.TaskEngineConcurrency,
	}, entry)

	policies := taskengine.DefaultPolicies()
	scmPolicy := policies[taskengine.JobKindSCMWorkflow]
	scmPolicy.DefaultQueue = config.TaskEngineSCMQueue
	policies[taskengine.JobKindSCMWorkflow] = scmPolicy

	scheduler, err := taskengine.NewScheduler(platform, policies)
	if err != nil {
		return nil, nil, fmt.Errorf("bootstrap task engine scheduler: %w", err)
	}
	return scheduler, platform, nil
}

func parseOTLPHeaders(raw string) map[string]string {
	if strings.TrimSpace(raw) == "" {
		return map[string]string{}
	}
	headers := map[string]string{}
	parts := strings.Split(raw, ",")
	for _, part := range parts {
		pair := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(pair) != 2 {
			continue
		}
		key := strings.TrimSpace(pair[0])
		value := strings.TrimSpace(pair[1])
		if key == "" {
			continue
		}
		headers[key] = value
	}
	return headers
}
