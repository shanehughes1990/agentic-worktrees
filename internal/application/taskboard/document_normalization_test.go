package taskboard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeDirectoryDocumentsCanonicalUTF8AndStableOrder(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("second\r\nline"), 0o600); err != nil {
		t.Fatalf("write b.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte{0xEF, 0xBB, 0xBF, 'f', 'i', 'r', 's', 't', '\n', 0xFF}, 0o600); err != nil {
		t.Fatalf("write a.txt: %v", err)
	}

	documents, err := NormalizeDirectoryDocuments(dir, DefaultDocumentNormalizers())
	if err != nil {
		t.Fatalf("unexpected normalization error: %v", err)
	}
	if len(documents) != 2 {
		t.Fatalf("expected 2 normalized documents, got %d", len(documents))
	}
	if documents[0].RelativePath != "a.txt" || documents[1].RelativePath != "b.txt" {
		t.Fatalf("expected stable lexical ordering, got %#v", documents)
	}
	if strings.Contains(documents[0].Content, "\r") {
		t.Fatalf("expected canonical newlines, got %q", documents[0].Content)
	}
	if strings.Contains(documents[0].Content, "\uFEFF") {
		t.Fatalf("expected bom stripped, got %q", documents[0].Content)
	}
}

type extensionNormalizer struct {
	extension string
	value     string
}

func (normalizer extensionNormalizer) Supports(relativePath string) bool {
	return strings.HasSuffix(relativePath, normalizer.extension)
}

func (normalizer extensionNormalizer) Normalize(string, []byte) (string, error) {
	return normalizer.value, nil
}

func TestNormalizeDirectoryDocumentsUsesFirstMatchingNormalizer(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "doc.md"), []byte("unused"), 0o600); err != nil {
		t.Fatalf("write doc.md: %v", err)
	}

	documents, err := NormalizeDirectoryDocuments(dir, []DocumentNormalizer{
		extensionNormalizer{extension: ".md", value: "markdown"},
		CanonicalUTF8DocumentNormalizer{},
	})
	if err != nil {
		t.Fatalf("unexpected normalization error: %v", err)
	}
	if len(documents) != 1 {
		t.Fatalf("expected 1 normalized document, got %d", len(documents))
	}
	if documents[0].Content != "markdown" {
		t.Fatalf("expected first matching normalizer result, got %q", documents[0].Content)
	}
}
