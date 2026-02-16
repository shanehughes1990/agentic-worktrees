package app

import (
	"context"

	"github.com/sirupsen/logrus"
)

func (runtime *Runtime) runWorker(_ context.Context) error {
	runtime.logger.WithFields(logrus.Fields{
		"component": "worker",
		"app":       runtime.config.Name,
	}).Info("worker bootstrapped")
	return nil
}
