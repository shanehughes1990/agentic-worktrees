package taskengine

import (
	"context"
	"fmt"
	"strings"
)

type WorkerCapability struct {
	Kind JobKind
}

type WorkerCapabilityAdvertisement struct {
	WorkerID     string
	Capabilities []WorkerCapability
}

func (advertisement WorkerCapabilityAdvertisement) Validate() error {
	if strings.TrimSpace(advertisement.WorkerID) == "" {
		return fmt.Errorf("%w: worker_id is required", ErrInvalidWorkerCapabilityAdvertisement)
	}
	if len(advertisement.Capabilities) == 0 {
		return fmt.Errorf("%w: at least one capability is required", ErrInvalidWorkerCapabilityAdvertisement)
	}

	seen := make(map[JobKind]struct{}, len(advertisement.Capabilities))
	for index, capability := range advertisement.Capabilities {
		if strings.TrimSpace(string(capability.Kind)) == "" {
			return fmt.Errorf("%w: capability kind is required at index %d", ErrInvalidWorkerCapabilityAdvertisement, index)
		}
		if _, exists := seen[capability.Kind]; exists {
			return fmt.Errorf("%w: duplicate capability kind %q", ErrInvalidWorkerCapabilityAdvertisement, capability.Kind)
		}
		seen[capability.Kind] = struct{}{}
	}

	return nil
}

type WorkerCapabilityAdvertiser interface {
	Advertise(ctx context.Context, advertisement WorkerCapabilityAdvertisement) error
}
