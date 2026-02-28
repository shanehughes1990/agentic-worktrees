package scm

import (
	applicationscm "agentic-orchestrator/internal/application/scm"
	"agentic-orchestrator/internal/domain/failures"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"testing"
)

func TestInMemoryRepoLeaseManagerAcquireAndRelease(t *testing.T) {
	manager := NewInMemoryRepoLeaseManager()
	request := applicationscm.RepoLeaseAcquireRequest{CacheKey: domainscm.RepoCacheKey("github/acme/repo"), OwnerID: "run-1:task-1:feature-one", Token: "id-1"}

	lease, err := manager.Acquire(context.Background(), request)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if err := manager.Release(context.Background(), lease); err != nil {
		t.Fatalf("release: %v", err)
	}
}

func TestInMemoryRepoLeaseManagerRejectsConflictingOwner(t *testing.T) {
	manager := NewInMemoryRepoLeaseManager()
	_, err := manager.Acquire(context.Background(), applicationscm.RepoLeaseAcquireRequest{CacheKey: domainscm.RepoCacheKey("github/acme/repo"), OwnerID: "owner-1", Token: "id-1"})
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	_, secondErr := manager.Acquire(context.Background(), applicationscm.RepoLeaseAcquireRequest{CacheKey: domainscm.RepoCacheKey("github/acme/repo"), OwnerID: "owner-2", Token: "id-2"})
	if !failures.IsClass(secondErr, failures.ClassTransient) {
		t.Fatalf("expected transient conflict, got %q (%v)", failures.ClassOf(secondErr), secondErr)
	}
}
