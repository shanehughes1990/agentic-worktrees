package taskboard

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type NormalizedDocument struct {
	RelativePath string
	Content      string
}

type DocumentNormalizer interface {
	Supports(relativePath string) bool
	Normalize(relativePath string, content []byte) (string, error)
}

type CanonicalUTF8DocumentNormalizer struct{}

func (CanonicalUTF8DocumentNormalizer) Supports(string) bool {
	return true
}

func (CanonicalUTF8DocumentNormalizer) Normalize(_ string, content []byte) (string, error) {
	canonical := strings.ToValidUTF8(string(content), "\uFFFD")
	canonical = strings.TrimPrefix(canonical, "\uFEFF")
	canonical = strings.ReplaceAll(canonical, "\r\n", "\n")
	canonical = strings.ReplaceAll(canonical, "\r", "\n")
	canonical = strings.TrimSpace(canonical)
	return canonical, nil
}

func DefaultDocumentNormalizers() []DocumentNormalizer {
	return []DocumentNormalizer{CanonicalUTF8DocumentNormalizer{}}
}

func NormalizeDirectoryDocuments(directory string, normalizers []DocumentNormalizer) ([]NormalizedDocument, error) {
	cleanDirectory := strings.TrimSpace(directory)
	if cleanDirectory == "" {
		return nil, fmt.Errorf("directory is required")
	}

	if len(normalizers) == 0 {
		normalizers = DefaultDocumentNormalizers()
	}

	paths := make([]string, 0, 32)
	if err := filepath.WalkDir(cleanDirectory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		relativePath, err := filepath.Rel(cleanDirectory, path)
		if err != nil {
			return fmt.Errorf("build relative path for %s: %w", path, err)
		}
		paths = append(paths, filepath.ToSlash(relativePath))
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk directory %s: %w", cleanDirectory, err)
	}

	sort.Strings(paths)
	documents := make([]NormalizedDocument, 0, len(paths))
	for _, relativePath := range paths {
		normalizer := pickDocumentNormalizer(relativePath, normalizers)
		if normalizer == nil {
			continue
		}
		content, err := os.ReadFile(filepath.Join(cleanDirectory, filepath.FromSlash(relativePath)))
		if err != nil {
			return nil, fmt.Errorf("read document %s: %w", relativePath, err)
		}
		normalized, err := normalizer.Normalize(relativePath, content)
		if err != nil {
			return nil, fmt.Errorf("normalize document %s: %w", relativePath, err)
		}
		if strings.TrimSpace(normalized) == "" {
			continue
		}
		documents = append(documents, NormalizedDocument{
			RelativePath: relativePath,
			Content:      normalized,
		})
	}

	return documents, nil
}

func pickDocumentNormalizer(relativePath string, normalizers []DocumentNormalizer) DocumentNormalizer {
	for _, normalizer := range normalizers {
		if normalizer == nil || !normalizer.Supports(relativePath) {
			continue
		}
		return normalizer
	}
	return nil
}
