package filestore

import (
	"agentic-orchestrator/internal/domain/failures"
	"testing"
)

func TestProfileValidateRequiresSupportedKinds(t *testing.T) {
	profile := Profile{StoreKind: "", DeliveryKind: DeliveryKindNone, Root: "projects"}
	if err := profile.Validate(); !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error for store kind, got %v", err)
	}
}

func TestProfileValidateCompatibilityRejectsLocalWithCDN(t *testing.T) {
	profile := Profile{StoreKind: StoreKindLocalDisk, DeliveryKind: DeliveryKindGoogleCDN, Root: "/tmp/store"}
	if err := profile.ValidateCompatibility(false); !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal incompatibility error, got %v", err)
	}
}

func TestProfileValidateCompatibilityRequiresGCSAndGoogleCDNForCDNFlows(t *testing.T) {
	profile := Profile{StoreKind: StoreKindGoogleCloudStorage, DeliveryKind: DeliveryKindGoogleCDN, Root: "projects"}
	if err := profile.ValidateCompatibility(true); err != nil {
		t.Fatalf("expected gcs/google-cdn profile to be compatible for cdn flows, got %v", err)
	}
}
