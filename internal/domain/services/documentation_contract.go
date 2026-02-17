package services

import (
	"context"
	"fmt"
	"strings"
)

type DocumentationSourceFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (f DocumentationSourceFile) Validate() error {
	if strings.TrimSpace(f.Path) == "" {
		return fmt.Errorf("path is required")
	}
	return nil
}

type DocumentationFileLoader interface {
	LoadDocumentationFiles(ctx context.Context, rootDirectory string, maxDepth int) ([]DocumentationSourceFile, error)
}
