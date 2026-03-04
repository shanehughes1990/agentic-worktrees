package ingestion

import (
	applicationingestion "agentic-orchestrator/internal/application/ingestion"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGitRepositorySynchronizerSyncFallsBackToUnscopedRepositoryCache(t *testing.T) {
	root := t.TempDir()
	unscopedRepoPath := filepath.Join(root, "unscoped", "repositories", "agentic-worktrees")
	if err := os.MkdirAll(unscopedRepoPath, 0o755); err != nil {
		t.Fatalf("mkdir unscoped cache path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(unscopedRepoPath, "README.md"), []byte("cached"), 0o644); err != nil {
		t.Fatalf("write cached file: %v", err)
	}

	sandbox := t.TempDir()
	synchronizer, err := NewGitRepositorySynchronizer("true", root)
	if err != nil {
		t.Fatalf("new synchronizer: %v", err)
	}

	err = synchronizer.Sync(context.Background(), "agentic", sandbox, "main", []applicationingestion.SourceRepository{{
		RepositoryID:  "agentic-repo-1",
		RepositoryURL: "https://github.com/shanehughes1990/agentic-worktrees",
	}})
	if err != nil {
		t.Fatalf("sync with unscoped fallback: %v", err)
	}

	copiedPath := filepath.Join(sandbox, "repos", "agentic-worktrees", "README.md")
	if _, statErr := os.Stat(copiedPath); statErr != nil {
		t.Fatalf("expected repository to be copied into sandbox from unscoped cache: %v", statErr)
	}
}

func TestGitRepositorySynchronizerSyncReturnsErrorWhenNoCacheExists(t *testing.T) {
	root := t.TempDir()
	sandbox := t.TempDir()
	synchronizer, err := NewGitRepositorySynchronizer("true", root)
	if err != nil {
		t.Fatalf("new synchronizer: %v", err)
	}

	err = synchronizer.Sync(context.Background(), "agentic", sandbox, "main", []applicationingestion.SourceRepository{{
		RepositoryID:  "agentic-repo-1",
		RepositoryURL: "https://github.com/shanehughes1990/agentic-worktrees",
	}})
	if err == nil {
		t.Fatal("expected missing cache error")
	}
}
