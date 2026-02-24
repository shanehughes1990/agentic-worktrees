package core

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/logging/logruslogger"
	queueasynq "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq"
	"github.com/shanehughes1990/agentic-worktrees/internal/interface/dashboard"
)

type Runtime struct {
	worker *queueasynq.Server
	ui     *dashboard.UI
}

func Init() (*Runtime, error) {
	cfg, err := LoadAppConfigFromEnv()
	if err != nil {
		return nil, err
	}

	logger, err := logruslogger.New(cfg.Logging.Format, cfg.Logging.Level)
	if err != nil {
		return nil, err
	}

	queueCfg, err := queueasynq.NewConfig(cfg.Redis.URI)
	if err != nil {
		return nil, err
	}
	queueCfg = queueCfg.WithLogger(logruslogger.NewAsynqAdapter(logger))

	return &Runtime{
		worker: queueasynq.NewServer(queueCfg),
		ui:     dashboard.New(),
	}, nil
}

func (runtime *Runtime) Run() error {
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	workerErr := make(chan error, 1)
	go func() {
		workerErr <- runtime.worker.Run(workerCtx, nil)
	}()

	uiErr := make(chan error, 1)
	go func() {
		uiErr <- runtime.ui.Run()
	}()

	sigCtx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	shutdownDeadline := 5 * time.Second

	waitWorker := func() error {
		timer := time.NewTimer(shutdownDeadline)
		defer timer.Stop()
		select {
		case err := <-workerErr:
			if err != nil && !errors.Is(err, asynq.ErrServerClosed) {
				return fmt.Errorf("asynq worker stopped: %w", err)
			}
			return nil
		case <-timer.C:
			return fmt.Errorf("shutdown timeout waiting for worker")
		}
	}

	waitUI := func() error {
		timer := time.NewTimer(shutdownDeadline)
		defer timer.Stop()
		select {
		case err := <-uiErr:
			return err
		case <-timer.C:
			return fmt.Errorf("shutdown timeout waiting for dashboard")
		}
	}

	select {
	case err := <-workerErr:
		if err != nil && !errors.Is(err, asynq.ErrServerClosed) {
			runtime.ui.Stop()
			_ = waitUI()
			return fmt.Errorf("asynq worker stopped: %w", err)
		}
		runtime.ui.Stop()
		return waitUI()

	case err := <-uiErr:
		cancelWorker()
		workerStopErr := waitWorker()
		if err != nil {
			return err
		}
		return workerStopErr

	case <-sigCtx.Done():
		cancelWorker()
		if err := waitWorker(); err != nil {
			runtime.ui.Stop()
			_ = waitUI()
			return err
		}
		runtime.ui.Stop()
		return waitUI()
	}
}
