package bootstrap

import (
	"agentic-orchestrator/internal/application/taskengine"
	"agentic-orchestrator/internal/infrastructure/healthcheck"
	"agentic-orchestrator/internal/infrastructure/observability"
	asynqengine "agentic-orchestrator/internal/infrastructure/queue/asynq"
	"agentic-orchestrator/internal/interface/graphql/graph"
	"agentic-orchestrator/internal/interface/graphql/resolvers"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
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
}

func InitAPI() (*APIApp, error) {
	config, err := LoadAPIConfigFromEnv()
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

	resolver := resolvers.NewResolver(taskScheduler)
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
	healthPlatform.Mount(mux)

	return &APIApp{
		config:                config,
		httpServer:            &http.Server{Addr: fmt.Sprintf(":%d", config.APIPort), Handler: mux},
		observabilityPlatform: observabilityPlatform,
		healthPlatform:        healthPlatform,
		taskScheduler:         taskScheduler,
		taskEnginePlatform:    taskEnginePlatform,
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
