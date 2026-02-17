package logger

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewDefaultsToInfoAndJSON(t *testing.T) {
	logger, err := New("", "")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}
	if logger.GetLevel() != logrus.InfoLevel {
		t.Fatalf("expected info level, got %s", logger.GetLevel())
	}
	if _, ok := logger.Formatter.(*logrus.JSONFormatter); !ok {
		t.Fatalf("expected json formatter")
	}
}

func TestNewTextFormat(t *testing.T) {
	logger, err := New("debug", "text")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}
	if logger.GetLevel() != logrus.DebugLevel {
		t.Fatalf("expected debug level, got %s", logger.GetLevel())
	}
	if _, ok := logger.Formatter.(*logrus.TextFormatter); !ok {
		t.Fatalf("expected text formatter")
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	if _, err := New("bad-level", "json"); err == nil {
		t.Fatalf("expected invalid level error")
	}
	if _, err := New("info", "xml"); err == nil {
		t.Fatalf("expected invalid format error")
	}
}
