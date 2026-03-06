package ingestion

import (
	applicationingestion "agentic-orchestrator/internal/application/ingestion"
	"context"
	"fmt"
	"os"
	"os/exec"
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

func TestGitRepositorySynchronizerSyncFallsBackToOriginWhenNoCacheExists(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not found")
	}
	root := t.TempDir()
	sandbox := t.TempDir()
	remoteURL := initLocalGitRemote(t, root)
	synchronizer, err := NewGitRepositorySynchronizer("git", root)
	if err != nil {
		t.Fatalf("new synchronizer: %v", err)
	}

	err = synchronizer.Sync(context.Background(), "agentic", sandbox, "main", []applicationingestion.SourceRepository{{
		RepositoryID:  "agentic-repo-1",
		RepositoryURL: remoteURL,
	}})
	if err != nil {
		t.Fatalf("sync fallback to origin clone: %v", err)
	}

	clonedReadme := filepath.Join(sandbox, "repos", "remote", "README.md")
	if _, statErr := os.Stat(clonedReadme); statErr != nil {
		t.Fatalf("expected cloned repository file after origin fallback: %v", statErr)
	}
}

func initLocalGitRemote(t *testing.T, root string) string {
	t.Helper()
	remote := filepath.Join(root, "remote.git")
	work := filepath.Join(root, "work")
	runGit(t, root, "init", "--bare", remote)
	runGit(t, root, "init", work)
	runGit(t, work, "config", "user.email", "test@example.com")
	runGit(t, work, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(work, "README.md"), []byte("origin fallback"), 0o644); err != nil {
		t.Fatalf("write worktree readme: %v", err)
	}
	runGit(t, work, "add", "README.md")
	runGit(t, work, "commit", "-m", "initial commit")
	runGit(t, work, "branch", "-M", "main")
	runGit(t, work, "remote", "add", "origin", remote)
	runGit(t, work, "push", "-u", "origin", "main")
	return remote
}

func runGit(t *testing.T, workingDirectory string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = workingDirectory
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v (%s)", args, err, fmt.Sprintf("%s", output))
	}
}
