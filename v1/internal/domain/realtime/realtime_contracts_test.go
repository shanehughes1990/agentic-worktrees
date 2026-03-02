package realtime

import "testing"

func TestEvaluateAPIToWorkerCompatibility_Success(t *testing.T) {
	requirements := APIContractRequirements{Required: []Capability{{
		Contract:     ContractWorkerRegistry,
		Version:      "v1",
		SubContracts: []SubContract{SubContractHeartbeatRequest},
	}}}
	advertisement := WorkerContractAdvertisement{Implemented: []Capability{{
		Contract:     ContractWorkerRegistry,
		Version:      "v1",
		SubContracts: []SubContract{SubContractHeartbeatRequest, SubContractHeartbeatResponse},
	}}}

	result := EvaluateAPIToWorkerCompatibility(requirements, advertisement)
	if !result.Compatible {
		t.Fatalf("expected compatible=true, got false: %v", result.Issues)
	}
}

func TestEvaluateAPIToWorkerCompatibility_Failure(t *testing.T) {
	requirements := APIContractRequirements{Required: []Capability{{
		Contract:     ContractWorkerRegistry,
		Version:      "v2",
		SubContracts: []SubContract{SubContractHeartbeatRequest, SubContractHeartbeatResponse},
	}}}
	advertisement := WorkerContractAdvertisement{Implemented: []Capability{{
		Contract:     ContractWorkerRegistry,
		Version:      "v1",
		SubContracts: []SubContract{SubContractHeartbeatRequest},
	}}}

	result := EvaluateAPIToWorkerCompatibility(requirements, advertisement)
	if result.Compatible {
		t.Fatalf("expected compatible=false")
	}
	if len(result.Issues) == 0 {
		t.Fatalf("expected issues")
	}
}

func TestEvaluateWorkerToAPICompatibility_Success(t *testing.T) {
	requirements := WorkerContractRequirements{Required: []Capability{{
		Contract:     ContractWorkerRegistry,
		Version:      "v1",
		SubContracts: []SubContract{SubContractHeartbeatResponse},
	}}}
	advertisement := APIContractAdvertisement{Implemented: []Capability{{
		Contract:     ContractWorkerRegistry,
		Version:      "v1",
		SubContracts: []SubContract{SubContractHeartbeatRequest, SubContractHeartbeatResponse},
	}}}

	result := EvaluateWorkerToAPICompatibility(requirements, advertisement)
	if !result.Compatible {
		t.Fatalf("expected compatible=true, got false: %v", result.Issues)
	}
}

func TestEvaluateWorkerToAPICompatibility_Failure(t *testing.T) {
	requirements := WorkerContractRequirements{Required: []Capability{{
		Contract:     ContractWorkerRegistry,
		Version:      "v1",
		SubContracts: []SubContract{SubContractHeartbeatResponse},
	}}}
	advertisement := APIContractAdvertisement{Implemented: []Capability{{
		Contract:     ContractWorkerRegistry,
		Version:      "v1",
		SubContracts: []SubContract{SubContractHeartbeatRequest},
	}}}

	result := EvaluateWorkerToAPICompatibility(requirements, advertisement)
	if result.Compatible {
		t.Fatalf("expected compatible=false")
	}
	if len(result.Issues) == 0 {
		t.Fatalf("expected issues")
	}
}
