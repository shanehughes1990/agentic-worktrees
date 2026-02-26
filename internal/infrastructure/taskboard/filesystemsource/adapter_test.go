package filesystemsource

import (
	"context"
	"os"
	"path/filepath"
	"slices"
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

func TestAdapterListFolderAppliesOptions(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "root.md"), []byte("root"), 0o600); err != nil {
		t.Fatalf("write root.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(directory, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "sub", "keep.txt"), []byte("keep"), 0o600); err != nil {
		t.Fatalf("write keep.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "sub", "skip.tmp"), []byte("skip"), 0o600); err != nil {
		t.Fatalf("write skip.tmp: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(directory, "ignored"), 0o755); err != nil {
		t.Fatalf("mkdir ignored: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "ignored", "doc.md"), []byte("ignored"), 0o600); err != nil {
		t.Fatalf("write ignored/doc.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(directory, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, ".git", "config"), []byte("[core]"), 0o600); err != nil {
		t.Fatalf("write .git/config: %v", err)
	}

	entries, err := NewAdapter().List(context.Background(), domaintaskboard.SourceMetadata{
		Identity: domaintaskboard.SourceIdentity{
			Kind:    domaintaskboard.SourceKindFolder,
			Locator: directory,
		},
	}, domaintaskboard.SourceListOptions{
		WalkDepth:        1,
		IgnorePaths:      []string{"ignored"},
		IgnoreExtensions: []string{".tmp"},
	})
	if err != nil {
		t.Fatalf("list source folder: %v", err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		paths = append(paths, entry.RelativePath)
	}
	if !slices.Equal(paths, []string{"root.md", "sub/keep.txt"}) {
		t.Fatalf("unexpected listed paths: %#v", paths)
	}
	for _, entry := range entries {
		if entry.Metadata == nil {
			t.Fatalf("expected metadata to be populated for %s", entry.RelativePath)
		}
		metadataRelativePath, ok := entry.Metadata.Attributes["relative_path"].(string)
		if !ok || metadataRelativePath != entry.RelativePath {
			t.Fatalf("expected metadata relative path %s, got %#v", entry.RelativePath, entry.Metadata.Attributes["relative_path"])
		}
	}
}
