package filesystem

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

type Adapter struct{}

var (
	_ domaintaskboard.SourceLister = (*Adapter)(nil)
	_ domaintaskboard.SourceReader = (*Adapter)(nil)
)

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (adapter *Adapter) List(ctx context.Context, source domaintaskboard.SourceMetadata, options domaintaskboard.SourceListOptions) ([]domaintaskboard.SourceListEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := source.ValidateBasics(); err != nil {
		return nil, err
	}

	switch source.Identity.Kind {
	case domaintaskboard.SourceKindFile:
		return adapter.listFile(source.Identity)
	case domaintaskboard.SourceKindFolder:
		return adapter.listFolder(ctx, source.Identity, options)
	default:
		return nil, fmt.Errorf("unsupported source kind: %s", source.Identity.Kind)
	}
}

func (adapter *Adapter) Read(ctx context.Context, source domaintaskboard.SourceIdentity) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := source.ValidateBasics(); err != nil {
		return nil, err
	}
	if source.Kind != domaintaskboard.SourceKindFile {
		return nil, fmt.Errorf("source kind must be file")
	}

	info, err := os.Stat(strings.TrimSpace(source.Locator))
	if err != nil {
		return nil, fmt.Errorf("stat source file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("source locator must be a file")
	}

	content, err := os.ReadFile(strings.TrimSpace(source.Locator))
	if err != nil {
		return nil, fmt.Errorf("read source file: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return content, nil
}

func (adapter *Adapter) listFile(identity domaintaskboard.SourceIdentity) ([]domaintaskboard.SourceListEntry, error) {
	cleanLocator := strings.TrimSpace(identity.Locator)
	info, err := os.Stat(cleanLocator)
	if err != nil {
		return nil, fmt.Errorf("stat source file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("source locator must be a file")
	}
	fileIdentity := domaintaskboard.SourceIdentity{
		Kind:    domaintaskboard.SourceKindFile,
		Locator: cleanLocator,
	}

	return []domaintaskboard.SourceListEntry{
		{
			Identity:     fileIdentity,
			RelativePath: filepath.ToSlash(filepath.Base(cleanLocator)),
			Metadata:     mapFilesystemObjectMetadata(fileIdentity, filepath.ToSlash(filepath.Base(cleanLocator)), info),
		},
	}, nil
}

func (adapter *Adapter) listFolder(ctx context.Context, identity domaintaskboard.SourceIdentity, options domaintaskboard.SourceListOptions) ([]domaintaskboard.SourceListEntry, error) {
	cleanLocator := strings.TrimSpace(identity.Locator)
	info, err := os.Stat(cleanLocator)
	if err != nil {
		return nil, fmt.Errorf("stat source folder: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source locator must be a directory")
	}

	cleanWalkDepth := options.WalkDepth
	cleanIgnorePaths := normalizeIgnorePaths(options.IgnorePaths)
	cleanIgnoreExtensions := normalizeIgnoreExtensions(options.IgnoreExtensions)

	entries := make([]domaintaskboard.SourceListEntry, 0, 32)
	if err := filepath.WalkDir(cleanLocator, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		relativePath, err := filepath.Rel(cleanLocator, path)
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
		fileIdentity := domaintaskboard.SourceIdentity{
			Kind:    domaintaskboard.SourceKindFile,
			Locator: filepath.Join(cleanLocator, filepath.FromSlash(relativePath)),
		}
		fileInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("load file info for %s: %w", path, err)
		}
		entries = append(entries, domaintaskboard.SourceListEntry{
			Identity:     fileIdentity,
			RelativePath: relativePath,
			Metadata:     mapFilesystemObjectMetadata(fileIdentity, relativePath, fileInfo),
		})
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk source folder %s: %w", cleanLocator, err)
	}

	sort.Slice(entries, func(i int, j int) bool {
		if entries[i].RelativePath == entries[j].RelativePath {
			return entries[i].Identity.Locator < entries[j].Identity.Locator
		}
		return entries[i].RelativePath < entries[j].RelativePath
	})
	return entries, nil
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

func mapFilesystemObjectMetadata(identity domaintaskboard.SourceIdentity, relativePath string, info fs.FileInfo) *domaintaskboard.SourceMetadata {
	attributes := map[string]any{
		"relative_path": relativePath,
	}
	if info != nil {
		attributes["size_bytes"] = info.Size()
		attributes["mode"] = info.Mode().String()
		attributes["mod_time_unix"] = info.ModTime().UTC().Unix()
	}
	return &domaintaskboard.SourceMetadata{
		Identity:   identity,
		Attributes: attributes,
	}
}
