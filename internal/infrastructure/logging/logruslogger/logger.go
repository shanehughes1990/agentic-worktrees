package logruslogger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(format, level, filePath string) (*logrus.Logger, error) {
	logger := logrus.New()

	cleanPath := strings.TrimSpace(filePath)
	if cleanPath == "" {
		return nil, fmt.Errorf("log file path is required")
	}

	rotatingWriter, err := newRotatingWriter(cleanPath)
	if err != nil {
		return nil, err
	}
	logger.SetOutput(rotatingWriter)

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
	logger, err := New(os.Getenv("LOG_FORMAT"), os.Getenv("LOG_LEVEL"), envOrDefault("LOG_FILE_PATH", "logs/app.log"))
	if err != nil {
		fallback := logrus.New()
		fallback.SetOutput(os.Stderr)
		fallback.SetLevel(logrus.InfoLevel)
		fallback.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
		fallback.Error(err)
		return fallback
	}
	return logger
}

func newRotatingWriter(filePath string) (*lumberjack.Logger, error) {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	rotatingWriter := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    100,
		MaxBackups: 30,
		MaxAge:     30,
		Compress:   true,
	}

	if info, err := os.Stat(filePath); err == nil && info.Size() > 0 {
		if err := rotatingWriter.Rotate(); err != nil {
			return nil, fmt.Errorf("rotate existing startup log: %w", err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("stat log file: %w", err)
	}

	return rotatingWriter, nil
}

func envOrDefault(key, fallbackValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallbackValue
	}
	return value
}
