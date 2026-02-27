package logruslogger

import "github.com/sirupsen/logrus"

type AsynqAdapter struct {
	logger *logrus.Logger
}

func NewAsynqAdapter(logger *logrus.Logger) *AsynqAdapter {
	return &AsynqAdapter{logger: logger}
}

func (adapter *AsynqAdapter) Debug(args ...interface{}) {
	adapter.logger.Debug(args...)
}

func (adapter *AsynqAdapter) Info(args ...interface{}) {
	adapter.logger.Info(args...)
}

func (adapter *AsynqAdapter) Warn(args ...interface{}) {
	adapter.logger.Warn(args...)
}

func (adapter *AsynqAdapter) Error(args ...interface{}) {
	adapter.logger.Error(args...)
}

func (adapter *AsynqAdapter) Fatal(args ...interface{}) {
	adapter.logger.Fatal(args...)
}
