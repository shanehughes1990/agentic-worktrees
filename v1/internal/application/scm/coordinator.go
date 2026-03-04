package scm

import (
	"agentic-orchestrator/internal/domain/failures"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"errors"
	"fmt"
	"strings"
)

type RepoLeaseAcquireRequest struct {
	CacheKey domainscm.RepoCacheKey
	OwnerID  string
	Token    string
}

func (request RepoLeaseAcquireRequest) Validate() error {
	if err := request.CacheKey.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(request.OwnerID) == "" {
		return failures.WrapTerminal(errors.New("owner_id is required"))
	}
	if strings.TrimSpace(request.Token) == "" {
		return failures.WrapTerminal(errors.New("token is required"))
	}
	return nil
}

type RepoLeaseManager interface {
	Acquire(ctx context.Context, request RepoLeaseAcquireRequest) (domainscm.RepoLease, error)
	Release(ctx context.Context, lease domainscm.RepoLease) error
}

type EnsureRepositoryCoordinator struct {
	orchestrator domainscm.Orchestrator
	leaseManager RepoLeaseManager
}

func NewEnsureRepositoryCoordinator(orchestrator domainscm.Orchestrator, leaseManager RepoLeaseManager) (*EnsureRepositoryCoordinator, error) {
	if orchestrator == nil {
		return nil, failures.WrapTerminal(errors.New("scm orchestrator is required"))
	}
	return &EnsureRepositoryCoordinator{orchestrator: orchestrator, leaseManager: leaseManager}, nil
}

func (coordinator *EnsureRepositoryCoordinator) Ensure(ctx context.Context, request EnsureRepositoryRequest) (domainscm.RepositoryState, error) {
	if coordinator == nil || coordinator.orchestrator == nil {
		return domainscm.RepositoryState{}, failures.WrapTerminal(errors.New("ensure repository coordinator is not initialized"))
	}
	if err := request.Validate(); err != nil {
		return domainscm.RepositoryState{}, err
	}
	strategy := request.Spec.EffectiveSyncStrategy()
	if strategy == domainscm.SyncStrategyRebase {
		panic("scm rebase sync strategy is not implemented")
	}
	request.Spec.SyncStrategy = strategy

	lease, acquired, err := coordinator.acquireLease(ctx, request)
	if err != nil {
		return domainscm.RepositoryState{}, err
	}
	if acquired {
		defer func() {
			_ = coordinator.leaseManager.Release(ctx, lease)
		}()
	}

	state, ensureErr := coordinator.orchestrator.EnsureRepository(ctx, request.Repository, request.Spec)
	if ensureErr != nil {
		return domainscm.RepositoryState{}, ensureErr
	}
	if err := state.Validate(); err != nil {
		return domainscm.RepositoryState{}, err
	}
	return state, nil
}

func (coordinator *EnsureRepositoryCoordinator) acquireLease(ctx context.Context, request EnsureRepositoryRequest) (domainscm.RepoLease, bool, error) {
	if coordinator.leaseManager == nil {
		return domainscm.RepoLease{}, false, nil
	}
	acquireRequest := RepoLeaseAcquireRequest{
		CacheKey: domainscm.RepoCacheKeyFromRepository(request.Repository),
		OwnerID:  fmt.Sprintf("%s:%s:%s", request.Metadata.CorrelationIDs.RunID, request.Metadata.CorrelationIDs.TaskID, request.Spec.TargetBranch),
		Token:    request.Metadata.IdempotencyKey,
	}
	if err := acquireRequest.Validate(); err != nil {
		return domainscm.RepoLease{}, false, err
	}
	lease, err := coordinator.leaseManager.Acquire(ctx, acquireRequest)
	if err != nil {
		return domainscm.RepoLease{}, false, err
	}
	if err := lease.Validate(); err != nil {
		return domainscm.RepoLease{}, false, err
	}
	return lease, true, nil
}
