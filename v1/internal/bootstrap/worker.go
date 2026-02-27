package bootstrap

import (
	"agentic-orchestrator/internal/infrastructure/healthcheck"
	"agentic-orchestrator/internal/infrastructure/observability"
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"
)

type WorkerApp struct {
	config                WorkerConfig
	observabilityPlatform *observability.Platform
	healthPlatform        *healthcheck.Platform
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

	return &WorkerApp{
		config:                config,
		observabilityPlatform: observabilityPlatform,
		healthPlatform:        healthPlatform,
	}, nil
}

func (app *WorkerApp) Run() error {
	if app == nil {
		return errors.New("worker app is not initialized")
	}

	entry := app.observabilityPlatform.ServiceEntry()
	if entry != nil {
		entry.WithFields(map[string]any{
			"runtime": "worker",
			"env":     app.config.Environment,
		}).Info("worker noop runtime blocking until shutdown signal")
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
