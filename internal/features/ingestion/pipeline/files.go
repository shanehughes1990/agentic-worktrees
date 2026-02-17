package pipeline

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxScopeFileBytes = 2 * 1024 * 1024

var supportedScopeExtensions = map[string]struct{}{
	".md":       {},
	".markdown": {},
	".txt":      {},
	".rst":      {},
}

type ScopeFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func CollectScopeFiles(scopePath string) ([]ScopeFile, error) {
	info, err := os.Stat(scopePath)
	if err != nil {
		return nil, fmt.Errorf("stat scope path: %w", err)
	}

	if !info.IsDir() {
		file, err := readScopeFile(scopePath, filepath.Dir(scopePath))
		if err != nil {
			return nil, err
		}
		return []ScopeFile{file}, nil
	}

	files := make([]ScopeFile, 0)
	walkErr := filepath.WalkDir(scopePath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		file, err := readScopeFile(path, scopePath)
		if err != nil {
			if isUnsupportedScopeTypeError(err) {
				return nil
			}
			return err
		}
		files = append(files, file)
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk scope directory: %w", walkErr)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no supported scope files found in %q", scopePath)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return files, nil
}

type unsupportedScopeTypeError struct {
	path string
}

func (e unsupportedScopeTypeError) Error() string {
	return fmt.Sprintf("unsupported scope file type: %s", e.path)
}

func isUnsupportedScopeTypeError(err error) bool {
	_, ok := err.(unsupportedScopeTypeError)
	return ok
}

func readScopeFile(path string, baseDir string) (ScopeFile, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if _, ok := supportedScopeExtensions[ext]; !ok {
		return ScopeFile{}, unsupportedScopeTypeError{path: path}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return ScopeFile{}, fmt.Errorf("read scope file %s: %w", path, err)
	}
	if len(content) > maxScopeFileBytes {
		return ScopeFile{}, fmt.Errorf("scope file too large %s (%d bytes, max %d)", path, len(content), maxScopeFileBytes)
	}

	relPath := path
	if baseDir != "" {
		if rel, relErr := filepath.Rel(baseDir, path); relErr == nil {
			relPath = rel
		}
	}
	relPath = filepath.ToSlash(relPath)

	return ScopeFile{Path: relPath, Content: string(content)}, nil
}
