package worker

import (
	"testing"
)

func TestApplicationRootPathUsesWorkerArtifactRootDirectory(t *testing.T) {
	config := Config{Worker: WorkerConfig{ArtifactRootDirectory: "/tmp/worker-root"}}
	if got := config.ApplicationRootPath(); got != "/tmp/worker-root" {
		t.Fatalf("expected application root path to use worker artifact root directory, got %q", got)
	}
}


func TestApplicationRootPathFallsBackToDefaultWhenEmpty(t *testing.T) {
	config := Config{Worker: WorkerConfig{ArtifactRootDirectory: ""}}
	if got := config.ApplicationRootPath(); got != ".agentic-orchestrator" {
		t.Fatalf("expected default application root path, got %q", got)
	}
}
