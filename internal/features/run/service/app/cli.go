package app

import (
	"context"

	"github.com/sirupsen/logrus"
)

func (runtime *Runtime) runCLI(_ context.Context) error {
	runtime.logger.WithFields(logrus.Fields{
		"component": "cli",
		"app":       runtime.config.Name,
	}).Info("cli bootstrapped")
	return nil
}
