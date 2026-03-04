package scm

import (
	"agentic-orchestrator/internal/domain/failures"
	"testing"
)

func TestRepositoryValidateRequiresFields(t *testing.T) {
	err := (Repository{}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestRepoCacheKeyFromRepositoryBuildsNormalizedKey(t *testing.T) {
	key := RepoCacheKeyFromRepository(Repository{Provider: "GitHub", Owner: "Acme", Name: "Repo"})
	if key != RepoCacheKey("github/acme/repo") {
		t.Fatalf("expected github/acme/repo, got %q", key)
	}
}

func TestRepositoryValidateRejectsUnsupportedProvider(t *testing.T) {
	err := (Repository{Provider: "unsupported", Owner: "Acme", Name: "Repo"}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestRepoLeaseValidateRequiresFields(t *testing.T) {
	err := (RepoLease{}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestRepositorySpecValidateRequiresFields(t *testing.T) {
	err := (RepositorySpec{BaseBranch: "main", TargetBranch: "feature/one"}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestRepositorySpecValidateRejectsUnsupportedSyncStrategy(t *testing.T) {
	err := (RepositorySpec{BaseBranch: "main", TargetBranch: "feature/one", Path: "/tmp/repository", SyncStrategy: SyncStrategy("cherry-pick")}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestRepositorySpecEffectiveSyncStrategyDefaultsToMerge(t *testing.T) {
	spec := RepositorySpec{BaseBranch: "main", TargetBranch: "feature/one", Path: "/tmp/repository"}
	if spec.EffectiveSyncStrategy() != SyncStrategyMerge {
		t.Fatalf("expected merge default strategy, got %q", spec.EffectiveSyncStrategy())
	}
}

func TestMergeReadinessValidateRequiresReasonWhenNotMergeable(t *testing.T) {
	err := (MergeReadiness{CanMerge: false}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}
