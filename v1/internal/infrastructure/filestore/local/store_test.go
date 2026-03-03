package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLocalStoreRejectsPathsOutsideRoot(t *testing.T) {
	store, err := NewStore(Config{RootDirectory: t.TempDir()})
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	if _, err := store.CreateSignedUploadURL(context.Background(), "../../etc/passwd", "text/plain", zeroTime()); err == nil {
		t.Fatalf("expected path outside root to be rejected")
	}
}

func TestLocalStoreEnsuresArtifactSubfolders(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(Config{RootDirectory: root})
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	if err := store.EnsureArtifactSubfolders("project-1"); err != nil {
		t.Fatalf("ensure artifact subfolders: %v", err)
	}
	directories := []string{
		filepath.Join(root, "projects", "project-1", "artifacts"),
		filepath.Join(root, "projects", "project-1", "repositories", "source"),
		filepath.Join(root, "projects", "project-1", "worktrees"),
	}
	for _, directory := range directories {
		if infoErr := ensureDirectory(directory); infoErr != nil {
			t.Fatalf("expected directory %q: %v", directory, infoErr)
		}
	}
}

func ensureDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory")
	}
	return nil
}

func zeroTime() time.Time {
	return time.Time{}
}
