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

func NormalizeSourceDocuments(sourcePath string, sourceType IngestionSourceType, folderOptions FolderTraversalOptions, normalizers []DocumentNormalizer) ([]NormalizedDocument, error) {
	cleanSourcePath := strings.TrimSpace(sourcePath)
	if cleanSourcePath == "" {
		return nil, fmt.Errorf("source path is required")
	}

	if len(normalizers) == 0 {
		normalizers = DefaultDocumentNormalizers()
	}

	switch sourceType {
	case IngestionSourceTypeFile:
		return normalizeSingleDocument(cleanSourcePath, filepath.ToSlash(filepath.Base(cleanSourcePath)), normalizers)
	case IngestionSourceTypeFolder:
		return NormalizeDirectoryDocumentsWithOptions(cleanSourcePath, folderOptions, normalizers)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", sourceType)
	}
}

func NormalizeDirectoryDocuments(directory string, normalizers []DocumentNormalizer) ([]NormalizedDocument, error) {
	return NormalizeDirectoryDocumentsWithOptions(directory, FolderTraversalOptions{WalkDepth: -1}, normalizers)
}

func NormalizeDirectoryDocumentsWithOptions(directory string, options FolderTraversalOptions, normalizers []DocumentNormalizer) ([]NormalizedDocument, error) {
	cleanDirectory := strings.TrimSpace(directory)
	if cleanDirectory == "" {
		return nil, fmt.Errorf("directory is required")
	}

	cleanWalkDepth := options.WalkDepth
	cleanIgnorePaths := normalizeIgnorePaths(options.IgnorePaths)
	cleanIgnoreExtensions := normalizeIgnoreExtensions(options.IgnoreExtensions)

	paths := make([]string, 0, 32)
	if err := filepath.WalkDir(cleanDirectory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(cleanDirectory, path)
		if err != nil {
			return fmt.Errorf("build relative path for %s: %w", path, err)
		}
		relativePath = filepath.ToSlash(relativePath)
		if relativePath == "." {
			relativePath = ""
		}

		if strings.TrimSpace(relativePath) != "" && shouldIgnorePath(relativePath, cleanIgnorePaths) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if cleanWalkDepth >= 0 && pathDepth(relativePath) > cleanWalkDepth {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldIgnoreExtension(relativePath, cleanIgnoreExtensions) {
			return nil
		}
		paths = append(paths, relativePath)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk directory %s: %w", cleanDirectory, err)
	}

	sort.Strings(paths)
	documents := make([]NormalizedDocument, 0, len(paths))
	for _, relativePath := range paths {
		normalizedDocuments, err := normalizeSingleDocument(filepath.Join(cleanDirectory, filepath.FromSlash(relativePath)), relativePath, normalizers)
		if err != nil {
			return nil, err
		}
		documents = append(documents, normalizedDocuments...)
	}

	return documents, nil
}

func normalizeSingleDocument(absolutePath string, relativePath string, normalizers []DocumentNormalizer) ([]NormalizedDocument, error) {
	normalizer := pickDocumentNormalizer(relativePath, normalizers)
	if normalizer == nil {
		return nil, nil
	}
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("read document %s: %w", relativePath, err)
	}
	normalized, err := normalizer.Normalize(relativePath, content)
	if err != nil {
		return nil, fmt.Errorf("normalize document %s: %w", relativePath, err)
	}
	if strings.TrimSpace(normalized) == "" {
		return nil, nil
	}
	return []NormalizedDocument{{
		RelativePath: relativePath,
		Content:      normalized,
	}}, nil
}

func normalizeIgnorePaths(ignorePaths []string) []string {
	normalized := make([]string, 0, len(ignorePaths))
	for _, ignorePath := range ignorePaths {
		cleanPath := strings.TrimSpace(ignorePath)
		if cleanPath == "" {
			continue
		}
		cleanPath = filepath.ToSlash(filepath.Clean(cleanPath))
		cleanPath = strings.TrimPrefix(cleanPath, "./")
		cleanPath = strings.TrimPrefix(cleanPath, "/")
		if cleanPath == "" || cleanPath == "." || cleanPath == ".." {
			continue
		}
		normalized = append(normalized, cleanPath)
	}
	sort.Strings(normalized)
	return normalized
}

func normalizeIgnoreExtensions(ignoreExtensions []string) map[string]struct{} {
	normalized := make(map[string]struct{}, len(ignoreExtensions))
	for _, ignoreExtension := range ignoreExtensions {
		cleanExtension := strings.ToLower(strings.TrimSpace(ignoreExtension))
		if cleanExtension == "" {
			continue
		}
		if !strings.HasPrefix(cleanExtension, ".") {
			cleanExtension = "." + cleanExtension
		}
		normalized[cleanExtension] = struct{}{}
	}
	return normalized
}

func shouldIgnorePath(relativePath string, ignorePaths []string) bool {
	cleanRelativePath := strings.Trim(strings.TrimSpace(filepath.ToSlash(relativePath)), "/")
	if cleanRelativePath == "" {
		return false
	}
	for _, ignorePath := range ignorePaths {
		if cleanRelativePath == ignorePath || strings.HasPrefix(cleanRelativePath, ignorePath+"/") {
			return true
		}
	}
	return false
}

func shouldIgnoreExtension(relativePath string, ignoreExtensions map[string]struct{}) bool {
	if len(ignoreExtensions) == 0 {
		return false
	}
	cleanExtension := strings.ToLower(filepath.Ext(strings.TrimSpace(relativePath)))
	if cleanExtension == "" {
		return false
	}
	_, ignored := ignoreExtensions[cleanExtension]
	return ignored
}

func pathDepth(relativePath string) int {
	cleanRelativePath := strings.Trim(strings.TrimSpace(filepath.ToSlash(relativePath)), "/")
	if cleanRelativePath == "" {
		return 0
	}
	return strings.Count(cleanRelativePath, "/")
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
