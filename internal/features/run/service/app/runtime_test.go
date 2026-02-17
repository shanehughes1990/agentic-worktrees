package app

import (
	"context"
	"testing"
)

func seedAppEnv(t *testing.T) {
	t.Setenv("APP_NAME", "agentic-test")
	t.Setenv("APP_LOG_LEVEL", "info")
	t.Setenv("APP_LOG_FORMAT", "json")
	t.Setenv("APP_REDIS_ADDR", "127.0.0.1:6379")
	t.Setenv("APP_ASYNQ_QUEUE", "default")
	t.Setenv("APP_BOARD_PATH", "state/board.json")
	t.Setenv("APP_CHECKPOINT_PATH", "state/checkpoints.json")
	t.Setenv("APP_WORKER_CONCURRENCY", "2")
	t.Setenv("APP_COPILOT_ADK_BOARD_URL", "https://example.invalid/adk")
}

func TestInitCLI(t *testing.T) {
	seedAppEnv(t)
	runtime, err := Init(KindCLI)
	if err != nil {
		t.Fatalf("init cli runtime: %v", err)
	}
	if runtime.kind != KindCLI {
		t.Fatalf("expected cli kind")
	}
}

func TestInitWorker(t *testing.T) {
	seedAppEnv(t)
	runtime, err := Init(KindWorker)
	if err != nil {
		t.Fatalf("init worker runtime: %v", err)
	}
	if runtime.kind != KindWorker {
		t.Fatalf("expected worker kind")
	}
}

func TestInitRejectsInvalidKind(t *testing.T) {
	seedAppEnv(t)
	if _, err := Init(Kind("invalid")); err == nil {
		t.Fatalf("expected invalid kind error")
	}
}

func TestRunRejectsNilRuntime(t *testing.T) {
	if err := Run(context.Background(), nil); err == nil {
		t.Fatalf("expected nil runtime error")
	}
}
