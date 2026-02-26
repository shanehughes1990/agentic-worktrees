package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
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
		return nil, appgitflow.WrapTransient(err)
	}
	if err := source.ValidateBasics(); err != nil {
		return nil, appgitflow.WrapTerminal(err)
	}

	switch source.Identity.Kind {
	case domaintaskboard.SourceKindFile:
		return adapter.listFile(source.Identity)
	case domaintaskboard.SourceKindFolder:
		return adapter.listFolder(ctx, source.Identity, options)
	default:
		return nil, appgitflow.WrapTerminal(fmt.Errorf("unsupported source kind: %s", source.Identity.Kind))
	}
}

func (adapter *Adapter) Read(ctx context.Context, source domaintaskboard.SourceIdentity) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, appgitflow.WrapTransient(err)
	}
	if err := source.ValidateBasics(); err != nil {
		return nil, appgitflow.WrapTerminal(err)
	}
	if source.Kind != domaintaskboard.SourceKindFile {
		return nil, appgitflow.WrapTerminal(fmt.Errorf("source kind must be file"))
	}

	info, err := os.Stat(strings.TrimSpace(source.Locator))
	if err != nil {
		return nil, classifyFilesystemError(fmt.Errorf("stat source file: %w", err))
	}
	if info.IsDir() {
		return nil, appgitflow.WrapTerminal(fmt.Errorf("source locator must be a file"))
	}

	content, err := os.ReadFile(strings.TrimSpace(source.Locator))
	if err != nil {
		return nil, classifyFilesystemError(fmt.Errorf("read source file: %w", err))
	}
	if err := ctx.Err(); err != nil {
		return nil, appgitflow.WrapTransient(err)
	}
	return content, nil
}

func (adapter *Adapter) listFile(identity domaintaskboard.SourceIdentity) ([]domaintaskboard.SourceListEntry, error) {
	cleanLocator := strings.TrimSpace(identity.Locator)
	info, err := os.Stat(cleanLocator)
	if err != nil {
		return nil, classifyFilesystemError(fmt.Errorf("stat source file: %w", err))
	}
	if info.IsDir() {
		return nil, appgitflow.WrapTerminal(fmt.Errorf("source locator must be a file"))
	}

	return []domaintaskboard.SourceListEntry{
		{
			Identity: domaintaskboard.SourceIdentity{
				Kind:    domaintaskboard.SourceKindFile,
				Locator: cleanLocator,
			},
			RelativePath: filepath.ToSlash(filepath.Base(cleanLocator)),
		},
	}, nil
}

func (adapter *Adapter) listFolder(ctx context.Context, identity domaintaskboard.SourceIdentity, options domaintaskboard.SourceListOptions) ([]domaintaskboard.SourceListEntry, error) {
	cleanLocator := strings.TrimSpace(identity.Locator)
	info, err := os.Stat(cleanLocator)
	if err != nil {
		return nil, classifyFilesystemError(fmt.Errorf("stat source folder: %w", err))
	}
	if !info.IsDir() {
		return nil, appgitflow.WrapTerminal(fmt.Errorf("source locator must be a directory"))
	}

	cleanWalkDepth := options.WalkDepth
	cleanIgnorePaths := normalizeIgnorePaths(options.IgnorePaths)
	cleanIgnoreExtensions := normalizeIgnoreExtensions(options.IgnoreExtensions)

	relativePaths := make([]string, 0, 32)
	if err := filepath.WalkDir(cleanLocator, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return classifyFilesystemError(walkErr)
		}
		if err := ctx.Err(); err != nil {
			return appgitflow.WrapTransient(err)
		}

		relativePath, err := filepath.Rel(cleanLocator, path)
		if err != nil {
			return appgitflow.WrapTerminal(fmt.Errorf("build relative path for %s: %w", path, err))
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
		relativePaths = append(relativePaths, relativePath)
		return nil
	}); err != nil {
		return nil, classifyFilesystemError(fmt.Errorf("walk source folder %s: %w", cleanLocator, err))
	}

	sort.Strings(relativePaths)
	entries := make([]domaintaskboard.SourceListEntry, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		if err := ctx.Err(); err != nil {
			return nil, appgitflow.WrapTransient(err)
		}
		entries = append(entries, domaintaskboard.SourceListEntry{
			Identity: domaintaskboard.SourceIdentity{
				Kind:    domaintaskboard.SourceKindFile,
				Locator: filepath.Join(cleanLocator, filepath.FromSlash(relativePath)),
			},
			RelativePath: relativePath,
		})
	}
	return entries, nil
}

func classifyFilesystemError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return appgitflow.EnsureClassified(err, appgitflow.FailureClassTransient)
	}
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrPermission) || errors.Is(err, fs.ErrInvalid) {
		return appgitflow.EnsureClassified(err, appgitflow.FailureClassTerminal)
	}
	return appgitflow.EnsureClassified(err, appgitflow.FailureClassTransient)
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
