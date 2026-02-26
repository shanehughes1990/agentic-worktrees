package filesystemsource

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

func TestAdapterListMapsFilesystemTraversalObjectMetadata(t *testing.T) {
	directory := t.TempDir()
	filePath := filepath.Join(directory, "scope.md")
	if err := os.WriteFile(filePath, []byte("scope"), 0o600); err != nil {
		t.Fatalf("write scope.md: %v", err)
	}

	entries, err := NewAdapter().List(context.Background(), domaintaskboard.SourceMetadata{
		Identity: domaintaskboard.SourceIdentity{
			Kind:    domaintaskboard.SourceKindFolder,
			Locator: directory,
		},
	}, domaintaskboard.SourceListOptions{})
	if err != nil {
		t.Fatalf("list source folder: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].RelativePath != "scope.md" {
		t.Fatalf("expected relative path scope.md, got %s", entries[0].RelativePath)
	}
	if entries[0].Metadata == nil {
		t.Fatalf("expected metadata to be populated")
	}
	if entries[0].Metadata.Identity.Locator != filePath {
		t.Fatalf("expected metadata locator %s, got %s", filePath, entries[0].Metadata.Identity.Locator)
	}
	relativePath, ok := entries[0].Metadata.Attributes["relative_path"].(string)
	if !ok || relativePath != "scope.md" {
		t.Fatalf("expected metadata relative path scope.md, got %#v", entries[0].Metadata.Attributes["relative_path"])
	}
	if _, ok := entries[0].Metadata.Attributes["size_bytes"].(int64); !ok {
		t.Fatalf("expected metadata size_bytes to be int64, got %#v", entries[0].Metadata.Attributes["size_bytes"])
	}
}
