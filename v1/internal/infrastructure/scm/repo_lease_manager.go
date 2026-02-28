package scm

import (
	applicationscm "agentic-orchestrator/internal/application/scm"
	"agentic-orchestrator/internal/domain/failures"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"fmt"
	"sync"
)

type InMemoryRepoLeaseManager struct {
	mu     sync.Mutex
	leases map[domainscm.RepoCacheKey]domainscm.RepoLease
}

func NewInMemoryRepoLeaseManager() *InMemoryRepoLeaseManager {
	return &InMemoryRepoLeaseManager{leases: make(map[domainscm.RepoCacheKey]domainscm.RepoLease)}
}

func (manager *InMemoryRepoLeaseManager) Acquire(ctx context.Context, request applicationscm.RepoLeaseAcquireRequest) (domainscm.RepoLease, error) {
	_ = ctx
	if err := request.Validate(); err != nil {
		return domainscm.RepoLease{}, err
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()

	existing, exists := manager.leases[request.CacheKey]
	if exists {
		if existing.OwnerID == request.OwnerID && existing.Token == request.Token {
			return existing, nil
		}
		return domainscm.RepoLease{}, failures.WrapTransient(fmt.Errorf("repository cache lease already held for %s", request.CacheKey))
	}
	lease := domainscm.RepoLease{CacheKey: request.CacheKey, OwnerID: request.OwnerID, Token: request.Token}
	manager.leases[request.CacheKey] = lease
	return lease, nil
}

func (manager *InMemoryRepoLeaseManager) Release(ctx context.Context, lease domainscm.RepoLease) error {
	_ = ctx
	if err := lease.Validate(); err != nil {
		return err
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	existing, exists := manager.leases[lease.CacheKey]
	if !exists {
		return nil
	}
	if existing.OwnerID == lease.OwnerID && existing.Token == lease.Token {
		delete(manager.leases, lease.CacheKey)
	}
	return nil
}
