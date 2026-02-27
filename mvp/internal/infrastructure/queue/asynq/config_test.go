package asynq

import (
	"runtime"
	"testing"
)

func TestNewConfigRequiresRedisURI(t *testing.T) {
	_, err := NewConfig("   ")
	if err == nil {
		t.Fatalf("expected redis uri required error")
	}
}

func TestNewConfigDefaults(t *testing.T) {
	cfg, err := NewConfig("redis://localhost:6379/0")
	if err != nil {
		t.Fatalf("unexpected config error: %v", err)
	}

	if cfg.Concurrency != runtime.GOMAXPROCS(0) {
		t.Fatalf("unexpected concurrency: %d", cfg.Concurrency)
	}
	if cfg.Queues[QueueIngestion] != 6 || cfg.Queues[QueueAgent] != 3 || cfg.Queues[QueueDefault] != 1 {
		t.Fatalf("unexpected queue priorities: %#v", cfg.Queues)
	}
}
