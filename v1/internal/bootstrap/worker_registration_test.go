package bootstrap

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"testing"
)

type capabilityRecordingConsumer struct {
	registeredKinds []taskengine.JobKind
	advertisement   *taskengine.WorkerCapabilityAdvertisement
}

func (consumer *capabilityRecordingConsumer) Register(kind taskengine.JobKind, handler taskengine.Handler) error {
	consumer.registeredKinds = append(consumer.registeredKinds, kind)
	return nil
}

func (consumer *capabilityRecordingConsumer) Start() error {
	return nil
}

func (consumer *capabilityRecordingConsumer) Shutdown(ctx context.Context) error {
	return nil
}

func (consumer *capabilityRecordingConsumer) Advertise(ctx context.Context, advertisement taskengine.WorkerCapabilityAdvertisement) error {
	consumer.advertisement = &advertisement
	return nil
}

func TestRegisterWorkerJobsAdvertisesCapabilitiesForRegisteredKinds(t *testing.T) {
	consumer := &capabilityRecordingConsumer{}
	registrations := []workerJobRegistration{
		{kind: taskengine.JobKindIngestionAgent, label: "ingestion agent", handler: taskengine.HandlerFunc(func(context.Context, taskengine.Job) error { return nil })},
		{kind: taskengine.JobKindAgentWorkflow, label: "agent workflow", handler: taskengine.HandlerFunc(func(context.Context, taskengine.Job) error { return nil })},
		{kind: taskengine.JobKindSCMWorkflow, label: "scm workflow", handler: taskengine.HandlerFunc(func(context.Context, taskengine.Job) error { return nil })},
	}

	if err := registerWorkerJobs(context.Background(), consumer, "worker-1", registrations); err != nil {
		t.Fatalf("register worker jobs: %v", err)
	}

	expectedKinds := []taskengine.JobKind{taskengine.JobKindIngestionAgent, taskengine.JobKindAgentWorkflow, taskengine.JobKindSCMWorkflow}
	if len(consumer.registeredKinds) != len(expectedKinds) {
		t.Fatalf("expected %d registrations, got %d", len(expectedKinds), len(consumer.registeredKinds))
	}
	for index, expectedKind := range expectedKinds {
		if consumer.registeredKinds[index] != expectedKind {
			t.Fatalf("expected registered kind %q at index %d, got %q", expectedKind, index, consumer.registeredKinds[index])
		}
	}

	if consumer.advertisement == nil {
		t.Fatalf("expected worker capability advertisement")
	}
	if consumer.advertisement.WorkerID != "worker-1" {
		t.Fatalf("expected worker id worker-1, got %q", consumer.advertisement.WorkerID)
	}
	if len(consumer.advertisement.Capabilities) != len(expectedKinds) {
		t.Fatalf("expected %d capabilities, got %d", len(expectedKinds), len(consumer.advertisement.Capabilities))
	}
	for index, expectedKind := range expectedKinds {
		if consumer.advertisement.Capabilities[index].Kind != expectedKind {
			t.Fatalf("expected capability %q at index %d, got %q", expectedKind, index, consumer.advertisement.Capabilities[index].Kind)
		}
	}
}
