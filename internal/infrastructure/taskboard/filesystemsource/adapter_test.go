package filesystemsource

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

func TestAdapterReadMissingFileIsTerminal(t *testing.T) {
	_, err := NewAdapter().Read(context.Background(), domaintaskboard.SourceIdentity{
		Kind:    domaintaskboard.SourceKindFile,
		Locator: filepath.Join(t.TempDir(), "missing.md"),
	})
	if err == nil {
		t.Fatalf("expected read to fail for missing file")
	}
	if !appgitflow.IsTerminalFailure(err) {
		t.Fatalf("expected missing file error to be terminal, got: %v", err)
	}
}

func TestAdapterReadCanceledContextIsTransient(t *testing.T) {
	directory := t.TempDir()
	filePath := filepath.Join(directory, "source.md")
	if err := os.WriteFile(filePath, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write source.md: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewAdapter().Read(ctx, domaintaskboard.SourceIdentity{
		Kind:    domaintaskboard.SourceKindFile,
		Locator: filePath,
	})
	if err == nil {
		t.Fatalf("expected read to fail for canceled context")
	}
	if appgitflow.IsTerminalFailure(err) {
		t.Fatalf("expected canceled context error to be transient, got terminal: %v", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled context error, got: %v", err)
	}
}

func TestAdapterReadDirectoryIsTerminal(t *testing.T) {
	_, err := NewAdapter().Read(context.Background(), domaintaskboard.SourceIdentity{
		Kind:    domaintaskboard.SourceKindFile,
		Locator: t.TempDir(),
	})
	if err == nil {
		t.Fatalf("expected read to fail for directory")
	}
	if !appgitflow.IsTerminalFailure(err) {
		t.Fatalf("expected directory read error to be terminal, got: %v", err)
	}
}

func TestAdapterListMissingFolderIsTerminal(t *testing.T) {
	_, err := NewAdapter().List(context.Background(), domaintaskboard.SourceMetadata{
		Identity: domaintaskboard.SourceIdentity{
			Kind:    domaintaskboard.SourceKindFolder,
			Locator: filepath.Join(t.TempDir(), "missing"),
		},
	}, domaintaskboard.SourceListOptions{})
	if err == nil {
		t.Fatalf("expected list to fail for missing folder")
	}
	if !appgitflow.IsTerminalFailure(err) {
		t.Fatalf("expected missing folder error to be terminal, got: %v", err)
	}
}

func TestAdapterListFilePathIsTerminal(t *testing.T) {
	directory := t.TempDir()
	filePath := filepath.Join(directory, "source.md")
	if err := os.WriteFile(filePath, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write source.md: %v", err)
	}

	_, err := NewAdapter().List(context.Background(), domaintaskboard.SourceMetadata{
		Identity: domaintaskboard.SourceIdentity{
			Kind:    domaintaskboard.SourceKindFolder,
			Locator: filePath,
		},
	}, domaintaskboard.SourceListOptions{})
	if err == nil {
		t.Fatalf("expected list to fail for file source")
	}
	if !appgitflow.IsTerminalFailure(err) {
		t.Fatalf("expected file source list error to be terminal, got: %v", err)
	}
}
