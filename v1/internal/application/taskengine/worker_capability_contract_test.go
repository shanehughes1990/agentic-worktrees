package taskengine

import (
	"errors"
	"testing"
)

func TestWorkerCapabilityAdvertisementValidateRequiresWorkerID(t *testing.T) {
	err := (WorkerCapabilityAdvertisement{
		Capabilities: []WorkerCapability{{Kind: JobKindAgentWorkflow}},
	}).Validate()
	if !errors.Is(err, ErrInvalidWorkerCapabilityAdvertisement) {
		t.Fatalf("expected ErrInvalidWorkerCapabilityAdvertisement, got %v", err)
	}
}

func TestWorkerCapabilityAdvertisementValidateRequiresCapabilities(t *testing.T) {
	err := (WorkerCapabilityAdvertisement{
		WorkerID: "worker-1",
	}).Validate()
	if !errors.Is(err, ErrInvalidWorkerCapabilityAdvertisement) {
		t.Fatalf("expected ErrInvalidWorkerCapabilityAdvertisement, got %v", err)
	}
}

func TestWorkerCapabilityAdvertisementValidateRejectsEmptyCapabilityKind(t *testing.T) {
	err := (WorkerCapabilityAdvertisement{
		WorkerID:     "worker-1",
		Capabilities: []WorkerCapability{{Kind: ""}},
	}).Validate()
	if !errors.Is(err, ErrInvalidWorkerCapabilityAdvertisement) {
		t.Fatalf("expected ErrInvalidWorkerCapabilityAdvertisement, got %v", err)
	}
}

func TestWorkerCapabilityAdvertisementValidateRejectsDuplicateKinds(t *testing.T) {
	err := (WorkerCapabilityAdvertisement{
		WorkerID: "worker-1",
		Capabilities: []WorkerCapability{
			{Kind: JobKindSCMWorkflow},
			{Kind: JobKindSCMWorkflow},
		},
	}).Validate()
	if !errors.Is(err, ErrInvalidWorkerCapabilityAdvertisement) {
		t.Fatalf("expected ErrInvalidWorkerCapabilityAdvertisement, got %v", err)
	}
}

func TestWorkerCapabilityAdvertisementValidateAcceptsDistinctKinds(t *testing.T) {
	err := (WorkerCapabilityAdvertisement{
		WorkerID: "worker-1",
		Capabilities: []WorkerCapability{
			{Kind: JobKindIngestionAgent},
			{Kind: JobKindAgentWorkflow},
			{Kind: JobKindSCMWorkflow},
		},
	}).Validate()
	if err != nil {
		t.Fatalf("validate advertisement: %v", err)
	}
}
