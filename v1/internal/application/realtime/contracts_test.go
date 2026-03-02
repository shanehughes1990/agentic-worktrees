package realtime

import (
	"context"
	"testing"

	domainrealtime "agentic-orchestrator/internal/domain/realtime"
)

type testTransport struct{}

func (testTransport) Publish(ctx context.Context, client string, message domainrealtime.Message) error {
	return nil
}

func (testTransport) Subscribe(ctx context.Context, client string, contract domainrealtime.Contract, subContracts []domainrealtime.SubContract, handler func(domainrealtime.Message) error) error {
	return nil
}

type testAPIScope struct{}

func (testAPIScope) RequiredWorkerRegistryCapability() domainrealtime.Capability {
	return domainrealtime.Capability{Contract: domainrealtime.ContractWorkerRegistry, Version: "v1", SubContracts: []domainrealtime.SubContract{domainrealtime.SubContractHeartbeatRequest}}
}

func (testAPIScope) HandleWorkerRegistryMessage(ctx context.Context, message domainrealtime.Message) error {
	return nil
}

func (testAPIScope) RequiredCapabilities() domainrealtime.APIContractRequirements {
	return domainrealtime.APIContractRequirements{Required: []domainrealtime.Capability{testAPIScope{}.RequiredWorkerRegistryCapability()}}
}

func (testAPIScope) AdvertisedCapabilities() domainrealtime.APIContractAdvertisement {
	return domainrealtime.APIContractAdvertisement{Implemented: []domainrealtime.Capability{{
		Contract:     domainrealtime.ContractWorkerRegistry,
		Version:      "v1",
		SubContracts: []domainrealtime.SubContract{domainrealtime.SubContractHeartbeatRequest, domainrealtime.SubContractHeartbeatResponse},
	}}}
}

type apiScopeMissingHeartbeatResponse struct{ testAPIScope }

func (apiScopeMissingHeartbeatResponse) AdvertisedCapabilities() domainrealtime.APIContractAdvertisement {
	return domainrealtime.APIContractAdvertisement{Implemented: []domainrealtime.Capability{{
		Contract:     domainrealtime.ContractWorkerRegistry,
		Version:      "v1",
		SubContracts: []domainrealtime.SubContract{domainrealtime.SubContractHeartbeatRequest},
	}}}
}

type testWorkerScope struct{}

func (testWorkerScope) ImplementedWorkerRegistryCapability() domainrealtime.Capability {
	return domainrealtime.Capability{Contract: domainrealtime.ContractWorkerRegistry, Version: "v1", SubContracts: []domainrealtime.SubContract{domainrealtime.SubContractHeartbeatRequest, domainrealtime.SubContractHeartbeatResponse}}
}

func (testWorkerScope) HandleWorkerRegistryMessage(ctx context.Context, message domainrealtime.Message) error {
	return nil
}

func (testWorkerScope) RequiredCapabilities() domainrealtime.WorkerContractRequirements {
	return domainrealtime.WorkerContractRequirements{Required: []domainrealtime.Capability{{
		Contract:     domainrealtime.ContractWorkerRegistry,
		Version:      "v1",
		SubContracts: []domainrealtime.SubContract{domainrealtime.SubContractHeartbeatResponse},
	}}}
}

func (testWorkerScope) AdvertisedCapabilities() domainrealtime.WorkerContractAdvertisement {
	return domainrealtime.WorkerContractAdvertisement{Implemented: []domainrealtime.Capability{testWorkerScope{}.ImplementedWorkerRegistryCapability()}}
}

func TestEnsureAPIRuntimeBinding(t *testing.T) {
	binding := APIRuntimeBinding[string]{
		Transport: testTransport{},
		Client:    "postgres-client",
		Scope:     testAPIScope{},
	}

	if err := EnsureAPIRuntimeBinding(binding); err != nil {
		t.Fatalf("expected compatible api binding, got error: %v", err)
	}
}

func TestEnsureWorkerRuntimeBinding(t *testing.T) {
	binding := WorkerRuntimeBinding[string]{
		Transport: testTransport{},
		Client:    "postgres-client",
		Scope:     testWorkerScope{},
	}

	if err := EnsureWorkerRuntimeBinding(binding); err != nil {
		t.Fatalf("expected compatible worker binding, got error: %v", err)
	}
}

func TestEnsureCrossRuntimeCompatibility_Succeeds(t *testing.T) {
	handshake := CrossRuntimeHandshake{
		APIRequirements:     testAPIScope{}.RequiredCapabilities(),
		APIAdvertisement:    testAPIScope{}.AdvertisedCapabilities(),
		WorkerRequirements:  testWorkerScope{}.RequiredCapabilities(),
		WorkerAdvertisement: testWorkerScope{}.AdvertisedCapabilities(),
	}

	if err := EnsureCrossRuntimeCompatibility(handshake); err != nil {
		t.Fatalf("expected cross-runtime compatibility, got error: %v", err)
	}
}

func TestEnsureCrossRuntimeCompatibility_FailsWhenWorkerRequiresMissingAPISubContract(t *testing.T) {
	handshake := CrossRuntimeHandshake{
		APIRequirements:     testAPIScope{}.RequiredCapabilities(),
		APIAdvertisement:    apiScopeMissingHeartbeatResponse{}.AdvertisedCapabilities(),
		WorkerRequirements:  testWorkerScope{}.RequiredCapabilities(),
		WorkerAdvertisement: testWorkerScope{}.AdvertisedCapabilities(),
	}

	err := EnsureCrossRuntimeCompatibility(handshake)
	if err == nil {
		t.Fatalf("expected Worker->API compatibility error, got nil")
	}
}
