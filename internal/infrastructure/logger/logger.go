package logger

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const timestampFormat = time.RFC3339Nano

func New(level string, format string) (*logrus.Logger, error) {
	logger := logrus.New()

	trimmedLevel := strings.TrimSpace(level)
	if trimmedLevel == "" {
		trimmedLevel = logrus.InfoLevel.String()
	}

	parsedLevel, err := logrus.ParseLevel(trimmedLevel)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}
	logger.SetLevel(parsedLevel)

	trimmedFormat := strings.ToLower(strings.TrimSpace(format))
	switch trimmedFormat {
	case "", "json":
		logger.SetFormatter(&logrus.JSONFormatter{TimestampFormat: timestampFormat})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: timestampFormat})
	default:
		return nil, fmt.Errorf("unsupported log format %q", format)
	}

	return logger, nil
}
