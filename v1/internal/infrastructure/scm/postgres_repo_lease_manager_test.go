package scm

import (
	applicationscm "agentic-orchestrator/internal/application/scm"
	"agentic-orchestrator/internal/domain/failures"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newLeaseTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func TestPostgresRepoLeaseManagerAcquireAndRelease(t *testing.T) {
	manager, err := NewPostgresRepoLeaseManager(newLeaseTestDB(t))
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	request := applicationscm.RepoLeaseAcquireRequest{CacheKey: domainscm.RepoCacheKey("github/acme/repo"), OwnerID: "owner-1", Token: "id-1"}
	lease, err := manager.Acquire(context.Background(), request)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if err := manager.Release(context.Background(), lease); err != nil {
		t.Fatalf("release: %v", err)
	}
}

func TestPostgresRepoLeaseManagerAcquireIsIdempotent(t *testing.T) {
	manager, err := NewPostgresRepoLeaseManager(newLeaseTestDB(t))
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	request := applicationscm.RepoLeaseAcquireRequest{CacheKey: domainscm.RepoCacheKey("github/acme/repo"), OwnerID: "owner-1", Token: "id-1"}
	if _, err := manager.Acquire(context.Background(), request); err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	if _, err := manager.Acquire(context.Background(), request); err != nil {
		t.Fatalf("second acquire: %v", err)
	}
}

func TestPostgresRepoLeaseManagerRejectsConflictingOwner(t *testing.T) {
	manager, err := NewPostgresRepoLeaseManager(newLeaseTestDB(t))
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	_, _ = manager.Acquire(context.Background(), applicationscm.RepoLeaseAcquireRequest{CacheKey: domainscm.RepoCacheKey("github/acme/repo"), OwnerID: "owner-1", Token: "id-1"})
	_, conflictErr := manager.Acquire(context.Background(), applicationscm.RepoLeaseAcquireRequest{CacheKey: domainscm.RepoCacheKey("github/acme/repo"), OwnerID: "owner-2", Token: "id-2"})
	if !failures.IsClass(conflictErr, failures.ClassTransient) {
		t.Fatalf("expected transient conflict, got %q (%v)", failures.ClassOf(conflictErr), conflictErr)
	}
}
