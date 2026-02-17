package commands

import (
	"context"

	"github.com/sirupsen/logrus"
	urcli "github.com/urfave/cli/v3"

	infralogger "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/logger"
)

var (
	logLevel  = "info"
	logFormat = "json"
	logger    *logrus.Logger
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
		},
		Before: func(hookCtx context.Context, _ *urcli.Command) (context.Context, error) {
			configuredLogger, err := infralogger.New(logLevel, logFormat)
			if err != nil {
				return hookCtx, err
			}
			logger = configuredLogger
			return hookCtx, nil
		},
	}
}
