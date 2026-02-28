package bootstrap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBaseConfigRuntimePathsAreRootedUnderApplicationRootDirectory(t *testing.T) {
	config := BaseConfig{ApplicationRootDirectory: "/tmp/agentic-runtime"}

	if config.RepositoriesPath() != filepath.Join("/tmp/agentic-runtime", "repositories") {
		t.Fatalf("unexpected repositories path: %q", config.RepositoriesPath())
	}
	if config.RepositorySourcePath() != filepath.Join("/tmp/agentic-runtime", "repositories", "source") {
		t.Fatalf("unexpected repository source path: %q", config.RepositorySourcePath())
	}
	if config.WorktreesPath() != filepath.Join("/tmp/agentic-runtime", "worktrees") {
		t.Fatalf("unexpected worktrees path: %q", config.WorktreesPath())
	}
	if config.LogsPath() != filepath.Join("/tmp/agentic-runtime", "logs") {
		t.Fatalf("unexpected logs path: %q", config.LogsPath())
	}
	if config.TrackerPath() != filepath.Join("/tmp/agentic-runtime", "tracker") {
		t.Fatalf("unexpected tracker path: %q", config.TrackerPath())
	}
}

func TestEnsureRuntimeFilesystemCreatesRootedDirectories(t *testing.T) {
	tempDirectory := t.TempDir()
	config := BaseConfig{ApplicationRootDirectory: filepath.Join(tempDirectory, "runtime-root")}

	if err := ensureRuntimeFilesystem(config); err != nil {
		t.Fatalf("ensure runtime filesystem: %v", err)
	}
	expectedDirectories := []string{
		config.ApplicationRootPath(),
		config.RepositoriesPath(),
		config.RepositorySourcePath(),
		config.WorktreesPath(),
		config.LogsPath(),
		config.TrackerPath(),
	}
	for _, directory := range expectedDirectories {
		info, statErr := os.Stat(directory)
		if statErr != nil {
			t.Fatalf("stat %q: %v", directory, statErr)
		}
		if !info.IsDir() {
			t.Fatalf("expected %q to be a directory", directory)
		}
	}
}
