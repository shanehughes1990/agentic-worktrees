package realtime

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"agentic-orchestrator/internal/domain/failures"
)

type Role string

const (
	RoleAPI    Role = "api"
	RoleWorker Role = "worker"
)

type DeliverySemantics string

const (
	DeliveryAtLeastOnce DeliverySemantics = "at_least_once"
)

type Contract string

const (
	ContractWorkerRegistry Contract = "worker_registry"
)

type SubContract string

const (
	SubContractHeartbeatRequest  SubContract = "heartbeat.request"
	SubContractHeartbeatResponse SubContract = "heartbeat.response"
)

type CorrelationIDs struct {
	RunID         string
	TaskID        string
	JobID         string
	ProjectID     string
	CorrelationID string
}

func (ids CorrelationIDs) Validate() error {
	if strings.TrimSpace(ids.RunID) == "" {
		return failures.WrapTerminal(errors.New("run_id is required"))
	}
	if strings.TrimSpace(ids.TaskID) == "" {
		return failures.WrapTerminal(errors.New("task_id is required"))
	}
	if strings.TrimSpace(ids.JobID) == "" {
		return failures.WrapTerminal(errors.New("job_id is required"))
	}
	if strings.TrimSpace(ids.CorrelationID) == "" {
		return failures.WrapTerminal(errors.New("correlation_id is required"))
	}
	return nil
}

type Message struct {
	MessageID      string
	Contract       Contract
	SubContract    SubContract
	Version        string
	ProducerRole   Role
	Delivery       DeliverySemantics
	CorrelationIDs CorrelationIDs
	IdempotencyKey string
	OccurredAt     time.Time
	Payload        map[string]any
}

func (message Message) Validate() error {
	if strings.TrimSpace(message.MessageID) == "" {
		return failures.WrapTerminal(errors.New("message_id is required"))
	}
	if err := message.Contract.Validate(); err != nil {
		return err
	}
	if err := message.SubContract.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(message.Version) == "" {
		return failures.WrapTerminal(errors.New("version is required"))
	}
	if err := message.ProducerRole.Validate(); err != nil {
		return err
	}
	if message.Delivery != DeliveryAtLeastOnce {
		return failures.WrapTerminal(errors.New("delivery semantics is invalid"))
	}
	if err := message.CorrelationIDs.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(message.IdempotencyKey) == "" {
		return failures.WrapTerminal(errors.New("idempotency_key is required"))
	}
	if message.OccurredAt.IsZero() {
		return failures.WrapTerminal(errors.New("occurred_at is required"))
	}
	if message.Payload == nil {
		return failures.WrapTerminal(errors.New("payload is required"))
	}
	return nil
}

func (role Role) Validate() error {
	supported := []Role{RoleAPI, RoleWorker}
	if !slices.Contains(supported, role) {
		return failures.WrapTerminal(fmt.Errorf("unsupported realtime role %q", role))
	}
	return nil
}

func (contract Contract) Validate() error {
	supported := []Contract{ContractWorkerRegistry}
	if !slices.Contains(supported, contract) {
		return failures.WrapTerminal(fmt.Errorf("unsupported realtime contract %q", contract))
	}
	return nil
}

func (subContract SubContract) Validate() error {
	supported := []SubContract{
		SubContractHeartbeatRequest,
		SubContractHeartbeatResponse,
	}
	if !slices.Contains(supported, subContract) {
		return failures.WrapTerminal(fmt.Errorf("unsupported realtime sub_contract %q", subContract))
	}
	return nil
}

type Capability struct {
	Contract     Contract
	Version      string
	SubContracts []SubContract
}

func (capability Capability) Validate() error {
	if err := capability.Contract.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(capability.Version) == "" {
		return failures.WrapTerminal(errors.New("capability version is required"))
	}
	if len(capability.SubContracts) == 0 {
		return failures.WrapTerminal(errors.New("at least one sub_contract is required"))
	}
	for _, subContract := range capability.SubContracts {
		if err := subContract.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type APIContractRequirements struct {
	Required []Capability
}

func (requirements APIContractRequirements) Validate() error {
	if len(requirements.Required) == 0 {
		return failures.WrapTerminal(errors.New("at least one required capability is required"))
	}
	for _, capability := range requirements.Required {
		if err := capability.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type WorkerContractAdvertisement struct {
	Implemented []Capability
}

func (advertisement WorkerContractAdvertisement) Validate() error {
	if len(advertisement.Implemented) == 0 {
		return failures.WrapTerminal(errors.New("at least one implemented capability is required"))
	}
	for _, capability := range advertisement.Implemented {
		if err := capability.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type WorkerContractRequirements struct {
	Required []Capability
}

func (requirements WorkerContractRequirements) Validate() error {
	if len(requirements.Required) == 0 {
		return failures.WrapTerminal(errors.New("at least one required capability is required"))
	}
	for _, capability := range requirements.Required {
		if err := capability.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type APIContractAdvertisement struct {
	Implemented []Capability
}

func (advertisement APIContractAdvertisement) Validate() error {
	if len(advertisement.Implemented) == 0 {
		return failures.WrapTerminal(errors.New("at least one implemented capability is required"))
	}
	for _, capability := range advertisement.Implemented {
		if err := capability.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type CompatibilityResult struct {
	Compatible bool
	Issues     []string
}

func EvaluateAPIToWorkerCompatibility(requirements APIContractRequirements, advertisement WorkerContractAdvertisement) CompatibilityResult {
	return evaluateCompatibility(requirements.Required, advertisement.Implemented)
}

func EvaluateWorkerToAPICompatibility(requirements WorkerContractRequirements, advertisement APIContractAdvertisement) CompatibilityResult {
	return evaluateCompatibility(requirements.Required, advertisement.Implemented)
}

func evaluateCompatibility(requiredCapabilities []Capability, implementedCapabilities []Capability) CompatibilityResult {
	issues := make([]string, 0)
	if len(requiredCapabilities) == 0 {
		issues = append(issues, failures.WrapTerminal(errors.New("at least one required capability is required")).Error())
	}
	if len(implementedCapabilities) == 0 {
		issues = append(issues, failures.WrapTerminal(errors.New("at least one implemented capability is required")).Error())
	}
	if len(issues) > 0 {
		return CompatibilityResult{Compatible: false, Issues: issues}
	}

	for _, required := range requiredCapabilities {
		if err := required.Validate(); err != nil {
			issues = append(issues, err.Error())
			continue
		}
		implemented, found := findCapability(implementedCapabilities, required.Contract)
		if !found {
			issues = append(issues, fmt.Sprintf("missing required capability %s", required.Contract))
			continue
		}
		if implemented.Version != required.Version {
			issues = append(issues, fmt.Sprintf("version mismatch for %s: required=%s implemented=%s", required.Contract, required.Version, implemented.Version))
		}
		for _, requiredSubContract := range required.SubContracts {
			if !slices.Contains(implemented.SubContracts, requiredSubContract) {
				issues = append(issues, fmt.Sprintf("missing required sub_contract %s for %s", requiredSubContract, required.Contract))
			}
		}
	}

	return CompatibilityResult{Compatible: len(issues) == 0, Issues: issues}
}

func findCapability(capabilities []Capability, contract Contract) (Capability, bool) {
	for _, capability := range capabilities {
		if err := capability.Validate(); err != nil {
			continue
		}
		if capability.Contract == contract {
			return capability, true
		}
	}
	return Capability{}, false
}
