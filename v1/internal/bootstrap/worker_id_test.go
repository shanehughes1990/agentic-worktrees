package bootstrap

import "testing"

func TestBuildWorkerIDIncludesHostname(t *testing.T) {
	workerID := buildWorkerID("agentic-orchestrator", "worker-replica-2", 8081)
	if workerID != "agentic-orchestrator-worker-replica-2-worker-8081" {
		t.Fatalf("unexpected worker id: %q", workerID)
	}
}

func TestBuildWorkerIDNormalizesMissingValues(t *testing.T) {
	workerID := buildWorkerID("  ", "", 8081)
	if workerID != "worker-unknown-host-worker-8081" {
		t.Fatalf("unexpected fallback worker id: %q", workerID)
	}
}
