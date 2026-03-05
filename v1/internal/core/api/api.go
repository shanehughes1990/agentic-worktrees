package api

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	applicationrealtime "agentic-orchestrator/internal/application/realtime"
	applicationstream "agentic-orchestrator/internal/application/stream"
	"agentic-orchestrator/internal/application/taskengine"
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	applicationworker "agentic-orchestrator/internal/application/worker"
	"agentic-orchestrator/internal/core/shared/healthcheck"
	"agentic-orchestrator/internal/core/shared/observability"
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
	domainscm "agentic-orchestrator/internal/domain/scm"
	domainstream "agentic-orchestrator/internal/domain/stream"
	infraagent "agentic-orchestrator/internal/infrastructure/agent"
	postgresdb "agentic-orchestrator/internal/infrastructure/database/postgres"
	asynqengine "agentic-orchestrator/internal/infrastructure/queue/asynq"
	infrastructure_realtime "agentic-orchestrator/internal/infrastructure/realtime"
	infrascm "agentic-orchestrator/internal/infrastructure/scm"
	infrastreampostgres "agentic-orchestrator/internal/infrastructure/stream/postgres"
	infrataskenginepostgres "agentic-orchestrator/internal/infrastructure/taskengine/postgres"
	infratrackerpostgres "agentic-orchestrator/internal/infrastructure/tracker"
	"agentic-orchestrator/internal/interface/graphql/graph"
	"agentic-orchestrator/internal/interface/graphql/resolvers"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/vektah/gqlparser/v2/ast"
)

type APIApp struct {
	config                Config
	httpServer            *http.Server
	observabilityPlatform *observability.Platform
	healthPlatform        *healthcheck.Platform
	taskScheduler         *taskengine.Scheduler
	taskEnginePlatform    *asynqengine.APIPlatform
	databaseClient        *postgresdb.Client
	streamService         *applicationstream.Service
	sessionStateReader    *infraagent.SessionStateReader
	acpClient             *infraagent.ACPClient
	workerService         *applicationworker.Service
	realtimeTransport     domainrealtime.WorkerLifecycleTransport
	pendingHeartbeats     map[string]domainrealtime.HeartbeatRequest
	pendingHeartbeatsMu   sync.Mutex
}

type apiProjectRepositoryBranchCatalog struct {
	repositoryRootPath string
}

func (catalog *apiProjectRepositoryBranchCatalog) ListOriginBranches(ctx context.Context, projectID string, scm applicationcontrolplane.ProjectSCM, repository applicationcontrolplane.ProjectRepository) ([]string, string, error) {
	if strings.TrimSpace(scm.SCMProvider) != domainscm.ProviderGitHub {
		return nil, "", fmt.Errorf("unsupported scm provider %q", scm.SCMProvider)
	}
	if strings.TrimSpace(scm.SCMToken) == "" {
		return nil, "", fmt.Errorf("scm token is required")
	}
	owner, repositoryName, err := ownerRepositoryFromRepositoryURL(repository.RepositoryURL)
	if err != nil {
		return nil, "", err
	}
	rootPath := filepath.Join(strings.TrimSpace(catalog.repositoryRootPath), strings.TrimSpace(projectID), "repository-branch-catalog")
	if err := os.MkdirAll(rootPath, 0o755); err != nil {
		return nil, "", fmt.Errorf("create repository branch catalog path: %w", err)
	}
	repoPath := filepath.Join(rootPath, repositoryName)
	adapter, err := infrascm.NewGitHubAdapter(infrascm.GitHubAdapterConfig{
		RepoPath:           repoPath,
		RepositoryRootPath: rootPath,
		RepositoryURL:      strings.TrimSpace(repository.RepositoryURL),
	}, nil, infrascm.NewStaticTokenProvider(scm.SCMToken), infrascm.NewExecGitRunner())
	if err != nil {
		return nil, "", fmt.Errorf("init github adapter: %w", err)
	}
	branches, defaultBranch, err := adapter.ListOriginBranches(ctx, domainscm.Repository{Provider: domainscm.ProviderGitHub, Owner: owner, Name: repositoryName})
	if err != nil {
		return nil, "", err
	}
	return branches, defaultBranch, nil
}

func New() (*APIApp, error) {
	config, err := LoadConfigFromEnv()
	if err != nil {
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
	streamEventStore, err := infrastreampostgres.NewEventStore(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init postgres stream event store: %w", err)
	}
	streamService, err := applicationstream.NewService(streamEventStore)
	if err != nil {
		return nil, fmt.Errorf("init stream service: %w", err)
	}
	controlPlaneQueryRepository, err := infrataskenginepostgres.NewControlPlaneQueryRepository(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init control-plane query repository: %w", err)
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
	controlPlaneService, err := applicationcontrolplane.NewService(taskScheduler, controlPlaneQueryRepository, projectSetupRepository, taskEnginePlatform)
	if err != nil {
		return nil, fmt.Errorf("init control-plane service: %w", err)
	}
	controlPlaneService.SetProjectRepositoryBranchCatalog(&apiProjectRepositoryBranchCatalog{repositoryRootPath: filepath.Join(os.TempDir(), "agentic-orchestrator")})
	controlPlaneService.SetProjectDocumentRepository(projectDocumentRepository)
	if err := bootstrapRemoteStorage(config, controlPlaneService); err != nil {
		return nil, fmt.Errorf("bootstrap remote storage: %w", err)
	}
	workerRegistry, err := infrataskenginepostgres.NewWorkerRegistry(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init worker registry: %w", err)
	}
	workerService, err := applicationworker.NewService(workerRegistry)
	if err != nil {
		return nil, fmt.Errorf("init worker service: %w", err)
	}
	trackerStore, err := infratrackerpostgres.NewPostgresBoardStore(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init tracker board store: %w", err)
	}
	trackerService, err := applicationtracker.NewTaskMutationService(trackerStore)
	if err != nil {
		return nil, fmt.Errorf("init tracker service: %w", err)
	}
	realtimeTransport, err := infrastructure_realtime.NewPGNotifyTransport(databaseClient.DB(), config.App.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("init realtime transport: %w", err)
	}
	if _, err := workerService.EnsureBaseSettings(context.Background(), applicationworker.DefaultSettings(time.Now().UTC())); err != nil {
		return nil, fmt.Errorf("ensure base worker settings: %w", err)
	}
	sessionStateReader, err := infraagent.NewSessionStateReader(strings.TrimSpace(os.Getenv("API_COPILOT_SESSION_STATE_DIR")))
	if err != nil {
		return nil, fmt.Errorf("init session state reader: %w", err)
	}
	var acpClient *infraagent.ACPClient
	if strings.EqualFold(strings.TrimSpace(os.Getenv("API_COPILOT_ACP_ENABLED")), "true") {
		acpClient, err = infraagent.NewACPClient(strings.TrimSpace(os.Getenv("API_COPILOT_CLI_BINARY")), strings.TrimSpace(os.Getenv("API_COPILOT_ACP_WORKING_DIR")))
		if err != nil {
			return nil, fmt.Errorf("init acp client: %w", err)
		}
		streamService.SetPromptInjector(acpClient)
		streamService.SetHealthEvaluator(acpClient)
	}

	resolver := resolvers.NewResolver(taskScheduler, controlPlaneService, streamService, workerService, trackerService)
	server := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
	server.AddTransport(transport.Options{})
	server.AddTransport(transport.GET{})
	server.AddTransport(transport.POST{})
	server.AddTransport(transport.Websocket{Upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}})
	server.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	server.Use(extension.Introspection{})
	server.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})

	mux := http.NewServeMux()
	if config.API.EnablePlayground {
		mux.Handle(config.API.PlaygroundPath, playground.Handler("GraphQL playground", config.API.GraphQLPath))
	}
	mux.Handle(config.API.GraphQLPath, server)
	healthPlatform.Mount(mux)

	return &APIApp{
		config:                config,
		httpServer:            &http.Server{Addr: fmt.Sprintf(":%d", config.App.Port), Handler: mux},
		observabilityPlatform: observabilityPlatform,
		healthPlatform:        healthPlatform,
		taskScheduler:         taskScheduler,
		taskEnginePlatform:    taskEnginePlatform,
		databaseClient:        databaseClient,
		streamService:         streamService,
		sessionStateReader:    sessionStateReader,
		acpClient:             acpClient,
		workerService:         workerService,
		realtimeTransport:     realtimeTransport,
		pendingHeartbeats:     map[string]domainrealtime.HeartbeatRequest{},
	}, nil
}

func ownerRepositoryFromRepositoryURL(repositoryURL string) (string, string, error) {
	trimmedURL := strings.TrimSpace(repositoryURL)
	if trimmedURL == "" {
		return "", "", fmt.Errorf("repository_url is required")
	}
	parsedURL, err := url.Parse(trimmedURL)
	if err != nil {
		return "", "", fmt.Errorf("parse repository_url: %w", err)
	}
	pathParts := strings.Split(strings.Trim(strings.TrimSpace(parsedURL.Path), "/"), "/")
	if len(pathParts) < 2 {
		return "", "", fmt.Errorf("repository_url %q is missing owner/repository segments", trimmedURL)
	}
	owner := strings.TrimSpace(pathParts[len(pathParts)-2])
	repositoryName := strings.TrimSpace(strings.TrimSuffix(pathParts[len(pathParts)-1], ".git"))
	if owner == "" || repositoryName == "" {
		return "", "", fmt.Errorf("repository_url %q has invalid owner/repository", trimmedURL)
	}
	return owner, repositoryName, nil
}

func (app *APIApp) Run() error {
	if app == nil || app.httpServer == nil {
		return errors.New("api app is not initialized")
	}

	entry := app.observabilityPlatform.ServiceEntry()
	if entry != nil {
		entry.WithFields(map[string]any{
			"runtime":             "api",
			"addr":                app.httpServer.Addr,
			"env":                 app.config.OTEL.ServiceEnvironment,
			"task_engine_backend": "asynq",
		}).Info("runtime starting")
	}
	if app.acpClient != nil && app.streamService != nil {
		if err := app.acpClient.Start(context.Background(), func(ctx context.Context, event domainstream.Event) error {
			_, appendErr := app.streamService.AppendAndPublish(ctx, event)
			return appendErr
		}); err != nil {
			return fmt.Errorf("start copilot acp client: %w", err)
		}
	}
	if app.realtimeTransport == nil {
		return fmt.Errorf("realtime transport is not configured")
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go app.runRegistrationSubmissionListener(signalCtx)
	go app.runRegistrationStartupCatchup(signalCtx)
	go app.runHeartbeatResponseListener(signalCtx)
	go app.runHeartbeatRequestLoop(signalCtx)
	go app.runWorkerSessionStreamPublisher(signalCtx)

	serverErrCh := make(chan error, 1)
	go func() {
		err := app.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}
		serverErrCh <- nil
	}()

	select {
	case err := <-serverErrCh:
		if err != nil {
			if entry != nil {
				entry.WithError(err).WithField("runtime", "api").Error("runtime server error")
			}
			return fmt.Errorf("runtime server error: %w", err)
		}
		return nil
	case <-signalCtx.Done():
		if entry != nil {
			entry.WithField("runtime", "api").Info("shutdown signal received")
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.config.App.ShutdownTimeout)
	defer cancel()

	var shutdownErr error
	if err := app.httpServer.Shutdown(shutdownCtx); err != nil {
		if entry != nil {
			entry.WithError(err).WithField("runtime", "api").Error("shutdown http server failed")
		}
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown http server: %w", err))
	}
	if app.taskEnginePlatform != nil {
		if err := app.taskEnginePlatform.Shutdown(shutdownCtx); err != nil {
			if entry != nil {
				entry.WithError(err).WithField("runtime", "api").Error("shutdown task engine platform failed")
			}
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown task engine platform: %w", err))
		}
	}
	if app.acpClient != nil {
		if err := app.acpClient.Shutdown(); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown acp client: %w", err))
		}
	}
	if app.databaseClient != nil {
		if err := app.databaseClient.Close(); err != nil {
			if entry != nil {
				entry.WithError(err).WithField("runtime", "api").Error("shutdown postgres client failed")
			}
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown postgres client: %w", err))
		}
	}
	if err := app.healthPlatform.Shutdown(shutdownCtx); err != nil {
		if entry != nil {
			entry.WithError(err).WithField("runtime", "api").Error("shutdown health platform failed")
		}
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown health platform: %w", err))
	}
	if err := app.observabilityPlatform.Shutdown(shutdownCtx); err != nil {
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown observability platform: %w", err))
	}
	return shutdownErr
}

func (app *APIApp) runRegistrationStartupCatchup(ctx context.Context) {
	if app == nil || app.workerService == nil || app.realtimeTransport == nil {
		return
	}
	pending, err := app.workerService.ListPendingRegistrationSubmissions(ctx, 500)
	if err != nil {
		return
	}
	for _, submission := range pending {
		_ = app.processRegistrationSubmission(ctx, submission)
	}
}

func (app *APIApp) runRegistrationSubmissionListener(ctx context.Context) {
	if app == nil || app.workerService == nil || app.realtimeTransport == nil {
		return
	}
	_ = app.realtimeTransport.ListenRegistrationSubmissions(ctx, func(event domainrealtime.RegistrationSubmissionEvent) error {
		submission := domainrealtime.RegistrationSubmission{
			SubmissionID: event.SubmissionID,
			WorkerID:     event.WorkerID,
			RequestedAt:  event.RequestedAt,
			ExpiresAt:    event.ExpiresAt,
			Status:       domainrealtime.RegistrationStatusPending,
			Capabilities: event.Capabilities,
		}
		return app.processRegistrationSubmission(ctx, submission)
	})
}

func (app *APIApp) processRegistrationSubmission(ctx context.Context, submission domainrealtime.RegistrationSubmission) error {
	if app == nil || app.workerService == nil || app.realtimeTransport == nil {
		return fmt.Errorf("api app is not initialized")
	}
	if time.Now().UTC().After(submission.ExpiresAt) {
		resolved, err := app.workerService.ResolveRegistrationSubmission(ctx, submission.SubmissionID, false, []string{"registration request expired"})
		if err != nil {
			return err
		}
		return app.realtimeTransport.PublishRegistrationDecision(ctx, domainrealtime.RegistrationDecisionEvent{
			SubmissionID: resolved.SubmissionID,
			WorkerID:     resolved.WorkerID,
			Decision:     domainrealtime.RegistrationDecisionReject,
			Reasons:      resolved.RejectReasons,
			RespondedAt:  time.Now().UTC(),
		})
	}
	apiRequirements := domainrealtime.APIContractRequirements{Required: []domainrealtime.Capability{{
		Contract:     domainrealtime.ContractWorkerRegistry,
		Version:      "v1",
		SubContracts: []domainrealtime.SubContract{domainrealtime.SubContractHeartbeatRequest, domainrealtime.SubContractHeartbeatResponse},
	}}}
	workerAdvertisement := domainrealtime.WorkerContractAdvertisement{Implemented: submission.Capabilities}
	compatibilityErr := applicationrealtime.EnsureAPIToWorkerCompatibility(apiRequirements, workerAdvertisement)
	if compatibilityErr != nil {
		resolved, err := app.workerService.ResolveRegistrationSubmission(ctx, submission.SubmissionID, false, []string{compatibilityErr.Error()})
		if err != nil {
			return err
		}
		return app.realtimeTransport.PublishRegistrationDecision(ctx, domainrealtime.RegistrationDecisionEvent{
			SubmissionID: resolved.SubmissionID,
			WorkerID:     resolved.WorkerID,
			Decision:     domainrealtime.RegistrationDecisionReject,
			Reasons:      resolved.RejectReasons,
			RespondedAt:  time.Now().UTC(),
		})
	}
	resolved, err := app.workerService.ResolveRegistrationSubmission(ctx, submission.SubmissionID, true, nil)
	if err != nil {
		return err
	}
	_ = app.publishWorkerStreamEvent(ctx, domainstream.EventWorkerRegistrationAccepted, submission.SubmissionID, map[string]any{
		"submission_id": resolved.SubmissionID,
		"worker_id":     resolved.WorkerID,
		"decision":      string(domainrealtime.RegistrationDecisionAccept),
		"responded_at":  time.Now().UTC().Format(time.RFC3339Nano),
	})
	return app.realtimeTransport.PublishRegistrationDecision(ctx, domainrealtime.RegistrationDecisionEvent{
		SubmissionID: resolved.SubmissionID,
		WorkerID:     resolved.WorkerID,
		Decision:     domainrealtime.RegistrationDecisionAccept,
		RespondedAt:  time.Now().UTC(),
	})
}

func (app *APIApp) runHeartbeatRequestLoop(ctx context.Context) {
	if app == nil || app.workerService == nil || app.realtimeTransport == nil {
		return
	}
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			settings, err := app.workerService.GetSettings(ctx)
			if err != nil {
				continue
			}
			workers, err := app.workerService.ListWorkers(ctx, 500)
			if err != nil {
				continue
			}
			now := time.Now().UTC()
			for _, worker := range workers {
				if worker.State != domainrealtime.StateRegistered && worker.State != domainrealtime.StateHealthy {
					continue
				}
				request := domainrealtime.HeartbeatRequest{
					RequestID:   fmt.Sprintf("%s-%d", worker.WorkerID, now.UnixNano()),
					WorkerID:    worker.WorkerID,
					Epoch:       worker.Epoch,
					RequestedAt: now,
					DeadlineAt:  now.Add(settings.ResponseDeadline),
				}
				if err := app.realtimeTransport.PublishRequest(ctx, request); err != nil {
					continue
				}
				app.pendingHeartbeatsMu.Lock()
				app.pendingHeartbeats[request.RequestID] = request
				app.pendingHeartbeatsMu.Unlock()
			}
			app.enforceHeartbeatDeadlines(ctx)
		}
	}
}

func (app *APIApp) runHeartbeatResponseListener(ctx context.Context) {
	if app == nil || app.workerService == nil || app.realtimeTransport == nil {
		return
	}
	_ = app.realtimeTransport.ListenResponses(ctx, func(response domainrealtime.HeartbeatResponse) error {
		app.pendingHeartbeatsMu.Lock()
		request, exists := app.pendingHeartbeats[response.RequestID]
		if exists {
			delete(app.pendingHeartbeats, response.RequestID)
		}
		app.pendingHeartbeatsMu.Unlock()
		if !exists {
			return nil
		}
		settings, err := app.workerService.GetSettings(ctx)
		if err != nil {
			return nil
		}
		if _, err := app.workerService.RecordHeartbeat(ctx, response.WorkerID, response.Epoch, settings.HeartbeatInterval); err != nil {
			return nil
		}
		_ = app.publishWorkerStreamEvent(ctx, domainstream.EventWorkerHeartbeat, response.RequestID, map[string]any{
			"request_id":  response.RequestID,
			"worker_id":   response.WorkerID,
			"epoch":       response.Epoch,
			"responded_at": response.RespondedAt.UTC().Format(time.RFC3339Nano),
		})
		_ = request
		return nil
	})
}

func (app *APIApp) enforceHeartbeatDeadlines(ctx context.Context) {
	if app == nil || app.workerService == nil || app.realtimeTransport == nil {
		return
	}
	now := time.Now().UTC()
	expired := make([]domainrealtime.HeartbeatRequest, 0)
	app.pendingHeartbeatsMu.Lock()
	for key, request := range app.pendingHeartbeats {
		if now.After(request.DeadlineAt) {
			expired = append(expired, request)
			delete(app.pendingHeartbeats, key)
		}
	}
	app.pendingHeartbeatsMu.Unlock()
	for _, request := range expired {
		_, _ = app.workerService.Invalidate(ctx, request.WorkerID, request.Epoch)
		_, _ = app.workerService.ForceDeregister(ctx, request.WorkerID, request.Epoch)
		_ = app.publishWorkerStreamEvent(ctx, domainstream.EventWorkerInvalidated, request.RequestID, map[string]any{
			"request_id":  request.RequestID,
			"worker_id":   request.WorkerID,
			"epoch":       request.Epoch,
			"reason":      "heartbeat deadline missed",
			"deadline_at": request.DeadlineAt.UTC().Format(time.RFC3339Nano),
		})
		_ = app.realtimeTransport.PublishInvalidationIntent(ctx, domainrealtime.InvalidationIntent{
			WorkerID: request.WorkerID,
			Epoch:    request.Epoch,
			Reason:   "heartbeat deadline missed",
			IssuedAt: now,
		})
	}
}

func (app *APIApp) runWorkerSessionStreamPublisher(ctx context.Context) {
	if app == nil || app.workerService == nil || app.streamService == nil {
		return
	}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	previous := map[string]domainrealtime.Worker{}
	initialized := false
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			workers, err := app.workerService.ListWorkers(ctx, 500)
			if err != nil {
				continue
			}
			current := make(map[string]domainrealtime.Worker, len(workers))
			for _, worker := range workers {
				current[worker.WorkerID] = worker
			}

			if !initialized {
				previous = current
				initialized = true
				continue
			}

			for workerID, worker := range current {
				before, exists := previous[workerID]
				if !exists {
					_ = app.publishWorkerStreamEvent(ctx, domainstream.EventWorkerRegistrationAccepted, workerID, map[string]any{
						"worker_id":       worker.WorkerID,
						"epoch":           worker.Epoch,
						"state":           string(worker.State),
						"last_heartbeat":  worker.LastHeartbeat.UTC().Format(time.RFC3339Nano),
						"lease_expires_at": worker.LeaseExpiresAt.UTC().Format(time.RFC3339Nano),
						"updated_at":      worker.UpdatedAt.UTC().Format(time.RFC3339Nano),
					})
					continue
				}

				if before.Epoch != worker.Epoch ||
					before.State != worker.State ||
					!before.LastHeartbeat.Equal(worker.LastHeartbeat) ||
					!before.LeaseExpiresAt.Equal(worker.LeaseExpiresAt) {
					_ = app.publishWorkerStreamEvent(ctx, domainstream.EventWorkerHeartbeat, workerID, map[string]any{
						"worker_id":       worker.WorkerID,
						"epoch":           worker.Epoch,
						"state":           string(worker.State),
						"last_heartbeat":  worker.LastHeartbeat.UTC().Format(time.RFC3339Nano),
						"lease_expires_at": worker.LeaseExpiresAt.UTC().Format(time.RFC3339Nano),
						"updated_at":      worker.UpdatedAt.UTC().Format(time.RFC3339Nano),
					})
				}
			}

			for workerID, worker := range previous {
				if _, exists := current[workerID]; exists {
					continue
				}
				_ = app.publishWorkerStreamEvent(ctx, domainstream.EventWorkerInvalidated, workerID, map[string]any{
					"worker_id":       worker.WorkerID,
					"epoch":           worker.Epoch,
					"state":           string(worker.State),
					"reason":          "worker removed from registry",
					"last_heartbeat":  worker.LastHeartbeat.UTC().Format(time.RFC3339Nano),
					"lease_expires_at": worker.LeaseExpiresAt.UTC().Format(time.RFC3339Nano),
					"updated_at":      worker.UpdatedAt.UTC().Format(time.RFC3339Nano),
				})
			}

			previous = current
		}
	}
}

func (app *APIApp) publishWorkerStreamEvent(ctx context.Context, eventType domainstream.EventType, correlationID string, payload map[string]any) error {
	if app == nil || app.streamService == nil {
		return nil
	}
	resolvedCorrelationID := strings.TrimSpace(correlationID)
	if resolvedCorrelationID == "" {
		resolvedCorrelationID = fmt.Sprintf("worker-%d", time.Now().UTC().UnixNano())
	}
	_, err := app.streamService.AppendAndPublish(ctx, domainstream.Event{
		EventID:      fmt.Sprintf("%s-%d", strings.ReplaceAll(string(eventType), ".", "_"), time.Now().UTC().UnixNano()),
		OccurredAt:   time.Now().UTC(),
		Source:       domainstream.SourceWorker,
		EventType:    eventType,
		CorrelationIDs: domainstream.CorrelationIDs{CorrelationID: resolvedCorrelationID},
		Payload:      payload,
	})
	return err
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

func bootstrapTaskEngine(config Config, observabilityPlatform *observability.Platform) (*taskengine.Scheduler, *asynqengine.APIPlatform, error) {
	entry := observabilityPlatform.ServiceEntry()
	if entry != nil {
		entry = entry.WithFields(map[string]any{"component": "taskengine", "backend": "asynq"})
	}

	platform, err := asynqengine.NewAPIPlatform(asynqengine.APIConfig{
		RedisURL: config.App.RedisURL,
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

func bootstrapRemoteStorage(config Config, service *applicationcontrolplane.Service) error {
	if service == nil {
		return fmt.Errorf("control-plane service is required")
	}
	switch strings.ToLower(strings.TrimSpace(config.RemoteStorage.Type)) {
	case "gcs":
		service.SetProjectDocumentRootPrefix(config.RemoteStorage.BucketPrefix)
		service.SetProjectDocumentRemoteStorageType("gcs")
		service.SetProjectDocumentGoogleApplicationCredentialsPath(config.RemoteStorage.GoogleCloudStorage.ApplicationCredentialsFilePath)
		return nil
	default:
		return fmt.Errorf("unsupported remote storage type %q", config.RemoteStorage.Type)
	}
}
