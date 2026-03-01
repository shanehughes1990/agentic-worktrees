package postgres

import (
	"context"
	"testing"
	"time"
)

func TestNewGormLoggerAppliesDefaultSlowThreshold(t *testing.T) {
	raw := newGormLogger(nil, 0)
	adapter, ok := raw.(*gormLogAdapter)
	if !ok {
		t.Fatalf("expected *gormLogAdapter, got %T", raw)
	}
	if adapter.slowQueryThreshold != defaultSlowQueryThreshold {
		t.Fatalf("expected default threshold %s, got %s", defaultSlowQueryThreshold, adapter.slowQueryThreshold)
	}
}

func TestGormLogAdapterNilEntryIsNoop(t *testing.T) {
	adapter := &gormLogAdapter{entry: nil, slowQueryThreshold: time.Second}
	adapter.Info(context.Background(), "info")
	adapter.Warn(context.Background(), "warn")
	adapter.Error(context.Background(), "error")
	adapter.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "select 1", 1
	}, nil)
}

func TestIsSlowQuery(t *testing.T) {
	if !isSlowQuery(200*time.Millisecond, 100*time.Millisecond) {
		t.Fatalf("expected query to be classified as slow")
	}
	if isSlowQuery(50*time.Millisecond, 100*time.Millisecond) {
		t.Fatalf("expected query to not be classified as slow")
	}
}
