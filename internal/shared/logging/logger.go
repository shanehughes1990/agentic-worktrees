package logging

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const timestampFormat = time.RFC3339Nano

type Config struct {
	Level  string
	Format string
}

func New(cfg Config) (*logrus.Logger, error) {
	logger := logrus.New()

	level := strings.TrimSpace(cfg.Level)
	if level == "" {
		level = logrus.InfoLevel.String()
	}

	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}
	logger.SetLevel(parsedLevel)

	format := strings.ToLower(strings.TrimSpace(cfg.Format))
	switch format {
	case "", "json":
		logger.SetFormatter(&logrus.JSONFormatter{TimestampFormat: timestampFormat})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: timestampFormat})
	default:
		return nil, fmt.Errorf("unsupported log format %q", cfg.Format)
	}

	return logger, nil
}

func MustNew(cfg Config) *logrus.Logger {
	logger, err := New(cfg)
	if err != nil {
		panic(err)
	}
	return logger
}

func WithFields(logger *logrus.Logger, fields logrus.Fields) *logrus.Entry {
	if logger == nil {
		logger = logrus.New()
		logger.SetFormatter(&logrus.JSONFormatter{TimestampFormat: timestampFormat})
	}
	return logger.WithFields(fields)
}
