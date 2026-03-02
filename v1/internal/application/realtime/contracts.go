package realtime

import (
	"context"
	"errors"
	"fmt"

	domainrealtime "agentic-orchestrator/internal/domain/realtime"
)

var ErrContractsIncompatible = errors.New("realtime contracts are incompatible")

type Transport[C any] interface {
	Publish(ctx context.Context, client C, message domainrealtime.Message) error
	Subscribe(ctx context.Context, client C, contract domainrealtime.Contract, subContracts []domainrealtime.SubContract, handler func(domainrealtime.Message) error) error
}

type APIWorkerRegistryScope interface {
	RequiredWorkerRegistryCapability() domainrealtime.Capability
	HandleWorkerRegistryMessage(ctx context.Context, message domainrealtime.Message) error
}

type APIScope interface {
	APIWorkerRegistryScope
	RequiredCapabilities() domainrealtime.APIContractRequirements
	AdvertisedCapabilities() domainrealtime.APIContractAdvertisement
}

type WorkerWorkerRegistryScope interface {
	ImplementedWorkerRegistryCapability() domainrealtime.Capability
	HandleWorkerRegistryMessage(ctx context.Context, message domainrealtime.Message) error
}

type WorkerScope interface {
	WorkerWorkerRegistryScope
	RequiredCapabilities() domainrealtime.WorkerContractRequirements
	AdvertisedCapabilities() domainrealtime.WorkerContractAdvertisement
}

type APIRuntimeBinding[C any] struct {
	Transport Transport[C]
	Client    C
	Scope     APIScope
}

type WorkerRuntimeBinding[C any] struct {
	Transport Transport[C]
	Client    C
	Scope     WorkerScope
}

type CrossRuntimeHandshake struct {
	APIRequirements     domainrealtime.APIContractRequirements
	APIAdvertisement    domainrealtime.APIContractAdvertisement
	WorkerRequirements  domainrealtime.WorkerContractRequirements
	WorkerAdvertisement domainrealtime.WorkerContractAdvertisement
}

func EnsureAPIToWorkerCompatibility(apiRequirements domainrealtime.APIContractRequirements, workerAdvertisement domainrealtime.WorkerContractAdvertisement) error {
	result := domainrealtime.EvaluateAPIToWorkerCompatibility(apiRequirements, workerAdvertisement)
	if result.Compatible {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrContractsIncompatible, result.Issues)
}

func EnsureWorkerToAPICompatibility(workerRequirements domainrealtime.WorkerContractRequirements, apiAdvertisement domainrealtime.APIContractAdvertisement) error {
	result := domainrealtime.EvaluateWorkerToAPICompatibility(workerRequirements, apiAdvertisement)
	if result.Compatible {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrContractsIncompatible, result.Issues)
}

func EnsureCrossRuntimeCompatibility(handshake CrossRuntimeHandshake) error {
	if err := EnsureAPIToWorkerCompatibility(handshake.APIRequirements, handshake.WorkerAdvertisement); err != nil {
		return err
	}
	if err := EnsureWorkerToAPICompatibility(handshake.WorkerRequirements, handshake.APIAdvertisement); err != nil {
		return err
	}
	return nil
}

func EnsureAPIRuntimeBinding[C any](binding APIRuntimeBinding[C]) error {
	if binding.Transport == nil {
		return fmt.Errorf("transport is required")
	}
	if binding.Scope == nil {
		return fmt.Errorf("api scope is required")
	}
	if err := binding.Scope.RequiredCapabilities().Validate(); err != nil {
		return err
	}
	if err := binding.Scope.AdvertisedCapabilities().Validate(); err != nil {
		return err
	}
	return nil
}

func EnsureWorkerRuntimeBinding[C any](binding WorkerRuntimeBinding[C]) error {
	if binding.Transport == nil {
		return fmt.Errorf("transport is required")
	}
	if binding.Scope == nil {
		return fmt.Errorf("worker scope is required")
	}
	if err := binding.Scope.RequiredCapabilities().Validate(); err != nil {
		return err
	}
	if err := binding.Scope.AdvertisedCapabilities().Validate(); err != nil {
		return err
	}
	return nil
}
