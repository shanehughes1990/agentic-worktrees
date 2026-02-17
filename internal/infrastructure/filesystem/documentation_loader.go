package filesystem

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	domainservices "github.com/shanehughes1990/agentic-worktrees/internal/domain/services"
)

type DocumentationLoader struct{}

func NewDocumentationLoader() *DocumentationLoader {
	return &DocumentationLoader{}
}

func (l *DocumentationLoader) LoadDocumentationFiles(ctx context.Context, rootDirectory string, maxDepth int) ([]domainservices.DocumentationSourceFile, error) {
	if l == nil {
		return nil, fmt.Errorf("loader cannot be nil")
	}
	if strings.TrimSpace(rootDirectory) == "" {
		return nil, fmt.Errorf("root directory is required")
	}
	if maxDepth < 0 {
		return nil, fmt.Errorf("max depth must be zero or greater")
	}

	root := filepath.Clean(rootDirectory)
	documents := make([]domainservices.DocumentationSourceFile, 0, 32)

	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		depth := 0
		if rel != "." {
			depth = strings.Count(rel, string(filepath.Separator)) + 1
		}

		if entry.IsDir() {
			if rel != "." && depth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		if depth > maxDepth {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".md", ".mdx", ".txt", ".rst":
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			documents = append(documents, domainservices.DocumentationSourceFile{Path: path, Content: string(content)})
		}
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	sort.Slice(documents, func(i int, j int) bool { return documents[i].Path < documents[j].Path })
	return documents, nil
}
