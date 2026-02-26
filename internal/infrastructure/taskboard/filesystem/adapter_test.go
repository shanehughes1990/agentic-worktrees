package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

func TestAdapterListFileReturnsSingleEntry(t *testing.T) {
	directory := t.TempDir()
	filePath := filepath.Join(directory, "scope.md")
	if err := os.WriteFile(filePath, []byte("scope"), 0o600); err != nil {
		t.Fatalf("write scope.md: %v", err)
	}

	entries, err := NewAdapter().List(context.Background(), domaintaskboard.SourceMetadata{
		Identity: domaintaskboard.SourceIdentity{
			Kind:    domaintaskboard.SourceKindFile,
			Locator: filePath,
		},
	}, domaintaskboard.SourceListOptions{})
	if err != nil {
		t.Fatalf("list source file: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].RelativePath != "scope.md" {
		t.Fatalf("expected relative path scope.md, got %s", entries[0].RelativePath)
	}
	if entries[0].Identity.Locator != filePath {
		t.Fatalf("expected locator %s, got %s", filePath, entries[0].Identity.Locator)
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
}

func TestAdapterReadFile(t *testing.T) {
	directory := t.TempDir()
	filePath := filepath.Join(directory, "source.md")
	if err := os.WriteFile(filePath, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write source.md: %v", err)
	}

	content, err := NewAdapter().Read(context.Background(), domaintaskboard.SourceIdentity{
		Kind:    domaintaskboard.SourceKindFile,
		Locator: filePath,
	})
	if err != nil {
		t.Fatalf("read source file: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("unexpected source content: %q", string(content))
	}
}

func TestAdapterReadRejectsFolderIdentity(t *testing.T) {
	_, err := NewAdapter().Read(context.Background(), domaintaskboard.SourceIdentity{
		Kind:    domaintaskboard.SourceKindFolder,
		Locator: "/tmp",
	})
	if err == nil {
		t.Fatalf("expected read to reject folder identity")
	}
}
