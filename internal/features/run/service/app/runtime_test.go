package app

import (
	"context"
	"testing"
)

func TestInitAndRunCLI(t *testing.T) {
	t.Setenv("APP_NAME", "agentic-test")
	t.Setenv("APP_LOG_LEVEL", "info")
	t.Setenv("APP_LOG_FORMAT", "json")

	runtime, err := Init(KindCLI)
	if err != nil {
		t.Fatalf("init cli runtime: %v", err)
	}

	if err := Run(context.Background(), runtime); err != nil {
		t.Fatalf("run cli runtime: %v", err)
	}
}

func TestInitAndRunWorker(t *testing.T) {
	t.Setenv("APP_NAME", "agentic-test")
	t.Setenv("APP_LOG_LEVEL", "info")
	t.Setenv("APP_LOG_FORMAT", "json")

	runtime, err := Init(KindWorker)
	if err != nil {
		t.Fatalf("init worker runtime: %v", err)
	}

	if err := Run(context.Background(), runtime); err != nil {
		t.Fatalf("run worker runtime: %v", err)
	}
}

func TestInitRejectsInvalidKind(t *testing.T) {
	t.Setenv("APP_NAME", "agentic-test")
	t.Setenv("APP_LOG_LEVEL", "info")
	t.Setenv("APP_LOG_FORMAT", "json")

	if _, err := Init(Kind("invalid")); err == nil {
		t.Fatalf("expected invalid kind error")
	}
}

func TestRunRejectsNilRuntime(t *testing.T) {
	if err := Run(context.Background(), nil); err == nil {
		t.Fatalf("expected nil runtime error")
	}
}
