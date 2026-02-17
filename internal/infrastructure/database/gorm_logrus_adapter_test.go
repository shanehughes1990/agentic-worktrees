package database

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func newTestLogger() (*logrus.Logger, *bytes.Buffer) {
	var buffer bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buffer)
	logger.SetLevel(logrus.TraceLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	})
	return logger, &buffer
}

func TestNewGormLogrusAdapterAllowsNilLogger(t *testing.T) {
	adapter := NewGormLogrusAdapter(nil)
	typed, ok := adapter.(*gormLogrusAdapter)
	if !ok {
		t.Fatalf("expected concrete adapter type")
	}
	if typed.logger == nil {
		t.Fatalf("expected fallback logger")
	}
}

func TestGormLogrusAdapterLogModeCopies(t *testing.T) {
	logger, _ := newTestLogger()
	adapter := NewGormLogrusAdapter(logger).(*gormLogrusAdapter)
	adapted := adapter.LogMode(gormlogger.Error).(*gormLogrusAdapter)

	if adapter.logLevel != gormlogger.Info {
		t.Fatalf("expected original adapter log level unchanged")
	}
	if adapted.logLevel != gormlogger.Error {
		t.Fatalf("expected copied adapter with updated log level")
	}
}

func TestGormLogrusAdapterInfoWarnError(t *testing.T) {
	logger, buffer := newTestLogger()
	adapter := NewGormLogrusAdapter(logger).(*gormLogrusAdapter)
	ctx := context.Background()

	adapter.Info(ctx, "info %s", "message")
	adapter.Warn(ctx, "warn %s", "message")
	adapter.Error(ctx, "error %s", "message")

	output := buffer.String()
	if !strings.Contains(output, "info message") {
		t.Fatalf("expected debug info message in output")
	}
	if !strings.Contains(output, "warn message") {
		t.Fatalf("expected warn message in output")
	}
	if !strings.Contains(output, "error message") {
		t.Fatalf("expected error message in output")
	}
}

func TestGormLogrusAdapterInfoRespectsLevel(t *testing.T) {
	logger, buffer := newTestLogger()
	adapter := NewGormLogrusAdapter(logger).LogMode(gormlogger.Warn).(*gormLogrusAdapter)
	adapter.Info(context.Background(), "should not log")
	if buffer.Len() != 0 {
		t.Fatalf("expected no info logs below info level")
	}
}

func TestGormLogrusAdapterWarnRespectsLevel(t *testing.T) {
	logger, buffer := newTestLogger()
	adapter := NewGormLogrusAdapter(logger).LogMode(gormlogger.Error).(*gormLogrusAdapter)
	adapter.Warn(context.Background(), "should not log")
	if buffer.Len() != 0 {
		t.Fatalf("expected no warn logs below warn level")
	}
}

func TestGormLogrusAdapterErrorRespectsLevel(t *testing.T) {
	logger, buffer := newTestLogger()
	adapter := NewGormLogrusAdapter(logger).LogMode(gormlogger.Silent).(*gormLogrusAdapter)
	adapter.Error(context.Background(), "should not log")
	if buffer.Len() != 0 {
		t.Fatalf("expected no error logs below error level")
	}
}

func TestGormLogrusAdapterTraceSilentSkipsFormatter(t *testing.T) {
	logger, _ := newTestLogger()
	adapter := NewGormLogrusAdapter(logger).LogMode(gormlogger.Silent).(*gormLogrusAdapter)

	called := false
	adapter.Trace(context.Background(), time.Now(), func() (string, int64) {
		called = true
		return "select 1", 1
	}, nil)

	if called {
		t.Fatalf("expected trace formatter callback not called in silent mode")
	}
}

func TestGormLogrusAdapterTraceErrorBranch(t *testing.T) {
	logger, buffer := newTestLogger()
	adapter := NewGormLogrusAdapter(logger).LogMode(gormlogger.Error).(*gormLogrusAdapter)

	adapter.Trace(context.Background(), time.Now().Add(-time.Millisecond), func() (string, int64) {
		return "select 1", 1
	}, errors.New("boom"))

	if !strings.Contains(buffer.String(), "gorm trace error") {
		t.Fatalf("expected error trace log")
	}
}

func TestGormLogrusAdapterTraceIgnoresRecordNotFoundError(t *testing.T) {
	logger, buffer := newTestLogger()
	adapter := NewGormLogrusAdapter(logger).LogMode(gormlogger.Error).(*gormLogrusAdapter)

	adapter.Trace(context.Background(), time.Now().Add(-time.Millisecond), func() (string, int64) {
		return "select 1", 0
	}, gorm.ErrRecordNotFound)

	if strings.Contains(buffer.String(), "gorm trace error") {
		t.Fatalf("expected record not found to be ignored by error branch")
	}
}

func TestGormLogrusAdapterTraceSlowQueryBranch(t *testing.T) {
	logger, buffer := newTestLogger()
	adapter := NewGormLogrusAdapter(logger).LogMode(gormlogger.Warn).(*gormLogrusAdapter)
	adapter.slowThreshold = time.Microsecond

	adapter.Trace(context.Background(), time.Now().Add(-time.Millisecond), func() (string, int64) {
		return "select 1", 1
	}, nil)

	if !strings.Contains(buffer.String(), "slow_query") {
		t.Fatalf("expected slow query trace log")
	}
}

func TestGormLogrusAdapterTraceInfoBranch(t *testing.T) {
	logger, buffer := newTestLogger()
	adapter := NewGormLogrusAdapter(logger).LogMode(gormlogger.Info).(*gormLogrusAdapter)
	adapter.slowThreshold = time.Hour

	adapter.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "select 1", 1
	}, nil)

	if !strings.Contains(buffer.String(), "gorm trace elapsed=") {
		t.Fatalf("expected normal trace log")
	}
}

func TestGormLogrusAdapterTraceWarnFastQueryNoLog(t *testing.T) {
	logger, buffer := newTestLogger()
	adapter := NewGormLogrusAdapter(logger).LogMode(gormlogger.Warn).(*gormLogrusAdapter)
	adapter.slowThreshold = time.Hour

	adapter.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "select 1", 1
	}, nil)

	if buffer.Len() != 0 {
		t.Fatalf("expected no log for fast query at warn level")
	}
}
