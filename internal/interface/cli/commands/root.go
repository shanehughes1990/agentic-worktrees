package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	urcli "github.com/urfave/cli/v3"

	infradatabase "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/database"
	infralogger "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/logger"
)

var (
	logLevel    = "info"
	logFormat   = "json"
	databaseDSN = defaultDatabaseDSN()
	logger      *logrus.Logger
)

func NewRootCommand() *urcli.Command {
	return &urcli.Command{
		Name:  "cli",
		Usage: "bootstrap cli",
		Flags: []urcli.Flag{
			&urcli.StringFlag{
				Name:        "LOG_LEVEL",
				Value:       logLevel,
				Destination: &logLevel,
				Sources:     urcli.EnvVars("LOG_LEVEL"),
			},
			&urcli.StringFlag{
				Name:        "LOG_FORMAT",
				Value:       logFormat,
				Destination: &logFormat,
				Sources:     urcli.EnvVars("LOG_FORMAT"),
			},
			&urcli.StringFlag{
				Name:        "DATABASE_DSN",
				Value:       databaseDSN,
				Destination: &databaseDSN,
				Sources:     urcli.EnvVars("DATABASE_DSN"),
			},
		},
		Before: func(hookCtx context.Context, _ *urcli.Command) (context.Context, error) {
			configuredLogger, err := infralogger.New(logLevel, logFormat)
			if err != nil {
				return hookCtx, err
			}
			logger = configuredLogger

			if _, err := infradatabase.NewGormClient(logger, databaseDSN); err != nil {
				return hookCtx, fmt.Errorf("bootstrap database: %w", err)
			}
			return hookCtx, nil
		},
	}
}

func defaultDatabaseDSN() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "sqlite:///agentic-worktrees.db"
	}
	return "sqlite:///" + filepath.Join(cwd, "agentic-worktrees.db")
}
