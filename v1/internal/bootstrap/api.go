package bootstrap

import (
	applicationsupervisor "agentic-orchestrator/internal/application/supervisor"
	"agentic-orchestrator/internal/application/taskengine"
	domainsupervisor "agentic-orchestrator/internal/domain/supervisor"
	postgresdb "agentic-orchestrator/internal/infrastructure/database/postgres"
	"agentic-orchestrator/internal/infrastructure/healthcheck"
	"agentic-orchestrator/internal/infrastructure/observability"
	asynqengine "agentic-orchestrator/internal/infrastructure/queue/asynq"
	infrasupervisorpostgres "agentic-orchestrator/internal/infrastructure/supervisor/postgres"
	infrataskenginepostgres "agentic-orchestrator/internal/infrastructure/taskengine/postgres"
	"agentic-orchestrator/internal/interface/graphql/graph"
	"agentic-orchestrator/internal/interface/graphql/resolvers"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"strings"
	"syscall"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
)

type APIApp struct {
	config                APIConfig
	httpServer            *http.Server
	observabilityPlatform *observability.Platform
	healthPlatform        *healthcheck.Platform
	taskScheduler         *taskengine.Scheduler
	taskEnginePlatform    *asynqengine.Platform
	databaseClient        *postgresdb.Client
}

func InitAPI() (*APIApp, error) {
	config, err := LoadAPIConfigFromEnv()
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
	supervisorEventStore, err := infrasupervisorpostgres.NewEventStore(databaseClient.DB())
	if err != nil {
		return nil, fmt.Errorf("init postgres supervisor event store: %w", err)
	}
	supervisorService, err := applicationsupervisor.NewService(supervisorEventStore, nil)
	if err != nil {
		return nil, fmt.Errorf("init supervisor service: %w", err)
	}

	resolver := resolvers.NewResolver(taskScheduler, supervisorService)
	server := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
	server.AddTransport(transport.Options{})
	server.AddTransport(transport.GET{})
	server.AddTransport(transport.POST{})
	server.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	server.Use(extension.Introspection{})
	server.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})

	mux := http.NewServeMux()
	if config.EnablePlayground {
		mux.Handle(config.PlaygroundPath, playground.Handler("GraphQL playground", config.GraphQLPath))
	}
	mux.Handle(config.GraphQLPath, server)
	mux.HandleFunc("/supervisor/history", supervisorHistoryHandler(supervisorService))
	mux.HandleFunc("/supervisor/issue-approval", supervisorIssueApprovalHandler(supervisorService))
	healthPlatform.Mount(mux)

	return &APIApp{
		config:                config,
		httpServer:            &http.Server{Addr: fmt.Sprintf(":%d", config.APIPort), Handler: mux},
		observabilityPlatform: observabilityPlatform,
		healthPlatform:        healthPlatform,
		taskScheduler:         taskScheduler,
		taskEnginePlatform:    taskEnginePlatform,
		databaseClient:        databaseClient,
	}, nil
}

func supervisorHistoryHandler(supervisorService *applicationsupervisor.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if supervisorService == nil {
			http.Error(writer, "supervisor service is not configured", http.StatusInternalServerError)
			return
		}
		runID := strings.TrimSpace(request.URL.Query().Get("run_id"))
		taskID := strings.TrimSpace(request.URL.Query().Get("task_id"))
		jobID := strings.TrimSpace(request.URL.Query().Get("job_id"))
		if runID == "" || taskID == "" || jobID == "" {
			http.Error(writer, "run_id, task_id, and job_id are required", http.StatusBadRequest)
			return
		}
		history, err := supervisorService.History(request.Context(), domainsupervisor.CorrelationIDs{RunID: runID, TaskID: taskID, JobID: jobID})
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(map[string]any{"decisions": history}); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type supervisorIssueApprovalRequest struct {
	RunID          string `json:"run_id"`
	TaskID         string `json:"task_id"`
	JobID          string `json:"job_id"`
	Source         string `json:"source"`
	IssueReference string `json:"issue_reference"`
	ApprovedBy     string `json:"approved_by"`
}

func supervisorIssueApprovalHandler(supervisorService *applicationsupervisor.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if supervisorService == nil {
			http.Error(writer, "supervisor service is not configured", http.StatusInternalServerError)
			return
		}
		var approvalRequest supervisorIssueApprovalRequest
		if err := json.NewDecoder(request.Body).Decode(&approvalRequest); err != nil {
			http.Error(writer, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		runID := strings.TrimSpace(approvalRequest.RunID)
		taskID := strings.TrimSpace(approvalRequest.TaskID)
		jobID := strings.TrimSpace(approvalRequest.JobID)
		issueReference := strings.TrimSpace(approvalRequest.IssueReference)
		approvedBy := strings.TrimSpace(approvalRequest.ApprovedBy)
		if runID == "" || taskID == "" || jobID == "" || issueReference == "" || approvedBy == "" {
			http.Error(writer, "run_id, task_id, job_id, issue_reference, and approved_by are required", http.StatusBadRequest)
			return
		}
		decision, err := supervisorService.OnIssueApproved(
			request.Context(),
			taskengine.CorrelationIDs{RunID: runID, TaskID: taskID, JobID: jobID},
			strings.TrimSpace(approvalRequest.Source),
			issueReference,
			approvedBy,
		)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(map[string]any{"decision": decision}); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
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
			"env":                 app.config.Environment,
			"task_engine_backend": app.config.TaskEngineBackend,
		}).Info("runtime starting")
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

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.config.ShutdownTimeout)
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
