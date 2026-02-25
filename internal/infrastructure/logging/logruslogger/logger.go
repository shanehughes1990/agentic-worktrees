package logruslogger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	runtimeRoot := strings.TrimSpace(os.Getenv("APP_ROOT_DIR"))
	if runtimeRoot == "" {
		runtimeRoot = ".worktree"
	}
	logger, err := New(os.Getenv("LOG_FORMAT"), os.Getenv("LOG_LEVEL"), filepath.ToSlash(filepath.Join(runtimeRoot, "logs", "app.log")))
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
		MaxBackups: 1,
		MaxAge:     30,
		Compress:   false,
	}

	if info, err := os.Stat(filePath); err == nil && info.Size() > 0 {
		if err := rotatingWriter.Rotate(); err != nil {
			return nil, fmt.Errorf("rotate existing startup log: %w", err)
		}
		if err := pruneRotatedBackups(filePath, 1); err != nil {
			return nil, fmt.Errorf("prune rotated logs: %w", err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("stat log file: %w", err)
	}

	return rotatingWriter, nil
}

func pruneRotatedBackups(filePath string, maxBackups int) error {
	if maxBackups < 0 {
		maxBackups = 0
	}
	directory := filepath.Dir(filePath)
	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	rotatedPrefix := baseName + "-"

	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read log directory: %w", err)
	}

	rotatedPaths := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if !strings.HasPrefix(name, rotatedPrefix) {
			continue
		}
		if !strings.HasSuffix(name, ".log") && !strings.HasSuffix(name, ".log.gz") {
			continue
		}
		rotatedPaths = append(rotatedPaths, filepath.Join(directory, name))
	}

	if len(rotatedPaths) <= maxBackups {
		return nil
	}

	sort.Strings(rotatedPaths)
	deleteCount := len(rotatedPaths) - maxBackups
	for i := 0; i < deleteCount; i++ {
		if err := os.Remove(rotatedPaths[i]); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove rotated log %s: %w", rotatedPaths[i], err)
		}
	}

	return nil
}
