package filesystemsource

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

type Adapter struct{}

type failureClass string

const (
	failureClassTransient failureClass = "transient"
	failureClassTerminal  failureClass = "terminal"
)

type classifiedError struct {
	class failureClass
	err   error
}

func (err *classifiedError) Error() string {
	if err == nil || err.err == nil {
		return "classified error"
	}
	return err.err.Error()
}

func (err *classifiedError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.err
}

func (err *classifiedError) FailureClass() string {
	if err == nil {
		return ""
	}
	return string(err.class)
}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (adapter *Adapter) List(ctx context.Context, source domaintaskboard.SourceMetadata, options domaintaskboard.SourceListOptions) ([]domaintaskboard.SourceListEntry, error) {
	if err := source.ValidateBasics(); err != nil {
		return nil, wrapTerminal(err)
	}
	if source.Identity.Kind != domaintaskboard.SourceKindFolder {
		return nil, wrapTerminal(fmt.Errorf("source kind must be folder"))
	}

	cleanDirectory := strings.TrimSpace(source.Identity.Locator)
	if cleanDirectory == "" {
		return nil, wrapTerminal(fmt.Errorf("source locator is required"))
	}
	info, err := os.Stat(cleanDirectory)
	if err != nil {
		return nil, classifyFilesystemError(fmt.Errorf("stat source folder %s: %w", cleanDirectory, err))
	}
	if !info.IsDir() {
		return nil, wrapTerminal(fmt.Errorf("source locator must be a directory"))
	}
	cleanWalkDepth := options.WalkDepth
	cleanIgnorePaths := normalizeIgnorePaths(options.IgnorePaths)
	cleanIgnoreExtensions := normalizeIgnoreExtensions(options.IgnoreExtensions)

	entries := make([]domaintaskboard.SourceListEntry, 0, 32)
	if err := filepath.WalkDir(cleanDirectory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return classifyFilesystemError(walkErr)
		}
		if ctx != nil {
			select {
			case <-ctx.Done():
				return wrapTransient(ctx.Err())
			default:
			}
		}

		relativePath, err := filepath.Rel(cleanDirectory, path)
		if err != nil {
			return wrapTerminal(fmt.Errorf("build relative path for %s: %w", path, err))
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
		entries = append(entries, domaintaskboard.SourceListEntry{
			Identity: domaintaskboard.SourceIdentity{
				Kind:    domaintaskboard.SourceKindFile,
				Locator: path,
			},
			RelativePath: relativePath,
		})
		return nil
	}); err != nil {
		return nil, classifyFilesystemError(fmt.Errorf("walk directory %s: %w", cleanDirectory, err))
	}

	sort.Slice(entries, func(i int, j int) bool {
		if entries[i].RelativePath == entries[j].RelativePath {
			return entries[i].Identity.Locator < entries[j].Identity.Locator
		}
		return entries[i].RelativePath < entries[j].RelativePath
	})
	return entries, nil
}

func (adapter *Adapter) Read(ctx context.Context, source domaintaskboard.SourceIdentity) ([]byte, error) {
	if err := source.ValidateBasics(); err != nil {
		return nil, wrapTerminal(err)
	}
	if source.Kind != domaintaskboard.SourceKindFile {
		return nil, wrapTerminal(fmt.Errorf("source kind must be file"))
	}
	if ctx != nil {
		select {
		case <-ctx.Done():
			return nil, wrapTransient(ctx.Err())
		default:
		}
	}

	cleanLocator := strings.TrimSpace(source.Locator)
	info, err := os.Stat(cleanLocator)
	if err != nil {
		return nil, classifyFilesystemError(fmt.Errorf("stat source %s: %w", cleanLocator, err))
	}
	if info.IsDir() {
		return nil, wrapTerminal(fmt.Errorf("source locator must be a file"))
	}

	content, err := os.ReadFile(cleanLocator)
	if err != nil {
		return nil, classifyFilesystemError(fmt.Errorf("read source %s: %w", cleanLocator, err))
	}
	return content, nil
}

func (adapter *Adapter) ResolveWorkingDirectory(ctx context.Context, source domaintaskboard.SourceIdentity) (string, error) {
	if err := source.ValidateBasics(); err != nil {
		return "", wrapTerminal(err)
	}
	if ctx != nil {
		select {
		case <-ctx.Done():
			return "", wrapTransient(ctx.Err())
		default:
		}
	}

	cleanLocator := strings.TrimSpace(source.Locator)
	switch source.Kind {
	case domaintaskboard.SourceKindFolder:
		return cleanLocator, nil
	case domaintaskboard.SourceKindFile:
		return filepath.Dir(cleanLocator), nil
	default:
		return "", wrapTerminal(fmt.Errorf("source kind must be file or folder"))
	}
}

func classifyFilesystemError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return ensureClassified(err, failureClassTransient)
	}
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrPermission) || errors.Is(err, fs.ErrInvalid) {
		return ensureClassified(err, failureClassTerminal)
	}
	return ensureClassified(err, failureClassTransient)
}

func wrapTerminal(err error) error {
	if err == nil {
		return nil
	}
	return &classifiedError{class: failureClassTerminal, err: err}
}

func wrapTransient(err error) error {
	if err == nil {
		return nil
	}
	return &classifiedError{class: failureClassTransient, err: err}
}

func ensureClassified(err error, defaultClass failureClass) error {
	if err == nil {
		return nil
	}
	current := err
	for current != nil {
		_, ok := current.(*classifiedError)
		if ok {
			return err
		}
		wrapped, ok := current.(interface{ Unwrap() error })
		if !ok {
			break
		}
		current = wrapped.Unwrap()
	}
	if defaultClass == failureClassTerminal {
		return wrapTerminal(fmt.Errorf("%w", err))
	}
	return wrapTransient(fmt.Errorf("%w", err))
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
