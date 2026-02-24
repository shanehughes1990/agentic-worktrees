package logruslogger

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func New(format, level string) (*logrus.Logger, error) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	parsedLevel, err := logrus.ParseLevel(strings.TrimSpace(level))
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	logger.SetLevel(parsedLevel)

	switch strings.TrimSpace(format) {
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{})
	default:
		return nil, fmt.Errorf("invalid log format: %s", format)
	}

	return logger, nil
}

func NewFromEnv() *logrus.Logger {
	logger, err := New(os.Getenv("LOG_FORMAT"), os.Getenv("LOG_LEVEL"))
	if err != nil {
		fallback := logrus.New()
		fallback.SetOutput(os.Stdout)
		fallback.SetLevel(logrus.InfoLevel)
		fallback.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
		return fallback
	}
	return logger
}
