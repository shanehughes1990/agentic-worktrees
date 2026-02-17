package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectScopeFilesFromDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "02.md"), []byte("# two"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "01.txt"), []byte("one"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignore.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	files, err := CollectScopeFiles(dir)
	if err != nil {
		t.Fatalf("collect scope files: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].Path != "01.txt" || files[1].Path != "02.md" {
		t.Fatalf("expected sorted files, got %+v", files)
	}
}

func TestCollectScopeFilesFromSingleFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "scope.md")
	if err := os.WriteFile(file, []byte("# scope"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	files, err := CollectScopeFiles(file)
	if err != nil {
		t.Fatalf("collect scope files: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
}
