package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentationLoaderLoadDocumentationFiles(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.md"), []byte("A"), 0o644); err != nil {
		t.Fatalf("write a.md: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "nested", "b.txt"), []byte("B"), 0o644); err != nil {
		t.Fatalf("write b.txt: %v", err)
	}

	loader := NewDocumentationLoader()
	documents, err := loader.LoadDocumentationFiles(context.Background(), tmpDir, 2)
	if err != nil {
		t.Fatalf("load docs: %v", err)
	}
	if len(documents) != 2 {
		t.Fatalf("expected 2 docs, got %d", len(documents))
	}
}
