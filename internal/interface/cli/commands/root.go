package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
	urcli "github.com/urfave/cli/v3"

	"github.com/shanehughes1990/agentic-worktrees/internal/application"
)

var (
	logLevel       = "info"
	logFormat      = "json"
	databaseDSN    = defaultDatabaseDSN()
	rootDirectory  = "docs"
	traversalDepth = 2
	boardPrompt    = "Create a task board from the provided documentation files."
	boardModel     = "gpt-5.3-codex"
	logger         *logrus.Logger
)

type EnvironmentBootstrapper interface {
	Bootstrap(logLevel string, logFormat string, databaseDSN string) (*logrus.Logger, error)
}

type GenerateTaskBoardRunner interface {
	Run(ctx context.Context, input application.GenerateTaskBoardInput) (string, error)
}

type WorkerRuntimeRunner interface {
	Run(ctx context.Context) error
}

func NewRootCommand(bootstrapper EnvironmentBootstrapper, runner GenerateTaskBoardRunner, workerRunner WorkerRuntimeRunner) *urcli.Command {
	return &urcli.Command{
		Name:  "cli",
		Usage: "bootstrap cli",
		Flags: []urcli.Flag{
			&urcli.StringFlag{Name: "LOG_LEVEL", Value: logLevel, Destination: &logLevel, Sources: urcli.EnvVars("LOG_LEVEL")},
			&urcli.StringFlag{Name: "LOG_FORMAT", Value: logFormat, Destination: &logFormat, Sources: urcli.EnvVars("LOG_FORMAT")},
			&urcli.StringFlag{Name: "DATABASE_DSN", Value: databaseDSN, Destination: &databaseDSN, Sources: urcli.EnvVars("DATABASE_DSN")},
		},
		Commands: []*urcli.Command{newGenerateTaskBoardCommand(runner), newRunWorkerCommand(workerRunner)},
		Before: func(hookCtx context.Context, _ *urcli.Command) (context.Context, error) {
			if bootstrapper == nil {
				return hookCtx, fmt.Errorf("bootstrapper cannot be nil")
			}
			configuredLogger, err := bootstrapper.Bootstrap(logLevel, logFormat, databaseDSN)
			if err != nil {
				return hookCtx, err
			}
			logger = configuredLogger
			return hookCtx, nil
		},
	}
}

func newGenerateTaskBoardCommand(runner GenerateTaskBoardRunner) *urcli.Command {
	return &urcli.Command{
		Name:  "generate-task-board",
		Usage: "enqueue task-board generation from local docs payload",
		Flags: []urcli.Flag{
			&urcli.StringFlag{Name: "ROOT_DIRECTORY", Value: rootDirectory, Destination: &rootDirectory, Sources: urcli.EnvVars("ROOT_DIRECTORY")},
			&urcli.IntFlag{Name: "MAX_DEPTH", Value: traversalDepth, Destination: &traversalDepth, Sources: urcli.EnvVars("MAX_DEPTH")},
			&urcli.StringFlag{Name: "PROMPT", Value: boardPrompt, Destination: &boardPrompt, Sources: urcli.EnvVars("PROMPT")},
			&urcli.StringFlag{Name: "MODEL", Value: boardModel, Destination: &boardModel, Sources: urcli.EnvVars("MODEL")},
		},
		Action: func(actionCtx context.Context, _ *urcli.Command) error {
			if runner == nil {
				return fmt.Errorf("runner cannot be nil")
			}

			taskID, err := runner.Run(actionCtx, application.GenerateTaskBoardInput{
				RootDirectory: rootDirectory,
				MaxDepth:      traversalDepth,
				Prompt:        boardPrompt,
				Model:         boardModel,
			})
			if err != nil {
				return err
			}
			if logger != nil {
				logger.WithField("task_id", taskID).Info("enqueued generate-task-board")
			}
			return nil
		},
	}
}

func newRunWorkerCommand(workerRunner WorkerRuntimeRunner) *urcli.Command {
	return &urcli.Command{
		Name:  "run-worker",
		Usage: "run asynq worker runtime",
		Action: func(actionCtx context.Context, _ *urcli.Command) error {
			if workerRunner == nil {
				return fmt.Errorf("worker runner cannot be nil")
			}
			return workerRunner.Run(actionCtx)
		},
	}
}

func defaultDatabaseDSN() string {
	return "sqlite:///" + filepath.Join(".", "agentic-worktrees.db")
}
