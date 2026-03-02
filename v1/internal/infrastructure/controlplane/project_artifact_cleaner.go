package controlplane

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProjectArtifactCleanerConfig struct {
	ProjectsPath string
}

type ProjectArtifactCleaner struct {
	projectsPath string
}

func NewProjectArtifactCleaner(config ProjectArtifactCleanerConfig) (*ProjectArtifactCleaner, error) {
	projectsPath := strings.TrimSpace(config.ProjectsPath)
	if projectsPath == "" {
		return nil, fmt.Errorf("projects path is required")
	}
	return &ProjectArtifactCleaner{projectsPath: filepath.Clean(projectsPath)}, nil
}

func (cleaner *ProjectArtifactCleaner) CleanupProjectArtifacts(ctx context.Context, setup applicationcontrolplane.ProjectSetup) error {
	if cleaner == nil {
		return fmt.Errorf("project artifact cleaner is not configured")
	}
	projectID := strings.TrimSpace(setup.ProjectID)
	if projectID == "" {
		return fmt.Errorf("project_id is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	projectPath := filepath.Join(cleaner.projectsPath, projectID)
	if err := os.RemoveAll(projectPath); err != nil {
		return fmt.Errorf("remove project artifact path %q: %w", projectPath, err)
	}
	return nil
}

var _ applicationcontrolplane.ProjectCleanupManager = (*ProjectArtifactCleaner)(nil)
