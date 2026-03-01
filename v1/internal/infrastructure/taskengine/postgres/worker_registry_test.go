package postgres

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"testing"
)

func TestWorkerRegistryUpsert(t *testing.T) {
	registry, err := NewWorkerRegistry(newTestDB(t))
	if err != nil {
		t.Fatalf("new worker registry: %v", err)
	}
	advertisement := taskengine.WorkerCapabilityAdvertisement{
		WorkerID: "worker-1",
		Capabilities: []taskengine.WorkerCapability{
			{Kind: taskengine.JobKindIngestionAgent},
			{Kind: taskengine.JobKindSCMWorkflow},
		},
	}
	if err := registry.Upsert(context.Background(), advertisement); err != nil {
		t.Fatalf("upsert worker registry: %v", err)
	}
	var count int64
	if err := registry.db.Model(&workerRegistryRecord{}).Where("worker_id = ?", "worker-1").Count(&count).Error; err != nil {
		t.Fatalf("count worker registry records: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 worker registry record, got %d", count)
	}
}
