package logruslogger

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCreatesAndRotatesLogFile(t *testing.T) {
	logsDir := t.TempDir()
	logFilePath := filepath.Join(logsDir, "app.log")

	if err := os.WriteFile(logFilePath, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed log file: %v", err)
	}

	logger, err := New("text", "info", logFilePath)
	if err != nil {
		t.Fatalf("unexpected logger creation error: %v", err)
	}
	logger.Info("hello")

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("read current log file: %v", err)
	}
	if len(content) == 0 {
		t.Fatalf("expected current log content")
	}

	rotated, err := filepath.Glob(filepath.Join(logsDir, "app-*.log"))
	if err != nil {
		t.Fatalf("glob rotated logs: %v", err)
	}
	if len(rotated) == 0 {
		t.Fatalf("expected rotated startup log file")
	}
}

func TestNewRejectsInvalidLevel(t *testing.T) {
	_, err := New("text", "bad-level", filepath.Join(t.TempDir(), "app.log"))
	if err == nil {
		t.Fatalf("expected invalid level error")
	}
}

func TestNewKeepsSingleRotatedBackup(t *testing.T) {
	logsDir := t.TempDir()
	logFilePath := filepath.Join(logsDir, "app.log")

	for i := 0; i < 3; i++ {
		if err := os.WriteFile(logFilePath, []byte("seed"), 0o644); err != nil {
			t.Fatalf("seed log file: %v", err)
		}
		logger, err := New("text", "info", logFilePath)
		if err != nil {
			t.Fatalf("unexpected logger creation error: %v", err)
		}
		logger.Infof("boot %d", i)
	}

	rotated, err := filepath.Glob(filepath.Join(logsDir, "app-*.log*"))
	if err != nil {
		t.Fatalf("glob rotated logs: %v", err)
	}
	if len(rotated) > 1 {
		t.Fatalf("expected at most one rotated file, got %d (%v)", len(rotated), rotated)
	}
}
