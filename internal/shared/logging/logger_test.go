package logging

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewDefaultsToInfoJSON(t *testing.T) {
	logger, err := New(Config{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if logger.GetLevel() != logrus.InfoLevel {
		t.Fatalf("expected info level, got %s", logger.GetLevel())
	}

	formatter, ok := logger.Formatter.(*logrus.JSONFormatter)
	if !ok {
		t.Fatalf("expected JSON formatter")
	}
	if formatter.TimestampFormat != timestampFormat {
		t.Fatalf("expected JSON timestamp format %q, got %q", timestampFormat, formatter.TimestampFormat)
	}
}

func TestNewInvalidLevel(t *testing.T) {
	_, err := New(Config{Level: "not-a-level"})
	if err == nil {
		t.Fatalf("expected level parse error")
	}
}

func TestNewTextFormat(t *testing.T) {
	logger, err := New(Config{Format: "text", Level: "debug"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if logger.GetLevel() != logrus.DebugLevel {
		t.Fatalf("expected debug level, got %s", logger.GetLevel())
	}

	formatter, ok := logger.Formatter.(*logrus.TextFormatter)
	if !ok {
		t.Fatalf("expected text formatter")
	}
	if !formatter.FullTimestamp {
		t.Fatalf("expected full timestamps enabled")
	}
	if formatter.TimestampFormat != timestampFormat {
		t.Fatalf("expected text timestamp format %q, got %q", timestampFormat, formatter.TimestampFormat)
	}
}

func TestWithFieldsNilLogger(t *testing.T) {
	entry := WithFields(nil, logrus.Fields{"run_id": "r1"})
	if entry == nil {
		t.Fatalf("expected entry")
	}
	if entry.Data["run_id"] != "r1" {
		t.Fatalf("expected run_id field")
	}

	formatter, ok := entry.Logger.Formatter.(*logrus.JSONFormatter)
	if !ok {
		t.Fatalf("expected JSON formatter for nil logger path")
	}
	if formatter.TimestampFormat != timestampFormat {
		t.Fatalf("expected JSON timestamp format %q, got %q", timestampFormat, formatter.TimestampFormat)
	}
}
