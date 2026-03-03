package api

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	applicationstream "agentic-orchestrator/internal/application/stream"
	applicationsupervisor "agentic-orchestrator/internal/application/supervisor"
	"agentic-orchestrator/internal/application/taskengine"
	applicationworker "agentic-orchestrator/internal/application/worker"
	"agentic-orchestrator/internal/core/shared/healthcheck"
	"agentic-orchestrator/internal/core/shared/observability"
	domainstream "agentic-orchestrator/internal/domain/stream"
	infraagent "agentic-orchestrator/internal/infrastructure/agent"
	postgresdb "agentic-orchestrator/internal/infrastructure/database/postgres"
	asynqengine "agentic-orchestrator/internal/infrastructure/queue/asynq"
	infrastreampostgres "agentic-orchestrator/internal/infrastructure/stream/postgres"
	infrasupervisorpostgres "agentic-orchestrator/internal/infrastructure/supervisor/postgres"
	infrataskenginepostgres "agentic-orchestrator/internal/infrastructure/taskengine/postgres"
	"agentic-orchestrator/internal/interface/graphql/graph"
	"agentic-orchestrator/internal/interface/graphql/resolvers"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	workerCoordinator     *applicationworker.Coordinator
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
	supervisorEventStore, err := infrasupervisorpostgres.NewEventStore(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init postgres supervisor event store: %w", err)
	}
	supervisorService, err := applicationsupervisor.NewService(supervisorEventStore, nil)
	if err != nil {
		return nil, fmt.Errorf("init supervisor service: %w", err)
	}
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
	controlPlaneService, err := applicationcontrolplane.NewService(taskScheduler, supervisorService, controlPlaneQueryRepository, projectSetupRepository, taskEnginePlatform)
	if err != nil {
		return nil, fmt.Errorf("init control-plane service: %w", err)
	}
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
	if _, err := workerService.EnsureBaseSettings(context.Background(), applicationworker.DefaultSettings(time.Now().UTC())); err != nil {
		return nil, fmt.Errorf("ensure base worker settings: %w", err)
	}
	workerCoordinator, err := applicationworker.NewCoordinator(workerService, taskScheduler)
	if err != nil {
		return nil, fmt.Errorf("init worker coordinator: %w", err)
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

	resolver := resolvers.NewResolver(taskScheduler, supervisorService, controlPlaneService, streamService, workerService)
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
		workerCoordinator:     workerCoordinator,
	}, nil
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

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if app.workerCoordinator != nil {
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-signalCtx.Done():
					return
				case <-ticker.C:
					_ = app.workerCoordinator.ProbeAndEscalate(context.Background())
					_ = app.publishWorkerSummaryEvent(context.Background())
				}
			}
		}()
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

func (app *APIApp) publishWorkerSummaryEvent(ctx context.Context) error {
	if app == nil || app.workerService == nil || app.streamService == nil {
		return nil
	}
	workers, err := app.workerService.ListWorkers(ctx, 500)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	healthyWorkers := 0
	for _, worker := range workers {
		if strings.EqualFold(string(worker.State), "healthy") && worker.LeaseExpiresAt.After(now) {
			healthyWorkers++
		}
	}
	_, err = app.streamService.AppendAndPublish(ctx, domainstream.Event{
		EventID:    fmt.Sprintf("worker-summary-%d", now.UnixNano()),
		OccurredAt: now,
		Source:     domainstream.SourceWorker,
		EventType:  domainstream.EventWorkerHeartbeat,
		CorrelationIDs: domainstream.CorrelationIDs{
			RunID:         "worker",
			TaskID:        "summary",
			JobID:         "tick",
			CorrelationID: "worker-summary",
		},
		Payload: map[string]any{
			"total_workers":   len(workers),
			"healthy_workers": healthyWorkers,
			"observed_at":     now.Format(time.RFC3339Nano),
		},
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
