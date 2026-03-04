package worker

import (
	applicationagent "agentic-orchestrator/internal/application/agent"
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	applicationingestion "agentic-orchestrator/internal/application/ingestion"
	applicationscm "agentic-orchestrator/internal/application/scm"
	applicationsupervisor "agentic-orchestrator/internal/application/supervisor"
	"agentic-orchestrator/internal/application/taskengine"
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	applicationworker "agentic-orchestrator/internal/application/worker"
	"agentic-orchestrator/internal/core/shared/healthcheck"
	"agentic-orchestrator/internal/core/shared/observability"
	"agentic-orchestrator/internal/domain/failures"
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
	domainscm "agentic-orchestrator/internal/domain/scm"
	infrastructurecdngoogle "agentic-orchestrator/internal/infrastructure/cdn/google"
	postgresdb "agentic-orchestrator/internal/infrastructure/database/postgres"
	infrastructurefilestoregcs "agentic-orchestrator/internal/infrastructure/filestore/gcs"
	infrastructureingestion "agentic-orchestrator/internal/infrastructure/ingestion"
	asynqengine "agentic-orchestrator/internal/infrastructure/queue/asynq"
	infrastructure_realtime "agentic-orchestrator/internal/infrastructure/realtime"
	infrascm "agentic-orchestrator/internal/infrastructure/scm"
	infrasupervisorpostgres "agentic-orchestrator/internal/infrastructure/supervisor/postgres"
	infrasupervisortaskengine "agentic-orchestrator/internal/infrastructure/supervisor/taskengine"
	infrataskenginepostgres "agentic-orchestrator/internal/infrastructure/taskengine/postgres"
	infratrackerpostgres "agentic-orchestrator/internal/infrastructure/tracker"
	workerinterface "agentic-orchestrator/internal/interface/worker"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
)

type WorkerApp struct {
	config                    Config
	httpServer                *http.Server
	observabilityPlatform     *observability.Platform
	healthPlatform            *healthcheck.Platform
	taskScheduler             *taskengine.Scheduler
	taskEnginePlatform        taskengine.Consumer
	databaseClient            *postgresdb.Client
	checkpointStore           taskengine.CheckpointStore
	executionJournal          taskengine.ExecutionJournal
	projectSetupRepository    applicationcontrolplane.ProjectSetupRepository
	projectDocumentRepository applicationcontrolplane.ProjectDocumentRepository
	supervisorService         *applicationsupervisor.Service
	workerService             *applicationworker.Service
	realtimeTransport         domainrealtime.WorkerLifecycleTransport
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
	realtimeTransport, err := infrastructure_realtime.NewPGNotifyTransport(databaseClient.DB(), config.App.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("init realtime transport: %w", err)
	}
	scmTokenCrypto, err := infrataskenginepostgres.NewSCMTokenCrypto(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init scm token crypto: %w", err)
	}
	projectSetupRepository, err := infrataskenginepostgres.NewProjectSetupRepository(databaseClient.DB(), scmTokenCrypto)
	if err != nil {
		return nil, fmt.Errorf("init project setup repository: %w", err)
	}
	projectDocumentRepository, err := infrataskenginepostgres.NewProjectDocumentRepository(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init project document repository: %w", err)
	}
	if err := projectSetupRepository.MigrateLegacySCMTokensToEncrypted(context.Background()); err != nil {
		return nil, fmt.Errorf("migrate legacy scm tokens: %w", err)
	}

	mux := http.NewServeMux()
	healthPlatform.Mount(mux)

	return &WorkerApp{
		config:                    config,
		httpServer:                &http.Server{Addr: fmt.Sprintf(":%d", config.App.Port), Handler: mux},
		observabilityPlatform:     observabilityPlatform,
		healthPlatform:            healthPlatform,
		taskScheduler:             taskScheduler,
		taskEnginePlatform:        taskEnginePlatform,
		databaseClient:            databaseClient,
		checkpointStore:           checkpointStore,
		executionJournal:          executionJournal,
		projectSetupRepository:    projectSetupRepository,
		projectDocumentRepository: projectDocumentRepository,
		supervisorService:         supervisorService,
		workerService:             workerService,
		realtimeTransport:         realtimeTransport,
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
			"task_engine_backend": "asynq",
		}).Info("worker runtime starting")
	}

	if app.projectSetupRepository == nil {
		return fmt.Errorf("project setup repository is not initialized")
	}
	repoLeaseManager, err := infrascm.NewPostgresRepoLeaseManager(app.databaseClient.DB())
	if err != nil {
		return fmt.Errorf("init postgres repo lease manager: %w", err)
	}

	buildGitHubAdapter := func(projectID string, scm applicationcontrolplane.ProjectSCM, repositoryID string, repositoryURL string) (*infrascm.GitHubAdapter, error) {
		if strings.TrimSpace(scm.SCMProvider) != "github" {
			return nil, fmt.Errorf("unsupported scm provider %q", scm.SCMProvider)
		}
		if strings.TrimSpace(scm.SCMToken) == "" {
			return nil, fmt.Errorf("project scm_token is required")
		}
		_, repositoryName, parseErr := ownerRepositoryFromRepositoryURL(repositoryURL)
		if parseErr != nil {
			repositoryName = strings.TrimSpace(repositoryID)
		}
		if strings.TrimSpace(repositoryName) == "" {
			repositoryName = "repository"
		}
		repoPath := filepath.Join(app.config.ProjectPath(projectID), "repositories", repositoryName)
		if err := os.MkdirAll(repoPath, 0o755); err != nil {
			return nil, fmt.Errorf("create project repository path: %w", err)
		}
		repositoryRootPath := app.config.ProjectsPath()
		if err := os.MkdirAll(repositoryRootPath, 0o755); err != nil {
			return nil, fmt.Errorf("create project repository root path: %w", err)
		}
		githubAdapter, adapterErr := infrascm.NewGitHubAdapter(infrascm.GitHubAdapterConfig{
			RepoPath:         repoPath,
			RepositoryRootPath: repositoryRootPath,
			RepositoryURL:    strings.TrimSpace(repositoryURL),
		}, nil, infrascm.NewStaticTokenProvider(scm.SCMToken), infrascm.NewExecGitRunner())
		if adapterErr != nil {
			return nil, fmt.Errorf("init github scm adapter: %w", adapterErr)
		}
		return githubAdapter, nil
	}
	credentialBytes, err := os.ReadFile(app.config.RemoteStorage.GoogleCloudStorage.ApplicationCredentialsFilePath)
	if err != nil {
		return fmt.Errorf("read google application credentials: %w", err)
	}
	documentStore, err := infrastructurefilestoregcs.NewStore(context.Background(), infrastructurefilestoregcs.Config{
		ProjectID:          app.config.RemoteStorage.GoogleCloudStorage.ProjectID,
		Bucket:             app.config.RemoteStorage.GoogleCloudStorage.Bucket,
		ServiceAccountJSON: string(credentialBytes),
		RootPrefix:         app.config.RemoteStorage.BucketPrefix,
	})
	if err != nil {
		return fmt.Errorf("init gcs document store: %w", err)
	}

	trackerStore, err := infratrackerpostgres.NewPostgresBoardStore(app.databaseClient.DB())
	if err != nil {
		return fmt.Errorf("init tracker board store: %w", err)
	}
	ingestionArtifactFetcher := applicationingestion.ArtifactFetcherFunc(func(ctx context.Context, objectPath string, destinationPath string) error {
		if err := documentStore.DownloadObjectToFile(ctx, strings.TrimSpace(objectPath), strings.TrimSpace(destinationPath)); err != nil {
			if failures.ClassOf(err) == failures.ClassUnknown {
				return failures.WrapTransient(err)
			}
			return err
		}
		return nil
	})
	ingestionAgentRunner, err := infrastructureingestion.NewCopilotCLIRunner("")
	if err != nil {
		return fmt.Errorf("init ingestion copilot cli runner: %w", err)
	}
	ingestionRepositorySynchronizer, err := infrastructureingestion.NewGitRepositorySynchronizer("", app.config.ProjectsPath())
	if err != nil {
		return fmt.Errorf("init ingestion repository synchronizer: %w", err)
	}
	ingestionService, err := applicationingestion.NewService(trackerStore, ingestionArtifactFetcher, ingestionAgentRunner, ingestionRepositorySynchronizer)
	if err != nil {
		return fmt.Errorf("init ingestion service: %w", err)
	}
	ingestionHandler, err := workerinterface.NewIngestionAgentHandler(ingestionService)
	if err != nil {
		return fmt.Errorf("create ingestion agent handler: %w", err)
	}
	taskMutationService, err := applicationtracker.NewTaskMutationService(trackerStore)
	if err != nil {
		return fmt.Errorf("init tracker task mutation service: %w", err)
	}

	agentHandler, err := workerinterface.NewAgentWorkflowHandlerWithProjectSetupAndTracker(
		app.projectSetupRepository,
		taskMutationService,
		func(ctx context.Context, projectID string, scm applicationcontrolplane.ProjectSCM, repository applicationcontrolplane.ProjectRepository) (workerinterface.AgentRuntimeService, error) {
			_ = ctx
			githubAdapter, adapterErr := buildGitHubAdapter(projectID, scm, repository.RepositoryID, repository.RepositoryURL)
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
		func(ctx context.Context, projectID string, scm applicationcontrolplane.ProjectSCM, repository applicationcontrolplane.ProjectRepository) (workerinterface.SCMRuntimeService, error) {
			_ = ctx
			githubAdapter, adapterErr := buildGitHubAdapter(projectID, scm, repository.RepositoryID, repository.RepositoryURL)
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
	controlPlaneQueryRepository, err := infrataskenginepostgres.NewControlPlaneQueryRepository(app.databaseClient.DB())
	if err != nil {
		return fmt.Errorf("init control-plane query repository: %w", err)
	}
	documentTaskService, err := applicationcontrolplane.NewService(app.taskScheduler, app.supervisorService, controlPlaneQueryRepository, app.projectSetupRepository, nil)
	if err != nil {
		return fmt.Errorf("init document task service: %w", err)
	}
	cdnSigner, err := infrastructurecdngoogle.NewSigner(infrastructurecdngoogle.Config{
		BaseURL:      app.config.RemoteStorage.GoogleCloudStorage.CDNBaseURL,
		KeyName:      app.config.RemoteStorage.GoogleCloudStorage.CDNKeyName,
		SignedKeyB64: app.config.RemoteStorage.GoogleCloudStorage.CDNKeyValue,
	})
	if err != nil {
		return fmt.Errorf("init google cdn signer: %w", err)
	}

	if cdnSigner == nil {
		return fmt.Errorf("init document delivery signer: nil signer for configured profile")
	}
	documentTaskService.SetProjectDocumentRepository(app.projectDocumentRepository)
	documentTaskService.SetProjectFileStore(documentStore)
	documentTaskService.SetProjectCDNSigner(cdnSigner)
	documentTaskService.SetProjectDocumentRootPrefix(app.config.RemoteStorage.BucketPrefix)
	documentPrepareHandler, err := workerinterface.NewProjectDocumentPrepareUploadHandler(documentTaskService)
	if err != nil {
		return fmt.Errorf("create project document prepare-upload handler: %w", err)
	}
	documentDeleteHandler, err := workerinterface.NewProjectDocumentDeleteHandler(documentTaskService)
	if err != nil {
		return fmt.Errorf("create project document delete handler: %w", err)
	}
	hostname, hostnameErr := os.Hostname()
	if hostnameErr != nil {
		hostname = "unknown-host"
	}
	workerID := buildWorkerID(strings.TrimSpace(app.config.OTEL.ServiceName), hostname, app.config.App.Port)
	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	reconcileProjectSourcesOnce := func(ctx context.Context) {
		reconcileCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()
		if err := app.reconcileProjectSourceArtifacts(reconcileCtx, workerID, repoLeaseManager, buildGitHubAdapter); err != nil {
			if entry != nil {
				entry.WithError(err).WithField("runtime", "worker").Warn("project source reconciliation run failed")
			}
		}
	}
	reconcileProjectSourcesOnce(signalCtx)

	reconcileCron := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	if _, err := reconcileCron.AddFunc("@every "+app.config.Worker.ProjectSourceReconcileInterval.String(), func() {
		reconcileProjectSourcesOnce(signalCtx)
	}); err != nil {
		return fmt.Errorf("schedule project source reconciler: %w", err)
	}
	reconcileCron.Start()
	defer func() {
		stopCtx := reconcileCron.Stop()
		select {
		case <-stopCtx.Done():
		case <-time.After(5 * time.Second):
		}
	}()

	settings, err := waitForWorkerSettings(signalCtx, app.workerService, 2*time.Second)
	if err != nil {
		return fmt.Errorf("wait for worker settings: %w", err)
	}

	registrations := []workerJobRegistration{
		{kind: taskengine.JobKindIngestionAgent, handler: ingestionHandler, label: "ingestion agent"},
		{kind: taskengine.JobKindAgentWorkflow, handler: agentHandler, label: "agent workflow"},
		{kind: taskengine.JobKindSCMWorkflow, handler: scmHandler, label: "scm workflow"},
		{kind: taskengine.JobKindProjectDocumentUploadPrepare, handler: documentPrepareHandler, label: "project document prepare upload"},
		{kind: taskengine.JobKindProjectDocumentDelete, handler: documentDeleteHandler, label: "project document delete"},
	}
	if err := registerWorkerJobs(context.Background(), app.taskEnginePlatform, workerID, registrations); err != nil {
		return err
	}
	capabilities := workerCapabilities(registrations)
	registrationSubmissionID := fmt.Sprintf("%s-%d", workerID, time.Now().UTC().UnixNano())
	registrationCapabilities := []domainrealtime.Capability{{
		Contract:     domainrealtime.ContractWorkerRegistry,
		Version:      "v1",
		SubContracts: []domainrealtime.SubContract{domainrealtime.SubContractHeartbeatRequest, domainrealtime.SubContractHeartbeatResponse},
	}}
	registrationSubmission := domainrealtime.RegistrationSubmission{
		SubmissionID: registrationSubmissionID,
		WorkerID:     workerID,
		RequestedAt:  time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(20 * time.Second),
		Status:       domainrealtime.RegistrationStatusPending,
		Capabilities: registrationCapabilities,
	}
	if _, err := app.workerService.CreateRegistrationSubmission(context.Background(), registrationSubmission); err != nil {
		return fmt.Errorf("create registration submission: %w", err)
	}
	if app.realtimeTransport == nil {
		return fmt.Errorf("realtime transport is not initialized")
	}
	if err := app.realtimeTransport.PublishRegistrationSubmission(context.Background(), domainrealtime.RegistrationSubmissionEvent{
		SubmissionID: registrationSubmission.SubmissionID,
		WorkerID:     registrationSubmission.WorkerID,
		RequestedAt:  registrationSubmission.RequestedAt,
		ExpiresAt:    registrationSubmission.ExpiresAt,
		Capabilities: registrationSubmission.Capabilities,
	}); err != nil {
		return fmt.Errorf("publish registration submission: %w", err)
	}
	decision, err := waitForRegistrationDecision(signalCtx, app.realtimeTransport, app.workerService, registrationSubmission.SubmissionID, workerID, 20*time.Second)
	if err != nil {
		if _, revokeErr := app.workerService.RevokeRegistrationSubmission(context.Background(), registrationSubmission.SubmissionID, "registration decision timeout"); revokeErr != nil {
			return fmt.Errorf("wait registration decision: %w (revoke failed: %v)", err, revokeErr)
		}
		return fmt.Errorf("wait registration decision: %w", err)
	}
	if decision.Decision != domainrealtime.RegistrationDecisionAccept {
		_, _ = app.workerService.ResolveRegistrationSubmission(context.Background(), registrationSubmission.SubmissionID, false, decision.Reasons)
		return fmt.Errorf("registration rejected: %s", strings.Join(decision.Reasons, "; "))
	}
	if _, err := app.workerService.ResolveRegistrationSubmission(context.Background(), registrationSubmission.SubmissionID, true, nil); err != nil {
		return fmt.Errorf("resolve registration submission accepted: %w", err)
	}
	registeredWorker, err := app.workerService.Register(context.Background(), workerID, capabilities, settings.HeartbeatInterval)
	if err != nil {
		return fmt.Errorf("register worker lifecycle: %w", err)
	}

	invalidationErrCh := make(chan error, 1)
	go func() {
		listenerErr := app.realtimeTransport.ListenRequests(signalCtx, func(request domainrealtime.HeartbeatRequest) error {
			if strings.TrimSpace(request.WorkerID) != workerID {
				return nil
			}
			if request.Epoch != registeredWorker.Epoch {
				return nil
			}
			return app.realtimeTransport.PublishResponse(context.Background(), domainrealtime.HeartbeatResponse{
				RequestID:   request.RequestID,
				WorkerID:    workerID,
				Epoch:       request.Epoch,
				ReceivedAt:  time.Now().UTC(),
				RespondedAt: time.Now().UTC(),
				Healthy:     true,
			})
		})
		if listenerErr != nil && signalCtx.Err() == nil {
			invalidationErrCh <- fmt.Errorf("listen heartbeat requests: %w", listenerErr)
		}
	}()
	go func() {
		listenerErr := app.realtimeTransport.ListenInvalidationIntents(signalCtx, func(intent domainrealtime.InvalidationIntent) error {
			if strings.TrimSpace(intent.WorkerID) != workerID || intent.Epoch != registeredWorker.Epoch {
				return nil
			}
			invalidationErrCh <- fmt.Errorf("worker invalidated by api: %s", intent.Reason)
			return nil
		})
		if listenerErr != nil && signalCtx.Err() == nil {
			invalidationErrCh <- fmt.Errorf("listen invalidation intents: %w", listenerErr)
		}
	}()
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

	var runErr error

	select {
	case err := <-serverErrCh:
		if err != nil {
			runErr = fmt.Errorf("worker health server error: %w", err)
			if entry != nil {
				entry.WithError(err).WithField("runtime", "worker").Error("worker health server error")
			}
		} else {
			return nil
		}
	case err := <-invalidationErrCh:
		runErr = err
		if entry != nil {
			entry.WithError(err).WithField("runtime", "worker").Error("worker invalidation triggered shutdown")
		}
	case <-signalCtx.Done():
		if entry != nil {
			entry.WithField("runtime", "worker").Info("shutdown signal received")
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.config.App.ShutdownTimeout)
	defer cancel()

	var shutdownErr error
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
		if _, err := app.workerService.Deregister(shutdownCtx, workerID, registeredWorker.Epoch); err != nil {
			if _, forceErr := app.workerService.ForceDeregister(shutdownCtx, workerID, registeredWorker.Epoch); forceErr != nil {
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

func waitForRegistrationDecision(ctx context.Context, transport domainrealtime.WorkerLifecycleTransport, service *applicationworker.Service, submissionID string, workerID string, timeout time.Duration) (domainrealtime.RegistrationDecisionEvent, error) {
	if transport == nil {
		return domainrealtime.RegistrationDecisionEvent{}, fmt.Errorf("realtime transport is required")
	}
	if service == nil {
		return domainrealtime.RegistrationDecisionEvent{}, fmt.Errorf("worker service is required")
	}
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	decisionCh := make(chan domainrealtime.RegistrationDecisionEvent, 1)
	errCh := make(chan error, 1)
	statusTicker := time.NewTicker(500 * time.Millisecond)
	defer statusTicker.Stop()
	listenCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		listenErr := transport.ListenRegistrationDecisions(listenCtx, func(event domainrealtime.RegistrationDecisionEvent) error {
			if event.SubmissionID != submissionID || strings.TrimSpace(event.WorkerID) != strings.TrimSpace(workerID) {
				return nil
			}
			select {
			case decisionCh <- event:
			default:
			}
			cancel()
			return nil
		})
		if listenErr != nil && listenCtx.Err() == nil {
			errCh <- listenErr
		}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case decision := <-decisionCh:
			return decision, nil
		case err := <-errCh:
			return domainrealtime.RegistrationDecisionEvent{}, err
		case <-statusTicker.C:
			submission, err := service.GetRegistrationSubmission(ctx, submissionID)
			if err != nil {
				if errors.Is(err, applicationworker.ErrSubmissionNotFound) {
					continue
				}
				return domainrealtime.RegistrationDecisionEvent{}, err
			}
			if strings.TrimSpace(submission.WorkerID) != strings.TrimSpace(workerID) {
				return domainrealtime.RegistrationDecisionEvent{}, fmt.Errorf("registration decision worker mismatch")
			}
			switch submission.Status {
			case domainrealtime.RegistrationStatusAccepted:
				respondedAt := submission.ResolvedAt
				if respondedAt.IsZero() {
					respondedAt = time.Now().UTC()
				}
				return domainrealtime.RegistrationDecisionEvent{
					SubmissionID: submission.SubmissionID,
					WorkerID:     submission.WorkerID,
					Decision:     domainrealtime.RegistrationDecisionAccept,
					RespondedAt:  respondedAt,
				}, nil
			case domainrealtime.RegistrationStatusRejected:
				respondedAt := submission.ResolvedAt
				if respondedAt.IsZero() {
					respondedAt = time.Now().UTC()
				}
				return domainrealtime.RegistrationDecisionEvent{
					SubmissionID: submission.SubmissionID,
					WorkerID:     submission.WorkerID,
					Decision:     domainrealtime.RegistrationDecisionReject,
					Reasons:      submission.RejectReasons,
					RespondedAt:  respondedAt,
				}, nil
			case domainrealtime.RegistrationStatusRevoked:
				return domainrealtime.RegistrationDecisionEvent{}, fmt.Errorf("registration submission revoked")
			}
		case <-timer.C:
			return domainrealtime.RegistrationDecisionEvent{}, fmt.Errorf("registration decision timeout")
		case <-ctx.Done():
			return domainrealtime.RegistrationDecisionEvent{}, ctx.Err()
		}
	}
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
		LogFormat:    parseAppLogFormat(config.App.LogType),
		LogLevel:     parseAppLogLevel(config.App.LogLevel),
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

func bootstrapTaskEngine(config Config, observabilityPlatform *observability.Platform) (*taskengine.Scheduler, *asynqengine.WorkerPlatform, error) {
	entry := observabilityPlatform.ServiceEntry()
	if entry != nil {
		entry = entry.WithFields(map[string]any{"component": "taskengine", "backend": "asynq"})
	}

	platform, err := asynqengine.NewWorkerPlatform(asynqengine.WorkerConfig{
		RedisURL:    config.App.RedisURL,
		Concurrency: config.Worker.TaskConcurrencyLimit,
	}, entry)
	if err != nil {
		return nil, nil, fmt.Errorf("bootstrap asynq platform: %w", err)
	}

	policies := taskengine.DefaultPolicies()

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

func parseAppLogLevel(raw string) observability.LogLevel {
	level := strings.ToLower(strings.TrimSpace(raw))
	switch level {
	case "debug":
		return observability.LogLevelDebug
	case "warn":
		return observability.LogLevelWarn
	case "error":
		return observability.LogLevelError
	default:
		return observability.LogLevelInfo
	}
}

func parseAppLogFormat(raw string) observability.LogFormat {
	format := strings.ToLower(strings.TrimSpace(raw))
	if format == "json" {
		return observability.LogFormatJSON
	}
	return observability.LogFormatText
}

func (app *WorkerApp) reconcileProjectSourceArtifacts(ctx context.Context, workerID string, repoLeaseManager applicationscm.RepoLeaseManager, buildGitHubAdapter func(projectID string, scm applicationcontrolplane.ProjectSCM, repositoryID string, repositoryURL string) (*infrascm.GitHubAdapter, error)) error {
	if app == nil || app.projectSetupRepository == nil {
		return fmt.Errorf("project setup repository is not initialized")
	}
	entry := app.observabilityPlatform.ServiceEntry()
	startedAt := time.Now().UTC()
	if entry != nil {
		entry.WithFields(map[string]any{
			"runtime":   "worker",
			"worker_id": strings.TrimSpace(workerID),
		}).Debug("project source reconciliation run started")
	}
	setups, err := app.projectSetupRepository.ListProjectSetups(ctx, 500)
	if err != nil {
		return fmt.Errorf("list project setups: %w", err)
	}
	processedRepositories := 0
	syncedRepositories := 0
	skippedRepositories := 0
	for _, setup := range setups {
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry != nil {
			entry.WithFields(map[string]any{
				"runtime":       "worker",
				"worker_id":     strings.TrimSpace(workerID),
				"project_id":    strings.TrimSpace(setup.ProjectID),
				"scm_count":     len(setup.SCMs),
				"repo_count":    len(setup.Repositories),
				"reconcile_run": startedAt.Format(time.RFC3339Nano),
			}).Debug("reconciling project repositories")
		}
		scmByID := make(map[string]applicationcontrolplane.ProjectSCM, len(setup.SCMs))
		for _, scm := range setup.SCMs {
			scmByID[strings.TrimSpace(scm.SCMID)] = scm
		}
		for _, repository := range setup.Repositories {
			processedRepositories++
			projectID := strings.TrimSpace(setup.ProjectID)
			repositoryID := strings.TrimSpace(repository.RepositoryID)
			scmConfig, ok := scmByID[strings.TrimSpace(repository.SCMID)]
			if !ok {
				skippedRepositories++
				if entry != nil {
					entry.WithFields(map[string]any{
						"runtime":       "worker",
						"worker_id":     strings.TrimSpace(workerID),
						"project_id":    projectID,
						"repository_id": repositoryID,
						"scm_id":        strings.TrimSpace(repository.SCMID),
					}).Debug("skipping repository reconcile due to missing scm configuration")
				}
				continue
			}
			adapter, adapterErr := buildGitHubAdapter(setup.ProjectID, scmConfig, repository.RepositoryID, repository.RepositoryURL)
			if adapterErr != nil {
				skippedRepositories++
				if entry != nil {
					entry.WithError(adapterErr).WithFields(map[string]any{
						"runtime":       "worker",
						"worker_id":     strings.TrimSpace(workerID),
						"project_id":    projectID,
						"repository_id": repositoryID,
						"provider":      strings.TrimSpace(scmConfig.SCMProvider),
					}).Debug("failed to build scm adapter for repository reconcile")
				}
				continue
			}
			scmService, serviceErr := applicationscm.NewServiceWithLeaseManager(adapter, repoLeaseManager)
			if serviceErr != nil {
				skippedRepositories++
				if entry != nil {
					entry.WithError(serviceErr).WithFields(map[string]any{
						"runtime":       "worker",
						"worker_id":     strings.TrimSpace(workerID),
						"project_id":    projectID,
						"repository_id": repositoryID,
					}).Debug("failed to initialize scm service for repository reconcile")
				}
				continue
			}
			owner, repositoryName, parseErr := ownerRepositoryFromRepositoryURL(strings.TrimSpace(repository.RepositoryURL))
			if parseErr != nil {
				skippedRepositories++
				if entry != nil {
					entry.WithError(parseErr).WithFields(map[string]any{
						"runtime":        "worker",
						"worker_id":      strings.TrimSpace(workerID),
						"project_id":     projectID,
						"repository_id":  repositoryID,
						"repository_url": strings.TrimSpace(repository.RepositoryURL),
					}).Debug("failed to parse repository url for reconcile")
				}
				continue
			}
			repo := domainscm.Repository{Provider: strings.TrimSpace(scmConfig.SCMProvider), Owner: owner, Name: repositoryName}
			jobID := fmt.Sprintf("project-source-reconcile:%s:%s:%d", projectID, repositoryID, time.Now().UTC().UnixNano())
			idempotencyKey := fmt.Sprintf("project-source-reconcile:%s:%s:%s", projectID, repositoryID, strings.TrimSpace(workerID))
			metadata := applicationscm.Metadata{CorrelationIDs: taskengine.CorrelationIDs{RunID: "project-source-reconcile", TaskID: "project-source-reconcile", JobID: jobID, ProjectID: projectID}, IdempotencyKey: idempotencyKey}
			state, stateErr := scmService.SourceState(ctx, applicationscm.SourceStateRequest{Repository: repo, Metadata: metadata})
			if stateErr != nil {
				skippedRepositories++
				if entry != nil {
					entry.WithError(stateErr).WithFields(map[string]any{
						"runtime":       "worker",
						"worker_id":     strings.TrimSpace(workerID),
						"project_id":    projectID,
						"repository_id": repositoryID,
						"owner":         owner,
						"repository":    repositoryName,
					}).Debug("failed to read source state for repository reconcile")
				}
				continue
			}
			repositoryPath := filepath.Join(projectID, "repositories", repositoryName)
			if _, syncErr := scmService.SyncRepository(ctx, applicationscm.SyncRepositoryRequest{Repository: repo, Path: repositoryPath, Metadata: metadata}); syncErr == nil {
				syncedRepositories++
				if entry != nil {
					entry.WithFields(map[string]any{
						"runtime":       "worker",
						"worker_id":     strings.TrimSpace(workerID),
						"project_id":    projectID,
						"repository_id": repositoryID,
						"owner":         owner,
						"repository":    repositoryName,
						"repository_path": repositoryPath,
					}).Debug("repository reconcile sync completed")
				}
				continue
			}
			if _, ensureErr := scmService.EnsureRepository(ctx, applicationscm.EnsureRepositoryRequest{
				Repository: repo,
				Spec: domainscm.RepositorySpec{
					BaseBranch:   state.DefaultBranch,
					TargetBranch: state.DefaultBranch,
					Path:         repositoryPath,
					SyncStrategy: domainscm.SyncStrategyMerge,
				},
				Metadata: metadata,
			}); ensureErr != nil {
				skippedRepositories++
				if entry != nil {
					entry.WithError(ensureErr).WithFields(map[string]any{
						"runtime":       "worker",
						"worker_id":     strings.TrimSpace(workerID),
						"project_id":    projectID,
						"repository_id": repositoryID,
						"owner":         owner,
						"repository":    repositoryName,
						"repository_path": repositoryPath,
					}).Debug("failed to ensure repository during repository reconcile")
				}
				continue
			}
			if _, syncErr := scmService.SyncRepository(ctx, applicationscm.SyncRepositoryRequest{Repository: repo, Path: repositoryPath, Metadata: metadata}); syncErr != nil {
				skippedRepositories++
				if entry != nil {
					entry.WithError(syncErr).WithFields(map[string]any{
						"runtime":       "worker",
						"worker_id":     strings.TrimSpace(workerID),
						"project_id":    projectID,
						"repository_id": repositoryID,
						"owner":         owner,
						"repository":    repositoryName,
						"repository_path": repositoryPath,
					}).Debug("failed to sync repository after ensure during repository reconcile")
				}
				continue
			}
			syncedRepositories++
			if entry != nil {
				entry.WithFields(map[string]any{
					"runtime":       "worker",
					"worker_id":     strings.TrimSpace(workerID),
					"project_id":    projectID,
					"repository_id": repositoryID,
					"owner":         owner,
					"repository":    repositoryName,
					"repository_path": repositoryPath,
				}).Debug("repository reconcile ensure+sync completed")
			}
		}
	}
	if entry != nil {
		entry.WithFields(map[string]any{
			"runtime":                "worker",
			"worker_id":              strings.TrimSpace(workerID),
			"projects_seen":          len(setups),
			"repositories_processed": processedRepositories,
			"repositories_synced":    syncedRepositories,
			"repositories_skipped":   skippedRepositories,
			"duration_ms":            time.Since(startedAt).Milliseconds(),
		}).Debug("project source reconciliation run completed")
	}
	return nil
}

func ownerRepositoryFromRepositoryURL(repositoryURL string) (string, string, error) {
	trimmedURL := strings.TrimSpace(repositoryURL)
	if trimmedURL == "" {
		return "", "", fmt.Errorf("project repository_url is required")
	}
	if strings.HasPrefix(trimmedURL, "git@") {
		parts := strings.SplitN(trimmedURL, ":", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("project repository_url %q is invalid", trimmedURL)
		}
		return ownerRepositoryFromPath(parts[1], trimmedURL)
	}
	parsedURL, err := url.Parse(trimmedURL)
	if err != nil || parsedURL.Host == "" {
		return "", "", fmt.Errorf("project repository_url %q is invalid", trimmedURL)
	}
	return ownerRepositoryFromPath(parsedURL.Path, trimmedURL)
}

func ownerRepositoryFromPath(pathValue string, rawURL string) (string, string, error) {
	trimmedPath := strings.Trim(strings.TrimSpace(pathValue), "/")
	parts := strings.Split(trimmedPath, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("project repository_url %q must include owner and repository", rawURL)
	}
	owner := strings.TrimSpace(parts[0])
	repository := strings.TrimSuffix(strings.TrimSpace(parts[1]), ".git")
	if owner == "" || repository == "" {
		return "", "", fmt.Errorf("project repository_url %q must include owner and repository", rawURL)
	}
	return owner, repository, nil
}
